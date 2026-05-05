// Package api 是 Monitor 服务的 HTTP 路由层。
// 当前注册 GET /api/v1/monitor/overview 端点，后续扩展告警规则/事件接口。
// 负责 JSON 解析和错误 → HTTP 状态码映射，不处理聚合逻辑。
package api

import (
	"net/http"

	"crawler-platform/apps/monitor-service/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 创建 Gin 引擎并注册监控端点。
func NewRouter(monitorService *service.MonitorService) *gin.Engine {
	router := gin.Default()

	router.GET("/api/v1/monitor/overview", func(c *gin.Context) {
		overview, err := monitorService.Overview()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, overview)
	})

	return router
}
