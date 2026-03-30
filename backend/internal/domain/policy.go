package domain

import (
	"time"

	"github.com/google/uuid"
)

// PolicyStatus defines the lifecycle state of a policy
type PolicyStatus string

const (
	PolicyStatusDraft     PolicyStatus = "draft"
	PolicyStatusPublished PolicyStatus = "published"
	PolicyStatusArchived  PolicyStatus = "archived"
)

// Policy represents an access control policy in Boni PAM
type Policy struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Version     int              `json:"version"`
	Status      PolicyStatus     `json:"status"`
	Definition  PolicyDefinition `json:"definition"`
	CreatedBy   uuid.UUID        `json:"created_by"`
	CreatedAt   time.Time        `json:"created_at"`
	PublishedAt *time.Time       `json:"published_at,omitempty"`
}

// PolicyDefinition contains the actual logic of the policy
type PolicyDefinition struct {
	SchemaVersion string       `json:"schema_version"`
	Description   string       `json:"description,omitempty"`
	Rules         []PolicyRule `json:"rules"`
	DefaultEffect string       `json:"default_effect"` // "allow" or "deny"
}

// PolicyRule defines a single permission/restriction atom
type PolicyRule struct {
	ID          string            `json:"id,omitempty"`
	Effect      string            `json:"effect"`    // "allow" or "deny"
	Subjects    []string          `json:"subjects"`  // Role names or user IDs (supports wildcards e.g. "role:*")
	Resources   []string          `json:"resources"` // Asset IDs or tags (supports wildcards)
	Actions     []string          `json:"actions"`   // e.g. "ssh.connect", "vault.read"
	Conditions  []PolicyCondition `json:"conditions,omitempty"`
	Obligations []string          `json:"obligations,omitempty"` // e.g. "record_session", "require_mfa"
}

// PolicyCondition defines requirements for a rule to apply
type PolicyCondition struct {
	Attribute string      `json:"attribute"` // e.g. "time", "source_ip", "risk_score"
	Operator  string      `json:"operator"`  // e.g. "between", "in", "gte"
	Value     interface{} `json:"value"`
}

type CreatePolicyRequest struct {
	Name       string           `json:"name" binding:"required"`
	Definition PolicyDefinition `json:"definition" binding:"required"`
	CreatedBy  string           `json:"created_by"`
}

type UpdatePolicyRequest struct {
	Name       string            `json:"name"`
	Definition *PolicyDefinition `json:"definition"`
}

type PolicyEvaluationRequest struct {
	Subject    string                 `json:"subject" binding:"required"`
	Resource   string                 `json:"resource" binding:"required"`
	Action     string                 `json:"action" binding:"required"`
	Attributes map[string]interface{} `json:"attributes"`
}

type PolicyEvaluationResponse struct {
	PolicyID      uuid.UUID `json:"policy_id"`
	Decision      string    `json:"decision"`
	Obligations   []string  `json:"obligations,omitempty"`
	MatchedRuleID []string  `json:"matched_rule_ids,omitempty"`
}
