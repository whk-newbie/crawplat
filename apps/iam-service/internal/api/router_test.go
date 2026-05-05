// Package api 包含 IAM HTTP 路由层的 httptest 集成测试。
package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/iam-service/internal/repo"
	"crawler-platform/apps/iam-service/internal/service"
	"github.com/gin-gonic/gin"
)

func newTestRouter(seedAdmin bool) *gin.Engine {
	return NewRouter(service.NewAuthService("secret", repo.NewUserRepo(seedAdmin)))
}

func TestLoginHandlerReturnsToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTestRouter(true)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"token":`) {
		t.Fatalf("expected token response, got %s", w.Body.String())
	}
}

func TestLoginHandlerRejectsMissingUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTestRouter(true)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLoginHandlerRejectsSeedAdminWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTestRouter(false)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestLoginHandlerReturnsFriendlyErrorForBadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTestRouter(true)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid request body") {
		t.Fatalf("expected friendly error message, got %s", w.Body.String())
	}
}

func TestRegisterHandlerCreatesUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTestRouter(false)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"username":"newuser","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"username":"newuser"`) {
		t.Fatalf("expected username in response, got %s", w.Body.String())
	}
}

func TestRegisterHandlerRejectsDuplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTestRouter(true)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"username":"admin","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRegisterHandlerRejectsMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTestRouter(false)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"username":"user"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
