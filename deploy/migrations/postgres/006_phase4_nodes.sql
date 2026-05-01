CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    capabilities_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    last_seen_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS node_heartbeats (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    seen_at TIMESTAMPTZ NOT NULL,
    capabilities_json JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_node_heartbeats_node_id_seen_at
    ON node_heartbeats(node_id, seen_at DESC);
