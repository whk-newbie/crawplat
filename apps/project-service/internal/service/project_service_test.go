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

func (r *fakeProjectRepo) List(_ context.Context, orgID string, limit, offset int) ([]model.Project, error) {
	if limit <= 0 {
		limit = 20
	}
	var filtered []model.Project
	for _, p := range r.projects {
		if orgID == "" || p.OrganizationID == orgID {
			filtered = append(filtered, p)
		}
	}
	if offset >= len(filtered) {
		return []model.Project{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	projects := make([]model.Project, end-offset)
	copy(projects, filtered[offset:end])
	return projects, nil
}

func (r *fakeProjectRepo) ExistsByCode(_ context.Context, orgID, code string) (bool, error) {
	for _, p := range r.projects {
		if (orgID == "" || p.OrganizationID == orgID) && p.Code == code {
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

	project, err := svc.Create("org-1", "core-crawlers", "Core Crawlers")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if project.ID == "" {
		t.Fatal("expected generated id")
	}
	if project.OrganizationID != "org-1" {
		t.Fatalf("expected orgID org-1, got %s", project.OrganizationID)
	}

	got := repo.mustGet(project.ID)
	if got.Code != "core-crawlers" || got.Name != "Core Crawlers" {
		t.Fatalf("unexpected persisted project: %#v", got)
	}
}

func TestProjectServiceListReturnsRepoProjects(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	created, err := svc.Create("org-1", "core-crawlers", "Core Crawlers")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	projects, err := svc.List("org-1", 20, 0)
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

func TestProjectServiceListFiltersByOrg(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	svc.Create("org-1", "code-1", "Project 1")
	svc.Create("org-2", "code-2", "Project 2")

	projects, err := svc.List("org-1", 20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project for org-1, got %d", len(projects))
	}
	if projects[0].Code != "code-1" {
		t.Fatalf("expected code-1, got %s", projects[0].Code)
	}
}

func TestProjectServiceCreateRejectsDuplicateCode(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	if _, err := svc.Create("org-1", "core-crawlers", "Core Crawlers"); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}

	_, err := svc.Create("org-1", "core-crawlers", "Another Name")
	if err != ErrProjectCodeExists {
		t.Fatalf("expected ErrProjectCodeExists, got: %v", err)
	}
}

func TestProjectServiceCreateAllowsSameCodeInDifferentOrgs(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	if _, err := svc.Create("org-1", "core-crawlers", "Core Crawlers"); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}

	// Same code in different org should succeed
	project, err := svc.Create("org-2", "core-crawlers", "Another Org Crawlers")
	if err != nil {
		t.Fatalf("second Create with same code in different org returned error: %v", err)
	}
	if project.OrganizationID != "org-2" {
		t.Fatalf("expected orgID org-2, got %s", project.OrganizationID)
	}
}

func TestProjectServiceListPagination(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := NewProjectService(repo)

	for i := 0; i < 5; i++ {
		if _, err := svc.Create("org-1", "code-"+string(rune('a'+i)), "Project"); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	projects, err := svc.List("org-1", 3, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects with limit=3, got %d", len(projects))
	}

	projects, err = svc.List("org-1", 3, 2)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects with limit=3 offset=2, got %d", len(projects))
	}
}
