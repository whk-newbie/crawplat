package api

import (
	"context"
	"errors"
	"net/http"

	"crawler-platform/apps/execution-service/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(executionService *service.ExecutionService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/executions", func(c *gin.Context) {
		var req struct {
			ProjectID string   `json:"projectId" binding:"required"`
			SpiderID  string   `json:"spiderId" binding:"required"`
			Image     string   `json:"image" binding:"required"`
			Command   []string `json:"command"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, err := executionService.CreateManual(context.Background(), service.CreateManualInput{
			ProjectID: req.ProjectID,
			SpiderID:  req.SpiderID,
			Image:     req.Image,
			Command:   req.Command,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
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

		entry, err := executionService.AppendLog(context.Background(), c.Param("id"), req.Message)
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, entry)
	})

	router.GET("/api/v1/executions/:id", func(c *gin.Context) {
		exec, err := executionService.Get(context.Background(), c.Param("id"))
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	router.GET("/api/v1/executions/:id/logs", func(c *gin.Context) {
		logs, err := executionService.GetLogs(context.Background(), c.Param("id"))
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, logs)
	})

	return router
}
