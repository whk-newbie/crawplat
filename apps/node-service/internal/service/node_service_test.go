package service

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"
)

type fakeLiveRepo struct {
	nodes       map[string]Node
	upsertCalls int
	listCalls   int
	upsertErr   error
	listErr     error
}

func (r *fakeLiveRepo) UpsertHeartbeat(_ context.Context, name string, capabilities []string) (Node, error) {
	if r.upsertErr != nil {
		return Node{}, r.upsertErr
	}
	node := Node{
		ID:           name,
		Name:         name,
		Status:       "online",
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   time.Unix(1700000000, 0).UTC(),
	}
	r.upsertCalls++
	if r.nodes == nil {
		r.nodes = map[string]Node{}
	}
	r.nodes[name] = node
	return node, nil
}

func (r *fakeLiveRepo) ListOnline(_ context.Context) ([]Node, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	r.listCalls++
	nodes := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return nodes, nil
}

type fakeCatalogRepo struct {
	nodes       map[string]Node
	upsertCalls int
	listCalls   int
	upsertErr   error
	listErr     error
}

func (r *fakeCatalogRepo) UpsertCatalog(_ context.Context, name string, capabilities []string, seenAt time.Time) (Node, error) {
	if r.upsertErr != nil {
		return Node{}, r.upsertErr
	}
	r.upsertCalls++
	node := Node{
		ID:           name,
		Name:         name,
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   seenAt,
	}
	if r.nodes == nil {
		r.nodes = map[string]Node{}
	}
	r.nodes[name] = node
	return node, nil
}

func (r *fakeCatalogRepo) ListCatalog(_ context.Context) ([]Node, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	r.listCalls++
	nodes := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return nodes, nil
}

func TestHeartbeatWritesLiveAndCatalog(t *testing.T) {
	live := &fakeLiveRepo{}
	catalog := &fakeCatalogRepo{}
	svc := NewNodeServiceWithCatalog(live, catalog)

	node, err := svc.Heartbeat("node-a", []string{"docker", "python", "go"})
	if err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}
	if node.Status != "online" {
		t.Fatalf("expected online, got %s", node.Status)
	}
	if live.upsertCalls != 1 || catalog.upsertCalls != 1 {
		t.Fatalf("expected one upsert for both repos, got live=%d catalog=%d", live.upsertCalls, catalog.upsertCalls)
	}
	catalogNode, ok := catalog.nodes["node-a"]
	if !ok {
		t.Fatalf("expected node-a in catalog")
	}
	if !catalogNode.LastSeenAt.Equal(node.LastSeenAt) {
		t.Fatalf("expected catalog seenAt=%v, got %v", node.LastSeenAt, catalogNode.LastSeenAt)
	}
}

func TestListMergesCatalogAndLiveWithOnlineOfflineStatus(t *testing.T) {
	catalogSeenAt := time.Unix(1700000000, 0).UTC()
	liveSeenAt := time.Unix(1700001000, 0).UTC()

	live := &fakeLiveRepo{
		nodes: map[string]Node{
			"node-a": {
				ID:           "node-a",
				Name:         "node-a",
				Status:       "online",
				Capabilities: []string{"go"},
				LastSeenAt:   liveSeenAt,
			},
			"node-c": {
				ID:           "node-c",
				Name:         "node-c",
				Status:       "online",
				Capabilities: []string{"python"},
				LastSeenAt:   liveSeenAt,
			},
		},
	}
	catalog := &fakeCatalogRepo{
		nodes: map[string]Node{
			"node-a": {
				ID:           "node-a",
				Name:         "node-a",
				Capabilities: []string{"go"},
				LastSeenAt:   catalogSeenAt,
			},
			"node-b": {
				ID:           "node-b",
				Name:         "node-b",
				Capabilities: []string{"docker"},
				LastSeenAt:   catalogSeenAt,
			},
		},
	}
	svc := NewNodeServiceWithCatalog(live, catalog)

	nodes, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}

	byID := make(map[string]Node, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = node
	}

	if byID["node-a"].Status != "online" || !byID["node-a"].LastSeenAt.Equal(liveSeenAt) {
		t.Fatalf("expected node-a online with live seenAt, got %#v", byID["node-a"])
	}
	if byID["node-b"].Status != "offline" {
		t.Fatalf("expected node-b offline, got %#v", byID["node-b"])
	}
	if byID["node-c"].Status != "online" {
		t.Fatalf("expected node-c online, got %#v", byID["node-c"])
	}
}

func TestHeartbeatReturnsCatalogError(t *testing.T) {
	live := &fakeLiveRepo{}
	catalog := &fakeCatalogRepo{upsertErr: errors.New("catalog unavailable")}
	svc := NewNodeServiceWithCatalog(live, catalog)

	if _, err := svc.Heartbeat("node-a", []string{"go"}); err == nil {
		t.Fatalf("expected error when catalog upsert fails")
	}
}
