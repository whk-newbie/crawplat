package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueAndParseToken(t *testing.T) {
	token, err := IssueToken("secret", "user-1")
	if err != nil {
		t.Fatalf("IssueToken returned error: %v", err)
	}

	claims, err := ParseToken("secret", token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}

	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
}

func TestIssueTokenRejectsEmptySecret(t *testing.T) {
	_, err := IssueToken("", "user-1")
	if err == nil {
		t.Fatal("expected IssueToken to fail for empty secret")
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	token, err := IssueToken("secret", "user-1")
	if err != nil {
		t.Fatalf("IssueToken returned error: %v", err)
	}

	_, err = ParseToken("other-secret", token)
	if err == nil {
		t.Fatal("expected ParseToken to fail for invalid signature")
	}
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "user-1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString returned error: %v", err)
	}

	_, err = ParseToken("secret", token)
	if err == nil {
		t.Fatal("expected ParseToken to fail for expired token")
	}
}
