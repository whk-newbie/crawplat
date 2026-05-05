package api

import (
	"net/http"
	"strings"

	"crawler-platform/apps/gateway/internal/proxy"

	"github.com/gin-gonic/gin"
)

const apiPrefix = "/api/"
const stableAPIVersion = "v1"

type apiVersionConfig struct {
	supported []string
}

type authConfig struct {
	enforceJWT bool
	jwtSecret  string
}

// registerAPIVersion 注册一个可对外暴露的 API 版本，并把非稳定版本重写到当前稳定上游路径。
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

	protectedGroup.Any("/spiders/:spiderId/versions", proxy.ProxyTo("spider-service"))
	protectedGroup.Any("/spiders/:spiderId/versions/*path", proxy.ProxyTo("spider-service"))

	protectedGroup.Any("/executions", proxy.ProxyTo("execution-service"))
	protectedGroup.Any("/executions/*path", proxy.ProxyTo("execution-service"))

	protectedGroup.Any("/nodes", proxy.ProxyTo("node-service"))
	protectedGroup.Any("/nodes/*path", proxy.ProxyTo("node-service"))

	protectedGroup.Any("/datasources", proxy.ProxyTo("datasource-service"))
	protectedGroup.Any("/datasources/*path", proxy.ProxyTo("datasource-service"))

	protectedGroup.Any("/schedules", proxy.ProxyTo("scheduler-service"))
	protectedGroup.Any("/schedules/*path", proxy.ProxyTo("scheduler-service"))

	protectedGroup.Any("/monitor/*path", proxy.ProxyTo("monitor-service"))
}

// registerUnsupportedAPIVersion 为形如 /api/vN 的未知版本提供稳定的 404 错误边界。
func registerUnsupportedAPIVersion(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, apiPrefix) {
			respondError(c, http.StatusNotFound, "unsupported api version or route")
			return
		}
		respondError(c, http.StatusNotFound, "route not found")
	})
}

// requireAPIVersion 标记当前请求命中的 API 版本，并在需要时改写到稳定上游版本。
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

// requireJWT 统一保护公共业务 API，显式区分缺少凭证和凭证无效两类失败。
func requireJWT(secret string) gin.HandlerFunc {
	secret = strings.TrimSpace(secret)
	return func(c *gin.Context) {
		if secret == "" {
			respondError(c, http.StatusUnauthorized, "jwt secret is not configured")
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			respondError(c, http.StatusUnauthorized, "missing bearer token")
			return
		}
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			respondError(c, http.StatusUnauthorized, "invalid authorization scheme")
			return
		}

		token := strings.TrimSpace(authHeader[7:])
		if token == "" {
			respondError(c, http.StatusUnauthorized, "missing bearer token")
			return
		}
		if token != secret {
			respondError(c, http.StatusUnauthorized, "invalid bearer token")
			return
		}
		c.Next()
	}
}
