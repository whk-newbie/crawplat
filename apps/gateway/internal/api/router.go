package api

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"crawler-platform/apps/gateway/internal/proxy"
	commonauth "crawler-platform/packages/go-common/auth"

	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

type authConfig struct {
	enforceJWT    bool
	jwtSecret     string
	internalToken string
}

type rateLimitConfig struct {
	enabled       bool
	windowSeconds int
	maxRequests   int
}

type observabilityConfig struct {
	requestLogEnabled bool
	requestIDHeader   string
	trustRequestID    bool
}

func NewRouter() *gin.Engine {
	return newRouter(loadAuthConfig(), loadRateLimitConfig(), loadObservabilityConfig())
}

func newRouter(cfg authConfig, rlCfg rateLimitConfig, obsCfg observabilityConfig) *gin.Engine {
	router := gin.Default()
	router.Use(attachRequestID(obsCfg))
	if obsCfg.requestLogEnabled {
		router.Use(logRequest(obsCfg))
	}

	router.Any("/api/v1/auth", proxy.ProxyTo("iam-service"))
	router.Any("/api/v1/auth/*path", proxy.ProxyTo("iam-service"))

	api := router.Group("/api/v1")
	if cfg.enforceJWT {
		api.Use(requireJWT(cfg.jwtSecret))
	}
	if rlCfg.enabled {
		api.Use(requireRateLimit(rlCfg))
	}

	api.Any("/projects", proxy.ProxyTo("project-service"))

	api.Any("/projects/:projectId/spiders", proxy.ProxyTo("spider-service"))

	api.Any("/executions", proxy.ProxyTo("execution-service"))
	api.Any("/executions/*path", proxy.ProxyTo("execution-service"))
	internalExecution := router.Group("/internal/v1/executions", requireInternalToken(cfg.internalToken))
	internalExecution.Any("/claim", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/start", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/logs", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/complete", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/fail", proxy.ProxyTo("execution-service"))

	api.Any("/nodes", proxy.ProxyTo("node-service"))
	api.Any("/nodes/*path", proxy.ProxyTo("node-service"))

	api.Any("/datasources", proxy.ProxyTo("datasource-service"))
	api.Any("/datasources/*path", proxy.ProxyTo("datasource-service"))

	api.Any("/schedules", proxy.ProxyTo("scheduler-service"))
	api.Any("/monitor/*path", proxy.ProxyTo("monitor-service"))

	return router
}

func loadAuthConfig() authConfig {
	internalToken := os.Getenv("INTERNAL_API_TOKEN")
	if internalToken == "" {
		internalToken = os.Getenv("JWT_SECRET")
	}
	return authConfig{
		enforceJWT:    envBool("GATEWAY_ENFORCE_JWT", false),
		jwtSecret:     os.Getenv("JWT_SECRET"),
		internalToken: internalToken,
	}
}

func loadRateLimitConfig() rateLimitConfig {
	return rateLimitConfig{
		enabled:       envBool("GATEWAY_RATE_LIMIT_ENABLED", false),
		windowSeconds: envInt("GATEWAY_RATE_LIMIT_WINDOW_SECONDS", 60),
		maxRequests:   envInt("GATEWAY_RATE_LIMIT_MAX_REQUESTS", 120),
	}
}

func loadObservabilityConfig() observabilityConfig {
	header := strings.TrimSpace(os.Getenv("GATEWAY_REQUEST_ID_HEADER"))
	if header == "" {
		header = "X-Request-Id"
	}
	return observabilityConfig{
		requestLogEnabled: envBool("GATEWAY_REQUEST_LOG_ENABLED", true),
		requestIDHeader:   header,
		trustRequestID:    envBool("GATEWAY_TRUST_REQUEST_ID", true),
	}
}

func envBool(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func requireInternalToken(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" || c.GetHeader(internalTokenHeader) != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized internal route"})
			return
		}
		c.Next()
	}
}

func requireJWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(authz, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "jwt auth is misconfigured"})
			return
		}
		if _, err := commonauth.ParseToken(secret, token); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}
		c.Next()
	}
}
