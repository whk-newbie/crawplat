package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"crawler-platform/apps/spider-service/internal/model"
	"github.com/google/uuid"
)

var (
	ErrInvalidLanguage = errors.New("invalid language")
	ErrInvalidRuntime  = errors.New("invalid runtime")
	ErrImageRequired   = errors.New("image is required for docker runtime")
	ErrSpiderNotFound  = errors.New("spider not found")
)

type SpiderService struct {
	repo Repository
}

type Repository interface {
	Create(ctx context.Context, spider model.Spider) error
	ListByProject(ctx context.Context, projectID string, limit, offset int) ([]model.Spider, error)
	CountByProject(ctx context.Context, projectID string) (int64, error)
	Get(ctx context.Context, id string) (model.Spider, bool, error)
	CreateVersion(ctx context.Context, spiderID, image string, command []string) (model.SpiderVersion, error)
	ListVersions(ctx context.Context, spiderID string) ([]model.SpiderVersion, error)
}

type memoryRepository struct {
	mu       sync.Mutex
	spiders  []model.Spider
	versions map[string][]model.SpiderVersion
}

func (r *memoryRepository) Create(_ context.Context, spider model.Spider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.spiders = append(r.spiders, spider)
	if r.versions == nil {
		r.versions = map[string][]model.SpiderVersion{}
	}
	r.versions[spider.ID] = []model.SpiderVersion{{
		ID:        uuid.NewString(),
		SpiderID:  spider.ID,
		Version:   1,
		Image:     spider.Image,
		Command:   append([]string(nil), spider.Command...),
		CreatedAt: time.Now().UTC(),
	}}
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

func (r *memoryRepository) CreateVersion(_ context.Context, spiderID, image string, command []string) (model.SpiderVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.spiders {
		if r.spiders[i].ID != spiderID {
			continue
		}
		existing := r.versions[spiderID]
		nextVersion := 1
		if len(existing) > 0 {
			nextVersion = existing[0].Version + 1
		}
		created := model.SpiderVersion{
			ID:        uuid.NewString(),
			SpiderID:  spiderID,
			Version:   nextVersion,
			Image:     image,
			Command:   append([]string(nil), command...),
			CreatedAt: time.Now().UTC(),
		}
		r.versions[spiderID] = append([]model.SpiderVersion{created}, existing...)
		r.spiders[i].Image = image
		r.spiders[i].Command = append([]string(nil), command...)
		return created, nil
	}

	return model.SpiderVersion{}, ErrSpiderNotFound
}

func (r *memoryRepository) ListVersions(_ context.Context, spiderID string) ([]model.SpiderVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions := r.versions[spiderID]
	out := make([]model.SpiderVersion, 0, len(versions))
	for _, version := range versions {
		version.Command = append([]string(nil), version.Command...)
		out = append(out, version)
	}
	return out, nil
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

func (s *SpiderService) CreateVersion(spiderID, image string, command []string) (model.SpiderVersion, error) {
	spider, found, err := s.repo.Get(context.Background(), spiderID)
	if err != nil {
		return model.SpiderVersion{}, err
	}
	if !found {
		return model.SpiderVersion{}, ErrSpiderNotFound
	}

	trimmedImage := strings.TrimSpace(image)
	if spider.Runtime == "docker" && trimmedImage == "" {
		return model.SpiderVersion{}, ErrImageRequired
	}
	if trimmedImage == "" {
		trimmedImage = spider.Image
	}
	if len(command) == 0 {
		command = spider.Command
	}

	return s.repo.CreateVersion(context.Background(), spiderID, trimmedImage, append([]string(nil), command...))
}

func (s *SpiderService) ListVersions(spiderID string) ([]model.SpiderVersion, error) {
	_, found, err := s.repo.Get(context.Background(), spiderID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrSpiderNotFound
	}
	return s.repo.ListVersions(context.Background(), spiderID)
}
