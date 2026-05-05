// 该文件为 Spider 的 PostgreSQL 持久化层，封装对 spiders 表的 CRUD 操作。
// 因为 PostgreSQL 不支持直接存储 Go 切片，将 Command 字段序列化为 JSON 字节数组存入/读出。
// 依赖 model.Spider 结构体，不包含业务校验逻辑（校验由 service 层负责）。
package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"crawler-platform/apps/spider-service/internal/model"
)

// PostgresRepository 封装 *sql.DB，实现 Repository 接口，提供对 spiders 表的数据库操作。
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository 基于已有的数据库连接创建 PostgresRepository 实例。
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create 将一条 Spider 记录写入 spiders 表。Command 切片会被 JSON 序列化为字符串后存储。
// 返回 error 表示数据库写入失败或 JSON 序列化失败。
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

// ListByProject 查询指定项目下的所有 Spider，按创建时间降序排列（created_at DESC, id DESC）。
// 返回的 Command 字段从 JSON 字节数组反序列化为 Go 切片。
func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string) ([]model.Spider, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, name, language, runtime, image, command
		FROM spiders
		WHERE project_id = $1
		ORDER BY created_at DESC, id DESC
	`, projectID)
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

// Get 按主键 id 查询单条 Spider 记录。
// 返回值：找到时返回 (spider, true, nil)；未找到时返回 (empty, false, nil)；数据库错误返回 (empty, false, error)。
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
