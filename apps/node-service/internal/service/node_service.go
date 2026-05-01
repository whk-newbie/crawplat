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

type NodeSession struct {
	StartedAt       time.Time `json:"startedAt"`
	EndedAt         time.Time `json:"endedAt"`
	HeartbeatCount  int       `json:"heartbeatCount"`
	DurationSeconds int64     `json:"durationSeconds"`
}

type NodeSessionSummary struct {
	TotalSessions              int   `json:"totalSessions"`
	TotalHeartbeatCount        int   `json:"totalHeartbeatCount"`
	TotalOnlineDurationSeconds int64 `json:"totalOnlineDurationSeconds"`
}

type NodeSessionsResult struct {
	Sessions []NodeSession      `json:"sessions"`
	Summary  NodeSessionSummary `json:"summary"`
}

type NodeDetail struct {
	Node             Node            `json:"node"`
	HeartbeatHistory []NodeHeartbeat `json:"heartbeatHistory"`
	RecentExecutions []NodeExecution `json:"recentExecutions"`
}

type ExecutionQuery struct {
	Limit  int
	Offset int
	Status string
	From   *time.Time
	To     *time.Time
}

type DetailQuery struct {
	HeartbeatLimit int
	ExecutionQuery ExecutionQuery
}

var ErrNodeNotFound = errors.New("node not found")

type NodeService struct {
	liveRepo    LiveRepository
	catalogRepo CatalogRepository
}

type LiveRepository interface {
	UpsertHeartbeat(ctx context.Context, name string, capabilities []string) (Node, error)
	ListOnline(ctx context.Context) ([]Node, error)
}

