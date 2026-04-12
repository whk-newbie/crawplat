CREATE INDEX IF NOT EXISTS idx_spiders_project_id ON spiders(project_id);
CREATE INDEX IF NOT EXISTS idx_datasources_project_id ON datasources(project_id);
CREATE INDEX IF NOT EXISTS idx_executions_project_id ON executions(project_id);
CREATE INDEX IF NOT EXISTS idx_executions_status_created_at ON executions(status, created_at DESC);
