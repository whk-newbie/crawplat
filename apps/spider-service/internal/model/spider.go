package model

import "time"

type Spider struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"projectId"`
	Name      string   `json:"name"`
	Language  string   `json:"language"`
	Runtime   string   `json:"runtime"`
	Image     string   `json:"image,omitempty"`
	Command   []string `json:"command,omitempty"`
}

type SpiderVersion struct {
	ID              string    `json:"id"`
	SpiderID        string    `json:"spiderId"`
	Version         int       `json:"version"`
	RegistryAuthRef string    `json:"registryAuthRef,omitempty"`
	Image           string    `json:"image"`
	Command         []string  `json:"command"`
	CreatedAt       time.Time `json:"createdAt"`
}