type CatalogRepository interface {
	UpsertCatalog(ctx context.Context, name string, capabilities []string, seenAt time.Time) (Node, error)
	ListCatalog(ctx context.Context) ([]Node, error)
	GetByID(ctx context.Context, nodeID string) (Node, error)
	ListHeartbeatHistory(ctx context.Context, nodeID string, limit int) ([]NodeHeartbeat, error)
	ListRecentExecutions(ctx context.Context, nodeID string, query ExecutionQuery) ([]NodeExecution, error)
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

func (r *memoryRepository) UpsertCatalog(_ context.Context, name string, capabilities []string, seenAt time.Time) (Node, error) {
	node := Node{
		ID:           name,
		Name:         name,
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   seenAt,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.nodes == nil {
		r.nodes = make(map[string]Node)
	}
	if existing, ok := r.nodes[name]; ok {
		node.Status = existing.Status
	}
	r.nodes[name] = node
	return node, nil
}

func (r *memoryRepository) ListCatalog(_ context.Context) ([]Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodes := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
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

func (r *memoryRepository) ListHeartbeatHistory(ctx context.Context, nodeID string, limit int) ([]NodeHeartbeat, error) {
	node, err := r.GetByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	history := []NodeHeartbeat{{
		SeenAt:       node.LastSeenAt,
		Capabilities: append([]string(nil), node.Capabilities...),
	}}
	if limit <= 0 || len(history) <= limit {
		return history, nil
	}
	return history[:limit], nil
}

func (r *memoryRepository) ListRecentExecutions(_ context.Context, nodeID string, _ ExecutionQuery) ([]NodeExecution, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.nodes[nodeID]; !ok {
		return nil, ErrNodeNotFound
	}
	return []NodeExecution{}, nil
}

func NewNodeService(repos ...LiveRepository) *NodeService {
	mem := &memoryRepository{nodes: make(map[string]Node)}
	liveRepo := LiveRepository(mem)
	if len(repos) > 0 && repos[0] != nil {
		liveRepo = repos[0]
	}

	return NewNodeServiceWithCatalog(liveRepo, mem)
}

func NewNodeServiceWithCatalog(liveRepo LiveRepository, catalogRepo CatalogRepository) *NodeService {
	if liveRepo == nil {
		liveRepo = &memoryRepository{nodes: make(map[string]Node)}
	}
	if catalogRepo == nil {
		catalogRepo = &memoryRepository{nodes: make(map[string]Node)}
	}
	return &NodeService{
		liveRepo:    liveRepo,
		catalogRepo: catalogRepo,
	}
}

func (s *NodeService) Heartbeat(name string, capabilities []string) (Node, error) {
	node, err := s.liveRepo.UpsertHeartbeat(context.Background(), name, capabilities)
	if err != nil {
		return Node{}, err
	}

	seenAt := node.LastSeenAt
	if seenAt.IsZero() {
		seenAt = time.Now()
	}

	if _, err := s.catalogRepo.UpsertCatalog(context.Background(), name, capabilities, seenAt); err != nil {
		return Node{}, err
	}
	return node, nil
}

func (s *NodeService) List() ([]Node, error) {
	catalogNodes, err := s.catalogRepo.ListCatalog(context.Background())
	if err != nil {
		return nil, err
	}
	liveNodes, err := s.liveRepo.ListOnline(context.Background())
	if err != nil {
		return nil, err
	}

	liveByID := make(map[string]Node, len(liveNodes))
	for _, node := range liveNodes {
		liveByID[node.ID] = node
	}

	nodes := make([]Node, 0, len(catalogNodes)+len(liveNodes))
	for _, node := range catalogNodes {
		if live, ok := liveByID[node.ID]; ok {
			node.Status = "online"
			if !live.LastSeenAt.IsZero() {
				node.LastSeenAt = live.LastSeenAt
			}
			delete(liveByID, node.ID)
		} else {
			node.Status = "offline"
		}
		nodes = append(nodes, node)
	}

	// Include nodes that are live but missing from catalog to avoid dropping recent heartbeats.
	for _, live := range liveByID {
		if live.Status == "" {
			live.Status = "online"
		}
		nodes = append(nodes, live)
	}

	return nodes, nil
}

func (s *NodeService) Detail(id string, query DetailQuery) (NodeDetail, error) {
	node, err := s.catalogRepo.GetByID(context.Background(), id)
	if err != nil {
		return NodeDetail{}, err
	}

	node.Status = "offline"
	liveNodes, err := s.liveRepo.ListOnline(context.Background())
	if err != nil {
		return NodeDetail{}, err
	}
	for _, live := range liveNodes {
		if live.ID != id {
			continue
		}
		node.Status = "online"
		if !live.LastSeenAt.IsZero() {
			node.LastSeenAt = live.LastSeenAt
		}
		if len(live.Capabilities) > 0 {
			node.Capabilities = append([]string(nil), live.Capabilities...)
		}
		break
	}

	history, err := s.catalogRepo.ListHeartbeatHistory(context.Background(), id, query.HeartbeatLimit)
	if err != nil {
		return NodeDetail{}, err
	}
	recentExecutions, err := s.catalogRepo.ListRecentExecutions(context.Background(), id, query.ExecutionQuery)
	if err != nil {
		return NodeDetail{}, err
	}

	return NodeDetail{
		Node:             node,
		HeartbeatHistory: history,
		RecentExecutions: recentExecutions,
	}, nil
}

func (s *NodeService) Sessions(id string, limit int, gapSeconds int) (NodeSessionsResult, error) {
	if _, err := s.catalogRepo.GetByID(context.Background(), id); err != nil {
		return NodeSessionsResult{}, err
	}

	historyLimit := limit * 20
	if historyLimit < 200 {
		historyLimit = 200
	}
	if historyLimit > 5000 {
		historyLimit = 5000
	}

	history, err := s.catalogRepo.ListHeartbeatHistory(context.Background(), id, historyLimit)
	if err != nil {
		return NodeSessionsResult{}, err
	}
	sessions := aggregateSessions(history, time.Duration(gapSeconds)*time.Second)
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}

	summary := NodeSessionSummary{
		TotalSessions: len(sessions),
	}
	for _, s := range sessions {
		summary.TotalHeartbeatCount += s.HeartbeatCount
		summary.TotalOnlineDurationSeconds += s.DurationSeconds
	}

	return NodeSessionsResult{
		Sessions: sessions,
		Summary:  summary,
	}, nil
}

func aggregateSessions(history []NodeHeartbeat, gap time.Duration) []NodeSession {
	if len(history) == 0 {
		return []NodeSession{}
	}

	sessions := make([]NodeSession, 0)
	current := NodeSession{
		StartedAt:      history[0].SeenAt,
		EndedAt:        history[0].SeenAt,
		HeartbeatCount: 1,
	}
	prevSeenAt := history[0].SeenAt

	for i := 1; i < len(history); i++ {
		seenAt := history[i].SeenAt
		if prevSeenAt.Sub(seenAt) <= gap {
			current.StartedAt = seenAt
			current.HeartbeatCount++
			prevSeenAt = seenAt
			continue
		}

		current.DurationSeconds = int64(current.EndedAt.Sub(current.StartedAt).Seconds())
		sessions = append(sessions, current)

		current = NodeSession{
			StartedAt:      seenAt,
			EndedAt:        seenAt,
			HeartbeatCount: 1,
		}
		prevSeenAt = seenAt
	}

	current.DurationSeconds = int64(current.EndedAt.Sub(current.StartedAt).Seconds())
	sessions = append(sessions, current)
	return sessions
}
