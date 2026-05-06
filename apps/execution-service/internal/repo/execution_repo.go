// 执行数据仓库（PostgreSQL）。
// 封装 executions 表的 CRUD 操作和状态转换，使用条件 UPDATE（WHERE status = 'expected'）保证状态机正确性。
// 状态转换失败时通过 ensureTransitionRowsAffected 区分「执行不存在」与「状态不匹配」两种错误。
// 不管理日志存储（MongoDB log_repo）和执行队列（Redis queue）。
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

// ExecutionRepository 封装了对 PostgreSQL executions 表的所有操作。
type ExecutionRepository struct {
	db *sql.DB
}

// NewExecutionRepository 创建仓库实例。db 应为已验证连接的 *sql.DB 对象。
func NewExecutionRepository(db *sql.DB) *ExecutionRepository {
	return &ExecutionRepository{db: db}
}

// Create 将一个新执行记录插入 executions 表。
// Command 字段以 JSON 数组形式存储（如 ["./crawler"]），通过 json.Marshal 序列化。
// 返回传入的 exec 对象（不含数据库生成的字段，因为 ID 已由调用方生成）。
func (r *ExecutionRepository) Create(ctx context.Context, exec model.Execution) (model.Execution, error) {
	command, err := json.Marshal(exec.Command)
	if err != nil {
		return model.Execution{}, err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO executions (id, project_id, spider_id, spider_version, registry_auth_ref, cpu_cores, memory_mb, timeout_seconds, node_id, status, trigger_source, image, command, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, organization_id, started_at, finished_at, error_message, retried_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`, exec.ID, exec.ProjectID, exec.SpiderID, nullableString(exec.SpiderVersion), nullableString(exec.RegistryAuthRef), exec.CpuCores, exec.MemoryMB, exec.TimeoutSeconds, nullableString(exec.NodeID), exec.Status, exec.TriggerSource, exec.Image, string(command), exec.RetryLimit, exec.RetryCount, exec.RetryDelaySeconds, nullableString(exec.RetryOfExecutionID), nullableString(exec.OrganizationID), exec.StartedAt, exec.FinishedAt, nullableString(exec.ErrorMessage), exec.RetriedAt)
	if err != nil {
		return model.Execution{}, err
	}

	return exec, nil
}

// Get 根据 ID 查询单条执行记录。
// 返回值：找不到时返回 service.ErrExecutionNotFound；数据库错误时返回原始 error。
// 可空字段（NodeID、ErrorMessage、StartedAt 等）通过 sql.Null* 类型安全处理。
func (r *ExecutionRepository) Get(ctx context.Context, id string) (model.Execution, error) {
	var exec model.Execution
	var (
		nodeID            sql.NullString
		spiderVersion     sql.NullString
		registryAuthRef   sql.NullString
		commandRaw        []byte
		errorMessage      sql.NullString
		retryOfExecutionID sql.NullString
		startedAt         sql.NullTime
		finishedAt        sql.NullTime
		retriedAt         sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, organization_id, spider_id, spider_version, registry_auth_ref, cpu_cores, memory_mb, timeout_seconds, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
		FROM executions
		WHERE id = $1
	`, id).Scan(
		&exec.ID,
		&exec.ProjectID,
			&exec.OrganizationID,
		&exec.SpiderID,
		&spiderVersion,
		&registryAuthRef,
		&exec.CpuCores,
		&exec.MemoryMB,
		&exec.TimeoutSeconds,
		&nodeID,
		&exec.Status,
		&exec.TriggerSource,
		&exec.Image,
		&commandRaw,
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
	if spiderVersion.Valid {
		exec.SpiderVersion = spiderVersion.String
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

// List 分页查询执行记录，支持按 status 过滤。
// limit 和 offset 控制分页；status 为空时返回所有状态。
func (r *ExecutionRepository) List(ctx context.Context, orgID string, limit, offset int, status string) (*service.ListResult, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM executions`
	args := []any{}
	countQuery += ` WHERE ($1 = '' OR organization_id = $1)`
	args = append(args, orgID)
	if status != "" {
		countQuery += ` AND status = $2`
		args = append(args, status)
	}
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	query := `
		SELECT id, project_id, spider_id, spider_version, registry_auth_ref, cpu_cores, memory_mb, timeout_seconds, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
		FROM executions
		WHERE ($1 = '' OR organization_id = $1)
	`
	queryArgs := []any{orgID}
	if status != "" {
		query += ` AND status = $2`
		queryArgs = append(queryArgs, status)
	}
	query += ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []model.Execution
	for rows.Next() {
		var exec model.Execution
		var (
			nodeID             sql.NullString
			spiderVersion      sql.NullString
			registryAuthRef    sql.NullString
			commandRaw         []byte
			errorMessage       sql.NullString
			retryOfExecutionID sql.NullString
			startedAt          sql.NullTime
			finishedAt         sql.NullTime
			retriedAt          sql.NullTime
		)
		if err := rows.Scan(
			&exec.ID, &exec.ProjectID, &exec.SpiderID, &spiderVersion, &registryAuthRef,
			&exec.CpuCores, &exec.MemoryMB, &exec.TimeoutSeconds, &nodeID,
			&exec.Status, &exec.TriggerSource, &exec.Image, &commandRaw,
			&errorMessage, &exec.CreatedAt, &startedAt, &finishedAt,
			&exec.RetryLimit, &exec.RetryCount, &exec.RetryDelaySeconds,
			&retryOfExecutionID, &retriedAt,
		); err != nil {
			return nil, err
		}
		if nodeID.Valid {
			exec.NodeID = nodeID.String
		}
		if spiderVersion.Valid {
			exec.SpiderVersion = spiderVersion.String
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
			return nil, err
		}
		executions = append(executions, exec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &service.ListResult{Executions: executions, Total: total}, nil
}

// MarkRunning 将执行状态从 pending 转换为 running。
// 关键：使用条件 UPDATE (WHERE status = 'pending') 实现乐观并发控制——只有当前状态为 pending 时才允许转换。
// 同时写入 nodeID 和 startedAt，并清空 finished_at 和 error_message（如果这是重试执行）。
// 如果 update 影响行数为 0：
//   - 执行不存在 → 返回 ErrExecutionNotFound
//   - 执行状态不是 pending → 返回 ErrInvalidExecutionState
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

// Complete 将执行状态从 running 转换为 succeeded，写入 finished_at 并清空 error_message。
// 同样使用条件 UPDATE (WHERE status = 'running') 保证状态转换原子性。
// 转换失败时通过 ensureTransitionRowsAffected 返回 ErrInvalidExecutionState 或 ErrExecutionNotFound。
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

// Fail 将执行状态从 running 转换为 failed，写入 finished_at 和 error_message。
// errorMessage 记录了失败原因，会在重试查询时作为候选筛选依据（通过 error_message IS NOT NULL 间接筛选）。
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

// Delete 从 executions 表中删除指定记录。
// 仅由 rollbackCreate 在创建流程失败时调用，用于清理已写入 PostgreSQL 但后续步骤失败的部分创建数据。
func (r *ExecutionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM executions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

// ClaimNextRetryCandidate 原子地查找并标记下一个重试候选执行。
//
// 筛选条件（在内层子查询中）：
//   - status = 'failed'：已失败
//   - retried_at IS NULL：尚未被重试（乐观锁）
//   - retry_limit > retry_count：仍有重试配额
//   - finished_at IS NOT NULL：已记录失败时间
//   - finished_at + retry_delay_seconds <= now：延迟已过
// 排序：按 finished_at 升序（最早失败的优先）、id 升序（确定性排序）。
//
// 外层 UPDATE 原子地设置 retried_at = now，防止并发重试同一执行。
// 使用 UPDATE ... RETURNING 在单次查询中完成锁定和读取，避免 SELECT FOR UPDATE 导致的额外锁争用。
//
// 返回值：(执行记录, true, nil) 表示找到候选；(零值, false, nil) 表示没有符合条件的候选。
func (r *ExecutionRepository) ClaimNextRetryCandidate(ctx context.Context, now time.Time) (model.Execution, bool, error) {
	var exec model.Execution
	var (
		nodeID            sql.NullString
		spiderVersion     sql.NullString
		registryAuthRef   sql.NullString
		commandRaw        []byte
		errorMessage      sql.NullString
		startedAt         sql.NullTime
		finishedAt        sql.NullTime
		retryOfExecutionID sql.NullString
		retriedAt         sql.NullTime
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
		RETURNING id, project_id, organization_id, spider_id, spider_version, registry_auth_ref, cpu_cores, memory_mb, timeout_seconds, node_id, status, trigger_source, image, command, error_message, created_at, started_at, finished_at, retry_limit, retry_count, retry_delay_seconds, retry_of_execution_id, retried_at
	`, now, now).Scan(
		&exec.ID,
		&exec.ProjectID,
			&exec.OrganizationID,
		&exec.SpiderID,
		&spiderVersion,
		&registryAuthRef,
		&exec.CpuCores,
		&exec.MemoryMB,
		&exec.TimeoutSeconds,
		&nodeID,
		&exec.Status,
		&exec.TriggerSource,
		&exec.Image,
		&commandRaw,
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
	if spiderVersion.Valid {
		exec.SpiderVersion = spiderVersion.String
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

// ResetRetryClaim 将执行记录的 retried_at 重置为 NULL，撤销 ClaimNextRetryCandidate 的乐观锁标记。
// 调用场景：ClaimNextRetryCandidate 成功后，Create 新重试执行失败时，需要回滚乐观锁以便下次重试周期再次尝试。
func (r *ExecutionRepository) ResetRetryClaim(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE executions SET retried_at = NULL WHERE id = $1`, id)
	return err
}

// nullableString 将空字符串映射为 nil（数据库 NULL），非空字符串保持原值。
// 用于处理 Optional 字段（如 NodeID、Error、RetryOfExecutionID），这些字段在数据库中可为 NULL。
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

// ensureRowsAffected 检查 SQL 操作是否影响了至少一行。不影响任何行时返回 ErrExecutionNotFound。
// 用于 Delete 等操作——如果 ID 不存在，应返回 404 而非静默成功。
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

// ensureTransitionRowsAffected 检查状态转换 UPDATE 的结果并返回合适的错误类型。
//
// 逻辑：
//  1. 如果影响行数 > 0 → 转换成功，返回 nil
//  2. 如果影响行数 = 0 → 进一步查询执行是否存在：
//     a. 记录不存在 → 返回 ErrExecutionNotFound（执行已被删除或 ID 无效）
//     b. 记录存在但状态不匹配 → 返回 ErrInvalidExecutionState（例如尝试 complete 一个 pending 执行）
//
// 为什么这样设计：条件 UPDATE（WHERE status = 'expected'）可能因两种原因影响 0 行，
// 通过二次查询区分这两种情况，让调用方可以采取不同的错误处理策略。
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
