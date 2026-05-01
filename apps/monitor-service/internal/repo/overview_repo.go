package repo

import (
	"context"
	"database/sql"

	"crawler-platform/apps/monitor-service/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	nodeKeyPrefix = "nodes:"
	nodeIndexKey  = "nodes:online"
)

type OverviewRepository struct {
	db     *sql.DB
	client *redis.Client
}

func NewOverviewRepository(db *sql.DB, client *redis.Client) *OverviewRepository {
	return &OverviewRepository{db: db, client: client}
}

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
