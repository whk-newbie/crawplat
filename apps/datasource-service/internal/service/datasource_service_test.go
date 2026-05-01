package service

import (
	"context"
	"testing"

	"crawler-platform/apps/datasource-service/internal/model"
)

type fakeDatasourceRepo struct {
	datasources []model.Datasource
}

func (r *fakeDatasourceRepo) Create(_ context.Context, datasource model.Datasource) error {
	r.datasources = append(r.datasources, datasource)
	return nil
}

func (r *fakeDatasourceRepo) ListByProject(_ context.Context, projectID string, limit, offset int) ([]model.Datasource, error) {
	var all []model.Datasource
	for _, datasource := range r.datasources {
		if projectID == "" || datasource.ProjectID == projectID {
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

func (r *fakeDatasourceRepo) CountByProject(_ context.Context, projectID string) (int64, error) {
	var count int64
	for _, datasource := range r.datasources {
		if projectID == "" || datasource.ProjectID == projectID {
			count++
		}
	}
	return count, nil
}

func (r *fakeDatasourceRepo) Get(_ context.Context, id string) (model.Datasource, bool, error) {
	for _, datasource := range r.datasources {
		if datasource.ID == id {
			return datasource, true, nil
		}
	}
	return model.Datasource{}, false, nil
}

func TestCreateDatasourceRejectsUnknownType(t *testing.T) {
	svc := NewDatasourceService(&fakeDatasourceRepo{})
	_, err := svc.Create("project-1", "main", "mysql", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateDatasourcePersistsThroughRepo(t *testing.T) {
	repo := &fakeDatasourceRepo{}
	svc := NewDatasourceService(repo)

	datasource, err := svc.Create("project-1", "main", "postgresql", map[string]string{"schema": "public"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if datasource.ID == "" {
		t.Fatal("expected generated id")
	}
	if len(repo.datasources) != 1 {
		t.Fatalf("expected repo to persist one datasource, got %d", len(repo.datasources))
	}
	if repo.datasources[0].Config["schema"] != "public" {
		t.Fatalf("expected config to persist, got %#v", repo.datasources[0].Config)
	}
}

func TestDatasourceServiceListAndReadUseRepo(t *testing.T) {
	repo := &fakeDatasourceRepo{}
	svc := NewDatasourceService(repo)

	created, err := svc.Create("project-1", "main", "redis", map[string]string{"db": "0"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	listed, total, err := svc.List("project-1", 20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("unexpected total: %d", total)
	}
	if len(listed) != 1 {
		t.Fatalf("unexpected list length: %#v", listed)
	}
	if listed[0].ID != created.ID || listed[0].Config["db"] != "0" {
		t.Fatalf("unexpected list result: %#v", listed[0])
	}

	result, err := svc.Test(created.ID)
	if err != nil {
		t.Fatalf("Test returned error: %v", err)
	}
	if result.DatasourceID != created.ID {
		t.Fatalf("unexpected test result: %#v", result)
	}

	preview, err := svc.Preview(created.ID)
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	if preview.DatasourceType != "redis" {
		t.Fatalf("unexpected preview result: %#v", preview)
	}
}
