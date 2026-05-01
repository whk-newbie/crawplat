CREATE TABLE IF NOT EXISTS alert_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    rule_type TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    webhook_url TEXT NOT NULL,
    cooldown_seconds INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 5,
    offline_grace_seconds INTEGER NOT NULL DEFAULT 60,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alert_events (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    rule_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    dedupe_key TEXT NOT NULL,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    delivery_status TEXT NOT NULL,
    webhook_status_code INTEGER,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled_type
    ON alert_rules(enabled, rule_type);

CREATE INDEX IF NOT EXISTS idx_alert_events_rule_dedupe_created
    ON alert_events(rule_id, dedupe_key, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_alert_events_created_at
    ON alert_events(created_at DESC);
