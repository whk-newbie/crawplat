// Package model 的监控总览数据结构。
// Overview 聚合执行和节点两维度的统计数据，由 repo 层查询后组装。
package model

type ExecutionSummary struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

type NodeSummary struct {
	Total   int `json:"total"`
	Online  int `json:"online"`
	Offline int `json:"offline"`
}

type Overview struct {
	Executions ExecutionSummary `json:"executions"`
	Nodes      NodeSummary      `json:"nodes"`
}
