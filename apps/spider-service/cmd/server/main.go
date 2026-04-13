package main

import (
	"log"

	"crawler-platform/apps/spider-service/internal/api"
	spiderrepo "crawler-platform/apps/spider-service/internal/repo"
	"crawler-platform/apps/spider-service/internal/service"
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

	router := api.NewRouter(service.NewSpiderService(spiderrepo.NewPostgresRepository(db)))
	if err := router.Run(":8083"); err != nil {
		log.Fatal(err)
	}
}
