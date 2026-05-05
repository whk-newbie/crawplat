package api

import (
	"net/http"

	"crawler-platform/apps/monitor-service/internal/service"
	"github.com/gin-gonic/gin"
)

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
