// Package api 是 Project 服务的 HTTP 路由层。
// 注册 POST /api/v1/projects（创建）和 GET /api/v1/projects（列表）。
// 负责 JSON 解析、分页参数提取和错误 → HTTP 状态码映射，不处理业务逻辑。
package api

import (
	"net/http"
	"strconv"

	"crawler-platform/apps/project-service/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 创建 Gin 路由引擎并注册项目相关端点。
// POST 响应 201（成功）、400（参数错误）、409（code 冲突）。
// GET 列表支持 ?limit=<n>&offset=<n> 分页，默认 limit=20。
func NewRouter(projectService *service.ProjectService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/projects", func(c *gin.Context) {
		var req struct {
			Code string `json:"code" binding:"required"`
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		project, err := projectService.Create(req.Code, req.Name)
		if err != nil {
			if err == service.ErrProjectCodeExists {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, project)
	})

	router.GET("/api/v1/projects", func(c *gin.Context) {
		limit := parseQueryInt(c, "limit", 20)
		offset := parseQueryInt(c, "offset", 0)

		projects, err := projectService.List(limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, projects)
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
