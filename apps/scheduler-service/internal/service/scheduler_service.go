package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"crawler-platform/apps/scheduler-service/internal/model"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var ErrInvalidSchedule = errors.New("invalid schedule")

const maxCatchUpRunsPerPoll = 16

type SchedulerService struct {
	repo            Repository
	executionClient ExecutionClient
	parser          cron.Parser
	now             func() time.Time
}

type Repository interface {
	Create(ctx context.Context, schedule model.Schedule) error
	List(ctx context.Context) ([]model.Schedule, error)
	AdvanceLastMaterialized(ctx context.Context, id string, previous *time.Time, next time.Time) (bool, error)
	RestoreLastMaterialized(ctx context.Context, id string, previous *time.Time, current time.Time) error
}

type ExecutionClient interface {
	Create(ctx context.Context, input CreateExecutionInput) (string, error)
	MaterializeRetry(ctx context.Context) (bool, error)
}

type CreateExecutionInput struct {
	ScheduleID         string
	ProjectID          string
	SpiderID           string
	Image              string
	Command            []string
	TriggerSource      string
	ScheduledFor       time.Time
	RetryLimit         int
	RetryCount         int
	RetryDelaySeconds  int
}

type Option func(*SchedulerService)

type memoryRepository struct {
	mu        sync.Mutex
	schedules []model.Schedule
}

type noopExecutionClient struct{}

type HTTPExecutionClient struct {
	baseURL       string
	internalToken string
	client        *http.Client
}

func (r *memoryRepository) Create(_ context.Context, schedule model.Schedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schedules = append(r.schedules, schedule)
	return nil
}

func (r *memoryRepository) List(_ context.Context) ([]model.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	schedules := make([]model.Schedule, len(r.schedules))
	copy(schedules, r.schedules)
	return schedules, nil
}

func (r *memoryRepository) AdvanceLastMaterialized(_ context.Context, id string, previous *time.Time, next time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, schedule := range r.schedules {
		if schedule.ID != id {
			continue
		}
		if !timesEqual(schedule.LastMaterializedAt, previous) {
			return false, nil
		}
		r.schedules[i].LastMaterializedAt = &next
		return true, nil
	}
	return false, nil
}

func (r *memoryRepository) RestoreLastMaterialized(_ context.Context, id string, previous *time.Time, current time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, schedule := range r.schedules {
		if schedule.ID != id {
			continue
		}
		if schedule.LastMaterializedAt == nil || !schedule.LastMaterializedAt.Equal(current) {
			return nil
		}
		r.schedules[i].LastMaterializedAt = previous
		return nil
	}
	return nil
}

func (noopExecutionClient) Create(_ context.Context, _ CreateExecutionInput) (string, error) {
	return "", nil
}

func (noopExecutionClient) MaterializeRetry(_ context.Context) (bool, error) {
	return false, nil
}

