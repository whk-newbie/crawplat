// 执行服务入口文件。
// 负责加载配置，初始化 PostgreSQL（执行状态）、Redis（任务队列）、MongoDB（执行日志）三个外部依赖，
// 组装依赖注入链（Repository → Service → Router），最后在 :8085 端口启动 HTTP 服务。
// 不负责业务逻辑、路由定义或数据操作——这些由 internal 包处理。
package main

import (
	"context"
	"log"

	"crawler-platform/apps/execution-service/internal/api"
	"crawler-platform/apps/execution-service/internal/queue"
	execrepo "crawler-platform/apps/execution-service/internal/repo"
	"crawler-platform/apps/execution-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonmongox "crawler-platform/packages/go-common/mongox"
	commonpostgres "crawler-platform/packages/go-common/postgres"
	commonredis "crawler-platform/packages/go-common/redisx"
)

// main 是执行服务的入口函数。
// 按顺序完成：加载配置 → 连接 PostgreSQL → 连接 Redis → 连接 MongoDB → 组装服务 → 启动 HTTP。
// 任一初始化步骤失败都会直接 log.Fatal 退出，因为服务无法在缺少依赖的情况下正常启动。
func main() {
	cfg := commonconfig.Load()

	db, err := commonpostgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	redisClient, err := commonredis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	mongoClient, err := commonmongox.Connect(context.Background(), cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = mongoClient.Disconnect(context.Background())
	}()

	router := api.NewRouter(service.NewExecutionService(
		execrepo.NewExecutionRepository(db),
		execrepo.NewMongoLogRepository(mongoClient.Database("crawler")),
		queue.NewRedisQueue(redisClient),
	))
	if err := router.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
