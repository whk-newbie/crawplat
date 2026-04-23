package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/scheduler-service/internal/model"
	"crawler-platform/apps/scheduler-service/internal/service"
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

func TestListSchedulesReturnsLowerCaseJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSchedulerService(nil, nil)
	if _, err := svc.Create("project-1", "spider-1", "nightly", "0 * * * *", "crawler/go-echo:latest", []string{"./go-echo"}, true, 0, 0); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var schedules []model.Schedule
	if err := json.Unmarshal(w.Body.Bytes(), &schedules); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedules))
	}
	if schedules[0].Name != "nightly" {
		t.Fatalf("unexpected schedule: %+v", schedules[0])
	}
}
