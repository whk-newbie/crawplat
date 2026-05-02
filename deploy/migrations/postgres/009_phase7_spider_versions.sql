CREATE TABLE IF NOT EXISTS spider_versions (
    id TEXT PRIMARY KEY,
    spider_id TEXT NOT NULL REFERENCES spiders(id) ON DELETE CASCADE,
    version INT NOT NULL,
    image TEXT NOT NULL DEFAULT '',
    command JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (spider_id, version)
);

CREATE INDEX IF NOT EXISTS idx_spider_versions_spider_version
    ON spider_versions (spider_id, version DESC);

INSERT INTO spider_versions (id, spider_id, version, image, command, created_at)
SELECT s.id || '-v1', s.id, 1, s.image, s.command, s.created_at
FROM spiders s
WHERE NOT EXISTS (
    SELECT 1 FROM spider_versions v WHERE v.spider_id = s.id
);
