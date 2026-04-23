package model

type Schedule struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"projectId"`
	SpiderID  string   `json:"spiderId"`
	Name      string   `json:"name"`
	CronExpr  string   `json:"cronExpr"`
	Enabled   bool     `json:"enabled"`
	Image     string   `json:"image"`
	Command   []string `json:"command,omitempty"`
}
