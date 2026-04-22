package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/execution-service/internal/model"
	"crawler-platform/apps/execution-service/internal/service"
	"github.com/gin-gonic/gin"
)

type apiFakeExecutionRepo struct {
	executions map[string]model.Execution
	deleteErr  error
}

func (r *apiFakeExecutionRepo) Create(_ context.Context, exec model.Execution) (model.Execution, error) {
	if r.executions == nil {
		r.executions = map[string]model.Execution{}
	}
	r.executions[exec.ID] = exec
	return exec, nil
}

func (r *apiFakeExecutionRepo) Get(_ context.Context, id string) (model.Execution, error) {
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, service.ErrExecutionNotFound
	}
	return exec, nil
}

func (r *apiFakeExecutionRepo) Delete(_ context.Context, id string) error {
	delete(r.executions, id)
	return r.deleteErr
}

type apiFakeLogRepo struct {
	logs      map[string][]model.ExecutionLog
	appendErr error
	listErr   error
}

func (r *apiFakeLogRepo) Init(_ context.Context, executionID string) error {
	if r.logs == nil {
		r.logs = map[string][]model.ExecutionLog{}
	}
	if _, ok := r.logs[executionID]; !ok {
		r.logs[executionID] = nil
	}
	return nil
}

func (r *apiFakeLogRepo) Append(_ context.Context, entry model.ExecutionLog) error {
	if r.appendErr != nil {
		return r.appendErr
	}
	r.logs[entry.ExecutionID] = append(r.logs[entry.ExecutionID], entry)
	return nil
}

func (r *apiFakeLogRepo) List(_ context.Context, executionID string) ([]model.ExecutionLog, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return append([]model.ExecutionLog(nil), r.logs[executionID]...), nil
}

type apiFakeQueue struct {
	lastEnqueued string
	err          error
}

func (q *apiFakeQueue) Enqueue(_ context.Context, executionID string) error {
	q.lastEnqueued = executionID
	return q.err
}

func newAPITestService() (*service.ExecutionService, *apiFakeExecutionRepo, *apiFakeLogRepo, *apiFakeQueue) {
	execRepo := &apiFakeExecutionRepo{}
	logRepo := &apiFakeLogRepo{}
	queue := &apiFakeQueue{}
	return service.NewExecutionService(execRepo, logRepo, queue), execRepo, logRepo, queue
}

func TestCreateExecutionReturnsPendingExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(func() *service.ExecutionService {
		svc, _, _, _ := newAPITestService()
		return svc
	}())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", strings.NewReader(`{"projectId":"project-1","spiderId":"spider-1","image":"crawler/go-echo:latest","command":["./go-echo"]}`))
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
	if exec.ProjectID != "project-1" || exec.SpiderID != "spider-1" || exec.Status != "pending" || exec.TriggerSource != "manual" {
		t.Fatalf("unexpected execution contents: %+v", exec)
	}
	if exec.Image != "crawler/go-echo:latest" {
		t.Fatalf("unexpected image: %+v", exec)
	}
	if len(exec.Command) != 1 || exec.Command[0] != "./go-echo" {
		t.Fatalf("unexpected command: %+v", exec.Command)
	}
}

func TestAppendLogAndReadExecutionLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _, _ := newAPITestService()
	exec, err := svc.CreateManual(context.Background(), service.CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
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

	svc, _, _, _ := newAPITestService()
	exec, err := svc.CreateManual(context.Background(), service.CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	if _, err := svc.AppendLog(context.Background(), exec.ID, "started"); err != nil {
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

func TestGetExecutionReturnsInternalServerErrorForBackendFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, logRepo, _ := newAPITestService()
	exec, err := svc.CreateManual(context.Background(), service.CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	logRepo.listErr = errors.New("mongo unavailable")
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/"+exec.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestGetExecutionReturnsNotFoundForMissingExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _, _ := newAPITestService()
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/missing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestAppendLogReturnsNotFoundForMissingExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _, _ := newAPITestService()
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions/missing/logs", strings.NewReader(`{"message":"started"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}
