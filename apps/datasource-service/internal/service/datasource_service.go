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
var ErrDatasourceConfigInvalid = errors.New("invalid datasource config")
var ErrDatasourceProbeFailed = errors.New("datasource probe failed")

type Datasource = model.Datasource

type DatasourceService struct {
	repo   Repository
	prober Prober
}

type Repository interface {
	Create(ctx context.Context, datasource model.Datasource) error
	ListByProject(ctx context.Context, projectID string, limit, offset int) ([]model.Datasource, error)
	CountByProject(ctx context.Context, projectID string) (int64, error)
	Get(ctx context.Context, id string) (model.Datasource, bool, error)
}

type Prober interface {
	Test(ctx context.Context, datasource model.Datasource) (model.TestResult, error)
	Preview(ctx context.Context, datasource model.Datasource) (model.PreviewResult, error)
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

func (r *memoryRepository) ListByProject(_ context.Context, projectID string, limit, offset int) ([]model.Datasource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var all []model.Datasource
	for _, datasource := range r.datasources {
		if projectID == "" || datasource.ProjectID == projectID {
			datasource.Config = cloneConfig(datasource.Config)
			all = append(all, datasource)
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
	for _, datasource := range r.datasources {
		if projectID == "" || datasource.ProjectID == projectID {
			count++
		}
	}
	return count, nil
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
	svc := &DatasourceService{prober: newLiveDatasourceProber()}
	if len(repos) > 0 && repos[0] != nil {
		svc.repo = repos[0]
		return svc
	}
	svc.repo = &memoryRepository{}
	return svc
}

func (s *DatasourceService) WithProber(prober Prober) *DatasourceService {
	if prober != nil {
		s.prober = prober
	}
	return s
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

func (s *DatasourceService) List(projectID string, limit, offset int) ([]Datasource, int64, error) {
	datasources, err := s.repo.ListByProject(context.Background(), projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByProject(context.Background(), projectID)
	if err != nil {
		return nil, 0, err
	}
	return datasources, total, nil
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
	result, err := s.prober.Test(context.Background(), datasource)
	if err != nil {
		return model.TestResult{}, err
	}
	if result.DatasourceID == "" {
		result.DatasourceID = datasource.ID
	}
	return result, nil
}

func (s *DatasourceService) Preview(id string) (model.PreviewResult, error) {
	datasource, ok, err := s.Get(id)
	if err != nil {
		return model.PreviewResult{}, err
	}
	if !ok {
		return model.PreviewResult{}, ErrDatasourceNotFound
	}
	result, err := s.prober.Preview(context.Background(), datasource)
	if err != nil {
		return model.PreviewResult{}, err
	}
	if result.DatasourceID == "" {
		result.DatasourceID = datasource.ID
	}
	if result.DatasourceType == "" {
		result.DatasourceType = datasource.Type
	}
	return result, nil
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
