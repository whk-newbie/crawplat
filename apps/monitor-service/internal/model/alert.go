package model

import "time"

const (
	AlertRuleTypeExecutionFailed = "execution_failed"
	AlertRuleTypeNodeOffline     = "node_offline"
)

type AlertRule struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	RuleType            string    `json:"ruleType"`
	Enabled             bool      `json:"enabled"`
	WebhookURL          string    `json:"webhookUrl"`
	CooldownSeconds     int       `json:"cooldownSeconds"`
	TimeoutSeconds      int       `json:"timeoutSeconds"`
	OfflineGraceSeconds int       `json:"offlineGraceSeconds"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type AlertRulePatch struct {
	Name                *string
	Enabled             *bool
	WebhookURL          *string
	CooldownSeconds     *int
	TimeoutSeconds      *int
	OfflineGraceSeconds *int
	UpdatedAt           time.Time
}

type AlertEvent struct {
	ID                string    `json:"id"`
	RuleID            string    `json:"ruleId"`
	RuleType          string    `json:"ruleType"`
	EntityType        string    `json:"entityType"`
	EntityID          string    `json:"entityId"`
	DedupeKey         string    `json:"dedupeKey"`
	Payload           string    `json:"payload"`
	DeliveryStatus    string    `json:"deliveryStatus"`
	WebhookStatusCode int       `json:"webhookStatusCode,omitempty"`
	ErrorMessage      string    `json:"errorMessage,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
}

type AlertEventRecord struct {
	ID                string
	RuleID            string
	RuleType          string
	EntityType        string
	EntityID          string
	DedupeKey         string
	Payload           string
	DeliveryStatus    string
	WebhookStatusCode int
	ErrorMessage      string
	CreatedAt         time.Time
}

type FailedExecutionCandidate struct {
	ExecutionID string
	ProjectID   string
	SpiderID    string
	Error       string
	OccurredAt  time.Time
}

type OfflineNodeCandidate struct {
	NodeID     string
	NodeName   string
	LastSeenAt time.Time
}
