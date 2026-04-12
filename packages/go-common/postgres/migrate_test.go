package postgres

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crawler-platform/packages/go-common/config"
)

func TestConfigRequiresPersistenceSettings(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://crawler:crawler@postgres:5432/crawler?sslmode=disable")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("MONGO_URI", "mongodb://mongo:27017")

	cfg := config.Load()

	if cfg.PostgresDSN == "" || cfg.RedisAddr == "" || cfg.MongoURI == "" {
		t.Fatalf("expected persistence config to be populated: %#v", cfg)
	}
}

func TestPhase2MigrationAssetsExist(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	sqlDir := filepath.Join(root, "deploy", "migrations", "postgres")

	firstMigration, err := os.ReadFile(filepath.Join(sqlDir, "001_phase2_core_tables.sql"))
	if err != nil {
		t.Fatalf("ReadFile core tables returned error: %v", err)
	}
	secondMigration, err := os.ReadFile(filepath.Join(sqlDir, "002_phase2_execution_indexes.sql"))
	if err != nil {
		t.Fatalf("ReadFile execution indexes returned error: %v", err)
	}
	if !strings.Contains(string(firstMigration), "projects") {
		t.Fatalf("expected core migration to define projects table")
	}
	if !strings.Contains(string(secondMigration), "CREATE INDEX") {
		t.Fatalf("expected secondary migration to define indexes")
	}

	scriptPath := filepath.Join(root, "deploy", "scripts", "migrate-postgres.sh")
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("Stat migrate-postgres.sh returned error: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("expected migrate-postgres.sh to be executable")
	}
}
