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

func TestAssetService_TestConnection(t *testing.T) {
	s := NewAssetService()

	sshAsset, err := s.CreateAsset(domain.CreateAssetRequest{
		Name: "ssh-node-1",
		Type: "ssh",
		Host: "ssh.internal.local",
		ConnectionMetadata: map[string]interface{}{
			"username":    "ops",
			"auth_method": "key",
		},
	})
	if err != nil {
		t.Fatalf("CreateAsset(ssh) returned error: %v", err)
	}

	t.Run("Success", func(t *testing.T) {
		result, err := s.TestConnection(sshAsset.ID, 5)
		if err != nil {
			t.Fatalf("TestConnection returned error: %v", err)
		}
		if result.Status != "ok" {
			t.Fatalf("expected status ok, got %s", result.Status)
		}
		if result.Protocol != "ssh" {
			t.Fatalf("expected protocol ssh, got %s", result.Protocol)
		}
		if result.TimeoutUsedS != 5 {
			t.Fatalf("expected timeout 5, got %d", result.TimeoutUsedS)
		}
	})

	t.Run("FailureByHostPattern", func(t *testing.T) {
		badAsset, err := s.CreateAsset(domain.CreateAssetRequest{
			Name: "db-bad",
			Type: "database",
			Host: "unreachable-db.local",
			ConnectionMetadata: map[string]interface{}{
				"engine":   "postgres",
				"database": "app",
				"username": "app_user",
			},
		})
		if err != nil {
			t.Fatalf("CreateAsset(database) returned error: %v", err)
		}

		result, err := s.TestConnection(badAsset.ID, 3)
		if err != nil {
			t.Fatalf("TestConnection returned error: %v", err)
		}
		if result.Status != "failed" {
			t.Fatalf("expected status failed, got %s", result.Status)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := s.TestConnection("ast-missing", 1)
		if err == nil {
			t.Fatalf("expected error for unknown asset")
		}
	})
}

func TestAssetService_OwnershipWorkflow(t *testing.T) {
	s := NewAssetService()

	asset, err := s.CreateAsset(domain.CreateAssetRequest{
		Name: "owner-test-node",
		Type: "ssh",
		Host: "owner.internal.local",
		ConnectionMetadata: map[string]interface{}{
			"username":    "ops",
			"auth_method": "key",
		},
	})
	if err != nil {
		t.Fatalf("CreateAsset returned error: %v", err)
	}

	assigned, err := s.AssignOwner(asset.ID, domain.AssignAssetOwnerRequest{
		Owner:      "alice",
		AssignedBy: "admin",
		Reason:     "initial ownership",
	})
	if err != nil {
		t.Fatalf("AssignOwner returned error: %v", err)
	}
	if assigned.Owner != "alice" {
		t.Fatalf("expected owner alice, got %s", assigned.Owner)
	}

	transfer, err := s.RequestOwnershipTransfer(asset.ID, domain.RequestAssetOwnershipTransferRequest{
		NewOwner:    "bob",
		RequestedBy: "alice",
		Reason:      "team rotation",
	})
	if err != nil {
		t.Fatalf("RequestOwnershipTransfer returned error: %v", err)
	}
	if transfer.Status != domain.AssetTransferStatusPending {
		t.Fatalf("expected pending transfer, got %s", transfer.Status)
	}

	reviewed, err := s.ReviewOwnershipTransfer(transfer.ID, domain.ReviewAssetOwnershipTransferRequest{
		Approved:   true,
		ReviewedBy: "manager-1",
		Comment:    "approved",
	})
	if err != nil {
		t.Fatalf("ReviewOwnershipTransfer returned error: %v", err)
	}
	if reviewed.Status != domain.AssetTransferStatusApproved {
		t.Fatalf("expected approved transfer, got %s", reviewed.Status)
	}

	updatedAsset, err := s.GetAsset(asset.ID)
	if err != nil {
		t.Fatalf("GetAsset returned error: %v", err)
	}
	if updatedAsset.Owner != "bob" {
		t.Fatalf("expected owner bob after approval, got %s", updatedAsset.Owner)
	}

	transfers := s.ListOwnershipTransfers(asset.ID)
	if len(transfers) != 1 {
		t.Fatalf("expected one ownership transfer, got %d", len(transfers))
	}

	events := s.ListAssetAuditEvents(asset.ID)
	if len(events) < 3 {
		t.Fatalf("expected at least 3 audit events, got %d", len(events))
	}
}
