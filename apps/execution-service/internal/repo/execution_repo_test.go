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

func TestExecutionRepositoryMarkRunning(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	startedAt := time.Now().UTC()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE executions
		SET node_id = $2, status = 'running', started_at = $3, finished_at = NULL, error_message = NULL
		WHERE id = $1 AND status = 'pending'
	`)).
		WithArgs("e1", "node-1", startedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "node_id", "status", "trigger_source", "image", "command", "error_message", "created_at", "started_at", "finished_at"}).
		AddRow("e1", "p1", "s1", "node-1", "running", "manual", "crawler/go:latest", `["./crawler"]`, nil, startedAt, startedAt, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at
		FROM executions
		WHERE id = $1
	`)).
		WithArgs("e1").
		WillReturnRows(rows)

	exec, err := repo.MarkRunning(context.Background(), "e1", "node-1", startedAt)
	if err != nil {
		t.Fatalf("MarkRunning returned error: %v", err)
	}
	if exec.Status != "running" || exec.NodeID != "node-1" || exec.StartedAt == nil {
		t.Fatalf("unexpected running execution: %+v", exec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestExecutionRepositoryComplete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	finishedAt := time.Now().UTC()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE executions
		SET status = 'succeeded', finished_at = $2, error_message = NULL
		WHERE id = $1 AND status = 'running'
	`)).
		WithArgs("e1", finishedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "node_id", "status", "trigger_source", "image", "command", "error_message", "created_at", "started_at", "finished_at"}).
		AddRow("e1", "p1", "s1", "node-1", "succeeded", "manual", "crawler/go:latest", `["./crawler"]`, nil, finishedAt, nil, finishedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at
		FROM executions
		WHERE id = $1
	`)).
		WithArgs("e1").
		WillReturnRows(rows)

	exec, err := repo.Complete(context.Background(), "e1", finishedAt)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if exec.Status != "succeeded" || exec.FinishedAt == nil {
		t.Fatalf("unexpected completed execution: %+v", exec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestExecutionRepositoryFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	finishedAt := time.Now().UTC()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE executions
		SET status = 'failed', finished_at = $2, error_message = $3
		WHERE id = $1 AND status = 'running'
	`)).
		WithArgs("e1", finishedAt, "exit status 1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "node_id", "status", "trigger_source", "image", "command", "error_message", "created_at", "started_at", "finished_at"}).
		AddRow("e1", "p1", "s1", "node-1", "failed", "manual", "crawler/go:latest", `["./crawler"]`, "exit status 1", finishedAt, nil, finishedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at
		FROM executions
		WHERE id = $1
	`)).
		WithArgs("e1").
		WillReturnRows(rows)

	exec, err := repo.Fail(context.Background(), "e1", "exit status 1", finishedAt)
	if err != nil {
		t.Fatalf("Fail returned error: %v", err)
	}
	if exec.Status != "failed" || exec.ErrorMessage != "exit status 1" || exec.FinishedAt == nil {
		t.Fatalf("unexpected failed execution: %+v", exec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestExecutionRepositoryCompleteReturnsInvalidStateWhenNotRunning(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	finishedAt := time.Now().UTC()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE executions
		SET status = 'succeeded', finished_at = $2, error_message = NULL
		WHERE id = $1 AND status = 'running'
	`)).
		WithArgs("e1", finishedAt).
		WillReturnResult(sqlmock.NewResult(0, 0))

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "node_id", "status", "trigger_source", "image", "command", "error_message", "created_at", "started_at", "finished_at"}).
		AddRow("e1", "p1", "s1", "node-1", "pending", "manual", "crawler/go:latest", `["./crawler"]`, nil, finishedAt, nil, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at
		FROM executions
		WHERE id = $1
	`)).
		WithArgs("e1").
		WillReturnRows(rows)

	_, err = repo.Complete(context.Background(), "e1", finishedAt)
	if !errors.Is(err, service.ErrInvalidExecutionState) {
		t.Fatalf("expected ErrInvalidExecutionState, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
