// Package repo 的单元测试，使用 go-sqlmock 模拟 PostgreSQL 数据库交互。
// 不依赖真实数据库，通过 mock SQL 期望验证 SQL 语句的正确性和参数绑定。
package repo

import (
	"context"
	"regexp"
	"testing"

	"crawler-platform/apps/datasource-service/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
)

// TestPostgresRepositoryCreate 验证 Create 方法执行正确的 INSERT 语句，
// 包括 SQL 模板匹配、参数值校验以及 config_json 的 JSON 序列化。
func TestPostgresRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	datasource := model.Datasource{
		ID:        "d1",
		ProjectID: "p1",
		Name:      "main",
		Type:      "postgresql",
		Readonly:  true,
		Config:    map[string]string{"schema": "public"},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO datasources (id, project_id, name, type, readonly, config_json)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)).
		WithArgs(datasource.ID, datasource.ProjectID, datasource.Name, datasource.Type, datasource.Readonly, `{"schema":"public"}`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), datasource); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}

// TestPostgresRepositoryListByProject 验证 ListByProject 执行正确的 SELECT 查询，
// 包括 WHERE 条件、ORDER BY 排序以及 config_json 的反序列化。
func TestPostgresRepositoryListByProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "type", "readonly", "config_json"}).
		AddRow("d1", "p1", "main", "redis", true, `{"db":"0"}`)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, project_id, name, type, readonly, config_json
		FROM datasources
		WHERE project_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`)).WithArgs("p1", 20, 0).WillReturnRows(rows)

	datasources, err := repo.ListByProject(context.Background(), "p1", 20, 0)
	if err != nil {
		t.Fatalf("ListByProject returned error: %v", err)
	}
	if len(datasources) != 1 || datasources[0].Config["db"] != "0" {
		t.Fatalf("unexpected datasources: %#v", datasources)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet returned error: %v", err)
	}
}
