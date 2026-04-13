package service

import (
	"context"
	"testing"
)

type fakeNodeRepo struct {
	nodes map[string]Node
}

func (r *fakeNodeRepo) UpsertHeartbeat(_ context.Context, name string, capabilities []string) (Node, error) {
	node := Node{
		ID:           name,
		Name:         name,
		Status:       "online",
		Capabilities: append([]string(nil), capabilities...),
	}
	if r.nodes == nil {
		r.nodes = map[string]Node{}
	}
	r.nodes[name] = node
	return node, nil
}

func (r *fakeNodeRepo) ListOnline(_ context.Context) ([]Node, error) {
	nodes := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func TestHeartbeatMarksNodeOnline(t *testing.T) {
	svc := NewNodeService(&fakeNodeRepo{})
	node, err := svc.Heartbeat("node-a", []string{"docker", "python", "go"})
	if err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}
	if node.Status != "online" {
		t.Fatalf("expected online, got %s", node.Status)
	}
}

func TestListReturnsRepoNodes(t *testing.T) {
	repo := &fakeNodeRepo{}
	svc := NewNodeService(repo)
	if _, err := svc.Heartbeat("node-a", []string{"docker", "go"}); err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}

	nodes, err := svc.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "node-a" {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}
}
