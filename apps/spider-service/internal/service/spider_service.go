package service

import (
	"context"
	"errors"
	"sync"

	"crawler-platform/apps/spider-service/internal/model"
	"github.com/google/uuid"
)

var (
	ErrInvalidLanguage = errors.New("invalid language")
	ErrInvalidRuntime  = errors.New("invalid runtime")
	ErrImageRequired   = errors.New("image is required for docker runtime")
)

type SpiderService struct {
	repo Repository
}

type Repository interface {
	Create(ctx context.Context, spider model.Spider) error
	ListByProject(ctx context.Context, projectID string, limit, offset int) ([]model.Spider, error)
	CountByProject(ctx context.Context, projectID string) (int64, error)
	Get(ctx context.Context, id string) (model.Spider, bool, error)
}

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

func (r *memoryRepository) ListByProject(_ context.Context, projectID string, limit, offset int) ([]model.Spider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var all []model.Spider
	for _, spider := range r.spiders {
		if spider.ProjectID == projectID {
			spider.Command = append([]string(nil), spider.Command...)
			all = append(all, spider)
		}
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *memoryRepository) CountByProject(_ context.Context, projectID string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for _, spider := range r.spiders {
		if spider.ProjectID == projectID {
			count++
		}
	}
	return count, nil
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

func NewSpiderService(repos ...Repository) *SpiderService {
	if len(repos) > 0 && repos[0] != nil {
		return &SpiderService{repo: repos[0]}
	}
	return &SpiderService{repo: &memoryRepository{}}
}

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

func (s *SpiderService) List(projectID string, limit, offset int) ([]model.Spider, int64, error) {
	spiders, err := s.repo.ListByProject(context.Background(), projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByProject(context.Background(), projectID)
	if err != nil {
		return nil, 0, err
	}
	return spiders, total, nil
}
