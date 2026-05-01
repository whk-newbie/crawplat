package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"crawler-platform/apps/node-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
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

		id := c.Param("id")
		if err := validateNodeID(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		node, err := nodeService.Heartbeat(id, req.Capabilities)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, node)
	})

	router.GET("/api/v1/nodes", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		p := httpx.DefaultPagination(limit, offset)

		nodes, total, err := nodeService.List(p.Limit, p.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, httpx.PaginatedResponse{
			Items:  nodes,
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
		})
	})

	router.GET("/api/v1/nodes/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := validateNodeID(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		heartbeatLimit, err := parsePositiveIntQuery(c, "limit", 20)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
			return
		}
		executionLimit, err := parsePositiveIntQuery(c, "executionLimit", 20)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "executionLimit must be a positive integer"})
			return
		}
		executionOffset, err := parseNonNegativeIntQuery(c, "executionOffset", 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "executionOffset must be a non-negative integer"})
			return
		}
		executionStatus := c.Query("executionStatus")
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

		detail, err := nodeService.Detail(id, service.DetailQuery{
			HeartbeatLimit: heartbeatLimit,
			ExecutionQuery: service.ExecutionQuery{
				Limit:  executionLimit,
				Offset: executionOffset,
				Status: executionStatus,
				From:   executionFrom,
				To:     executionTo,
			},
		})
		if err != nil {
			if errors.Is(err, service.ErrNodeNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, detail)
	})

	router.GET("/api/v1/nodes/:id/sessions", func(c *gin.Context) {
		id := c.Param("id")
		if err := validateNodeID(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		limit, err := parsePositiveIntQuery(c, "limit", 20)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
			return
		}

		gapSeconds, err := parseIntRangeQuery(c, "gapSeconds", 60, 1, 3600)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gapSeconds must be an integer between 1 and 3600"})
			return
		}

		sessions, err := nodeService.Sessions(id, limit, gapSeconds)
		if err != nil {
			if errors.Is(err, service.ErrNodeNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sessions)
	})

	return router
}

func parsePositiveIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 0, errors.New("invalid positive int")
	}
	return parsed, nil
}

func parseNonNegativeIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 0 {
		return 0, errors.New("invalid non-negative int")
	}
	return parsed, nil
}

func parseIntRangeQuery(c *gin.Context, key string, defaultValue int, min int, max int) (int, error) {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < min || parsed > max {
		return 0, errors.New("out of range")
	}
	return parsed, nil
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
