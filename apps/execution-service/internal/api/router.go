package api

import (
	"context"
	"errors"
	"net/http"
	"os"

	"crawler-platform/apps/execution-service/internal/service"
	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

func NewRouter(executionService *service.ExecutionService) *gin.Engine {
	router := gin.Default()

	createExecutionHandler := func(c *gin.Context) {
		var req struct {
			ProjectID     string   `json:"projectId" binding:"required"`
			SpiderID      string   `json:"spiderId" binding:"required"`
			Image         string   `json:"image" binding:"required"`
			Command       []string `json:"command"`
			TriggerSource string   `json:"triggerSource"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, err := executionService.Create(context.Background(), service.CreateExecutionInput{
			ProjectID:     req.ProjectID,
			SpiderID:      req.SpiderID,
			Image:         req.Image,
			Command:       req.Command,
			TriggerSource: req.TriggerSource,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, exec)
	}

	appendLogHandler := func(c *gin.Context) {
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
	}

	getExecutionHandler := func(c *gin.Context) {
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
	}

	getLogsHandler := func(c *gin.Context) {
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
	}

	router.POST("/api/v1/executions", createExecutionHandler)
	router.POST("/api/v1/executions/:id/logs", appendLogHandler)
	router.GET("/api/v1/executions/:id", getExecutionHandler)
	router.GET("/api/v1/executions/:id/logs", getLogsHandler)

	internalExecution := router.Group("/internal/v1/executions", requireInternalToken())

	internalExecution.POST("/claim", func(c *gin.Context) {
		var req struct {
			NodeID string `json:"nodeId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, ok, err := executionService.ClaimNext(context.Background(), req.NodeID)
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else if errors.Is(err, service.ErrInvalidExecutionState) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		if !ok {
			c.Status(http.StatusNoContent)
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	internalExecution.POST("/:id/start", func(c *gin.Context) {
		var req struct {
			NodeID string `json:"nodeId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, err := executionService.Start(context.Background(), c.Param("id"), req.NodeID)
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else if errors.Is(err, service.ErrInvalidExecutionState) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	internalExecution.POST("/:id/logs", appendLogHandler)

	internalExecution.POST("/:id/complete", func(c *gin.Context) {
		exec, err := executionService.Complete(context.Background(), c.Param("id"))
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else if errors.Is(err, service.ErrInvalidExecutionState) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	internalExecution.POST("/:id/fail", func(c *gin.Context) {
		var req struct {
			Error string `json:"error" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, err := executionService.Fail(context.Background(), c.Param("id"), req.Error)
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else if errors.Is(err, service.ErrInvalidExecutionState) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	return router
}

func requireInternalToken() gin.HandlerFunc {
	token := os.Getenv("INTERNAL_API_TOKEN")
	if token == "" {
		token = os.Getenv("JWT_SECRET")
	}

	return func(c *gin.Context) {
		if token == "" || c.GetHeader(internalTokenHeader) != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized internal route"})
			return
		}
		c.Next()
	}
}
