package model

import "time"

type Schedule struct {
	ID                 string     `json:"id"`
	ProjectID          string     `json:"projectId"`
	SpiderID           string     `json:"spiderId"`
	SpiderVersion      int        `json:"spiderVersion,omitempty"`
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
