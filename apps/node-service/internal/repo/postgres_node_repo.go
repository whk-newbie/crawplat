package repo

import (
	"context"
	"database/sql"
	"encoding/json"
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
