package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/execution-service/internal/model"
	"crawler-platform/apps/execution-service/internal/service"
	"github.com/gin-gonic/gin"
)

func TestCreateExecutionReturnsPendingExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewExecutionService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", strings.NewReader(`{"taskId":"task-1","spiderVersionId":"spider-v1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var exec model.Execution
	if err := json.Unmarshal(w.Body.Bytes(), &exec); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if exec.ID == "" {
		t.Fatal("expected generated id")
	}
	if exec.TaskID != "task-1" || exec.SpiderVersionID != "spider-v1" || exec.Status != "pending" || exec.TriggerSource != "manual" {
		t.Fatalf("unexpected execution contents: %+v", exec)
	}
}

func TestAppendLogAndReadExecutionLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewExecutionService()
	exec := svc.CreateManual("task-1", "spider-v1")
	router := NewRouter(svc)

	appendReq := httptest.NewRequest(http.MethodPost, "/api/v1/executions/"+exec.ID+"/logs", strings.NewReader(`{"message":"started"}`))
	appendReq.Header.Set("Content-Type", "application/json")
	appendW := httptest.NewRecorder()
	router.ServeHTTP(appendW, appendReq)

	if appendW.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", appendW.Code)
	}

	getLogsReq := httptest.NewRequest(http.MethodGet, "/api/v1/executions/"+exec.ID+"/logs", nil)
	getLogsW := httptest.NewRecorder()
	router.ServeHTTP(getLogsW, getLogsReq)

	if getLogsW.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getLogsW.Code)
	}

	var logs []model.ExecutionLog
	if err := json.Unmarshal(getLogsW.Body.Bytes(), &logs); err != nil {
		t.Fatalf("failed to decode logs response: %v", err)
	}
	if len(logs) != 1 || logs[0].Message != "started" {
		t.Fatalf("unexpected logs response: %+v", logs)
	}
}

func TestGetExecutionIncludesStoredLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewExecutionService()
	exec := svc.CreateManual("task-1", "spider-v1")
	if _, err := svc.AppendLog(exec.ID, "started"); err != nil {
		t.Fatalf("expected append log success, got error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/"+exec.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var got model.Execution
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode execution response: %v", err)
	}
	if len(got.Logs) != 1 || got.Logs[0].Message != "started" {
		t.Fatalf("unexpected execution logs: %+v", got.Logs)
	}
}
