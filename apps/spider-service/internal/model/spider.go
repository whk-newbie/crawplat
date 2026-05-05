// 该文件定义 Spider 领域模型结构体及其 JSON 序列化标签。
// 纯数据结构，不包含业务逻辑、验证规则或持久化操作——这些职责属于 service 和 repo 层。
package model

// Spider 表示一个爬虫配置实体，包含运行时环境（语言、运行时类型、镜像、启动命令）信息。
// Command 字段在 JSON 序列化时若为空则省略（omitempty），Image 同理。
type Spider struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	Runtime   string `json:"runtime"`
	Image     string   `json:"image,omitempty"`
	Command   []string `json:"command,omitempty"`
}
