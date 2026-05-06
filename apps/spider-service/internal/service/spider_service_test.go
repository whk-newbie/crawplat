package service

import (
	"context"
	"errors"
	"testing"

	"crawler-platform/apps/spider-service/internal/model"
)

type fakeSpiderRepo struct {
	spiders  []model.Spider
	versions []model.SpiderVersion
}

func (r *fakeSpiderRepo) Create(_ context.Context, spider model.Spider) error {
	r.spiders = append(r.spiders, spider)
	return nil
}

func (r *fakeSpiderRepo) ListByProject(_ context.Context, orgID, projectID string, limit, offset int) ([]model.Spider, error) {
	if limit <= 0 {
		limit = 20
	}
	var spiders []model.Spider
	for _, spider := range r.spiders {
		if orgID != "" && spider.OrganizationID != orgID {
			continue
		}
		if spider.ProjectID == projectID {
			spiders = append(spiders, spider)
		}
	}
	if offset >= len(spiders) {
		return []model.Spider{}, nil
	}
	end := offset + limit
	if end > len(spiders) {
		end = len(spiders)
	}
	return spiders[offset:end], nil
}

func (r *fakeSpiderRepo) Get(_ context.Context, id string) (model.Spider, bool, error) {
	for _, spider := range r.spiders {
		if spider.ID == id {
			return spider, true, nil
		}
	}
	return model.Spider{}, false, nil
}

func (r *fakeSpiderRepo) CreateVersion(_ context.Context, version model.SpiderVersion) error {
	r.versions = append(r.versions, version)
	return nil
}

func (r *fakeSpiderRepo) ListVersions(_ context.Context, spiderID string) ([]model.SpiderVersion, error) {
	var versions []model.SpiderVersion
	for _, v := range r.versions {
		if v.SpiderID == spiderID {
			versions = append(versions, v)
		}
	}
	return versions, nil
}

func (r *fakeSpiderRepo) ListRegistryAuthRefs(_ context.Context, _ string) ([]model.RegistryAuthRef, error) {
	return []model.RegistryAuthRef{}, nil
}

func TestCreateSpiderRejectsUnknownLanguage(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	_, err := svc.Create("", "p1", "bad", "ruby", "docker", "crawler/go:latest", []string{"./crawler"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrInvalidLanguage) {
		t.Fatalf("expected invalid language error, got %v", err)
	}
}

func TestCreateSpiderRejectsInvalidRuntime(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	_, err := svc.Create("", "p1", "bad", "go", "vm", "crawler/go:latest", []string{"./crawler"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(err, ErrInvalidRuntime) {
		t.Fatalf("expected invalid runtime error, got %v", err)
	}
}

func TestCreateSpiderRequiresImageForDockerRuntime(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	_, err := svc.Create("", "p1", "crawler", "go", "docker", "", []string{"./crawler"})
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

	spider, err := svc.Create("org-1", "p1", "crawler", "python", "docker", "crawler/python:latest", []string{"python", "main.py"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	if spider.ID == "" {
		t.Fatal("expected generated id")
	}
	if spider.OrganizationID != "org-1" {
		t.Fatalf("expected OrganizationID org-1, got %s", spider.OrganizationID)
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

	listed, err := svc.List("org-1", "p1", 20, 0)
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

	spiderA, err := svc.Create("", "p1", "crawler-a", "go", "docker", "crawler/go-a:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	spiderB, err := svc.Create("", "p2", "crawler-b", "python", "host", "", []string{"python", "main.py"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}

	listedP1, err := svc.List("", "p1", 20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listedP1) != 1 {
		t.Fatalf("expected 1 spider for p1, got %d", len(listedP1))
	}
	if listedP1[0].ID != spiderA.ID {
		t.Fatalf("expected p1 spider %+v, got %+v", spiderA, listedP1[0])
	}

	listedP2, err := svc.List("", "p2", 20, 0)
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

func TestCreateVersionRequiresExistingSpider(t *testing.T) {
	repo := &fakeSpiderRepo{}
	svc := NewSpiderService(repo)

	_, err := svc.CreateVersion("nonexistent", "v1.0", "img:latest", "", nil)
	if !errors.Is(err, ErrSpiderNotFound) {
		t.Fatalf("expected ErrSpiderNotFound, got: %v", err)
	}
}

func TestCreateVersionAndList(t *testing.T) {
	repo := &fakeSpiderRepo{}
	svc := NewSpiderService(repo)

	spider, err := svc.Create("org-1", "p1", "crawler", "go", "docker", "img:latest", nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	v, err := svc.CreateVersion(spider.ID, "v1.0", "img:v1", "my-registry", []string{"./run"})
	if err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	if v.ID == "" {
		t.Fatal("expected generated id for version")
	}
	if v.SpiderID != spider.ID {
		t.Fatalf("expected spiderID %s, got %s", spider.ID, v.SpiderID)
	}
	if v.OrganizationID != "org-1" {
		t.Fatalf("expected OrganizationID org-1 on version, got %s", v.OrganizationID)
	}
	if v.RegistryAuthRef != "my-registry" {
		t.Fatalf("expected registryAuthRef my-registry, got %s", v.RegistryAuthRef)
	}

	versions, err := svc.ListVersions(spider.ID)
	if err != nil {
		t.Fatalf("ListVersions returned error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
}

func TestListRegistryAuthRefs(t *testing.T) {
	repo := &fakeSpiderRepo{}
	svc := NewSpiderService(repo)

	refs, err := svc.ListRegistryAuthRefs("p1")
	if err != nil {
		t.Fatalf("ListRegistryAuthRefs returned error: %v", err)
	}
	if refs == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestSpiderServiceListFiltersByOrg(t *testing.T) {
	repo := &fakeSpiderRepo{}
	svc := NewSpiderService(repo)

	if _, err := svc.Create("org-a", "p1", "spider-org-a", "go", "host", "", nil); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create("org-b", "p1", "spider-org-b", "go", "host", "", nil); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	listed, err := svc.List("org-a", "p1", 20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 spider for org-a, got %d", len(listed))
	}
	if listed[0].Name != "spider-org-a" {
		t.Fatalf("expected spider-org-a, got %s", listed[0].Name)
	}
}

func TestSpiderServiceCreateAllowsSameNameInDifferentOrgs(t *testing.T) {
	repo := &fakeSpiderRepo{}
	svc := NewSpiderService(repo)

	s1, err := svc.Create("org-a", "p1", "crawler", "go", "host", "", nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	s2, err := svc.Create("org-b", "p1", "crawler", "go", "host", "", nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if s1.ID == s2.ID {
		t.Fatal("expected different IDs for different orgs")
	}

	listedA, _ := svc.List("org-a", "p1", 20, 0)
	listedB, _ := svc.List("org-b", "p1", 20, 0)
	if len(listedA) != 1 || listedA[0].OrganizationID != "org-a" {
		t.Fatalf("org-a should see 1 spider, got %d", len(listedA))
	}
	if len(listedB) != 1 || listedB[0].OrganizationID != "org-b" {
		t.Fatalf("org-b should see 1 spider, got %d", len(listedB))
	}
}
