package service

import (
	"context"
	"testing"

	"crawler-platform/apps/project-service/internal/model"
)

type fakeProjectRepo struct {
	projects []model.Project
}

func (r *fakeProjectRepo) Create(_ context.Context, project model.Project) error {
	r.projects = append(r.projects, project)
	return nil
}

func (r *fakeProjectRepo) List(_ context.Context, limit, offset int) ([]model.Project, error) {
	if offset >= len(r.projects) {
		return nil, nil
	}
	end := offset + limit
	if end > len(r.projects) {
		end = len(r.projects)
	}
	result := make([]model.Project, end-offset)
	copy(result, r.projects[offset:end])
	return result, nil
}

func (r *fakeProjectRepo) Count(_ context.Context) (int64, error) {
	return int64(len(r.projects)), nil
}

func (r *fakeProjectRepo) mustGet(id string) model.Project {
	for _, project := range r.projects {
		if project.ID == id {
			return project
		}
	}
	return model.Project{}
}

func TestProjectServiceCreatePersistsThroughRepo(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	project, err := svc.Create("core-crawlers", "Core Crawlers")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if project.ID == "" {
		t.Fatal("expected generated id")
	}

	got := repo.mustGet(project.ID)
	if got.Code != "core-crawlers" || got.Name != "Core Crawlers" {
		t.Fatalf("unexpected persisted project: %#v", got)
	}
}

func TestProjectServiceListReturnsRepoProjects(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	created, err := svc.Create("core-crawlers", "Core Crawlers")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	projects, total, err := svc.List(20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0] != created {
		t.Fatalf("expected list to return created project, got %+v want %+v", projects[0], created)
	}
}

func TestProjectServiceListPagination(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	for i := 0; i < 5; i++ {
		if _, err := svc.Create("code", "name"); err != nil {
			t.Fatal(err)
		}
	}

	projects, total, err := svc.List(2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Fatalf("expected total 5, got %d", total)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

	projects2, _, err := svc.List(2, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects2) != 1 {
		t.Fatalf("expected 1 project at offset 4, got %d", len(projects2))
	}
}
