// 执行服务核心业务逻辑层。
//
// 职责：管理执行的完整生命周期——创建、认领、状态转换（pending→running→succeeded/failed）、日志追加、重试物化。
// 依赖三层存储：PostgreSQL（执行状态，通过 ExecutionRepository）、MongoDB（执行日志，通过 LogRepository）、Redis（任务队列，通过 Queue）。
//
// 关键设计决策：
//  1. 创建流程（Create）使用试运行模式——先写 PostgreSQL，再写 MongoDB，最后入队 Redis；任一步失败则回滚（rollbackCreate）
//  2. 认领循环（ClaimNext）使用 Claim → MarkRunning → Ack/Release 三阶段模式，保证队列安全
//  3. 重试物化（MaterializeRetry）使用 retried_at 作为乐观锁，在单次 UPDATE...RETURNING 中原子完成锁定和读取
//  4. Complete/Fail 对已终态的执行做幂等处理——如果状态已经是 succeeded/failed，只补做 Ack 操作
//
// 不负责：执行资源限制的强制执行（由节点服务处理）、爬虫版本的解析（由 SpiderVersionResolver 处理）。
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"crawler-platform/apps/execution-service/internal/model"
	"github.com/google/uuid"
)

// 预定义的业务错误，用于在 repo → service → api 层之间传递结构化的错误语义。
var ErrExecutionNotFound = errors.New("execution not found")
var ErrInvalidExecutionState = errors.New("invalid execution state transition")
var ErrSpiderVersionNotFound = errors.New("spider version not found")

// ExecutionService 执行服务核心结构体，组合三个外部依赖接口。
// 使用接口而非具体类型，便于单元测试时注入 fake 实现。
type ExecutionService struct {
	execRepo ExecutionRepository
	logRepo  LogRepository
	queue    Queue
}

