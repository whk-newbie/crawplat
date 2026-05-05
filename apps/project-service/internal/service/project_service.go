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
	List(ctx context.Context) ([]model.Project, error)
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

func (r *memoryRepository) List(_ context.Context) ([]model.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	projects := make([]model.Project, len(r.projects))
	copy(projects, r.projects)
	return projects, nil
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

func (s *ProjectService) List() ([]model.Project, error) {
	return s.repo.List(context.Background())
}
