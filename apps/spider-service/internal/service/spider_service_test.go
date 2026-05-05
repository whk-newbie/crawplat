package service

import (
	"context"
	"errors"
	"testing"

	"crawler-platform/apps/spider-service/internal/model"
)

type fakeSpiderRepo struct {
	spiders []model.Spider
}

func (r *fakeSpiderRepo) Create(_ context.Context, spider model.Spider) error {
	r.spiders = append(r.spiders, spider)
	return nil
}

func (r *fakeSpiderRepo) ListByProject(_ context.Context, projectID string) ([]model.Spider, error) {
	var spiders []model.Spider
	for _, spider := range r.spiders {
		if spider.ProjectID == projectID {
			spiders = append(spiders, spider)
		}
	}
	return spiders, nil
}

func (r *fakeSpiderRepo) Get(_ context.Context, id string) (model.Spider, bool, error) {
	for _, spider := range r.spiders {
		if spider.ID == id {
			return spider, true, nil
		}
	}
	return model.Spider{}, false, nil
}

func TestCreateSpiderRejectsUnknownLanguage(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	_, err := svc.Create("p1", "bad", "ruby", "docker", "crawler/go:latest", []string{"./crawler"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrInvalidLanguage) {
		t.Fatalf("expected invalid language error, got %v", err)
	}
}

func TestCreateSpiderRejectsInvalidRuntime(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	_, err := svc.Create("p1", "bad", "go", "vm", "crawler/go:latest", []string{"./crawler"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrInvalidRuntime) {
		t.Fatalf("expected invalid runtime error, got %v", err)
	}
}

func TestCreateSpiderRequiresImageForDockerRuntime(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	_, err := svc.Create("p1", "crawler", "go", "docker", "", []string{"./crawler"})
	if err == nil {
		t.Fatal("expected missing image to fail")
	}
	if !errors.Is(err, ErrImageRequired) {
		t.Fatalf("expected image required error, got %v", err)
	}
}

func TestCreateSpiderAssignsIDAndPersistsSpider(t *testing.T) {
	repo := &fakeSpiderRepo{}
	svc := NewSpiderService(repo)

	spider, err := svc.Create("p1", "crawler", "python", "docker", "crawler/python:latest", []string{"python", "main.py"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	if spider.ID == "" {
		t.Fatal("expected generated id")
	}
	if spider.ProjectID != "p1" || spider.Name != "crawler" || spider.Language != "python" || spider.Runtime != "docker" {
		t.Fatalf("unexpected spider contents: %+v", spider)
	}
	if spider.Image != "crawler/python:latest" {
		t.Fatalf("expected image to persist, got %s", spider.Image)
	}
	if len(spider.Command) != 2 || spider.Command[0] != "python" || spider.Command[1] != "main.py" {
		t.Fatalf("expected command to persist, got %#v", spider.Command)
	}

	listed, err := svc.List("p1")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 spider, got %d", len(listed))
	}
	if listed[0].ID != spider.ID || listed[0].Image != spider.Image {
		t.Fatalf("expected list to return created spider, got %+v want %+v", listed[0], spider)
	}
}

func TestListFiltersByProjectID(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})

	spiderA, err := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go-a:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	spiderB, err := svc.Create("p2", "crawler-b", "python", "host", "", []string{"python", "main.py"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}

	listedP1, err := svc.List("p1")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listedP1) != 1 {
		t.Fatalf("expected 1 spider for p1, got %d", len(listedP1))
	}
	if listedP1[0].ID != spiderA.ID {
		t.Fatalf("expected p1 spider %+v, got %+v", spiderA, listedP1[0])
	}

	listedP2, err := svc.List("p2")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listedP2) != 1 {
		t.Fatalf("expected 1 spider for p2, got %d", len(listedP2))
	}
	if listedP2[0].ID != spiderB.ID {
		t.Fatalf("expected p2 spider %+v, got %+v", spiderB, listedP2[0])
	}
}
