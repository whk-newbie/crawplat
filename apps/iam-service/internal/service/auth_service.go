// Package service 是 IAM 认证业务逻辑层。
// 负责用户登录认证、JWT Token 签发。
// 不处理 HTTP 请求解析和路由注册——这些属于 api 层。
// 不处理用户数据存取——该职责属于 repo 层。
package service

import (
	"errors"

	"crawler-platform/apps/iam-service/internal/repo"
	"crawler-platform/packages/go-common/auth"
)

// ErrInvalidCredentials 表示认证凭证无效（用户名或密码错误）。
// 出于安全考虑，不区分"用户不存在"和"密码错误"的具体原因。
var ErrInvalidCredentials = errors.New("invalid credentials")

// AuthService 处理登录认证，依赖 JWT secret 和用户仓储。
type AuthService struct {
	secret string
	users  *repo.UserRepo
}

// NewAuthService 创建认证服务实例。enableSeedAdmin 控制是否预置管理员账号。
func NewAuthService(secret string, enableSeedAdmin bool) *AuthService {
	return &AuthService{
		secret: secret,
		users:  repo.NewUserRepo(enableSeedAdmin),
	}
}

// Login 验证用户名密码并签发 JWT Token。
// 输入：username（登录名）、password（明文密码，MVP 阶段）。
// 成功返回签发的 JWT Token 字符串，失败返回 ErrInvalidCredentials。
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
