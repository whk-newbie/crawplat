// Package service 中的数据源探针（prober）实现，负责真实连接测试和数据预览。
//
// 本文件是核心探针层——所有连接测试和数据预览都是真实操作，非 mock：
//   - 连接测试：对目标数据源建立真实 TCP 连接并执行 Ping/认证
//   - 数据预览：对目标数据源执行真实查询，返回元数据（表名、key 列表、集合名等）
//
// 与谁交互：直接与外部数据源建立连接——
//   - PostgreSQL：通过 pgx/v5 驱动
//   - Redis：通过 go-redis/v9 客户端
//   - MongoDB：通过 mongo-driver/v2 客户端
//
// 所有探针操作均设置 5 秒超时（probeTimeout），防止外部数据源不可达时无限阻塞。
// 错误分类：配置校验错误（ErrDatasourceConfigInvalid）vs 连接/查询失败（ErrDatasourceProbeFailed）。
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

// probeTimeout 定义了连接测试和数据预览的统一超时时间。
// 设置为 5 秒以在"快速反馈给前端"和"容忍慢网络"之间取得平衡。
const probeTimeout = 5 * time.Second

// Prober 定义了数据源探针的接口。
// 实现者必须支持连接测试（Test）和数据预览（Preview），两者均为真实操作。
type Prober interface {
	Test(ctx context.Context, datasource model.Datasource) (model.TestResult, error)
	Preview(ctx context.Context, datasource model.Datasource) (model.PreviewResult, error)
}

// newLiveDatasourceProber 创建真实探针实例，所有操作都直接连接外部数据源。
func newLiveDatasourceProber() Prober {
	return &liveDatasourceProber{}
}

// liveDatasourceProber 是 Prober 接口的真实实现，连接外部 PostgreSQL/Redis/MongoDB。
type liveDatasourceProber struct{}

// Test 对指定数据源执行真实连接测试。
// 内部创建带超时的 context（5 秒），根据数据源类型路由到对应的测试函数。
// 返回 TestResult.Status="ok" 表示连接成功；失败时返回包裹了 ErrDatasourceProbeFailed 或 ErrDatasourceConfigInvalid 的错误。
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

// Preview 对指定数据源执行真实数据预览查询。
// 内部创建带超时的 context（5 秒），根据数据源类型路由到对应的预览函数。
// 返回的 Rows 结构取决于数据源类型：PostgreSQL 返回表名，Redis 返回 key 和类型，MongoDB 返回数据库/集合名。
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

// testPostgres 使用 pgx 驱动连接 PostgreSQL 并执行 Ping，验证连接可用性。
// 连接失败时返回 ErrDatasourceProbeFailed，便于前端根据错误类型展示不同的失败原因。
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

// previewPostgres 查询 information_schema.tables 获取 schema 下的表名列表（最多 5 条）。
// schema 默认使用 "public"，可通过 config 中的 "schema" 键覆盖。
// 这样前端即可展示"该 PostgreSQL 数据源中有哪些表"的预览信息。
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

// testRedis 创建 Redis 客户端并执行 PING 命令验证连接。
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

// previewRedis 使用 SCAN 命令扫描前 5 个 key，返回 key 名、类型及字符串类型的值截断（最多 120 字符）。
// 注意：SCAN 在生产环境可能较慢，此处限制 count 为 5 以避免对 Redis 造成压力。
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

// testMongo 创建 MongoDB 客户端并对 Primary 节点执行 Ping，验证连接及副本集状态。
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

// previewMongo 预览 MongoDB 结构信息：
//   - 如果未指定 database，则列出前 5 个数据库名
//   - 如果指定了 database，则列出该数据库下前 5 个集合名
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

// postgresDSN 从 config map 中构建 PostgreSQL 连接字符串（DSN）。
// 支持两种配置方式：
//   1. 直接提供完整 URI/DSN（优先级最高）
//   2. 分别提供 host、port、user、password、database 等字段，由本函数拼接
// 未指定 sslmode 时默认使用 "disable"（开发环境常用）。
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

// newRedisClient 从 config map 创建 Redis 客户端。
// 支持 "addr" 或 "host:port" 两种地址配置方式，默认端口 6379。
// db 参数默认为 0。
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

// newMongoClient 从 config map 中的 "uri" 字段创建 MongoDB 客户端。
// MongoDB 必须提供完整 URI 连接字符串，不支持分别指定 host/port。
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

// firstNonEmpty 从 config map 中按优先级返回第一个非空的配置值。
// keys 按优先级降序排列——用于支持同一配置项的多种键名（如 "user" vs "username"）。
func firstNonEmpty(cfg map[string]string, keys ...string) string {
	for _, key := range keys {
		if strings.TrimSpace(cfg[key]) != "" {
			return strings.TrimSpace(cfg[key])
		}
	}
	return ""
}

// truncateValue 截断字符串到指定最大长度，超出部分用 "..." 表示。
// 用于 Redis 字符串类型的值预览，避免超长 value 撑爆前端展示。
// 当 maxLen <= 0 或字符串未超过限制时返回原值，不做修改。
func truncateValue(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	if maxLen <= 3 {
		return value[:maxLen]
	}
	return value[:maxLen-3] + "..."
}
