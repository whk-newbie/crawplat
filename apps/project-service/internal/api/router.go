package api

import (
	"net/http"
	"strconv"

	"crawler-platform/apps/project-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
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

		project, err := projectService.Create(req.Code, req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, project)
	})

	router.GET("/api/v1/projects", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		p := httpx.DefaultPagination(limit, offset)

		projects, total, err := projectService.List(p.Limit, p.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, httpx.PaginatedResponse{
			Items:  projects,
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
		})
	})

	return router
}
