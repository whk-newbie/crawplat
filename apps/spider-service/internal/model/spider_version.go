// Package model 定义 Spider 版本模型，不含业务逻辑。
package model

// SpiderVersion 表示爬虫的一个可部署版本。
// OrganizationID 是租户隔离标识，当前允许为空（向后兼容），Phase 4 加固时改为必填。
type SpiderVersion struct {
	ID              string   `json:"id"`
	SpiderID        string   `json:"spiderId"`
	OrganizationID  string   `json:"organizationId,omitempty"`
	Version         string   `json:"version"`
	Image           string   `json:"image,omitempty"`
	RegistryAuthRef string   `json:"registryAuthRef,omitempty"`
	Command         []string `json:"command,omitempty"`
}
