package main

import (
	"log"

	"crawler-platform/apps/project-service/internal/api"
	"crawler-platform/apps/project-service/internal/service"
)

func main() {
	router := api.NewRouter(service.NewProjectService())
	if err := router.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
