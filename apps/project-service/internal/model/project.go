// Package model 定义 Project 服务的数据结构。
// Project 是全层共享的领域模型，不含 GORM 标签或校验逻辑。
package model

// Project 表示一个爬虫项目，ID 由 service 层生成（UUID）。
type Project struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
