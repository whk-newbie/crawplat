package queue

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newRedisQueue(t *testing.T) *RedisQueue {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run returned error: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return NewRedisQueue(client)
}

func TestQueuePushAndClaim(t *testing.T) {
	queue := newRedisQueue(t)
	if err := queue.Enqueue(context.Background(), "exec-1"); err != nil {
		t.Fatalf("Enqueue returned error: %v", err)
	}
	id, err := queue.Claim(context.Background())
	if err != nil || id != "exec-1" {
		t.Fatalf("unexpected claim result: %q %v", id, err)
	}
}

func TestQueueClaimReturnsEmptyWhenNoWork(t *testing.T) {
	queue := newRedisQueue(t)
	id, err := queue.Claim(context.Background())
	if err != nil {
		t.Fatalf("Claim returned error: %v", err)
	}
	if id != "" {
		t.Fatalf("expected no work, got %q", id)
	}
}