// ExecutionRepository 定义执行状态持久化接口，由 PostgreSQL 实现。
type ExecutionRepository interface {
	Create(ctx context.Context, exec model.Execution) (model.Execution, error)
	Get(ctx context.Context, id string) (model.Execution, error)
	List(ctx context.Context, orgID string, limit, offset int, status string) (*ListResult, error)
	MarkRunning(ctx context.Context, id, nodeID string, startedAt time.Time) (model.Execution, error)
	Complete(ctx context.Context, id string, finishedAt time.Time) (model.Execution, error)
	Fail(ctx context.Context, id, errorMessage string, finishedAt time.Time) (model.Execution, error)
	ClaimNextRetryCandidate(ctx context.Context, now time.Time) (model.Execution, bool, error)
	ResetRetryClaim(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

// ListResult 包含分页执行列表查询的结果。
type ListResult struct {
	Executions []model.Execution
	Total      int64
}

// LogRepository 定义日志持久化接口，由 MongoDB 实现。
type LogRepository interface {
	Init(ctx context.Context, executionID string) error
	Append(ctx context.Context, entry model.ExecutionLog) error
	List(ctx context.Context, executionID string) ([]model.ExecutionLog, error)
}

// Queue 定义任务队列接口，由 Redis 实现。
type Queue interface {
	Enqueue(ctx context.Context, executionID string) error
	Claim(ctx context.Context) (string, error)
	Ack(ctx context.Context, executionID string) error
	Release(ctx context.Context, executionID string) error
}

// CreateManualInput 是手动创建执行的输入参数（简化版，固定 triggerSource = "manual"）。
type CreateManualInput struct {
	ProjectID string
	SpiderID  string
	Image     string
	Command   []string
}

// CreateExecutionInput 是创建执行的完整输入参数。
// RetryOfExecutionID 非空时表示这是一个重试执行，形成执行链用于追踪血缘关系。
// TriggerSource 可选，为空时默认 "manual"。
type CreateExecutionInput struct {
	ProjectID          string
	SpiderID           string
	SpiderVersion      string
	RegistryAuthRef    string
	Image              string
	Command            []string
	TriggerSource      string
	OrgID              string
	CpuCores           float64
	MemoryMB           int
	TimeoutSeconds     int
	RetryLimit         int
	RetryCount         int
	RetryDelaySeconds  int
	RetryOfExecutionID string
}

// NewExecutionService 创建执行服务实例。所有依赖必须非 nil。
func NewExecutionService(execRepo ExecutionRepository, logRepo LogRepository, queue Queue) *ExecutionService {
	return &ExecutionService{execRepo: execRepo, logRepo: logRepo, queue: queue}
}

// CreateManual 以手动模式创建执行，固定 triggerSource = "manual"。
// 是对 Create 的简化封装，用于手工触发的爬虫任务。
func (s *ExecutionService) CreateManual(ctx context.Context, input CreateManualInput) (model.Execution, error) {
	return s.Create(ctx, CreateExecutionInput{
		ProjectID:     input.ProjectID,
		SpiderID:      input.SpiderID,
		Image:         input.Image,
		Command:       input.Command,
		TriggerSource: "manual",
	})
}

// Create 执行创建流程：构建 Execution 模型 → 写入 PostgreSQL → 初始化 MongoDB 日志 → 入队 Redis。
//
// 设计上采用"试运行回滚"模式而非事务：
//  1. 生成 UUID，构建 pending 状态的 Execution 对象
//  2. 写入 PostgreSQL（executions 表）
//  3. 初始化 MongoDB 日志存储（为 executionID 预留日志集合/索引）
//  4. 将 executionID 加入 Redis pending 队列
//
// 如果步骤 3 或 4 失败，调用 rollbackCreate 从 PostgreSQL 删除已写入的记录。
// Command 字段通过 append([]string(nil), input.Command...) 防御性拷贝，避免外部修改影响内部状态。
//
// 为什么不使用分布式事务：两阶段提交或 Saga 模式引入了额外的复杂性，
// 而这里的回滚策略（删除已写入的 PostgreSQL 记录）在大多数失败场景下（MongoDB 或 Redis 不可用）是安全的，
// 且实现简单。极低的竞态窗口内（Delete 失败且回滚也失败）会将错误 join 返回供调用方决策。
func (s *ExecutionService) Create(ctx context.Context, input CreateExecutionInput) (model.Execution, error) {
	triggerSource := input.TriggerSource
	if triggerSource == "" {
		triggerSource = "manual"
	}

	exec := model.Execution{
		ID:                uuid.NewString(),
		ProjectID:         input.ProjectID,
		OrganizationID:    input.OrgID,
		SpiderID:          input.SpiderID,
		SpiderVersion:     input.SpiderVersion,
		RegistryAuthRef:   input.RegistryAuthRef,
		Status:            "pending",
		TriggerSource:     triggerSource,
		Image:             input.Image,
		Command:           append([]string(nil), input.Command...),
		CpuCores:          input.CpuCores,
		MemoryMB:          input.MemoryMB,
		TimeoutSeconds:    input.TimeoutSeconds,
		RetryLimit:        input.RetryLimit,
		RetryCount:        input.RetryCount,
		RetryDelaySeconds: input.RetryDelaySeconds,
		RetryOfExecutionID: input.RetryOfExecutionID,
		CreatedAt:         time.Now().UTC(),
	}

	created, err := s.execRepo.Create(ctx, exec)
	if err != nil {
		return model.Execution{}, err
	}
	if err := s.logRepo.Init(ctx, created.ID); err != nil {
		return model.Execution{}, s.rollbackCreate(ctx, created.ID, err)
	}
	if err := s.queue.Enqueue(ctx, created.ID); err != nil {
		return model.Execution{}, s.rollbackCreate(ctx, created.ID, err)
	}

	return created, nil
}

// MaterializeRetry 将符合条件的 failed 执行物化为新的重试执行。
//
// 流程：
//  1. 调用 ClaimNextRetryCandidate 原子查找并锁定下一个重试候选
//     - 锁定机制：设置 retried_at = now，防止并发重试同一执行
//  2. 如果未找到候选（所有失败执行都已重试或不满足条件）→ 返回 (零值, false, nil)
//  3. 基于候选执行创建新的执行：
//     - triggerSource = "retry"
//     - retryCount = candidate.RetryCount + 1（递增）
//     - retryOfExecutionID = candidate.ID（形成执行链）
//     - 其他字段（image、command、retryLimit 等）从候选复制
//  4. 如果 Create 失败 → 调用 ResetRetryClaim 回滚乐观锁，将 retried_at 复位为 NULL
//  5. 如果 Create 成功 → 新执行自动进入 pending 队列，等待节点认领
//
// 为什么 retryCount 在 MaterializeRetry 中递增而非在 Fail 中：
// Fail 阶段的错误可能是暂时的（如 OOM），递增 retryCount 应该在确定要创建重试执行时才进行，
// 这样可以支持"重试策略调整"——例如在 Fail 之后、MaterializeRetry 之前修改 retryLimit。
func (s *ExecutionService) MaterializeRetry(ctx context.Context) (model.Execution, bool, error) {
	candidate, ok, err := s.execRepo.ClaimNextRetryCandidate(ctx, time.Now().UTC())
	if err != nil || !ok {
		return model.Execution{}, ok, err
	}

	created, err := s.Create(ctx, CreateExecutionInput{
		ProjectID:          candidate.ProjectID,
		SpiderID:           candidate.SpiderID,
		SpiderVersion:      candidate.SpiderVersion,
		RegistryAuthRef:    candidate.RegistryAuthRef,
		Image:              candidate.Image,
		Command:            candidate.Command,
		TriggerSource:      "retry",
		CpuCores:           candidate.CpuCores,
		MemoryMB:           candidate.MemoryMB,
		TimeoutSeconds:     candidate.TimeoutSeconds,
		RetryLimit:         candidate.RetryLimit,
		RetryCount:         candidate.RetryCount + 1,
		RetryDelaySeconds:  candidate.RetryDelaySeconds,
		RetryOfExecutionID: candidate.ID,
	})
	if err != nil {
		if resetErr := s.execRepo.ResetRetryClaim(ctx, candidate.ID); resetErr != nil {
			return model.Execution{}, false, errors.Join(err, resetErr)
		}
		return model.Execution{}, false, err
	}

	return created, true, nil
}

// Get 查询执行详情，会自动附带该执行的全部日志（通过 logRepo.List 查询 MongoDB）。
// 调用方使用 errors.Is(err, ErrExecutionNotFound) 来区分"执行不存在"和其他错误。
func (s *ExecutionService) Get(ctx context.Context, id string) (model.Execution, error) {
	exec, err := s.execRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		return model.Execution{}, err
	}

	logs, err := s.logRepo.List(ctx, id)
	if err != nil {
		return model.Execution{}, err
	}
	exec.Logs = logs
	return exec, nil
}

// ListExecutions 分页查询执行记录，支持按 status 过滤。
func (s *ExecutionService) ListExecutions(ctx context.Context, orgID string, limit, offset int, status string) (*ListResult, error) {
	return s.execRepo.List(ctx, orgID, limit, offset, status)
}

// AppendLog 向指定执行追加一条日志。
// 先校验执行是否存在（通过 execRepo.Get），再生成 UUID 并写入 MongoDB。
// 执行不存在时返回 ErrExecutionNotFound——即使 MongoDB 写入成功，没有关联的执行记录也是无效日志。
func (s *ExecutionService) AppendLog(ctx context.Context, executionID, message string) (model.ExecutionLog, error) {
	if _, err := s.execRepo.Get(ctx, executionID); errors.Is(err, ErrExecutionNotFound) {
		return model.ExecutionLog{}, ErrExecutionNotFound
	} else if err != nil {
		return model.ExecutionLog{}, err
	}

	entry := model.ExecutionLog{
		ID:          uuid.NewString(),
		ExecutionID: executionID,
		Message:     message,
		CreatedAt:   time.Now().UTC(),
	}
	return entry, s.logRepo.Append(ctx, entry)
}

// GetLogs 查询指定执行的全部日志，按 created_at 升序返回。
// 与 Get 不同的是，GetLogs 不附带执行本身的状态信息，仅返回日志切片。
func (s *ExecutionService) GetLogs(ctx context.Context, executionID string) ([]model.ExecutionLog, error) {
	if _, err := s.execRepo.Get(ctx, executionID); err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return nil, ErrExecutionNotFound
		}
		return nil, err
	}

	logs, err := s.logRepo.List(ctx, executionID)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// ClaimNext 从 Redis 队列认领下一个待执行任务，并将其状态转换为 running。
