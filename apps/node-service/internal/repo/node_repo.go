package repo

import (
	"context"
	"encoding/json"
	"time"

	"crawler-platform/apps/node-service/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	nodeKeyPrefix = "nodes:"
	nodeIndexKey  = "nodes:online"
)

type RedisRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepository(client *redis.Client, ttl time.Duration) *RedisRepository {
	return &RedisRepository{client: client, ttl: ttl}
}

func (r *RedisRepository) UpsertHeartbeat(ctx context.Context, orgID, name string, capabilities []string) (service.Node, error) {
	node := service.Node{
		ID:           name,
		Name:         name,
		Status:       "online",
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   time.Now(),
	}

	payload, err := json.Marshal(node)
	if err != nil {
		return service.Node{}, err
	}

	key := nodeKeyPrefix + name
	if err := r.client.Set(ctx, key, payload, r.ttl).Err(); err != nil {
		return service.Node{}, err
	}
	if err := r.client.SAdd(ctx, nodeIndexKey, name).Err(); err != nil {
		return service.Node{}, err
	}
	return node, nil
}

func (r *RedisRepository) ListOnline(ctx context.Context, orgID string, limit, offset int) ([]service.Node, error) {
	ids, err := r.client.SMembers(ctx, nodeIndexKey).Result()
	if err != nil {
		return nil, err
	}

	nodes := make([]service.Node, 0, len(ids))
	for _, id := range ids {
		payload, err := r.client.Get(ctx, nodeKeyPrefix+id).Result()
		if err == redis.Nil {
			_ = r.client.SRem(ctx, nodeIndexKey, id).Err()
			continue
		}
		if err != nil {
			return nil, err
		}

		var node service.Node
		if err := json.Unmarshal([]byte(payload), &node); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	if offset >= len(nodes) {
		return []service.Node{}, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(nodes) {
		end = len(nodes)
	}
	return nodes[offset:end], nil
}

func (r *RedisRepository) GetByID(ctx context.Context, orgID, nodeID string) (service.Node, error) {
	payload, err := r.client.Get(ctx, nodeKeyPrefix+nodeID).Result()
	if err == redis.Nil {
		return service.Node{}, service.ErrNodeNotFound
	}
	if err != nil {
		return service.Node{}, err
	}

	var node service.Node
	if err := json.Unmarshal([]byte(payload), &node); err != nil {
		return service.Node{}, err
	}
	return node, nil
}

func (r *RedisRepository) ListRecentExecutions(ctx context.Context, orgID, nodeID string, query service.ExecutionQuery) ([]service.NodeExecution, error) {
	return []service.NodeExecution{}, nil
}
