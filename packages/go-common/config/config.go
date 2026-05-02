package config

import "os"

type App struct {
	HTTPAddr            string
	PostgresDSN         string
	RedisAddr           string
	MongoURI            string
	JWTSecret           string
	InternalAPIToken    string
	ExecutionServiceURL string
	SpiderServiceURL    string
}

func Load() App {
	return App{
		HTTPAddr:            envOrDefault("HTTP_ADDR", ":8080"),
		PostgresDSN:         os.Getenv("POSTGRES_DSN"),
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		MongoURI:            os.Getenv("MONGO_URI"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		InternalAPIToken:    envOrDefault("INTERNAL_API_TOKEN", os.Getenv("JWT_SECRET")),
		ExecutionServiceURL: envOrDefault("EXECUTION_SERVICE_URL", "http://execution-service:8085"),
		SpiderServiceURL:    envOrDefault("SPIDER_SERVICE_URL", "http://spider-service:8083"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
