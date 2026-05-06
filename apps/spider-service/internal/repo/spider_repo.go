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
			INSERT INTO spiders (id, project_id, name, language, runtime, image, command, organization_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, spider.ID, spider.ProjectID, spider.Name, spider.Language, spider.Runtime, spider.Image, string(command), nullIfEmpty(spider.OrganizationID))
	return err
}

func (r *PostgresRepository) ListByProject(ctx context.Context, orgID, projectID string, limit, offset int) ([]model.Spider, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
			SELECT id, project_id, name, language, runtime, image, command
			FROM spiders
			WHERE ($1 = '' OR organization_id = $1) AND project_id = $2
			ORDER BY created_at DESC, id DESC
			LIMIT $3 OFFSET $4
		`, orgID, projectID, limit, offset)
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

func (r *PostgresRepository) CreateVersion(ctx context.Context, version model.SpiderVersion) error {
	command, err := json.Marshal(version.Command)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
			INSERT INTO spider_versions (id, spider_id, version, image, registry_auth_ref, command, organization_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, version.ID, version.SpiderID, version.Version, version.Image, version.RegistryAuthRef, string(command), nullIfEmpty(version.OrganizationID))
	return err
}

func (r *PostgresRepository) ListVersions(ctx context.Context, spiderID string) ([]model.SpiderVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
			SELECT id, spider_id, version, image, registry_auth_ref, command
			FROM spider_versions
			WHERE spider_id = $1
			ORDER BY created_at DESC, id DESC
		`, spiderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []model.SpiderVersion
	for rows.Next() {
		var v model.SpiderVersion
		var commandRaw []byte
		if err := rows.Scan(&v.ID, &v.SpiderID, &v.Version, &v.Image, &v.RegistryAuthRef, &commandRaw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(commandRaw, &v.Command); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}

func (r *PostgresRepository) ListRegistryAuthRefs(ctx context.Context, projectID string) ([]model.RegistryAuthRef, error) {
	rows, err := r.db.QueryContext(ctx, `
			SELECT DISTINCT sv.registry_auth_ref, ''
			FROM spider_versions sv
			JOIN spiders s ON s.id = sv.spider_id
			WHERE s.project_id = $1 AND sv.registry_auth_ref != ''
		`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []model.RegistryAuthRef
	for rows.Next() {
		var ref model.RegistryAuthRef
		if err := rows.Scan(&ref.Ref, &ref.Server); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return refs, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
