ALTER TABLE spider_versions
    ADD COLUMN IF NOT EXISTS registry_auth_ref TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_spider_versions_registry_auth_ref
    ON spider_versions (registry_auth_ref);
