// Package repo 是 Project 服务的 PostgreSQL 持久层。
// 负责 projects 表的 CRUD，不包含业务逻辑——校验和 ID 生成属于 service 层。
// 依赖 database/sql + pgx driver，不依赖 ORM。
package repo

import (
	"context"
	"database/sql"

	"crawler-platform/apps/project-service/internal/model"
)

// PostgresRepository 基于 *sql.DB 的 Project 持久化实现。
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository 创建 PostgreSQL 仓储实例。
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create 向 projects 表插入一条记录。
func (r *PostgresRepository) Create(ctx context.Context, project model.Project) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO projects (id, code, name)
		VALUES ($1, $2, $3)
	`, project.ID, project.Code, project.Name)
	return err
}

// List 查询所有项目，按 created_at DESC, id DESC 排序。
// rows.Err() 检查确保迭代过程中无被忽略的错误。
func (r *PostgresRepository) List(ctx context.Context) ([]model.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name
		FROM projects
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var project model.Project
		if err := rows.Scan(&project.ID, &project.Code, &project.Name); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}
