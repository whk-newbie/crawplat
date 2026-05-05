// 该文件为 Spider 业务服务层，负责核心业务逻辑：创建爬虫时的参数校验、ID 生成和持久化；
// 按项目 ID 列表查询爬虫。
//
// 通过 Repository 接口与持久化层解耦：生产环境注入 PostgresRepository，测试或零配置启动时
// 回退到内置的 memoryRepository（线程安全的内存存储）。
//
// 不负责 HTTP 协议处理（由 api 层负责）和数据库 SQL 编写（由 repo 层负责）。
package service

import (
	"context"
	"errors"
	"sync"

	"crawler-platform/apps/spider-service/internal/model"
	"github.com/google/uuid"
)

// 业务校验哨兵错误，用于在 api 层映射为相应的 HTTP 状态码。
var (
	ErrInvalidLanguage = errors.New("invalid language")
	ErrInvalidRuntime  = errors.New("invalid runtime")
	ErrImageRequired   = errors.New("image is required for docker runtime")
)

// SpiderService 提供 Spider 的业务操作，依赖 Repository 接口进行持久化。
type SpiderService struct {
	repo Repository
}

// Repository 定义 Spider 持久化操作的抽象接口，由 PostgresRepository 或 memoryRepository 实现。
// 通过接口解耦，使 Service 层不直接依赖具体的数据库实现。
type Repository interface {
	Create(ctx context.Context, spider model.Spider) error
	ListByProject(ctx context.Context, projectID string) ([]model.Spider, error)
	Get(ctx context.Context, id string) (model.Spider, bool, error)
}

// memoryRepository 是 Repository 接口的线程安全内存实现，用于测试和零配置启动场景。
// 使用 sync.Mutex 保护并发访问；ListByProject 和 Get 返回数据时会深拷贝 Command 切片，
// 防止调用方意外修改内部状态。
type memoryRepository struct {
	mu      sync.Mutex
	spiders []model.Spider
}

func (r *memoryRepository) Create(_ context.Context, spider model.Spider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.spiders = append(r.spiders, spider)
	return nil
}

func (r *memoryRepository) ListByProject(_ context.Context, projectID string) ([]model.Spider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var spiders []model.Spider
	for _, spider := range r.spiders {
		if spider.ProjectID == projectID {
			spider.Command = append([]string(nil), spider.Command...)
			spiders = append(spiders, spider)
		}
	}
	return spiders, nil
}

func (r *memoryRepository) Get(_ context.Context, id string) (model.Spider, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, spider := range r.spiders {
		if spider.ID == id {
			spider.Command = append([]string(nil), spider.Command...)
			return spider, true, nil
		}
	}
	return model.Spider{}, false, nil
}

// NewSpiderService 创建 SpiderService 实例。接受可选的 Repository 实现：
//   - 若传入非 nil Repository，使用该实现（生产环境传入 PostgresRepository）
//   - 若未传入或传入 nil，使用内置的 memoryRepository（适合测试和零配置开发）
func NewSpiderService(repos ...Repository) *SpiderService {
	if len(repos) > 0 && repos[0] != nil {
		return &SpiderService{repo: repos[0]}
	}
	return &SpiderService{repo: &memoryRepository{}}
}

// Create 创建一条新的爬虫配置记录。
//
// 校验规则：
//   - language 仅允许 "go" 或 "python"
//   - runtime 仅允许 "docker" 或 "host"
//   - docker 运行时必须提供 image（host 运行时 image 可为空）
//
// 通过后生成 UUID 作为 Spider 的唯一标识，Command 切片会被深拷贝后存储。
// 返回创建的 Spider 对象（含生成的 ID）或相应的哨兵错误。
func (s *SpiderService) Create(projectID, name, language, runtime, image string, command []string) (model.Spider, error) {
	if language != "go" && language != "python" {
		return model.Spider{}, ErrInvalidLanguage
	}
	if runtime != "docker" && runtime != "host" {
		return model.Spider{}, ErrInvalidRuntime
	}
	if runtime == "docker" && image == "" {
		return model.Spider{}, ErrImageRequired
	}

	spider := model.Spider{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Name:      name,
		Language:  language,
		Runtime:   runtime,
		Image:     image,
		Command:   append([]string(nil), command...),
	}

	if err := s.repo.Create(context.Background(), spider); err != nil {
		return model.Spider{}, err
	}
	return spider, nil
}

// List 返回指定项目下所有爬虫的列表。委托 Repository.ListByProject 执行查询，
// 内部使用 context.Background()，适合无超时要求的场景。
func (s *SpiderService) List(projectID string) ([]model.Spider, error) {
	return s.repo.ListByProject(context.Background(), projectID)
}
