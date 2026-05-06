// Package repo 的单元测试，使用 sqlmock 模拟数据库交互。
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
	project := model.Project{ID: "p1", Code: "crawler", Name: "Crawler", OrganizationID: "org-1"}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO projects (id, code, name, organization_id)
		VALUES ($1, $2, $3, $4)
	`)).
		WithArgs(project.ID, project.Code, project.Name, project.OrganizationID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), project); err != nil {
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
	project := model.Project{ID: "p1", Code: "crawler", Name: "Crawler"}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO projects (id, code, name, organization_id)
		VALUES ($1, $2, $3, $4)
	`)).
		WithArgs(project.ID, project.Code, project.Name, nil).
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
		WHERE ($1 = '' OR organization_id = $1)
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`)).WithArgs("org-1", 20, 0).WillReturnRows(rows)

	projects, err := repo.List(context.Background(), "org-1", 20, 0)
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

func TestPostgresRepositoryExistsByCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(SELECT 1 FROM projects WHERE ($1 = '' OR organization_id = $1) AND code = $2)
	`)).WithArgs("org-1", "crawler").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ExistsByCode(context.Background(), "org-1", "crawler")
	if err != nil {
		t.Fatalf("ExistsByCode returned error: %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
