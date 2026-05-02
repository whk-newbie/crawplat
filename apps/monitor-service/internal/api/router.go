package api

import (
	"errors"
	"net/http"
	"strconv"

	"crawler-platform/apps/monitor-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
	"github.com/gin-gonic/gin"
)

func NewRouter(monitorService *service.MonitorService) *gin.Engine {
	router := gin.Default()

	router.GET("/api/v1/monitor/overview", func(c *gin.Context) {
		overview, err := monitorService.Overview()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, overview)
	})

	router.POST("/api/v1/monitor/alerts/rules", func(c *gin.Context) {
		var req struct {
			Name                string `json:"name" binding:"required"`
			RuleType            string `json:"ruleType" binding:"required"`
			Enabled             *bool  `json:"enabled"`
			WebhookURL          string `json:"webhookUrl" binding:"required"`
			CooldownSeconds     int    `json:"cooldownSeconds"`
			TimeoutSeconds      int    `json:"timeoutSeconds"`
			OfflineGraceSeconds int    `json:"offlineGraceSeconds"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		enabled := true
		if req.Enabled != nil {
			enabled = *req.Enabled
		}
		rule, err := monitorService.CreateAlertRule(service.CreateAlertRuleInput{
			Name:                req.Name,
			RuleType:            req.RuleType,
			Enabled:             enabled,
			WebhookURL:          req.WebhookURL,
			CooldownSeconds:     req.CooldownSeconds,
			TimeoutSeconds:      req.TimeoutSeconds,
			OfflineGraceSeconds: req.OfflineGraceSeconds,
		})
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidRuleType),
				errors.Is(err, service.ErrInvalidWebhookURL),
				errors.Is(err, service.ErrInvalidRuleName):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusCreated, rule)
	})

	router.GET("/api/v1/monitor/alerts/rules", func(c *gin.Context) {
		rules, err := monitorService.ListAlertRules()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, rules)
	})

	router.PATCH("/api/v1/monitor/alerts/rules/:id", func(c *gin.Context) {
		var req struct {
			Name                *string `json:"name"`
			Enabled             *bool   `json:"enabled"`
			WebhookURL          *string `json:"webhookUrl"`
			CooldownSeconds     *int    `json:"cooldownSeconds"`
			TimeoutSeconds      *int    `json:"timeoutSeconds"`
			OfflineGraceSeconds *int    `json:"offlineGraceSeconds"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rule, err := monitorService.UpdateAlertRule(c.Param("id"), service.UpdateAlertRuleInput{
			Name:                req.Name,
			Enabled:             req.Enabled,
			WebhookURL:          req.WebhookURL,
			CooldownSeconds:     req.CooldownSeconds,
			TimeoutSeconds:      req.TimeoutSeconds,
			OfflineGraceSeconds: req.OfflineGraceSeconds,
		})
		if err != nil {
			switch {
			case errors.Is(err, service.ErrRuleNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case errors.Is(err, service.ErrInvalidWebhookURL),
				errors.Is(err, service.ErrInvalidRuleName):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, rule)
	})

	router.GET("/api/v1/monitor/alerts/events", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		p := httpx.DefaultPagination(limit, offset)

		items, total, normalizedLimit, normalizedOffset, err := monitorService.ListAlertEvents(p.Limit, p.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, httpx.PaginatedResponse{
			Items:  items,
			Total:  total,
			Limit:  normalizedLimit,
			Offset: normalizedOffset,
		})
	})

	return router
}
