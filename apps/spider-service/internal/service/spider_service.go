// 该文件为 Spider 业务服务层，负责核心业务逻辑：创建爬虫时的参数校验、ID 生成和持久化；
// 按项目 ID 列表查询爬虫；管理爬虫版本和镜像仓库认证引用。
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
	ErrSpiderNotFound  = errors.New("spider not found")
)

// SpiderService 提供 Spider 的业务操作，依赖 Repository 接口进行持久化。
type SpiderService struct {
	repo Repository
}

// Repository 定义 Spider 持久化操作的抽象接口，由 PostgresRepository 或 memoryRepository 实现。
// 通过接口解耦，使 Service 层不直接依赖具体的数据库实现。
type Repository interface {
	Create(ctx context.Context, spider model.Spider) error
	ListByProject(ctx context.Context, projectID string, limit, offset int) ([]model.Spider, error)
	Get(ctx context.Context, id string) (model.Spider, bool, error)
	CreateVersion(ctx context.Context, version model.SpiderVersion) error
	ListVersions(ctx context.Context, spiderID string) ([]model.SpiderVersion, error)
	ListRegistryAuthRefs(ctx context.Context, projectID string) ([]model.RegistryAuthRef, error)
}

// memoryRepository 是 Repository 接口的线程安全内存实现，用于测试和零配置启动场景。
type memoryRepository struct {
	mu       sync.Mutex
	spiders  []model.Spider
	versions []model.SpiderVersion
}

func (r *memoryRepository) Create(_ context.Context, spider model.Spider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.spiders = append(r.spiders, spider)
	return nil
}

func (r *memoryRepository) ListByProject(_ context.Context, projectID string, limit, offset int) ([]model.Spider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limit <= 0 {
		limit = 20
	}

	var spiders []model.Spider
	for _, spider := range r.spiders {
		if spider.ProjectID == projectID {
			spider.Command = append([]string(nil), spider.Command...)
			spiders = append(spiders, spider)
		}
	}

	if offset >= len(spiders) {
		return []model.Spider{}, nil
	}
	end := offset + limit
	if end > len(spiders) {
		end = len(spiders)
	}
	return spiders[offset:end], nil
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

func (r *memoryRepository) CreateVersion(_ context.Context, version model.SpiderVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.versions = append(r.versions, version)
	return nil
}

func (r *memoryRepository) ListVersions(_ context.Context, spiderID string) ([]model.SpiderVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var versions []model.SpiderVersion
	for _, v := range r.versions {
		if v.SpiderID == spiderID {
			v.Command = append([]string(nil), v.Command...)
			versions = append(versions, v)
		}
	}
	return versions, nil
}

func (r *memoryRepository) ListRegistryAuthRefs(_ context.Context, _ string) ([]model.RegistryAuthRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return []model.RegistryAuthRef{}, nil
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

// List 分页返回指定项目下爬虫列表。limit <= 0 时由 repo 层使用默认值。
func (s *SpiderService) List(projectID string, limit, offset int) ([]model.Spider, error) {
	return s.repo.ListByProject(context.Background(), projectID, limit, offset)
}

// CreateVersion 为指定 Spider 创建一个新版本。
// 先校验 Spider 是否存在，再生成版本记录。
func (s *SpiderService) CreateVersion(spiderID, version, image, registryAuthRef string, command []string) (model.SpiderVersion, error) {
	_, ok, err := s.repo.Get(context.Background(), spiderID)
	if err != nil {
		return model.SpiderVersion{}, err
	}
	if !ok {
		return model.SpiderVersion{}, ErrSpiderNotFound
	}

	v := model.SpiderVersion{
		ID:              uuid.NewString(),
		SpiderID:        spiderID,
		Version:         version,
		Image:           image,
		RegistryAuthRef: registryAuthRef,
		Command:         append([]string(nil), command...),
	}
	if err := s.repo.CreateVersion(context.Background(), v); err != nil {
		return model.SpiderVersion{}, err
	}
	return v, nil
}

// ListVersions 返回指定 Spider 的所有版本。
func (s *SpiderService) ListVersions(spiderID string) ([]model.SpiderVersion, error) {
	return s.repo.ListVersions(context.Background(), spiderID)
}

// ListRegistryAuthRefs 返回项目关联的镜像仓库认证引用列表。
func (s *SpiderService) ListRegistryAuthRefs(projectID string) ([]model.RegistryAuthRef, error) {
	return s.repo.ListRegistryAuthRefs(context.Background(), projectID)
}
