package api

import (
	"net/http"

	"crawler-platform/apps/project-service/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(projectService *service.ProjectService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/projects", func(c *gin.Context) {
		var req struct {
			Code string `json:"code" binding:"required"`
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		project := projectService.Create(req.Code, req.Name)
		c.JSON(http.StatusCreated, project)
	})

	router.GET("/api/v1/projects", func(c *gin.Context) {
		c.JSON(http.StatusOK, projectService.List())
	})

	return router
}
