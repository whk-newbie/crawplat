package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/iam-service/internal/service"
	"github.com/gin-gonic/gin"
)

func TestLoginHandlerReturnsToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewAuthService("secret", true))
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

	router := NewRouter(service.NewAuthService("secret", true))
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

	router := NewRouter(service.NewAuthService("secret", false))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
