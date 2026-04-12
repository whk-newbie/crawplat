package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func IssueToken(secret, subject string) (string, error) {
	if secret == "" {
		return "", errors.New("jwt secret is required")
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}).SignedString([]byte(secret))
}

func ParseToken(secret, token string) (*jwt.RegisteredClaims, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	return claims, err
}
