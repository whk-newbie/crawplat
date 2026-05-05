package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// OrgMembership 表示用户在某个组织中的成员关系及角色。
type OrgMembership struct {
	OrganizationID string `json:"id"`
	Role           string `json:"role"`
}

// UserClaims 是 JWT 的自定义声明，在标准 RegisteredClaims 之上扩展了多租户所需字段。
// OrganizationID 是用户当前活跃的组织，Organizations 是用户所属的全部组织及其角色。
type UserClaims struct {
	jwt.RegisteredClaims
	OrganizationID string          `json:"org_id,omitempty"`
	Organizations  []OrgMembership `json:"orgs,omitempty"`
}

// IssueToken 签发仅包含 subject 的基础 JWT Token（向后兼容，用于无组织上下文的场景）。
func IssueToken(secret, subject string) (string, error) {
	return IssueTokenWithClaims(secret, UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})
}

// IssueTokenWithClaims 签发包含完整 UserClaims 的 JWT Token，用于登录时嵌入组织成员信息。
func IssueTokenWithClaims(secret string, claims UserClaims) (string, error) {
	if secret == "" {
		return "", errors.New("jwt secret is required")
	}
	if claims.Subject == "" {
		return "", errors.New("jwt subject is required")
	}
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken 解析 JWT Token 并返回基础 RegisteredClaims（向后兼容）。
func ParseToken(secret, token string) (*jwt.RegisteredClaims, error) {
	claims, err := ParseUserToken(secret, token)
	if err != nil {
		return nil, err
	}
	return &claims.RegisteredClaims, nil
}

// ParseUserToken 解析 JWT Token 并返回完整的 UserClaims，包含组织成员信息。
func ParseUserToken(secret, token string) (*UserClaims, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}
	claims := &UserClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
