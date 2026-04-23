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
