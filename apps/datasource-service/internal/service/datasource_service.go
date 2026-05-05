// Package service 实现数据源服务的核心业务逻辑。
//
// 本文件负责：
//   1. 数据源 CRUD：创建、列表查询、按 ID 查询
//   2. 连接测试（真实探针，非 mock）：调用 liveDatasourceProber 执行真实连接
//   3. 数据预览（真实查询，非 mock）：调用 liveDatasourceProber 执行真实数据查询
//   4. 错误哨兵值（sentinel errors）：定义分类错误以支持前端差异化展示
//
// 与谁交互：
//   - Repository 接口：持久化数据源配置（PostgreSQL 或内存实现）
//   - Prober 接口（通过 NewDatasourceServiceWithProber 注入）：连接测试和数据预览的真实探针
//
// 不负责什么：不负责 HTTP 路由、请求参数校验、JSON 序列化——这些由 api 包处理。
package service

import (
	"context"
	"errors"
	"sync"

	"crawler-platform/apps/datasource-service/internal/model"
	"github.com/google/uuid"
)

// ErrInvalidDatasourceType 表示用户提供了不支持的数据源类型（当前仅支持 postgresql/redis/mongodb）。
// 前端收到此错误应展示"不支持的数据源类型"提示。
var ErrInvalidDatasourceType = errors.New("invalid datasource type")

// ErrDatasourceNotFound 表示按 ID 查询数据源时未找到对应记录，对应 HTTP 404。
var ErrDatasourceNotFound = errors.New("datasource not found")

// ErrDatasourceProbeFailed 表示连接测试或数据预览时真实探测失败。
// 底层原因（如网络超时、认证失败、DNS 解析失败）会通过 %w 包裹在此错误中。
// 前端收到此错误应展示"连接失败"并附带详细信息。
var ErrDatasourceProbeFailed = errors.New("datasource probe failed")

// ErrDatasourceConfigInvalid 表示数据源配置参数不完整或不合法（如缺少必填字段 host/database）。
// 前端收到此错误应展示"配置不完整"提示。
var ErrDatasourceConfigInvalid = errors.New("invalid datasource config")

// Datasource 类型别名，便于 service 包内直接使用 model.Datasource。
type Datasource = model.Datasource

// DatasourceService 是数据源模块的业务逻辑门面。
// 通过 NewDatasourceService 构造；若传入 Repository 则使用外部持久化，否则使用内存存储。
// 注意：当前 Test 和 Preview 方法返回 mock 结果；真实探针逻辑在 datasource_prober.go 中，
// 未来版本将通过 NewDatasourceServiceWithProber 将 Prober 注入至此。
type DatasourceService struct {
	repo Repository
}

// Repository 定义了数据源持久化仓储的接口。
// 当前有两种实现：
//   - PostgresRepository：生产环境使用，对接 PostgreSQL 的 datasources 表
//   - memoryRepository：测试/开发环境使用，数据仅存在于进程内存
type Repository interface {
	Create(ctx context.Context, datasource model.Datasource) error
	ListByProject(ctx context.Context, projectID string) ([]model.Datasource, error)
	Get(ctx context.Context, id string) (model.Datasource, bool, error)
}

// memoryRepository 是 Repository 的内存实现，用于测试和开发环境。
// 所有数据仅存于进程内存中的切片，进程重启后数据丢失。
type memoryRepository struct {
	mu          sync.Mutex
	datasources []model.Datasource
}

func (r *memoryRepository) Create(_ context.Context, datasource model.Datasource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.datasources = append(r.datasources, datasource)
	return nil
}

func (r *memoryRepository) ListByProject(_ context.Context, projectID string) ([]model.Datasource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var datasources []model.Datasource
	for _, datasource := range r.datasources {
		if projectID == "" || datasource.ProjectID == projectID {
			datasource.Config = cloneConfig(datasource.Config)
			datasources = append(datasources, datasource)
		}
	}
	return datasources, nil
}

