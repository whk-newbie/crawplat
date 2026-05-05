package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func ResolveServiceURL(name string) string {
	switch name {
	case "iam-service":
		return "http://iam-service:8081"
	case "project-service":
		return "http://project-service:8082"
	case "spider-service":
		return "http://spider-service:8083"
	case "execution-service":
		return "http://execution-service:8085"
	case "node-service":
		return "http://node-service:8084"
	case "datasource-service":
		return "http://datasource-service:8086"
	case "scheduler-service":
		return "http://scheduler-service:8087"
	case "monitor-service":
		return "http://monitor-service:8088"
	default:
		return ""
	}
}

func ProxyTo(serviceName string) gin.HandlerFunc {
	target := ResolveServiceURL(serviceName)
	if target == "" {
		return func(c *gin.Context) {
			c.JSON(http.StatusBadGateway, gin.H{"error": "unknown upstream service"})
		}
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid upstream service url"})
		}
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(gin.H{"error": "upstream service unavailable"})
	}

	return func(c *gin.Context) {
		reverseProxy.ServeHTTP(c.Writer, c.Request)
	}
}
