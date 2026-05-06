package service

// Package service 的告警评估与 Webhook 投递逻辑。
// 核心流程：轮询告警规则 → 查询候选实体（失败执行/离线节点）→
// 冷却检查 → 构建 payload → Webhook 投递 → 事件持久化。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"crawler-platform/apps/monitor-service/internal/model"

	"github.com/google/uuid"
)

const (
	alertBatchLimit              = 100
	defaultPollPeriod            = 15 * time.Second
	defaultNodeOfflinePollPeriod = 5 * time.Second
)

var (
	ErrInvalidRuleName   = errors.New("invalid rule name")
	ErrInvalidRuleType   = errors.New("invalid rule type")
	ErrInvalidWebhookURL = errors.New("invalid webhook url")
	ErrInvalidRuleID     = errors.New("invalid rule id")
	ErrAlertRuleNotFound = errors.New("alert rule not found")
)

type CreateAlertRuleInput struct {
	Name                string
	RuleType            string
	Enabled             bool
	WebhookURL          string
	CooldownSeconds     int
	TimeoutSeconds      int
	OfflineGraceSeconds int
}

type UpdateAlertRuleInput struct {
	ID                  string
	Name                *string
	Enabled             *bool
	WebhookURL          *string
	CooldownSeconds     *int
	TimeoutSeconds      *int
	OfflineGraceSeconds *int
}

type webhookPayload struct {
	RuleID     string      `json:"ruleId"`
	RuleName   string      `json:"ruleName"`
	RuleType   string      `json:"ruleType"`
	EntityType string      `json:"entityType"`
	EntityID   string      `json:"entityId"`
	OccurredAt time.Time   `json:"occurredAt"`
	Data       interface{} `json:"data"`
}

type httpWebhookDeliverer struct{}

