// Postgres 组织仓储实现，将组织数据持久化到 PostgreSQL organizations 和 organization_members 表。
package repo

import (
	"database/sql"
	"fmt"

	"crawler-platform/apps/iam-service/internal/model"
)

// PostgresOrgRepo 使用 PostgreSQL 存取组织与成员数据。
type PostgresOrgRepo struct {
	db *sql.DB
}

// NewPostgresOrgRepo 创建 Postgres 组织仓储。
func NewPostgresOrgRepo(db *sql.DB) *PostgresOrgRepo {
	return &PostgresOrgRepo{db: db}
}

// CreateOrganization 创建新组织并返回，同时将创建者（按 username 查找）添加为 admin。
func (r *PostgresOrgRepo) CreateOrganization(name, slug, createdByUsername string) (model.Organization, error) {
	var org model.Organization
	err := r.db.QueryRow(
		`INSERT INTO organizations (id, name, slug) VALUES (gen_random_uuid()::text, $1, $2)
		 RETURNING id, name, slug`,
		name, slug,
	).Scan(&org.ID, &org.Name, &org.Slug)
	if err != nil {
		return model.Organization{}, fmt.Errorf("create organization: %w", err)
	}

	_, err = r.db.Exec(
		`INSERT INTO organization_members (organization_id, user_id, role)
		 SELECT $1, id, 'admin' FROM users WHERE username = $2`,
		org.ID, createdByUsername,
	)
	if err != nil {
		return model.Organization{}, fmt.Errorf("add admin member: %w", err)
	}
	return org, nil
}

// FindMembershipsByUser 查询用户所属的全部组织及角色。
func (r *PostgresOrgRepo) FindMembershipsByUser(username string) ([]model.OrgMembership, error) {
	rows, err := r.db.Query(
		`SELECT om.organization_id, o.name, om.role
		 FROM organization_members om
		 JOIN users u ON u.id = om.user_id
		 JOIN organizations o ON o.id = om.organization_id
		 WHERE u.username = $1
		 ORDER BY o.created_at ASC`,
		username,
	)
	if err != nil {
		return nil, fmt.Errorf("find memberships: %w", err)
	}
	defer rows.Close()

	var memberships []model.OrgMembership
	for rows.Next() {
		var m model.OrgMembership
		if err := rows.Scan(&m.OrganizationID, &m.OrganizationName, &m.Role); err != nil {
			return nil, fmt.Errorf("scan membership: %w", err)
		}
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}
