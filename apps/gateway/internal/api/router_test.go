package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func TestNewRouterIncludesGatewayRoutes(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	router := NewRouter()

	routes := router.Routes()
	seen := map[string]bool{}
	for _, route := range routes {
		seen[route.Path] = true
	}

	assertRouteRegistered(t, seen, "/internal/v1/executions/claim")
	assertRouteRegistered(t, seen, "/internal/v1/executions/:id/start")
	assertRouteRegistered(t, seen, "/internal/v1/executions/:id/logs")
	assertRouteRegistered(t, seen, "/internal/v1/executions/:id/complete")
	assertRouteRegistered(t, seen, "/internal/v1/executions/:id/fail")
	assertRouteRegistered(t, seen, "/internal/v1/executions/retries/materialize")
	assertRouteRegistered(t, seen, "/api/v1/projects/*path")
	assertRouteRegistered(t, seen, "/api/v1/spiders/:spiderId/versions")
	assertRouteRegistered(t, seen, "/api/v1/schedules/*path")
	assertRouteRegistered(t, seen, "/api/v1/monitor/*path")
}

func TestInternalExecutionRoutesRequireToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/executions/claim", strings.NewReader(`{"nodeId":"node-1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusUnauthorized, "unauthorized internal route")
}

func TestInternalExecutionRoutesRejectWhenTokenNotConfigured(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/executions/claim", strings.NewReader(`{"nodeId":"node-1"}`))
	req.Header.Set(internalTokenHeader, "test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusUnauthorized, "internal token is not configured")
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

func TestRequestIDIsReturnedAndTrusted(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v9/projects", nil)
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "req-123" {
		t.Fatalf("expected request id to be returned, got %q", got)
	}
}

func TestUnsupportedAPIVersionUsesUnifiedError(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v9/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusNotFound, "unsupported api version or route")
}

func TestConfiguredAPIVersionIsRegisteredAndRewritten(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	t.Setenv("GATEWAY_API_SUPPORTED_VERSIONS", "v1,v2")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v2/projects", nil)
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatalf("expected configured api version to be routed, got %d", w.Code)
	}
	if got := w.Header().Get("X-API-Version"); got != "v2" {
		t.Fatalf("expected response api version v2, got %q", got)
	}
}

func TestJWTProtectionRequiresBearerToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	t.Setenv("GATEWAY_ENFORCE_JWT", "true")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusUnauthorized, "missing bearer token")
}

func TestJWTProtectionRejectsInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	t.Setenv("GATEWAY_ENFORCE_JWT", "true")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusUnauthorized, "invalid bearer token")
}

func TestJWTProtectionAllowsAuthRoutes(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-token")
	t.Setenv("GATEWAY_ENFORCE_JWT", "true")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"u"}`))
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatal("expected auth route to bypass jwt middleware")
	}
}

func assertRouteRegistered(t *testing.T, seen map[string]bool, path string) {
	t.Helper()
	if !seen[path] {
		t.Fatalf("expected route %s to be registered", path)
	}
}

func assertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantError string) {
	t.Helper()
	if w.Code != wantStatus {
		t.Fatalf("expected status %d, got %d with body %s", wantStatus, w.Code, w.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected json error response, got %q: %v", w.Body.String(), err)
	}
	if body["error"] != wantError {
		t.Fatalf("expected error %q, got %q", wantError, body["error"])
	}
}

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
}

func (r *closeNotifyRecorder) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}
