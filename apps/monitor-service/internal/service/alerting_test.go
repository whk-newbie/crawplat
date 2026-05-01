package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"crawler-platform/apps/monitor-service/internal/model"
)

type fakeAlertRepo struct {
	rules              []model.AlertRule
	failedExecutions   []model.FailedExecutionCandidate
	offlineNodes       []model.OfflineNodeCandidate
	lastEventByDedupe  map[string]time.Time
	savedEvents        []model.AlertEventRecord
	createdRules       []model.AlertRule
	createRuleErr      error
	listAlertRulesErr  error
	listFailedExecErr  error
	listOfflineNodeErr error
	saveEventErr       error
}

func (r *fakeAlertRepo) Overview(_ context.Context) (model.Overview, error) {
	return model.Overview{}, nil
}
func (r *fakeAlertRepo) CreateAlertRule(_ context.Context, rule model.AlertRule) (model.AlertRule, error) {
	if r.createRuleErr != nil {
		return model.AlertRule{}, r.createRuleErr
	}
	r.createdRules = append(r.createdRules, rule)
	return rule, nil
}
func (r *fakeAlertRepo) ListAlertRules(_ context.Context) ([]model.AlertRule, error) {
	if r.listAlertRulesErr != nil {
		return nil, r.listAlertRulesErr
	}
	return append([]model.AlertRule(nil), r.rules...), nil
}
func (r *fakeAlertRepo) ListAlertEvents(_ context.Context, _, _ int) ([]model.AlertEvent, error) {
	return nil, nil
}
func (r *fakeAlertRepo) CountAlertEvents(_ context.Context) (int64, error) { return 0, nil }
func (r *fakeAlertRepo) ListFailedExecutionsSince(_ context.Context, _ time.Time, _ int) ([]model.FailedExecutionCandidate, error) {
	if r.listFailedExecErr != nil {
		return nil, r.listFailedExecErr
	}
	return append([]model.FailedExecutionCandidate(nil), r.failedExecutions...), nil
}
func (r *fakeAlertRepo) ListOfflineNodes(_ context.Context, _ time.Time, _ int) ([]model.OfflineNodeCandidate, error) {
	if r.listOfflineNodeErr != nil {
		return nil, r.listOfflineNodeErr
	}
	return append([]model.OfflineNodeCandidate(nil), r.offlineNodes...), nil
}
func (r *fakeAlertRepo) LastAlertEventAt(_ context.Context, ruleID, dedupeKey string) (*time.Time, error) {
	if r.lastEventByDedupe == nil {
		return nil, nil
	}
	key := ruleID + ":" + dedupeKey
	if t, ok := r.lastEventByDedupe[key]; ok {
		copied := t
		return &copied, nil
	}
	return nil, nil
}
func (r *fakeAlertRepo) SaveAlertEvent(_ context.Context, event model.AlertEventRecord) error {
	if r.saveEventErr != nil {
		return r.saveEventErr
	}
	r.savedEvents = append(r.savedEvents, event)
	return nil
}

type fakeDeliverer struct {
	statusCode int
	err        error
}

func (d *fakeDeliverer) Deliver(_ context.Context, _ string, _ []byte, _ time.Duration) (int, error) {
	if d.err != nil {
		return d.statusCode, d.err
	}
	if d.statusCode == 0 {
		return 200, nil
	}
	return d.statusCode, nil
}

func TestCreateAlertRuleDefaultsForNodeOffline(t *testing.T) {
	repo := &fakeAlertRepo{}
	svc := NewMonitorService(repo)

	rule, err := svc.CreateAlertRule(CreateAlertRuleInput{
		Name:       "node offline",
		RuleType:   model.AlertRuleTypeNodeOffline,
		Enabled:    true,
		WebhookURL: "https://example.com/webhook",
	})
	if err != nil {
		t.Fatalf("CreateAlertRule returned error: %v", err)
	}
	if rule.TimeoutSeconds != 3 {
		t.Fatalf("expected node offline default timeout 3s, got %d", rule.TimeoutSeconds)
	}
	if rule.CooldownSeconds != 60 {
		t.Fatalf("expected node offline default cooldown 60s, got %d", rule.CooldownSeconds)
	}
	if rule.OfflineGraceSeconds != 60 {
		t.Fatalf("expected default offline grace 60s, got %d", rule.OfflineGraceSeconds)
	}
}

func TestEvaluateAlertsSavesSentEventForFailedExecution(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeAlertRepo{
		rules: []model.AlertRule{{
			ID:                  "rule-1",
			Name:                "failed exec",
			RuleType:            model.AlertRuleTypeExecutionFailed,
			Enabled:             true,
			WebhookURL:          "https://example.com/webhook",
			CooldownSeconds:     120,
			TimeoutSeconds:      5,
			OfflineGraceSeconds: 60,
			CreatedAt:           now.Add(-time.Hour),
			UpdatedAt:           now.Add(-time.Hour),
		}},
		failedExecutions: []model.FailedExecutionCandidate{{
			ExecutionID: "exec-1",
			ProjectID:   "p1",
			SpiderID:    "s1",
			Error:       "exit status 1",
			OccurredAt:  now.Add(-time.Minute),
		}},
	}
	svc := NewMonitorService(repo).WithWebhookDeliverer(&fakeDeliverer{statusCode: 200})

	if err := svc.EvaluateAlerts(context.Background()); err != nil {
		t.Fatalf("EvaluateAlerts returned error: %v", err)
	}
	if len(repo.savedEvents) != 1 {
		t.Fatalf("expected 1 saved alert event, got %d", len(repo.savedEvents))
	}
	if repo.savedEvents[0].DeliveryStatus != "sent" || repo.savedEvents[0].EntityID != "exec-1" {
		t.Fatalf("unexpected saved event: %+v", repo.savedEvents[0])
	}
}

func TestEvaluateAlertsSkipsExecutionWithinCooldown(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeAlertRepo{
		rules: []model.AlertRule{{
			ID:                  "rule-1",
			Name:                "failed exec",
			RuleType:            model.AlertRuleTypeExecutionFailed,
			Enabled:             true,
			WebhookURL:          "https://example.com/webhook",
			CooldownSeconds:     120,
			TimeoutSeconds:      5,
			OfflineGraceSeconds: 60,
			CreatedAt:           now.Add(-time.Hour),
			UpdatedAt:           now.Add(-time.Hour),
		}},
		failedExecutions: []model.FailedExecutionCandidate{{
			ExecutionID: "exec-1",
			ProjectID:   "p1",
			SpiderID:    "s1",
			Error:       "exit status 1",
			OccurredAt:  now.Add(-time.Minute),
		}},
		lastEventByDedupe: map[string]time.Time{
			"rule-1:execution:exec-1": now,
		},
	}
	svc := NewMonitorService(repo).WithWebhookDeliverer(&fakeDeliverer{statusCode: 200})

	if err := svc.EvaluateAlerts(context.Background()); err != nil {
		t.Fatalf("EvaluateAlerts returned error: %v", err)
	}
	if len(repo.savedEvents) != 0 {
		t.Fatalf("expected 0 saved alert events, got %d", len(repo.savedEvents))
	}
}

func TestCreateAlertRuleRejectsInvalidInput(t *testing.T) {
	repo := &fakeAlertRepo{}
	svc := NewMonitorService(repo)

	if _, err := svc.CreateAlertRule(CreateAlertRuleInput{}); !errors.Is(err, ErrInvalidRuleName) {
		t.Fatalf("expected ErrInvalidRuleName, got %v", err)
	}
}
