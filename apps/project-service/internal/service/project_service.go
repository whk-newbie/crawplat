// Package service 是 Project 服务业务逻辑层，负责项目 CRUD。
// 通过 Repository 接口隔离持久化细节，支持 PostgreSQL 和内存两种实现。
// UUID 生成在此层完成，repo 层仅做存取。
package service

import (
	"context"
	"errors"
	"sync"

	"crawler-platform/apps/project-service/internal/model"
	"github.com/google/uuid"
)

// ErrProjectCodeExists 表示项目 code 已被占用。
var ErrProjectCodeExists = errors.New("project code already exists")

// ProjectService 处理项目的创建和查询。
type ProjectService struct {
	repo Repository
}

// Repository 定义项目持久化接口，由 PostgreSQL 或内存实现满足。
type Repository interface {
	Create(ctx context.Context, project model.Project) error
	List(ctx context.Context, limit, offset int) ([]model.Project, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
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

func (r *memoryRepository) List(_ context.Context, limit, offset int) ([]model.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limit <= 0 {
		limit = 20
	}
	if offset >= len(r.projects) {
		return []model.Project{}, nil
	}
	end := offset + limit
	if end > len(r.projects) {
		end = len(r.projects)
	}

	projects := make([]model.Project, end-offset)
	copy(projects, r.projects[offset:end])
	return projects, nil
}

func (r *memoryRepository) ExistsByCode(_ context.Context, code string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.projects {
		if p.Code == code {
			return true, nil
		}
	}
	return false, nil
}

// NewProjectService 创建 ProjectService。接受可选 Repository，为 nil 时回退内存实现。
// 使用可变参数是为了方便调用方省略参数（测试场景），仅取第一个非 nil 值。
func NewProjectService(repos ...Repository) *ProjectService {
	if len(repos) > 0 && repos[0] != nil {
		return &ProjectService{repo: repos[0]}
	}
	return &ProjectService{repo: &memoryRepository{}}
}

// Create 创建项目：先检查 code 唯一性，再生成 UUID 并持久化。
// code 已存在时返回 ErrProjectCodeExists。
func (s *ProjectService) Create(code, name string) (model.Project, error) {
	exists, err := s.repo.ExistsByCode(context.Background(), code)
	if err != nil {
		return model.Project{}, err
	}
	if exists {
		return model.Project{}, ErrProjectCodeExists
	}

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

// List 分页返回项目列表，由 repo 层控制排序。
// limit <= 0 时由 repo 层使用默认值。
func (s *ProjectService) List(limit, offset int) ([]model.Project, error) {
	return s.repo.List(context.Background(), limit, offset)
}
