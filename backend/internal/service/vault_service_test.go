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
