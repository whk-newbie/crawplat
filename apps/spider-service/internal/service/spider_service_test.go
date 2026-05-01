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

func (r *fakeSpiderRepo) ListByProject(_ context.Context, projectID string, limit, offset int) ([]model.Spider, error) {
	var all []model.Spider
	for _, spider := range r.spiders {
		if spider.ProjectID == projectID {
			all = append(all, spider)
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

func (r *fakeSpiderRepo) CountByProject(_ context.Context, projectID string) (int64, error) {
	var count int64
	for _, spider := range r.spiders {
		if spider.ProjectID == projectID {
			count++
		}
	}
	return count, nil
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
	if !errors.Is(err, ErrInvalidLanguage) {
		t.Fatalf("expected invalid language error, got %v", err)
	}
}

func TestCreateSpiderRejectsInvalidRuntime(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	_, err := svc.Create("p1", "bad", "go", "vm", "crawler/go:latest", []string{"./crawler"})
	if !errors.Is(err, ErrInvalidRuntime) {
		t.Fatalf("expected invalid runtime error, got %v", err)
	}
}

func TestCreateSpiderRequiresImageForDockerRuntime(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	_, err := svc.Create("p1", "crawler", "go", "docker", "", []string{"./crawler"})
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

	listed, total, err := svc.List("p1", 20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(listed) != 1 || listed[0].ID != spider.ID {
		t.Fatalf("expected list to return created spider")
	}
}

func TestListFiltersByProjectID(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})

	spiderA, _ := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go-a:latest", []string{"./crawler-a"})
	spiderB, _ := svc.Create("p2", "crawler-b", "python", "host", "", []string{"python", "main.py"})

	listedP1, totalP1, _ := svc.List("p1", 20, 0)
	if totalP1 != 1 || len(listedP1) != 1 || listedP1[0].ID != spiderA.ID {
		t.Fatalf("expected p1 spider %s, got %+v", spiderA.ID, listedP1)
	}

	listedP2, totalP2, _ := svc.List("p2", 20, 0)
	if totalP2 != 1 || len(listedP2) != 1 || listedP2[0].ID != spiderB.ID {
		t.Fatalf("expected p2 spider %s, got %+v", spiderB.ID, listedP2)
	}
}
