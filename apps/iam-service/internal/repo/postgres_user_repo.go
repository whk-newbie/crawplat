// Postgres 用户仓储实现。
// 将用户数据持久化到 PostgreSQL users 表，支持用户名唯一约束和 bcrypt 密码哈希。
// 与 MemoryUserRepo 实现相同的 UserRepository 接口，生产环境优先使用本实现。
package repo

import (
	"database/sql"
	"fmt"
	"strings"

	"crawler-platform/apps/iam-service/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

// PostgresUserRepo 使用 PostgreSQL users 表存取用户数据。
type PostgresUserRepo struct {
	db *sql.DB
}

// NewPostgresUserRepo 创建 Postgres 用户仓储。
// enableSeedAdmin 为 true 时，在 users 表为空的情况下预置 admin/admin123。
func NewPostgresUserRepo(db *sql.DB, enableSeedAdmin bool) (*PostgresUserRepo, error) {
	repo := &PostgresUserRepo{db: db}
	if enableSeedAdmin {
		if err := repo.seedAdmin(); err != nil {
			return nil, fmt.Errorf("seed admin: %w", err)
		}
	}
	return repo, nil
}

// FindByUsername 按用户名查询用户，未找到返回 ErrUserNotFound。
func (r *PostgresUserRepo) FindByUsername(username string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(
		`SELECT id, username, password_hash FROM users WHERE username = $1`,
		strings.TrimSpace(username),
	).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err == sql.ErrNoRows {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}

// Create 创建新用户。用户名已存在时返回 ErrUserAlreadyExists。
func (r *PostgresUserRepo) Create(user model.User) error {
	hash := user.PasswordHash
	if hash == "" {
		hash = user.Password
	}
	_, err := r.db.Exec(
		`INSERT INTO users (id, username, password_hash) VALUES (gen_random_uuid()::text, $1, $2)`,
		strings.TrimSpace(user.Username),
		hash,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %s", ErrUserAlreadyExists, user.Username)
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// seedAdmin 在 users 表为空时预置管理员账号（admin/admin123），密码使用 bcrypt 哈希存储。
func (r *PostgresUserRepo) seedAdmin() error {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	return r.Create(model.User{
		Username:     "admin",
		PasswordHash: string(hash),
	})
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unique") ||
		strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "23505")
}
