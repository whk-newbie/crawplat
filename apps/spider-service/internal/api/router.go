package api

import (
	"net/http"
	"strconv"

	"crawler-platform/apps/spider-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
	"github.com/gin-gonic/gin"
)

func NewRouter(spiderService *service.SpiderService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/projects/:projectId/spiders", func(c *gin.Context) {
		var req struct {
			Name     string   `json:"name" binding:"required"`
			Language string   `json:"language" binding:"required"`
			Runtime  string   `json:"runtime" binding:"required"`
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
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		p := httpx.DefaultPagination(limit, offset)

		spiders, total, err := spiderService.List(c.Param("projectId"), p.Limit, p.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, httpx.PaginatedResponse{
			Items:  spiders,
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
		})
	})

	router.POST("/api/v1/spiders/:spiderId/versions", func(c *gin.Context) {
		var req struct {
			Image   string   `json:"image"`
			Command []string `json:"command"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		version, err := spiderService.CreateVersion(c.Param("spiderId"), req.Image, req.Command)
		if err != nil {
			switch err {
			case service.ErrSpiderNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case service.ErrImageRequired:
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusCreated, version)
	})

	router.GET("/api/v1/spiders/:spiderId/versions", func(c *gin.Context) {
		versions, err := spiderService.ListVersions(c.Param("spiderId"))
		if err != nil {
			switch err {
			case service.ErrSpiderNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, versions)
	})

	return router
}
