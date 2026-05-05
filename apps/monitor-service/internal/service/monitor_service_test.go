package service

import (
	"context"
	"errors"
	"testing"

	"crawler-platform/apps/monitor-service/internal/model"
)

type fakeSummaryRepository struct {
	overview model.Overview
	err      error
}

func (r *fakeSummaryRepository) Overview(_ context.Context) (model.Overview, error) {
	if r.err != nil {
		return model.Overview{}, r.err
	}
	return r.overview, nil
}

func TestMonitorServiceOverviewReturnsRepositorySummary(t *testing.T) {
	repo := &fakeSummaryRepository{
		overview: model.Overview{
			Executions: model.ExecutionSummary{
				Total:     12,
				Pending:   3,
				Running:   2,
				Succeeded: 6,
				Failed:    1,
			},
			Nodes: model.NodeSummary{
				Total:   4,
				Online:  3,
				Offline: 1,
			},
		},
	}
	svc := NewMonitorService(repo)

	overview, err := svc.Overview()
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}
	if overview != repo.overview {
		t.Fatalf("expected overview %+v, got %+v", repo.overview, overview)
	}
}

func TestMonitorServiceOverviewUsesMemoryFallback(t *testing.T) {
	svc := NewMonitorService()

	overview, err := svc.Overview()
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}
	if overview.Executions.Total != 0 || overview.Executions.Pending != 0 || overview.Executions.Running != 0 || overview.Executions.Succeeded != 0 || overview.Executions.Failed != 0 {
		t.Fatalf("expected zero execution counts, got %+v", overview.Executions)
	}
	if overview.Nodes.Total != 0 || overview.Nodes.Online != 0 || overview.Nodes.Offline != 0 {
		t.Fatalf("expected zero node counts, got %+v", overview.Nodes)
	}
}

func TestMonitorServiceOverviewReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("summary unavailable")
	svc := NewMonitorService(&fakeSummaryRepository{err: expectedErr})

	_, err := svc.Overview()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
