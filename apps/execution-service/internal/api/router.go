package api

import (
	"net/http"

	"crawler-platform/apps/execution-service/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(executionService *service.ExecutionService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/executions", func(c *gin.Context) {
		var req struct {
			TaskID          string `json:"taskId" binding:"required"`
			SpiderVersionID string `json:"spiderVersionId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec := executionService.CreateManual(req.TaskID, req.SpiderVersionID)
		c.JSON(http.StatusCreated, exec)
	})

	router.POST("/api/v1/executions/:id/logs", func(c *gin.Context) {
		var req struct {
			Message string `json:"message" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		entry, err := executionService.AppendLog(c.Param("id"), req.Message)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, entry)
	})

	router.GET("/api/v1/executions/:id", func(c *gin.Context) {
		exec, ok := executionService.Get(c.Param("id"))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	router.GET("/api/v1/executions/:id/logs", func(c *gin.Context) {
		logs, ok := executionService.GetLogs(c.Param("id"))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			return
		}

		c.JSON(http.StatusOK, logs)
	})

	return router
}
