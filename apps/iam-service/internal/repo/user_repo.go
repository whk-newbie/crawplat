// Package repo 是用户仓储层，负责用户数据的存取。
// 当前采用内存存储（map），不依赖外部数据库。
// 仅做数据存取，不做业务校验——认证逻辑属于 service 层。
package repo

import (
	"errors"
	"fmt"

	"crawler-platform/apps/iam-service/internal/model"
)

// ErrUserNotFound 表示用户不存在。service 层在 Login 失败时会统一返回
// ErrInvalidCredentials，不直接暴露此错误，避免用户枚举攻击。
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists 表示用户名已被注册。
var ErrUserAlreadyExists = errors.New("user already exists")

// UserRepo 是内存用户仓储，非并发安全。
type UserRepo struct {
	users map[string]model.User
}

// NewUserRepo 创建用户仓储。enableSeedAdmin 为 true 时预置 admin/admin123。
func NewUserRepo(enableSeedAdmin bool) *UserRepo {
	users := make(map[string]model.User)
	if enableSeedAdmin {
		users["admin"] = model.User{
			Username: "admin",
			Password: "admin123",
		}
	}

	return &UserRepo{users: users}
}

// FindByUsername 按用户名查找用户，未找到返回 ErrUserNotFound。
func (r *UserRepo) FindByUsername(username string) (model.User, error) {
	user, ok := r.users[username]
	if !ok {
		return model.User{}, ErrUserNotFound
	}
	return user, nil
}

// Create 创建新用户，用户名已存在时返回 ErrUserAlreadyExists。
func (r *UserRepo) Create(user model.User) error {
	if _, ok := r.users[user.Username]; ok {
		return fmt.Errorf("%w: %s", ErrUserAlreadyExists, user.Username)
	}
	r.users[user.Username] = user
	return nil
}
