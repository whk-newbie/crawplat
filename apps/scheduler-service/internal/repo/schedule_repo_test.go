package repo

import (
	"context"
	"regexp"
	"testing"
	"time"

	"crawler-platform/apps/scheduler-service/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	schedule := model.Schedule{
		ID:                "sched-1",
		ProjectID:         "project-1",
		SpiderID:          "spider-1",
		SpiderVersion:     2,
		RegistryAuthRef:   "ghcr-prod",
		Name:              "nightly",
		CronExpr:          "0 * * * *",
		Enabled:           true,
		Image:             "crawler/go-echo:latest",
		Command:           []string{"./go-echo"},
		RetryLimit:        2,
		RetryDelaySeconds: 30,
		CreatedAt:         time.Date(2026, 4, 23, 23, 40, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO scheduled_tasks (id, project_id, spider_id, spider_version, registry_auth_ref, name, cron_expr, enabled, image, command, retry_limit, retry_delay_seconds, created_at, last_materialized_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12, $13, $14)
	`)).
		WithArgs(schedule.ID, schedule.ProjectID, schedule.SpiderID, schedule.SpiderVersion, schedule.RegistryAuthRef, schedule.Name, schedule.CronExpr, schedule.Enabled, schedule.Image, `["./go-echo"]`, schedule.RetryLimit, schedule.RetryDelaySeconds, schedule.CreatedAt, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), schedule); err != nil {
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
	createdAt := time.Date(2026, 4, 23, 23, 40, 0, 0, time.UTC)
	lastMaterializedAt := createdAt.Add(5 * time.Minute)
	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "spider_version", "registry_auth_ref", "name", "cron_expr", "enabled", "image", "command", "retry_limit", "retry_delay_seconds", "created_at", "last_materialized_at"}).
		AddRow("sched-1", "project-1", "spider-1", 2, "ghcr-prod", "nightly", "0 * * * *", true, "crawler/go-echo:latest", `["./go-echo"]`, 2, 30, createdAt, lastMaterializedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, name, cron_expr, enabled, image, command, retry_limit, retry_delay_seconds, created_at, last_materialized_at
		FROM scheduled_tasks
		ORDER BY created_at DESC, id DESC
	`)).WillReturnRows(rows)

	schedules, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(schedules) != 1 || schedules[0].ID != "sched-1" {
		t.Fatalf("unexpected schedules: %#v", schedules)
	}
	if schedules[0].RetryLimit != 2 || schedules[0].RetryDelaySeconds != 30 {
		t.Fatalf("unexpected retry config: %#v", schedules[0])
	}
	if schedules[0].SpiderVersion != 2 {
		t.Fatalf("unexpected spiderVersion: %#v", schedules[0])
	}
	if schedules[0].RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("unexpected registryAuthRef: %#v", schedules[0])
	}
	if schedules[0].LastMaterializedAt == nil || !schedules[0].LastMaterializedAt.Equal(lastMaterializedAt) {
		t.Fatalf("unexpected lastMaterializedAt: %#v", schedules[0].LastMaterializedAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

func TestPostgresRepositoryAdvanceLastMaterialized(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	next := time.Date(2026, 4, 23, 23, 45, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta(`
			UPDATE scheduled_tasks
			SET last_materialized_at = $2
			WHERE id = $1 AND last_materialized_at IS NULL
		`)).
		WithArgs("sched-1", next).
		WillReturnResult(sqlmock.NewResult(0, 1))

	claimed, err := repo.AdvanceLastMaterialized(context.Background(), "sched-1", nil, next)
	if err != nil {
		t.Fatalf("AdvanceLastMaterialized returned error: %v", err)
	}
	if !claimed {
		t.Fatal("expected claim to succeed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
