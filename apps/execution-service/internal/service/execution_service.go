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
var ErrInvalidExecutionState = errors.New("invalid execution state transition")

type ExecutionService struct {
	execRepo ExecutionRepository
	logRepo  LogRepository
	queue    Queue
}

type ExecutionRepository interface {
	Create(ctx context.Context, exec model.Execution) (model.Execution, error)
	Get(ctx context.Context, id string) (model.Execution, error)
	MarkRunning(ctx context.Context, id, nodeID string, startedAt time.Time) (model.Execution, error)
	Complete(ctx context.Context, id string, finishedAt time.Time) (model.Execution, error)
	Fail(ctx context.Context, id, errorMessage string, finishedAt time.Time) (model.Execution, error)
	ClaimNextRetryCandidate(ctx context.Context, now time.Time) (model.Execution, bool, error)
	ResetRetryClaim(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

type LogRepository interface {
	Init(ctx context.Context, executionID string) error
	Append(ctx context.Context, entry model.ExecutionLog) error
	List(ctx context.Context, executionID string) ([]model.ExecutionLog, error)
}

type Queue interface {
	Enqueue(ctx context.Context, executionID string) error
	Claim(ctx context.Context) (string, error)
	Ack(ctx context.Context, executionID string) error
	Release(ctx context.Context, executionID string) error
}

type CreateManualInput struct {
	ProjectID      string
	SpiderID       string
	SpiderVersion  int
	Image          string
	Command        []string
	CPUCores       float64
	MemoryMB       int
	TimeoutSeconds int
}

type CreateExecutionInput struct {
	ProjectID          string
	SpiderID           string
	SpiderVersion      int
	Image              string
	Command            []string
	CPUCores           float64
	MemoryMB           int
	TimeoutSeconds     int
	TriggerSource      string
	RetryLimit         int
	RetryCount         int
	RetryDelaySeconds  int
	RetryOfExecutionID string
}

func NewExecutionService(execRepo ExecutionRepository, logRepo LogRepository, queue Queue) *ExecutionService {
	return &ExecutionService{execRepo: execRepo, logRepo: logRepo, queue: queue}
}

func (s *ExecutionService) CreateManual(ctx context.Context, input CreateManualInput) (model.Execution, error) {
	return s.Create(ctx, CreateExecutionInput{
		ProjectID:      input.ProjectID,
		SpiderID:       input.SpiderID,
		SpiderVersion:  input.SpiderVersion,
		Image:          input.Image,
		Command:        input.Command,
		CPUCores:       input.CPUCores,
		MemoryMB:       input.MemoryMB,
		TimeoutSeconds: input.TimeoutSeconds,
		TriggerSource:  "manual",
	})
}

func (s *ExecutionService) Create(ctx context.Context, input CreateExecutionInput) (model.Execution, error) {
	triggerSource := input.TriggerSource
	if triggerSource == "" {
		triggerSource = "manual"
	}

	exec := model.Execution{
		ID:                 uuid.NewString(),
		ProjectID:          input.ProjectID,
		SpiderID:           input.SpiderID,
		SpiderVersion:      input.SpiderVersion,
		Status:             "pending",
		TriggerSource:      triggerSource,
		Image:              input.Image,
		Command:            append([]string(nil), input.Command...),
		CPUCores:           input.CPUCores,
		MemoryMB:           input.MemoryMB,
		TimeoutSeconds:     input.TimeoutSeconds,
		RetryLimit:         input.RetryLimit,
		RetryCount:         input.RetryCount,
		RetryDelaySeconds:  input.RetryDelaySeconds,
		RetryOfExecutionID: input.RetryOfExecutionID,
		CreatedAt:          time.Now().UTC(),
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

func (s *ExecutionService) MaterializeRetry(ctx context.Context) (model.Execution, bool, error) {
	candidate, ok, err := s.execRepo.ClaimNextRetryCandidate(ctx, time.Now().UTC())
	if err != nil || !ok {
		return model.Execution{}, ok, err
	}

	created, err := s.Create(ctx, CreateExecutionInput{
		ProjectID:          candidate.ProjectID,
		SpiderID:           candidate.SpiderID,
		SpiderVersion:      candidate.SpiderVersion,
		Image:              candidate.Image,
		Command:            candidate.Command,
		CPUCores:           candidate.CPUCores,
		MemoryMB:           candidate.MemoryMB,
		TimeoutSeconds:     candidate.TimeoutSeconds,
		TriggerSource:      "retry",
		RetryLimit:         candidate.RetryLimit,
		RetryCount:         candidate.RetryCount + 1,
		RetryDelaySeconds:  candidate.RetryDelaySeconds,
		RetryOfExecutionID: candidate.ID,
	})
	if err != nil {
		if resetErr := s.execRepo.ResetRetryClaim(ctx, candidate.ID); resetErr != nil {
			return model.Execution{}, false, errors.Join(err, resetErr)
		}
		return model.Execution{}, false, err
	}

	return created, true, nil
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

func (s *ExecutionService) ClaimNext(ctx context.Context, nodeID string) (model.Execution, bool, error) {
	for {
		executionID, err := s.queue.Claim(ctx)
		if err != nil {
			return model.Execution{}, false, err
		}
		if executionID == "" {
			return model.Execution{}, false, nil
		}

		exec, err := s.transitionToRunning(ctx, executionID, nodeID, time.Now().UTC())
		if err != nil {
			if errors.Is(err, ErrExecutionNotFound) || errors.Is(err, ErrInvalidExecutionState) {
				if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
					return model.Execution{}, false, errors.Join(err, fmt.Errorf("ack claimed execution %s: %w", executionID, ackErr))
				}
				continue
			}
			if releaseErr := s.queue.Release(ctx, executionID); releaseErr != nil {
				return model.Execution{}, false, errors.Join(err, fmt.Errorf("release execution %s: %w", executionID, releaseErr))
			}
			return model.Execution{}, false, err
		}
		return exec, true, nil
	}
}

func (s *ExecutionService) Start(ctx context.Context, executionID, nodeID string) (model.Execution, error) {
	exec, err := s.execRepo.Get(ctx, executionID)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		return model.Execution{}, err
	}
	if exec.Status != "running" {
		return model.Execution{}, ErrInvalidExecutionState
	}
	if exec.NodeID != "" && exec.NodeID != nodeID {
		return model.Execution{}, ErrInvalidExecutionState
	}
	return exec, nil
}

func (s *ExecutionService) Complete(ctx context.Context, executionID string) (model.Execution, error) {
	exec, err := s.execRepo.Complete(ctx, executionID, time.Now().UTC())
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		if errors.Is(err, ErrInvalidExecutionState) {
			current, currentErr := s.execRepo.Get(ctx, executionID)
			if currentErr != nil {
				if errors.Is(currentErr, ErrExecutionNotFound) {
					return model.Execution{}, ErrExecutionNotFound
				}
				return model.Execution{}, currentErr
			}
			if current.Status == "succeeded" {
				if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
					return model.Execution{}, fmt.Errorf("ack completed execution %s: %w", executionID, ackErr)
				}
				return current, nil
			}
			return model.Execution{}, ErrInvalidExecutionState
		}
		return model.Execution{}, err
	}
	if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
		return model.Execution{}, fmt.Errorf("ack completed execution %s: %w", executionID, ackErr)
	}
	return exec, nil
}

