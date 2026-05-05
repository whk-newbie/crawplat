// Package model 定义数据源服务的核心数据结构，包括数据源实体、连接测试结果和数据预览结果。
// 本文件仅负责数据结构定义与 JSON 序列化标签，不包含任何业务逻辑或数据库操作。
package model

// Datasource 表示一个已注册的外部数据源配置。
// Config 字段为扁平化的 key-value 配置（如 host、port、password 等），不同数据源类型使用不同键值。
// Readonly 字段固定为 true，表示爬虫平台对所有外部数据源均为只读访问。
// OrganizationID 是租户隔离标识，当前允许为空（向后兼容），Phase 4 加固时改为必填。
type Datasource struct {
	ID             string            `json:"id"`
	ProjectID      string            `json:"projectId"`
	OrganizationID string            `json:"organizationId,omitempty"`
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Readonly       bool              `json:"readonly"`
	Config         map[string]string `json:"config,omitempty"`
}

// TestResult 表示连接测试的结果，由真实探针（非 mock）探测外部数据源后返回。
// Status 为 "ok" 表示连接成功，其他值表示失败。
// Message 包含成功确认或失败原因的描述信息。
type TestResult struct {
	DatasourceID string `json:"datasourceId"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// PreviewResult 表示数据预览的结果，由真实查询（非 mock）返回外部数据源的结构信息。
// Rows 中的每条记录为 key-value 键值对，其结构取决于数据源类型：
//   - PostgreSQL: 返回表名（schema + table）
//   - Redis: 返回 key、类型以及字符串类型的值截断
//   - MongoDB: 返回数据库名或集合名
type PreviewResult struct {
	DatasourceID   string              `json:"datasourceId"`
	DatasourceType string              `json:"datasourceType"`
	Rows           []map[string]string `json:"rows"`
}
