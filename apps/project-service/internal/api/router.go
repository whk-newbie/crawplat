// Package api 是 Project 服务的 HTTP 路由层。
// 注册 POST /api/v1/projects（创建）和 GET /api/v1/projects（列表）。
// 负责 JSON 解析和错误 → HTTP 状态码映射，不处理业务逻辑。
package api

import (
	"net/http"

	"crawler-platform/apps/project-service/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 创建 Gin 路由引擎并注册项目相关端点。
// POST 响应 201（成功）或 400（参数错误）。
// GET 列表返回 200，空列表返回 []（非 null）。
func NewRouter(projectService *service.ProjectService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/projects", func(c *gin.Context) {
		var req struct {
			Code string `json:"code" binding:"required"`
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		project, err := projectService.Create(req.Code, req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, project)
	})

	router.GET("/api/v1/projects", func(c *gin.Context) {
		projects, err := projectService.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, projects)
	})

	return router
}
