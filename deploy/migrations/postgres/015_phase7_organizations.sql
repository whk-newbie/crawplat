-- 多租户基础表：组织与成员关系
-- organizations 是租户隔离的顶层容器，每个组织拥有独立的项目、爬虫、执行等资源

CREATE TABLE IF NOT EXISTS organizations (
    id         TEXT        PRIMARY KEY,
    name       TEXT        NOT NULL,
    slug       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- organization_members 记录用户与组织的成员关系及角色
-- role 取值：'admin'（管理员）、'member'（普通成员）
CREATE TABLE IF NOT EXISTS organization_members (
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            TEXT NOT NULL DEFAULT 'member',
    PRIMARY KEY (organization_id, user_id)
);
