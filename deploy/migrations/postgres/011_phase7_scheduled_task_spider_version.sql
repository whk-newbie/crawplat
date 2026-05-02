ALTER TABLE scheduled_tasks
    ADD COLUMN IF NOT EXISTS spider_version INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_spider_version
    ON scheduled_tasks (spider_id, spider_version);