func (s *ExecutionService) Fail(ctx context.Context, executionID, errorMessage string) (model.Execution, error) {
	exec, err := s.execRepo.Fail(ctx, executionID, errorMessage, time.Now().UTC())
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		if errors.Is(err, ErrInvalidExecutionState) {
			current, currentErr := s.execRepo.Get(ctx, executionID)
			if currentErr != nil {
				if errors.Is(currentErr, ErrExecutionNotFound) {
					return model.Execution{}, ErrExecutionNotFound
				}
				return model.Execution{}, currentErr
			}
			if current.Status == "failed" {
				if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
					return model.Execution{}, fmt.Errorf("ack failed execution %s: %w", executionID, ackErr)
				}
				return current, nil
			}
			return model.Execution{}, ErrInvalidExecutionState
		}
		return model.Execution{}, err
	}
	if ackErr := s.queue.Ack(ctx, executionID); ackErr != nil {
		return model.Execution{}, fmt.Errorf("ack failed execution %s: %w", executionID, ackErr)
	}
	return exec, nil
}

func (s *ExecutionService) rollbackCreate(ctx context.Context, executionID string, cause error) error {
	if deleteErr := s.execRepo.Delete(ctx, executionID); deleteErr != nil {
		return errors.Join(cause, fmt.Errorf("rollback execution %s: %w", executionID, deleteErr))
	}
	return cause
}

func (s *ExecutionService) transitionToRunning(ctx context.Context, executionID, nodeID string, startedAt time.Time) (model.Execution, error) {
	current, err := s.execRepo.Get(ctx, executionID)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		return model.Execution{}, err
	}
	if current.Status == "running" {
		return current, nil
	}
	if current.Status != "pending" {
		return model.Execution{}, ErrInvalidExecutionState
	}

	exec, err := s.execRepo.MarkRunning(ctx, executionID, nodeID, startedAt)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return model.Execution{}, ErrExecutionNotFound
		}
		if errors.Is(err, ErrInvalidExecutionState) {
			return model.Execution{}, ErrInvalidExecutionState
		}
		return model.Execution{}, err
	}
	return exec, nil
}
