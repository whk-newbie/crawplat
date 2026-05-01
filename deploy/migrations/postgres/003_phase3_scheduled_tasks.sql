CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    spider_id TEXT NOT NULL REFERENCES spiders(id),
    name TEXT NOT NULL,
    cron_expr TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    image TEXT NOT NULL,
    command JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_project_id ON scheduled_tasks (project_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_enabled ON scheduled_tasks (enabled, created_at DESC);
