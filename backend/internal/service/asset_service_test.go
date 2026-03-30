package service

import (
	"testing"

	"boni-pam/internal/domain"
)

func TestAssetService_CreateAsset(t *testing.T) {
	s := NewAssetService()

	t.Run("CreateSSHAssetSuccess", func(t *testing.T) {
		asset, err := s.CreateAsset(domain.CreateAssetRequest{
			Name:        "prod-bastion",
			Type:        "ssh",
			Host:        "10.0.0.10",
			Environment: "prod",
			Owner:       "platform-team",
			Criticality: "high",
			Groups:      []string{"linux", "bastion"},
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
		if asset.Environment != "prod" || asset.Criticality != "high" {
			t.Fatalf("expected normalized tags to be persisted")
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

	t.Run("RejectInvalidTagging", func(t *testing.T) {
		_, err := s.CreateAsset(domain.CreateAssetRequest{
			Name:        "bad-tagged-asset",
			Type:        "ssh",
			Host:        "10.0.0.11",
			Environment: "qa",
			ConnectionMetadata: map[string]interface{}{
				"username":    "ec2-user",
				"auth_method": "password",
			},
		})
		if err == nil {
			t.Fatalf("expected invalid environment error but got nil")
		}
	})
}

func TestAssetService_UpdateTaggingAndFilter(t *testing.T) {
	s := NewAssetService()

	asset, err := s.CreateAsset(domain.CreateAssetRequest{
		Name: "app-db",
		Type: "database",
		Host: "db.internal",
		ConnectionMetadata: map[string]interface{}{
			"engine":   "postgres",
			"database": "app",
			"username": "app_user",
		},
	})
	if err != nil {
		t.Fatalf("CreateAsset returned error: %v", err)
	}

	updated, err := s.UpdateAssetTagging(asset.ID, domain.UpdateAssetTaggingRequest{
		Environment: "staging",
		Owner:       "data-team",
		Criticality: "critical",
		Groups:      []string{"databases", "tier1", "databases"},
	})
	if err != nil {
		t.Fatalf("UpdateAssetTagging returned error: %v", err)
	}

	if updated.Environment != "staging" || updated.Criticality != "critical" {
		t.Fatalf("expected updated tagging fields")
	}
	if len(updated.Groups) != 2 {
		t.Fatalf("expected deduped groups, got %#v", updated.Groups)
	}

	filtered := s.ListAssets("staging", "data-team", "critical", "tier1")
	if len(filtered) != 1 {
		t.Fatalf("expected one filtered asset, got %d", len(filtered))
	}
}
