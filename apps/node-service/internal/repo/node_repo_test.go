package repo

import (
	"context"
	"testing"
	"time"

	"crawler-platform/apps/node-service/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newRedisNodeRepo(t *testing.T) *RedisRepository {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run returned error: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return NewRedisRepository(client, 30*time.Second)
}

func TestNodeRepoStoresHeartbeatWithTTL(t *testing.T) {
	repo := newRedisNodeRepo(t)
	node, err := repo.UpsertHeartbeat(context.Background(), "node-1", []string{"docker", "go"})
	if err != nil {
		t.Fatalf("UpsertHeartbeat returned error: %v", err)
	}
	if node.Name != "node-1" || node.Status != "online" {
		t.Fatalf("unexpected node: %#v", node)
	}

	nodes, err := repo.ListOnline(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("ListOnline returned error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "node-1" {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}
}

func TestNodeRepoGetByID(t *testing.T) {
	repo := newRedisNodeRepo(t)
	if _, err := repo.UpsertHeartbeat(context.Background(), "node-1", []string{"docker", "go"}); err != nil {
		t.Fatalf("UpsertHeartbeat returned error: %v", err)
	}

	node, err := repo.GetByID(context.Background(), "node-1")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if node.ID != "node-1" || len(node.Capabilities) != 2 {
		t.Fatalf("unexpected node: %#v", node)
	}
}

func TestNodeRepoGetByIDNotFound(t *testing.T) {
	repo := newRedisNodeRepo(t)
	_, err := repo.GetByID(context.Background(), "missing")
	if err != service.ErrNodeNotFound {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

var _ service.Repository = (*RedisRepository)(nil)
