// 该文件为 PostgreSQL 持久化层的单元测试，使用 go-sqlmock 模拟数据库交互，
// 验证 SQL 语句的正确性和结果映射逻辑（包括 Command 字段的 JSON 序列化/反序列化）。
// 覆盖 spiders、spider_versions 表及 registry_auth_refs 查询。
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
		ID:        "s1",
		ProjectID: "p1",
		Name:      "crawler",
		Language:  "go",
		Runtime:   "docker",
		Image:     "crawler/go:latest",
		Command:   []string{"./crawler"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO spiders (id, project_id, name, language, runtime, image, command)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)).
		WithArgs(spider.ID, spider.ProjectID, spider.Name, spider.Language, spider.Runtime, spider.Image, `["./crawler"]`).
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
		WHERE project_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`)).WithArgs("p1", 20, 0).WillReturnRows(rows)

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
		Version:         "v1.0",
		Image:           "img:v1",
		RegistryAuthRef: "my-registry",
		Command:         []string{"./run"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO spider_versions (id, spider_id, version, image, registry_auth_ref, command)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)).
		WithArgs(version.ID, version.SpiderID, version.Version, version.Image, version.RegistryAuthRef, `["./run"]`).
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
