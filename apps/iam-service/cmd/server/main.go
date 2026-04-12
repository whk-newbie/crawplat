package main

import (
	"log"
	"os"

	"crawler-platform/apps/iam-service/internal/api"
	"crawler-platform/apps/iam-service/internal/service"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}

	router := api.NewRouter(service.NewAuthService(secret))
	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
