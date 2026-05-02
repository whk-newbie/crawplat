package repo

import (
	"context"
	"regexp"
	"testing"
	"time"

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
		ID:        "s1",
		ProjectID: "p1",
		Name:      "crawler",
		Language:  "go",
		Runtime:   "docker",
		Image:     "crawler/go:latest",
		Command:   []string{"./crawler"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO spiders (id, project_id, name, language, runtime, image, command)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)).
		WithArgs(spider.ID, spider.ProjectID, spider.Name, spider.Language, spider.Runtime, spider.Image, `["./crawler"]`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO spider_versions (id, spider_id, version, registry_auth_ref, image, command)
		VALUES ($1, $2, 1, '', $3, $4)
	`)).
		WithArgs(sqlmock.AnyArg(), spider.ID, spider.Image, `["./crawler"]`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.Create(context.Background(), spider); err != nil {
		t.Fatalf("Create returned error: %v", err)
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
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM spiders WHERE id = $1 FOR UPDATE`)).
		WithArgs("s1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("s1"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(MAX(version), 0) FROM spider_versions WHERE spider_id = $1`)).
		WithArgs("s1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO spider_versions (id, spider_id, version, registry_auth_ref, image, command)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`)).
		WithArgs(sqlmock.AnyArg(), "s1", 2, "ghcr-prod", "crawler/go:v2", `["./crawler","--fast"]`).
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE spiders
		SET image = $2, command = $3
		WHERE id = $1
	`)).
		WithArgs("s1", "crawler/go:v2", `["./crawler","--fast"]`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	version, err := repo.CreateVersion(context.Background(), "s1", "ghcr-prod", "crawler/go:v2", []string{"./crawler", "--fast"})
	if err != nil {
		t.Fatalf("CreateVersion returned error: %v", err)
	}
	if version.SpiderID != "s1" || version.Version != 2 || version.Image != "crawler/go:v2" {
		t.Fatalf("unexpected version: %+v", version)
	}
	if version.RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("unexpected registry auth ref: %+v", version)
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
	rows := sqlmock.NewRows([]string{"id", "spider_id", "version", "registry_auth_ref", "image", "command", "created_at"}).
		AddRow("v2", "s1", 2, "ghcr-prod", "crawler/go:v2", `["./crawler","--fast"]`, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)).
		AddRow("v1", "s1", 1, "", "crawler/go:latest", `["./crawler"]`, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, spider_id, version, registry_auth_ref, image, command, created_at
		FROM spider_versions
		WHERE spider_id = $1
		ORDER BY version DESC
	`)).
		WithArgs("s1").
		WillReturnRows(rows)

	versions, err := repo.ListVersions(context.Background(), "s1")
	if err != nil {
		t.Fatalf("ListVersions returned error: %v", err)
	}
	if len(versions) != 2 || versions[0].Version != 2 || versions[1].Version != 1 {
		t.Fatalf("unexpected versions: %+v", versions)
	}
	if versions[0].RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("unexpected registry auth ref: %+v", versions[0])
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
		WHERE project_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`)).
		WithArgs("p1", 20, 0).
		WillReturnRows(rows)

	spiders, err := repo.ListByProject(context.Background(), "p1", 20, 0)
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
