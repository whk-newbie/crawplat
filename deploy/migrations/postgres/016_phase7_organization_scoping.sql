-- 为所有资源表添加 organization_id 外键列（当前允许 NULL，Phase 4 加固时设为 NOT NULL）
-- 对于已有 project_id 的子资源表，organization_id 是冗余但必要的反范式列
-- 它避免了跨表 JOIN 查询，确保每个服务层可以直接按 org 过滤

ALTER TABLE projects       ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id);
ALTER TABLE spiders        ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id);
ALTER TABLE datasources    ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id);
ALTER TABLE executions     ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id);
ALTER TABLE scheduled_tasks ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id);
ALTER TABLE spider_versions ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id);
ALTER TABLE nodes          ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id);
ALTER TABLE alert_rules    ADD COLUMN IF NOT EXISTS organization_id TEXT REFERENCES organizations(id);

-- 索引：按 organization_id 查询是每个服务最频繁的过滤条件
CREATE INDEX IF NOT EXISTS idx_projects_org_id        ON projects(organization_id);
CREATE INDEX IF NOT EXISTS idx_spiders_org_id         ON spiders(organization_id);
CREATE INDEX IF NOT EXISTS idx_datasources_org_id     ON datasources(organization_id);
CREATE INDEX IF NOT EXISTS idx_executions_org_id      ON executions(organization_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_org_id ON scheduled_tasks(organization_id);
CREATE INDEX IF NOT EXISTS idx_spider_versions_org_id ON spider_versions(organization_id);
CREATE INDEX IF NOT EXISTS idx_nodes_org_id           ON nodes(organization_id);
CREATE INDEX IF NOT EXISTS idx_alert_rules_org_id     ON alert_rules(organization_id);
