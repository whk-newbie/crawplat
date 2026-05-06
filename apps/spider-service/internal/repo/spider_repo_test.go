package repo

import (
	"context"
	"regexp"
	"testing"

	"crawler-platform/apps/spider-service/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	spider := model.Spider{
		ID:             "s1",
		ProjectID:      "p1",
		OrganizationID: "org-1",
		Name:           "crawler",
		Language:       "go",
		Runtime:        "docker",
		Image:          "crawler/go:latest",
		Command:        []string{"./crawler"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO spiders (id, project_id, name, language, runtime, image, command, organization_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`)).
		WithArgs(spider.ID, spider.ProjectID, spider.Name, spider.Language, spider.Runtime, spider.Image, `["./crawler"]`, spider.OrganizationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), spider); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestPostgresRepositoryCreateWithEmptyOrg(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	spider := model.Spider{
		ID:        "s1",
		ProjectID: "p1",
		Name:      "crawler",
		Language:  "go",
		Runtime:   "docker",
		Image:     "crawler/go:latest",
		Command:   []string{"./crawler"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO spiders (id, project_id, name, language, runtime, image, command, organization_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`)).
		WithArgs(spider.ID, spider.ProjectID, spider.Name, spider.Language, spider.Runtime, spider.Image, `["./crawler"]`, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), spider); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestPostgresRepositoryListByProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "language", "runtime", "image", "command"}).
		AddRow("s1", "p1", "crawler", "go", "docker", "crawler/go:latest", `["./crawler"]`)

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, project_id, name, language, runtime, image, command
			FROM spiders
			WHERE ($1 = '' OR organization_id = $1) AND project_id = $2
			ORDER BY created_at DESC, id DESC
			LIMIT $3 OFFSET $4
		`)).WithArgs("org-1", "p1", 20, 0).WillReturnRows(rows)

	spiders, err := repo.ListByProject(context.Background(), "org-1", "p1", 20, 0)
	if err != nil {
		t.Fatalf("ListByProject returned error: %v", err)
	}
	if len(spiders) != 1 || spiders[0].Command[0] != "./crawler" {
		t.Fatalf("unexpected spiders: %#v", spiders)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestPostgresRepositoryListByProjectEmptyOrg(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "language", "runtime", "image", "command"}).
		AddRow("s1", "p1", "crawler", "go", "docker", "crawler/go:latest", `["./crawler"]`)

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, project_id, name, language, runtime, image, command
			FROM spiders
			WHERE ($1 = '' OR organization_id = $1) AND project_id = $2
			ORDER BY created_at DESC, id DESC
			LIMIT $3 OFFSET $4
		`)).WithArgs("", "p1", 20, 0).WillReturnRows(rows)

	spiders, err := repo.ListByProject(context.Background(), "", "p1", 20, 0)
	if err != nil {
		t.Fatalf("ListByProject returned error: %v", err)
	}
	if len(spiders) != 1 {
		t.Fatalf("expected 1 spider with empty orgID, got %d", len(spiders))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestPostgresRepositoryCreateVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	version := model.SpiderVersion{
		ID:              "v1",
		SpiderID:        "s1",
		OrganizationID:  "org-1",
		Version:         "v1.0",
		Image:           "img:v1",
		RegistryAuthRef: "my-registry",
		Command:         []string{"./run"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
			INSERT INTO spider_versions (id, spider_id, version, image, registry_auth_ref, command, organization_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`)).
		WithArgs(version.ID, version.SpiderID, version.Version, version.Image, version.RegistryAuthRef, `["./run"]`, version.OrganizationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.CreateVersion(context.Background(), version); err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestPostgresRepositoryListVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	rows := sqlmock.NewRows([]string{"id", "spider_id", "version", "image", "registry_auth_ref", "command"}).
		AddRow("v1", "s1", "v1.0", "img:v1", "my-registry", `["./run"]`)

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, spider_id, version, image, registry_auth_ref, command
			FROM spider_versions
			WHERE spider_id = $1
			ORDER BY created_at DESC, id DESC
		`)).WithArgs("s1").WillReturnRows(rows)

	versions, err := repo.ListVersions(context.Background(), "s1")
	if err != nil {
		t.Fatalf("ListVersions returned error: %v", err)
	}
	if len(versions) != 1 || versions[0].Version != "v1.0" {
		t.Fatalf("unexpected versions: %#v", versions)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