func (d *httpWebhookDeliverer) Deliver(ctx context.Context, webhookURL string, payload []byte, timeout time.Duration) (int, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func (s *MonitorService) StartAlertLoop(ctx context.Context, interval time.Duration) {
	s.StartAlertLoops(ctx, interval, interval)
}

func (s *MonitorService) StartAlertLoops(ctx context.Context, executionFailedInterval, nodeOfflineInterval time.Duration) {
	if executionFailedInterval <= 0 {
		executionFailedInterval = defaultPollPeriod
	}
	if nodeOfflineInterval <= 0 {
		nodeOfflineInterval = defaultNodeOfflinePollPeriod
	}
	if executionFailedInterval == nodeOfflineInterval {
		s.startAlertTicker(ctx, executionFailedInterval, func(loopCtx context.Context) error {
			return s.EvaluateAlerts(loopCtx)
		})
		return
	}

	s.startAlertTicker(ctx, executionFailedInterval, func(loopCtx context.Context) error {
		return s.EvaluateAlertsByRuleTypes(loopCtx, model.AlertRuleTypeExecutionFailed)
	})
	s.startAlertTicker(ctx, nodeOfflineInterval, func(loopCtx context.Context) error {
		return s.EvaluateAlertsByRuleTypes(loopCtx, model.AlertRuleTypeNodeOffline)
	})
}

func (s *MonitorService) startAlertTicker(ctx context.Context, interval time.Duration, evaluate func(context.Context) error) {
	if interval <= 0 {
		interval = defaultPollPeriod
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			_ = evaluate(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// CreateAlertRule 校验并创建告警规则，对未设置的超时/冷却参数取默认值。
// 节点离线类的默认超时 3s、冷却 60s；执行失败类默认超时 5s、冷却 120s。
func (s *MonitorService) CreateAlertRule(orgID string, input CreateAlertRuleInput) (model.AlertRule, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.AlertRule{}, ErrInvalidRuleName
	}
	if input.RuleType != model.AlertRuleTypeExecutionFailed && input.RuleType != model.AlertRuleTypeNodeOffline {
		return model.AlertRule{}, ErrInvalidRuleType
	}
	webhookURL := strings.TrimSpace(input.WebhookURL)
	if webhookURL == "" {
		return model.AlertRule{}, ErrInvalidWebhookURL
	}

	now := time.Now().UTC()
	rule := model.AlertRule{
		ID:                  uuid.NewString(),
		OrganizationID:      orgID,
		Name:                name,
		RuleType:            input.RuleType,
		Enabled:             input.Enabled,
		WebhookURL:          webhookURL,
		CooldownSeconds:     normalizeCooldown(input.CooldownSeconds, input.RuleType),
		TimeoutSeconds:      normalizeTimeout(input.TimeoutSeconds, input.RuleType),
		OfflineGraceSeconds: normalizeOfflineGrace(input.OfflineGraceSeconds),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	return s.repo.CreateAlertRule(context.Background(), orgID, rule)
}

func (s *MonitorService) EvaluateAlerts(ctx context.Context) error {
	return s.EvaluateAlertsByRuleTypes(ctx)
}

// EvaluateAlertsByRuleTypes 遍历已启用告警规则，按类型分流评估。
// 可选参数 ruleTypes 为空时评估所有类型。
func (s *MonitorService) EvaluateAlertsByRuleTypes(ctx context.Context, ruleTypes ...string) error {
	rules, err := s.repo.ListAlertRules(ctx, "")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	allowedRuleTypes := map[string]struct{}{}
	for _, ruleType := range ruleTypes {
		allowedRuleTypes[ruleType] = struct{}{}
	}
	filterByType := len(allowedRuleTypes) > 0

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if filterByType {
			if _, ok := allowedRuleTypes[rule.RuleType]; !ok {
				continue
			}
		}
		switch rule.RuleType {
		case model.AlertRuleTypeExecutionFailed:
			if err := s.evaluateExecutionFailedRule(ctx, rule, now); err != nil {
				return err
			}
		case model.AlertRuleTypeNodeOffline:
			if err := s.evaluateNodeOfflineRule(ctx, rule, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MonitorService) evaluateExecutionFailedRule(ctx context.Context, rule model.AlertRule, now time.Time) error {
	candidates, err := s.repo.ListFailedExecutionsSince(ctx, "", rule.CreatedAt, alertBatchLimit)
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		dedupeKey := "execution:" + candidate.ExecutionID
		if !s.shouldNotify(ctx, rule, dedupeKey, now) {
			continue
		}
		payload := webhookPayload{
			RuleID:     rule.ID,
			RuleName:   rule.Name,
			RuleType:   rule.RuleType,
			EntityType: "execution",
			EntityID:   candidate.ExecutionID,
			OccurredAt: candidate.OccurredAt,
			Data: map[string]string{
				"projectId":    candidate.ProjectID,
				"spiderId":     candidate.SpiderID,
				"errorMessage": candidate.Error,
			},
		}
		if err := s.emitEvent(ctx, rule, dedupeKey, payload, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *MonitorService) evaluateNodeOfflineRule(ctx context.Context, rule model.AlertRule, now time.Time) error {
	threshold := now.Add(-time.Duration(rule.OfflineGraceSeconds) * time.Second)
	candidates, err := s.repo.ListOfflineNodes(ctx, "", threshold, alertBatchLimit)
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		dedupeKey := "node:" + candidate.NodeID
		if !s.shouldNotify(ctx, rule, dedupeKey, now) {
			continue
		}
		payload := webhookPayload{
			RuleID:     rule.ID,
			RuleName:   rule.Name,
			RuleType:   rule.RuleType,
			EntityType: "node",
			EntityID:   candidate.NodeID,
			OccurredAt: now,
			Data: map[string]string{
				"nodeName":   candidate.NodeName,
				"lastSeenAt": candidate.LastSeenAt.Format(time.RFC3339),
			},
		}
		if err := s.emitEvent(ctx, rule, dedupeKey, payload, now); err != nil {
			return err
		}
	}
	return nil
}

// shouldNotify 判断是否应发送告警：若无历史事件则发送；若上条事件在冷却期内则跳过。
func (s *MonitorService) shouldNotify(ctx context.Context, rule model.AlertRule, dedupeKey string, now time.Time) bool {
	last, err := s.repo.LastAlertEventAt(ctx, "", rule.ID, dedupeKey)
	if err != nil {
		return false
	}
	if last == nil {
		return true
	}
	cooldown := time.Duration(rule.CooldownSeconds) * time.Second
	return last.Add(cooldown).Before(now)
}

// emitEvent 投递 Webhook 并持久化告警事件记录。投递失败时 status 标记为 "failed"。
func (s *MonitorService) emitEvent(ctx context.Context, rule model.AlertRule, dedupeKey string, payload webhookPayload, now time.Time) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	statusCode, deliverErr := s.deliverer.Deliver(ctx, rule.WebhookURL, body, time.Duration(rule.TimeoutSeconds)*time.Second)
	status := "sent"
	errMsg := ""
	if deliverErr != nil {
		status = "failed"
		errMsg = deliverErr.Error()
	}

	record := model.AlertEventRecord{
		ID:                uuid.NewString(),
		RuleID:            rule.ID,
		RuleType:          rule.RuleType,
		EntityType:        payload.EntityType,
		EntityID:          payload.EntityID,
		DedupeKey:         dedupeKey,
		Payload:           string(body),
		DeliveryStatus:    status,
		WebhookStatusCode: statusCode,
		ErrorMessage:      errMsg,
		CreatedAt:         now,
	}
	if err := s.repo.SaveAlertEvent(ctx, "", record); err != nil {
		return err
	}
	return nil
}

func normalizeCooldown(input int, ruleType string) int {
	if input > 0 {
		return input
	}
	if ruleType == model.AlertRuleTypeNodeOffline {
		return 60
	}
	return 120
}

func normalizeTimeout(input int, ruleType string) int {
	if input > 0 {
		return input
	}
	if ruleType == model.AlertRuleTypeNodeOffline {
		return 3
	}
	return 5
}

func normalizeOfflineGrace(input int) int {
	if input > 0 {
		return input
	}
	return 60
}

var _ interface {
	Deliver(ctx context.Context, webhookURL string, payload []byte, timeout time.Duration) (int, error)
} = (*httpWebhookDeliverer)(nil)
