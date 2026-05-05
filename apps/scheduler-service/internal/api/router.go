// Package api 定义调度服务（scheduler-service）的 HTTP 路由层。
//
// 该文件负责：
//   - 注册 POST /api/v1/schedules（创建调度）和 GET /api/v1/schedules（列出调度）两个端点。
//   - 请求参数绑定与校验（使用 gin.ShouldBindJSON）。
//   - 错误码映射：参数错误返回 400，内部错误返回 500。
//   - 创建成功返回 201，列表成功返回 200。
//
// 与谁交互：
//   - service.SchedulerService：所有请求均委托给 Service 层处理。
//
// 不负责：
//   - 不做身份认证（由 gateway 负责）。
//   - 不执行具体业务逻辑（由 Service 层负责）。
//   - 不处理分页参数（当前 List 接口未实现分页，返回所有调度）。
package api

import (
	"net/http"

	"crawler-platform/apps/scheduler-service/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 创建并配置 Gin 路由引擎，注册所有调度相关 API 端点。
// 参数 schedulerService 为上层业务服务，路由层不直接操作 Repository 或 ExecutionClient。
func NewRouter(schedulerService *service.SchedulerService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/schedules", func(c *gin.Context) {
		var req struct {
			ProjectID         string   `json:"projectId" binding:"required"`
			SpiderID          string   `json:"spiderId" binding:"required"`
			Name              string   `json:"name" binding:"required"`
			CronExpr          string   `json:"cronExpr" binding:"required"`
			Enabled           bool     `json:"enabled"`
			Image             string   `json:"image" binding:"required"`
			Command           []string `json:"command"`
			RetryLimit        int      `json:"retryLimit"`
			RetryDelaySeconds int      `json:"retryDelaySeconds"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		schedule, err := schedulerService.Create(req.ProjectID, req.SpiderID, req.Name, req.CronExpr, req.Image, req.Command, req.Enabled, req.RetryLimit, req.RetryDelaySeconds)
		if err != nil {
			if err == service.ErrInvalidSchedule {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, schedule)
	})

	router.GET("/api/v1/schedules", func(c *gin.Context) {
		schedules, err := schedulerService.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, schedules)
	})

	return router
}
