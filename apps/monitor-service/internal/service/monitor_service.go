package service

import (
	"context"

	"crawler-platform/apps/monitor-service/internal/model"
)

type MonitorService struct {
	repo Repository
}

type Repository interface {
	Overview(ctx context.Context) (model.Overview, error)
}

type memoryRepository struct{}

func (r *memoryRepository) Overview(_ context.Context) (model.Overview, error) {
	return model.Overview{}, nil
}

func NewMonitorService(repos ...Repository) *MonitorService {
	if len(repos) > 0 && repos[0] != nil {
		return &MonitorService{repo: repos[0]}
	}
	return &MonitorService{repo: &memoryRepository{}}
}

func (s *MonitorService) Overview() (model.Overview, error) {
	return s.repo.Overview(context.Background())
}
