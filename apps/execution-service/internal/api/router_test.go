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

	"crawler-platform/apps/execution-service/internal/model"
	"crawler-platform/apps/execution-service/internal/service"
	"github.com/gin-gonic/gin"
)

type apiFakeExecutionRepo struct {
	executions      map[string]model.Execution
	retryCandidates []string
	deleteErr       error
}

func (r *apiFakeExecutionRepo) Create(_ context.Context, exec model.Execution) (model.Execution, error) {
	if r.executions == nil {
		r.executions = map[string]model.Execution{}
	}
	r.executions[exec.ID] = exec
	return exec, nil
}

func (r *apiFakeExecutionRepo) ListByProject(_ context.Context, query service.ListExecutionsQuery) ([]model.Execution, error) {
	all := make([]model.Execution, 0)
	for _, exec := range r.executions {
		if exec.ProjectID != query.ProjectID {
			continue
		}
		if query.Status != "" && exec.Status != query.Status {
			continue
		}
		if query.From != nil && exec.CreatedAt.Before(*query.From) {
			continue
		}
		if query.To != nil && exec.CreatedAt.After(*query.To) {
			continue
		}
		all = append(all, exec)
	}
	if query.Offset >= len(all) {
		return []model.Execution{}, nil
	}
	end := query.Offset + query.Limit
	if end > len(all) {
		end = len(all)
	}
	return append([]model.Execution(nil), all[query.Offset:end]...), nil
}

func (r *apiFakeExecutionRepo) CountByProject(_ context.Context, query service.ListExecutionsQuery) (int64, error) {
	var total int64
	for _, exec := range r.executions {
		if exec.ProjectID != query.ProjectID {
			continue
		}
		if query.Status != "" && exec.Status != query.Status {
			continue
		}
		if query.From != nil && exec.CreatedAt.Before(*query.From) {
			continue
		}
		if query.To != nil && exec.CreatedAt.After(*query.To) {
			continue
		}
		total++
	}
	return total, nil
}

func (r *apiFakeExecutionRepo) Get(_ context.Context, id string) (model.Execution, error) {
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, service.ErrExecutionNotFound
	}
	return exec, nil
}

func (r *apiFakeExecutionRepo) MarkRunning(_ context.Context, id, nodeID string, startedAt time.Time) (model.Execution, error) {
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, service.ErrExecutionNotFound
	}
	if exec.Status != "pending" && exec.Status != "running" {
		return model.Execution{}, service.ErrInvalidExecutionState
	}
	exec.NodeID = nodeID
	exec.Status = "running"
	exec.StartedAt = &startedAt
	r.executions[id] = exec
	return exec, nil
}

func (r *apiFakeExecutionRepo) Complete(_ context.Context, id string, finishedAt time.Time) (model.Execution, error) {
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, service.ErrExecutionNotFound
	}
	if exec.Status != "running" {
		return model.Execution{}, service.ErrInvalidExecutionState
	}
	exec.Status = "succeeded"
	exec.FinishedAt = &finishedAt
	r.executions[id] = exec
	return exec, nil
}

func (r *apiFakeExecutionRepo) Fail(_ context.Context, id, errorMessage string, finishedAt time.Time) (model.Execution, error) {
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, service.ErrExecutionNotFound
	}
	if exec.Status != "running" {
		return model.Execution{}, service.ErrInvalidExecutionState
	}
	exec.Status = "failed"
	exec.ErrorMessage = errorMessage
	exec.FinishedAt = &finishedAt
	r.executions[id] = exec
	return exec, nil
}

func (r *apiFakeExecutionRepo) Delete(_ context.Context, id string) error {
	delete(r.executions, id)
	return r.deleteErr
}

func (r *apiFakeExecutionRepo) ClaimNextRetryCandidate(_ context.Context, _ time.Time) (model.Execution, bool, error) {
	if len(r.retryCandidates) == 0 {
		return model.Execution{}, false, nil
	}
	id := r.retryCandidates[0]
	r.retryCandidates = r.retryCandidates[1:]
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, false, service.ErrExecutionNotFound
	}
	now := time.Now().UTC()
	exec.RetriedAt = &now
	r.executions[id] = exec
	return exec, true, nil
}

func (r *apiFakeExecutionRepo) ResetRetryClaim(_ context.Context, _ string) error {
	return nil
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
	nextClaimed  []string
	acked        []string
	released     []string
	err          error
}

func (q *apiFakeQueue) Enqueue(_ context.Context, executionID string) error {
	q.lastEnqueued = executionID
	return q.err
}

func (q *apiFakeQueue) Claim(_ context.Context) (string, error) {
	if q.err != nil {
		return "", q.err
	}
	if len(q.nextClaimed) == 0 {
		return "", nil
	}
	id := q.nextClaimed[0]
	q.nextClaimed = q.nextClaimed[1:]
	return id, nil
}

func (q *apiFakeQueue) Ack(_ context.Context, executionID string) error {
	q.acked = append(q.acked, executionID)
	return q.err
}

