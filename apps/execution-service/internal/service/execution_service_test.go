package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"crawler-platform/apps/execution-service/internal/model"
)

type fakeExecutionRepo struct {
	created         []model.Execution
	executions      map[string]model.Execution
	retryCandidates []string
	deleted         []string
	deleteErr       error
	markErr         error
}

func newFakeExecutionRepo() *fakeExecutionRepo {
	return &fakeExecutionRepo{executions: map[string]model.Execution{}}
}

func (r *fakeExecutionRepo) Create(_ context.Context, exec model.Execution) (model.Execution, error) {
	r.created = append(r.created, exec)
	r.executions[exec.ID] = exec
	return exec, nil
}

func (r *fakeExecutionRepo) ListByProject(_ context.Context, projectID string, limit, offset int) ([]model.Execution, error) {
	all := make([]model.Execution, 0)
	for _, exec := range r.executions {
		if exec.ProjectID == projectID {
			all = append(all, exec)
		}
	}
	if offset >= len(all) {
		return []model.Execution{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return append([]model.Execution(nil), all[offset:end]...), nil
}

func (r *fakeExecutionRepo) CountByProject(_ context.Context, projectID string) (int64, error) {
	var total int64
	for _, exec := range r.executions {
		if exec.ProjectID == projectID {
			total++
		}
	}
	return total, nil
}

func (r *fakeExecutionRepo) Get(_ context.Context, id string) (model.Execution, error) {
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, ErrExecutionNotFound
	}
	return exec, nil
}

func (r *fakeExecutionRepo) MarkRunning(_ context.Context, id, nodeID string, startedAt time.Time) (model.Execution, error) {
	if r.markErr != nil {
		return model.Execution{}, r.markErr
	}
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, ErrExecutionNotFound
	}
	if exec.Status != "pending" && exec.Status != "running" {
		return model.Execution{}, ErrInvalidExecutionState
	}
	exec.NodeID = nodeID
	exec.Status = "running"
	exec.StartedAt = &startedAt
	r.executions[id] = exec
	return exec, nil
}

func (r *fakeExecutionRepo) Complete(_ context.Context, id string, finishedAt time.Time) (model.Execution, error) {
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, ErrExecutionNotFound
	}
	if exec.Status != "running" {
		return model.Execution{}, ErrInvalidExecutionState
	}
	exec.Status = "succeeded"
	exec.FinishedAt = &finishedAt
	r.executions[id] = exec
	return exec, nil
}

func (r *fakeExecutionRepo) Fail(_ context.Context, id, errorMessage string, finishedAt time.Time) (model.Execution, error) {
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, ErrExecutionNotFound
	}
	if exec.Status != "running" {
		return model.Execution{}, ErrInvalidExecutionState
	}
	exec.Status = "failed"
	exec.ErrorMessage = errorMessage
	exec.FinishedAt = &finishedAt
	r.executions[id] = exec
	return exec, nil
}

func (r *fakeExecutionRepo) Delete(_ context.Context, id string) error {
	r.deleted = append(r.deleted, id)
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.executions, id)
	return nil
}

func (r *fakeExecutionRepo) ClaimNextRetryCandidate(_ context.Context, _ time.Time) (model.Execution, bool, error) {
	if len(r.retryCandidates) == 0 {
		return model.Execution{}, false, nil
	}
	id := r.retryCandidates[0]
	r.retryCandidates = r.retryCandidates[1:]
	exec, ok := r.executions[id]
	if !ok {
		return model.Execution{}, false, ErrExecutionNotFound
	}
	now := time.Now().UTC()
	exec.RetriedAt = &now
	r.executions[id] = exec
	return exec, true, nil
}

func (r *fakeExecutionRepo) ResetRetryClaim(_ context.Context, id string) error {
	exec, ok := r.executions[id]
	if !ok {
		return ErrExecutionNotFound
	}
	exec.RetriedAt = nil
	r.executions[id] = exec
	return nil
}

type fakeLogRepo struct {
	initialized []string
	appended    []model.ExecutionLog
	logs        map[string][]model.ExecutionLog
	initErr     error
	listErr     error
}

func newFakeLogRepo() *fakeLogRepo {
	return &fakeLogRepo{logs: map[string][]model.ExecutionLog{}}
}

func (r *fakeLogRepo) Init(_ context.Context, executionID string) error {
	if r.initErr != nil {
		return r.initErr
	}
	r.initialized = append(r.initialized, executionID)
	if _, ok := r.logs[executionID]; !ok {
		r.logs[executionID] = nil
	}
	return nil
}

