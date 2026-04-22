package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"crawler-platform/apps/execution-service/internal/model"
	"crawler-platform/apps/execution-service/internal/service"
)

type ExecutionRepository struct {
	db *sql.DB
}

func NewExecutionRepository(db *sql.DB) *ExecutionRepository {
	return &ExecutionRepository{db: db}
}

func (r *ExecutionRepository) Create(ctx context.Context, exec model.Execution) (model.Execution, error) {
	command, err := json.Marshal(exec.Command)
	if err != nil {
		return model.Execution{}, err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO executions (id, project_id, spider_id, node_id, status, trigger_source, image, command, started_at, finished_at, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, exec.ID, exec.ProjectID, exec.SpiderID, nullableString(exec.NodeID), exec.Status, exec.TriggerSource, exec.Image, string(command), exec.StartedAt, exec.FinishedAt, nullableString(exec.ErrorMessage))
	if err != nil {
		return model.Execution{}, err
	}

	return exec, nil
}

func (r *ExecutionRepository) Get(ctx context.Context, id string) (model.Execution, error) {
	var exec model.Execution
	var (
		nodeID       sql.NullString
		commandRaw   []byte
		errorMessage sql.NullString
		startedAt    sql.NullTime
		finishedAt   sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, spider_id, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at
		FROM executions
		WHERE id = $1
	`, id).Scan(
		&exec.ID,
		&exec.ProjectID,
		&exec.SpiderID,
		&nodeID,
		&exec.Status,
		&exec.TriggerSource,
		&exec.Image,
		&commandRaw,
		&errorMessage,
		&exec.CreatedAt,
		&startedAt,
		&finishedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Execution{}, service.ErrExecutionNotFound
		}
		return model.Execution{}, err
	}

	if nodeID.Valid {
		exec.NodeID = nodeID.String
	}
	if errorMessage.Valid {
		exec.ErrorMessage = errorMessage.String
	}
	if startedAt.Valid {
		t := startedAt.Time
		exec.StartedAt = &t
	}
	if finishedAt.Valid {
		t := finishedAt.Time
		exec.FinishedAt = &t
	}
	if err := json.Unmarshal(commandRaw, &exec.Command); err != nil {
		return model.Execution{}, err
	}

	return exec, nil
}

func (r *ExecutionRepository) MarkRunning(ctx context.Context, id, nodeID string, startedAt time.Time) (model.Execution, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE executions
		SET node_id = $2, status = 'running', started_at = $3, finished_at = NULL, error_message = NULL
		WHERE id = $1 AND status = 'pending'
	`, id, nodeID, startedAt)
	if err != nil {
		return model.Execution{}, err
	}
	if err := r.ensureTransitionRowsAffected(ctx, result, id); err != nil {
		return model.Execution{}, err
	}
	return r.Get(ctx, id)
}

func (r *ExecutionRepository) Complete(ctx context.Context, id string, finishedAt time.Time) (model.Execution, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE executions
		SET status = 'succeeded', finished_at = $2, error_message = NULL
		WHERE id = $1 AND status = 'running'
	`, id, finishedAt)
	if err != nil {
		return model.Execution{}, err
	}
	if err := r.ensureTransitionRowsAffected(ctx, result, id); err != nil {
		return model.Execution{}, err
	}
	return r.Get(ctx, id)
}

func (r *ExecutionRepository) Fail(ctx context.Context, id, errorMessage string, finishedAt time.Time) (model.Execution, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE executions
		SET status = 'failed', finished_at = $2, error_message = $3
		WHERE id = $1 AND status = 'running'
	`, id, finishedAt, errorMessage)
	if err != nil {
		return model.Execution{}, err
	}
	if err := r.ensureTransitionRowsAffected(ctx, result, id); err != nil {
		return model.Execution{}, err
	}
	return r.Get(ctx, id)
}

func (r *ExecutionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM executions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func ensureRowsAffected(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return service.ErrExecutionNotFound
	}
	return nil
}

func (r *ExecutionRepository) ensureTransitionRowsAffected(ctx context.Context, result sql.Result, executionID string) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected > 0 {
		return nil
	}

	if _, err := r.Get(ctx, executionID); err != nil {
		if errors.Is(err, service.ErrExecutionNotFound) {
			return service.ErrExecutionNotFound
		}
		return err
	}

	return service.ErrInvalidExecutionState
}
