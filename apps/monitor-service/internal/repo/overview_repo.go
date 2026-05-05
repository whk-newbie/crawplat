// Package repo 是 Monitor 服务的持久层，直接访问 PostgreSQL 和 Redis。
// 负责：执行/节点聚合查询、告警规则 CRUD、告警事件持久化、失败执行/离线节点轮询。
// 不包含告警评估逻辑——该职责属于 service/alerting。
package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"crawler-platform/apps/monitor-service/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	nodeKeyPrefix = "nodes:"
	nodeIndexKey  = "nodes:online"
)

// OverviewRepository 通过 PostgreSQL 聚合执行/节点统计，通过 Redis 判定在线节点。
// 同时管理告警规则和告警事件的持久化。
type OverviewRepository struct {
	db     *sql.DB
	client *redis.Client
}

func NewOverviewRepository(db *sql.DB, client *redis.Client) *OverviewRepository {
	return &OverviewRepository{db: db, client: client}
}

// Overview 返回聚合后的执行和节点统计概览。
func (r *OverviewRepository) Overview(ctx context.Context) (model.Overview, error) {
	executions, err := r.executionSummary(ctx)
	if err != nil {
		return model.Overview{}, err
	}

	nodes, err := r.nodeSummary(ctx)
	if err != nil {
		return model.Overview{}, err
	}

	return model.Overview{
		Executions: executions,
		Nodes:      nodes,
	}, nil
}

func (r *OverviewRepository) executionSummary(ctx context.Context) (model.ExecutionSummary, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM executions GROUP BY status`)
	if err != nil {
		return model.ExecutionSummary{}, err
	}
	defer rows.Close()

	var summary model.ExecutionSummary
	for rows.Next() {
		var (
			status string
			count  int
		)
		if err := rows.Scan(&status, &count); err != nil {
			return model.ExecutionSummary{}, err
		}

		summary.Total += count
		switch status {
		case "pending":
			summary.Pending = count
		case "running":
			summary.Running = count
		case "succeeded":
			summary.Succeeded = count
		case "failed":
			summary.Failed = count
		}
	}

	if err := rows.Err(); err != nil {
		return model.ExecutionSummary{}, err
	}

	return summary, nil
}

func (r *OverviewRepository) nodeSummary(ctx context.Context) (model.NodeSummary, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes`).Scan(&total); err != nil {
		return model.NodeSummary{}, err
	}

	online, err := r.countOnlineNodes(ctx)
	if err != nil {
		return model.NodeSummary{}, err
	}

	offline := total - online
	if offline < 0 {
		offline = 0
	}

	return model.NodeSummary{
		Total:   total,
		Online:  online,
		Offline: offline,
	}, nil
}

// countOnlineNodes 通过 Redis SMEMBERS + GET 判定在线节点数。
// 惰性清理：若节点键已过期但仍在 SET 中，则从 SET 中移除，避免离线计数漂移。
func (r *OverviewRepository) countOnlineNodes(ctx context.Context) (int, error) {
	nodeIDs, err := r.client.SMembers(ctx, nodeIndexKey).Result()
	if err != nil {
		return 0, err
	}

	online := 0
	for _, nodeID := range nodeIDs {
		_, err := r.client.Get(ctx, nodeKeyPrefix+nodeID).Result()
		if err == redis.Nil {
			_ = r.client.SRem(ctx, nodeIndexKey, nodeID).Err()
			continue
		}
		if err != nil {
			return 0, err
		}
		online++
	}
	return online, nil
}

func (r *OverviewRepository) CreateAlertRule(ctx context.Context, rule model.AlertRule) (model.AlertRule, error) {
	var created model.AlertRule
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO alert_rules (
			id, name, rule_type, enabled, webhook_url, cooldown_seconds, timeout_seconds, offline_grace_seconds, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, name, rule_type, enabled, webhook_url, cooldown_seconds, timeout_seconds, offline_grace_seconds, created_at, updated_at
	`,
		rule.ID, rule.Name, rule.RuleType, rule.Enabled, rule.WebhookURL, rule.CooldownSeconds,
		rule.TimeoutSeconds, rule.OfflineGraceSeconds, rule.CreatedAt, rule.UpdatedAt,
	).Scan(
		&created.ID,
		&created.Name,
		&created.RuleType,
		&created.Enabled,
		&created.WebhookURL,
		&created.CooldownSeconds,
		&created.TimeoutSeconds,
		&created.OfflineGraceSeconds,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return model.AlertRule{}, err
	}
	return created, nil
}

func (r *OverviewRepository) UpdateAlertRule(ctx context.Context, id string, patch model.AlertRulePatch) (model.AlertRule, bool, error) {
	setParts := make([]string, 0, 8)
	args := make([]any, 0, 10)
	argPos := 1

	if patch.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *patch.Name)
		argPos++
	}
	if patch.Enabled != nil {
		setParts = append(setParts, fmt.Sprintf("enabled = $%d", argPos))
		args = append(args, *patch.Enabled)
		argPos++
	}
	if patch.WebhookURL != nil {
		setParts = append(setParts, fmt.Sprintf("webhook_url = $%d", argPos))
		args = append(args, *patch.WebhookURL)
		argPos++
	}
	if patch.CooldownSeconds != nil {
		setParts = append(setParts, fmt.Sprintf("cooldown_seconds = $%d", argPos))
		args = append(args, *patch.CooldownSeconds)
		argPos++
	}
	if patch.TimeoutSeconds != nil {
		setParts = append(setParts, fmt.Sprintf("timeout_seconds = $%d", argPos))
		args = append(args, *patch.TimeoutSeconds)
		argPos++
	}
	if patch.OfflineGraceSeconds != nil {
		setParts = append(setParts, fmt.Sprintf("offline_grace_seconds = $%d", argPos))
		args = append(args, *patch.OfflineGraceSeconds)
		argPos++
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argPos))
	args = append(args, patch.UpdatedAt)
	argPos++

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE alert_rules
		SET %s
		WHERE id = $%d
		RETURNING id, name, rule_type, enabled, webhook_url, cooldown_seconds, timeout_seconds, offline_grace_seconds, created_at, updated_at
	`, strings.Join(setParts, ", "), argPos)

	var updated model.AlertRule
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&updated.ID,
		&updated.Name,
		&updated.RuleType,
		&updated.Enabled,
		&updated.WebhookURL,
		&updated.CooldownSeconds,
		&updated.TimeoutSeconds,
		&updated.OfflineGraceSeconds,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AlertRule{}, false, nil
		}
		return model.AlertRule{}, false, err
	}
	return updated, true, nil
}