func (q *apiFakeQueue) Release(_ context.Context, executionID string) error {
	q.released = append(q.released, executionID)
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

func TestCreateExecutionAcceptsScheduledTriggerSource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(func() *service.ExecutionService {
		svc, _, _, _ := newAPITestService()
		return svc
	}())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", strings.NewReader(`{"projectId":"project-1","spiderId":"spider-1","image":"crawler/go-echo:latest","command":["./go-echo"],"triggerSource":"scheduled"}`))
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
	if exec.TriggerSource != "scheduled" {
		t.Fatalf("expected scheduled trigger source, got %+v", exec)
	}
}

func TestCreateExecutionAcceptsRetryMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(func() *service.ExecutionService {
		svc, _, _, _ := newAPITestService()
		return svc
	}())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", strings.NewReader(`{"projectId":"project-1","spiderId":"spider-1","spiderVersion":3,"registryAuthRef":"ghcr-prod","image":"crawler/go-echo:latest","command":["./go-echo"],"triggerSource":"retry","retryLimit":3,"retryCount":1,"retryDelaySeconds":45,"retryOfExecutionId":"exec-root","cpuCores":2,"memoryMB":1024,"timeoutSeconds":180}`))
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
	if exec.RetryLimit != 3 || exec.RetryCount != 1 || exec.RetryDelaySeconds != 45 || exec.RetryOfExecutionID != "exec-root" {
		t.Fatalf("expected retry metadata in response, got %+v", exec)
	}
	if exec.SpiderVersion != 3 {
		t.Fatalf("expected spider version 3 in response, got %+v", exec)
	}
	if exec.RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("expected registryAuthRef in response, got %+v", exec)
	}
	if exec.CPUCores != 2 || exec.MemoryMB != 1024 || exec.TimeoutSeconds != 180 {
		t.Fatalf("expected resource limits in response, got %+v", exec)
	}
}

func TestListExecutionsByProject(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, execRepo, _, _ := newAPITestService()
	_, _ = svc.Create(context.Background(), service.CreateExecutionInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go:v1",
		Command:   []string{"./crawler"},
	})
	_, _ = svc.Create(context.Background(), service.CreateExecutionInput{
		ProjectID: "project-2",
		SpiderID:  "spider-2",
		Image:     "crawler/go:v2",
		Command:   []string{"./crawler"},
	})
	if len(execRepo.executions) < 2 {
		t.Fatalf("expected seeded executions, got %d", len(execRepo.executions))
	}

	router := NewRouter(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions?projectId=project-1&limit=20&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}

	var payload struct {
		Items  []model.Execution `json:"items"`
		Total  int64             `json:"total"`
		Limit  int               `json:"limit"`
		Offset int               `json:"offset"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Total != 1 || len(payload.Items) != 1 {
		t.Fatalf("unexpected paginated payload: %+v", payload)
	}
	if payload.Items[0].ProjectID != "project-1" {
		t.Fatalf("expected project-1 execution, got %+v", payload.Items[0])
	}
}

func TestListExecutionsByProjectWithStatusAndTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, repo, _, _ := newAPITestService()
	repo.executions = map[string]model.Execution{
		"exec-1": {
			ID:            "exec-1",
			ProjectID:     "project-1",
			SpiderID:      "spider-1",
			Status:        "failed",
			TriggerSource: "manual",
			Image:         "crawler/go:v1",
			Command:       []string{"./crawler"},
			CreatedAt:     time.Date(2026, 5, 1, 1, 0, 0, 0, time.UTC),
		},
		"exec-2": {
			ID:            "exec-2",
			ProjectID:     "project-1",
			SpiderID:      "spider-2",
			Status:        "succeeded",
			TriggerSource: "manual",
			Image:         "crawler/go:v2",
			Command:       []string{"./crawler"},
			CreatedAt:     time.Date(2026, 5, 1, 2, 0, 0, 0, time.UTC),
		},
		"exec-3": {
			ID:            "exec-3",
			ProjectID:     "project-1",
			SpiderID:      "spider-3",
			Status:        "pending",
			TriggerSource: "manual",
			Image:         "crawler/go:v3",
			Command:       []string{"./crawler"},
			CreatedAt:     time.Date(2026, 5, 1, 3, 0, 0, 0, time.UTC),
		},
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions?projectId=project-1&executionStatus=failed&executionFrom=2026-05-01T00:30:00Z&executionTo=2026-05-01T01:30:00Z&limit=20&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
	var payload struct {
		Items []model.Execution `json:"items"`
		Total int64             `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Total != 1 || len(payload.Items) != 1 || payload.Items[0].ID != "exec-1" {
		t.Fatalf("unexpected filtered payload: %+v", payload)
	}
}

func TestListExecutionsRejectsInvalidTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(func() *service.ExecutionService {
		svc, _, _, _ := newAPITestService()
		return svc
	}())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions?projectId=project-1&executionFrom=not-a-time", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestListExecutionsRequiresProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(func() *service.ExecutionService {
		svc, _, _, _ := newAPITestService()
		return svc
	}())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "projectId is required") {
		t.Fatalf("expected projectId error, got %s", w.Body.String())
	}
}

