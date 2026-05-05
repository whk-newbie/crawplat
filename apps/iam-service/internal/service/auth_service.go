package service

import (
	"errors"

	"crawler-platform/apps/iam-service/internal/repo"
	"crawler-platform/packages/go-common/auth"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	secret string
	users  *repo.UserRepo
}

func NewAuthService(secret string, enableSeedAdmin bool) *AuthService {
	return &AuthService{
		secret: secret,
		users:  repo.NewUserRepo(enableSeedAdmin),
	}
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if user.Password != password {
		return "", ErrInvalidCredentials
	}
	return auth.IssueToken(s.secret, user.Username)
}
