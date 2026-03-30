package service

import (
	"boni-pam/internal/domain"
	"testing"
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
