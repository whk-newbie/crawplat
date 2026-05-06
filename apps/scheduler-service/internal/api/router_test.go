// Package api 测试：验证路由层的请求参数绑定、响应格式和状态码。
//
// 该文件负责：
//   - TestCreateScheduleReturnsLowerCaseJSON：验证创建调度的请求/响应 JSON 字段为小写 camelCase。
//   - TestCreateScheduleAcceptsRetryConfiguration：验证 retryLimit 和 retryDelaySeconds 参数能正确透传。
//   - TestListSchedulesReturnsLowerCaseJSON：验证列表接口返回的 JSON 字段格式。
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

// TestCreateScheduleReturnsLowerCaseJSON 验证 POST /api/v1/schedules 返回的 JSON 使用小写 camelCase 键名。
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

// TestCreateScheduleAcceptsRetryConfiguration 验证重试配置（retryLimit、retryDelaySeconds）能正确传入并返回。
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

// TestListSchedulesReturnsLowerCaseJSON 验证 GET /api/v1/schedules 返回小写 camelCase JSON 数组。
func TestListSchedulesReturnsLowerCaseJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSchedulerService(nil, nil)
	if _, err := svc.Create("", "project-1", "spider-1", "", "", "nightly", "0 * * * *", "crawler/go-echo:latest", []string{"./go-echo"}, true, 0, 0); err != nil {
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
