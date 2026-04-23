package service

import (
	"context"
	"reflect"
	"testing"

	"crawler-platform/apps/scheduler-service/internal/model"
)

type fakeScheduleRepo struct {
	schedules []model.Schedule
}

func (r *fakeScheduleRepo) Create(_ context.Context, schedule model.Schedule) error {
	r.schedules = append(r.schedules, schedule)
	return nil
}

func (r *fakeScheduleRepo) List(_ context.Context) ([]model.Schedule, error) {
	schedules := make([]model.Schedule, len(r.schedules))
	copy(schedules, r.schedules)
	return schedules, nil
}

func (r *fakeScheduleRepo) mustGet(id string) model.Schedule {
	for _, schedule := range r.schedules {
		if schedule.ID == id {
			return schedule
		}
	}
	return model.Schedule{}
}

func TestSchedulerServiceCreatePersistsThroughRepo(t *testing.T) {
	repo := &fakeScheduleRepo{}
	svc := NewSchedulerService(repo)

	schedule, err := svc.Create("project-1", "spider-1", "nightly", "0 * * * *", "crawler/go-echo:latest", []string{"./go-echo"}, true)
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
}

func TestSchedulerServiceCreateRejectsMissingFields(t *testing.T) {
	svc := NewSchedulerService(&fakeScheduleRepo{})

	_, err := svc.Create("", "spider-1", "nightly", "0 * * * *", "crawler/go-echo:latest", nil, true)
	if err != ErrInvalidSchedule {
		t.Fatalf("expected ErrInvalidSchedule, got %v", err)
	}
}

func TestSchedulerServiceListReturnsRepoSchedules(t *testing.T) {
	repo := &fakeScheduleRepo{}
	svc := NewSchedulerService(repo)

	created, err := svc.Create("project-1", "spider-1", "nightly", "0 * * * *", "crawler/go-echo:latest", []string{"./go-echo"}, true)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	schedules, err := svc.List()
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
