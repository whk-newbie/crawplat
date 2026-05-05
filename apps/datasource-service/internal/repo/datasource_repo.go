// Package repo 负责数据源配置的持久化存储，当前基于 PostgreSQL 实现。
// 本文件仅处理数据源的创建、按项目列表查询和按 ID 查询——不负责更新、删除或任何业务逻辑。
// 与谁交互：直接与 PostgreSQL（datasources 表）交互，通过 database/sql 标准库执行 SQL。
package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"crawler-platform/apps/datasource-service/internal/model"
)

// PostgresRepository 使用 PostgreSQL 作为数据源配置的持久化存储。
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository 创建基于 PostgreSQL 的数据源仓库实例。
// db 必须为已建立连接的 *sql.DB 对象，由调用方（main.go）注入。
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create 向 datasources 表中插入一条新的数据源记录。
// Config 字段以 JSON 格式序列化后存入 config_json 列。
func (r *PostgresRepository) Create(ctx context.Context, datasource model.Datasource) error {
	configJSON, err := json.Marshal(datasource.Config)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO datasources (id, project_id, name, type, readonly, config_json)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, datasource.ID, datasource.ProjectID, datasource.Name, datasource.Type, datasource.Readonly, string(configJSON))
	return err
}

// ListByProject 按 projectID 查询该项目的所有数据源，按创建时间降序排列。
// 返回的 []model.Datasource 中的 Config 字段从 JSON 反序列化后返回。
func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string) ([]model.Datasource, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, name, type, readonly, config_json
		FROM datasources
		WHERE project_id = $1
		ORDER BY created_at DESC, id DESC
	`, projectID)
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

// Get 按 ID 查询单个数据源。
// 返回三个值：数据源实体、是否找到、以及可能的错误。
// 如果数据源不存在，返回 (empty Datasource, false, nil) 而非错误——调用方据此判断 404。
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

// scanner 抽象 *sql.Row 和 *sql.Rows 共有的 Scan 方法，便于 scanDatasource 复用。
type scanner interface {
	Scan(dest ...any) error
}

// scanDatasource 从数据库行扫描结果中反序列化 model.Datasource。
// Config 列存储为 JSONB/JSON，在此反序列化为 map[string]string；
// 若 config_json 为 NULL 则初始化为空 map，避免调用方空指针。
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
