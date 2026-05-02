package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"crawler-platform/apps/monitor-service/internal/model"
	"github.com/google/uuid"
)

type MonitorService struct {
	repo      Repository
	deliverer WebhookDeliverer
}

type Repository interface {
	Overview(ctx context.Context) (model.Overview, error)
	CreateAlertRule(ctx context.Context, rule model.AlertRule) (model.AlertRule, error)
	UpdateAlertRule(ctx context.Context, id string, patch model.AlertRulePatch) (model.AlertRule, bool, error)
	ListAlertRules(ctx context.Context) ([]model.AlertRule, error)
	ListAlertEvents(ctx context.Context, limit, offset int) ([]model.AlertEvent, error)
	CountAlertEvents(ctx context.Context) (int64, error)
	ListFailedExecutionsSince(ctx context.Context, since time.Time, limit int) ([]model.FailedExecutionCandidate, error)
	ListOfflineNodes(ctx context.Context, before time.Time, limit int) ([]model.OfflineNodeCandidate, error)
	LastAlertEventAt(ctx context.Context, ruleID, dedupeKey string) (*time.Time, error)
	SaveAlertEvent(ctx context.Context, event model.AlertEventRecord) error
}

type WebhookDeliverer interface {
	Deliver(ctx context.Context, webhookURL string, payload []byte, timeout time.Duration) (statusCode int, err error)
}

type memoryRepository struct{}

func (r *memoryRepository) Overview(_ context.Context) (model.Overview, error) {
	return model.Overview{}, nil
}
func (r *memoryRepository) CreateAlertRule(_ context.Context, rule model.AlertRule) (model.AlertRule, error) {
	return rule, nil
}
func (r *memoryRepository) UpdateAlertRule(_ context.Context, _ string, _ model.AlertRulePatch) (model.AlertRule, bool, error) {
	return model.AlertRule{}, false, nil
}
func (r *memoryRepository) ListAlertRules(_ context.Context) ([]model.AlertRule, error) {
	return nil, nil
}
func (r *memoryRepository) ListAlertEvents(_ context.Context, _ int, _ int) ([]model.AlertEvent, error) {
	return nil, nil
}
func (r *memoryRepository) CountAlertEvents(_ context.Context) (int64, error) { return 0, nil }
func (r *memoryRepository) ListFailedExecutionsSince(_ context.Context, _ time.Time, _ int) ([]model.FailedExecutionCandidate, error) {
	return nil, nil
}
func (r *memoryRepository) ListOfflineNodes(_ context.Context, _ time.Time, _ int) ([]model.OfflineNodeCandidate, error) {
	return nil, nil
}
func (r *memoryRepository) LastAlertEventAt(_ context.Context, _ string, _ string) (*time.Time, error) {
	return nil, nil
}
func (r *memoryRepository) SaveAlertEvent(_ context.Context, _ model.AlertEventRecord) error {
	return nil
}

func NewMonitorService(repos ...Repository) *MonitorService {
	svc := &MonitorService{deliverer: &httpWebhookDeliverer{}}
	if len(repos) > 0 && repos[0] != nil {
		svc.repo = repos[0]
		return svc
	}
	svc.repo = &memoryRepository{}
	return svc
}

func (s *MonitorService) WithWebhookDeliverer(deliverer WebhookDeliverer) *MonitorService {
	if deliverer != nil {
		s.deliverer = deliverer
	}
	return s
}

func (s *MonitorService) Overview() (model.Overview, error) {
	return s.repo.Overview(context.Background())
}

type CreateAlertRuleInput struct {
	Name                string
	RuleType            string
	Enabled             bool
	WebhookURL          string
	CooldownSeconds     int
	TimeoutSeconds      int
	OfflineGraceSeconds int
}

var (
	ErrInvalidRuleType   = errors.New("invalid alert rule type")
	ErrInvalidWebhookURL = errors.New("webhook url is required")
	ErrInvalidRuleName   = errors.New("rule name is required")
	ErrRuleNotFound      = errors.New("alert rule not found")
)

func (s *MonitorService) CreateAlertRule(input CreateAlertRuleInput) (model.AlertRule, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.AlertRule{}, ErrInvalidRuleName
	}
	ruleType := strings.TrimSpace(input.RuleType)
	if ruleType != model.AlertRuleTypeExecutionFailed && ruleType != model.AlertRuleTypeNodeOffline {
		return model.AlertRule{}, ErrInvalidRuleType
	}
	webhookURL := strings.TrimSpace(input.WebhookURL)
	if webhookURL == "" {
		return model.AlertRule{}, ErrInvalidWebhookURL
	}

	now := time.Now().UTC()
	rule := model.AlertRule{
		ID:                  uuid.NewString(),
		Name:                name,
		RuleType:            ruleType,
		Enabled:             input.Enabled,
		WebhookURL:          webhookURL,
		CooldownSeconds:     normalizeCooldown(input.CooldownSeconds, ruleType),
		TimeoutSeconds:      normalizeTimeout(input.TimeoutSeconds, ruleType),
		OfflineGraceSeconds: normalizeOfflineGrace(input.OfflineGraceSeconds),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	return s.repo.CreateAlertRule(context.Background(), rule)
}

func (s *MonitorService) ListAlertRules() ([]model.AlertRule, error) {
	return s.repo.ListAlertRules(context.Background())
}

type UpdateAlertRuleInput struct {
	Name                *string
	Enabled             *bool
	WebhookURL          *string
	CooldownSeconds     *int
	TimeoutSeconds      *int
	OfflineGraceSeconds *int
}

func (s *MonitorService) UpdateAlertRule(id string, input UpdateAlertRuleInput) (model.AlertRule, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.AlertRule{}, ErrRuleNotFound
	}

	patch := model.AlertRulePatch{
		Enabled:             input.Enabled,
		CooldownSeconds:     input.CooldownSeconds,
		TimeoutSeconds:      input.TimeoutSeconds,
		OfflineGraceSeconds: input.OfflineGraceSeconds,
		UpdatedAt:           time.Now().UTC(),
	}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return model.AlertRule{}, ErrInvalidRuleName
		}
		patch.Name = &trimmed
	}
	if input.WebhookURL != nil {
		trimmed := strings.TrimSpace(*input.WebhookURL)
		if trimmed == "" {
			return model.AlertRule{}, ErrInvalidWebhookURL
		}
		patch.WebhookURL = &trimmed
	}
	if patch.CooldownSeconds != nil && *patch.CooldownSeconds <= 0 {
		return model.AlertRule{}, errors.New("cooldownSeconds must be positive")
	}
	if patch.TimeoutSeconds != nil && *patch.TimeoutSeconds <= 0 {
		return model.AlertRule{}, errors.New("timeoutSeconds must be positive")
	}
	if patch.OfflineGraceSeconds != nil && *patch.OfflineGraceSeconds <= 0 {
		return model.AlertRule{}, errors.New("offlineGraceSeconds must be positive")
	}

	updated, ok, err := s.repo.UpdateAlertRule(context.Background(), id, patch)
	if err != nil {
		return model.AlertRule{}, err
	}
	if !ok {
		return model.AlertRule{}, ErrRuleNotFound
	}
	return updated, nil
}

func (s *MonitorService) ListAlertEvents(limit, offset int) ([]model.AlertEvent, int64, int, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.repo.ListAlertEvents(context.Background(), limit, offset)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	total, err := s.repo.CountAlertEvents(context.Background())
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, limit, offset, nil
}
