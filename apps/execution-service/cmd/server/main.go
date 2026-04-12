package main

import (
	"log"

	"crawler-platform/apps/execution-service/internal/api"
	"crawler-platform/apps/execution-service/internal/service"
)

func main() {
	router := api.NewRouter(service.NewExecutionService())
	if err := router.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
