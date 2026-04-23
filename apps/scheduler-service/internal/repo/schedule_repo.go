package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

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

	var lastMaterializedAt any
	if schedule.LastMaterializedAt != nil {
		lastMaterializedAt = *schedule.LastMaterializedAt
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO scheduled_tasks (id, project_id, spider_id, name, cron_expr, enabled, image, command, created_at, last_materialized_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10)
	`, schedule.ID, schedule.ProjectID, schedule.SpiderID, schedule.Name, schedule.CronExpr, schedule.Enabled, schedule.Image, string(commandJSON), schedule.CreatedAt, lastMaterializedAt)
	return err
}

func (r *PostgresRepository) List(ctx context.Context) ([]model.Schedule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, spider_id, name, cron_expr, enabled, image, command, created_at, last_materialized_at
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
		var lastMaterializedAt sql.NullTime
		if err := rows.Scan(&schedule.ID, &schedule.ProjectID, &schedule.SpiderID, &schedule.Name, &schedule.CronExpr, &schedule.Enabled, &schedule.Image, &commandJSON, &schedule.CreatedAt, &lastMaterializedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(commandJSON), &schedule.Command); err != nil {
			return nil, err
		}
		if lastMaterializedAt.Valid {
			value := lastMaterializedAt.Time.UTC()
			schedule.LastMaterializedAt = &value
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *PostgresRepository) AdvanceLastMaterialized(ctx context.Context, id string, previous *time.Time, next time.Time) (bool, error) {
	var result sql.Result
	var err error
	if previous == nil {
		result, err = r.db.ExecContext(ctx, `
			UPDATE scheduled_tasks
			SET last_materialized_at = $2
			WHERE id = $1 AND last_materialized_at IS NULL
		`, id, next)
	} else {
		result, err = r.db.ExecContext(ctx, `
			UPDATE scheduled_tasks
			SET last_materialized_at = $3
			WHERE id = $1 AND last_materialized_at = $2
		`, id, *previous, next)
	}
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}

func (r *PostgresRepository) RestoreLastMaterialized(ctx context.Context, id string, previous *time.Time, current time.Time) error {
	if previous == nil {
		_, err := r.db.ExecContext(ctx, `
			UPDATE scheduled_tasks
			SET last_materialized_at = NULL
			WHERE id = $1 AND last_materialized_at = $2
		`, id, current)
		return err
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE scheduled_tasks
		SET last_materialized_at = $3
		WHERE id = $1 AND last_materialized_at = $2
	`, id, current, *previous)
	return err
}
