package service

import (
	"context"
	"testing"
	"time"

	"boni-pam/internal/domain"
)

func TestPolicyService_ValidatePolicy(t *testing.T) {
	s := NewPolicyService()
	ctx := context.Background()

	tests := []struct {
		name       string
		definition domain.PolicyDefinition
		wantErr    bool
	}{
		{
			name: "Valid Policy v1.0",
			definition: domain.PolicyDefinition{
				SchemaVersion: "1.0",
				Description:   "Test policy",
				DefaultEffect: "deny",
				Rules: []domain.PolicyRule{
					{
						Effect:    "allow",
						Subjects:  []string{"role:admin"},
						Resources: []string{"asset:server-1"},
						Actions:   []string{"ssh.connect"},
						Conditions: []domain.PolicyCondition{
							{
								Attribute: "time",
								Operator:  "between",
								Value:     []string{"09:00", "17:00"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Schema Version",
			definition: domain.PolicyDefinition{
				SchemaVersion: "2.0",
				DefaultEffect: "deny",
				Rules:         []domain.PolicyRule{},
			},
			wantErr: true,
		},
		{
			name: "Missing Required Fields",
			definition: domain.PolicyDefinition{
				SchemaVersion: "1.0",
				// Missing DefaultEffect and Rules
			},
			wantErr: true,
		},
		{
			name: "Invalid Rule (Missing Subject)",
			definition: domain.PolicyDefinition{
				SchemaVersion: "1.0",
				DefaultEffect: "deny",
				Rules: []domain.PolicyRule{
					{
						Effect:    "allow",
						Resources: []string{"*"},
						Actions:   []string{"*"},
						// Missing Subjects
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidatePolicy(ctx, tt.definition)
			if (err != nil) != tt.wantErr {
				t.Errorf("PolicyService.ValidatePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPolicyService_EvaluatePolicy(t *testing.T) {
	s := NewPolicyService()
	ctx := context.Background()

	created, err := s.CreatePolicy(ctx, domain.CreatePolicyRequest{
		Name: "runtime-eval",
		Definition: domain.PolicyDefinition{
			SchemaVersion: "1.0",
			DefaultEffect: "deny",
			Rules: []domain.PolicyRule{
				{
					ID:          "allow-admin-ssh",
					Effect:      "allow",
					Subjects:    []string{"role:admin"},
					Resources:   []string{"asset:*"},
					Actions:     []string{"ssh.connect"},
					Obligations: []string{"record_session"},
				},
				{
					ID:        "deny-high-risk",
					Effect:    "deny",
					Subjects:  []string{"role:*"},
					Resources: []string{"asset:*"},
					Actions:   []string{"ssh.connect"},
					Conditions: []domain.PolicyCondition{
						{Attribute: "risk_score", Operator: "gte", Value: 80},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreatePolicy returned error: %v", err)
	}

	t.Run("AllowOnMatchingRule", func(t *testing.T) {
		result, err := s.EvaluatePolicy(ctx, created.ID.String(), domain.PolicyEvaluationRequest{
			Subject:  "role:admin",
			Resource: "asset:server-1",
			Action:   "ssh.connect",
			Attributes: map[string]interface{}{
				"risk_score": 20,
			},
		})
		if err != nil {
			t.Fatalf("EvaluatePolicy returned error: %v", err)
		}
		if result.Decision != "allow" {
			t.Fatalf("expected allow decision, got %s", result.Decision)
		}
		if len(result.Obligations) != 1 || result.Obligations[0] != "record_session" {
			t.Fatalf("unexpected obligations: %#v", result.Obligations)
		}
	})

	t.Run("DenyOverridesAllow", func(t *testing.T) {
		result, err := s.EvaluatePolicy(ctx, created.ID.String(), domain.PolicyEvaluationRequest{
			Subject:  "role:admin",
			Resource: "asset:server-1",
			Action:   "ssh.connect",
			Attributes: map[string]interface{}{
				"risk_score": 90,
			},
		})
		if err != nil {
			t.Fatalf("EvaluatePolicy returned error: %v", err)
		}
		if result.Decision != "deny" {
			t.Fatalf("expected deny decision, got %s", result.Decision)
		}
		if len(result.MatchedRuleID) == 0 {
			t.Fatalf("expected matched rules, got none")
		}
	})

	t.Run("DefaultEffectWhenNoMatch", func(t *testing.T) {
		result, err := s.EvaluatePolicy(ctx, created.ID.String(), domain.PolicyEvaluationRequest{
			Subject:  "role:viewer",
			Resource: "asset:server-2",
			Action:   "vault.read",
		})
		if err != nil {
			t.Fatalf("EvaluatePolicy returned error: %v", err)
		}
		if result.Decision != "deny" {
			t.Fatalf("expected default deny decision, got %s", result.Decision)
		}
	})
}

func TestPolicyService_PublishAndRollback(t *testing.T) {
	s := NewPolicyService()
	ctx := context.Background()

	created, err := s.CreatePolicy(ctx, domain.CreatePolicyRequest{
		Name: "publish-rollback",
		Definition: domain.PolicyDefinition{
			SchemaVersion: "1.0",
			DefaultEffect: "deny",
			Rules: []domain.PolicyRule{
				{Effect: "allow", Subjects: []string{"role:admin"}, Resources: []string{"asset:*"}, Actions: []string{"ssh.connect"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreatePolicy returned error: %v", err)
	}

	publishResult, err := s.PublishPolicy(created.ID.String())
	if err != nil {
		t.Fatalf("PublishPolicy returned error: %v", err)
	}
	if publishResult.Status != domain.PolicyStatusPublished {
		t.Fatalf("expected published status, got %s", publishResult.Status)
	}

	updated, err := s.UpdatePolicy(ctx, created.ID.String(), domain.UpdatePolicyRequest{
		Definition: &domain.PolicyDefinition{
			SchemaVersion: "1.0",
			DefaultEffect: "allow",
			Rules: []domain.PolicyRule{
				{Effect: "allow", Subjects: []string{"role:*"}, Resources: []string{"asset:*"}, Actions: []string{"ssh.connect"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdatePolicy returned error: %v", err)
	}
	if updated.Status != domain.PolicyStatusDraft {
		t.Fatalf("expected draft status after update, got %s", updated.Status)
	}

	rolledBack, err := s.RollbackPolicy(ctx, created.ID.String(), 1)
	if err != nil {
		t.Fatalf("RollbackPolicy returned error: %v", err)
	}
	if rolledBack.Status != domain.PolicyStatusPublished {
		t.Fatalf("expected published status after rollback, got %s", rolledBack.Status)
	}
	if rolledBack.Definition.DefaultEffect != "deny" {
		t.Fatalf("expected rollback to version 1 definition, got default_effect=%s", rolledBack.Definition.DefaultEffect)
	}
}

func TestContextResolverService_Resolve(t *testing.T) {
	resolver := NewContextResolverService()
	risk := 72.5

	resolved := resolver.Resolve(domain.PolicyEvaluationRequest{
		SourceIP:    "10.0.0.5",
		DeviceID:    "device-123",
		DeviceTrust: "TRUSTED",
		RiskScore:   &risk,
		RequestTime: "2026-03-30T14:35:00Z",
		Attributes: map[string]interface{}{
			"custom_flag": true,
		},
	})

	if resolved["source_ip"] != "10.0.0.5" {
		t.Fatalf("expected source_ip to be resolved")
	}
	if resolved["device_trust"] != "trusted" {
		t.Fatalf("expected normalized device_trust")
	}
	if resolved["risk_score"] != 72.5 {
		t.Fatalf("expected risk_score to be resolved")
	}
	if resolved["time"] != "14:35" {
		t.Fatalf("expected time to be derived from request_time, got %v", resolved["time"])
	}
	if resolved["day_of_week"] != "monday" {
		t.Fatalf("expected day_of_week to be derived from request_time, got %v", resolved["day_of_week"])
	}
	if resolved["custom_flag"] != true {
		t.Fatalf("expected custom attributes to be preserved")
	}
}

func TestPolicyService_EvaluatePolicy_WithResolvedContext(t *testing.T) {
	s := NewPolicyService()
	ctx := context.Background()

	created, err := s.CreatePolicy(ctx, domain.CreatePolicyRequest{
		Name: "context-aware-eval",
		Definition: domain.PolicyDefinition{
			SchemaVersion: "1.0",
			DefaultEffect: "deny",
			Rules: []domain.PolicyRule{
				{
					Effect:    "allow",
					Subjects:  []string{"role:admin"},
					Resources: []string{"asset:*"},
					Actions:   []string{"ssh.connect"},
					Conditions: []domain.PolicyCondition{
						{Attribute: "device_trust", Operator: "eq", Value: "trusted"},
						{Attribute: "risk_score", Operator: "lt", Value: 70},
						{Attribute: "time", Operator: "between", Value: []string{"09:00", "17:00"}},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreatePolicy returned error: %v", err)
	}

	risk := 55.0
	result, err := s.EvaluatePolicy(ctx, created.ID.String(), domain.PolicyEvaluationRequest{
		Subject:     "role:admin",
		Resource:    "asset:server-1",
		Action:      "ssh.connect",
		SourceIP:    "10.1.1.9",
		DeviceTrust: "trusted",
		RiskScore:   &risk,
		RequestTime: time.Date(2026, time.March, 30, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("EvaluatePolicy returned error: %v", err)
	}
	if result.Decision != "allow" {
		t.Fatalf("expected allow decision, got %s", result.Decision)
	}
}
