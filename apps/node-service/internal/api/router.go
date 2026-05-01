package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
		nodes, err := nodeService.List()
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

		limit := 20
		if raw := c.Query("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
				return
			}
			limit = parsed
		}

		detail, err := nodeService.Detail(id, limit)
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
