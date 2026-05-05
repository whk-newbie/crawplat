package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// respondError 统一 gateway 对外错误响应，保持 {"error": "message"} 契约稳定。
func respondError(c *gin.Context, status int, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		message = http.StatusText(status)
	}
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
