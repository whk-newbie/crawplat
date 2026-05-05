// Package service 是 IAM 认证业务逻辑层。
// 负责用户登录认证、用户注册、JWT Token 签发。
// 不处理 HTTP 请求解析和路由注册——这些属于 api 层。
// 不处理用户数据存取——该职责通过 UserRepository 接口委托给 repo 层。
package service

import (
	"errors"
	"strings"

	"crawler-platform/apps/iam-service/internal/model"
	"crawler-platform/packages/go-common/auth"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials 表示认证凭证无效（用户名或密码错误）。
// 出于安全考虑，不区分"用户不存在"和"密码错误"的具体原因。
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrUserAlreadyExists 表示注册时用户名已被占用。
var ErrUserAlreadyExists = errors.New("user already exists")

// UserRepository 定义用户数据访问接口，由 repo 层实现。
type UserRepository interface {
	FindByUsername(username string) (model.User, error)
	Create(user model.User) error
}

// AuthService 处理登录认证和用户注册，依赖 JWT secret 和 UserRepository 接口。
type AuthService struct {
	secret string
	users  UserRepository
}

// NewAuthService 创建认证服务实例，接受 UserRepository 接口实现。
func NewAuthService(secret string, users UserRepository) *AuthService {
	return &AuthService{
		secret: secret,
		users:  users,
	}
}

// Login 验证用户名密码并签发 JWT Token。
// 输入自动做 TrimSpace 处理，成功返回 JWT Token，失败返回 ErrInvalidCredentials。
func (s *AuthService) Login(username, password string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := s.users.FindByUsername(username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	hashed := user.PasswordHash
	if hashed == "" {
		hashed = user.Password
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return auth.IssueToken(s.secret, user.Username)
}

// Register 注册新用户。
// 用户名自动做 TrimSpace 处理，不允许空用户名或空密码。
// 用户名已存在时返回 ErrUserAlreadyExists。
func (s *AuthService) Register(username, password string) (model.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return model.User{}, errors.New("username and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, errors.New("failed to hash password")
	}
	user := model.User{Username: username, PasswordHash: string(hash)}
	if err := s.users.Create(user); err != nil {
		return model.User{}, ErrUserAlreadyExists
	}
	return user, nil
}
