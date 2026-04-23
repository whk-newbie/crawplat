package main

import (
	"log"

	"crawler-platform/apps/monitor-service/internal/api"
	"crawler-platform/apps/monitor-service/internal/service"
)

func main() {
	router := api.NewRouter(service.NewMonitorService())
	if err := router.Run(":8088"); err != nil {
		log.Fatal(err)
	}
}
