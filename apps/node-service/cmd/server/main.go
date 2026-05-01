package main

import (
	"log"
	"time"

	"crawler-platform/apps/node-service/internal/api"
	noderepo "crawler-platform/apps/node-service/internal/repo"
	"crawler-platform/apps/node-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonpostgres "crawler-platform/packages/go-common/postgres"
	commonredis "crawler-platform/packages/go-common/redisx"
)

func main() {
	cfg := commonconfig.Load()

	pgDB, err := commonpostgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer pgDB.Close()

	client, err := commonredis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	liveRepo := noderepo.NewRedisRepository(client, 30*time.Second)
	catalogRepo := noderepo.NewPostgresNodeRepository(pgDB)
	router := api.NewRouter(service.NewNodeServiceWithCatalog(liveRepo, catalogRepo))
	if err := router.Run(":8084"); err != nil {
		log.Fatal(err)
	}
}
