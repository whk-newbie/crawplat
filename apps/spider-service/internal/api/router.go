package api

import (
	"net/http"

	"crawler-platform/apps/spider-service/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(spiderService *service.SpiderService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/projects/:projectId/spiders", func(c *gin.Context) {
		var req struct {
			Name     string `json:"name" binding:"required"`
			Language string `json:"language" binding:"required"`
			Runtime  string `json:"runtime" binding:"required"`
			Image    string   `json:"image"`
			Command  []string `json:"command"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		spider, err := spiderService.Create(c.Param("projectId"), req.Name, req.Language, req.Runtime, req.Image, req.Command)
		if err != nil {
			switch err {
			case service.ErrInvalidLanguage, service.ErrInvalidRuntime, service.ErrImageRequired:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, spider)
	})

	router.GET("/api/v1/projects/:projectId/spiders", func(c *gin.Context) {
		spiders, err := spiderService.List(c.Param("projectId"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, spiders)
	})

	return router
}
