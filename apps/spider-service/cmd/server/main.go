package main

import (
	"log"

	"crawler-platform/apps/spider-service/internal/api"
	"crawler-platform/apps/spider-service/internal/service"
)

func main() {
	router := api.NewRouter(service.NewSpiderService())
	if err := router.Run(":8083"); err != nil {
		log.Fatal(err)
	}
}
