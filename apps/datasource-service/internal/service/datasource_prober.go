package service

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"crawler-platform/apps/datasource-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const probeTimeout = 5 * time.Second

type Prober interface {
	Test(ctx context.Context, datasource model.Datasource) (model.TestResult, error)
	Preview(ctx context.Context, datasource model.Datasource) (model.PreviewResult, error)
}

func newLiveDatasourceProber() Prober {
	return &liveDatasourceProber{}
}

type liveDatasourceProber struct{}

func (p *liveDatasourceProber) Test(ctx context.Context, datasource model.Datasource) (model.TestResult, error) {
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	switch datasource.Type {
	case "postgresql":
		if err := testPostgres(probeCtx, datasource.Config); err != nil {
			return model.TestResult{}, err
		}
	case "redis":
		if err := testRedis(probeCtx, datasource.Config); err != nil {
			return model.TestResult{}, err
		}
	case "mongodb":
		if err := testMongo(probeCtx, datasource.Config); err != nil {
			return model.TestResult{}, err
		}
	default:
		return model.TestResult{}, ErrInvalidDatasourceType
	}

	return model.TestResult{
		DatasourceID: datasource.ID,
		Status:       "ok",
		Message:      "connection test passed",
	}, nil
}

func (p *liveDatasourceProber) Preview(ctx context.Context, datasource model.Datasource) (model.PreviewResult, error) {
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	var (
		rows []map[string]string
		err  error
	)

	switch datasource.Type {
	case "postgresql":
		rows, err = previewPostgres(probeCtx, datasource.Config)
	case "redis":
		rows, err = previewRedis(probeCtx, datasource.Config)
	case "mongodb":
		rows, err = previewMongo(probeCtx, datasource.Config)
	default:
		return model.PreviewResult{}, ErrInvalidDatasourceType
	}
	if err != nil {
		return model.PreviewResult{}, err
	}

	return model.PreviewResult{
		DatasourceID:   datasource.ID,
		DatasourceType: datasource.Type,
		Rows:           rows,
	}, nil
}

func testPostgres(ctx context.Context, cfg map[string]string) error {
	dsn, err := postgresDSN(cfg)
	if err != nil {
		return err
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	defer conn.Close(context.Background())

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	return nil
}

func previewPostgres(ctx context.Context, cfg map[string]string) ([]map[string]string, error) {
	dsn, err := postgresDSN(cfg)
	if err != nil {
		return nil, err
	}
	schema := firstNonEmpty(cfg, "schema")
	if schema == "" {
		schema = "public"
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		ORDER BY table_name
		LIMIT 5
	`, schema)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	defer rows.Close()

	previewRows := make([]map[string]string, 0, 5)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
		}
		previewRows = append(previewRows, map[string]string{
			"schema": schema,
			"table":  tableName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	return previewRows, nil
}

func testRedis(ctx context.Context, cfg map[string]string) error {
	client, err := newRedisClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	return nil
}

func previewRedis(ctx context.Context, cfg map[string]string) ([]map[string]string, error) {
	client, err := newRedisClient(cfg)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	keys, _, err := client.Scan(ctx, 0, "*", 5).Result()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}

	previewRows := make([]map[string]string, 0, len(keys))
	for _, key := range keys {
		typ, err := client.Type(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
		}
		row := map[string]string{
			"key":  key,
			"type": typ,
		}
		if typ == "string" {
			if value, err := client.Get(ctx, key).Result(); err == nil {
				row["value"] = truncateValue(value, 120)
			}
		}
		previewRows = append(previewRows, row)
	}
	return previewRows, nil
}

func testMongo(ctx context.Context, cfg map[string]string) error {
	client, err := newMongoClient(cfg)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	return nil
}

func previewMongo(ctx context.Context, cfg map[string]string) ([]map[string]string, error) {
	client, err := newMongoClient(cfg)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	database := firstNonEmpty(cfg, "database", "db")
	if database == "" {
		names, err := client.ListDatabaseNames(ctx, bson.D{})
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
		}
		if len(names) > 5 {
			names = names[:5]
		}
		rows := make([]map[string]string, 0, len(names))
		for _, name := range names {
			rows = append(rows, map[string]string{"database": name})
		}
		return rows, nil
	}

	collections, err := client.Database(database).ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	if len(collections) > 5 {
		collections = collections[:5]
	}
	rows := make([]map[string]string, 0, len(collections))
	for _, name := range collections {
		rows = append(rows, map[string]string{
			"database":   database,
			"collection": name,
		})
	}
	return rows, nil
}

func postgresDSN(cfg map[string]string) (string, error) {
	if uri := firstNonEmpty(cfg, "uri", "dsn"); uri != "" {
		return uri, nil
	}

	host := strings.TrimSpace(cfg["host"])
	port := strings.TrimSpace(cfg["port"])
	if host == "" {
		addr := strings.TrimSpace(firstNonEmpty(cfg, "addr", "address"))
		if addr != "" {
			parsedHost, parsedPort, err := net.SplitHostPort(addr)
			if err == nil {
				host = parsedHost
				port = parsedPort
			} else if strings.Count(addr, ":") == 0 {
				host = addr
			}
		}
	}
	if host == "" {
		return "", fmt.Errorf("%w: postgresql host/uri is required", ErrDatasourceConfigInvalid)
	}
	if port == "" {
		port = "5432"
	}

	user := strings.TrimSpace(firstNonEmpty(cfg, "user", "username"))
	database := strings.TrimSpace(firstNonEmpty(cfg, "database", "dbname"))
	if user == "" || database == "" {
		return "", fmt.Errorf("%w: postgresql user/database is required", ErrDatasourceConfigInvalid)
	}
	password := cfg["password"]
	sslmode := strings.TrimSpace(firstNonEmpty(cfg, "sslmode"))
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		url.QueryEscape(database),
		url.QueryEscape(sslmode),
	), nil
}

func newRedisClient(cfg map[string]string) (*redis.Client, error) {
	addr := strings.TrimSpace(firstNonEmpty(cfg, "addr", "address"))
	if addr == "" {
		host := strings.TrimSpace(cfg["host"])
		if host == "" {
			return nil, fmt.Errorf("%w: redis addr/host is required", ErrDatasourceConfigInvalid)
		}
		port := strings.TrimSpace(cfg["port"])
		if port == "" {
			port = "6379"
		}
		addr = net.JoinHostPort(host, port)
	}

	db := 0
	if rawDB := strings.TrimSpace(cfg["db"]); rawDB != "" {
		parsedDB, err := strconv.Atoi(rawDB)
		if err != nil {
			return nil, fmt.Errorf("%w: redis db must be integer", ErrDatasourceConfigInvalid)
		}
		db = parsedDB
	}

	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg["password"],
		DB:       db,
	}), nil
}

func newMongoClient(cfg map[string]string) (*mongo.Client, error) {
	uri := strings.TrimSpace(firstNonEmpty(cfg, "uri"))
	if uri == "" {
		return nil, fmt.Errorf("%w: mongodb uri is required", ErrDatasourceConfigInvalid)
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatasourceProbeFailed, err)
	}
	return client, nil
}

func firstNonEmpty(cfg map[string]string, keys ...string) string {
	for _, key := range keys {
		if strings.TrimSpace(cfg[key]) != "" {
			return strings.TrimSpace(cfg[key])
		}
	}
	return ""
}

func truncateValue(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	if maxLen <= 3 {
		return value[:maxLen]
	}
	return value[:maxLen-3] + "..."
}