func TestCreateExecutionResolvesSpiderVersionWhenImageMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc, _, _, _ := newAPITestService()
	svc = svc.WithSpiderVersionResolver(&service.StaticSpiderVersionResolver{
		Version: 4,
		Image:   "crawler/go:v4",
		Command: []string{"./crawler", "--v4"},
	})
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/executions", strings.NewReader(`{"projectId":"project-1","spiderId":"spider-1","spiderVersion":4}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", w.Code, w.Body.String())
	}

	var exec model.Execution
	if err := json.Unmarshal(w.Body.Bytes(), &exec); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if exec.SpiderVersion != 4 || exec.Image != "crawler/go:v4" {
		t.Fatalf("expected resolved execution payload, got %+v", exec)
	}
}

func TestClaimNextExecutionMarksRunning(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-token")

	svc, _, _, queue := newAPITestService()
	exec, err := svc.CreateManual(context.Background(), service.CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	queue.nextClaimed = []string{exec.ID}
	router := NewRouter(svc)

	req := newInternalJSONRequest(http.MethodPost, "/internal/v1/executions/claim", `{"nodeId":"node-1"}`)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var claimed model.Execution
	if err := json.Unmarshal(w.Body.Bytes(), &claimed); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if claimed.Status != "running" || claimed.NodeID != "node-1" {
		t.Fatalf("unexpected claimed execution: %+v", claimed)
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

func TestCompleteExecutionMarksSucceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-token")

	svc, _, _, queue := newAPITestService()
	exec, err := svc.CreateManual(context.Background(), service.CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	queue.nextClaimed = []string{exec.ID}
	if _, ok, err := svc.ClaimNext(context.Background(), "node-1"); err != nil || !ok {
		t.Fatalf("ClaimNext returned ok=%v err=%v", ok, err)
	}
	router := NewRouter(svc)

	req := newInternalJSONRequest(http.MethodPost, "/internal/v1/executions/"+exec.ID+"/complete", "")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestStartExecutionReturnsClaimedExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-token")

	svc, _, _, queue := newAPITestService()
	exec, err := svc.CreateManual(context.Background(), service.CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	queue.nextClaimed = []string{exec.ID}
	if _, ok, err := svc.ClaimNext(context.Background(), "node-1"); err != nil || !ok {
		t.Fatalf("ClaimNext returned ok=%v err=%v", ok, err)
	}
	router := NewRouter(svc)

	req := newInternalJSONRequest(http.MethodPost, "/internal/v1/executions/"+exec.ID+"/start", `{"nodeId":"node-1"}`)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestStartExecutionRejectsPendingState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-token")

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

	req := newInternalJSONRequest(http.MethodPost, "/internal/v1/executions/"+exec.ID+"/start", `{"nodeId":"node-1"}`)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
}

func TestFailExecutionMarksFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-token")

	svc, _, _, queue := newAPITestService()
	exec, err := svc.CreateManual(context.Background(), service.CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	queue.nextClaimed = []string{exec.ID}
	if _, ok, err := svc.ClaimNext(context.Background(), "node-1"); err != nil || !ok {
		t.Fatalf("ClaimNext returned ok=%v err=%v", ok, err)
	}
	router := NewRouter(svc)

	req := newInternalJSONRequest(http.MethodPost, "/internal/v1/executions/"+exec.ID+"/fail", `{"error":"exit status 1"}`)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestMaterializeRetryCreatesExecution(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-token")

	svc, execRepo, _, _ := newAPITestService()
	finishedAt := time.Now().UTC()
	execRepo.executions = map[string]model.Execution{
		"failed-1": {
			ID:                "failed-1",
			ProjectID:         "project-1",
			SpiderID:          "spider-1",
			Status:            "failed",
			TriggerSource:     "scheduled",
			Image:             "crawler/go-echo:latest",
			Command:           []string{"./go-echo"},
			RetryLimit:        2,
			RetryCount:        0,
			RetryDelaySeconds: 30,
			FinishedAt:        &finishedAt,
		},
	}
	execRepo.retryCandidates = []string{"failed-1"}
	router := NewRouter(svc)

	req := newInternalJSONRequest(http.MethodPost, "/internal/v1/executions/retries/materialize", "")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var exec model.Execution
	if err := json.Unmarshal(w.Body.Bytes(), &exec); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if exec.TriggerSource != "retry" || exec.RetryCount != 1 || exec.RetryOfExecutionID != "failed-1" {
		t.Fatalf("unexpected retry execution: %+v", exec)
	}
}

func TestCompleteExecutionRejectsPendingState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-token")

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

	req := newInternalJSONRequest(http.MethodPost, "/internal/v1/executions/"+exec.ID+"/complete", "")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
}

func TestInternalExecutionRoutesRequireToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-token")

	svc, _, _, _ := newAPITestService()
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/executions/claim", strings.NewReader(`{"nodeId":"node-1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func newInternalJSONRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalTokenHeader, "test-token")
	return req
}