//
// 认领循环的完整语义（for 循环内的每一步）：
//  1. queue.Claim：原子地从 pending 队列取一个 ID，移入 inflight 队列
//     - 返回空字符串 → 队列无任务，返回 (零值, false, nil)
//  2. transitionToRunning：尝试将执行状态从 pending 更新为 running
//     - 成功 → 返回执行记录，队列中该 ID 保留在 inflight 等待最终 Ack/Release
//     - ErrExecutionNotFound → 执行已被删除 → Ack 清理 inflight 中的无效 ID，continue 尝试下一个
//     - ErrInvalidExecutionState → 状态已被其他节点改变 → Ack 清理，continue
//     - 其他错误（DB 故障）→ Release 归还到 pending 队列，返回错误让调用方重试
//
// 为什么使用 Ack 而非 Release 处理 NotFound/InvalidState：
// 这两种情况是永久性错误——执行已被删除或状态已不可逆——
// 将无效 ID 归还到队列只会导致无限循环。Ack 从 inflight 删除它，让循环继续处理下一个有效任务。
//
// 为什么使用 Release 处理其他错误：
// DB 暂时不可用是暂时性故障——Release 将任务放回 pending 队列头部，
// 等 DB 恢复后可以被重新认领。使用 LPush 放回头部是为了避免饥饿：
// 如果反复 RPush 到尾部，而故障期间新任务不断入队，该任务可能永远排不到前面。
func (s *ExecutionService) ClaimNext(ctx context.Context, nodeID string) (model.Execution, bool, error) {
	for {
		executionID, err := s.queue.Claim(ctx)
		if err != nil {
			return model.Execution{}, false, err
		}
		if executionID == "" {
			return model.Execution{}, false, nil
		}

		exec, err := s.transitionToRunning(ctx, executionID, nodeID, time.Now().UTC())
		if err != nil {
			if errors.Is(err, ErrExecutionNotFound) || errors.Is(err, ErrInvalidExecutionState) {
				if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
					return model.Execution{}, false, errors.Join(err, fmt.Errorf("ack claimed execution %s: %w", executionID, ackErr))
				}
				continue
			}
			if releaseErr := s.queue.Release(ctx, executionID); releaseErr != nil {
				return model.Execution{}, false, errors.Join(err, fmt.Errorf("release execution %s: %w", executionID, releaseErr))
			}
			return model.Execution{}, false, err
		}
		return exec, true, nil
	}
}

