package model

import "time"

type Execution struct {
	ID              string         `json:"id"`
	TaskID          string         `json:"taskId"`
	SpiderVersionID string         `json:"spiderVersionId"`
	Status          string         `json:"status"`
	TriggerSource   string         `json:"triggerSource"`
	CreatedAt       time.Time      `json:"createdAt"`
	Logs            []ExecutionLog `json:"logs,omitempty"`
}

type ExecutionLog struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"executionId"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"createdAt"`
}
