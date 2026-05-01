package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commonauth "crawler-platform/packages/go-common/auth"
)

func TestNewRouterDoesNotPanic(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected router construction to not panic, got %v", r)
		}
	}()

	router := NewRouter()
	if router == nil {
		t.Fatal("expected router to be non-nil")
	}
}

func TestNewRouterIncludesInternalExecutionRoutes(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	router := NewRouter()

	routes := router.Routes()
	seen := map[string]bool{}
	for _, route := range routes {
		seen[route.Path] = true
	}

	if !seen["/internal/v1/executions/claim"] {
		t.Fatal("expected internal execution claim route to be registered")
	}
	if !seen["/internal/v1/executions/:id/start"] {
		t.Fatal("expected internal execution start route to be registered")
	}
	if !seen["/internal/v1/executions/:id/logs"] {
		t.Fatal("expected internal execution log route to be registered")
	}
	if !seen["/internal/v1/executions/:id/complete"] {
		t.Fatal("expected internal execution complete route to be registered")
	}
	if !seen["/internal/v1/executions/:id/fail"] {
		t.Fatal("expected internal execution fail route to be registered")
	}
	if !seen["/api/v1/schedules"] {
		t.Fatal("expected scheduler route to be registered")
	}
	if !seen["/api/v1/monitor/*path"] {
		t.Fatal("expected monitor route to be registered")
	}
}

func TestInternalExecutionRoutesRequireToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/executions/claim", strings.NewReader(`{"nodeId":"node-1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 without token, got %d", w.Code)
	}
}

func TestInternalExecutionRoutesAcceptValidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/executions/claim", strings.NewReader(`{"nodeId":"node-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalTokenHeader, "test-token")
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("expected valid token to pass internal auth, got %d", w.Code)
	}
}

func TestPublicRoutesBypassJWTWhenDisabled(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("expected jwt-disabled gateway to bypass auth, got %d", w.Code)
	}
}

func TestPublicRoutesRequireJWTWhenEnabled(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "true")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without bearer token, got %d", w.Code)
	}
}

func TestPublicRoutesAcceptValidJWTWhenEnabled(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "true")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	token, err := commonauth.IssueToken("test-secret", "user-1")
	if err != nil {
		t.Fatalf("IssueToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("expected valid token to pass jwt auth, got %d", w.Code)
	}
}

func TestAuthRoutesBypassJWTWhenEnabled(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "true")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("expected auth routes to bypass jwt auth, got %d", w.Code)
	}
}

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
}

func (r *closeNotifyRecorder) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}
