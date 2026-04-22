package repo

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"crawler-platform/apps/execution-service/internal/model"
	"crawler-platform/apps/execution-service/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestExecutionRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	now := time.Now().UTC()
	exec := model.Execution{
		ID:            "e1",
		ProjectID:     "p1",
		SpiderID:      "s1",
		Status:        "pending",
		TriggerSource: "manual",
		Image:         "crawler/go:latest",
		Command:       []string{"./crawler"},
		StartedAt:     &now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO executions (id, project_id, spider_id, node_id, status, trigger_source, image, command, started_at, finished_at, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`)).
		WithArgs(exec.ID, exec.ProjectID, exec.SpiderID, nil, exec.Status, exec.TriggerSource, exec.Image, `["./crawler"]`, exec.StartedAt, exec.FinishedAt, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := repo.Create(context.Background(), exec); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestExecutionRepositoryGetReturnsBackendError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	expectedErr := errors.New("postgres unavailable")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at
		FROM executions
		WHERE id = $1
	`)).
		WithArgs("missing").
		WillReturnError(expectedErr)

	_, err = repo.Get(context.Background(), "missing")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected backend error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestExecutionRepositoryGetMapsNoRowsToNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at
		FROM executions
		WHERE id = $1
	`)).
		WithArgs("missing").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "spider_id", "node_id", "status", "trigger_source", "image", "command", "error_message", "created_at", "started_at", "finished_at"}))

	_, err = repo.Get(context.Background(), "missing")
	if err == nil || err != service.ErrExecutionNotFound {
		t.Fatalf("expected ErrExecutionNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
