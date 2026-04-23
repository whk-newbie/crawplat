package repo

import (
	"context"
	"encoding/json"
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
	mock.ExpectQuery(`SELECT status, COUNT\(\*\) FROM executions GROUP BY status`).
		WillReturnRows(rows)

	liveNode := struct {
		ID           string    `json:"id"`
		Name         string    `json:"name"`
		Status       string    `json:"status"`
		Capabilities []string  `json:"capabilities"`
		LastSeenAt   time.Time `json:"lastSeenAt"`
	}{
		ID:           "node-1",
		Name:         "node-1",
		Status:       "online",
		Capabilities: []string{"docker"},
		LastSeenAt:   time.Unix(1700000000, 0).UTC(),
	}
	payload, err := json.Marshal(liveNode)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	if err := client.Set(context.Background(), "nodes:node-1", payload, time.Minute).Err(); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if err := client.SAdd(context.Background(), "nodes:online", "node-1", "node-stale").Err(); err != nil {
		t.Fatalf("SAdd returned error: %v", err)
	}

	repository := NewOverviewRepository(db, client)

	overview, err := repository.Overview(context.Background())
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
			Total:   1,
			Online:  1,
			Offline: 0,
		},
	}
	if overview != expected {
		t.Fatalf("expected overview %+v, got %+v", expected, overview)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
