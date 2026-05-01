package api

import (
	"net/http"

	"crawler-platform/apps/iam-service/internal/service"
	"github.com/gin-gonic/gin"
)

func NewRouter(authService *service.AuthService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := authService.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	router.POST("/api/v1/auth/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required,min=6"`
			Email    string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := authService.Register(req.Username, req.Password, req.Email)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}

		c.JSON(http.StatusCreated, user)
	})

	return router
}
