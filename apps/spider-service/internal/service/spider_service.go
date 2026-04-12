package service

import (
	"errors"
	"sync"

	"crawler-platform/apps/spider-service/internal/model"
	"github.com/google/uuid"
)

var (
	ErrInvalidLanguage = errors.New("invalid language")
	ErrInvalidRuntime  = errors.New("invalid runtime")
)

type SpiderService struct {
	mu      sync.Mutex
	spiders []model.Spider
}

func NewSpiderService() *SpiderService {
	return &SpiderService{}
}

func (s *SpiderService) Create(projectID, name, language, runtime string) (model.Spider, error) {
	if language != "go" && language != "python" {
		return model.Spider{}, ErrInvalidLanguage
	}
	if runtime != "docker" && runtime != "host" {
		return model.Spider{}, ErrInvalidRuntime
	}

	spider := model.Spider{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Name:      name,
		Language:  language,
		Runtime:   runtime,
	}

	s.mu.Lock()
	s.spiders = append(s.spiders, spider)
	s.mu.Unlock()

	return spider, nil
}

func (s *SpiderService) List(projectID string) []model.Spider {
	s.mu.Lock()
	defer s.mu.Unlock()

	spiders := make([]model.Spider, 0, len(s.spiders))
	for _, spider := range s.spiders {
		if spider.ProjectID == projectID {
			spiders = append(spiders, spider)
		}
	}
	return spiders
}
