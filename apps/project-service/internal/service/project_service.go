// Package service 是 Project 服务业务逻辑层，负责项目 CRUD。
// 通过 Repository 接口隔离持久化细节，支持 PostgreSQL 和内存两种实现。
// UUID 生成在此层完成，repo 层仅做存取。
package service

import (
	"context"
	"sync"

	"crawler-platform/apps/project-service/internal/model"
	"github.com/google/uuid"
)

// ProjectService 处理项目的创建和查询。
type ProjectService struct {
	repo Repository
}

// Repository 定义项目持久化接口，由 PostgreSQL 或内存实现满足。
type Repository interface {
	Create(ctx context.Context, project model.Project) error
	List(ctx context.Context) ([]model.Project, error)
}

// memoryRepository 基于内存切片的轻量实现，用于测试和 MVP 阶段。
// 互斥锁保证并发安全，List 返回深拷贝避免外部修改污染内部状态。
type memoryRepository struct {
	mu       sync.Mutex
	projects []model.Project
}

func (r *memoryRepository) Create(_ context.Context, project model.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.projects = append(r.projects, project)
	return nil
}

func (r *memoryRepository) List(_ context.Context) ([]model.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	projects := make([]model.Project, len(r.projects))
	copy(projects, r.projects)
	return projects, nil
}

// NewProjectService 创建 ProjectService。接受可选 Repository，为 nil 时回退内存实现。
// 使用可变参数是为了方便调用方省略参数（测试场景），仅取第一个非 nil 值。
func NewProjectService(repos ...Repository) *ProjectService {
	if len(repos) > 0 && repos[0] != nil {
		return &ProjectService{repo: repos[0]}
	}
	return &ProjectService{repo: &memoryRepository{}}
}

// Create 创建项目：生成 UUID 并持久化。成功返回完整的 Project 对象。
func (s *ProjectService) Create(code, name string) (model.Project, error) {
	project := model.Project{
		ID:   uuid.NewString(),
		Code: code,
		Name: name,
	}

	if err := s.repo.Create(context.Background(), project); err != nil {
		return model.Project{}, err
	}
	return project, nil
}

// List 返回全量项目列表，由 repo 层控制排序。
func (s *ProjectService) List() ([]model.Project, error) {
	return s.repo.List(context.Background())
}
