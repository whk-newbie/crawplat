package api

import (
	"strings"

	"crawler-platform/apps/gateway/internal/proxy"

	"github.com/gin-gonic/gin"
)

const apiPrefix = "/api/"
const stableAPIVersion = "v1"

type apiVersionConfig struct {
	supported []string
}

func registerAPIVersion(router *gin.Engine, version string, cfg authConfig, rateLimitHandler gin.HandlerFunc) {
	versionGroup := router.Group(apiPrefix + version)
	versionGroup.Use(requireAPIVersion(version))

	authGroup := versionGroup.Group("/auth")
	if rateLimitHandler != nil {
		authGroup.Use(rateLimitHandler)
	}
	authGroup.Any("", proxy.ProxyTo("iam-service"))
	authGroup.Any("/*path", proxy.ProxyTo("iam-service"))

	protectedGroup := versionGroup.Group("")
	if cfg.enforceJWT {
		protectedGroup.Use(requireJWT(cfg.jwtSecret))
	}
	if rateLimitHandler != nil {
		protectedGroup.Use(rateLimitHandler)
	}

	protectedGroup.Any("/projects", proxy.ProxyTo("project-service"))
	protectedGroup.Any("/projects/:projectId/spiders", proxy.ProxyTo("spider-service"))

	protectedGroup.Any("/executions", proxy.ProxyTo("execution-service"))
	protectedGroup.Any("/executions/*path", proxy.ProxyTo("execution-service"))

	protectedGroup.Any("/nodes", proxy.ProxyTo("node-service"))
	protectedGroup.Any("/nodes/*path", proxy.ProxyTo("node-service"))

	protectedGroup.Any("/datasources", proxy.ProxyTo("datasource-service"))
	protectedGroup.Any("/datasources/*path", proxy.ProxyTo("datasource-service"))

	protectedGroup.Any("/schedules", proxy.ProxyTo("scheduler-service"))
	protectedGroup.Any("/monitor/*path", proxy.ProxyTo("monitor-service"))
}

func requireAPIVersion(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Version", version)
		if version != stableAPIVersion {
			rewriteAPIVersionPrefix(c, version, stableAPIVersion)
		}
		c.Next()
	}
}

func rewriteAPIVersionPrefix(c *gin.Context, fromVersion string, toVersion string) {
	fromPrefix := apiPrefix + fromVersion
	toPrefix := apiPrefix + toVersion

	if strings.HasPrefix(c.Request.URL.Path, fromPrefix) {
		c.Request.URL.Path = toPrefix + strings.TrimPrefix(c.Request.URL.Path, fromPrefix)
	}

	if c.Request.URL.RawPath != "" && strings.HasPrefix(c.Request.URL.RawPath, fromPrefix) {
		c.Request.URL.RawPath = toPrefix + strings.TrimPrefix(c.Request.URL.RawPath, fromPrefix)
	}
}

func normalizeAPIVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return ""
	}
	if len(version) < 2 || version[0] != 'v' {
		return ""
	}
	for _, ch := range version[1:] {
		if ch < '0' || ch > '9' {
			return ""
		}
	}
	return version
}
