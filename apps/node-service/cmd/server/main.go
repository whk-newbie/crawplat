// 文件职责：node-service 的启动入口。
// 负责：加载公共配置，创建 Redis 客户端，构建 Service -> Repository 依赖链，
//       启动 HTTP 服务器（监听 :8084 端口）。
// 与谁交互：依赖 go-common 的 config（配置加载）、redisx（Redis 客户端创建），
//           以及本服务的 api、repo、service 子包。
// 不负责：具体的业务逻辑、路由定义、数据存取 —— 这些分别由 service、api、repo 包负责。
package main

import (
	"log"
	"time"

	"crawler-platform/apps/node-service/internal/api"
	noderepo "crawler-platform/apps/node-service/internal/repo"
	"crawler-platform/apps/node-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonredis "crawler-platform/packages/go-common/redisx"
)

// main 启动 node-service 服务端。
// 执行顺序：加载配置 → 创建 Redis 客户端 → 构建依赖图（RedisRepository → NodeService → Router）→ 启动 HTTP 服务。
// 注意：默认使用 Redis 存储，TTL 为 30 秒。如果 Redis 连接失败，服务直接退出（log.Fatal）。
func main() {
	cfg := commonconfig.Load()
	client, err := commonredis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	router := api.NewRouter(service.NewNodeService(noderepo.NewRedisRepository(client, 30*time.Second)))
	if err := router.Run(":8084"); err != nil {
		log.Fatal(err)
	}
}
