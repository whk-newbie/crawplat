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
	if _, err := svc.Heartbeat("node-a", []string{"docker", "python", "go"}); err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}
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

func TestNodeDetailRouteReturnsNodeDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewNodeService()
	if _, err := svc.Heartbeat("node-a", []string{"docker", "python", "go"}); err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/nodes/node-a?limit=5&executionLimit=3&executionOffset=0&executionStatus=succeeded&executionFrom=2023-11-14T22:12:00Z&executionTo=2023-11-14T22:16:00Z",
		nil,
	)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var detail struct {
		Node             service.Node            `json:"node"`
		HeartbeatHistory []service.NodeHeartbeat `json:"heartbeatHistory"`
		RecentExecutions []service.NodeExecution `json:"recentExecutions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if detail.Node.ID != "node-a" {
		t.Fatalf("expected node id node-a, got %s", detail.Node.ID)
	}
	if len(detail.HeartbeatHistory) == 0 {
		t.Fatalf("expected heartbeat history, got %#v", detail.HeartbeatHistory)
	}
	if detail.RecentExecutions == nil {
		t.Fatalf("expected recentExecutions array, got nil")
	}
}

func TestNodeDetailRouteReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewNodeService())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/missing", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "node not found") {
		t.Fatalf("expected node not found error, got %s", w.Body.String())
	}
}

func TestNodeDetailRouteRejectsInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewNodeService())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a?limit=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for non-number limit, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a?limit=0", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for non-positive limit, got %d", w2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a?executionLimit=-1", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for executionLimit, got %d", w3.Code)
	}

	req4 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a?executionOffset=-1", nil)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for executionOffset, got %d", w4.Code)
	}

	req5 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a?executionFrom=not-time", nil)
	w5 := httptest.NewRecorder()
	router.ServeHTTP(w5, req5)
	if w5.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid executionFrom, got %d", w5.Code)
	}

	req6 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a?executionTo=not-time", nil)
	w6 := httptest.NewRecorder()
	router.ServeHTTP(w6, req6)
	if w6.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid executionTo, got %d", w6.Code)
	}

	req7 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a?executionFrom=2023-11-14T22:16:00Z&executionTo=2023-11-14T22:12:00Z", nil)
	w7 := httptest.NewRecorder()
	router.ServeHTTP(w7, req7)
	if w7.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for executionFrom > executionTo, got %d", w7.Code)
	}
}

func TestNodeSessionsRouteReturnsSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewNodeService()
	if _, err := svc.Heartbeat("node-a", []string{"go"}); err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a/sessions?limit=2&gapSeconds=60", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var response struct {
		Sessions []service.NodeSession      `json:"sessions"`
		Summary  service.NodeSessionSummary `json:"summary"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Sessions) == 0 {
		t.Fatalf("expected non-empty sessions")
	}
	if response.Summary.TotalSessions == 0 || response.Summary.TotalHeartbeatCount == 0 {
		t.Fatalf("expected non-empty summary, got %#v", response.Summary)
	}
}

func TestNodeSessionsRouteRejectsInvalidParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(service.NewNodeService())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a/sessions?limit=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for limit=0, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a/sessions?gapSeconds=0", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for gapSeconds=0, got %d", w2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-a/sessions?gapSeconds=3601", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for gapSeconds=3601, got %d", w3.Code)
	}
}
