package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"crawler-platform/apps/execution-service/internal/model"
	"github.com/google/uuid"
)

var ErrExecutionNotFound = errors.New("execution not found")

type ExecutionService struct {
	execRepo ExecutionRepository
	logRepo  LogRepository
	queue    Queue
}

type ExecutionRepository interface {
	Create(ctx context.Context, exec model.Execution) (model.Execution, error)
	Get(ctx context.Context, id string) (model.Execution, error)
	Delete(ctx context.Context, id string) error
}

type LogRepository interface {
	Init(ctx context.Context, executionID string) error
	Append(ctx context.Context, entry model.ExecutionLog) error
	List(ctx context.Context, executionID string) ([]model.ExecutionLog, error)
}

type Queue interface {
	Enqueue(ctx context.Context, executionID string) error
}

type CreateManualInput struct {
	ProjectID string
	SpiderID  string
	Image     string
	Command   []string
}

func NewExecutionService(execRepo ExecutionRepository, logRepo LogRepository, queue Queue) *ExecutionService {
	return &ExecutionService{execRepo: execRepo, logRepo: logRepo, queue: queue}
}

func (s *ExecutionService) CreateManual(ctx context.Context, input CreateManualInput) (model.Execution, error) {
	exec := model.Execution{
		ID:            uuid.NewString(),
		ProjectID:     input.ProjectID,
		SpiderID:      input.SpiderID,
		Status:        "pending",
		TriggerSource: "manual",
		Image:         input.Image,
		Command:       append([]string(nil), input.Command...),
		CreatedAt:     time.Now().UTC(),
	}

	created, err := s.execRepo.Create(ctx, exec)
	if err != nil {
		return model.Execution{}, err
	}
	if err := s.logRepo.Init(ctx, created.ID); err != nil {
		return model.Execution{}, s.rollbackCreate(ctx, created.ID, err)
	}
	if err := s.queue.Enqueue(ctx, created.ID); err != nil {
		return model.Execution{}, s.rollbackCreate(ctx, created.ID, err)
	}

	return created, nil
}

func (s *ExecutionService) Get(ctx context.Context, id string) (model.Execution, error) {
	exec, err := s.execRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		return model.Execution{}, err
	}

	logs, err := s.logRepo.List(ctx, id)
	if err != nil {
		return model.Execution{}, err
	}
	exec.Logs = logs
	return exec, nil
}

func (s *ExecutionService) AppendLog(ctx context.Context, executionID, message string) (model.ExecutionLog, error) {
	if _, err := s.execRepo.Get(ctx, executionID); errors.Is(err, ErrExecutionNotFound) {
		return model.ExecutionLog{}, ErrExecutionNotFound
	} else if err != nil {
		return model.ExecutionLog{}, err
	}

	entry := model.ExecutionLog{
		ID:          uuid.NewString(),
		ExecutionID: executionID,
		Message:     message,
		CreatedAt:   time.Now().UTC(),
	}
	return entry, s.logRepo.Append(ctx, entry)
}

func (s *ExecutionService) GetLogs(ctx context.Context, executionID string) ([]model.ExecutionLog, error) {
	if _, err := s.execRepo.Get(ctx, executionID); err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return nil, ErrExecutionNotFound
		}
		return nil, err
	}

	logs, err := s.logRepo.List(ctx, executionID)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *ExecutionService) rollbackCreate(ctx context.Context, executionID string, cause error) error {
	if deleteErr := s.execRepo.Delete(ctx, executionID); deleteErr != nil {
		return errors.Join(cause, fmt.Errorf("rollback execution %s: %w", executionID, deleteErr))
	}
	return cause
}
