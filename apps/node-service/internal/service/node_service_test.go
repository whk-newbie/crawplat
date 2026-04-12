package service

import "testing"

func TestHeartbeatMarksNodeOnline(t *testing.T) {
	svc := NewNodeService()
	node := svc.Heartbeat("node-a", []string{"docker", "python", "go"})
	if node.Status != "online" {
		t.Fatalf("expected online, got %s", node.Status)
	}
}
