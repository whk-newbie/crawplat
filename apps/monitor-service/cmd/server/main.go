package main

import (
	"context"
	"log"
	"os"
	"time"

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

	monitorService := service.NewMonitorService(monitorrepo.NewOverviewRepository(db, redisClient))
	pollInterval := 15 * time.Second
	if raw := os.Getenv("MONITOR_ALERT_POLL_INTERVAL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			log.Printf("invalid MONITOR_ALERT_POLL_INTERVAL=%q, fallback to %s", raw, pollInterval)
		} else {
			pollInterval = parsed
		}
	}
	nodeOfflinePollInterval := 5 * time.Second
	if raw := os.Getenv("MONITOR_NODE_OFFLINE_ALERT_POLL_INTERVAL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			log.Printf("invalid MONITOR_NODE_OFFLINE_ALERT_POLL_INTERVAL=%q, fallback to %s", raw, nodeOfflinePollInterval)
		} else {
			nodeOfflinePollInterval = parsed
		}
	}
	monitorService.StartAlertLoops(context.Background(), pollInterval, nodeOfflinePollInterval)

	router := api.NewRouter(monitorService)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
