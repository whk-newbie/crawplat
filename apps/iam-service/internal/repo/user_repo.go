package repo

import (
	"errors"

	"crawler-platform/apps/iam-service/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo struct {
	users map[string]model.User
}

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

func (r *UserRepo) FindByUsername(username string) (model.User, error) {
	user, ok := r.users[username]
	if !ok {
		return model.User{}, ErrUserNotFound
	}
	return user, nil
}
