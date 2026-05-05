// Package model 定义 IAM 服务的数据模型。
// 本文件定义 Organization 和 OrgMembership 结构，用于多租户组织管理。
package model

// Organization 表示一个租户组织，是多租户隔离的顶层容器。
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// OrgMembership 表示用户在某个组织中的成员角色。
type OrgMembership struct {
	OrganizationID string `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	Role           string `json:"role"`
}
