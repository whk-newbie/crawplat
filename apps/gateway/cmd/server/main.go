package main

import (
	"log"

	"crawler-platform/apps/gateway/internal/api"
)

func main() {
	router := api.NewRouter()
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
