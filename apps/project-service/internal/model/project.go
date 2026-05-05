// Package model 定义 Project 服务的数据结构。
// Project 是全层共享的领域模型，不含 GORM 标签或校验逻辑。
package model

// Project 表示一个爬虫项目，ID 由 service 层生成（UUID）。
// OrganizationID 是租户隔离标识，当前允许为空（向后兼容），Phase 4 加固时改为必填。
type Project struct {
	ID             string `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	OrganizationID string `json:"organizationId,omitempty"`
}
