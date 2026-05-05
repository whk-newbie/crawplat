// 内存组织仓储实现，仅用于开发和测试。
package repo

import (
	"fmt"
	"sync"

	"crawler-platform/apps/iam-service/internal/model"
)

// MemoryOrgRepo 使用内存 map 存储组织与成员数据，非并发安全。
type MemoryOrgRepo struct {
	mu      sync.Mutex
	orgs    map[string]model.Organization
	members map[string][]model.OrgMembership // key: username
	nextID  int
}

// NewMemoryOrgRepo 创建内存组织仓储。
func NewMemoryOrgRepo() *MemoryOrgRepo {
	return &MemoryOrgRepo{
		orgs:    make(map[string]model.Organization),
		members: make(map[string][]model.OrgMembership),
		nextID:  1,
	}
}

// CreateOrganization 创建新组织并返回，同时将创建者添加为 admin。
func (r *MemoryOrgRepo) CreateOrganization(name, slug, createdByUsername string) (model.Organization, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("mem-org-%d", r.nextID)
	r.nextID++
	org := model.Organization{ID: id, Name: name, Slug: slug}
	r.orgs[id] = org
	r.members[createdByUsername] = append(r.members[createdByUsername], model.OrgMembership{
		OrganizationID:   id,
		OrganizationName: name,
		Role:             "admin",
	})
	return org, nil
}

// FindMembershipsByUser 查询用户所属的全部组织及角色。
func (r *MemoryOrgRepo) FindMembershipsByUser(username string) ([]model.OrgMembership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[username]
	if !ok {
		return nil, nil
	}
	return m, nil
}
