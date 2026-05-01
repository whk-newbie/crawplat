package service

import (
	"context"
	"sync"

	"crawler-platform/apps/project-service/internal/model"
	"github.com/google/uuid"
)

type ProjectService struct {
	repo Repository
}

type Repository interface {
	Create(ctx context.Context, project model.Project) error
	List(ctx context.Context, limit, offset int) ([]model.Project, error)
	Count(ctx context.Context) (int64, error)
}

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

	if offset >= len(r.projects) {
		return nil, nil
	}
	end := offset + limit
	if end > len(r.projects) {
		end = len(r.projects)
	}
	result := make([]model.Project, end-offset)
	copy(result, r.projects[offset:end])
	return result, nil
}

func (r *memoryRepository) Count(_ context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.projects)), nil
}

func NewProjectService(repos ...Repository) *ProjectService {
	if len(repos) > 0 && repos[0] != nil {
		return &ProjectService{repo: repos[0]}
	}
	return &ProjectService{repo: &memoryRepository{}}
}

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

func (s *ProjectService) List(limit, offset int) ([]model.Project, int64, error) {
	projects, err := s.repo.List(context.Background(), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(context.Background())
	if err != nil {
		return nil, 0, err
	}
	return projects, total, err
}