func NewHTTPExecutionClient(baseURL, internalToken string) *HTTPExecutionClient {
	return &HTTPExecutionClient{
		baseURL:       baseURL,
		internalToken: internalToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *HTTPExecutionClient) Create(ctx context.Context, input CreateExecutionInput) (string, error) {
	body, err := json.Marshal(map[string]any{
		"projectId":     input.ProjectID,
		"spiderId":      input.SpiderID,
		"image":         input.Image,
		"command":       input.Command,
		"triggerSource": input.TriggerSource,
		"scheduleId":    input.ScheduleID,
		"scheduledFor":  input.ScheduledFor.Format(time.RFC3339),
		"retryLimit":    input.RetryLimit,
		"retryCount":    input.RetryCount,
		"retryDelaySeconds": input.RetryDelaySeconds,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/executions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("execution create returned status %d", resp.StatusCode)
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ID == "" {
		return "", errors.New("execution create returned empty id")
	}
	return payload.ID, nil
}

func (c *HTTPExecutionClient) MaterializeRetry(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/executions/retries/materialize", nil)
	if err != nil {
		return false, err
	}
	if c.internalToken != "" {
		req.Header.Set("X-Internal-Token", c.internalToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		return true, nil
	case http.StatusNoContent:
		return false, nil
	default:
		return false, fmt.Errorf("retry materialization returned status %d", resp.StatusCode)
	}
}

func WithNow(now func() time.Time) Option {
	return func(s *SchedulerService) {
		s.now = now
	}
}

func NewSchedulerService(repo Repository, executionClient ExecutionClient, options ...Option) *SchedulerService {
	if repo == nil {
		repo = &memoryRepository{}
	}
	if executionClient == nil {
		executionClient = noopExecutionClient{}
	}

	svc := &SchedulerService{
		repo:            repo,
		executionClient: executionClient,
		parser:          cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		now:             func() time.Time { return time.Now().UTC() },
	}

	for _, option := range options {
		if option != nil {
			option(svc)
		}
	}
	return svc
}

func (s *SchedulerService) Create(projectID, spiderID, name, cronExpr, image string, command []string, enabled bool, retryLimit, retryDelaySeconds int) (model.Schedule, error) {
	if projectID == "" || spiderID == "" || name == "" || cronExpr == "" || image == "" {
		return model.Schedule{}, ErrInvalidSchedule
	}

	createdAt := s.now().UTC()
	schedule := model.Schedule{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		SpiderID:  spiderID,
		Name:      name,
		CronExpr:  cronExpr,
		Enabled:   enabled,
		Image:     image,
		Command:   append([]string(nil), command...),
		RetryLimit: retryLimit,
		RetryDelaySeconds: retryDelaySeconds,
		CreatedAt: createdAt,
	}

	if err := s.repo.Create(context.Background(), schedule); err != nil {
		return model.Schedule{}, err
	}
	return schedule, nil
}

func (s *SchedulerService) List() ([]model.Schedule, error) {
	return s.repo.List(context.Background())
}

func (s *SchedulerService) MaterializeDue(ctx context.Context) (int, error) {
	schedules, err := s.repo.List(ctx)
	if err != nil {
		return 0, err
	}

	now := s.now().UTC().Truncate(time.Minute)
	materialized := 0

	for _, schedule := range schedules {
		if !schedule.Enabled {
			continue
		}

		spec, err := s.parser.Parse(schedule.CronExpr)
		if err != nil {
			return materialized, err
		}

		base := schedule.CreatedAt.UTC().Add(-time.Minute)
		if schedule.LastMaterializedAt != nil {
			base = schedule.LastMaterializedAt.UTC()
		}

		for catchUp := 0; catchUp < maxCatchUpRunsPerPoll; catchUp++ {
			next := spec.Next(base).UTC()
			if next.After(now) {
				break
			}

			previous := schedule.LastMaterializedAt
			claimed, err := s.repo.AdvanceLastMaterialized(ctx, schedule.ID, previous, next)
			if err != nil {
				return materialized, err
			}
			if !claimed {
				break
			}

			_, err = s.executionClient.Create(ctx, CreateExecutionInput{
				ScheduleID:        schedule.ID,
				ProjectID:         schedule.ProjectID,
				SpiderID:          schedule.SpiderID,
				Image:             schedule.Image,
				Command:           append([]string(nil), schedule.Command...),
				TriggerSource:     "scheduled",
				ScheduledFor:      next,
				RetryLimit:        schedule.RetryLimit,
				RetryCount:        0,
				RetryDelaySeconds: schedule.RetryDelaySeconds,
			})
			if err != nil {
				if rollbackErr := s.repo.RestoreLastMaterialized(ctx, schedule.ID, previous, next); rollbackErr != nil {
					return materialized, errors.Join(err, rollbackErr)
				}
				return materialized, err
			}

			schedule.LastMaterializedAt = &next
			base = next
			materialized++
		}
	}

	return materialized, nil
}

func (s *SchedulerService) Run(ctx context.Context, pollInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if _, err := s.MaterializeDue(ctx); err != nil {
			return err
		}
		for retryBatch := 0; retryBatch < maxCatchUpRunsPerPoll; retryBatch++ {
			materializedRetry, err := s.executionClient.MaterializeRetry(ctx)
			if err != nil {
				return err
			}
			if !materializedRetry {
				break
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func timesEqual(left, right *time.Time) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return left.Equal(*right)
	}
}
