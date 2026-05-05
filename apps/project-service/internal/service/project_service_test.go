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

func (r *fakeProjectRepo) List(_ context.Context, limit, offset int) ([]model.Project, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset >= len(r.projects) {
		return []model.Project{}, nil
	}
	end := offset + limit
	if end > len(r.projects) {
		end = len(r.projects)
	}
	projects := make([]model.Project, end-offset)
	copy(projects, r.projects[offset:end])
	return projects, nil
}

func (r *fakeProjectRepo) ExistsByCode(_ context.Context, code string) (bool, error) {
	for _, p := range r.projects {
		if p.Code == code {
			return true, nil
		}
	}
	return false, nil
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

	projects, err := svc.List(20, 0)
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

func TestProjectServiceCreateRejectsDuplicateCode(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	if _, err := svc.Create("core-crawlers", "Core Crawlers"); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}

	_, err := svc.Create("core-crawlers", "Another Name")
	if err != ErrProjectCodeExists {
		t.Fatalf("expected ErrProjectCodeExists, got: %v", err)
	}
}

func TestProjectServiceListPagination(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	for i := 0; i < 5; i++ {
		if _, err := svc.Create("code-"+string(rune('a'+i)), "Project"); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	// 限制 3 条
	projects, err := svc.List(3, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects with limit=3, got %d", len(projects))
	}

	// 偏移 2 条，限制 3 条
	projects, err = svc.List(3, 2)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects with limit=3 offset=2, got %d", len(projects))
	}
}
