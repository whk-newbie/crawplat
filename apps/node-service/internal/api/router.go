// 文件职责：HTTP API 路由层，基于 Gin 框架注册节点相关的 REST 端点。
// 负责：请求解析、参数校验（如 nodeID 合法性）、调用 service 层方法、返回 JSON 响应。
//       错误统一包装为 {"error": "message"} 格式，符合 API_DESIGN.md 契约。
// 与谁交互：依赖 internal/service.NodeService（业务逻辑），被 main.go 调用以启动 HTTP 服务。
// 不负责：业务逻辑、数据持久化、在线/离线状态判定 —— 这些由 service 和 repo 包负责。
// 已注册端点：POST /api/v1/nodes/:id/heartbeat（心跳上报）、GET /api/v1/nodes（节点列表）。
//           节点详情（GET /api/v1/nodes/:id）和会话（GET /api/v1/nodes/:id/sessions）尚未实现。
package api

import (
	"fmt"
	"net/http"

	"crawler-platform/apps/node-service/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 创建并配置 Gin 路由引擎。
// 注册两个端点：
//   - POST /api/v1/nodes/:id/heartbeat：处理节点心跳上报，会校验 JSON body（capabilities 字段）和 nodeID 格式。
//   - GET  /api/v1/nodes：获取所有在线节点列表。
// 所有路由均通过 service.NodeService 间接访问数据层，路由层不感知存储实现（Redis/内存/Postgres）。
func NewRouter(nodeService *service.NodeService) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/nodes/:id/heartbeat", func(c *gin.Context) {
		var req struct {
			Capabilities []string `json:"capabilities"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id := c.Param("id")
		if err := validateNodeID(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		node, err := nodeService.Heartbeat(id, req.Capabilities)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, node)
	})

	router.GET("/api/v1/nodes", func(c *gin.Context) {
		nodes, err := nodeService.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nodes)
	})

	return router
}

// validateNodeID 校验 nodeID 的格式合法性。
// 只允许字母（大小写）、数字、短横线、下划线、点号，不允许空字符串。
// 目的：防止非法字符污染 Redis key 或造成路径注入，接口契约约定 nodeID 必须是 path 安全的。
func validateNodeID(id string) error {
	if id == "" {
		return fmt.Errorf("node id is required")
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			continue
		}
		return fmt.Errorf("invalid node id %q: use only letters, numbers, dash, underscore, or dot", id)
	}
	return nil
}
