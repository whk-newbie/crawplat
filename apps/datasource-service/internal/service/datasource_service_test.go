// Package service 的单元测试，验证 DatasourceService 的业务逻辑。
// 使用 fakeDatasourceRepo（内存实现）替代真实仓储，不依赖 PostgreSQL 或外部数据源。
// 测试覆盖：类型校验、仓储持久化、列表查询、连接测试和数据预览。
package service

import (
	"context"
	"testing"

	"crawler-platform/apps/datasource-service/internal/model"
)

type fakeDatasourceRepo struct {
	datasources []model.Datasource
}

func (r *fakeDatasourceRepo) Create(_ context.Context, datasource model.Datasource) error {
	r.datasources = append(r.datasources, datasource)
	return nil
}

func (r *fakeDatasourceRepo) ListByProject(_ context.Context, projectID string) ([]model.Datasource, error) {
	var datasources []model.Datasource
	for _, datasource := range r.datasources {
		if projectID == "" || datasource.ProjectID == projectID {
			datasources = append(datasources, datasource)
		}
	}
	return datasources, nil
}

func (r *fakeDatasourceRepo) Get(_ context.Context, id string) (model.Datasource, bool, error) {
	for _, datasource := range r.datasources {
		if datasource.ID == id {
			return datasource, true, nil
		}
	}
	return model.Datasource{}, false, nil
}

// TestCreateDatasourceRejectsUnknownType 验证创建数据源时对不支持类型（如 mysql）的校验拒绝。
func TestCreateDatasourceRejectsUnknownType(t *testing.T) {
	svc := NewDatasourceService(&fakeDatasourceRepo{})
	_, err := svc.Create("project-1", "main", "mysql", nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// TestCreateDatasourcePersistsThroughRepo 验证 Create 方法正确委托给 Repository 持久化，
// 包括 UUID 自动生成和 Config map 数据保存。
func TestCreateDatasourcePersistsThroughRepo(t *testing.T) {
	repo := &fakeDatasourceRepo{}
	svc := NewDatasourceService(repo)

	datasource, err := svc.Create("project-1", "main", "postgresql", map[string]string{"schema": "public"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if datasource.ID == "" {
		t.Fatal("expected generated id")
	}
	if len(repo.datasources) != 1 {
		t.Fatalf("expected repo to persist one datasource, got %d", len(repo.datasources))
	}
	if repo.datasources[0].Config["schema"] != "public" {
		t.Fatalf("expected config to persist, got %#v", repo.datasources[0].Config)
	}
}

// TestDatasourceServiceListAndReadUseRepo 验证 List、Test、Preview 方法正确委托给 Repository，
// 覆盖完整的创建→查询→测试→预览生命周期。
func TestDatasourceServiceListAndReadUseRepo(t *testing.T) {
	repo := &fakeDatasourceRepo{}
	svc := NewDatasourceService(repo)

	created, err := svc.Create("project-1", "main", "redis", map[string]string{"db": "0"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	listed, err := svc.List("project-1")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("unexpected list length: %#v", listed)
	}
	if listed[0].ID != created.ID || listed[0].Config["db"] != "0" {
		t.Fatalf("unexpected list result: %#v", listed[0])
	}

	result, err := svc.Test(created.ID)
	if err != nil {
		t.Fatalf("Test returned error: %v", err)
	}
	if result.DatasourceID != created.ID {
		t.Fatalf("unexpected test result: %#v", result)
	}

	preview, err := svc.Preview(created.ID)
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	if preview.DatasourceType != "redis" {
		t.Fatalf("unexpected preview result: %#v", preview)
	}
}
