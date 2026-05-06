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

func (r *fakeDatasourceRepo) ListByProject(_ context.Context, orgID, projectID string, limit, offset int) ([]model.Datasource, error) {
	var datasources []model.Datasource
	for _, datasource := range r.datasources {
		if orgID != "" && datasource.OrganizationID != orgID {
			continue
		}
		if projectID == "" || datasource.ProjectID == projectID {
			datasources = append(datasources, datasource)
		}
	}
	if offset >= len(datasources) {
		return []model.Datasource{}, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(datasources) {
		end = len(datasources)
	}
	return datasources[offset:end], nil
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
	_, err := svc.Create("", "project-1", "main", "mysql", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateDatasourcePersistsThroughRepo(t *testing.T) {
	repo := &fakeDatasourceRepo{}
	svc := NewDatasourceService(repo)

	datasource, err := svc.Create("org-1", "project-1", "main", "postgresql", map[string]string{"schema": "public"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if datasource.ID == "" {
		t.Fatal("expected generated id")
	}
	if datasource.OrganizationID != "org-1" {
		t.Fatalf("expected OrganizationID org-1, got %s", datasource.OrganizationID)
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

	created, err := svc.Create("org-1", "project-1", "main", "redis", map[string]string{"db": "0"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	listed, err := svc.List("org-1", "project-1", 20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
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

func TestDatasourceServiceListFiltersByOrg(t *testing.T) {
	repo := &fakeDatasourceRepo{}
	svc := NewDatasourceService(repo)

	if _, err := svc.Create("org-a", "p1", "ds-org-a", "redis", nil); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create("org-b", "p1", "ds-org-b", "redis", nil); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	listed, err := svc.List("org-a", "p1", 20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 datasource for org-a, got %d", len(listed))
	}
	if listed[0].Name != "ds-org-a" {
		t.Fatalf("expected ds-org-a, got %s", listed[0].Name)
	}
}

func TestDatasourceServiceCreateAllowsSameNameInDifferentOrgs(t *testing.T) {
	repo := &fakeDatasourceRepo{}
	svc := NewDatasourceService(repo)

	ds1, err := svc.Create("org-a", "p1", "main", "redis", nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	ds2, err := svc.Create("org-b", "p1", "main", "redis", nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if ds1.ID == ds2.ID {
		t.Fatal("expected different IDs for different orgs")
	}

	listedA, _ := svc.List("org-a", "p1", 20, 0)
	listedB, _ := svc.List("org-b", "p1", 20, 0)
	if len(listedA) != 1 || listedA[0].OrganizationID != "org-a" {
		t.Fatalf("org-a should see 1 datasource, got %d", len(listedA))
	}
	if len(listedB) != 1 || listedB[0].OrganizationID != "org-b" {
		t.Fatalf("org-b should see 1 datasource, got %d", len(listedB))
	}
}
