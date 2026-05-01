package repo

import (
	"context"
	"database/sql"

	"crawler-platform/apps/project-service/internal/model"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, project model.Project) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO projects (id, code, name)
		VALUES ($1, $2, $3)
	`, project.ID, project.Code, project.Name)
	return err
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]model.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name
		FROM projects
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
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

func (r *PostgresRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects`).Scan(&count)
	return count, err
}
