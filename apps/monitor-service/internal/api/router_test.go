package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"crawler-platform/apps/monitor-service/internal/model"
	"crawler-platform/apps/monitor-service/internal/service"
	"github.com/gin-gonic/gin"
)

type fakeSummaryRepository struct {
	overview      model.Overview
	err           error
	rules         []model.AlertRule
	events        []model.AlertEvent
	totalEvents   int64
	createRuleErr error
}

func (r *fakeSummaryRepository) Overview(_ context.Context) (model.Overview, error) {
	if r.err != nil {
		return model.Overview{}, r.err
	}
	return r.overview, nil
}

func (r *fakeSummaryRepository) CreateAlertRule(_ context.Context, rule model.AlertRule) (model.AlertRule, error) {
	if r.createRuleErr != nil {
		return model.AlertRule{}, r.createRuleErr
	}
	r.rules = append(r.rules, rule)
	return rule, nil
}
func (r *fakeSummaryRepository) UpdateAlertRule(_ context.Context, id string, patch model.AlertRulePatch) (model.AlertRule, bool, error) {
	for i := range r.rules {
		if r.rules[i].ID != id {
			continue
		}
		if patch.Name != nil {
			r.rules[i].Name = *patch.Name
		}
		if patch.Enabled != nil {
			r.rules[i].Enabled = *patch.Enabled
		}
		if patch.WebhookURL != nil {
			r.rules[i].WebhookURL = *patch.WebhookURL
		}
		if patch.CooldownSeconds != nil {
			r.rules[i].CooldownSeconds = *patch.CooldownSeconds
		}
		if patch.TimeoutSeconds != nil {
			r.rules[i].TimeoutSeconds = *patch.TimeoutSeconds
		}
		if patch.OfflineGraceSeconds != nil {
			r.rules[i].OfflineGraceSeconds = *patch.OfflineGraceSeconds
		}
		r.rules[i].UpdatedAt = patch.UpdatedAt
		return r.rules[i], true, nil
	}
	return model.AlertRule{}, false, nil
}
func (r *fakeSummaryRepository) ListAlertRules(_ context.Context) ([]model.AlertRule, error) {
	return append([]model.AlertRule(nil), r.rules...), nil
}
func (r *fakeSummaryRepository) ListAlertEvents(_ context.Context, limit, offset int) ([]model.AlertEvent, error) {
	if offset >= len(r.events) {
		return nil, nil
	}
	end := offset + limit
	if end > len(r.events) {
		end = len(r.events)
	}
	return append([]model.AlertEvent(nil), r.events[offset:end]...), nil
}
func (r *fakeSummaryRepository) CountAlertEvents(_ context.Context) (int64, error) {
	if r.totalEvents > 0 {
		return r.totalEvents, nil
	}
	return int64(len(r.events)), nil
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

func TestOverviewRouteReturnsSummaryJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewMonitorService(&fakeSummaryRepository{
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
	})
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitor/overview", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var overview model.Overview
	if err := json.Unmarshal(w.Body.Bytes(), &overview); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if overview.Executions.Total != 12 || overview.Executions.Pending != 3 || overview.Executions.Running != 2 || overview.Executions.Succeeded != 6 || overview.Executions.Failed != 1 {
		t.Fatalf("unexpected execution summary: %+v", overview.Executions)
	}
	if overview.Nodes.Total != 4 || overview.Nodes.Online != 3 || overview.Nodes.Offline != 1 {
		t.Fatalf("unexpected node summary: %+v", overview.Nodes)
	}
	if !strings.Contains(w.Body.String(), `"executions":`) || !strings.Contains(w.Body.String(), `"nodes":`) || !strings.Contains(w.Body.String(), `"succeeded":`) || !strings.Contains(w.Body.String(), `"offline":`) {
		t.Fatalf("expected lower-case JSON keys, got %s", w.Body.String())
	}
}

func TestOverviewRouteReturnsInternalServerErrorWhenSummaryFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewMonitorService(&fakeSummaryRepository{err: errors.New("boom")}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitor/overview", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "boom") {
		t.Fatalf("expected error body to mention boom, got %s", w.Body.String())
	}
}

func TestAlertRulesAndEventsRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &fakeSummaryRepository{
		rules: []model.AlertRule{{
			ID:                  "rule-1",
			Name:                "exec failed",
			RuleType:            model.AlertRuleTypeExecutionFailed,
			Enabled:             true,
			WebhookURL:          "https://example.com/hook",
			CooldownSeconds:     120,
			TimeoutSeconds:      5,
			OfflineGraceSeconds: 60,
			CreatedAt:           time.Now().UTC(),
			UpdatedAt:           time.Now().UTC(),
		}},
		events: []model.AlertEvent{{
			ID:             "evt-1",
			RuleID:         "rule-1",
			RuleType:       model.AlertRuleTypeExecutionFailed,
			EntityType:     "execution",
			EntityID:       "exec-1",
			DedupeKey:      "execution:exec-1",
			Payload:        `{"entityId":"exec-1"}`,
			DeliveryStatus: "sent",
			CreatedAt:      time.Now().UTC(),
		}},
	}
	router := NewRouter(service.NewMonitorService(repo))

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/monitor/alerts/rules", strings.NewReader(`{
		"name":"node offline",
		"ruleType":"node_offline",
		"webhookUrl":"https://example.com/offline"
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d body=%s", createResp.Code, createResp.Body.String())
	}

	listRulesReq := httptest.NewRequest(http.MethodGet, "/api/v1/monitor/alerts/rules", nil)
	listRulesResp := httptest.NewRecorder()
	router.ServeHTTP(listRulesResp, listRulesReq)
	if listRulesResp.Code != http.StatusOK || !strings.Contains(listRulesResp.Body.String(), `"ruleType":"execution_failed"`) {
		t.Fatalf("unexpected list rules response: %d %s", listRulesResp.Code, listRulesResp.Body.String())
	}

	listEventsReq := httptest.NewRequest(http.MethodGet, "/api/v1/monitor/alerts/events?limit=10&offset=0", nil)
	listEventsResp := httptest.NewRecorder()
	router.ServeHTTP(listEventsResp, listEventsReq)
	if listEventsResp.Code != http.StatusOK {
		t.Fatalf("expected list events status 200, got %d body=%s", listEventsResp.Code, listEventsResp.Body.String())
	}
	if !strings.Contains(listEventsResp.Body.String(), `"deliveryStatus":"sent"`) || !strings.Contains(listEventsResp.Body.String(), `"total":1`) {
		t.Fatalf("unexpected list events payload: %s", listEventsResp.Body.String())
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/monitor/alerts/rules/rule-1", strings.NewReader(`{"enabled":false}`))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp := httptest.NewRecorder()
	router.ServeHTTP(patchResp, patchReq)
	if patchResp.Code != http.StatusOK {
		t.Fatalf("expected patch status 200, got %d body=%s", patchResp.Code, patchResp.Body.String())
	}
	if !strings.Contains(patchResp.Body.String(), `"enabled":false`) {
		t.Fatalf("expected patched rule to be disabled, got %s", patchResp.Body.String())
	}
}
