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
	ErrSpiderNotFound  = errors.New("spider not found")
)

type SpiderService struct {
	repo Repository
}

type Repository interface {
	Create(ctx context.Context, spider model.Spider) error
	ListByProject(ctx context.Context, orgID, projectID string, limit, offset int) ([]model.Spider, error)
	Get(ctx context.Context, id string) (model.Spider, bool, error)
	CreateVersion(ctx context.Context, version model.SpiderVersion) error
	ListVersions(ctx context.Context, spiderID string) ([]model.SpiderVersion, error)
	ListRegistryAuthRefs(ctx context.Context, projectID string) ([]model.RegistryAuthRef, error)
}

type memoryRepository struct {
	mu       sync.Mutex
	spiders  []model.Spider
	versions []model.SpiderVersion
}

func (r *memoryRepository) Create(_ context.Context, spider model.Spider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.spiders = append(r.spiders, spider)
	return nil
}

func (r *memoryRepository) ListByProject(_ context.Context, orgID, projectID string, limit, offset int) ([]model.Spider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limit <= 0 {
		limit = 20
	}

	var spiders []model.Spider
	for _, spider := range r.spiders {
		if orgID != "" && spider.OrganizationID != orgID {
			continue
		}
		if spider.ProjectID == projectID {
			spider.Command = append([]string(nil), spider.Command...)
			spiders = append(spiders, spider)
		}
	}

	if offset >= len(spiders) {
		return []model.Spider{}, nil
	}
	end := offset + limit
	if end > len(spiders) {
		end = len(spiders)
	}
	return spiders[offset:end], nil
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

func (r *memoryRepository) CreateVersion(_ context.Context, version model.SpiderVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.versions = append(r.versions, version)
	return nil
}

func (r *memoryRepository) ListVersions(_ context.Context, spiderID string) ([]model.SpiderVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var versions []model.SpiderVersion
	for _, v := range r.versions {
		if v.SpiderID == spiderID {
			v.Command = append([]string(nil), v.Command...)
			versions = append(versions, v)
		}
	}
	return versions, nil
}

func (r *memoryRepository) ListRegistryAuthRefs(_ context.Context, _ string) ([]model.RegistryAuthRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return []model.RegistryAuthRef{}, nil
}

func NewSpiderService(repos ...Repository) *SpiderService {
	if len(repos) > 0 && repos[0] != nil {
		return &SpiderService{repo: repos[0]}
	}
	return &SpiderService{repo: &memoryRepository{}}
}

func (s *SpiderService) Create(orgID, projectID, name, language, runtime, image string, command []string) (model.Spider, error) {
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
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		ProjectID:      projectID,
		Name:           name,
		Language:       language,
		Runtime:        runtime,
		Image:          image,
		Command:        append([]string(nil), command...),
	}

	if err := s.repo.Create(context.Background(), spider); err != nil {
		return model.Spider{}, err
	}
	return spider, nil
}

func (s *SpiderService) List(orgID, projectID string, limit, offset int) ([]model.Spider, error) {
	return s.repo.ListByProject(context.Background(), orgID, projectID, limit, offset)
}

func (s *SpiderService) CreateVersion(spiderID, version, image, registryAuthRef string, command []string) (model.SpiderVersion, error) {
	spider, ok, err := s.repo.Get(context.Background(), spiderID)
	if err != nil {
		return model.SpiderVersion{}, err
	}
	if !ok {
		return model.SpiderVersion{}, ErrSpiderNotFound
	}

	v := model.SpiderVersion{
		ID:              uuid.NewString(),
		SpiderID:        spiderID,
		OrganizationID:  spider.OrganizationID,
		Version:         version,
		Image:           image,
		RegistryAuthRef: registryAuthRef,
		Command:         append([]string(nil), command...),
	}
	if err := s.repo.CreateVersion(context.Background(), v); err != nil {
		return model.SpiderVersion{}, err
	}
	return v, nil
}

func (s *SpiderService) ListVersions(spiderID string) ([]model.SpiderVersion, error) {
	return s.repo.ListVersions(context.Background(), spiderID)
}

func (s *SpiderService) ListRegistryAuthRefs(projectID string) ([]model.RegistryAuthRef, error) {
	return s.repo.ListRegistryAuthRefs(context.Background(), projectID)
}
