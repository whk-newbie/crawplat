package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/iam-service/internal/model"
	"crawler-platform/apps/iam-service/internal/repo"
	"crawler-platform/apps/iam-service/internal/service"
	"github.com/gin-gonic/gin"
)

type fakeUserRepo struct {
	users map[string]model.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]model.User)}
}

func (r *fakeUserRepo) FindByUsername(_ context.Context, username string) (model.User, error) {
	u, ok := r.users[username]
	if !ok {
		return model.User{}, repo.ErrUserNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) Create(_ context.Context, user model.User) error {
	if _, exists := r.users[user.Username]; exists {
		return errors.New("duplicate username")
	}
	r.users[user.Username] = user
	return nil
}

func newSeededService(t *testing.T) *service.AuthService {
	t.Helper()
	fake := newFakeUserRepo()
	svc := service.NewAuthService("secret", fake)
	if _, err := svc.Register("admin", "admin123", "admin@localhost"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return svc
}

func TestLoginHandlerReturnsToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(newSeededService(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"token":`) {
		t.Fatalf("expected token response, got %s", w.Body.String())
	}
}

func TestLoginHandlerRejectsMissingUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(newSeededService(t))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLoginHandlerRejectsUnknownUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := newFakeUserRepo()
	router := NewRouter(service.NewAuthService("secret", fake))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestRegisterHandlerCreatesUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := newFakeUserRepo()
	router := NewRouter(service.NewAuthService("secret", fake))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"username":"newuser","password":"password123","email":"new@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"username":"newuser"`) {
		t.Fatalf("expected user in response, got %s", w.Body.String())
	}
}

func TestRegisterHandlerRejectsShortPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := newFakeUserRepo()
	router := NewRouter(service.NewAuthService("secret", fake))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"username":"user","password":"short"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
