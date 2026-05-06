// Package service 是 Monitor 服务的业务逻辑层。
// 核心职责：监控总览聚合、告警规则管理、告警评估（轮询）、Webhook 通知投递。
// 不直接访问数据库——该职责属于 repo 层。
package service

import (
	"context"
	"crawler-platform/apps/monitor-service/internal/model"
	"time"
)

type MonitorService struct {
	repo      Repository
	deliverer WebhookDeliverer
}

type WebhookDeliverer interface {
	Deliver(ctx context.Context, webhookURL string, payload []byte, timeout time.Duration) (int, error)
}

type Repository interface {
	Overview(ctx context.Context, orgID string) (model.Overview, error)
	CreateAlertRule(ctx context.Context, orgID string, rule model.AlertRule) (model.AlertRule, error)
	UpdateAlertRule(ctx context.Context, orgID, id string, patch model.AlertRulePatch) (model.AlertRule, bool, error)
	ListAlertRules(ctx context.Context, orgID string) ([]model.AlertRule, error)
	ListAlertEvents(ctx context.Context, orgID string, limit, offset int) ([]model.AlertEvent, error)
	CountAlertEvents(ctx context.Context, orgID string) (int64, error)
	ListFailedExecutionsSince(ctx context.Context, orgID string, since time.Time, limit int) ([]model.FailedExecutionCandidate, error)
	ListOfflineNodes(ctx context.Context, orgID string, threshold time.Time, limit int) ([]model.OfflineNodeCandidate, error)
	LastAlertEventAt(ctx context.Context, orgID, ruleID, dedupeKey string) (*time.Time, error)
	SaveAlertEvent(ctx context.Context, orgID string, event model.AlertEventRecord) error
}

type memoryRepository struct{}

func (r *memoryRepository) Overview(_ context.Context, _ string) (model.Overview, error) {
	return model.Overview{}, nil
}
func (r *memoryRepository) CreateAlertRule(_ context.Context, _ string, rule model.AlertRule) (model.AlertRule, error) {
	return rule, nil
}
func (r *memoryRepository) UpdateAlertRule(_ context.Context, _, _ string, patch model.AlertRulePatch) (model.AlertRule, bool, error) {
	return model.AlertRule{}, false, nil
}
func (r *memoryRepository) ListAlertRules(_ context.Context, _ string) ([]model.AlertRule, error) {
	return nil, nil
}
func (r *memoryRepository) ListAlertEvents(_ context.Context, _ string, _, _ int) ([]model.AlertEvent, error) {
	return nil, nil
}
func (r *memoryRepository) CountAlertEvents(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (r *memoryRepository) ListFailedExecutionsSince(_ context.Context, _ string, _ time.Time, _ int) ([]model.FailedExecutionCandidate, error) {
	return nil, nil
}
func (r *memoryRepository) ListOfflineNodes(_ context.Context, _ string, _ time.Time, _ int) ([]model.OfflineNodeCandidate, error) {
	return nil, nil
}
func (r *memoryRepository) LastAlertEventAt(_ context.Context, _, _, _ string) (*time.Time, error) {
	return nil, nil
}
func (r *memoryRepository) SaveAlertEvent(_ context.Context, _ string, _ model.AlertEventRecord) error {
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

func (s *MonitorService) Overview(orgID string) (model.Overview, error) {
	return s.repo.Overview(context.Background(), orgID)
}

func (s *MonitorService) UpdateAlertRule(orgID string, input UpdateAlertRuleInput) (model.AlertRule, error) {
	if input.ID == "" {
		return model.AlertRule{}, ErrInvalidRuleID
	}
	patch := model.AlertRulePatch{
		Name:                input.Name,
		Enabled:             input.Enabled,
		WebhookURL:          input.WebhookURL,
		CooldownSeconds:     input.CooldownSeconds,
		TimeoutSeconds:      input.TimeoutSeconds,
		OfflineGraceSeconds: input.OfflineGraceSeconds,
		UpdatedAt:           time.Now().UTC(),
	}
	rule, found, err := s.repo.UpdateAlertRule(context.Background(), orgID, input.ID, patch)
	if err != nil {
		return model.AlertRule{}, err
	}
	if !found {
		return model.AlertRule{}, ErrAlertRuleNotFound
	}
	return rule, nil
}

func (s *MonitorService) ListAlertRules(orgID string) ([]model.AlertRule, error) {
	return s.repo.ListAlertRules(context.Background(), orgID)
}

func (s *MonitorService) ListAlertEvents(orgID string, limit, offset int) ([]model.AlertEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.ListAlertEvents(context.Background(), orgID, limit, offset)
}
