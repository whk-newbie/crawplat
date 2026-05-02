package api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"crawler-platform/apps/execution-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

func NewRouter(executionService *service.ExecutionService) *gin.Engine {
	router := gin.Default()

	createExecutionHandler := func(c *gin.Context) {
		var req struct {
			ProjectID          string   `json:"projectId" binding:"required"`
			SpiderID           string   `json:"spiderId" binding:"required"`
			SpiderVersion      int      `json:"spiderVersion"`
			RegistryAuthRef    string   `json:"registryAuthRef"`
			Image              string   `json:"image"`
			Command            []string `json:"command"`
			CPUCores           float64  `json:"cpuCores"`
			MemoryMB           int      `json:"memoryMB"`
			TimeoutSeconds     int      `json:"timeoutSeconds"`
			TriggerSource      string   `json:"triggerSource"`
			RetryLimit         int      `json:"retryLimit"`
			RetryCount         int      `json:"retryCount"`
			RetryDelaySeconds  int      `json:"retryDelaySeconds"`
			RetryOfExecutionID string   `json:"retryOfExecutionId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, err := executionService.Create(context.Background(), service.CreateExecutionInput{
			ProjectID:          req.ProjectID,
			SpiderID:           req.SpiderID,
			SpiderVersion:      req.SpiderVersion,
			RegistryAuthRef:    req.RegistryAuthRef,
			Image:              req.Image,
			Command:            req.Command,
			CPUCores:           req.CPUCores,
			MemoryMB:           req.MemoryMB,
			TimeoutSeconds:     req.TimeoutSeconds,
			TriggerSource:      req.TriggerSource,
			RetryLimit:         req.RetryLimit,
			RetryCount:         req.RetryCount,
			RetryDelaySeconds:  req.RetryDelaySeconds,
			RetryOfExecutionID: req.RetryOfExecutionID,
		})
		if err != nil {
			switch {
			case errors.Is(err, service.ErrExecutionImageRequired):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.Is(err, service.ErrSpiderVersionNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
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
	router.GET("/api/v1/executions", func(c *gin.Context) {
		projectID := c.Query("projectId")
		if projectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "projectId is required"})
			return
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		p := httpx.DefaultPagination(limit, offset)
		spiderID := strings.TrimSpace(c.Query("spiderId"))
		status := strings.TrimSpace(c.Query("executionStatus"))
		triggerSource := strings.TrimSpace(c.Query("executionTriggerSource"))
		executionFrom, err := parseRFC3339Query(c, "executionFrom")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "executionFrom must be RFC3339"})
			return
		}
		executionTo, err := parseRFC3339Query(c, "executionTo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "executionTo must be RFC3339"})
			return
		}
		if executionFrom != nil && executionTo != nil && executionFrom.After(*executionTo) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "executionFrom must be before or equal to executionTo"})
			return
		}

		items, total, err := executionService.List(context.Background(), service.ListExecutionsQuery{
			ProjectID: projectID,
			SpiderID:  spiderID,
			Status:    status,
			Trigger:   triggerSource,
			From:      executionFrom,
			To:        executionTo,
			Limit:     p.Limit,
			Offset:    p.Offset,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, httpx.PaginatedResponse{
			Items:  items,
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
		})
	})
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

	internalExecution.POST("/retries/materialize", func(c *gin.Context) {
		exec, ok, err := executionService.MaterializeRetry(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusCreated, exec)
	})

	return router
}

func parseRFC3339Query(c *gin.Context, key string) (*time.Time, error) {
	raw := c.Query(key)
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
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
