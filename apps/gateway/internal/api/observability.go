package api

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDContextKey = "request_id"
const requestPathContextKey = "request_path"

type observabilityConfig struct {
	requestIDHeader string
	trustRequestID  bool
}

// attachRequestID 负责读取或生成 request-id，并同时写入请求、响应和上下文，便于上游透传与日志关联。
func attachRequestID(cfg observabilityConfig) gin.HandlerFunc {
	header := strings.TrimSpace(cfg.requestIDHeader)
	if header == "" {
		header = "X-Request-ID"
	}

	return func(c *gin.Context) {
		requestID := ""
		if cfg.trustRequestID {
			requestID = sanitizeRequestID(c.GetHeader(header))
		}
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(requestPathContextKey, c.Request.URL.Path)
		c.Request.Header.Set(header, requestID)
		c.Request.Header.Set("X-Request-ID", requestID)
		c.Writer.Header().Set(header, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Set(requestIDContextKey, requestID)
		c.Next()
	}
}

// logRequest 输出结构化 access log，只记录路由、状态和 request-id，避免把 token 等敏感信息写入日志。
func logRequest(cfg observabilityConfig) gin.HandlerFunc {
	header := strings.TrimSpace(cfg.requestIDHeader)
	if header == "" {
		header = "X-Request-ID"
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		requestID := c.GetString(requestIDContextKey)
		if requestID == "" {
			requestID = sanitizeRequestID(c.Writer.Header().Get(header))
		}
		requestPath := c.GetString(requestPathContextKey)
		if requestPath == "" {
			requestPath = c.Request.URL.Path
		}

		log.Printf(
			"gateway_request method=%s path=%s status=%d duration_ms=%d ip=%s request_id=%s user_agent=%q",
			c.Request.Method,
			requestPath,
			c.Writer.Status(),
			latency.Milliseconds(),
			c.ClientIP(),
			requestID,
			c.Request.UserAgent(),
		)
	}
}

func newRequestID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}

func sanitizeRequestID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return ""
	}
	for _, ch := range value {
		if ch < 33 || ch > 126 {
			return ""
		}
	}
	return value
}
