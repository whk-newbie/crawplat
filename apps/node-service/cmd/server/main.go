package main

import (
	"log"

	"crawler-platform/apps/node-service/internal/api"
	"crawler-platform/apps/node-service/internal/service"
)

func main() {
	router := api.NewRouter(service.NewNodeService())
	if err := router.Run(":8084"); err != nil {
		log.Fatal(err)
	}
}
