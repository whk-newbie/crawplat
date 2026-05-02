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
		INSERT INTO executions (id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, started_at, finished_at, error_message, retried_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`, exec.ID, exec.ProjectID, exec.SpiderID, exec.SpiderVersion, nullableString(exec.RegistryAuthRef), nullableString(exec.NodeID), exec.Status, exec.TriggerSource, exec.Image, string(command), exec.CPUCores, exec.MemoryMB, exec.TimeoutSeconds, exec.RetryLimit, exec.RetryCount, exec.RetryDelaySeconds, nullableString(exec.RetryOfExecutionID), exec.StartedAt, exec.FinishedAt, nullableString(exec.ErrorMessage), exec.RetriedAt)
	if err != nil {
		return model.Execution{}, err
	}

	return exec, nil
}

func (r *ExecutionRepository) Get(ctx context.Context, id string) (model.Execution, error) {
	var exec model.Execution
	var (
		nodeID             sql.NullString
		registryAuthRef    sql.NullString
		commandRaw         []byte
		errorMessage       sql.NullString
		retryOfExecutionID sql.NullString
		startedAt          sql.NullTime
		finishedAt         sql.NullTime
		retriedAt          sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, node_id, status, trigger_source, image, command, cpu_cores, memory_mb, timeout_seconds, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
		FROM executions
		WHERE id = $1
	`, id).Scan(
		&exec.ID,
		&exec.ProjectID,
		&exec.SpiderID,
		&exec.SpiderVersion,
		&registryAuthRef,
		&nodeID,
		&exec.Status,
		&exec.TriggerSource,
		&exec.Image,
		&commandRaw,
		&exec.CPUCores,
		&exec.MemoryMB,
		&exec.TimeoutSeconds,
		&errorMessage,
		&exec.CreatedAt,
		&startedAt,
		&finishedAt,
		&exec.RetryLimit,
		&exec.RetryCount,
		&exec.RetryDelaySeconds,
		&retryOfExecutionID,
		&retriedAt,
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
	if registryAuthRef.Valid {
		exec.RegistryAuthRef = registryAuthRef.String
	}
	if errorMessage.Valid {
		exec.ErrorMessage = errorMessage.String
	}
	if retryOfExecutionID.Valid {
		exec.RetryOfExecutionID = retryOfExecutionID.String
	}
	if startedAt.Valid {
		t := startedAt.Time
		exec.StartedAt = &t
	}
	if finishedAt.Valid {
		t := finishedAt.Time
		exec.FinishedAt = &t
	}
	if retriedAt.Valid {
		t := retriedAt.Time
		exec.RetriedAt = &t
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

func (r *ExecutionRepository) ClaimNextRetryCandidate(ctx context.Context, now time.Time) (model.Execution, bool, error) {
	var exec model.Execution
	var (
		nodeID             sql.NullString
		registryAuthRef    sql.NullString
		commandRaw         []byte
		errorMessage       sql.NullString
		startedAt          sql.NullTime
		finishedAt         sql.NullTime
		retryOfExecutionID sql.NullString
		retriedAt          sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, `
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
	`, now, now).Scan(
		&exec.ID,
		&exec.ProjectID,
		&exec.SpiderID,
		&exec.SpiderVersion,
		&registryAuthRef,
		&nodeID,
		&exec.Status,
		&exec.TriggerSource,
		&exec.Image,
		&commandRaw,
		&exec.CPUCores,
		&exec.MemoryMB,
		&exec.TimeoutSeconds,
		&errorMessage,
		&exec.CreatedAt,
		&startedAt,
		&finishedAt,
		&exec.RetryLimit,
		&exec.RetryCount,
		&exec.RetryDelaySeconds,
		&retryOfExecutionID,
		&retriedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Execution{}, false, nil
		}
		return model.Execution{}, false, err
	}

	if nodeID.Valid {
		exec.NodeID = nodeID.String
	}
	if registryAuthRef.Valid {
		exec.RegistryAuthRef = registryAuthRef.String
	}
	if errorMessage.Valid {
		exec.ErrorMessage = errorMessage.String
	}
	if retryOfExecutionID.Valid {
		exec.RetryOfExecutionID = retryOfExecutionID.String
	}
	if startedAt.Valid {
		t := startedAt.Time
		exec.StartedAt = &t
	}
	if finishedAt.Valid {
		t := finishedAt.Time
		exec.FinishedAt = &t
	}
	if retriedAt.Valid {
		t := retriedAt.Time
		exec.RetriedAt = &t
	}
	if err := json.Unmarshal(commandRaw, &exec.Command); err != nil {
		return model.Execution{}, false, err
	}

	return exec, true, nil
}

func (r *ExecutionRepository) ResetRetryClaim(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE executions SET retried_at = NULL WHERE id = $1`, id)
	return err
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
