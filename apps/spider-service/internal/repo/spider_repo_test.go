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
