ALTER TABLE executions
    ADD COLUMN IF NOT EXISTS registry_auth_ref TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_executions_registry_auth_ref
    ON executions (registry_auth_ref);
