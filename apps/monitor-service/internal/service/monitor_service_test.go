package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"crawler-platform/apps/monitor-service/internal/model"
)

type fakeSummaryRepository struct {
	overview model.Overview
	err      error
	rules    map[string]model.AlertRule
}

func (r *fakeSummaryRepository) Overview(_ context.Context) (model.Overview, error) {
	if r.err != nil {
		return model.Overview{}, r.err
	}
	return r.overview, nil
}

func (r *fakeSummaryRepository) CreateAlertRule(_ context.Context, rule model.AlertRule) (model.AlertRule, error) {
	if r.rules == nil {
		r.rules = map[string]model.AlertRule{}
	}
	r.rules[rule.ID] = rule
	return rule, nil
}
func (r *fakeSummaryRepository) UpdateAlertRule(_ context.Context, id string, patch model.AlertRulePatch) (model.AlertRule, bool, error) {
	if r.rules == nil {
		return model.AlertRule{}, false, nil
	}
	rule, ok := r.rules[id]
	if !ok {
		return model.AlertRule{}, false, nil
	}
	if patch.Name != nil {
		rule.Name = *patch.Name
	}
	if patch.Enabled != nil {
		rule.Enabled = *patch.Enabled
	}
	if patch.WebhookURL != nil {
		rule.WebhookURL = *patch.WebhookURL
	}
	if patch.CooldownSeconds != nil {
		rule.CooldownSeconds = *patch.CooldownSeconds
	}
	if patch.TimeoutSeconds != nil {
		rule.TimeoutSeconds = *patch.TimeoutSeconds
	}
	if patch.OfflineGraceSeconds != nil {
		rule.OfflineGraceSeconds = *patch.OfflineGraceSeconds
	}
	rule.UpdatedAt = patch.UpdatedAt
	r.rules[id] = rule
	return rule, true, nil
}
func (r *fakeSummaryRepository) ListAlertRules(_ context.Context) ([]model.AlertRule, error) {
	return nil, nil
}
func (r *fakeSummaryRepository) ListAlertEvents(_ context.Context, _ int, _ int) ([]model.AlertEvent, error) {
	return nil, nil
}
func (r *fakeSummaryRepository) CountAlertEvents(_ context.Context) (int64, error) {
	return 0, nil
}
func (r *fakeSummaryRepository) ListFailedExecutionsSince(_ context.Context, _ time.Time, _ int) ([]model.FailedExecutionCandidate, error) {
	return nil, nil
}
func (r *fakeSummaryRepository) ListOfflineNodes(_ context.Context, _ time.Time, _ int) ([]model.OfflineNodeCandidate, error) {
	return nil, nil
}
func (r *fakeSummaryRepository) LastAlertEventAt(_ context.Context, _, _ string) (*time.Time, error) {
	return nil, nil
}
func (r *fakeSummaryRepository) SaveAlertEvent(_ context.Context, _ model.AlertEventRecord) error {
	return nil
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

func TestUpdateAlertRuleUpdatesEnabled(t *testing.T) {
	repo := &fakeSummaryRepository{
		rules: map[string]model.AlertRule{
			"rule-1": {
				ID:       "rule-1",
				Name:     "node offline",
				RuleType: model.AlertRuleTypeNodeOffline,
				Enabled:  true,
			},
		},
	}
	svc := NewMonitorService(repo)
	enabled := false
	got, err := svc.UpdateAlertRule("rule-1", UpdateAlertRuleInput{Enabled: &enabled})
	if err != nil {
		t.Fatalf("UpdateAlertRule returned error: %v", err)
	}
	if got.Enabled {
		t.Fatalf("expected enabled false, got %+v", got)
	}
}