func (r *fakeLogRepo) Append(_ context.Context, entry model.ExecutionLog) error {
	r.appended = append(r.appended, entry)
	r.logs[entry.ExecutionID] = append(r.logs[entry.ExecutionID], entry)
	return nil
}

func (r *fakeLogRepo) List(_ context.Context, executionID string) ([]model.ExecutionLog, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	logs := append([]model.ExecutionLog(nil), r.logs[executionID]...)
	return logs, nil
}

type fakeQueue struct {
	lastEnqueued string
	nextClaimed  []string
	acked        []string
	released     []string
	err          error
}

func (q *fakeQueue) Enqueue(_ context.Context, executionID string) error {
	q.lastEnqueued = executionID
	return q.err
}

func (q *fakeQueue) Claim(_ context.Context) (string, error) {
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

func (q *fakeQueue) Ack(_ context.Context, executionID string) error {
	q.acked = append(q.acked, executionID)
	return q.err
}

func (q *fakeQueue) Release(_ context.Context, executionID string) error {
	q.released = append(q.released, executionID)
	return q.err
}

type fakeSpiderVersionResolver struct {
	version         int
	registryAuthRef string
	image           string
	command         []string
	err             error
}

func (r *fakeSpiderVersionResolver) Resolve(_ context.Context, _ string, _ int) (int, string, string, []string, error) {
	if r.err != nil {
		return 0, "", "", nil, r.err
	}
	return r.version, r.registryAuthRef, r.image, append([]string(nil), r.command...), nil
}

func TestCreateManualEnqueuesPendingExecution(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}

	if exec.Status != "pending" {
		t.Fatalf("expected pending status, got %s", exec.Status)
	}
	if exec.TriggerSource != "manual" {
		t.Fatalf("expected manual trigger source, got %s", exec.TriggerSource)
	}
	if execRepo.created == nil || len(execRepo.created) != 1 {
		t.Fatalf("expected execution to be persisted once, got %+v", execRepo.created)
	}
	if logRepo.initialized == nil || len(logRepo.initialized) != 1 || logRepo.initialized[0] != exec.ID {
		t.Fatalf("expected log storage to be initialized for %s, got %+v", exec.ID, logRepo.initialized)
	}
	if queue.lastEnqueued != exec.ID {
		t.Fatalf("expected execution %s to be enqueued, got %s", exec.ID, queue.lastEnqueued)
	}
	if exec.ProjectID != "project-1" || exec.SpiderID != "spider-1" || exec.Image != "crawler/go-echo:latest" {
		t.Fatalf("unexpected execution fields: %+v", exec)
	}
	if got := exec.Command; len(got) != 1 || got[0] != "./go-echo" {
		t.Fatalf("unexpected command: %+v", got)
	}
}

