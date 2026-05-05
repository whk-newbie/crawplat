// Package service 测试：验证 SchedulerService 的创建、列表和物化核心逻辑。
//
// 该文件负责：
//   - TestSchedulerServiceCreatePersistsThroughRepo：验证 Create 正确持久化并通过 Repo 读取。
//   - TestSchedulerServiceCreateRejectsMissingFields：验证参数校验（必填字段为空）。
//   - TestSchedulerServiceListReturnsRepoSchedules：验证 List 返回已创建的调度。
//   - TestMaterializeDueBackfillsScheduledExecutions：验证物化逻辑能根据 cron 生成正确的执行记录数，
//     并验证 RetryLimit/RetryDelaySeconds 透传、last_materialized_at 更新。
//   - TestMaterializeDueSkipsDisabledSchedules：验证禁用的调度不生成执行记录。
//   - TestMaterializeDueIsIdempotentPerTick：验证同一 tick 内重复调用不重复物化（去重）。
//
// fakeScheduleRepo 和 fakeExecutionClient 提供可控的测试替身，不依赖任何外部服务。
package service

import (
	"context"
	"reflect"
	"testing"
	"time"

	"crawler-platform/apps/scheduler-service/internal/model"
)

// fakeScheduleRepo 是 Repository 的测试替身，在内存中存储调度数据。
type fakeScheduleRepo struct {
	schedules []model.Schedule
}

func (r *fakeScheduleRepo) Create(_ context.Context, schedule model.Schedule) error {
	r.schedules = append(r.schedules, schedule)
	return nil
}

func (r *fakeScheduleRepo) List(_ context.Context, limit, offset int) ([]model.Schedule, error) {
	schedules := make([]model.Schedule, len(r.schedules))
	copy(schedules, r.schedules)
	return schedules, nil
}

func (r *fakeScheduleRepo) AdvanceLastMaterialized(_ context.Context, id string, previous *time.Time, next time.Time) (bool, error) {
	for i, schedule := range r.schedules {
		if schedule.ID != id {
			continue
		}
		if !sameTimePtrs(schedule.LastMaterializedAt, previous) {
			return false, nil
		}
		r.schedules[i].LastMaterializedAt = &next
		return true, nil
	}
	return false, nil
}

func (r *fakeScheduleRepo) RestoreLastMaterialized(_ context.Context, id string, previous *time.Time, current time.Time) error {
	for i, schedule := range r.schedules {
		if schedule.ID != id {
			continue
		}
		if schedule.LastMaterializedAt == nil || !schedule.LastMaterializedAt.Equal(current) {
			return nil
		}
		r.schedules[i].LastMaterializedAt = previous
		return nil
	}
	return nil
}

func (r *fakeScheduleRepo) mustGet(id string) model.Schedule {
	for _, schedule := range r.schedules {
		if schedule.ID == id {
			return schedule
		}
	}
	return model.Schedule{}
}

// fakeExecutionClient 是 ExecutionClient 的测试替身，记录所有 Create 请求并返回假执行 ID。
// err 字段可注入错误以模拟执行服务故障。
type fakeExecutionClient struct {
	requests []CreateExecutionInput
	err      error
}

func (c *fakeExecutionClient) Create(_ context.Context, input CreateExecutionInput) (string, error) {
	if c.err != nil {
		return "", c.err
	}
	c.requests = append(c.requests, input)
	return "exec-" + input.ScheduleID, nil
}

func (c *fakeExecutionClient) MaterializeRetry(_ context.Context) (bool, error) {
	return false, c.err
}

// TestSchedulerServiceCreatePersistsThroughRepo 验证 Create 写出的数据能通过 Repo 完整读取。
func TestSchedulerServiceCreatePersistsThroughRepo(t *testing.T) {
	repo := &fakeScheduleRepo{}
	svc := NewSchedulerService(repo, nil)

	schedule, err := svc.Create("project-1", "spider-1", "", "", "nightly", "0 * * * *", "crawler/go-echo:latest", []string{"./go-echo"}, true, 0, 0)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if schedule.ID == "" {
		t.Fatal("expected generated id")
	}

	got := repo.mustGet(schedule.ID)
	if got.Name != "nightly" || got.CronExpr != "0 * * * *" || got.Image != "crawler/go-echo:latest" {
		t.Fatalf("unexpected persisted schedule: %#v", got)
	}
	if got.CreatedAt.IsZero() {
		t.Fatalf("expected createdAt to be set, got %#v", got)
	}
	if got.RetryLimit != 0 || got.RetryDelaySeconds != 0 {
		t.Fatalf("expected retry defaults to be zeroed, got %#v", got)
	}
}

// TestSchedulerServiceCreateRejectsMissingFields 验证缺少必填字段时返回 ErrInvalidSchedule。
func TestSchedulerServiceCreateRejectsMissingFields(t *testing.T) {
	svc := NewSchedulerService(&fakeScheduleRepo{}, nil)

	_, err := svc.Create("", "spider-1", "", "", "nightly", "0 * * * *", "crawler/go-echo:latest", nil, true, 0, 0)
	if err != ErrInvalidSchedule {
		t.Fatalf("expected ErrInvalidSchedule, got %v", err)
	}
}

