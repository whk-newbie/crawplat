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

type PostgresNodeRepository struct {
	db *sql.DB
}

func NewPostgresNodeRepository(db *sql.DB) *PostgresNodeRepository {
	return &PostgresNodeRepository{db: db}
}

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
			node            service.Node
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
