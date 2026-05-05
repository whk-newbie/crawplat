// Package service 的单元测试：验证 Login/Register 成功/失败路径和种子账号行为。
package service

import (
	"errors"
	"testing"

	"crawler-platform/apps/iam-service/internal/repo"
)

func TestLoginReturnsTokenForSeedUser(t *testing.T) {
	svc := NewAuthService("secret", repo.NewUserRepo(true))
	token, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	svc := NewAuthService("secret", repo.NewUserRepo(true))

	_, err := svc.Login("admin", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got: %v", err)
	}
}

func TestLoginRejectsSeedAdminWhenDisabled(t *testing.T) {
	svc := NewAuthService("secret", repo.NewUserRepo(false))

	_, err := svc.Login("admin", "admin123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got: %v", err)
	}
}

func TestLoginTrimsWhitespace(t *testing.T) {
	svc := NewAuthService("secret", repo.NewUserRepo(true))
	token, err := svc.Login("  admin  ", "admin123")
	if err != nil {
		t.Fatalf("expected login success with trimmed username, got error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestRegisterCreatesUser(t *testing.T) {
	svc := NewAuthService("secret", repo.NewUserRepo(false))
	user, err := svc.Register("newuser", "password123")
	if err != nil {
		t.Fatalf("expected register success, got error: %v", err)
	}
	if user.Username != "newuser" {
		t.Fatalf("expected username newuser, got %s", user.Username)
	}

	// 注册成功后应该可以登录
	_, err = svc.Login("newuser", "password123")
	if err != nil {
		t.Fatalf("expected login success after register, got error: %v", err)
	}
}

func TestRegisterRejectsDuplicateUsername(t *testing.T) {
	svc := NewAuthService("secret", repo.NewUserRepo(true))
	_, err := svc.Register("admin", "password123")
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestRegisterRejectsEmptyFields(t *testing.T) {
	svc := NewAuthService("secret", repo.NewUserRepo(false))

	_, err := svc.Register("", "password")
	if err == nil {
		t.Fatal("expected error for empty username")
	}

	_, err = svc.Register("user", "")
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}
