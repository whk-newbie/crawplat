// Package service 的单元测试，使用 fake repository 验证 Create/List 行为。
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

func (r *fakeProjectRepo) List(_ context.Context) ([]model.Project, error) {
	projects := make([]model.Project, len(r.projects))
	copy(projects, r.projects)
	return projects, nil
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

	projects, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0] != created {
		t.Fatalf("expected list to return created project, got %+v want %+v", projects[0], created)
	}
}
