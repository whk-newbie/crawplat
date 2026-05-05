// Redis 队列层单元测试。
// 使用 miniredis（内存 Redis mock）验证 Enqueue/Claim/Ack/Release 的 FIFO 语义和原子性。
// 不依赖真实 Redis 实例。
package queue

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newRedisQueue 创建基于 miniredis 的测试队列实例，测试结束后自动清理。
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
	if err := queue.Ack(context.Background(), "exec-1"); err != nil {
		t.Fatalf("Ack returned error: %v", err)
	}
}

func TestQueueClaimIsFIFO(t *testing.T) {
	queue := newRedisQueue(t)
	for _, id := range []string{"exec-1", "exec-2"} {
		if err := queue.Enqueue(context.Background(), id); err != nil {
			t.Fatalf("Enqueue returned error: %v", err)
		}
	}

	first, err := queue.Claim(context.Background())
	if err != nil || first != "exec-1" {
		t.Fatalf("unexpected first claim result: %q %v", first, err)
	}
	second, err := queue.Claim(context.Background())
	if err != nil || second != "exec-2" {
		t.Fatalf("unexpected second claim result: %q %v", second, err)
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

func TestQueueReleaseReturnsClaimedExecutionToPending(t *testing.T) {
	queue := newRedisQueue(t)
	if err := queue.Enqueue(context.Background(), "exec-1"); err != nil {
		t.Fatalf("Enqueue returned error: %v", err)
	}
	id, err := queue.Claim(context.Background())
	if err != nil || id != "exec-1" {
		t.Fatalf("unexpected claim result: %q %v", id, err)
	}
	if err := queue.Release(context.Background(), "exec-1"); err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
	id, err = queue.Claim(context.Background())
	if err != nil || id != "exec-1" {
		t.Fatalf("unexpected re-claim result: %q %v", id, err)
	}
}
