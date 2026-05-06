// Package repo 的单元测试。使用 sqlmock + miniredis 验证 Overview 聚合逻辑。
package repo

import (
	"context"
	"testing"
	"time"

	"crawler-platform/apps/monitor-service/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestOverviewRepositoryOverviewAggregatesExecutionsAndNodes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run returned error: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	rows := sqlmock.NewRows([]string{"status", "count"}).
		AddRow("pending", 2).
		AddRow("running", 1).
		AddRow("succeeded", 5).
		AddRow("failed", 3)
		mock.ExpectQuery(`SELECT status, COUNT\(\*\) FROM executions WHERE \(\$1 = '' OR organization_id = \$1\) GROUP BY status`).
			WithArgs("").
		WillReturnRows(rows)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM nodes WHERE \(\$1 = '' OR organization_id = \$1\)`).
			WithArgs("").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	if err := client.Set(context.Background(), "nodes:node-1", "ok", time.Minute).Err(); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if err := client.SAdd(context.Background(), "nodes:online", "node-1", "node-stale").Err(); err != nil {
		t.Fatalf("SAdd returned error: %v", err)
	}
	if err := client.Set(context.Background(), "nodes:node-legacy", "legacy", time.Minute).Err(); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	repository := NewOverviewRepository(db, client)

	overview, err := repository.Overview(context.Background(), "")
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}

	expected := model.Overview{
		Executions: model.ExecutionSummary{
			Total:     11,
			Pending:   2,
			Running:   1,
			Succeeded: 5,
			Failed:    3,
		},
		Nodes: model.NodeSummary{
			Total:   2,
			Online:  1,
			Offline: 1,
		},
	}
	if overview != expected {
		t.Fatalf("expected overview %+v, got %+v", expected, overview)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestOverviewRepositoryOverviewDoesNotReturnNegativeOfflineNodes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run returned error: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	rows := sqlmock.NewRows([]string{"status", "count"}).AddRow("running", 1)
		mock.ExpectQuery(`SELECT status, COUNT\(\*\) FROM executions WHERE \(\$1 = '' OR organization_id = \$1\) GROUP BY status`).
			WithArgs("").
		WillReturnRows(rows)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM nodes WHERE \(\$1 = '' OR organization_id = \$1\)`).
			WithArgs("").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	if err := client.Set(context.Background(), "nodes:node-1", "ok", time.Minute).Err(); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if err := client.Set(context.Background(), "nodes:node-2", "ok", time.Minute).Err(); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if err := client.SAdd(context.Background(), "nodes:online", "node-1", "node-2").Err(); err != nil {
		t.Fatalf("SAdd returned error: %v", err)
	}

	repository := NewOverviewRepository(db, client)
	overview, err := repository.Overview(context.Background(), "")
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}

	if overview.Nodes.Total != 1 || overview.Nodes.Online != 2 || overview.Nodes.Offline != 0 {
		t.Fatalf("unexpected node summary: %+v", overview.Nodes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
