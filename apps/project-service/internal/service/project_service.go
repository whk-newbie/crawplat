package service

import (
	"sync"

	"crawler-platform/apps/project-service/internal/model"
	"github.com/google/uuid"
)

type ProjectService struct {
	mu       sync.Mutex
	projects []model.Project
}

func NewProjectService() *ProjectService {
	return &ProjectService{}
}

func (s *ProjectService) Create(code, name string) model.Project {
	project := model.Project{
		ID:   uuid.NewString(),
		Code: code,
		Name: name,
	}

	s.mu.Lock()
	s.projects = append(s.projects, project)
	s.mu.Unlock()

	return project
}

func (s *ProjectService) List() []model.Project {
	s.mu.Lock()
	defer s.mu.Unlock()

	projects := make([]model.Project, len(s.projects))
	copy(projects, s.projects)
	return projects
}
