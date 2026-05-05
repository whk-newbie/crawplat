// Package model 定义 Spider 版本模型，不含业务逻辑。
package model

// SpiderVersion 表示爬虫的一个可部署版本。
type SpiderVersion struct {
	ID              string   `json:"id"`
	SpiderID        string   `json:"spiderId"`
	Version         string   `json:"version"`
	Image           string   `json:"image,omitempty"`
	RegistryAuthRef string   `json:"registryAuthRef,omitempty"`
	Command         []string `json:"command,omitempty"`
}
