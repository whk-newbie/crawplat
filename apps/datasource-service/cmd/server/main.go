package main

import (
	"log"

	"crawler-platform/apps/datasource-service/internal/api"
	"crawler-platform/apps/datasource-service/internal/service"
)

func main() {
	router := api.NewRouter(service.NewDatasourceService())
	if err := router.Run(":8086"); err != nil {
		log.Fatal(err)
	}
}
