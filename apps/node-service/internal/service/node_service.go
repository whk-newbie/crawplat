package service

import (
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
	mu    sync.Mutex
	nodes map[string]Node
}

func NewNodeService() *NodeService {
	return &NodeService{
		nodes: make(map[string]Node),
	}
}

func (s *NodeService) Heartbeat(name string, capabilities []string) Node {
	node := Node{
		ID:           name,
		Name:         name,
		Status:       "online",
		Capabilities: append([]string(nil), capabilities...),
		LastSeenAt:   time.Now(),
	}

	s.mu.Lock()
	s.nodes[name] = node
	s.mu.Unlock()

	return node
}

func (s *NodeService) List() []Node {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodes := make([]Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}
