package main

import (
	"context"
	"log"

	"crawler-platform/apps/execution-service/internal/api"
	"crawler-platform/apps/execution-service/internal/queue"
	execrepo "crawler-platform/apps/execution-service/internal/repo"
	"crawler-platform/apps/execution-service/internal/service"
	commonconfig "crawler-platform/packages/go-common/config"
	commonmongox "crawler-platform/packages/go-common/mongox"
	commonpostgres "crawler-platform/packages/go-common/postgres"
	commonredis "crawler-platform/packages/go-common/redisx"
)

func main() {
	cfg := commonconfig.Load()

	db, err := commonpostgres.Open(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	redisClient, err := commonredis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	mongoClient, err := commonmongox.Connect(context.Background(), cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = mongoClient.Disconnect(context.Background())
	}()

	router := api.NewRouter(service.NewExecutionService(
		execrepo.NewExecutionRepository(db),
		execrepo.NewMongoLogRepository(mongoClient.Database("crawler")),
		queue.NewRedisQueue(redisClient),
	))
	if err := router.Run(":8085"); err != nil {
		log.Fatal(err)
	}
}
