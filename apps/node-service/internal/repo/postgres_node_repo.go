// 文件职责：基于 PostgreSQL 的节点仓库实现（实现 service.Repository + service.CatalogRepository 接口）。
// 负责：
//   1. 节点目录管理（nodes 表）：持久化节点元信息，不依赖 TTL，支持按 ID 查询。
//   2. 心跳历史记录（node_heartbeats 表）：每次心跳写入一条记录，支持按时间倒序查询历史。
//   3. 执行记录过滤查询（executions 表）：按节点 id、状态、时间范围、分页参数动态构建查询。
// 与谁交互：依赖 database/sql 和 PostgreSQL，被 main.go 注入到 NodeService。
// 与 Redis 仓库的区别：
//   - Redis 仓库通过 TTL 自动过期实现"在线/离线"判定（过期即离线）。
//   - Postgres 仓库持久化节点目录，不自动判断在线/离线（ListOnline 返回全部目录节点）。
//   - Postgres 仓库额外支持：节点详情查询、心跳历史、执行记录过滤。
// 不负责：API 路由处理、业务逻辑（由 api、service 包负责）、Redis 存储逻辑。
package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"crawler-platform/apps/node-service/internal/service"
)

// PostgresNodeRepository 基于 PostgreSQL 持久化节点数据。
// 涉及三张表：nodes（节点目录）、node_heartbeats（心跳历史）、executions（执行记录）。
type PostgresNodeRepository struct {
	db *sql.DB
}

// NewPostgresNodeRepository 创建 PostgreSQL 仓库实例。
func NewPostgresNodeRepository(db *sql.DB) *PostgresNodeRepository {
	return &PostgresNodeRepository{db: db}
}

// UpsertHeartbeat 满足 service.Repository 接口的心跳处理入口。
// 内部调用 UpsertCatalog，使用当前 UTC 时间作为心跳时间。
func (r *PostgresNodeRepository) UpsertHeartbeat(ctx context.Context, name string, capabilities []string) (service.Node, error) {
	return r.UpsertCatalog(ctx, name, capabilities, time.Now().UTC())
}

