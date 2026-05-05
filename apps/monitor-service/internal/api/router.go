package api

import (
	"net/http"
	"strconv"

	"crawler-platform/apps/monitor-service/internal/service"
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
			Enabled             bool   `json:"enabled"`
			WebhookURL          string `json:"webhookUrl" binding:"required"`
			CooldownSeconds     int    `json:"cooldownSeconds"`
			TimeoutSeconds      int    `json:"timeoutSeconds"`
			OfflineGraceSeconds int    `json:"offlineGraceSeconds"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		rule, err := monitorService.CreateAlertRule(service.CreateAlertRuleInput{
			Name:                req.Name,
			RuleType:            req.RuleType,
			Enabled:             req.Enabled,
			WebhookURL:          req.WebhookURL,
			CooldownSeconds:     req.CooldownSeconds,
			TimeoutSeconds:      req.TimeoutSeconds,
			OfflineGraceSeconds: req.OfflineGraceSeconds,
		})
		if err != nil {
			switch err {
			case service.ErrInvalidRuleName, service.ErrInvalidRuleType, service.ErrInvalidWebhookURL:
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		rule, err := monitorService.UpdateAlertRule(service.UpdateAlertRuleInput{
			ID:                  c.Param("id"),
			Name:                req.Name,
			Enabled:             req.Enabled,
			WebhookURL:          req.WebhookURL,
			CooldownSeconds:     req.CooldownSeconds,
			TimeoutSeconds:      req.TimeoutSeconds,
			OfflineGraceSeconds: req.OfflineGraceSeconds,
		})
		if err != nil {
			if err == service.ErrAlertRuleNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else if err == service.ErrInvalidRuleID {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, rule)
	})

	router.GET("/api/v1/monitor/alerts/events", func(c *gin.Context) {
		limit := parseQueryInt(c, "limit", 20)
		offset := parseQueryInt(c, "offset", 0)
		events, err := monitorService.ListAlertEvents(limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, events)
	})

	return router
}
