package repo

import (
	"context"
	"regexp"
	"testing"

	"crawler-platform/apps/datasource-service/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	datasource := model.Datasource{
		ID:        "d1",
		ProjectID: "p1",
		Name:      "main",
		Type:      "postgresql",
		Readonly:  true,
		Config:    map[string]string{"schema": "public"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO datasources (id, project_id, name, type, readonly, config_json)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)).
		WithArgs(datasource.ID, datasource.ProjectID, datasource.Name, datasource.Type, datasource.Readonly, `{"schema":"public"}`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), datasource); err != nil {
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
	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "type", "readonly", "config_json"}).
		AddRow("d1", "p1", "main", "redis", true, `{"db":"0"}`)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, name, type, readonly, config_json
		FROM datasources
		WHERE project_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`)).WithArgs("p1", 20, 0).WillReturnRows(rows)

	datasources, err := repo.ListByProject(context.Background(), "p1", 20, 0)
	if err != nil {
		t.Fatalf("ListByProject returned error: %v", err)
	}
	if len(datasources) != 1 || datasources[0].Config["db"] != "0" {
		t.Fatalf("unexpected datasources: %#v", datasources)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
