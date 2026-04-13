package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const executionQueue = "executions:pending"

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
	id, err := q.client.LPop(ctx, executionQueue).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}
