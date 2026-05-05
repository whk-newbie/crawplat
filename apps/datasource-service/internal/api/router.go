// Package api 定义数据源服务的 HTTP 路由和处理函数，基于 Gin 框架。
//
// 本文件负责：
//   - 注册 RESTful API 路由：POST/GET datasources、POST test、POST preview
//   - HTTP 请求参数绑定与校验（通过 Gin 的 ShouldBindJSON）
//   - 错误到 HTTP 状态码的映射（如 ErrInvalidDatasourceType → 400、通用错误 → 500）
//   - JSON 响应序列化
//
// 路由表（与 API_DESIGN.md 契约一致）：
//   POST   /api/v1/datasources        - 创建数据源
//   GET    /api/v1/datasources         - 按项目列出数据源
//   POST   /api/v1/datasources/:id/test    - 连接测试
//   POST   /api/v1/datasources/:id/preview - 数据预览
//
// 不负责什么：不包含业务逻辑、数据校验、持久化——全部委托给 DatasourceService。
package api

import (
	"net/http"

	"crawler-platform/apps/datasource-service/internal/service"
	"github.com/gin-gonic/gin"
)

// NewRouter 创建并配置 Gin 路由器，注册所有数据源 API 端点。
// datasourceService 由调用方（main.go）注入，实现依赖反转。
// 返回的 *gin.Engine 可直接通过 Run 方法启动 HTTP 服务。
func NewRouter(datasourceService *service.DatasourceService) *gin.Engine {
	router := gin.Default()

	// POST /api/v1/datasources - 创建数据源
	// 请求体需包含 projectId、name、type（必填），config 为可选的 key-value 配置。
	// 成功返回 201 和新创建的数据源 JSON；类型不合法返回 400；其他错误返回 500。
	router.POST("/api/v1/datasources", func(c *gin.Context) {
		var req struct {
			ProjectID string            `json:"projectId" binding:"required"`
			Name      string            `json:"name" binding:"required"`
			Type      string            `json:"type" binding:"required"`
			Config    map[string]string `json:"config"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		datasource, err := datasourceService.Create(req.ProjectID, req.Name, req.Type, req.Config)
		if err != nil {
			if err == service.ErrInvalidDatasourceType {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, datasource)
	})

	// GET /api/v1/datasources?projectId=<id> - 按项目列出数据源
	// projectId 为可选查询参数；不传时（memoryRepository 模式）返回所有数据源。
	router.GET("/api/v1/datasources", func(c *gin.Context) {
		datasources, err := datasourceService.List(c.Query("projectId"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, datasources)
	})

	// POST /api/v1/datasources/:id/test - 连接测试
	// 对指定数据源执行连接测试，当前版本返回 mock 结果。
	// 数据源不存在时返回 404。
	router.POST("/api/v1/datasources/:id/test", func(c *gin.Context) {
		result, err := datasourceService.Test(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	// POST /api/v1/datasources/:id/preview - 数据预览
	// 对指定数据源执行数据预览，当前版本返回 mock 样本数据。
	// 数据源不存在时返回 404。
	router.POST("/api/v1/datasources/:id/preview", func(c *gin.Context) {
		result, err := datasourceService.Preview(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	return router
}
