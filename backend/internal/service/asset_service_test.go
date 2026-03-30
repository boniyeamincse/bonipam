package service

import (
	"testing"

	"boni-pam/internal/domain"
)

func TestAssetService_CreateAsset(t *testing.T) {
	s := NewAssetService()

	t.Run("CreateSSHAssetSuccess", func(t *testing.T) {
		asset, err := s.CreateAsset(domain.CreateAssetRequest{
			Name: "prod-bastion",
			Type: "ssh",
			Host: "10.0.0.10",
			ConnectionMetadata: map[string]interface{}{
				"username":    "ec2-user",
				"auth_method": "key",
			},
		})
		if err != nil {
			t.Fatalf("CreateAsset returned error: %v", err)
		}
		if asset.Port != 22 {
			t.Fatalf("expected default ssh port 22, got %d", asset.Port)
		}
	})

	t.Run("RejectInvalidDatabaseMetadata", func(t *testing.T) {
		_, err := s.CreateAsset(domain.CreateAssetRequest{
			Name: "db01",
			Type: "database",
			Host: "db.local",
			ConnectionMetadata: map[string]interface{}{
				"engine": "postgres",
				// missing database and username
			},
		})
		if err == nil {
			t.Fatalf("expected validation error but got nil")
		}
	})

	t.Run("RejectUnsupportedType", func(t *testing.T) {
		_, err := s.CreateAsset(domain.CreateAssetRequest{
			Name: "redis01",
			Type: "cache",
			Host: "cache.local",
			ConnectionMetadata: map[string]interface{}{
				"username": "svc",
			},
		})
		if err == nil {
			t.Fatalf("expected unsupported type error but got nil")
		}
	})
}
