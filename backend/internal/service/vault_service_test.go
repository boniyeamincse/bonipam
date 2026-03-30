package service

import (
	"boni-pam/internal/domain"
	"testing"
	"time"
)

func TestVaultService_StoreAndGetSecret(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	stored, err := s.StoreSecret(domain.CreateSecretRequest{
		Name:  "db-password",
		Value: "super-secret-value",
		Metadata: map[string]string{
			"env": "dev",
		},
	})
	if err != nil {
		t.Fatalf("StoreSecret returned error: %v", err)
	}

	got, err := s.GetSecret(stored.ID)
	if err != nil {
		t.Fatalf("GetSecret returned error: %v", err)
	}
	if got.Value != "super-secret-value" {
		t.Fatalf("expected decrypted value to match, got %q", got.Value)
	}
	if got.KEKVersion != "v1" {
		t.Fatalf("expected kek version v1, got %s", got.KEKVersion)
	}
}

func TestVaultService_NameUniqueness(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	_, err = s.StoreSecret(domain.CreateSecretRequest{Name: "api-key", Value: "abc"})
	if err != nil {
		t.Fatalf("first store failed: %v", err)
	}

	_, err = s.StoreSecret(domain.CreateSecretRequest{Name: "API-KEY", Value: "def"})
	if err == nil {
		t.Fatalf("expected duplicate name error")
	}
}

func TestVaultService_IssueCredential(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	issued, err := s.IssueCredential(domain.IssueCredentialRequest{
		TargetType: "database",
		TargetID:   "asset-db-01",
		Role:       "readonly",
		TTLSeconds: 600,
		Metadata: map[string]string{
			"env": "dev",
		},
	})
	if err != nil {
		t.Fatalf("IssueCredential returned error: %v", err)
	}
	if issued.LeaseID == "" {
		t.Fatalf("expected lease id to be set")
	}
	if issued.Username == "" || issued.Password == "" {
		t.Fatalf("expected generated username and password")
	}
	if issued.LeaseSeconds != 600 {
		t.Fatalf("expected lease seconds 600, got %d", issued.LeaseSeconds)
	}
	if !issued.ExpiresAt.After(issued.IssuedAt) {
		t.Fatalf("expected expires_at to be after issued_at")
	}
	actualTTL := issued.ExpiresAt.Sub(issued.IssuedAt)
	if actualTTL < 599*time.Second || actualTTL > 601*time.Second {
		t.Fatalf("expected ttl close to 600s, got %v", actualTTL)
	}
}

func TestVaultService_IssueCredentialValidation(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	_, err = s.IssueCredential(domain.IssueCredentialRequest{
		TargetType: "k8s",
		TargetID:   "asset-01",
		Role:       "admin",
	})
	if err == nil {
		t.Fatalf("expected unsupported target type error")
	}

	_, err = s.IssueCredential(domain.IssueCredentialRequest{
		TargetType: "ssh",
		TargetID:   "asset-01",
		Role:       "operator",
		TTLSeconds: 30,
	})
	if err == nil {
		t.Fatalf("expected ttl validation error")
	}
}

func TestVaultService_CreateRotationPolicy(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	policy, err := s.CreateRotationPolicy(domain.CreateRotationPolicyRequest{
		TargetType:      "database",
		TargetID:        "asset-db-01",
		Role:            "readonly",
		IntervalSeconds: 3600,
		TTLSeconds:      600,
	})
	if err != nil {
		t.Fatalf("CreateRotationPolicy returned error: %v", err)
	}
	if policy.PolicyID == "" {
		t.Fatalf("expected policy id to be set")
	}
	if !policy.Enabled {
		t.Fatalf("expected policy to be enabled")
	}
	if policy.IntervalSeconds != 3600 {
		t.Fatalf("expected interval 3600, got %d", policy.IntervalSeconds)
	}

	// Duplicate should fail
	_, err = s.CreateRotationPolicy(domain.CreateRotationPolicyRequest{
		TargetType:      "database",
		TargetID:        "asset-db-01",
		Role:            "readonly",
		IntervalSeconds: 3600,
	})
	if err == nil {
		t.Fatalf("expected duplicate rotation policy error")
	}
}

