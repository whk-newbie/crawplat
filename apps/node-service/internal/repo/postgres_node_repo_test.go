// 文件职责：PostgreSQL 仓库的单元测试。
// 测试范围：
//   - UpsertCatalog/ListCatalog：节点写入与列表查询
//   - GetByID：按 ID 查找节点及 not-found 错误处理
//   - ListHeartbeatHistory：心跳历史查询（多行返回）
//   - ListRecentExecutions：执行记录过滤（基础查询、状态过滤、时间范围过滤）
//   - 接口满足检查（编译时验证 PostgresNodeRepository 实现了 service.CatalogRepository 接口）
// 使用 go-sqlmock 模拟 PostgreSQL 数据库，无需真实数据库实例。
package repo

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"crawler-platform/apps/node-service/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresNodeRepositoryUpsertAndListCatalog(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresNodeRepository(db)
	seenAt := time.Unix(1700000000, 0).UTC()

	mock.ExpectExec(`INSERT INTO nodes`).
		WithArgs("node-1", "node-1", `["docker","python"]`, seenAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO node_heartbeats`).
		WithArgs("node-1", seenAt, `["docker","python"]`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	node, err := repo.UpsertCatalog(context.Background(), "", "node-1", []string{"docker", "python"}, seenAt)
	if err != nil {
		t.Fatalf("UpsertCatalog returned error: %v", err)
	}
	if node.ID != "node-1" || node.Name != "node-1" {
		t.Fatalf("unexpected node: %#v", node)
	}

	rows := sqlmock.NewRows([]string{"id", "name", "capabilities_json", "last_seen_at", "organization_id"}).
		AddRow("node-1", "node-1", `["docker","python"]`, seenAt, "")
	mock.ExpectQuery(`SELECT id, name, capabilities_json, last_seen_at, organization_id FROM nodes WHERE \(\$1 = '' OR organization_id = \$1\) ORDER BY name ASC`).
			WithArgs("").
		WillReturnRows(rows)

	nodes, err := repo.ListCatalog(context.Background(), "")
	if err != nil {
		t.Fatalf("ListCatalog returned error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].ID != "node-1" || len(nodes[0].Capabilities) != 2 {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}

	mock.ExpectQuery(`SELECT id, name, capabilities_json, last_seen_at, organization_id FROM nodes WHERE id = \$1`).
		WithArgs("node-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "capabilities_json", "last_seen_at", "organization_id"}).
			AddRow("node-1", "node-1", `["docker","python"]`, seenAt, ""))

	detailNode, err := repo.GetByID(context.Background(), "", "node-1")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if detailNode.ID != "node-1" || len(detailNode.Capabilities) != 2 {
		t.Fatalf("unexpected detail node: %#v", detailNode)
	}

	mock.ExpectQuery(`SELECT seen_at, capabilities_json FROM node_heartbeats WHERE node_id = \$1 ORDER BY seen_at DESC LIMIT \$2`).
		WithArgs("node-1", 2).
		WillReturnRows(sqlmock.NewRows([]string{"seen_at", "capabilities_json"}).
			AddRow(seenAt, `["docker","python"]`).
			AddRow(seenAt.Add(-time.Minute), `["docker"]`))

	history, err := repo.ListHeartbeatHistory(context.Background(), "node-1", 2)
	if err != nil {
		t.Fatalf("ListHeartbeatHistory returned error: %v", err)
	}
	if len(history) != 2 || history[0].SeenAt.IsZero() {
		t.Fatalf("unexpected heartbeat history: %#v", history)
	}

	startedAt := seenAt.Add(10 * time.Second)
	finishedAt := seenAt.Add(40 * time.Second)
	mock.ExpectQuery(`SELECT id, project_id, spider_id, status, trigger_source, created_at, started_at, finished_at FROM executions WHERE node_id = \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs("node-1", 3, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "spider_id", "status", "trigger_source", "created_at", "started_at", "finished_at"}).
			AddRow("exec-1", "project-1", "spider-1", "succeeded", "manual", seenAt, startedAt, finishedAt).
			AddRow("exec-2", "project-1", "spider-2", "failed", "scheduled", seenAt.Add(-time.Minute), nil, nil))

	executions, err := repo.ListRecentExecutions(context.Background(), "", "node-1", service.ExecutionQuery{
		Limit:  3,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListRecentExecutions returned error: %v", err)
	}
	if len(executions) != 2 || executions[0].ID != "exec-1" || executions[1].ID != "exec-2" {
		t.Fatalf("unexpected recent executions: %#v", executions)
	}
	if executions[0].StartedAt == nil || executions[0].FinishedAt == nil {
		t.Fatalf("expected startedAt/finishedAt on first execution, got %#v", executions[0])
	}
	if executions[1].StartedAt != nil || executions[1].FinishedAt != nil {
		t.Fatalf("expected nil startedAt/finishedAt on second execution, got %#v", executions[1])
	}

	mock.ExpectQuery(`SELECT id, project_id, spider_id, status, trigger_source, created_at, started_at, finished_at FROM executions WHERE node_id = \$1 AND status = \$2 ORDER BY created_at DESC LIMIT \$3 OFFSET \$4`).
		WithArgs("node-1", "failed", 5, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "spider_id", "status", "trigger_source", "created_at", "started_at", "finished_at"}))

	_, err = repo.ListRecentExecutions(context.Background(), "", "node-1", service.ExecutionQuery{
		Limit:  5,
		Offset: 10,
		Status: "failed",
	})
	if err != nil {
		t.Fatalf("ListRecentExecutions with status filter returned error: %v", err)
	}

	from := seenAt.Add(-2 * time.Hour)
	to := seenAt.Add(2 * time.Hour)
	mock.ExpectQuery(`SELECT id, project_id, spider_id, status, trigger_source, created_at, started_at, finished_at FROM executions WHERE node_id = \$1 AND status = \$2 AND created_at >= \$3 AND created_at <= \$4 ORDER BY created_at DESC LIMIT \$5 OFFSET \$6`).
		WithArgs("node-1", "succeeded", from, to, 20, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "spider_id", "status", "trigger_source", "created_at", "started_at", "finished_at"}))

	_, err = repo.ListRecentExecutions(context.Background(), "", "node-1", service.ExecutionQuery{
		Limit:  20,
		Offset: 2,
		Status: "succeeded",
		From:   &from,
		To:     &to,
	})
	if err != nil {
		t.Fatalf("ListRecentExecutions with time range returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPostgresNodeRepositoryGetByIDNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresNodeRepository(db)
	mock.ExpectQuery(`SELECT id, name, capabilities_json, last_seen_at, organization_id FROM nodes WHERE id = \$1`).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), "", "missing")
	if !errors.Is(err, service.ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

var _ service.CatalogRepository = (*PostgresNodeRepository)(nil)
