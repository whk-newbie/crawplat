package service

import (
	"errors"
	"sync"
	"time"

	"crawler-platform/apps/execution-service/internal/model"
	"github.com/google/uuid"
)

var ErrExecutionNotFound = errors.New("execution not found")

type ExecutionService struct {
	mu         sync.Mutex
	executions map[string]model.Execution
	logs       map[string][]model.ExecutionLog
}

func NewExecutionService() *ExecutionService {
	return &ExecutionService{
		executions: map[string]model.Execution{},
		logs:       map[string][]model.ExecutionLog{},
	}
}

func (s *ExecutionService) CreateManual(taskID, spiderVersionID string) model.Execution {
	exec := model.Execution{
		ID:              uuid.NewString(),
		TaskID:          taskID,
		SpiderVersionID: spiderVersionID,
		Status:          "pending",
		TriggerSource:   "manual",
		CreatedAt:       time.Now(),
	}

	s.mu.Lock()
	s.executions[exec.ID] = exec
	s.logs[exec.ID] = []model.ExecutionLog{}
	s.mu.Unlock()

	return exec
}

func (s *ExecutionService) Get(id string) (model.Execution, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exec, ok := s.executions[id]
	if !ok {
		return model.Execution{}, false
	}

	exec.Logs = append([]model.ExecutionLog(nil), s.logs[id]...)
	return exec, true
}

func (s *ExecutionService) AppendLog(executionID, message string) (model.ExecutionLog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.executions[executionID]; !ok {
		return model.ExecutionLog{}, ErrExecutionNotFound
	}

	entry := model.ExecutionLog{
		ID:          uuid.NewString(),
		ExecutionID: executionID,
		Message:     message,
		CreatedAt:   time.Now(),
	}
	s.logs[executionID] = append(s.logs[executionID], entry)
	return entry, nil
}

func (s *ExecutionService) GetLogs(executionID string) ([]model.ExecutionLog, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.executions[executionID]; !ok {
		return nil, false
	}

	logs := make([]model.ExecutionLog, len(s.logs[executionID]))
	copy(logs, s.logs[executionID])
	return logs, true
}
