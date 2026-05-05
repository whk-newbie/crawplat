package api

import (
	"net/http"

	"crawler-platform/apps/datasource-service/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(datasourceService *service.DatasourceService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/datasources", func(c *gin.Context) {
		var req struct {
			ProjectID string `json:"projectId" binding:"required"`
			Name      string `json:"name" binding:"required"`
			Type      string `json:"type" binding:"required"`
			Config    map[string]string `json:"config"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		datasource, err := datasourceService.Create(req.ProjectID, req.Name, req.Type, req.Config)
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
		datasources, err := datasourceService.List(c.Query("projectId"))
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
