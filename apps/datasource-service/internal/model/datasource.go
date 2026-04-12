package model

type Datasource struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Readonly  bool   `json:"readonly"`
}

type TestResult struct {
	DatasourceID string `json:"datasourceId"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type PreviewResult struct {
	DatasourceID   string              `json:"datasourceId"`
	DatasourceType string              `json:"datasourceType"`
	Rows           []map[string]string `json:"rows"`
}
