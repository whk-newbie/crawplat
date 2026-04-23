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
