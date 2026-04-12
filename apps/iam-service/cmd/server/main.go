package main

import (
	"log"
	"os"
	"strings"

	"crawler-platform/apps/iam-service/internal/api"
	"crawler-platform/apps/iam-service/internal/service"
)

func main() {
	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || strings.TrimSpace(secret) == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	enableSeedAdmin := strings.EqualFold(os.Getenv("IAM_ENABLE_SEED_ADMIN"), "true")
	router := api.NewRouter(service.NewAuthService(secret, enableSeedAdmin))
	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
