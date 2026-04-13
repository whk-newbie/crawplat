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

	nodes, err := repo.ListOnline(context.Background())
	if err != nil {
		t.Fatalf("ListOnline returned error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "node-1" {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}
}

var _ interface {
	UpsertHeartbeat(context.Context, string, []string) (service.Node, error)
	ListOnline(context.Context) ([]service.Node, error)
} = (*RedisRepository)(nil)