func TestVaultService_GetRotationPolicy(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	created, err := s.CreateRotationPolicy(domain.CreateRotationPolicyRequest{
		TargetType:      "ssh",
		TargetID:        "bastion-01",
		Role:            "operator",
		IntervalSeconds: 7200,
	})
	if err != nil {
		t.Fatalf("CreateRotationPolicy returned error: %v", err)
	}

	got, err := s.GetRotationPolicy(created.PolicyID)
	if err != nil {
		t.Fatalf("GetRotationPolicy returned error: %v", err)
	}
	if got.PolicyID != created.PolicyID {
		t.Fatalf("expected policy id to match")
	}

	_, err = s.GetRotationPolicy("nonexistent-policy-id")
	if err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestVaultService_TriggerRotation(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	policy, err := s.CreateRotationPolicy(domain.CreateRotationPolicyRequest{
		TargetType:      "api",
		TargetID:        "svc-api-01",
		Role:            "consumer",
		IntervalSeconds: 3600,
		TTLSeconds:      300,
	})
	if err != nil {
		t.Fatalf("CreateRotationPolicy returned error: %v", err)
	}

	result, err := s.TriggerRotation(policy.PolicyID)
	if err != nil {
		t.Fatalf("TriggerRotation returned error: %v", err)
	}
	if result.LeaseID == "" || result.Username == "" {
		t.Fatalf("expected lease and username to be set after rotation")
	}
	if result.RotatedAt.IsZero() {
		t.Fatalf("expected rotated_at to be set")
	}

	// Policy should reflect last rotated timestamp
	updated, err := s.GetRotationPolicy(policy.PolicyID)
	if err != nil {
		t.Fatalf("GetRotationPolicy after trigger returned error: %v", err)
	}
	if updated.LastRotatedAt == nil || updated.LastRotatedAt.IsZero() {
		t.Fatalf("expected last_rotated_at to be set after trigger")
	}
}

func TestVaultService_GetLeaseStatus(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	issued, err := s.IssueCredential(domain.IssueCredentialRequest{
		TargetType: "database",
		TargetID:   "db-status-01",
		Role:       "readonly",
		TTLSeconds: 300,
	})
	if err != nil {
		t.Fatalf("IssueCredential returned error: %v", err)
	}

	st, err := s.GetLeaseStatus(issued.LeaseID)
	if err != nil {
		t.Fatalf("GetLeaseStatus returned error: %v", err)
	}
	if st.Status != "active" {
		t.Fatalf("expected status active, got %s", st.Status)
	}
	if st.Revoked {
		t.Fatalf("expected lease to not be revoked")
	}

	_, err = s.GetLeaseStatus("nonexistent-lease-id")
	if err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestVaultService_RevokeLease(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	issued, err := s.IssueCredential(domain.IssueCredentialRequest{
		TargetType: "ssh",
		TargetID:   "bastion-revoke-01",
		Role:       "operator",
		TTLSeconds: 300,
	})
	if err != nil {
		t.Fatalf("IssueCredential returned error: %v", err)
	}

	err = s.RevokeLease(issued.LeaseID, "session terminated")
	if err != nil {
		t.Fatalf("RevokeLease returned error: %v", err)
	}

	st, err := s.GetLeaseStatus(issued.LeaseID)
	if err != nil {
		t.Fatalf("GetLeaseStatus after revoke returned error: %v", err)
	}
	if st.Status != "revoked" {
		t.Fatalf("expected status revoked, got %s", st.Status)
	}
	if !st.Revoked {
		t.Fatalf("expected revoked to be true")
	}
	if st.RevokedAt == nil || st.RevokedAt.IsZero() {
		t.Fatalf("expected revoked_at to be set")
	}
	if st.RevokeReason != "session terminated" {
		t.Fatalf("expected revoke reason to match, got %q", st.RevokeReason)
	}

	// Revoking an already revoked lease should fail
	err = s.RevokeLease(issued.LeaseID, "again")
	if err == nil {
		t.Fatalf("expected already revoked error")
	}
}

func TestVaultService_RevokeLeasesByTarget(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	target := "bulk-target-01"
	for i := 0; i < 3; i++ {
		_, err = s.IssueCredential(domain.IssueCredentialRequest{
			TargetType: "database",
			TargetID:   target,
			Role:       "readonly",
			TTLSeconds: 300,
		})
		if err != nil {
			t.Fatalf("IssueCredential[%d] returned error: %v", i, err)
		}
	}

	result, err := s.RevokeLeasesByTarget(target, "policy trigger")
	if err != nil {
		t.Fatalf("RevokeLeasesByTarget returned error: %v", err)
	}
	if result.Revoked != 3 {
		t.Fatalf("expected 3 leases revoked, got %d", result.Revoked)
	}
	if result.TargetID != target {
		t.Fatalf("expected target id to match")
	}

	// Second call should revoke 0 (all already revoked)
	result2, err := s.RevokeLeasesByTarget(target, "duplicate")
	if err != nil {
		t.Fatalf("second RevokeLeasesByTarget returned error: %v", err)
	}
	if result2.Revoked != 0 {
		t.Fatalf("expected 0 additional revocations, got %d", result2.Revoked)
	}
}

// Suppress unused import warning for time package (used in other tests).
var _ = time.Now
