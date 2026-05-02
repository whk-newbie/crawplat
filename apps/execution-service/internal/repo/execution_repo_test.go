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
		ID:                 "e1",
		ProjectID:          "p1",
		SpiderID:           "s1",
		Status:             "pending",
		TriggerSource:      "manual",
		Image:              "crawler/go:latest",
		Command:            []string{"./crawler"},
		CPUCores:           1.5,
		MemoryMB:           768,
		TimeoutSeconds:     120,
		RetryLimit:         3,
		RetryCount:         1,
		RetryDelaySeconds:  45,
		RetryOfExecutionID: "e0",
		StartedAt:          &now,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO executions (id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, started_at, finished_at, error_message, retried_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`)).
		WithArgs(exec.ID, exec.ProjectID, exec.SpiderID, exec.SpiderVersion, nil, nil, exec.Status, exec.TriggerSource, exec.Image, `["./crawler"]`, exec.CPUCores, exec.MemoryMB, exec.TimeoutSeconds, exec.RetryLimit, exec.RetryCount, exec.RetryDelaySeconds, exec.RetryOfExecutionID, exec.StartedAt, exec.FinishedAt, nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := repo.Create(context.Background(), exec); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestExecutionRepositoryListByProjectAndCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	now := time.Date(2026, 5, 2, 13, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "node_id", "status", "trigger_source", "image", "command", "cpu_cores", "memory_mb", "timeout_seconds", "error_message", "created_at", "started_at", "finished_at", "retry_limit", "retry_count", "retry_delay_seconds", "retry_of_execution_id", "retried_at"}).
		AddRow("e2", "p1", "s2", 2, "ghcr-prod", nil, "pending", "manual", "crawler/go:v2", `["./crawler","--v2"]`, 0.5, 512, 120, nil, now, nil, nil, 0, 0, 0, nil, nil).
		AddRow("e1", "p1", "s1", 1, nil, "node-a", "running", "scheduled", "crawler/go:v1", `["./crawler"]`, 1.0, 1024, 300, nil, now.Add(-time.Minute), now.Add(-time.Minute), nil, 2, 1, 30, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
		FROM executions
		WHERE project_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`)).
		WithArgs("p1", 20, 0).
		WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(1)
		FROM executions
		WHERE project_id = $1
	`)).
		WithArgs("p1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	items, err := repo.ListByProject(context.Background(), service.ListExecutionsQuery{
		ProjectID: "p1",
		Limit:     20,
		Offset:    0,
	})
	if err != nil {
		t.Fatalf("ListByProject returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "e2" || items[0].RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].ID != "e1" || items[1].NodeID != "node-a" || items[1].Status != "running" {
		t.Fatalf("unexpected second item: %+v", items[1])
	}

	total, err := repo.CountByProject(context.Background(), service.ListExecutionsQuery{ProjectID: "p1"})
	if err != nil {
		t.Fatalf("CountByProject returned error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestExecutionRepositoryListByProjectAndCountWithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	from := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 2, 14, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "node_id", "status", "trigger_source", "image", "command", "cpu_cores", "memory_mb", "timeout_seconds", "error_message", "created_at", "started_at", "finished_at", "retry_limit", "retry_count", "retry_delay_seconds", "retry_of_execution_id", "retried_at"}).
		AddRow("e3", "p1", "s3", 3, nil, nil, "failed", "manual", "crawler/go:v3", `["./crawler"]`, 1.0, 512, 120, "boom", from.Add(30*time.Minute), nil, nil, 0, 0, 0, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
		FROM executions
		WHERE project_id = $1 AND status = $2 AND trigger_source = $3 AND spider_id = $4 AND created_at >= $5 AND created_at <= $6
		ORDER BY created_at DESC, id DESC
		LIMIT $7 OFFSET $8
	`)).
		WithArgs("p1", "failed", "manual", "s3", from, to, 10, 0).
		WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(1)
		FROM executions
		WHERE project_id = $1 AND status = $2 AND trigger_source = $3 AND spider_id = $4 AND created_at >= $5 AND created_at <= $6
	`)).
		WithArgs("p1", "failed", "manual", "s3", from, to).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	query := service.ListExecutionsQuery{
		ProjectID: "p1",
		Status:    "failed",
		Trigger:   "manual",
		SpiderID:  "s3",
		From:      &from,
		To:        &to,
		Limit:     10,
		Offset:    0,
	}
	items, err := repo.ListByProject(context.Background(), query)
	if err != nil {
		t.Fatalf("ListByProject returned error: %v", err)
	}
	if len(items) != 1 || items[0].ID != "e3" {
		t.Fatalf("unexpected filtered items: %+v", items)
	}
	total, err := repo.CountByProject(context.Background(), query)
	if err != nil {
		t.Fatalf("CountByProject returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestExecutionRepositoryClaimNextRetryCandidate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewExecutionRepository(db)
	now := time.Date(2026, 4, 23, 23, 55, 0, 0, time.UTC)
	finishedAt := now.Add(-time.Minute)
	retriedAt := now

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "node_id", "status", "trigger_source", "image", "command", "cpu_cores", "memory_mb", "timeout_seconds", "error_message", "created_at", "started_at", "finished_at", "retry_limit", "retry_count", "retry_delay_seconds", "retry_of_execution_id", "retried_at"}).
		AddRow("e1", "p1", "s1", 1, nil, nil, "failed", "scheduled", "crawler/go:latest", `["./crawler"]`, 0.5, 512, 90, "boom", now.Add(-2*time.Minute), nil, finishedAt, 3, 0, 30, nil, retriedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE executions
		SET retried_at = $2
		WHERE id = (
			SELECT id
			FROM executions
			WHERE status = 'failed'
			  AND retried_at IS NULL
			  AND retry_limit > retry_count
			  AND finished_at IS NOT NULL
			  AND finished_at + make_interval(secs => retry_delay_seconds) <= $1
			ORDER BY finished_at ASC, id ASC
			LIMIT 1
		)
		RETURNING id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
	`)).
		WithArgs(now, now).
		WillReturnRows(rows)

	exec, ok, err := repo.ClaimNextRetryCandidate(context.Background(), now)
	if err != nil {
		t.Fatalf("ClaimNextRetryCandidate returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected retry candidate")
	}
	if exec.ID != "e1" || exec.RetryLimit != 3 || exec.RetryDelaySeconds != 30 || exec.RetriedAt == nil {
		t.Fatalf("unexpected retry candidate: %+v", exec)
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
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
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
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
		FROM executions
		WHERE id = $1
	`)).
		WithArgs("missing").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "node_id", "status", "trigger_source", "image", "command", "cpu_cores", "memory_mb", "timeout_seconds", "error_message", "created_at", "started_at", "finished_at", "retry_limit", "retry_count", "retry_delay_seconds", "retry_of_execution_id", "retried_at"}))

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

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "node_id", "status", "trigger_source", "image", "command", "cpu_cores", "memory_mb", "timeout_seconds", "error_message", "created_at", "started_at", "finished_at", "retry_limit", "retry_count", "retry_delay_seconds", "retry_of_execution_id", "retried_at"}).
		AddRow("e1", "p1", "s1", 1, nil, "node-1", "running", "manual", "crawler/go:latest", `["./crawler"]`, 0.25, 256, 30, nil, startedAt, startedAt, nil, 0, 0, 0, nil, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
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

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "node_id", "status", "trigger_source", "image", "command", "cpu_cores", "memory_mb", "timeout_seconds", "error_message", "created_at", "started_at", "finished_at", "retry_limit", "retry_count", "retry_delay_seconds", "retry_of_execution_id", "retried_at"}).
		AddRow("e1", "p1", "s1", 1, nil, "node-1", "succeeded", "manual", "crawler/go:latest", `["./crawler"]`, 1.0, 1024, 600, nil, finishedAt, nil, finishedAt, 0, 0, 0, nil, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
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

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "node_id", "status", "trigger_source", "image", "command", "cpu_cores", "memory_mb", "timeout_seconds", "error_message", "created_at", "started_at", "finished_at", "retry_limit", "retry_count", "retry_delay_seconds", "retry_of_execution_id", "retried_at"}).
		AddRow("e1", "p1", "s1", 1, nil, "node-1", "failed", "manual", "crawler/go:latest", `["./crawler"]`, 2.0, 2048, 1200, "exit status 1", finishedAt, nil, finishedAt, 0, 0, 0, nil, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
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

	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "node_id", "status", "trigger_source", "image", "command", "cpu_cores", "memory_mb", "timeout_seconds", "error_message", "created_at", "started_at", "finished_at", "retry_limit", "retry_count", "retry_delay_seconds", "retry_of_execution_id", "retried_at"}).
		AddRow("e1", "p1", "s1", 1, nil, "node-1", "pending", "manual", "crawler/go:latest", `["./crawler"]`, 0.1, 128, 10, nil, finishedAt, nil, nil, 0, 0, 0, nil, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
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
