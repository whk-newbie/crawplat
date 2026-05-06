// Package repo 定义调度服务（scheduler-service）的数据访问层（Repository）。
//
// 该文件负责：
//   - 实现 PostgresRepository，封装对 PostgreSQL scheduled_tasks 表的所有 SQL 操作。
//   - Create: 插入一条新的调度记录，command 字段序列化为 JSONB 存储。
//   - List: 查询所有调度记录，按 created_at DESC, id DESC 排序。
//   - AdvanceLastMaterialized: CAS（Compare-And-Swap）更新 last_materialized_at 游标，
//     用于防止并发物化重复——只有当当前值等于 previous 时才更新为 next。
//   - RestoreLastMaterialized: 回滚游标到之前的值，用于执行创建失败时的补偿操作。
//
// 与谁交互：
//   - PostgreSQL（通过 database/sql）：直接执行 SQL 语句。
//   - model.Schedule：使用领域模型作为数据传输对象。
//
// 不负责：
//   - 不做 cron 表达式解析（由 service 层负责）。
//   - 不做物化循环逻辑（由 service 层负责）。
//   - 不做连接池管理（由 go-common/postgres 包负责）。
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

// NewPostgresRepository 创建 PostgresRepository 实例，传入已初始化的 *sql.DB 连接。
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create 插入一条 Schedule 记录到 scheduled_tasks 表。
// command 字段会被序列化为 JSONB 存储；last_materialized_at 在首次创建时为 NULL。
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
		INSERT INTO scheduled_tasks (id, project_id, organization_id, spider_id, spider_version, registry_auth_ref, name, cron_expr, enabled, image, command, retry_limit, retry_delay_seconds, created_at, last_materialized_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12, $13, $14, $15, $16)
	`, schedule.ID, schedule.ProjectID, nullableString(schedule.OrganizationID), schedule.SpiderID, nullableString(schedule.SpiderVersion), nullableString(schedule.RegistryAuthRef), schedule.Name, schedule.CronExpr, schedule.Enabled, schedule.Image, string(commandJSON), schedule.RetryLimit, schedule.RetryDelaySeconds, schedule.CreatedAt, lastMaterializedAt)
	return err
}

// List 分页查询 Schedule 记录，按创建时间降序排列。limit <= 0 时默认返回 20 条。
func (r *PostgresRepository) List(ctx context.Context, orgID string, limit, offset int) ([]model.Schedule, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, organization_id, spider_id, spider_version, registry_auth_ref, name, cron_expr, enabled, image, command, retry_limit, retry_delay_seconds, created_at, last_materialized_at
		FROM scheduled_tasks		WHERE ($1 = '' OR organization_id = $1)
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []model.Schedule
	for rows.Next() {
		var schedule model.Schedule
		var commandJSON string
		var spiderVersion, registryAuthRef sql.NullString
		var lastMaterializedAt sql.NullTime
		if err := rows.Scan(&schedule.ID, &schedule.ProjectID, &schedule.OrganizationID, &schedule.SpiderID, &spiderVersion, &registryAuthRef, &schedule.Name, &schedule.CronExpr, &schedule.Enabled, &schedule.Image, &commandJSON, &schedule.RetryLimit, &schedule.RetryDelaySeconds, &schedule.CreatedAt, &lastMaterializedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(commandJSON), &schedule.Command); err != nil {
			return nil, err
		}
		if spiderVersion.Valid {
			schedule.SpiderVersion = spiderVersion.String
		}
		if registryAuthRef.Valid {
			schedule.RegistryAuthRef = registryAuthRef.String
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

// AdvanceLastMaterialized 通过 CAS（Compare-And-Swap）原子推进 last_materialized_at 游标。
//
// 设计目的：在多实例部署或高并发场景下，防止同一时间点被重复物化。
// - previous 为 nil 时：仅当 last_materialized_at IS NULL（首次物化）才更新。
// - previous 非 nil 时：仅当 last_materialized_at = previous 时才更新为 next。
// 返回 true 表示 CAS 成功（rowsAffected == 1），即调用者获得了该时间点的物化权。
// 返回 false 表示 CAS 失败（被其他实例抢先或游标已变更），调用者应跳过该时间点。
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

// RestoreLastMaterialized 回滚 last_materialized_at 游标到之前的值。
//
// 使用场景：当 AdvanceLastMaterialized 成功后，后续的 executionClient.Create 失败时，
// 需要将游标回退，以便下次物化循环重试该时间点。
// - previous 为 nil 时：将游标从 current 回退为 NULL。
// - previous 非 nil 时：将游标从 current 回退为 previous。
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

// nullableString 将空字符串映射为 nil（数据库 NULL），非空字符串保持原值。
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
