package api

import (
	"net/http"

	"crawler-platform/apps/node-service/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(nodeService *service.NodeService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/nodes/:id/heartbeat", func(c *gin.Context) {
		var req struct {
			Capabilities []string `json:"capabilities"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		node := nodeService.Heartbeat(c.Param("id"), req.Capabilities)
		c.JSON(http.StatusOK, node)
	})

	router.GET("/api/v1/nodes", func(c *gin.Context) {
		c.JSON(http.StatusOK, nodeService.List())
	})

	return router
}
