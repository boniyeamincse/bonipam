package service

import (
	"context"
	"testing"

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
