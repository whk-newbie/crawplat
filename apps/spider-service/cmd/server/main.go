// 该文件为 Spider 服务的启动入口，负责加载配置、初始化 PostgreSQL 数据库连接、
// 组装依赖链（Repository → Service → Router）并启动 Gin HTTP 服务器（默认监听 :8083）。
// 不负责路由注册、业务校验或持久化逻辑——这些分别由 api、service、repo 层处理。
package main

import (
	"log"

	"crawler-platform/apps/spider-service/internal/api"
	spiderrepo "crawler-platform/apps/spider-service/internal/repo"
	"crawler-platform/apps/spider-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonpostgres "crawler-platform/packages/go-common/postgres"
)

// main 启动 Spider 服务：加载公共配置，打开数据库连接，组装依赖并启动 HTTP 服务器。
// 如果数据库连接失败或 HTTP 服务退出，会触发 log.Fatal 终止进程。
func main() {
	cfg := commonconfig.Load()
	db, err := commonpostgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := api.NewRouter(service.NewSpiderService(spiderrepo.NewPostgresRepository(db)))
	if err := router.Run(":8083"); err != nil {
		log.Fatal(err)
	}
}
