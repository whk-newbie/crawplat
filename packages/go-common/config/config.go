package config

type App struct {
	HTTPAddr    string
	PostgresDSN string
	RedisAddr   string
	MongoURI    string
	JWTSecret   string
}