// UpsertCatalog 写入节点目录和心跳历史。
// 在一个非事务性操作中执行两次写入：
//   1. UPSERT INTO nodes —— 插入或更新节点目录（id 为主键冲突时更新 name/capabilities/last_seen_at/updated_at）。
//      注意：此表不返回 status 字段，因为 Postgres 仓库不自动计算在线状态，status 需在上层计算。
//   2. INSERT INTO node_heartbeats —— 追记一条心跳历史记录（每次调用都会新增一行）。
// 返回的 Node 快照不含 Status（零值），上层需根据业务逻辑附加状态。
func (r *PostgresNodeRepository) UpsertCatalog(ctx context.Context, name string, capabilities []string, seenAt time.Time) (service.Node, error) {
	capabilitiesJSON, err := json.Marshal(capabilities)
	if err != nil {
		return service.Node{}, err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO nodes (id, name, capabilities_json, last_seen_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    capabilities_json = EXCLUDED.capabilities_json,
		    last_seen_at = EXCLUDED.last_seen_at,
		    updated_at = NOW()
	`, name, name, string(capabilitiesJSON), seenAt)
	if err != nil {
		return service.Node{}, err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO node_heartbeats (node_id, seen_at, capabilities_json)
		VALUES ($1, $2, $3)
	`, name, seenAt, string(capabilitiesJSON))
	if err != nil {
		return service.Node{}, err
	}

	return service.Node{
		ID:           name,
		Name:         name,
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   seenAt,
	}, nil
}

// ListCatalog 返回 nodes 表中所有节点的完整列表（按 name 升序排列）。
// 返回的 Node 不含 Status 字段（零值），因为 Postgres 仓库不负责在线判定。
func (r *PostgresNodeRepository) ListCatalog(ctx context.Context) ([]service.Node, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, capabilities_json, last_seen_at
		FROM nodes
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]service.Node, 0)
	for rows.Next() {
		var (
			node           service.Node
			capabilitiesRaw []byte
		)
		if err := rows.Scan(&node.ID, &node.Name, &capabilitiesRaw, &node.LastSeenAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(capabilitiesRaw, &node.Capabilities); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

// ListOnline 满足 service.Repository 接口，直接委托给 ListCatalog。
// Postgres 模式不自动判断在线/离线 —— 返回所有节点目录中的记录。
// 在线判定需由上层根据 last_seen_at 与当前时间的差值来实现。
func (r *PostgresNodeRepository) ListOnline(ctx context.Context) ([]service.Node, error) {
	return r.ListCatalog(ctx)
}

// GetByID 根据节点 ID 查找节点详情。
// 用于 GET /api/v1/nodes/:id 端点。
// 如果节点在 nodes 表中不存在，返回 service.ErrNodeNotFound。
func (r *PostgresNodeRepository) GetByID(ctx context.Context, nodeID string) (service.Node, error) {
	var (
		node            service.Node
		capabilitiesRaw []byte
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, capabilities_json, last_seen_at
		FROM nodes
		WHERE id = $1
	`, nodeID).Scan(&node.ID, &node.Name, &capabilitiesRaw, &node.LastSeenAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.Node{}, service.ErrNodeNotFound
		}
		return service.Node{}, err
	}
	if err := json.Unmarshal(capabilitiesRaw, &node.Capabilities); err != nil {
		return service.Node{}, err
	}
	return node, nil
}

// ListHeartbeatHistory 查询指定节点的最近心跳历史记录。
// 按 seen_at 降序排列，limit 参数控制返回条数。
// 用于节点详情页绘制心跳时间序列图，或用于会话切分计算（通过心跳间隔 gapSeconds 划分 session）。
// 每次心跳上报都会在此表新增一条记录，所以历史数据持续增长。
func (r *PostgresNodeRepository) ListHeartbeatHistory(ctx context.Context, nodeID string, limit int) ([]service.NodeHeartbeat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT seen_at, capabilities_json
		FROM node_heartbeats
		WHERE node_id = $1
		ORDER BY seen_at DESC
		LIMIT $2
	`, nodeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]service.NodeHeartbeat, 0)
	for rows.Next() {
		var (
			heartbeat       service.NodeHeartbeat
			capabilitiesRaw []byte
		)
		if err := rows.Scan(&heartbeat.SeenAt, &capabilitiesRaw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(capabilitiesRaw, &heartbeat.Capabilities); err != nil {
			return nil, err
		}
		history = append(history, heartbeat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return history, nil
}

// ListRecentExecutions 查询指定节点上的近期执行记录，支持多重过滤。
// 过滤条件由 ExecutionQuery 提供，均为可选：
//   - Status：按执行状态精确匹配（如 "succeeded"、"failed"）。
//   - From/To：按 created_at 时间范围过滤（闭区间）。
//   - Limit/Offset：分页参数。
// 动态构建 SQL WHERE 子句：只添加 Query 中非空的过滤条件，利用 $1, $2, ... 占位符避免注入。
// started_at/finished_at 可能为 NULL（任务尚未开始或尚未结束），用 sql.NullTime 处理后再转为 *time.Time。
// 结果按 created_at DESC 排序，返回最近完成的在前。
func (r *PostgresNodeRepository) ListRecentExecutions(ctx context.Context, nodeID string, query service.ExecutionQuery) ([]service.NodeExecution, error) {
	args := []any{nodeID}
	where := []string{"node_id = $1"}
	argPos := 2
	if strings.TrimSpace(query.Status) != "" {
		where = append(where, fmt.Sprintf("status = $%d", argPos))
		args = append(args, query.Status)
		argPos++
	}
	if query.From != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", argPos))
		args = append(args, *query.From)
		argPos++
	}
	if query.To != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", argPos))
		args = append(args, *query.To)
		argPos++
	}

	sqlQuery := fmt.Sprintf(`
		SELECT id, project_id, spider_id, status, trigger_source, created_at, started_at, finished_at
		FROM executions
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d
		OFFSET $%d
	`, strings.Join(where, " AND "), argPos, argPos+1)
	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executions := make([]service.NodeExecution, 0)
	for rows.Next() {
		var (
			exec       service.NodeExecution
			startedAt  sql.NullTime
			finishedAt sql.NullTime
		)
		if err := rows.Scan(
			&exec.ID,
			&exec.ProjectID,
			&exec.SpiderID,
			&exec.Status,
			&exec.TriggerSource,
			&exec.CreatedAt,
			&startedAt,
			&finishedAt,
		); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			t := startedAt.Time
			exec.StartedAt = &t
		}
		if finishedAt.Valid {
			t := finishedAt.Time
			exec.FinishedAt = &t
		}
		executions = append(executions, exec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return executions, nil
}
