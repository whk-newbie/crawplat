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
