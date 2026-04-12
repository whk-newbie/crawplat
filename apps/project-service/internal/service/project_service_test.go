package service

import "testing"

func TestCreateProjectAssignsID(t *testing.T) {
	svc := NewProjectService()
	project := svc.Create("core-crawlers", "Core Crawlers")
	if project.ID == "" {
		t.Fatal("expected generated id")
	}
}

func TestListReturnsCreatedProjects(t *testing.T) {
	svc := NewProjectService()

	created := svc.Create("core-crawlers", "Core Crawlers")
	projects := svc.List()

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0] != created {
		t.Fatalf("expected list to return created project, got %+v want %+v", projects[0], created)
	}
}
