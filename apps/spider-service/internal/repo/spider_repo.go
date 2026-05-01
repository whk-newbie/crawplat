package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"crawler-platform/apps/spider-service/internal/model"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, spider model.Spider) error {
	command, err := json.Marshal(spider.Command)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO spiders (id, project_id, name, language, runtime, image, command)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, spider.ID, spider.ProjectID, spider.Name, spider.Language, spider.Runtime, spider.Image, string(command))
	return err
}

func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string, limit, offset int) ([]model.Spider, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, name, language, runtime, image, command
		FROM spiders
		WHERE project_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spiders []model.Spider
	for rows.Next() {
		var spider model.Spider
		var commandRaw []byte
		if err := rows.Scan(&spider.ID, &spider.ProjectID, &spider.Name, &spider.Language, &spider.Runtime, &spider.Image, &commandRaw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(commandRaw, &spider.Command); err != nil {
			return nil, err
		}
		spiders = append(spiders, spider)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return spiders, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (model.Spider, bool, error) {
	var spider model.Spider
	var commandRaw []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, language, runtime, image, command
		FROM spiders
		WHERE id = $1
	`, id).Scan(&spider.ID, &spider.ProjectID, &spider.Name, &spider.Language, &spider.Runtime, &spider.Image, &commandRaw)
	if err == sql.ErrNoRows {
		return model.Spider{}, false, nil
	}
	if err != nil {
		return model.Spider{}, false, err
	}
	if err := json.Unmarshal(commandRaw, &spider.Command); err != nil {
		return model.Spider{}, false, err
	}
	return spider, true, nil
}

func (r *PostgresRepository) CountByProject(ctx context.Context, projectID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM spiders WHERE project_id = $1`, projectID).Scan(&count)
	return count, err
}
