package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"crawler-platform/apps/spider-service/internal/model"
	"github.com/google/uuid"
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

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO spiders (id, project_id, name, language, runtime, image, command)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, spider.ID, spider.ProjectID, spider.Name, spider.Language, spider.Runtime, spider.Image, string(command))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO spider_versions (id, spider_id, version, registry_auth_ref, image, command)
		VALUES ($1, $2, 1, '', $3, $4)
	`, uuid.NewString(), spider.ID, spider.Image, string(command))
	if err != nil {
		return err
	}

	return tx.Commit()
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

func (r *PostgresRepository) CreateVersion(ctx context.Context, spiderID, registryAuthRef, image string, command []string) (model.SpiderVersion, error) {
	commandRaw, err := json.Marshal(command)
	if err != nil {
		return model.SpiderVersion{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.SpiderVersion{}, err
	}
	defer tx.Rollback()

	var lockedID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM spiders WHERE id = $1 FOR UPDATE`, spiderID).Scan(&lockedID); err != nil {
		return model.SpiderVersion{}, err
	}

	var currentMax int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM spider_versions WHERE spider_id = $1`, spiderID).Scan(&currentMax); err != nil {
		return model.SpiderVersion{}, err
	}
	nextVersion := currentMax + 1

	created := model.SpiderVersion{
		ID:              uuid.NewString(),
		SpiderID:        spiderID,
		Version:         nextVersion,
		RegistryAuthRef: registryAuthRef,
		Image:           image,
		Command:         append([]string(nil), command...),
	}
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO spider_versions (id, spider_id, version, registry_auth_ref, image, command)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`, created.ID, created.SpiderID, created.Version, created.RegistryAuthRef, created.Image, string(commandRaw)).Scan(&created.CreatedAt); err != nil {
		return model.SpiderVersion{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE spiders
		SET image = $2, command = $3
		WHERE id = $1
	`, spiderID, image, string(commandRaw)); err != nil {
		return model.SpiderVersion{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.SpiderVersion{}, err
	}
	created.CreatedAt = created.CreatedAt.UTC()
	return created, nil
}

func (r *PostgresRepository) ListVersions(ctx context.Context, spiderID string) ([]model.SpiderVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, spider_id, version, registry_auth_ref, image, command, created_at
		FROM spider_versions
		WHERE spider_id = $1
		ORDER BY version DESC
	`, spiderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []model.SpiderVersion
	for rows.Next() {
		var version model.SpiderVersion
		var commandRaw []byte
		var createdAt time.Time
		if err := rows.Scan(&version.ID, &version.SpiderID, &version.Version, &version.RegistryAuthRef, &version.Image, &commandRaw, &createdAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(commandRaw, &version.Command); err != nil {
			return nil, err
		}
		version.CreatedAt = createdAt.UTC()
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}

func (r *PostgresRepository) ListRegistryAuthRefsByProject(ctx context.Context, projectID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT sv.registry_auth_ref
		FROM spider_versions sv
		INNER JOIN spiders s ON s.id = sv.spider_id
		WHERE s.project_id = $1 AND sv.registry_auth_ref <> ''
		ORDER BY sv.registry_auth_ref ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []string
	for rows.Next() {
		var ref string
		if err := rows.Scan(&ref); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return refs, nil
}
