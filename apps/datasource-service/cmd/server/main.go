// 数据源服务（datasource-service）的入口文件。
//
// 负责：
//   1. 加载全局配置（commonconfig）
//   2. 建立 PostgreSQL 连接（用于数据源元数据存储，不是外部数据源）
//   3. 依赖注入并启动 HTTP 服务（端口 8086）
//
// 依赖链：config → PostgresRepository → DatasourceService → Gin Router → HTTP server
// 不负责：HTTP 路由定义（由 api 包处理）、业务逻辑（由 service 包处理）、探针操作（由 Prober 处理）。
package main

import (
	"log"

	"crawler-platform/apps/datasource-service/internal/api"
	datasourcerepo "crawler-platform/apps/datasource-service/internal/repo"
	"crawler-platform/apps/datasource-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonpostgres "crawler-platform/packages/go-common/postgres"
)

// main 是应用入口，完成依赖注入并启动 HTTP 服务。
// 启动流程：
//   1. Load 全局配置（从环境变量/配置文件读取 PostgresDSN 等）
//   2. Open PostgreSQL 连接池（用于存取 datasources 表）
//   3. 构建依赖链：PostgresRepository → DatasourceService → Gin Router
//   4. 在 :8086 端口启动 HTTP 服务
func main() {
	cfg := commonconfig.Load()
	db, err := commonpostgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := api.NewRouter(service.NewDatasourceService(datasourcerepo.NewPostgresRepository(db)))
	if err := router.Run(":8086"); err != nil {
		log.Fatal(err)
	}
}
