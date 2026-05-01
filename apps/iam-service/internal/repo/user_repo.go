package repo

import (
	"context"
	"database/sql"
	"errors"

	"crawler-platform/apps/iam-service/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (model.User, error)
	Create(ctx context.Context, user model.User) error
}

type PostgresUserRepo struct {
	db *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) FindByUsername(ctx context.Context, username string) (model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, email, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	return u, err
}

func (r *PostgresUserRepo) Create(ctx context.Context, user model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, username, password_hash, email, created_at) VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Username, user.PasswordHash, user.Email, user.CreatedAt,
	)
	return err
}
