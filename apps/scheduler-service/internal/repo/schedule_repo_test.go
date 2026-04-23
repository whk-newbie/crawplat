package repo

import (
	"context"
	"regexp"
	"testing"

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
		ID:        "sched-1",
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Name:      "nightly",
		CronExpr:  "0 * * * *",
		Enabled:   true,
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO scheduled_tasks (id, project_id, spider_id, name, cron_expr, enabled, image, command)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
	`)).
		WithArgs(schedule.ID, schedule.ProjectID, schedule.SpiderID, schedule.Name, schedule.CronExpr, schedule.Enabled, schedule.Image, `["./go-echo"]`).
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
	rows := sqlmock.NewRows([]string{"id", "project_id", "spider_id", "name", "cron_expr", "enabled", "image", "command"}).
		AddRow("sched-1", "project-1", "spider-1", "nightly", "0 * * * *", true, "crawler/go-echo:latest", `["./go-echo"]`)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, spider_id, name, cron_expr, enabled, image, command
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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
