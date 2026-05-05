// Package main 是调度服务（scheduler-service）的启动入口。
//
// 该文件负责：
//   - 加载配置，初始化 PostgreSQL 连接。
//   - 组装 Repository、ExecutionClient 和 SchedulerService。
//   - 启动后台 materialization 循环（定时扫描调度，生成执行记录）。
//   - 启动 HTTP API 服务（端口 8087），对外暴露调度 CRUD 接口。
//
// 与谁交互：
//   - Repository（PostgreSQL）：持久化 Schedule 记录和 last_materialized_at 游标。
//   - ExecutionClient（HTTP）：调用 execution-service 的 /api/v1/executions 创建执行记录，
//     以及 /internal/v1/executions/retries/materialize 执行重试物化。
//
// 不负责：
//   - 不执行具体的爬虫任务（由 execution-service 负责）。
//   - 不做身份认证和授权（由 gateway 负责）。
package main

import (
	"context"
	"log"
	"time"

	"crawler-platform/apps/scheduler-service/internal/api"
	schedulerrepo "crawler-platform/apps/scheduler-service/internal/repo"
	"crawler-platform/apps/scheduler-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonpostgres "crawler-platform/packages/go-common/postgres"
)

// main 启动调度服务：初始化依赖、启动后台循环、启动 HTTP 服务。
// 后台 goroutine 每 15 秒扫描一次调度表，将到期调度物化为执行记录，
// 同时调用 execution-service 的重试物化接口，确保失败任务被重新调度。
func main() {
	cfg := commonconfig.Load()
	db, err := commonpostgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	schedulerService := service.NewSchedulerService(
		schedulerrepo.NewPostgresRepository(db),
		service.NewHTTPExecutionClient(cfg.ExecutionServiceURL, cfg.JWTSecret),
	)
	go func() {
		if err := schedulerService.Run(context.Background(), 15*time.Second); err != nil {
			log.Printf("scheduler loop stopped: %v", err)
		}
	}()

	router := api.NewRouter(schedulerService)
	if err := router.Run(":8087"); err != nil {
		log.Fatal(err)
	}
}
