package model

import "time"

type Execution struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"projectId"`
	SpiderID      string         `json:"spiderId"`
	NodeID        string         `json:"nodeId,omitempty"`
	Status        string         `json:"status"`
	TriggerSource string         `json:"triggerSource"`
	Image         string         `json:"image"`
	Command       []string       `json:"command"`
	ErrorMessage  string         `json:"errorMessage,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	StartedAt     *time.Time     `json:"startedAt,omitempty"`
	FinishedAt    *time.Time     `json:"finishedAt,omitempty"`
	Logs          []ExecutionLog `json:"logs,omitempty"`
}

type ExecutionLog struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"executionId"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"createdAt"`
}
