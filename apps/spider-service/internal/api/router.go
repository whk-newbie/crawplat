// 该文件为 Spider 服务的 HTTP 路由层，负责注册 Gin 路由和请求处理函数。
//
// 路由端点（API contract）：
//   - POST /api/v1/projects/:projectId/spiders —— 创建爬虫
//   - GET  /api/v1/projects/:projectId/spiders —— 列表查询爬虫（分页）
//   - POST /api/v1/spiders/:spiderId/versions —— 创建爬虫版本
//   - GET  /api/v1/spiders/:spiderId/versions —— 查询爬虫版本列表
//   - GET  /api/v1/projects/:projectId/registry-auth-refs —— 查询镜像仓库认证引用
//
// Handler 负责解析请求参数、调用 Service 层处理业务、根据 Service 返回的哨兵错误
// 映射 HTTP 状态码（BadRequest / NotFound / InternalServerError）。
//
// 不包含任何业务逻辑——所有校验和持久化委托给 service.SpiderService。
package api

import (
	"net/http"
	"strconv"

	"crawler-platform/apps/spider-service/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 创建并配置 Gin 引擎，注册 Spider 相关路由。
// 接收 *service.SpiderService 作为依赖，路由处理函数通过闭包捕获该依赖。
func NewRouter(spiderService *service.SpiderService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/projects/:projectId/spiders", func(c *gin.Context) {
		var req struct {
			Name     string   `json:"name" binding:"required"`
			Language string   `json:"language" binding:"required"`
			Runtime  string   `json:"runtime" binding:"required"`
			Image    string   `json:"image"`
			Command  []string `json:"command"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		spider, err := spiderService.Create(c.Param("projectId"), req.Name, req.Language, req.Runtime, req.Image, req.Command)
		if err != nil {
			switch err {
			case service.ErrInvalidLanguage, service.ErrInvalidRuntime, service.ErrImageRequired:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, spider)
	})

	router.GET("/api/v1/projects/:projectId/spiders", func(c *gin.Context) {
		limit := parseQueryInt(c, "limit", 20)
		offset := parseQueryInt(c, "offset", 0)

		spiders, err := spiderService.List(c.Param("projectId"), limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, spiders)
	})

	router.POST("/api/v1/spiders/:spiderId/versions", func(c *gin.Context) {
		var req struct {
			Version         string   `json:"version" binding:"required"`
			Image           string   `json:"image"`
			RegistryAuthRef string   `json:"registryAuthRef"`
			Command         []string `json:"command"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		v, err := spiderService.CreateVersion(c.Param("spiderId"), req.Version, req.Image, req.RegistryAuthRef, req.Command)
		if err != nil {
			if err == service.ErrSpiderNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, v)
	})

	router.GET("/api/v1/spiders/:spiderId/versions", func(c *gin.Context) {
		versions, err := spiderService.ListVersions(c.Param("spiderId"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, versions)
	})

	router.GET("/api/v1/projects/:projectId/registry-auth-refs", func(c *gin.Context) {
		refs, err := spiderService.ListRegistryAuthRefs(c.Param("projectId"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, refs)
	})

	return router
}

// parseQueryInt 从查询参数中解析整数，解析失败或未提供时返回默认值。
func parseQueryInt(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}
