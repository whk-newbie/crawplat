package api

import (
	"net/http"
	"os"

	"crawler-platform/apps/gateway/internal/proxy"

	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.Any("/api/v1/auth", proxy.ProxyTo("iam-service"))
	router.Any("/api/v1/auth/*path", proxy.ProxyTo("iam-service"))

	router.Any("/api/v1/projects", proxy.ProxyTo("project-service"))

	router.Any("/api/v1/projects/:projectId/spiders", proxy.ProxyTo("spider-service"))

	router.Any("/api/v1/executions", proxy.ProxyTo("execution-service"))
	router.Any("/api/v1/executions/*path", proxy.ProxyTo("execution-service"))
	internalExecution := router.Group("/internal/v1/executions", requireInternalToken())
	internalExecution.Any("/claim", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/start", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/logs", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/complete", proxy.ProxyTo("execution-service"))
	internalExecution.Any("/:id/fail", proxy.ProxyTo("execution-service"))

	router.Any("/api/v1/nodes", proxy.ProxyTo("node-service"))
	router.Any("/api/v1/nodes/*path", proxy.ProxyTo("node-service"))

	router.Any("/api/v1/datasources", proxy.ProxyTo("datasource-service"))
	router.Any("/api/v1/datasources/*path", proxy.ProxyTo("datasource-service"))

	router.Any("/api/v1/schedules", proxy.ProxyTo("scheduler-service"))

	return router
}

func requireInternalToken() gin.HandlerFunc {
	token := os.Getenv("INTERNAL_API_TOKEN")
	if token == "" {
		token = os.Getenv("JWT_SECRET")
	}

	return func(c *gin.Context) {
		if token == "" || c.GetHeader(internalTokenHeader) != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized internal route"})
			return
		}
		c.Next()
	}
}
