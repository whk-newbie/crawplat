package service

import "testing"

func TestLoginReturnsTokenForSeedUser(t *testing.T) {
	svc := NewAuthService("secret")
	token, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}
