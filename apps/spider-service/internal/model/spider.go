package model

type Spider struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	Runtime   string `json:"runtime"`
}
