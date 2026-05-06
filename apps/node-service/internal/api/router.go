package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"crawler-platform/apps/node-service/internal/service"
	"github.com/gin-gonic/gin"
)

// parseQueryInt 解析查询参数为整数，解析失败时返回默认值。
func parseQueryInt(c *gin.Context, key string, defaultVal int) int {
	val, err := strconv.Atoi(c.Query(key))
	if err != nil || val < 0 {
		return defaultVal
	}
	return val
}

func NewRouter(nodeService *service.NodeService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/nodes/:id/heartbeat", func(c *gin.Context) {
		var req struct {
			Capabilities []string `json:"capabilities"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		id := c.Param("id")
		if err := validateNodeID(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orgID := c.GetHeader("X-Org-ID")
		node, err := nodeService.Heartbeat(orgID, id, req.Capabilities)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, node)
	})

	router.GET("/api/v1/nodes", func(c *gin.Context) {
		limit := parseQueryInt(c, "limit", 20)
		offset := parseQueryInt(c, "offset", 0)
		orgID := c.GetHeader("X-Org-ID")
		nodes, err := nodeService.List(orgID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nodes)
	})

	router.GET("/api/v1/nodes/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := validateNodeID(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orgID := c.GetHeader("X-Org-ID")
		node, err := nodeService.GetByID(orgID, id)
		if err != nil {
			if err == service.ErrNodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, node)
	})

	router.GET("/api/v1/nodes/:id/sessions", func(c *gin.Context) {
		id := c.Param("id")
		if err := validateNodeID(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		limit := parseQueryInt(c, "limit", 20)
		status := c.Query("executionStatus")
		var from, to *time.Time
		if fromStr := c.Query("executionFrom"); fromStr != "" {
			if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
				from = &t
			}
		}
		if toStr := c.Query("executionTo"); toStr != "" {
			if t, err := time.Parse(time.RFC3339, toStr); err == nil {
				to = &t
			}
		}
		executionOffset := parseQueryInt(c, "executionOffset", 0)

		query := service.ExecutionQuery{
			Status: status,
			From:   from,
			To:     to,
			Limit:  limit,
			Offset: executionOffset,
		}
		orgID := c.GetHeader("X-Org-ID")
		executions, err := nodeService.ListRecentExecutions(orgID, id, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, executions)
	})

	return router
}

func validateNodeID(id string) error {
	if id == "" {
		return fmt.Errorf("node id is required")
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			continue
		}
		return fmt.Errorf("invalid node id %q: use only letters, numbers, dash, underscore, or dot", id)
	}
	return nil
}
