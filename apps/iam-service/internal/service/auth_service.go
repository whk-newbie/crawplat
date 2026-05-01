package service

import (
	"context"
	"errors"
	"time"

	"crawler-platform/apps/iam-service/internal/model"
	"crawler-platform/apps/iam-service/internal/repo"
	"crawler-platform/packages/go-common/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	secret string
	users  repo.UserRepository
}

func NewAuthService(secret string, users repo.UserRepository) *AuthService {
	return &AuthService{
		secret: secret,
		users:  users,
	}
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.users.FindByUsername(context.Background(), username)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return auth.IssueToken(s.secret, user.Username)
}

func (s *AuthService) Register(username, password, email string) (model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	user := model.User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
		Email:        email,
		CreatedAt:    time.Now().UTC(),
	}
	return user, s.users.Create(context.Background(), user)
}
