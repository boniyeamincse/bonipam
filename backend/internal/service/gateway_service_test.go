package service

import (
	"boni-pam/internal/domain"
	"testing"
)

func TestGatewayService_InitiateSession(t *testing.T) {
	s := NewGatewayService()

	resp, err := s.InitiateSession(domain.InitiateSessionRequest{
		UserID:        "user-001",
		TargetAssetID: "asset-bastion-01",
		Protocol:      "ssh",
		Metadata: map[string]string{
			"env": "prod",
		},
	})
	if err != nil {
		t.Fatalf("InitiateSession returned error: %v", err)
	}
	if resp.SessionID == "" {
		t.Fatalf("expected session_id to be set")
	}
	if resp.Status != "active" {
		t.Fatalf("expected status active, got %s", resp.Status)
	}
	if resp.Protocol != "ssh" {
		t.Fatalf("expected protocol ssh, got %s", resp.Protocol)
	}
	if resp.ActivatedAt == nil {
		t.Fatalf("expected activated_at to be set")
	}
}

func TestGatewayService_InitiateSession_UnsupportedProtocol(t *testing.T) {
	s := NewGatewayService()
	_, err := s.InitiateSession(domain.InitiateSessionRequest{
		UserID:        "user-001",
		TargetAssetID: "asset-01",
		Protocol:      "ftp",
	})
	if err == nil {
		t.Fatalf("expected unsupported protocol error")
	}
}

func TestGatewayService_GetSession(t *testing.T) {
	s := NewGatewayService()

	created, err := s.InitiateSession(domain.InitiateSessionRequest{
		UserID:        "user-002",
		TargetAssetID: "asset-rdp-01",
		Protocol:      "rdp",
	})
	if err != nil {
		t.Fatalf("InitiateSession returned error: %v", err)
	}

	got, err := s.GetSession(created.SessionID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if got.SessionID != created.SessionID {
		t.Fatalf("expected session id to match")
	}

	_, err = s.GetSession("nonexistent-gw-id")
	if err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestGatewayService_TerminateSession(t *testing.T) {
	s := NewGatewayService()

	created, err := s.InitiateSession(domain.InitiateSessionRequest{
		UserID:        "user-003",
		TargetAssetID: "asset-http-01",
		Protocol:      "http",
	})
	if err != nil {
		t.Fatalf("InitiateSession returned error: %v", err)
	}

	err = s.TerminateSession(created.SessionID, "user disconnected")
	if err != nil {
		t.Fatalf("TerminateSession returned error: %v", err)
	}

	got, err := s.GetSession(created.SessionID)
	if err != nil {
		t.Fatalf("GetSession after terminate returned error: %v", err)
	}
	if got.Status != "terminated" {
		t.Fatalf("expected status terminated, got %s", got.Status)
	}
	if got.TerminatedAt == nil {
		t.Fatalf("expected terminated_at to be set")
	}
	if got.TerminateReason != "user disconnected" {
		t.Fatalf("expected terminate reason to match, got %q", got.TerminateReason)
	}

	// Terminating an already terminated session should fail
	err = s.TerminateSession(created.SessionID, "again")
	if err == nil {
		t.Fatalf("expected already terminated error")
	}
}
