package main

import (
	"log"

	"crawler-platform/apps/monitor-service/internal/api"
	monitorrepo "crawler-platform/apps/monitor-service/internal/repo"
	"crawler-platform/apps/monitor-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonpostgres "crawler-platform/packages/go-common/postgres"
	commonredis "crawler-platform/packages/go-common/redisx"
)

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

	router := api.NewRouter(service.NewMonitorService(monitorrepo.NewOverviewRepository(db, redisClient)))
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
