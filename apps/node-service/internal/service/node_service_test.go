// 文件职责：NodeService 服务层的单元测试。
// 测试范围：
//   - Heartbeat 标记节点为 "online"
//   - List 返回所有已记录心跳的节点
// 使用 fakeNodeRepo（内存 map）作为测试替身，验证服务层的委托逻辑和不变式。
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

func (r *fakeNodeRepo) ListOnline(_ context.Context, limit, offset int) ([]Node, error) {
	nodes := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	if offset >= len(nodes) {
		return []Node{}, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(nodes) {
		end = len(nodes)
	}
	return nodes[offset:end], nil
}

func (r *fakeNodeRepo) GetByID(_ context.Context, nodeID string) (Node, error) {
	node, ok := r.nodes[nodeID]
	if !ok {
		return Node{}, ErrNodeNotFound
	}
	return node, nil
}

func (r *fakeNodeRepo) ListRecentExecutions(_ context.Context, nodeID string, query ExecutionQuery) ([]NodeExecution, error) {
	return []NodeExecution{}, nil
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

	nodes, err := svc.List(20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "node-a" {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}
}

func TestGetByIDReturnsNode(t *testing.T) {
	repo := &fakeNodeRepo{}
	svc := NewNodeService(repo)
	if _, err := svc.Heartbeat("node-a", []string{"docker", "go"}); err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}

	node, err := svc.GetByID("node-a")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if node.ID != "node-a" || node.Status != "online" {
		t.Fatalf("unexpected node: %#v", node)
	}
}

func TestGetByIDReturnsNotFoundForUnknownNode(t *testing.T) {
	svc := NewNodeService(&fakeNodeRepo{})
	_, err := svc.GetByID("missing")
	if err != ErrNodeNotFound {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}
