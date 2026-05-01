package repo

import (
	"context"
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

	node, err := repo.UpsertCatalog(context.Background(), "node-1", []string{"docker", "python"}, seenAt)
	if err != nil {
		t.Fatalf("UpsertCatalog returned error: %v", err)
	}
	if node.ID != "node-1" || node.Name != "node-1" {
		t.Fatalf("unexpected node: %#v", node)
	}

	rows := sqlmock.NewRows([]string{"id", "name", "capabilities_json", "last_seen_at"}).
		AddRow("node-1", "node-1", `["docker","python"]`, seenAt)
	mock.ExpectQuery(`SELECT id, name, capabilities_json, last_seen_at FROM nodes ORDER BY name ASC`).
		WillReturnRows(rows)

	nodes, err := repo.ListCatalog(context.Background())
	if err != nil {
		t.Fatalf("ListCatalog returned error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].ID != "node-1" || len(nodes[0].Capabilities) != 2 {
		t.Fatalf("unexpected nodes: %#v", nodes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

var _ service.CatalogRepository = (*PostgresNodeRepository)(nil)
