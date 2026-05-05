// HTTP 路由和请求处理层。
// 负责注册所有执行服务的公开 API（/api/v1/executions）和内部 API（/internal/v1/executions），
// 包括请求参数绑定（ShouldBindJSON）、调用 service 层方法、错误码映射。
// 不包含业务逻辑——状态机、队列语义、持久化由 service 层处理；本层仅负责 HTTP 层面的输入输出。
package api

import (
	"context"
	"errors"
	"net/http"
	"os"

	"crawler-platform/apps/execution-service/internal/service"
	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

// NewRouter 创建并配置 gin 路由引擎，注册所有公开和内部路由。
//
// 公开 API（无需认证）：
//   - POST   /api/v1/executions         创建执行
//   - GET    /api/v1/executions/:id      查询执行详情
//   - GET    /api/v1/executions/:id/logs 查询执行日志
//   - POST   /api/v1/executions/:id/logs 追加日志
//
// 内部 API（需 X-Internal-Token 认证）：
//   - POST   /internal/v1/executions/claim                认领下一个待执行任务
//   - POST   /internal/v1/executions/:id/start            确认执行已启动
//   - POST   /internal/v1/executions/:id/complete         标记执行完成
//   - POST   /internal/v1/executions/:id/fail             标记执行失败
//   - POST   /internal/v1/executions/:id/logs             追加日志（内部调用）
//   - POST   /internal/v1/executions/retries/materialize  物化下一个重试候选
func NewRouter(executionService *service.ExecutionService) *gin.Engine {
	router := gin.Default()

	createExecutionHandler := func(c *gin.Context) {
		var req struct {
			ProjectID         string   `json:"projectId" binding:"required"`
			SpiderID          string   `json:"spiderId" binding:"required"`
			Image             string   `json:"image" binding:"required"`
			Command           []string `json:"command"`
			TriggerSource     string   `json:"triggerSource"`
			RetryLimit        int      `json:"retryLimit"`
			RetryCount        int      `json:"retryCount"`
			RetryDelaySeconds int      `json:"retryDelaySeconds"`
			RetryOfExecutionID string  `json:"retryOfExecutionId"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, err := executionService.Create(context.Background(), service.CreateExecutionInput{
			ProjectID:         req.ProjectID,
			SpiderID:          req.SpiderID,
			Image:             req.Image,
			Command:           req.Command,
			TriggerSource:     req.TriggerSource,
			RetryLimit:        req.RetryLimit,
			RetryCount:        req.RetryCount,
			RetryDelaySeconds: req.RetryDelaySeconds,
			RetryOfExecutionID: req.RetryOfExecutionID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, exec)
	}

	appendLogHandler := func(c *gin.Context) {
		var req struct {
			Message string `json:"message" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		entry, err := executionService.AppendLog(context.Background(), c.Param("id"), req.Message)
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, entry)
	}

	getExecutionHandler := func(c *gin.Context) {
		exec, err := executionService.Get(context.Background(), c.Param("id"))
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, exec)
	}

	getLogsHandler := func(c *gin.Context) {
		logs, err := executionService.GetLogs(context.Background(), c.Param("id"))
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, logs)
	}

	router.POST("/api/v1/executions", createExecutionHandler)
	router.POST("/api/v1/executions/:id/logs", appendLogHandler)
	router.GET("/api/v1/executions/:id", getExecutionHandler)
	router.GET("/api/v1/executions/:id/logs", getLogsHandler)

	internalExecution := router.Group("/internal/v1/executions", requireInternalToken())

	internalExecution.POST("/claim", func(c *gin.Context) {
		var req struct {
			NodeID string `json:"nodeId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, ok, err := executionService.ClaimNext(context.Background(), req.NodeID)
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else if errors.Is(err, service.ErrInvalidExecutionState) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		if !ok {
			c.Status(http.StatusNoContent)
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	internalExecution.POST("/:id/start", func(c *gin.Context) {
		var req struct {
			NodeID string `json:"nodeId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, err := executionService.Start(context.Background(), c.Param("id"), req.NodeID)
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else if errors.Is(err, service.ErrInvalidExecutionState) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	internalExecution.POST("/:id/logs", appendLogHandler)

	internalExecution.POST("/:id/complete", func(c *gin.Context) {
		exec, err := executionService.Complete(context.Background(), c.Param("id"))
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else if errors.Is(err, service.ErrInvalidExecutionState) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	internalExecution.POST("/:id/fail", func(c *gin.Context) {
		var req struct {
			Error string `json:"error" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		exec, err := executionService.Fail(context.Background(), c.Param("id"), req.Error)
		if err != nil {
			if errors.Is(err, service.ErrExecutionNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
			} else if errors.Is(err, service.ErrInvalidExecutionState) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, exec)
	})

	internalExecution.POST("/retries/materialize", func(c *gin.Context) {
		exec, ok, err := executionService.MaterializeRetry(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusCreated, exec)
	})

	return router
}

// requireInternalToken 返回一个 gin 中间件，校验请求头 X-Internal-Token。
// 优先读取环境变量 INTERNAL_API_TOKEN，回退到 JWT_SECRET。
// 如果环境变量为空（开发环境），任何请求都会被拒绝（安全优先原则）。
func requireInternalToken() gin.HandlerFunc {
	token := os.Getenv("INTERNAL_API_TOKEN")
	if token == "" {
		token = os.Getenv("JWT_SECRET")
	}

	return func(c *gin.Context) {
		if token == "" || c.GetHeader(internalTokenHeader) != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized internal route"})
			return
		}
		c.Next()
	}
}