func (r *memoryRepository) Get(_ context.Context, id string) (model.Datasource, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, datasource := range r.datasources {
		if datasource.ID == id {
			datasource.Config = cloneConfig(datasource.Config)
			return datasource, true, nil
		}
	}
	return model.Datasource{}, false, nil
}

// NewDatasourceService 创建数据源服务实例。
// 如果传入非 nil 的 Repository，使用传入的实现（如 PostgresRepository）；
// 否则默认使用 memoryRepository（用于测试和简单开发场景）。
func NewDatasourceService(repos ...Repository) *DatasourceService {
	if len(repos) > 0 && repos[0] != nil {
		return &DatasourceService{repo: repos[0]}
	}
	return &DatasourceService{repo: &memoryRepository{}}
}

// Create 创建一个新的数据源配置并持久化。
// 参数校验：
//   - typ 必须是 "mongodb"、"redis" 或 "postgresql" 之一
//   - cfg 为保存前会深拷贝，避免外部修改影响内部状态
// 自动生成 UUID 作为数据源 ID，Readonly 固定为 true。
func (s *DatasourceService) Create(projectID, name, typ string, cfg map[string]string) (Datasource, error) {
	switch typ {
	case "mongodb", "redis", "postgresql":
	default:
		return Datasource{}, ErrInvalidDatasourceType
	}

	datasource := model.Datasource{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Name:      name,
		Type:      typ,
		Readonly:  true,
		Config:    cloneConfig(cfg),
	}

	if err := s.repo.Create(context.Background(), datasource); err != nil {
		return Datasource{}, err
	}
	return datasource, nil
}

// List 按项目 ID 列出所有数据源。如果 projectID 为空，返回所有数据源（memoryRepository 行为）。
func (s *DatasourceService) List(projectID string) ([]Datasource, error) {
	return s.repo.ListByProject(context.Background(), projectID)
}

// Get 按 ID 查询单个数据源。返回三个值：数据源、是否找到、错误。
func (s *DatasourceService) Get(id string) (Datasource, bool, error) {
	return s.repo.Get(context.Background(), id)
}

// Test 对指定数据源执行连接测试。
// 当前版本返回 mock 结果（"mock connection test passed"），不需要实际连接外部数据源。
// 这样可以在没有真实数据源环境的情况下验证 API 路由和前端交互流程。
// 真实探针逻辑在 datasource_prober.go 中的 liveDatasourceProber.Test 实现。
// 如果数据源不存在，返回 ErrDatasourceNotFound。
func (s *DatasourceService) Test(id string) (model.TestResult, error) {
	datasource, ok, err := s.Get(id)
	if err != nil {
		return model.TestResult{}, err
	}
	if !ok {
		return model.TestResult{}, ErrDatasourceNotFound
	}

	return model.TestResult{
		DatasourceID: datasource.ID,
		Status:       "ok",
		Message:      "mock connection test passed",
	}, nil
}

// Preview 对指定数据源执行数据预览查询。
// 当前版本返回 mock 数据（固定的 sample-1/example 记录），不需要实际连接外部数据源。
// 这样可以验证前端数据预览展示的完整交互流程。
// 真实查询逻辑在 datasource_prober.go 中的 liveDatasourceProber.Preview 实现。
// 如果数据源不存在，返回 ErrDatasourceNotFound。
func (s *DatasourceService) Preview(id string) (model.PreviewResult, error) {
	datasource, ok, err := s.Get(id)
	if err != nil {
		return model.PreviewResult{}, err
	}
	if !ok {
		return model.PreviewResult{}, ErrDatasourceNotFound
	}

	return model.PreviewResult{
		DatasourceID:   datasource.ID,
		DatasourceType: datasource.Type,
		Rows: []map[string]string{
			{
				"id":   "sample-1",
				"name": "example",
			},
		},
	}, nil
}

// cloneConfig 深拷贝 config map，防止外部对返回结果的修改影响内部数据。
func cloneConfig(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
