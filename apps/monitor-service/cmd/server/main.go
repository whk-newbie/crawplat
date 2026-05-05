// Package main 是 Monitor 服务的程序入口。
// 负责创建 MonitorService 实例、注册 HTTP 路由并监听 :8088 端口。
package main

import (
	"log"

	"crawler-platform/apps/monitor-service/internal/api"
	"crawler-platform/apps/monitor-service/internal/service"
)

func main() {
	router := api.NewRouter(service.NewMonitorService())
	if err := router.Run(":8088"); err != nil {
		log.Fatal(err)
	}
}
