package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/scheduler-service/internal/model"
	"crawler-platform/apps/scheduler-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
	"github.com/gin-gonic/gin"
)

func TestCreateScheduleReturnsLowerCaseJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSchedulerService(nil, nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/schedules", strings.NewReader(`{"projectId":"project-1","spiderId":"spider-1","name":"nightly","cronExpr":"0 * * * *","enabled":true,"image":"crawler/go-echo:latest","command":["./go-echo"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var schedule model.Schedule
	if err := json.Unmarshal(w.Body.Bytes(), &schedule); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if schedule.ID == "" {
		t.Fatal("expected generated id")
	}
	if schedule.Name != "nightly" || schedule.CronExpr != "0 * * * *" {
		t.Fatalf("unexpected schedule: %+v", schedule)
	}
	if !strings.Contains(w.Body.String(), `"projectId":`) || !strings.Contains(w.Body.String(), `"cronExpr":`) {
		t.Fatalf("expected lower-case JSON keys, got %s", w.Body.String())
	}
}

func TestCreateScheduleAcceptsRetryConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSchedulerService(nil, nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/schedules", strings.NewReader(`{"projectId":"project-1","spiderId":"spider-1","name":"nightly","cronExpr":"0 * * * *","enabled":true,"image":"crawler/go-echo:latest","command":["./go-echo"],"retryLimit":2,"retryDelaySeconds":30}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var schedule model.Schedule
	if err := json.Unmarshal(w.Body.Bytes(), &schedule); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if schedule.RetryLimit != 2 || schedule.RetryDelaySeconds != 30 {
		t.Fatalf("expected retry config in response, got %+v", schedule)
	}
}

func TestCreateScheduleAcceptsSpiderVersionWithoutImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSchedulerService(nil, nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/schedules", strings.NewReader(`{"projectId":"project-1","spiderId":"spider-1","spiderVersion":2,"name":"nightly","cronExpr":"0 * * * *","enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", w.Code, w.Body.String())
	}

	var schedule model.Schedule
	if err := json.Unmarshal(w.Body.Bytes(), &schedule); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if schedule.SpiderVersion != 2 {
		t.Fatalf("expected spiderVersion=2, got %+v", schedule)
	}
}

func TestListSchedulesReturnsLowerCaseJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSchedulerService(nil, nil)
	if _, err := svc.Create("project-1", "spider-1", 0, "nightly", "0 * * * *", "crawler/go-echo:latest", []string{"./go-echo"}, true, 0, 0); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload struct {
		Items  []model.Schedule `json:"items"`
		Total  int64            `json:"total"`
		Limit  int              `json:"limit"`
		Offset int              `json:"offset"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload.Items) != 1 || payload.Total != 1 {
		t.Fatalf("unexpected schedule payload: %+v", payload)
	}
	if payload.Items[0].Name != "nightly" {
		t.Fatalf("unexpected schedule: %+v", payload.Items[0])
	}
	if payload.Limit != 20 || payload.Offset != 0 {
		t.Fatalf("expected default pagination, got limit=%d offset=%d", payload.Limit, payload.Offset)
	}
}

func TestListSchedulesRespectsPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSchedulerService(nil, nil)
	for i := 0; i < 3; i++ {
		if _, err := svc.Create("project-1", "spider-1", 0, "nightly", "0 * * * *", "crawler/go-echo:latest", []string{"./go-echo"}, true, 0, 0); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules?limit=1&offset=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var payload httpx.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Total != 3 || payload.Limit != 1 || payload.Offset != 1 {
		t.Fatalf("unexpected pagination payload: %+v", payload)
	}
}
