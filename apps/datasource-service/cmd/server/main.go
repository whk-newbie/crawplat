package main

import (
	"log"

	"crawler-platform/apps/datasource-service/internal/api"
	datasourcerepo "crawler-platform/apps/datasource-service/internal/repo"
	"crawler-platform/apps/datasource-service/internal/service"
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

	router := api.NewRouter(service.NewDatasourceService(datasourcerepo.NewPostgresRepository(db)))
	if err := router.Run(":8086"); err != nil {
		log.Fatal(err)
	}
}
