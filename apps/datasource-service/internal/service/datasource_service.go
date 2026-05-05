package service

import (
	"context"
	"errors"
	"sync"

	"crawler-platform/apps/datasource-service/internal/model"
	"github.com/google/uuid"
)

var ErrInvalidDatasourceType = errors.New("invalid datasource type")

var ErrDatasourceNotFound = errors.New("datasource not found")

var ErrDatasourceProbeFailed = errors.New("datasource probe failed")

var ErrDatasourceConfigInvalid = errors.New("invalid datasource config")

type Datasource = model.Datasource

type DatasourceService struct {
	repo Repository
}

type Repository interface {
	Create(ctx context.Context, datasource model.Datasource) error
	ListByProject(ctx context.Context, projectID string) ([]model.Datasource, error)
	Get(ctx context.Context, id string) (model.Datasource, bool, error)
}

type memoryRepository struct {
	mu          sync.Mutex
	datasources []model.Datasource
}

func (r *memoryRepository) Create(_ context.Context, datasource model.Datasource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.datasources = append(r.datasources, datasource)
	return nil
}

func (r *memoryRepository) ListByProject(_ context.Context, projectID string) ([]model.Datasource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var datasources []model.Datasource
	for _, datasource := range r.datasources {
		if projectID == "" || datasource.ProjectID == projectID {
			datasource.Config = cloneConfig(datasource.Config)
			datasources = append(datasources, datasource)
		}
	}
	return datasources, nil
}

func (r *memoryRepository) Get(_ context.Context, id string) (model.Datasource, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, datasource := range r.datasources {
		if datasource.ID == id {
			datasource.Config = cloneConfig(datasource.Config)
			return datasource, true, nil
		}
	}
	return model.Datasource{}, false, nil
}

func NewDatasourceService(repos ...Repository) *DatasourceService {
	if len(repos) > 0 && repos[0] != nil {
		return &DatasourceService{repo: repos[0]}
	}
	return &DatasourceService{repo: &memoryRepository{}}
}

func (s *DatasourceService) Create(projectID, name, typ string, cfg map[string]string) (Datasource, error) {
	switch typ {
	case "mongodb", "redis", "postgresql":
	default:
		return Datasource{}, ErrInvalidDatasourceType
	}

	datasource := model.Datasource{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Name:      name,
		Type:      typ,
		Readonly:  true,
		Config:    cloneConfig(cfg),
	}

	if err := s.repo.Create(context.Background(), datasource); err != nil {
		return Datasource{}, err
	}
	return datasource, nil
}

func (s *DatasourceService) List(projectID string) ([]Datasource, error) {
	return s.repo.ListByProject(context.Background(), projectID)
}

func (s *DatasourceService) Get(id string) (Datasource, bool, error) {
	return s.repo.Get(context.Background(), id)
}

func (s *DatasourceService) Test(id string) (model.TestResult, error) {
	datasource, ok, err := s.Get(id)
	if err != nil {
		return model.TestResult{}, err
	}
	if !ok {
		return model.TestResult{}, ErrDatasourceNotFound
	}

	return model.TestResult{
		DatasourceID: datasource.ID,
		Status:       "ok",
		Message:      "mock connection test passed",
	}, nil
}

func (s *DatasourceService) Preview(id string) (model.PreviewResult, error) {
	datasource, ok, err := s.Get(id)
	if err != nil {
		return model.PreviewResult{}, err
	}
	if !ok {
		return model.PreviewResult{}, ErrDatasourceNotFound
	}

	return model.PreviewResult{
		DatasourceID:   datasource.ID,
		DatasourceType: datasource.Type,
		Rows: []map[string]string{
			{
				"id":   "sample-1",
				"name": "example",
			},
		},
	}, nil
}

func cloneConfig(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
