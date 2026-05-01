package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	windowStart int64
	count       int
}

type fixedWindowLimiter struct {
	mu            sync.Mutex
	windowSeconds int64
	maxRequests   int
	buckets       map[string]rateBucket
}

func newFixedWindowLimiter(windowSeconds int, maxRequests int) *fixedWindowLimiter {
	if windowSeconds <= 0 {
		windowSeconds = 60
	}
	if maxRequests <= 0 {
		maxRequests = 120
	}
	return &fixedWindowLimiter{
		windowSeconds: int64(windowSeconds),
		maxRequests:   maxRequests,
		buckets:       map[string]rateBucket{},
	}
}

func (l *fixedWindowLimiter) allow(key string, now time.Time) bool {
	if key == "" {
		key = "unknown"
	}
	windowStart := now.Unix() / l.windowSeconds

	l.mu.Lock()
	defer l.mu.Unlock()

	bucket := l.buckets[key]
	if bucket.windowStart != windowStart {
		bucket = rateBucket{windowStart: windowStart, count: 0}
	}
	if bucket.count >= l.maxRequests {
		l.buckets[key] = bucket
		return false
	}
	bucket.count++
	l.buckets[key] = bucket
	return true
}

func requireRateLimit(cfg rateLimitConfig) gin.HandlerFunc {
	limiter := newFixedWindowLimiter(cfg.windowSeconds, cfg.maxRequests)
	return func(c *gin.Context) {
		if !limiter.allow(rateLimitKey(c), time.Now()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

func rateLimitKey(c *gin.Context) string {
	forwarded := strings.TrimSpace(c.GetHeader("X-Forwarded-For"))
	if forwarded != "" {
		first := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		if first != "" {
			return first
		}
	}
	return c.ClientIP()
}