// TestSchedulerServiceListReturnsRepoSchedules 验证 List 返回与 Create 输出一致的数据。
func TestSchedulerServiceListReturnsRepoSchedules(t *testing.T) {
	repo := &fakeScheduleRepo{}
	svc := NewSchedulerService(repo, nil)

	created, err := svc.Create("project-1", "spider-1", "", "", "nightly", "0 * * * *", "crawler/go-echo:latest", []string{"./go-echo"}, true, 0, 0)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	schedules, err := svc.List(20, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedules))
	}
	if !reflect.DeepEqual(schedules[0], created) {
		t.Fatalf("expected list to return created schedule, got %+v want %+v", schedules[0], created)
	}
}

// TestMaterializeDueBackfillsScheduledExecutions 验证从未物化的调度能正确生成追赶执行记录。
// 调度创建于 23:25，cron 为 "*/5 * * * *"，当前时间 23:45，应生成 5 条执行（23:25 到 23:45 每隔 5 分钟）。
// 同时验证 RetryLimit/RetryDelaySeconds 和 TriggerSource="scheduled" 透传到执行请求。
func TestMaterializeDueBackfillsScheduledExecutions(t *testing.T) {
	now := time.Date(2026, 4, 23, 23, 45, 0, 0, time.UTC)
	repo := &fakeScheduleRepo{
		schedules: []model.Schedule{{
			ID:        "sched-1",
			ProjectID: "project-1",
			SpiderID:  "spider-1",
			Name:      "nightly",
			CronExpr:  "*/5 * * * *",
			Enabled:   true,
			Image:     "crawler/go-echo:latest",
			Command:   []string{"./go-echo"},
			RetryLimit: 2,
			RetryDelaySeconds: 30,
			CreatedAt: now.Add(-20 * time.Minute),
		}},
	}
	client := &fakeExecutionClient{}
	svc := NewSchedulerService(repo, client, WithNow(func() time.Time { return now }))

	materialized, err := svc.MaterializeDue(context.Background())
	if err != nil {
		t.Fatalf("MaterializeDue returned error: %v", err)
	}
	if materialized != 5 {
		t.Fatalf("expected 5 materialized executions, got %d", materialized)
	}
	if len(client.requests) != 5 {
		t.Fatalf("expected 5 execution requests, got %d", len(client.requests))
	}
	if client.requests[0].TriggerSource != "scheduled" {
		t.Fatalf("expected scheduled trigger source, got %+v", client.requests[0])
	}
	if client.requests[0].RetryLimit != 2 || client.requests[0].RetryDelaySeconds != 30 || client.requests[0].RetryCount != 0 {
		t.Fatalf("expected retry config to flow into execution request, got %+v", client.requests[0])
	}
	got := repo.mustGet("sched-1")
	if got.LastMaterializedAt == nil || !got.LastMaterializedAt.Equal(now) {
		t.Fatalf("expected lastMaterializedAt=%s, got %#v", now, got.LastMaterializedAt)
	}
}

// TestMaterializeDueSkipsDisabledSchedules 验证 Enabled=false 的调度不生成任何执行记录。
func TestMaterializeDueSkipsDisabledSchedules(t *testing.T) {
	now := time.Date(2026, 4, 23, 23, 45, 0, 0, time.UTC)
	repo := &fakeScheduleRepo{
		schedules: []model.Schedule{{
			ID:        "sched-1",
			ProjectID: "project-1",
			SpiderID:  "spider-1",
			Name:      "nightly",
			CronExpr:  "*/5 * * * *",
			Enabled:   false,
			Image:     "crawler/go-echo:latest",
			Command:   []string{"./go-echo"},
			CreatedAt: now.Add(-20 * time.Minute),
		}},
	}
	client := &fakeExecutionClient{}
	svc := NewSchedulerService(repo, client, WithNow(func() time.Time { return now }))

	materialized, err := svc.MaterializeDue(context.Background())
	if err != nil {
		t.Fatalf("MaterializeDue returned error: %v", err)
	}
	if materialized != 0 {
		t.Fatalf("expected 0 materialized executions, got %d", materialized)
	}
	if len(client.requests) != 0 {
		t.Fatalf("expected 0 execution requests, got %d", len(client.requests))
	}
}

// TestMaterializeDueIsIdempotentPerTick 验证同一 tick 内两次调用 MaterializeDue 不会重复物化。
// 第一次物化 5 条，第二次应为 0（因为 last_materialized_at 已推进到 now）。
func TestMaterializeDueIsIdempotentPerTick(t *testing.T) {
	now := time.Date(2026, 4, 23, 23, 45, 0, 0, time.UTC)
	repo := &fakeScheduleRepo{
		schedules: []model.Schedule{{
			ID:        "sched-1",
			ProjectID: "project-1",
			SpiderID:  "spider-1",
			Name:      "nightly",
			CronExpr:  "*/5 * * * *",
			Enabled:   true,
			Image:     "crawler/go-echo:latest",
			Command:   []string{"./go-echo"},
			CreatedAt: now.Add(-20 * time.Minute),
		}},
	}
	client := &fakeExecutionClient{}
	svc := NewSchedulerService(repo, client, WithNow(func() time.Time { return now }))

	first, err := svc.MaterializeDue(context.Background())
	if err != nil {
		t.Fatalf("first MaterializeDue returned error: %v", err)
	}
	second, err := svc.MaterializeDue(context.Background())
	if err != nil {
		t.Fatalf("second MaterializeDue returned error: %v", err)
	}
	if first != 5 || second != 0 {
		t.Fatalf("expected materialization counts 5 then 0, got %d then %d", first, second)
	}
	if len(client.requests) != 5 {
		t.Fatalf("expected 5 execution requests across both runs, got %d", len(client.requests))
	}
}

func sameTimePtrs(left, right *time.Time) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return left.Equal(*right)
	}
}
