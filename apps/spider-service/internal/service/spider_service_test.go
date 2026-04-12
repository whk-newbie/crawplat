package service

import (
	"errors"
	"testing"
)

func TestCreateSpiderRejectsUnknownLanguage(t *testing.T) {
	svc := NewSpiderService()
	_, err := svc.Create("p1", "bad", "ruby", "docker")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrInvalidLanguage) {
		t.Fatalf("expected invalid language error, got %v", err)
	}
}

func TestCreateSpiderRejectsInvalidRuntime(t *testing.T) {
	svc := NewSpiderService()
	_, err := svc.Create("p1", "bad", "go", "vm")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrInvalidRuntime) {
		t.Fatalf("expected invalid runtime error, got %v", err)
	}
}

func TestCreateSpiderAssignsIDAndPersistsSpider(t *testing.T) {
	svc := NewSpiderService()

	spider, err := svc.Create("p1", "crawler", "python", "docker")
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	if spider.ID == "" {
		t.Fatal("expected generated id")
	}
	if spider.ProjectID != "p1" || spider.Name != "crawler" || spider.Language != "python" || spider.Runtime != "docker" {
		t.Fatalf("unexpected spider contents: %+v", spider)
	}

	listed := svc.List("p1")
	if len(listed) != 1 {
		t.Fatalf("expected 1 spider, got %d", len(listed))
	}
	if listed[0] != spider {
		t.Fatalf("expected list to return created spider, got %+v want %+v", listed[0], spider)
	}
}

func TestListFiltersByProjectID(t *testing.T) {
	svc := NewSpiderService()

	spiderA, err := svc.Create("p1", "crawler-a", "go", "docker")
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	spiderB, err := svc.Create("p2", "crawler-b", "python", "host")
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}

	listedP1 := svc.List("p1")
	if len(listedP1) != 1 {
		t.Fatalf("expected 1 spider for p1, got %d", len(listedP1))
	}
	if listedP1[0] != spiderA {
		t.Fatalf("expected p1 spider %+v, got %+v", spiderA, listedP1[0])
	}

	listedP2 := svc.List("p2")
	if len(listedP2) != 1 {
		t.Fatalf("expected 1 spider for p2, got %d", len(listedP2))
	}
	if listedP2[0] != spiderB {
		t.Fatalf("expected p2 spider %+v, got %+v", spiderB, listedP2[0])
	}
}
