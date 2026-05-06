// Package service 是调度服务（scheduler-service）的核心业务逻辑层。
//
// 该文件负责：
//   - Schedule CRUD：创建（含参数校验和 UUID 生成）和列表查询。
//   - MaterializeDue：调度物化主循环，扫描所有已启用调度，根据 cron 表达式计算下一次触发时间，
//     调用 AdvanceLastMaterialized（CAS 游标）去重后，通过 ExecutionClient 创建执行记录。
//   - Run：后台定时循环，交替执行"正向物化"和"重试物化"两个阶段。
//   - 重试物化：调用 execution-service 的 /internal/v1/executions/retries/materialize，
//     让执行服务将到期可重试的执行记录重新物化。
//
// 与谁交互：
//   - Repository（接口）：持久化 Schedule 和 last_materialized_at 游标。
//   - ExecutionClient（接口）：通过 HTTP 调用 execution-service 创建执行和触发重试物化。
//   - robfig/cron：解析 5 字段 cron 表达式，计算下一次触发时间。
//
// 不负责：
//   - 不执行实际的爬虫任务（由 execution-service 的 workers 负责）。
//   - 不管理重试策略（RetryLimit/RetryDelaySeconds 仅透传给 execution-service）。
//   - 不做 HTTP 路由（由 api 包负责）。
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"crawler-platform/apps/scheduler-service/internal/model"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// ErrInvalidSchedule 表示创建调度时参数校验失败（必填字段为空）。
var ErrInvalidSchedule = errors.New("invalid schedule")

// maxCatchUpRunsPerPoll 限制每次物化循环中单个调度最多追赶的次数（含正向物化和重试物化），
// 防止服务长时间停摆后恢复时一次性生成大量执行记录。
const maxCatchUpRunsPerPoll = 16

// SchedulerService 是调度服务的核心结构体，持有 Repository、ExecutionClient、cron 解析器和时钟源。
// now 字段默认为 time.Now().UTC()，可通过 WithNow 选项注入固定时钟，便于测试。
type SchedulerService struct {
	repo            Repository
	executionClient ExecutionClient
	parser          cron.Parser
	now             func() time.Time
}

// Repository 定义调度持久化接口，支持 PostgresRepository 和 memoryRepository 两种实现。
type Repository interface {
	Create(ctx context.Context, schedule model.Schedule) error
	List(ctx context.Context, orgID string, limit, offset int) ([]model.Schedule, error)
	AdvanceLastMaterialized(ctx context.Context, id string, previous *time.Time, next time.Time) (bool, error)
	RestoreLastMaterialized(ctx context.Context, id string, previous *time.Time, current time.Time) error
}

// ExecutionClient 定义与 execution-service 的通信接口。
// Create 用于创建新的执行记录；MaterializeRetry 触发执行服务的重试物化。
type ExecutionClient interface {
	Create(ctx context.Context, input CreateExecutionInput) (string, error)
	MaterializeRetry(ctx context.Context) (bool, error)
}

// CreateExecutionInput 是创建执行记录时传给 ExecutionClient 的参数结构体。
// TriggerSource 固定为 "scheduled"，RetryCount 初始为 0。
type CreateExecutionInput struct {
	ScheduleID         string
	ProjectID          string
	SpiderID           string
	SpiderVersion      string
	RegistryAuthRef    string
	Image              string
	Command            []string
	TriggerSource      string
	ScheduledFor       time.Time
	RetryLimit         int
	RetryCount         int
	RetryDelaySeconds  int
}

// Option 定义 SchedulerService 的配置选项函数类型，用于依赖注入和测试。
type Option func(*SchedulerService)

// memoryRepository 是 Repository 的内存实现，用于单元测试。
type memoryRepository struct {
	mu        sync.Mutex
	schedules []model.Schedule
}

// noopExecutionClient 是 ExecutionClient 的空操作实现，当未提供真实客户端时作为默认值。
type noopExecutionClient struct{}

