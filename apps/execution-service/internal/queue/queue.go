package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const (
	executionQueue         = "executions:pending"
	executionInflightQueue = "executions:inflight"
)

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

func (q *RedisQueue) Enqueue(ctx context.Context, executionID string) error {
	return q.client.RPush(ctx, executionQueue, executionID).Err()
}

func (q *RedisQueue) Claim(ctx context.Context) (string, error) {
	id, err := q.client.LMove(ctx, executionQueue, executionInflightQueue, "LEFT", "RIGHT").Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

func (q *RedisQueue) Ack(ctx context.Context, executionID string) error {
	return q.client.LRem(ctx, executionInflightQueue, 1, executionID).Err()
}

func (q *RedisQueue) Release(ctx context.Context, executionID string) error {
	pipe := q.client.TxPipeline()
	pipe.LRem(ctx, executionInflightQueue, 1, executionID)
	pipe.LPush(ctx, executionQueue, executionID)
	_, err := pipe.Exec(ctx)
	return err
}
