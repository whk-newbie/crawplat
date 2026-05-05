package service

import (
	"context"
	"errors"
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

type NodeHeartbeat struct {
	SeenAt       time.Time `json:"seenAt"`
	Capabilities []string  `json:"capabilities"`
}

type ExecutionQuery struct {
	Status string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

type NodeExecution struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	SpiderID      string     `json:"spiderId"`
	Status        string     `json:"status"`
	TriggerSource string     `json:"triggerSource"`
	CreatedAt     time.Time  `json:"createdAt"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	FinishedAt    *time.Time `json:"finishedAt,omitempty"`
}

var ErrNodeNotFound = errors.New("node not found")

type NodeService struct {
	repo Repository
}

type Repository interface {
	UpsertHeartbeat(ctx context.Context, name string, capabilities []string) (Node, error)
	ListOnline(ctx context.Context, limit, offset int) ([]Node, error)
	GetByID(ctx context.Context, nodeID string) (Node, error)
	ListRecentExecutions(ctx context.Context, nodeID string, query ExecutionQuery) ([]NodeExecution, error)
}

type CatalogRepository interface {
	Repository
	UpsertCatalog(ctx context.Context, name string, capabilities []string, seenAt time.Time) (Node, error)
	ListCatalog(ctx context.Context) ([]Node, error)
	ListHeartbeatHistory(ctx context.Context, nodeID string, limit int) ([]NodeHeartbeat, error)
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

func (r *memoryRepository) ListOnline(_ context.Context, limit, offset int) ([]Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

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

func (r *memoryRepository) GetByID(_ context.Context, nodeID string) (Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[nodeID]
	if !ok {
		return Node{}, ErrNodeNotFound
	}
	return node, nil
}

func (r *memoryRepository) ListRecentExecutions(_ context.Context, nodeID string, query ExecutionQuery) ([]NodeExecution, error) {
	return []NodeExecution{}, nil
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

func (s *NodeService) List(limit, offset int) ([]Node, error) {
	return s.repo.ListOnline(context.Background(), limit, offset)
}

func (s *NodeService) GetByID(nodeID string) (Node, error) {
	return s.repo.GetByID(context.Background(), nodeID)
}

func (s *NodeService) ListRecentExecutions(nodeID string, query ExecutionQuery) ([]NodeExecution, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	return s.repo.ListRecentExecutions(context.Background(), nodeID, query)
}