// Start 确认执行已在工作节点上启动运行。
// 校验逻辑：
//  1. 执行状态必须为 running——pending 状态的执行不能直接启动（应先通过 ClaimNext）
//  2. nodeID 必须匹配——只有认领该执行的工作节点才能确认启动
//
// 如果执行已经是 running 且 nodeID 为空（历史数据兼容），允许任何节点确认启动。
// 返回值：成功时返回执行记录；状态不匹配时返回 ErrInvalidExecutionState。
func (s *ExecutionService) Start(ctx context.Context, executionID, nodeID string) (model.Execution, error) {
	exec, err := s.execRepo.Get(ctx, executionID)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		return model.Execution{}, err
	}
	if exec.Status != "running" {
		return model.Execution{}, ErrInvalidExecutionState
	}
	if exec.NodeID != "" && exec.NodeID != nodeID {
		return model.Execution{}, ErrInvalidExecutionState
	}
	return exec, nil
}

// Complete 将执行状态从 running 转换为 succeeded，并 Ack 清理 inflight 队列。
//
// 幂等性处理：
// 如果转换失败且当前状态已经是 succeeded（例如 Complete 被重复调用）：
//   - 不返回错误，而是补做 Ack 操作并返回当前状态
//   - 为什么需要补 Ack：第一次 Complete 可能成功更新 PostgreSQL 但 Ack 失败（Redis 暂时不可用），
//     导致 inflight 队列中残留该执行 ID。重试 Complete 时检测到状态已是 succeeded，
//     补发 Ack 清理残留，保证最终一致性。
//
// 如果当前状态既不是 running 也不是 succeeded（例如 pending）→ 返回 ErrInvalidExecutionState。
func (s *ExecutionService) Complete(ctx context.Context, executionID string) (model.Execution, error) {
	exec, err := s.execRepo.Complete(ctx, executionID, time.Now().UTC())
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		if errors.Is(err, ErrInvalidExecutionState) {
			current, currentErr := s.execRepo.Get(ctx, executionID)
			if currentErr != nil {
				if errors.Is(currentErr, ErrExecutionNotFound) {
					return model.Execution{}, ErrExecutionNotFound
				}
				return model.Execution{}, currentErr
			}
			if current.Status == "succeeded" {
				if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
					return model.Execution{}, fmt.Errorf("ack completed execution %s: %w", executionID, ackErr)
				}
				return current, nil
			}
			return model.Execution{}, ErrInvalidExecutionState
		}
		return model.Execution{}, err
	}
	if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
		return model.Execution{}, fmt.Errorf("ack completed execution %s: %w", executionID, ackErr)
	}
	return exec, nil
}

