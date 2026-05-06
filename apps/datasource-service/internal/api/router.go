package api

import (
	"net/http"
	"strconv"

	"crawler-platform/apps/datasource-service/internal/service"
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

func NewRouter(datasourceService *service.DatasourceService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/datasources", func(c *gin.Context) {
		var req struct {
			ProjectID string            `json:"projectId" binding:"required"`
			Name      string            `json:"name" binding:"required"`
			Type      string            `json:"type" binding:"required"`
			Config    map[string]string `json:"config"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		orgID := c.GetHeader("X-Org-ID")
			datasource, err := datasourceService.Create(orgID, req.ProjectID, req.Name, req.Type, req.Config)
		if err != nil {
			if err == service.ErrInvalidDatasourceType {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, datasource)
	})

	router.GET("/api/v1/datasources", func(c *gin.Context) {
		limit := parseQueryInt(c, "limit", 20)
		offset := parseQueryInt(c, "offset", 0)
		orgID := c.GetHeader("X-Org-ID")
			datasources, err := datasourceService.List(orgID, c.Query("projectId"), limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, datasources)
	})

	router.POST("/api/v1/datasources/:id/test", func(c *gin.Context) {
		result, err := datasourceService.Test(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	router.POST("/api/v1/datasources/:id/preview", func(c *gin.Context) {
		result, err := datasourceService.Preview(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	return router
}
