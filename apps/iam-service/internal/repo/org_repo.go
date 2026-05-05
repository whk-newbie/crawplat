// Package repo 定义组织仓储接口和通用错误。
package repo

import "crawler-platform/apps/iam-service/internal/model"

// OrgRepository 定义组织数据访问接口，由 repo 层实现（内存 + Postgres）。
type OrgRepository interface {
	// CreateOrganization 创建新组织并返回，同时将创建者（按 username 查找）添加为 admin。
	CreateOrganization(name, slug, createdByUsername string) (model.Organization, error)

	// FindMembershipsByUser 查询用户所属的全部组织及其角色。
	FindMembershipsByUser(username string) ([]model.OrgMembership, error)
}
