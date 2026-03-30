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
