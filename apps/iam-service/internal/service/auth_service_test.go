// Package service 的单元测试：验证 Login 成功/失败路径和种子账号行为。
package service

import (
	"errors"
	"testing"
)

func TestLoginReturnsTokenForSeedUser(t *testing.T) {
	svc := NewAuthService("secret", true)
	token, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	svc := NewAuthService("secret", true)

	_, err := svc.Login("admin", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got: %v", err)
	}
}

func TestLoginRejectsSeedAdminWhenDisabled(t *testing.T) {
	svc := NewAuthService("secret", false)

	_, err := svc.Login("admin", "admin123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got: %v", err)
	}
}
