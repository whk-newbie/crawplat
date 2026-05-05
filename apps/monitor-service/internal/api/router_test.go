// Package api 的 HTTP 层集成测试。使用 fake repository 验证 /monitor/overview 端点。
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
	overview model.Overview
	err      error
}

func (r *fakeSummaryRepository) Overview(_ context.Context) (model.Overview, error) {
	if r.err != nil {
		return model.Overview{}, r.err
	}
	return r.overview, nil
}

func (r *fakeSummaryRepository) CreateAlertRule(_ context.Context, rule model.AlertRule) (model.AlertRule, error) {
	return rule, nil
}

func (r *fakeSummaryRepository) UpdateAlertRule(_ context.Context, _ string, _ model.AlertRulePatch) (model.AlertRule, bool, error) {
	return model.AlertRule{}, false, nil
}

func (r *fakeSummaryRepository) ListAlertRules(_ context.Context) ([]model.AlertRule, error) {
	return nil, nil
}

func (r *fakeSummaryRepository) ListAlertEvents(_ context.Context, _, _ int) ([]model.AlertEvent, error) {
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
