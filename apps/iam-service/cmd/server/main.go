// Package main 是 IAM 认证服务的程序入口。
// 负责加载配置、创建认证服务实例、注册 HTTP 路由并监听 :8081 端口。
// 不处理路由注册细节——该职责属于 internal/api。
package main

import (
	"log"
	"os"
	"strings"

	"crawler-platform/apps/iam-service/internal/api"
	"crawler-platform/apps/iam-service/internal/repo"
	"crawler-platform/apps/iam-service/internal/service"
)

func main() {
	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || strings.TrimSpace(secret) == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	enableSeedAdmin := strings.EqualFold(os.Getenv("IAM_ENABLE_SEED_ADMIN"), "true")
	router := api.NewRouter(service.NewAuthService(secret, repo.NewUserRepo(enableSeedAdmin)))
	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
