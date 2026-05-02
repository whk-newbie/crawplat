package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"crawler-platform/apps/spider-service/internal/model"
	"github.com/google/uuid"
)

type fakeSpiderRepo struct {
	spiders  []model.Spider
	versions map[string][]model.SpiderVersion
}

func (r *fakeSpiderRepo) Create(_ context.Context, spider model.Spider) error {
	r.spiders = append(r.spiders, spider)
	if r.versions == nil {
		r.versions = map[string][]model.SpiderVersion{}
	}
	r.versions[spider.ID] = []model.SpiderVersion{{
		ID:        uuid.NewString(),
		SpiderID:  spider.ID,
		Version:   1,
		Image:     spider.Image,
		Command:   append([]string(nil), spider.Command...),
		CreatedAt: time.Now().UTC(),
	}}
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

func (r *fakeSpiderRepo) CreateVersion(_ context.Context, spiderID, registryAuthRef, image string, command []string) (model.SpiderVersion, error) {
	for i := range r.spiders {
		if r.spiders[i].ID != spiderID {
			continue
		}
		existing := r.versions[spiderID]
		nextVersion := 1
		if len(existing) > 0 {
			nextVersion = existing[0].Version + 1
		}
		created := model.SpiderVersion{
			ID:              uuid.NewString(),
			SpiderID:        spiderID,
			Version:         nextVersion,
			RegistryAuthRef: registryAuthRef,
			Image:           image,
			Command:         append([]string(nil), command...),
			CreatedAt:       time.Now().UTC(),
		}
		r.versions[spiderID] = append([]model.SpiderVersion{created}, existing...)
		r.spiders[i].Image = image
		r.spiders[i].Command = append([]string(nil), command...)
		return created, nil
	}
	return model.SpiderVersion{}, ErrSpiderNotFound
}

func (r *fakeSpiderRepo) ListVersions(_ context.Context, spiderID string) ([]model.SpiderVersion, error) {
	versions := r.versions[spiderID]
	out := make([]model.SpiderVersion, 0, len(versions))
	for _, version := range versions {
		version.Command = append([]string(nil), version.Command...)
		out = append(out, version)
	}
	return out, nil
}

func (r *fakeSpiderRepo) ListRegistryAuthRefsByProject(_ context.Context, projectID string) ([]string, error) {
	refSet := map[string]struct{}{}
	for _, spider := range r.spiders {
		if spider.ProjectID != projectID {
			continue
		}
		for _, version := range r.versions[spider.ID] {
			if version.RegistryAuthRef == "" {
				continue
			}
			refSet[version.RegistryAuthRef] = struct{}{}
		}
	}

	var refs []string
	for ref := range refSet {
		refs = append(refs, ref)
	}
	return refs, nil
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

func TestCreateSpiderSeedsVersionAndListVersions(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})

	spider, err := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	versions, err := svc.ListVersions(spider.ID)
	if err != nil {
		t.Fatalf("ListVersions returned error: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].Version != 1 || versions[0].Image != "crawler/go:latest" {
		t.Fatalf("unexpected initial version: %+v", versions[0])
	}
}

func TestCreateVersionAppendsSequentialVersion(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	spider, err := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	version, err := svc.CreateVersion(spider.ID, "ghcr-prod", "crawler/go:v2", []string{"./crawler-a", "--fast"})
	if err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	if version.Version != 2 || version.Image != "crawler/go:v2" {
		t.Fatalf("unexpected created version: %+v", version)
	}
	if version.RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("expected registry auth ref to persist, got %+v", version)
	}

	versions, err := svc.ListVersions(spider.ID)
	if err != nil {
		t.Fatalf("ListVersions returned error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].Version != 2 || versions[1].Version != 1 {
		t.Fatalf("expected desc order by version, got %+v", versions)
	}
	if versions[0].RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("expected latest version to carry registry auth ref, got %+v", versions[0])
	}
}

func TestListRegistryAuthRefsByProjectReturnsUniqueValues(t *testing.T) {
	svc := NewSpiderService(&fakeSpiderRepo{})
	spiderP1A, _ := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go-a:latest", []string{"./crawler-a"})
	spiderP1B, _ := svc.Create("p1", "crawler-b", "go", "docker", "crawler/go-b:latest", []string{"./crawler-b"})
	spiderP2, _ := svc.Create("p2", "crawler-c", "go", "docker", "crawler/go-c:latest", []string{"./crawler-c"})

	if _, err := svc.CreateVersion(spiderP1A.ID, "ghcr-prod", "crawler/go-a:v2", []string{"./crawler-a"}); err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	if _, err := svc.CreateVersion(spiderP1B.ID, "harbor-ci", "crawler/go-b:v2", []string{"./crawler-b"}); err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	if _, err := svc.CreateVersion(spiderP1A.ID, "ghcr-prod", "crawler/go-a:v3", []string{"./crawler-a"}); err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	if _, err := svc.CreateVersion(spiderP2.ID, "dockerhub", "crawler/go-c:v2", []string{"./crawler-c"}); err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}

	refs, err := svc.ListRegistryAuthRefsByProject("p1")
	if err != nil {
		t.Fatalf("ListRegistryAuthRefsByProject returned error: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 unique refs, got %+v", refs)
	}
}
