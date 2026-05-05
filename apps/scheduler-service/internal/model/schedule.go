// Package model 定义调度服务（scheduler-service）的领域模型。
//
// 该文件负责：
//   - 定义 Schedule 结构体，包含定时任务的所有配置字段。
//   - LastMaterializedAt 字段作为物化游标，是防止重复生成执行记录的关键。
//
// 不负责：
//   - 不包含数据库操作逻辑（由 repo 层负责）。
//   - 不包含业务逻辑（由 service 层负责）。
package model

import "time"

// Schedule 表示一条定时爬虫调度规则。
//
// 字段说明：
//   - LastMaterializedAt: 上一次物化的时间点。作为游标用于去重和追赶控制：
//     每次物化时 AdvanceLastMaterialized 会 CAS 更新该字段，确保同一时间点不会被重复物化。
//     如果该字段为 nil，表示从未物化过，物化起点为 CreatedAt - 1 minute。
type Schedule struct {
	ID                 string     `json:"id"`
	ProjectID          string     `json:"projectId"`
	SpiderID           string     `json:"spiderId"`
	SpiderVersion      string     `json:"spiderVersion,omitempty"`
	RegistryAuthRef    string     `json:"registryAuthRef,omitempty"`
	Name               string     `json:"name"`
	CronExpr           string     `json:"cronExpr"`
	Enabled            bool       `json:"enabled"`
	Image              string     `json:"image"`
	Command            []string   `json:"command,omitempty"`
	RetryLimit         int        `json:"retryLimit"`
	RetryDelaySeconds  int        `json:"retryDelaySeconds"`
	CreatedAt          time.Time  `json:"createdAt"`
	LastMaterializedAt *time.Time `json:"lastMaterializedAt,omitempty"`
}
