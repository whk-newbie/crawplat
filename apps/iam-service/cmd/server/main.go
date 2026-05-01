package main

import (
	"log"
	"os"
	"strings"

	"crawler-platform/apps/iam-service/internal/api"
	"crawler-platform/apps/iam-service/internal/repo"
	"crawler-platform/apps/iam-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonpostgres "crawler-platform/packages/go-common/postgres"
)

func main() {
	cfg := commonconfig.Load()
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	db, err := commonpostgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := repo.NewPostgresUserRepo(db)
	authSvc := service.NewAuthService(cfg.JWTSecret, userRepo)

	if strings.EqualFold(os.Getenv("IAM_ENABLE_SEED_ADMIN"), "true") {
		if _, err := authSvc.Register("admin", "admin123", "admin@localhost"); err != nil {
			log.Printf("seed admin: %v (may already exist)", err)
		}
	}

	router := api.NewRouter(authSvc)
	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
