// Package main 是 IAM 认证服务的程序入口。
// 负责加载配置、创建认证服务实例、注册 HTTP 路由并监听 :8081 端口。
// 不处理路由注册细节——该职责属于 internal/api。
// 支持两种用户存储后端：设置 DATABASE_DSN 时使用 PostgreSQL，
// 否则回退到内存存储（仅用于开发/测试）。
package main

import (
	"log"
	"os"
	"strings"

	"crawler-platform/apps/iam-service/internal/api"
	"crawler-platform/apps/iam-service/internal/repo"
	"crawler-platform/apps/iam-service/internal/service"
	"crawler-platform/packages/go-common/postgres"
)

func main() {
	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || strings.TrimSpace(secret) == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	enableSeedAdmin := strings.EqualFold(os.Getenv("IAM_ENABLE_SEED_ADMIN"), "true")

	var userRepo service.UserRepository
	if dsn := os.Getenv("DATABASE_DSN"); strings.TrimSpace(dsn) != "" {
		db, err := postgres.Open(strings.TrimSpace(dsn))
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		if err := db.Ping(); err != nil {
			log.Fatalf("ping database: %v", err)
		}
		pgRepo, err := repo.NewPostgresUserRepo(db, enableSeedAdmin)
		if err != nil {
			log.Fatalf("init postgres user repo: %v", err)
		}
		userRepo = pgRepo
		log.Println("using postgres user repository")
	} else {
		userRepo = repo.NewUserRepo(enableSeedAdmin)
		log.Println("using in-memory user repository")
	}

	router := api.NewRouter(service.NewAuthService(secret, userRepo))
	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
