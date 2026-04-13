package service

import (
	"context"
	"sync"
	"time"
)

type Node struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Capabilities []string  `json:"capabilities"`
	LastSeenAt   time.Time `json:"lastSeenAt"`
}

type NodeService struct {
	repo Repository
}

type Repository interface {
	UpsertHeartbeat(ctx context.Context, name string, capabilities []string) (Node, error)
	ListOnline(ctx context.Context) ([]Node, error)
}

type memoryRepository struct {
	mu    sync.Mutex
	nodes map[string]Node
}

func (r *memoryRepository) UpsertHeartbeat(_ context.Context, name string, capabilities []string) (Node, error) {
	node := Node{
		ID:           name,
		Name:         name,
		Status:       "online",
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   time.Now(),
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nodes == nil {
		r.nodes = make(map[string]Node)
	}
	r.nodes[name] = node
	return node, nil
}

func (r *memoryRepository) ListOnline(_ context.Context) ([]Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodes := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func NewNodeService(repos ...Repository) *NodeService {
	if len(repos) > 0 && repos[0] != nil {
		return &NodeService{repo: repos[0]}
	}
	return &NodeService{repo: &memoryRepository{nodes: make(map[string]Node)}}
}

func (s *NodeService) Heartbeat(name string, capabilities []string) (Node, error) {
	return s.repo.UpsertHeartbeat(context.Background(), name, capabilities)
}

func (s *NodeService) List() ([]Node, error) {
	return s.repo.ListOnline(context.Background())
}