// HTTPExecutionClient 通过 HTTP 调用 execution-service 的 REST API。
// baseURL 为 execution-service 地址，internalToken 用于内部端点认证。
type HTTPExecutionClient struct {
	baseURL       string
	internalToken string
	client        *http.Client
}

// Create 将 Schedule 追加到内存切片中（线程安全）。
func (r *memoryRepository) Create(_ context.Context, schedule model.Schedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schedules = append(r.schedules, schedule)
	return nil
}

// List 分页返回内存中 Schedule 的副本（线程安全）。
func (r *memoryRepository) List(_ context.Context, orgID string, limit, offset int) ([]model.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	filtered := r.schedules
	if orgID != "" {
		filtered = nil
		for _, s := range r.schedules {
			if s.OrganizationID == orgID {
				filtered = append(filtered, s)
			}
		}
	}

	if limit <= 0 {
		limit = 20
	}
	if offset >= len(filtered) {
		return []model.Schedule{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	schedules := make([]model.Schedule, end-offset)
	copy(schedules, filtered[offset:end])
	return schedules, nil
}

// AdvanceLastMaterialized 在内存中模拟 CAS 游标推进（线程安全）。
// 与 PostgresRepository 语义一致：仅当 current == previous 时才更新。
func (r *memoryRepository) AdvanceLastMaterialized(_ context.Context, id string, previous *time.Time, next time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, schedule := range r.schedules {
		if schedule.ID != id {
			continue
		}
		if !timesEqual(schedule.LastMaterializedAt, previous) {
			return false, nil
		}
		r.schedules[i].LastMaterializedAt = &next
		return true, nil
	}
	return false, nil
}

// RestoreLastMaterialized 在内存中回滚游标（线程安全）。
func (r *memoryRepository) RestoreLastMaterialized(_ context.Context, id string, previous *time.Time, current time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, schedule := range r.schedules {
		if schedule.ID != id {
			continue
		}
		if schedule.LastMaterializedAt == nil || !schedule.LastMaterializedAt.Equal(current) {
			return nil
		}
		r.schedules[i].LastMaterializedAt = previous
		return nil
	}
	return nil
}

func (noopExecutionClient) Create(_ context.Context, _ CreateExecutionInput) (string, error) {
	return "", nil
}

func (noopExecutionClient) MaterializeRetry(_ context.Context) (bool, error) {
	return false, nil
}

// NewHTTPExecutionClient 创建 HTTPExecutionClient，设置 10 秒超时。
func NewHTTPExecutionClient(baseURL, internalToken string) *HTTPExecutionClient {
	return &HTTPExecutionClient{
		baseURL:       baseURL,
		internalToken: internalToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Create 向 execution-service 发送 POST /api/v1/executions 创建执行记录。
// 将 CreateExecutionInput 序列化为 JSON，期望返回 201 及执行 ID。
func (c *HTTPExecutionClient) Create(ctx context.Context, input CreateExecutionInput) (string, error) {
	body, err := json.Marshal(map[string]any{
		"projectId":         input.ProjectID,
		"spiderId":          input.SpiderID,
		"spiderVersion":     input.SpiderVersion,
		"registryAuthRef":   input.RegistryAuthRef,
		"image":             input.Image,
		"command":           input.Command,
		"triggerSource":     input.TriggerSource,
		"scheduleId":        input.ScheduleID,
		"scheduledFor":      input.ScheduledFor.Format(time.RFC3339),
		"retryLimit":        input.RetryLimit,
		"retryCount":        input.RetryCount,
		"retryDelaySeconds": input.RetryDelaySeconds,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/executions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("execution create returned status %d", resp.StatusCode)
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ID == "" {
		return "", errors.New("execution create returned empty id")
	}
	return payload.ID, nil
}

// MaterializeRetry 调用 execution-service 的内部重试物化端点。
// POST /internal/v1/executions/retries/materialize，使用 X-Internal-Token 做服务间认证。
// 返回 true（201 Created）表示有新的重试被物化；返回 false（204 No Content）表示无更多重试。
func (c *HTTPExecutionClient) MaterializeRetry(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/executions/retries/materialize", nil)
	if err != nil {
		return false, err
	}
	if c.internalToken != "" {
		req.Header.Set("X-Internal-Token", c.internalToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		return true, nil
	case http.StatusNoContent:
		return false, nil
	default:
		return false, fmt.Errorf("retry materialization returned status %d", resp.StatusCode)
	}
}

// WithNow 注入自定义时钟函数，主要用于单元测试中固定时间。
func WithNow(now func() time.Time) Option {
	return func(s *SchedulerService) {
		s.now = now
	}
}

// NewSchedulerService 创建 SchedulerService 实例。
// repo 和 executionClient 若为 nil 则使用内存实现作为默认值，便于测试。
// cron 解析器支持 5 字段标准格式（分 时 日 月 星期）。
func NewSchedulerService(repo Repository, executionClient ExecutionClient, options ...Option) *SchedulerService {
	if repo == nil {
		repo = &memoryRepository{}
	}
	if executionClient == nil {
		executionClient = noopExecutionClient{}
	}

	svc := &SchedulerService{
		repo:            repo,
		executionClient: executionClient,
		parser:          cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		now:             func() time.Time { return time.Now().UTC() },
	}

	for _, option := range options {
		if option != nil {
			option(svc)
		}
	}
	return svc
}

// Create 创建一条新的定时调度规则。
//
// 参数校验：projectID、spiderID、name、cronExpr、image 为必填字段，任缺其一返回 ErrInvalidSchedule。
// 命令切片会深拷贝后存储，防止外部修改。LastMaterializedAt 初始为 nil，表示从未物化。
// 调度 ID 通过 UUID v4 生成，创建时间取自 now() 时钟。
func (s *SchedulerService) Create(orgID, projectID, spiderID, spiderVersion, registryAuthRef, name, cronExpr, image string, command []string, enabled bool, retryLimit, retryDelaySeconds int) (model.Schedule, error) {
	if projectID == "" || spiderID == "" || name == "" || cronExpr == "" || image == "" {
		return model.Schedule{}, ErrInvalidSchedule
	}

	createdAt := s.now().UTC()
	schedule := model.Schedule{
		ID:              uuid.NewString(),
		ProjectID:       projectID,
		SpiderID:        spiderID,
		SpiderVersion:   spiderVersion,
		RegistryAuthRef: registryAuthRef,
		Name:            name,
		CronExpr:        cronExpr,
		Enabled:         enabled,
		Image:           image,
		Command:         append([]string(nil), command...),
		RetryLimit:      retryLimit,
		RetryDelaySeconds: retryDelaySeconds,
		CreatedAt:       createdAt,
	}

	if err := s.repo.Create(context.Background(), schedule); err != nil {
		return model.Schedule{}, err
	}
	return schedule, nil
}

// List 分页返回调度记录，limit <= 0 时由 repo 层使用默认值。
func (s *SchedulerService) List(orgID string, limit, offset int) ([]model.Schedule, error) {
	return s.repo.List(context.Background(), orgID, limit, offset)
}

// MaterializeDue 是调度物化的核心函数，扫描所有已启用调度并将到期时间点物化为执行记录。
//
// 物化流程（对每条启用的调度）：
//   1. 解析 cron 表达式，计算下一次触发时间。
//   2. 确定物化起点 base：
//      - 若从未物化过（LastMaterializedAt == nil）：base = CreatedAt - 1 minute（确保首个触发点不被遗漏）。
//      - 否则：base = LastMaterializedAt。
//   3. 从 base 开始逐步调用 cron.Next(base)，当 next > now 时停止。
//   4. 对每个 next <= now 的时间点，调用 AdvanceLastMaterialized 做 CAS 更新：
//      - 获得 CAS 锁（claimed == true）：创建执行记录。
//      - 未获得 CAS 锁：说明其他实例已处理，跳过该时间点。
//   5. 如果创建执行失败，调用 RestoreLastMaterialized 回滚游标，确保下次重试。
//   6. 每个调度最多追赶 maxCatchUpRunsPerPoll(16) 次，防止服务长时间停摆后一次性生成大量执行。
//
// 去重原理：
//   AdvanceLastMaterialized 使用 WHERE last_materialized_at = <previous> 的 CAS 条件更新，
//   确保同一时间点只有一个实例能成功物化。多实例部署时，只有第一个成功的实例会创建执行记录。
//
// 返回值：本次物化生成的执行记录总数。
func (s *SchedulerService) MaterializeDue(ctx context.Context) (int, error) {
	schedules, err := s.repo.List(ctx, "", 0, 0)
	if err != nil {
		return 0, err
	}

	now := s.now().UTC().Truncate(time.Minute)
	materialized := 0

	for _, schedule := range schedules {
		if !schedule.Enabled {
			continue
		}

		spec, err := s.parser.Parse(schedule.CronExpr)
		if err != nil {
			return materialized, err
		}

		base := schedule.CreatedAt.UTC().Add(-time.Minute)
		if schedule.LastMaterializedAt != nil {
			base = schedule.LastMaterializedAt.UTC()
		}

		for catchUp := 0; catchUp < maxCatchUpRunsPerPoll; catchUp++ {
			next := spec.Next(base).UTC()
			if next.After(now) {
				break
			}

			previous := schedule.LastMaterializedAt
			claimed, err := s.repo.AdvanceLastMaterialized(ctx, schedule.ID, previous, next)
			if err != nil {
				return materialized, err
			}
			if !claimed {
				break
			}

			_, err = s.executionClient.Create(ctx, CreateExecutionInput{
				ScheduleID:        schedule.ID,
				ProjectID:         schedule.ProjectID,
				SpiderID:          schedule.SpiderID,
				Image:             schedule.Image,
				Command:           append([]string(nil), schedule.Command...),
				TriggerSource:     "scheduled",
				ScheduledFor:      next,
				RetryLimit:        schedule.RetryLimit,
				RetryCount:        0,
				RetryDelaySeconds: schedule.RetryDelaySeconds,
			})
			if err != nil {
				if rollbackErr := s.repo.RestoreLastMaterialized(ctx, schedule.ID, previous, next); rollbackErr != nil {
					return materialized, errors.Join(err, rollbackErr)
				}
				return materialized, err
			}

			schedule.LastMaterializedAt = &next
			base = next
			materialized++
		}
	}

	return materialized, nil
}

// Run 启动后台物化循环，按 pollInterval 间隔重复执行。
//
// 每轮循环分两个阶段：
//   第一阶段：调用 MaterializeDue 做正向物化（扫描调度表，生成到期执行记录）。
//   第二阶段：循环调用 executionClient.MaterializeRetry 做重试物化，
//     让执行服务将可重试的失败执行重新调度。每次调用若返回 201（有新重试被物化）则继续，
//     返回 204（无更多重试）则退出；最多追赶 maxCatchUpRunsPerPoll 次防止风暴。
//
// 退出条件：ctx 被取消时返回 ctx.Err()；发生错误时立即返回 error。
func (s *SchedulerService) Run(ctx context.Context, pollInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if _, err := s.MaterializeDue(ctx); err != nil {
			return err
		}
		for retryBatch := 0; retryBatch < maxCatchUpRunsPerPoll; retryBatch++ {
			materializedRetry, err := s.executionClient.MaterializeRetry(ctx)
			if err != nil {
				return err
			}
			if !materializedRetry {
				break
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// timesEqual 安全比较两个 *time.Time 指针，两者都为 nil 时视为相等。
func timesEqual(left, right *time.Time) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return left.Equal(*right)
	}
}
