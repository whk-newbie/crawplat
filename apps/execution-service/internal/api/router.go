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
		projectID := firstQuery(c, "project_id", "projectId")
		limit, err := parseIntQuery(c, "limit", service.DefaultExecutionListLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
			return
		}
		offset, err := parseIntQuery(c, "offset", 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be an integer"})
			return
		}
		spiderID := firstQuery(c, "spider_id", "spiderId")
		nodeID := firstQuery(c, "node_id", "nodeId")
		status := firstQuery(c, "status", "executionStatus")
		triggerSource := firstQuery(c, "trigger_source", "executionTriggerSource")
		executionFrom, err := parseRFC3339Query(c, "from", "executionFrom")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "from must be RFC3339"})
			return
		}
		executionTo, err := parseRFC3339Query(c, "to", "executionTo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "to must be RFC3339"})
			return
		}

		items, total, err := executionService.List(context.Background(), service.ListExecutionsQuery{
			ProjectID: projectID,
			SpiderID:  spiderID,
			NodeID:    nodeID,
			Status:    status,
			Trigger:   triggerSource,
			From:      executionFrom,
			To:        executionTo,
			Limit:     limit,
			Offset:    offset,
			SortBy:    firstQuery(c, "sort_by", "sortBy"),
			SortOrder: firstQuery(c, "sort_order", "sortOrder"),
		})
		if err != nil {
			if errors.Is(err, service.ErrInvalidExecutionListQuery) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		normalized, _ := (service.ListExecutionsQuery{
			ProjectID: projectID,
			SpiderID:  spiderID,
			NodeID:    nodeID,
			Status:    status,
			Trigger:   triggerSource,
			From:      executionFrom,
			To:        executionTo,
			Limit:     limit,
			Offset:    offset,
			SortBy:    firstQuery(c, "sort_by", "sortBy"),
			SortOrder: firstQuery(c, "sort_order", "sortOrder"),
		}).Normalize()
		c.JSON(http.StatusOK, httpx.PaginatedResponse{
			Items:  items,
			Total:  total,
			Limit:  normalized.Limit,
			Offset: normalized.Offset,
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

func firstQuery(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			return value
		}
	}
	return ""
}

func parseIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(raw)
}

func parseRFC3339Query(c *gin.Context, keys ...string) (*time.Time, error) {
	raw := firstQuery(c, keys...)
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
