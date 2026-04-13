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
	ListByProject(ctx context.Context, projectID string) ([]model.Spider, error)
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

func (s *SpiderService) List(projectID string) ([]model.Spider, error) {
	return s.repo.ListByProject(context.Background(), projectID)
}
