package api

import (
	"crawler-platform/apps/gateway/internal/proxy"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.Any("/api/v1/auth", proxy.ProxyTo("iam-service"))
	router.Any("/api/v1/auth/*path", proxy.ProxyTo("iam-service"))

	router.Any("/api/v1/projects", proxy.ProxyTo("project-service"))

	router.Any("/api/v1/projects/:projectId/spiders", proxy.ProxyTo("spider-service"))

	router.Any("/api/v1/executions", proxy.ProxyTo("execution-service"))
	router.Any("/api/v1/executions/*path", proxy.ProxyTo("execution-service"))

	router.Any("/api/v1/nodes", proxy.ProxyTo("node-service"))
	router.Any("/api/v1/nodes/*path", proxy.ProxyTo("node-service"))

	router.Any("/api/v1/datasources", proxy.ProxyTo("datasource-service"))
	router.Any("/api/v1/datasources/*path", proxy.ProxyTo("datasource-service"))

	return router
}
