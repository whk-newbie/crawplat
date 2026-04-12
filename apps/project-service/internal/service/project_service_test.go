package service

import "testing"

func TestCreateProjectAssignsID(t *testing.T) {
	svc := NewProjectService()
	project := svc.Create("core-crawlers", "Core Crawlers")
	if project.ID == "" {
		t.Fatal("expected generated id")
	}
}
