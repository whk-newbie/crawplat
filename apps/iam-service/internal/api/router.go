// Package api 是 IAM 服务的 HTTP 路由层。
// 负责注册认证相关端点（POST /api/v1/auth/login、POST /api/v1/auth/register）、
// 解析请求参数、调用 service 层并将结果序列化为 JSON 响应。
// 不处理 JWT 签发逻辑——该职责属于 service/auth。
package api

import (
	"net/http"
	"strings"

	"crawler-platform/apps/iam-service/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 创建并配置 Gin 路由引擎。
// 注册 POST /api/v1/auth/login：接收 username/password JSON，
// 校验失败返回 400，认证失败返回 401，成功返回 {"token": "...", "organizations": [...]}。
// 注册 POST /api/v1/auth/register：接收 username/password JSON，
// 校验失败返回 400，用户名已存在返回 409，成功返回 201。
func NewRouter(authService *service.AuthService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		result, err := authService.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":         result.Token,
			"organizations": result.Memberships,
		})
	})

	router.POST("/api/v1/auth/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		user, err := authService.Register(req.Username, req.Password)
		if err != nil {
			if err == service.ErrUserAlreadyExists {
				c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"username": user.Username})
	})

	return router
}
