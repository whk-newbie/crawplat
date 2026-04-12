package service

import (
	"errors"
	"sync"

	"crawler-platform/apps/datasource-service/internal/model"
	"github.com/google/uuid"
)

var ErrInvalidDatasourceType = errors.New("invalid datasource type")

var ErrDatasourceNotFound = errors.New("datasource not found")

type Datasource = model.Datasource

type DatasourceService struct {
	mu          sync.Mutex
	datasources []model.Datasource
}

func NewDatasourceService() *DatasourceService {
	return &DatasourceService{}
}

func (s *DatasourceService) Create(projectID, name, typ string) (Datasource, error) {
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
	}

	s.mu.Lock()
	s.datasources = append(s.datasources, datasource)
	s.mu.Unlock()

	return datasource, nil
}

func (s *DatasourceService) List(projectID string) []Datasource {
	s.mu.Lock()
	defer s.mu.Unlock()

	datasources := make([]Datasource, 0, len(s.datasources))
	for _, datasource := range s.datasources {
		if projectID != "" && datasource.ProjectID != projectID {
			continue
		}
		datasources = append(datasources, datasource)
	}

	return datasources
}

func (s *DatasourceService) Get(id string) (Datasource, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, datasource := range s.datasources {
		if datasource.ID == id {
			return datasource, true
		}
	}

	return Datasource{}, false
}

func (s *DatasourceService) Test(id string) (model.TestResult, error) {
	datasource, ok := s.Get(id)
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
	datasource, ok := s.Get(id)
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
