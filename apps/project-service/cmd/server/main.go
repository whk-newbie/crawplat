// Package main 是 Project 服务的程序入口。
// 负责加载数据库配置、创建 PostgreSQL 仓储并注入 service → api 依赖链，最后监听 :8082。
package main

import (
	"log"

	"crawler-platform/apps/project-service/internal/api"
	projectrepo "crawler-platform/apps/project-service/internal/repo"
	"crawler-platform/apps/project-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonpostgres "crawler-platform/packages/go-common/postgres"
)

// main 组装依赖（config → db → repo → service → router）并启动 HTTP。
func main() {
	cfg := commonconfig.Load()
	db, err := commonpostgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := api.NewRouter(service.NewProjectService(projectrepo.NewPostgresRepository(db)))
	if err := router.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
