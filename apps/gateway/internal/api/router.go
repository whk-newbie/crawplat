package api

import (
	"os"
	"strconv"
	"strings"

	"crawler-platform/apps/gateway/internal/proxy"

	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

// NewRouter 构建 gateway 的统一入口路由，集中挂载观测、限流、鉴权与版本路由能力。
func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	obsCfg := observabilityConfig{
		requestIDHeader: getenvDefault("GATEWAY_REQUEST_ID_HEADER", "X-Request-ID"),
		trustRequestID:  getenvBoolDefault("GATEWAY_TRUST_REQUEST_ID", true),
	}
	router.Use(attachRequestID(obsCfg), logRequest(obsCfg))

	rateLimitHandler := requireRateLimit(rateLimitConfig{
		windowSeconds: getenvIntDefault("GATEWAY_RATE_LIMIT_WINDOW_SECONDS", 60),
		maxRequests:   getenvIntDefault("GATEWAY_RATE_LIMIT_MAX_REQUESTS", 120),
	})
	authCfg := authConfig{
		enforceJWT: getenvBoolDefault("GATEWAY_ENFORCE_JWT", false),
		jwtSecret:  os.Getenv("JWT_SECRET"),
	}

	for _, version := range supportedAPIVersions() {
		registerAPIVersion(router, version, authCfg, rateLimitHandler)
	}
	registerUnsupportedAPIVersion(router)

	internalExecution := router.Group("/internal/v1/executions", requireInternalToken())
	internalExecution.Any("/claim", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/start", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/logs", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/complete", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/fail", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/retries/materialize", proxy.ProxyTo("execution-service"))

	return router
}

// requireInternalToken 保护 agent / scheduler 调用的内部执行接口，避免公共客户端绕过 gateway 鉴权。
func requireInternalToken() gin.HandlerFunc {
	token := strings.TrimSpace(os.Getenv("INTERNAL_API_TOKEN"))
	if token == "" {
		token = strings.TrimSpace(os.Getenv("JWT_SECRET"))
	}

	return func(c *gin.Context) {
		if token == "" {
			respondError(c, 401, "internal token is not configured")
			return
		}
		if strings.TrimSpace(c.GetHeader(internalTokenHeader)) != token {
			respondError(c, 401, "unauthorized internal route")
			return
		}
		c.Next()
	}
}

func supportedAPIVersions() []string {
	raw := strings.TrimSpace(os.Getenv("GATEWAY_API_SUPPORTED_VERSIONS"))
	if raw == "" {
		return []string{stableAPIVersion}
	}

	seen := map[string]bool{}
	versions := make([]string, 0)
	for _, part := range strings.Split(raw, ",") {
		version := normalizeAPIVersion(part)
		if version == "" || seen[version] {
			continue
		}
		seen[version] = true
		versions = append(versions, version)
	}
	if len(versions) == 0 {
		return []string{stableAPIVersion}
	}
	return versions
}

func getenvDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getenvBoolDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvIntDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
