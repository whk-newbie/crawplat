ALTER TABLE scheduled_tasks
    ADD COLUMN IF NOT EXISTS registry_auth_ref TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_registry_auth_ref
    ON scheduled_tasks (registry_auth_ref);
