package service

import (
	"boni-pam/internal/domain"
	"testing"
)

func TestLocalKMSProvider_WrapUnwrapRoundTrip(t *testing.T) {
	p, err := NewLocalKMSProvider("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewLocalKMSProvider returned error: %v", err)
	}

	original := []byte("this-is-a-32-byte-data-encryption-k")
	wrapped, err := p.WrapKey(original)
	if err != nil {
		t.Fatalf("WrapKey returned error: %v", err)
	}
	if len(wrapped) != len(original) {
		t.Fatalf("expected wrapped length %d, got %d", len(original), len(wrapped))
	}

	unwrapped, err := p.UnwrapKey(wrapped)
	if err != nil {
		t.Fatalf("UnwrapKey returned error: %v", err)
	}
	for i := range original {
		if original[i] != unwrapped[i] {
			t.Fatalf("unwrapped key mismatch at index %d", i)
		}
	}
}

func TestLocalKMSProvider_Describe(t *testing.T) {
	p, err := NewLocalKMSProvider("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewLocalKMSProvider returned error: %v", err)
	}

	info := p.Describe()
	if info.Provider != "local" {
		t.Fatalf("expected provider=local, got %s", info.Provider)
	}
	if info.KeyID == "" {
		t.Fatalf("expected key_id to be set")
	}
	if info.KEKVersion == "" {
		t.Fatalf("expected kek_version to be set")
	}
}

func TestNewLocalKMSProvider_Validation(t *testing.T) {
	_, err := NewLocalKMSProvider("")
	if err == nil {
		t.Fatalf("expected error for empty master key")
	}

	_, err = NewLocalKMSProvider("short")
	if err == nil {
		t.Fatalf("expected error for short master key")
	}
}

func TestRemoteKMSProvider_ReturnsNotImplemented(t *testing.T) {
	p := NewRemoteKMSProvider("aws-kms", "https://kms.us-east-1.amazonaws.com", "arn:aws:kms:us-east-1:123:key/abc")

	_, err := p.WrapKey([]byte("somekey"))
	if err == nil {
		t.Fatalf("expected ErrKMSNotImplemented from WrapKey")
	}

	_, err = p.UnwrapKey([]byte("somewrapped"))
	if err == nil {
		t.Fatalf("expected ErrKMSNotImplemented from UnwrapKey")
	}

	info := p.Describe()
	if info.Provider != "aws-kms" {
		t.Fatalf("expected provider=aws-kms, got %s", info.Provider)
	}
	if info.Endpoint == "" {
		t.Fatalf("expected endpoint to be set")
	}
}

func TestVaultService_UsesKMSProvider(t *testing.T) {
	s, err := NewVaultService("test-master-key-123456")
	if err != nil {
		t.Fatalf("NewVaultService returned error: %v", err)
	}

	info := s.KMSInfo()
	if info.Provider != "local" {
		t.Fatalf("expected local KMS provider, got %s", info.Provider)
	}

	// Verify secrets still work end-to-end via KMS adapter
	stored, err := s.StoreSecret(domain.CreateSecretRequest{Name: "kms-test-key", Value: "kms-secret-val"})
	if err != nil {
		t.Fatalf("StoreSecret via KMS adapter returned error: %v", err)
	}
	got, err := s.GetSecret(stored.ID)
	if err != nil {
		t.Fatalf("GetSecret via KMS adapter returned error: %v", err)
	}
	if got.Value != "kms-secret-val" {
		t.Fatalf("decrypted value mismatch, got %q", got.Value)
	}
}