func TestCreateExecutionUsesProvidedTriggerSource(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.Create(context.Background(), CreateExecutionInput{
		ProjectID:          "project-1",
		SpiderID:           "spider-1",
		SpiderVersion:      2,
		RegistryAuthRef:    "ghcr-prod",
		Image:              "crawler/go-echo:latest",
		Command:            []string{"./go-echo"},
		TriggerSource:      "scheduled",
		RetryLimit:         3,
		RetryCount:         1,
		RetryDelaySeconds:  45,
		RetryOfExecutionID: "exec-root",
		CPUCores:           1.5,
		MemoryMB:           768,
		TimeoutSeconds:     120,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if exec.TriggerSource != "scheduled" {
		t.Fatalf("expected scheduled trigger source, got %+v", exec)
	}
	if exec.RetryLimit != 3 || exec.RetryCount != 1 || exec.RetryDelaySeconds != 45 || exec.RetryOfExecutionID != "exec-root" {
		t.Fatalf("expected retry metadata to persist, got %+v", exec)
	}
	if exec.CPUCores != 1.5 || exec.MemoryMB != 768 || exec.TimeoutSeconds != 120 {
		t.Fatalf("expected resource limits to persist, got %+v", exec)
	}
	if exec.SpiderVersion != 2 {
		t.Fatalf("expected spider version to persist, got %+v", exec)
	}
	if exec.RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("expected registry auth ref to persist, got %+v", exec)
	}
	if queue.lastEnqueued != exec.ID {
		t.Fatalf("expected execution %s to be enqueued, got %s", exec.ID, queue.lastEnqueued)
	}
}

func TestCreateExecutionResolvesSpiderVersionWhenImageAndCommandMissing(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	resolver := &fakeSpiderVersionResolver{
		version:         3,
		registryAuthRef: "ghcr-prod",
		image:           "crawler/go:v3",
		command:         []string{"./crawler", "--v3"},
	}
	svc := NewExecutionService(execRepo, logRepo, queue).WithSpiderVersionResolver(resolver)

	exec, err := svc.Create(context.Background(), CreateExecutionInput{
		ProjectID:     "project-1",
		SpiderID:      "spider-1",
		SpiderVersion: 3,
		TriggerSource: "manual",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if exec.SpiderVersion != 3 || exec.Image != "crawler/go:v3" {
		t.Fatalf("expected resolved spider version/image, got %+v", exec)
	}
	if len(exec.Command) != 2 || exec.Command[0] != "./crawler" || exec.Command[1] != "--v3" {
		t.Fatalf("expected resolved command, got %+v", exec.Command)
	}
	if exec.RegistryAuthRef != "ghcr-prod" {
		t.Fatalf("expected resolved registry auth ref, got %+v", exec)
	}
}

func TestCreateExecutionKeepsProvidedRegistryAuthRefOverResolvedOne(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	resolver := &fakeSpiderVersionResolver{
		version:         3,
		registryAuthRef: "ghcr-default",
		image:           "crawler/go:v3",
		command:         []string{"./crawler", "--v3"},
	}
	svc := NewExecutionService(execRepo, logRepo, queue).WithSpiderVersionResolver(resolver)

	exec, err := svc.Create(context.Background(), CreateExecutionInput{
		ProjectID:       "project-1",
		SpiderID:        "spider-1",
		SpiderVersion:   3,
		RegistryAuthRef: "ghcr-override",
		TriggerSource:   "manual",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if exec.RegistryAuthRef != "ghcr-override" {
		t.Fatalf("expected explicit registry auth ref to win, got %+v", exec)
	}
}

func TestCreateExecutionResolvesRegistryAuthRefWhenImageProvided(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	resolver := &fakeSpiderVersionResolver{
		version:         3,
		registryAuthRef: "ghcr-derived",
		image:           "crawler/go:v3-from-resolver",
		command:         []string{"./crawler", "--resolver"},
	}
	svc := NewExecutionService(execRepo, logRepo, queue).WithSpiderVersionResolver(resolver)

	exec, err := svc.Create(context.Background(), CreateExecutionInput{
		ProjectID:     "project-1",
		SpiderID:      "spider-1",
		SpiderVersion: 3,
		Image:         "crawler/go:v3-explicit",
		Command:       []string{"./crawler", "--explicit"},
		TriggerSource: "manual",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if exec.RegistryAuthRef != "ghcr-derived" {
		t.Fatalf("expected registry auth ref from resolver, got %+v", exec)
	}
	if exec.Image != "crawler/go:v3-explicit" {
		t.Fatalf("expected explicit image to remain unchanged, got %+v", exec)
	}
	if len(exec.Command) != 2 || exec.Command[1] != "--explicit" {
		t.Fatalf("expected explicit command to remain unchanged, got %+v", exec.Command)
	}
}

func TestMaterializeRetryCreatesNextAttempt(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	finishedAt := time.Now().UTC()
	failed := model.Execution{
		ID:                "failed-1",
		ProjectID:         "project-1",
		SpiderID:          "spider-1",
		Status:            "failed",
		TriggerSource:     "scheduled",
		Image:             "crawler/go-echo:latest",
		Command:           []string{"./go-echo"},
		RetryLimit:        3,
		RetryCount:        0,
		RetryDelaySeconds: 30,
		FinishedAt:        &finishedAt,
	}
	execRepo.executions[failed.ID] = failed
	execRepo.retryCandidates = []string{failed.ID}

	retried, ok, err := svc.MaterializeRetry(context.Background())
	if err != nil {
		t.Fatalf("MaterializeRetry returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected retry materialization to enqueue a new execution")
	}
	if retried.TriggerSource != "retry" {
		t.Fatalf("expected retry trigger source, got %+v", retried)
	}
	if retried.RetryCount != 1 || retried.RetryLimit != 3 || retried.RetryDelaySeconds != 30 {
		t.Fatalf("expected retry metadata to increment, got %+v", retried)
	}
	if retried.RetryOfExecutionID != failed.ID {
		t.Fatalf("expected retry_of to point at failed execution, got %+v", retried)
	}
}

func TestClaimNextExecutionMarksRunning(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	queue.nextClaimed = []string{exec.ID}

	claimed, ok, err := svc.ClaimNext(context.Background(), "node-1")
	if err != nil {
		t.Fatalf("ClaimNext returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ClaimNext to return a claimed execution")
	}
	if claimed.Status != "running" || claimed.NodeID != "node-1" || claimed.StartedAt == nil {
		t.Fatalf("expected running claimed execution, got %+v", claimed)
	}
	if len(queue.acked) != 0 {
		t.Fatalf("expected claimed execution to remain inflight until completion, got %+v", queue.acked)
	}
}

func TestClaimNextRequeuesExecutionWhenMarkRunningFails(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	execRepo.markErr = errors.New("postgres unavailable")
	queue.nextClaimed = []string{exec.ID}

	_, ok, err := svc.ClaimNext(context.Background(), "node-1")
	if err == nil || err.Error() != "postgres unavailable" {
		t.Fatalf("expected mark running error, got %v", err)
	}
	if ok {
		t.Fatal("expected no claimed execution on failure")
	}
	if len(queue.released) != 1 || queue.released[0] != exec.ID {
		t.Fatalf("expected failed claim to be released, got %+v", queue.released)
	}
}

func TestCreateManualRollsBackWhenLogInitFails(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	logRepo.initErr = errors.New("mongo unavailable")
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	_, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err == nil {
		t.Fatal("expected CreateManual to fail")
	}
	if len(execRepo.deleted) != 1 {
		t.Fatalf("expected persisted execution to be rolled back, got %+v", execRepo.deleted)
	}
	if queue.lastEnqueued != "" {
		t.Fatalf("expected queue not to be used, got %s", queue.lastEnqueued)
	}
}

func TestCreateManualRollsBackWhenQueueEnqueueFails(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{err: errors.New("redis unavailable")}
	svc := NewExecutionService(execRepo, logRepo, queue)

	_, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err == nil {
		t.Fatal("expected CreateManual to fail")
	}
	if len(execRepo.deleted) != 1 {
		t.Fatalf("expected persisted execution to be rolled back, got %+v", execRepo.deleted)
	}
}

func TestCreateManualReturnsJoinedErrorWhenRollbackFails(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	execRepo.deleteErr = errors.New("delete failed")
	logRepo := newFakeLogRepo()
	logRepo.initErr = errors.New("mongo unavailable")
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	_, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err == nil {
		t.Fatal("expected CreateManual to fail")
	}
	if !errors.Is(err, logRepo.initErr) {
		t.Fatalf("expected joined error to include init error, got %v", err)
	}
	if !strings.Contains(err.Error(), "rollback execution") || !strings.Contains(err.Error(), "delete failed") {
		t.Fatalf("expected rollback delete failure to be surfaced, got %v", err)
	}
}

func TestAppendLogPersistsThroughLogRepo(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}

	entry, err := svc.AppendLog(context.Background(), exec.ID, "started")
	if err != nil {
		t.Fatalf("AppendLog returned error: %v", err)
	}
	if entry.Message != "started" {
		t.Fatalf("expected log message started, got %s", entry.Message)
	}
	if len(logRepo.appended) != 1 || logRepo.appended[0].ExecutionID != exec.ID {
		t.Fatalf("expected log repo append for %s, got %+v", exec.ID, logRepo.appended)
	}
}

func TestGetAndGetLogsReadThroughRepos(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}

	createdAt := time.Now().UTC()
	extra := model.ExecutionLog{ID: "log-1", ExecutionID: exec.ID, Message: "started", CreatedAt: createdAt}
	logRepo.logs[exec.ID] = []model.ExecutionLog{extra}

	got, err := svc.Get(context.Background(), exec.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if len(got.Logs) != 1 || got.Logs[0].Message != "started" {
		t.Fatalf("unexpected execution logs: %+v", got.Logs)
	}

	logs, err := svc.GetLogs(context.Background(), exec.ID)
	if err != nil {
		t.Fatalf("GetLogs returned error: %v", err)
	}
	if len(logs) != 1 || logs[0].Message != "started" {
		t.Fatalf("unexpected logs: %+v", logs)
	}
}

func TestAppendLogReturnsNotFoundForMissingExecution(t *testing.T) {
	svc := NewExecutionService(newFakeExecutionRepo(), newFakeLogRepo(), &fakeQueue{})

	_, err := svc.AppendLog(context.Background(), "missing", "started")
	if !errors.Is(err, ErrExecutionNotFound) {
		t.Fatalf("expected ErrExecutionNotFound, got %v", err)
	}
}

func TestGetReturnsBackendError(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}

	logRepo.listErr = errors.New("mongo unavailable")
	_, err = svc.Get(context.Background(), exec.ID)
	if err == nil || err.Error() != "mongo unavailable" {
		t.Fatalf("expected backend error, got %v", err)
	}
}

func TestGetReturnsNotFound(t *testing.T) {
	svc := NewExecutionService(newFakeExecutionRepo(), newFakeLogRepo(), &fakeQueue{})

	_, err := svc.Get(context.Background(), "missing")
	if !errors.Is(err, ErrExecutionNotFound) {
		t.Fatalf("expected ErrExecutionNotFound, got %v", err)
	}
}

func TestCompleteMarksExecutionSucceeded(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
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

	got, err := svc.Complete(context.Background(), exec.ID)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if got.Status != "succeeded" || got.FinishedAt == nil {
		t.Fatalf("expected succeeded execution, got %+v", got)
	}
	if len(queue.acked) != 1 || queue.acked[0] != exec.ID {
		t.Fatalf("expected completed execution to be acked, got %+v", queue.acked)
	}
}

func TestStartReturnsClaimedExecution(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
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

	got, err := svc.Start(context.Background(), exec.ID, "node-1")
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if got.Status != "running" || got.NodeID != "node-1" || got.StartedAt == nil {
		t.Fatalf("expected running execution, got %+v", got)
	}
}

func TestStartRejectsPendingExecution(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}

	_, err = svc.Start(context.Background(), exec.ID, "node-1")
	if !errors.Is(err, ErrInvalidExecutionState) {
		t.Fatalf("expected ErrInvalidExecutionState, got %v", err)
	}
}

func TestStartRejectsFinishedExecution(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}
	finished := time.Now().UTC()
	stored := execRepo.executions[exec.ID]
	stored.Status = "succeeded"
	stored.FinishedAt = &finished
	execRepo.executions[exec.ID] = stored

	_, err = svc.Start(context.Background(), exec.ID, "node-1")
	if !errors.Is(err, ErrInvalidExecutionState) {
		t.Fatalf("expected ErrInvalidExecutionState, got %v", err)
	}
}

func TestFailMarksExecutionFailed(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
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

	got, err := svc.Fail(context.Background(), exec.ID, "exit status 1")
	if err != nil {
		t.Fatalf("Fail returned error: %v", err)
	}
	if got.Status != "failed" || got.ErrorMessage != "exit status 1" || got.FinishedAt == nil {
		t.Fatalf("expected failed execution, got %+v", got)
	}
	if len(queue.acked) != 1 || queue.acked[0] != exec.ID {
		t.Fatalf("expected failed execution to be acked, got %+v", queue.acked)
	}
}

func TestCompleteRetriesAckForAlreadySucceededExecution(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
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

	queue.err = errors.New("ack failed")
	if _, err := svc.Complete(context.Background(), exec.ID); err == nil {
		t.Fatal("expected Complete to surface ack failure")
	}

	queue.err = nil
	got, err := svc.Complete(context.Background(), exec.ID)
	if err != nil {
		t.Fatalf("retry Complete returned error: %v", err)
	}
	if got.Status != "succeeded" {
		t.Fatalf("expected succeeded execution on retry, got %+v", got)
	}
	if len(queue.acked) != 2 || queue.acked[0] != exec.ID || queue.acked[1] != exec.ID {
		t.Fatalf("expected ack retried for succeeded execution, got %+v", queue.acked)
	}
}

func TestFailRetriesAckForAlreadyFailedExecution(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
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

	queue.err = errors.New("ack failed")
	if _, err := svc.Fail(context.Background(), exec.ID, "exit status 1"); err == nil {
		t.Fatal("expected Fail to surface ack failure")
	}

	queue.err = nil
	got, err := svc.Fail(context.Background(), exec.ID, "exit status 1")
	if err != nil {
		t.Fatalf("retry Fail returned error: %v", err)
	}
	if got.Status != "failed" {
		t.Fatalf("expected failed execution on retry, got %+v", got)
	}
	if len(queue.acked) != 2 || queue.acked[0] != exec.ID || queue.acked[1] != exec.ID {
		t.Fatalf("expected ack retried for failed execution, got %+v", queue.acked)
	}
}

func TestCompleteRejectsPendingExecution(t *testing.T) {
	execRepo := newFakeExecutionRepo()
	logRepo := newFakeLogRepo()
	queue := &fakeQueue{}
	svc := NewExecutionService(execRepo, logRepo, queue)

	exec, err := svc.CreateManual(context.Background(), CreateManualInput{
		ProjectID: "project-1",
		SpiderID:  "spider-1",
		Image:     "crawler/go-echo:latest",
		Command:   []string{"./go-echo"},
	})
	if err != nil {
		t.Fatalf("CreateManual returned error: %v", err)
	}

	_, err = svc.Complete(context.Background(), exec.ID)
	if !errors.Is(err, ErrInvalidExecutionState) {
		t.Fatalf("expected ErrInvalidExecutionState, got %v", err)
	}
}
