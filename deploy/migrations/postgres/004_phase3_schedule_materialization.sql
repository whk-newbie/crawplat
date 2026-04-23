ALTER TABLE scheduled_tasks
ADD COLUMN IF NOT EXISTS last_materialized_at TIMESTAMPTZ NULL;
