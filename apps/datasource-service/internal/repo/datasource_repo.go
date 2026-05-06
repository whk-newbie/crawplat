package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"crawler-platform/apps/datasource-service/internal/model"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, datasource model.Datasource) error {
	configJSON, err := json.Marshal(datasource.Config)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO datasources (id, project_id, name, type, readonly, config_json, organization_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, datasource.ID, datasource.ProjectID, datasource.Name, datasource.Type, datasource.Readonly, string(configJSON), nullIfEmpty(datasource.OrganizationID))
	return err
}

func (r *PostgresRepository) ListByProject(ctx context.Context, orgID, projectID string, limit, offset int) ([]model.Datasource, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, name, type, readonly, config_json
		FROM datasources
		WHERE ($1 = '' OR organization_id = $1) AND project_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`, orgID, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasources []model.Datasource
	for rows.Next() {
		datasource, err := scanDatasource(rows)
		if err != nil {
			return nil, err
		}
		datasources = append(datasources, datasource)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return datasources, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (model.Datasource, bool, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, type, readonly, config_json
		FROM datasources
		WHERE id = $1
	`, id)

	datasource, err := scanDatasource(row)
	if err == sql.ErrNoRows {
		return model.Datasource{}, false, nil
	}
	if err != nil {
		return model.Datasource{}, false, err
	}
	return datasource, true, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanDatasource(scan scanner) (model.Datasource, error) {
	var datasource model.Datasource
	var configRaw []byte
	if err := scan.Scan(&datasource.ID, &datasource.ProjectID, &datasource.Name, &datasource.Type, &datasource.Readonly, &configRaw); err != nil {
		return model.Datasource{}, err
	}
	if err := json.Unmarshal(configRaw, &datasource.Config); err != nil {
		return model.Datasource{}, err
	}
	if datasource.Config == nil {
		datasource.Config = map[string]string{}
	}
	return datasource, nil
}
