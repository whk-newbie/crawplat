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

func TestPublicRoutesDoNotRateLimitWhenDisabled(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("GATEWAY_RATE_LIMIT_ENABLED", "false")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
		router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("expected no rate limit when disabled, request #%d got 429", i+1)
		}
	}
}

func TestPublicRoutesRateLimitWhenEnabled(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("GATEWAY_RATE_LIMIT_ENABLED", "true")
	t.Setenv("GATEWAY_RATE_LIMIT_WINDOW_SECONDS", "60")
	t.Setenv("GATEWAY_RATE_LIMIT_MAX_REQUESTS", "1")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	firstReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	firstReq.Header.Set("X-Forwarded-For", "10.0.0.2")
	first := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(first, firstReq)
	if first.Code == http.StatusTooManyRequests {
		t.Fatalf("expected first request to pass rate limit, got 429")
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	secondReq.Header.Set("X-Forwarded-For", "10.0.0.2")
	second := httptest.NewRecorder()
	router.ServeHTTP(second, secondReq)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request to hit rate limit, got %d", second.Code)
	}
}

func TestPublicRoutesRateLimitIsPerClientKey(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("GATEWAY_RATE_LIMIT_ENABLED", "true")
	t.Setenv("GATEWAY_RATE_LIMIT_WINDOW_SECONDS", "60")
	t.Setenv("GATEWAY_RATE_LIMIT_MAX_REQUESTS", "1")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	reqA := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	reqA.Header.Set("X-Forwarded-For", "10.0.0.3")
	wA := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(wA, reqA)
	if wA.Code == http.StatusTooManyRequests {
		t.Fatalf("expected first client request to pass, got 429")
	}

	reqB := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	reqB.Header.Set("X-Forwarded-For", "10.0.0.4")
	wB := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(wB, reqB)
	if wB.Code == http.StatusTooManyRequests {
		t.Fatalf("expected different client key to have separate quota, got 429")
	}
}

func TestGatewaySetsRequestIDHeaderWhenMissing(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	requestID := strings.TrimSpace(w.Header().Get("X-Request-Id"))
	if requestID == "" {
		t.Fatal("expected gateway to set X-Request-Id header")
	}
}

func TestGatewayTrustsIncomingRequestIDByDefault(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("X-Request-Id", "req-fixed-123")
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-Id"); got != "req-fixed-123" {
		t.Fatalf("expected gateway to preserve incoming request id, got %q", got)
	}
}

func TestGatewayCanDisableTrustIncomingRequestID(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("GATEWAY_TRUST_REQUEST_ID", "false")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("X-Request-Id", "req-fixed-override-me")
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	got := strings.TrimSpace(w.Header().Get("X-Request-Id"))
	if got == "" {
		t.Fatal("expected generated request id when trust incoming is disabled")
	}
	if got == "req-fixed-override-me" {
		t.Fatalf("expected gateway to override incoming request id, got %q", got)
	}
}

func TestConfiguredApiVersionIsRouted(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("GATEWAY_API_SUPPORTED_VERSIONS", "v1,v2")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	routes := router.Routes()
	seen := map[string]bool{}
	for _, route := range routes {
		seen[route.Path] = true
	}
	if !seen["/api/v2/projects"] {
		t.Fatal("expected configured api version route to be registered")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v2/projects", nil)
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if got := strings.TrimSpace(w.Header().Get("X-API-Version")); got != "v2" {
		t.Fatalf("expected response to advertise version v2, got %q", got)
	}
}

func TestConfiguredApiVersionRoutesAuthLogin(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("GATEWAY_API_SUPPORTED_VERSIONS", "v1,v2")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	router.ServeHTTP(w, req)

	if got := strings.TrimSpace(w.Header().Get("X-API-Version")); got != "v2" {
		t.Fatalf("expected response to advertise version v2, got %q", got)
	}
}

func TestUnsupportedApiVersionReturns404(t *testing.T) {
	t.Setenv("GATEWAY_ENFORCE_JWT", "false")
	t.Setenv("JWT_SECRET", "test-secret")
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v9/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected unsupported api version to return 404, got %d", w.Code)
	}
}

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
}

func (r *closeNotifyRecorder) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}
