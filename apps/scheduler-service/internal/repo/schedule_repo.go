package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"crawler-platform/apps/scheduler-service/internal/model"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, schedule model.Schedule) error {
	commandJSON, err := json.Marshal(schedule.Command)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO scheduled_tasks (id, project_id, spider_id, name, cron_expr, enabled, image, command)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
	`, schedule.ID, schedule.ProjectID, schedule.SpiderID, schedule.Name, schedule.CronExpr, schedule.Enabled, schedule.Image, string(commandJSON))
	return err
}

func (r *PostgresRepository) List(ctx context.Context) ([]model.Schedule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, spider_id, name, cron_expr, enabled, image, command
		FROM scheduled_tasks
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []model.Schedule
	for rows.Next() {
		var schedule model.Schedule
		var commandJSON string
		if err := rows.Scan(&schedule.ID, &schedule.ProjectID, &schedule.SpiderID, &schedule.Name, &schedule.CronExpr, &schedule.Enabled, &schedule.Image, &commandJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(commandJSON), &schedule.Command); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return schedules, nil
}
