ALTER TABLE executions
    ADD COLUMN IF NOT EXISTS spider_version INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_executions_spider_version
    ON executions (spider_id, spider_version);
