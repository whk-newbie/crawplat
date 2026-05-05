// Package service 是 IAM 认证业务逻辑层。
// 负责用户登录认证、用户注册、JWT Token 签发、组织创建与查询。
// 不处理 HTTP 请求解析和路由注册——这些属于 api 层。
// 不处理用户数据存取——该职责通过 UserRepository / OrgRepository 接口委托给 repo 层。
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

// LoginResult 包含登录成功后的 JWT Token 和用户所属组织列表。
type LoginResult struct {
	Token       string               `json:"token"`
	Memberships []model.OrgMembership `json:"organizations"`
}

// OrgRepository 定义组织数据访问接口，由 repo 层实现。
type OrgRepository interface {
	CreateOrganization(name, slug, createdByUserID string) (model.Organization, error)
	FindMembershipsByUser(username string) ([]model.OrgMembership, error)
}

// AuthService 处理登录认证和用户注册，依赖 JWT secret、UserRepository 和 OrgRepository。
type AuthService struct {
	secret string
	users  UserRepository
	orgs   OrgRepository
}

// NewAuthService 创建认证服务实例。
func NewAuthService(secret string, users UserRepository, orgs OrgRepository) *AuthService {
	return &AuthService{
		secret: secret,
		users:  users,
		orgs:   orgs,
	}
}

// Login 验证用户名密码并签发包含组织信息的 JWT Token。
// 成功返回 LoginResult（token + 组织列表），失败返回 ErrInvalidCredentials。
func (s *AuthService) Login(username, password string) (LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	user, err := s.users.FindByUsername(username)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	hashed := user.PasswordHash
	if hashed == "" {
		hashed = user.Password
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	memberships, err := s.orgs.FindMembershipsByUser(username)
	if err != nil {
		return LoginResult{}, err
	}

	claims := auth.UserClaims{
		Organizations: make([]auth.OrgMembership, len(memberships)),
	}
	for i, m := range memberships {
		claims.Organizations[i] = auth.OrgMembership{
			OrganizationID: m.OrganizationID,
			Role:           m.Role,
		}
		// 第一个组织设为当前活跃组织
		if i == 0 {
			claims.OrganizationID = m.OrganizationID
		}
	}
	claims.Subject = user.Username

	token, err := auth.IssueTokenWithClaims(s.secret, claims)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token, Memberships: memberships}, nil
}

// Register 注册新用户并自动创建个人组织。
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

	// 自动创建个人组织
	_, err = s.orgs.CreateOrganization(username+"'s Workspace", username, username)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