func (r *OverviewRepository) ListAlertRules(ctx context.Context) ([]model.AlertRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, rule_type, enabled, webhook_url, cooldown_seconds, timeout_seconds, offline_grace_seconds, created_at, updated_at
		FROM alert_rules
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]model.AlertRule, 0)
	for rows.Next() {
		var rule model.AlertRule
		if err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.RuleType,
			&rule.Enabled,
			&rule.WebhookURL,
			&rule.CooldownSeconds,
			&rule.TimeoutSeconds,
			&rule.OfflineGraceSeconds,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *OverviewRepository) ListAlertEvents(ctx context.Context, limit, offset int) ([]model.AlertEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, rule_id, rule_type, entity_type, entity_id, dedupe_key, payload_json::text, delivery_status, COALESCE(webhook_status_code, 0), COALESCE(error_message, ''), created_at
		FROM alert_events
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]model.AlertEvent, 0)
	for rows.Next() {
		var event model.AlertEvent
		if err := rows.Scan(
			&event.ID,
			&event.RuleID,
			&event.RuleType,
			&event.EntityType,
			&event.EntityID,
			&event.DedupeKey,
			&event.Payload,
			&event.DeliveryStatus,
			&event.WebhookStatusCode,
			&event.ErrorMessage,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *OverviewRepository) CountAlertEvents(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alert_events`).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *OverviewRepository) ListFailedExecutionsSince(ctx context.Context, since time.Time, limit int) ([]model.FailedExecutionCandidate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, spider_id, COALESCE(error_message, ''), COALESCE(finished_at, created_at) AS occurred_at
		FROM executions
		WHERE status = 'failed' AND COALESCE(finished_at, created_at) >= $1
		ORDER BY occurred_at DESC
		LIMIT $2
	`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]model.FailedExecutionCandidate, 0)
	for rows.Next() {
		var candidate model.FailedExecutionCandidate
		if err := rows.Scan(
			&candidate.ExecutionID,
			&candidate.ProjectID,
			&candidate.SpiderID,
			&candidate.Error,
			&candidate.OccurredAt,
		); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return candidates, nil
}

func (r *OverviewRepository) ListOfflineNodes(ctx context.Context, before time.Time, limit int) ([]model.OfflineNodeCandidate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, last_seen_at
		FROM nodes
		WHERE last_seen_at <= $1
		ORDER BY last_seen_at ASC
		LIMIT $2
	`, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]model.OfflineNodeCandidate, 0)
	for rows.Next() {
		var candidate model.OfflineNodeCandidate
		if err := rows.Scan(&candidate.NodeID, &candidate.NodeName, &candidate.LastSeenAt); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return candidates, nil
}

func (r *OverviewRepository) LastAlertEventAt(ctx context.Context, ruleID, dedupeKey string) (*time.Time, error) {
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT created_at
		FROM alert_events
		WHERE rule_id = $1 AND dedupe_key = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, ruleID, dedupeKey).Scan(&createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &createdAt, nil
}

func (r *OverviewRepository) SaveAlertEvent(ctx context.Context, event model.AlertEventRecord) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO alert_events (
			id, rule_id, rule_type, entity_type, entity_id, dedupe_key, payload_json, delivery_status, webhook_status_code, error_message, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9,$10,$11)
	`,
		event.ID,
		event.RuleID,
		event.RuleType,
		event.EntityType,
		event.EntityID,
		event.DedupeKey,
		event.Payload,
		event.DeliveryStatus,
		nullableInt(event.WebhookStatusCode),
		nullableString(event.ErrorMessage),
		event.CreatedAt,
	)
	return err
}

func nullableInt(v int) any {
	if v <= 0 {
		return nil
	}
	return v
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