// Fail 将执行状态从 running 转换为 failed，记录错误信息，并 Ack 清理 inflight 队列。
//
// 与 Complete 相同的幂等性模式：
// 如果转换失败且当前状态已经是 failed → 补做 Ack 并返回当前状态（处理 Ack 失败的补偿）。
// 如果当前状态既不是 running 也不是 failed → 返回 ErrInvalidExecutionState。
//
// errorMessage 记录了失败原因，在重试物化时可通过查询 error_message 字段了解失败历史。
func (s *ExecutionService) Fail(ctx context.Context, executionID, errorMessage string) (model.Execution, error) {
	exec, err := s.execRepo.Fail(ctx, executionID, errorMessage, time.Now().UTC())
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		if errors.Is(err, ErrInvalidExecutionState) {
			current, currentErr := s.execRepo.Get(ctx, executionID)
			if currentErr != nil {
				if errors.Is(currentErr, ErrExecutionNotFound) {
					return model.Execution{}, ErrExecutionNotFound
				}
				return model.Execution{}, currentErr
			}
			if current.Status == "failed" {
				if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
					return model.Execution{}, fmt.Errorf("ack failed execution %s: %w", executionID, ackErr)
				}
				return current, nil
			}
			return model.Execution{}, ErrInvalidExecutionState
		}
		return model.Execution{}, err
	}
	if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
		return model.Execution{}, fmt.Errorf("ack failed execution %s: %w", executionID, ackErr)
	}
	return exec, nil
}

// rollbackCreate 回滚失败的创建操作：从 PostgreSQL 删除已写入的执行记录。
//
// 调用场景：Create 流程中，PostgreSQL 写入成功后 MongoDB Init 或 Redis Enqueue 失败。
// 如果删除也失败（Delete 返回错误），使用 errors.Join 将原始错误和回滚错误合并返回。
// 这样调用方可以同时看到"为什么会失败"（cause）和"为什么清理也失败了"（deleteErr）。
func (s *ExecutionService) rollbackCreate(ctx context.Context, executionID string, cause error) error {
	if deleteErr := s.execRepo.Delete(ctx, executionID); deleteErr != nil {
		return errors.Join(cause, fmt.Errorf("rollback execution %s: %w", executionID, deleteErr))
	}
	return cause
}

// transitionToRunning 是 pending → running 状态转换的核心逻辑，由 ClaimNext 调用。
//
// 状态转换规则：
//  1. 执行不存在 → ErrExecutionNotFound → 调用方 Ack 清理 inflight 队列
//  2. 执行已经是 running → 直接返回当前状态（幂等：重复认领不报错）
//     这种场景可能发生在：前一个认领者成功 MarkRunning 但在返回前崩溃，当前认领者重新认领到同一 ID
//  3. 执行是 pending → 调用 MarkRunning（条件 UPDATE WHERE status = 'pending'）
//     - 成功 → 返回 updated 执行
//     - ErrInvalidExecutionState → 状态已被并发修改 → 返回错误供调用方 Ack 清理
//  4. 执行是其他终态（succeeded/failed）→ ErrInvalidExecutionState
//
// 为什么先 Get 再 MarkRunning：
// 条件 UPDATE 本身已经是原子的（WHERE status = 'pending'），但先 Get 可以区分以下场景：
//   a. 执行不存在（404）
//   b. 执行已经是 running（幂等返回）
//   c. 执行是其他状态但 MarkRunning 的 WHERE 子句已确保原子性
// 这样可以避免在 MarkRunning 失败后再做查询来区分错误类型。
func (s *ExecutionService) transitionToRunning(ctx context.Context, executionID, nodeID string, startedAt time.Time) (model.Execution, error) {
	current, err := s.execRepo.Get(ctx, executionID)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		return model.Execution{}, err
	}
	if current.Status == "running" {
		return current, nil
	}
	if current.Status != "pending" {
		return model.Execution{}, ErrInvalidExecutionState
	}

	exec, err := s.execRepo.MarkRunning(ctx, executionID, nodeID, startedAt)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		if errors.Is(err, ErrInvalidExecutionState) {
			return model.Execution{}, ErrInvalidExecutionState
		}
		return model.Execution{}, err
	}
	return exec, nil
}
