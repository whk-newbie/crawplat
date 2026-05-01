package api

import (
	"net/http"
	"strconv"

	"crawler-platform/apps/scheduler-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
	"github.com/gin-gonic/gin"
)

func NewRouter(schedulerService *service.SchedulerService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/schedules", func(c *gin.Context) {
		var req struct {
			ProjectID         string   `json:"projectId" binding:"required"`
			SpiderID          string   `json:"spiderId" binding:"required"`
			Name              string   `json:"name" binding:"required"`
			CronExpr          string   `json:"cronExpr" binding:"required"`
			Enabled           bool     `json:"enabled"`
			Image             string   `json:"image" binding:"required"`
			Command           []string `json:"command"`
			RetryLimit        int      `json:"retryLimit"`
			RetryDelaySeconds int      `json:"retryDelaySeconds"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		schedule, err := schedulerService.Create(req.ProjectID, req.SpiderID, req.Name, req.CronExpr, req.Image, req.Command, req.Enabled, req.RetryLimit, req.RetryDelaySeconds)
		if err != nil {
			if err == service.ErrInvalidSchedule {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, schedule)
	})

	router.GET("/api/v1/schedules", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		p := httpx.DefaultPagination(limit, offset)

		schedules, total, err := schedulerService.List(p.Limit, p.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, httpx.PaginatedResponse{
			Items:  schedules,
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
		})
	})

	return router
}
