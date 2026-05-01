package repo

import (
	"context"
	"regexp"
	"testing"

	"crawler-platform/apps/project-service/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	project := model.Project{ID: "p1", Code: "crawler", Name: "Crawler"}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO projects (id, code, name)
		VALUES ($1, $2, $3)
	`)).
		WithArgs(project.ID, project.Code, project.Name).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), project); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestPostgresRepositoryList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	rows := sqlmock.NewRows([]string{"id", "code", "name"}).AddRow("p1", "crawler", "Crawler")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, code, name
		FROM projects
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`)).
		WithArgs(20, 0).
		WillReturnRows(rows)

	projects, err := repo.List(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(projects) != 1 || projects[0].ID != "p1" {
		t.Fatalf("unexpected projects: %#v", projects)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
