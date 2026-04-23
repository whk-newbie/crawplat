package service

import (
	"context"
	"errors"
	"sync"

	"crawler-platform/apps/scheduler-service/internal/model"
	"github.com/google/uuid"
)

var ErrInvalidSchedule = errors.New("invalid schedule")

type SchedulerService struct {
	repo Repository
}

type Repository interface {
	Create(ctx context.Context, schedule model.Schedule) error
	List(ctx context.Context) ([]model.Schedule, error)
}

type memoryRepository struct {
	mu        sync.Mutex
	schedules []model.Schedule
}

func (r *memoryRepository) Create(_ context.Context, schedule model.Schedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schedules = append(r.schedules, schedule)
	return nil
}

func (r *memoryRepository) List(_ context.Context) ([]model.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	schedules := make([]model.Schedule, len(r.schedules))
	copy(schedules, r.schedules)
	return schedules, nil
}

func NewSchedulerService(repos ...Repository) *SchedulerService {
	if len(repos) > 0 && repos[0] != nil {
		return &SchedulerService{repo: repos[0]}
	}
	return &SchedulerService{repo: &memoryRepository{}}
}

func (s *SchedulerService) Create(projectID, spiderID, name, cronExpr, image string, command []string, enabled bool) (model.Schedule, error) {
	if projectID == "" || spiderID == "" || name == "" || cronExpr == "" || image == "" {
		return model.Schedule{}, ErrInvalidSchedule
	}

	schedule := model.Schedule{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		SpiderID:  spiderID,
		Name:      name,
		CronExpr:  cronExpr,
		Enabled:   enabled,
		Image:     image,
		Command:   append([]string(nil), command...),
	}

	if err := s.repo.Create(context.Background(), schedule); err != nil {
		return model.Schedule{}, err
	}
	return schedule, nil
}

func (s *SchedulerService) List() ([]model.Schedule, error) {
	return s.repo.List(context.Background())
}
