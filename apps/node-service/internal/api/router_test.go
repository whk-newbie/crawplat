package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/node-service/internal/service"
	"github.com/gin-gonic/gin"
)

func TestHeartbeatRoutePersistsNodeState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewNodeService()
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-a/heartbeat", strings.NewReader(`{"capabilities":["docker","python","go"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var node service.Node
	if err := json.Unmarshal(w.Body.Bytes(), &node); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if node.ID != "node-a" {
		t.Fatalf("expected node id node-a, got %s", node.ID)
	}
	if node.Status != "online" {
		t.Fatalf("expected status online, got %s", node.Status)
	}
	if len(node.Capabilities) != 3 || node.Capabilities[0] != "docker" || node.Capabilities[1] != "python" || node.Capabilities[2] != "go" {
		t.Fatalf("expected capabilities [docker python go], got %+v", node.Capabilities)
	}
}

func TestListRouteReturnsPersistedNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewNodeService()
	svc.Heartbeat("node-a", []string{"docker", "python", "go"})
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var nodes []service.Node
	if err := json.Unmarshal(w.Body.Bytes(), &nodes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].ID != "node-a" {
		t.Fatalf("expected node id node-a, got %s", nodes[0].ID)
	}
	if nodes[0].Status != "online" {
		t.Fatalf("expected status online, got %s", nodes[0].Status)
	}
	if len(nodes[0].Capabilities) != 3 || nodes[0].Capabilities[0] != "docker" || nodes[0].Capabilities[1] != "python" || nodes[0].Capabilities[2] != "go" {
		t.Fatalf("expected capabilities [docker python go], got %+v", nodes[0].Capabilities)
	}
}

func TestHeartbeatRouteRejectsInvalidNodeID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewNodeService())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node!a/heartbeat", strings.NewReader(`{"capabilities":["docker"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid node id") {
		t.Fatalf("expected invalid node id error, got %s", w.Body.String())
	}
}
