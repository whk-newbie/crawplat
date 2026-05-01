package service

import (
	"context"
	"errors"
	"testing"

	"crawler-platform/apps/iam-service/internal/model"
	"crawler-platform/apps/iam-service/internal/repo"
	"golang.org/x/crypto/bcrypt"
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

func seedAdmin(t *testing.T) (*AuthService, *fakeUserRepo) {
	t.Helper()
	fake := newFakeUserRepo()
	svc := NewAuthService("secret", fake)
	if _, err := svc.Register("admin", "admin123", "admin@localhost"); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	return svc, fake
}

func TestLoginReturnsTokenForSeedUser(t *testing.T) {
	svc, _ := seedAdmin(t)
	token, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	svc, _ := seedAdmin(t)
	_, err := svc.Login("admin", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got: %v", err)
	}
}

func TestLoginRejectsUnknownUser(t *testing.T) {
	fake := newFakeUserRepo()
	svc := NewAuthService("secret", fake)
	_, err := svc.Login("nobody", "password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got: %v", err)
	}
}

func TestRegisterCreatesUser(t *testing.T) {
	fake := newFakeUserRepo()
	svc := NewAuthService("secret", fake)

	user, err := svc.Register("testuser", "password123", "test@example.com")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if user.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %q", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("expected email 'test@example.com', got %q", user.Email)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
		t.Fatalf("password hash does not match: %v", err)
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	fake := newFakeUserRepo()
	svc := NewAuthService("secret", fake)

	if _, err := svc.Register("admin", "admin123", ""); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	if _, err := svc.Register("admin", "other", ""); err == nil {
		t.Fatal("expected error for duplicate username")
	}
}
