package service

import (
	"boni-pam/internal/domain"
	"fmt"

	"github.com/google/uuid"
)

type GatewayService struct {
	gatewayHost string
}

func NewGatewayService(gatewayHost string) *GatewayService {
	if gatewayHost == "" {
		gatewayHost = "gw-01.bonipam.internal"
	}
	return &GatewayService{gatewayHost: gatewayHost}
}

func (s *GatewayService) StartSession(req domain.StartSessionRequest) (domain.GatewaySession, error) {
	if req.UserID == "" || req.AssetID == "" || req.JITGrantID == "" {
		return domain.GatewaySession{}, fmt.Errorf("invalid start session request")
	}

	sessionID := "s-" + uuid.NewString()
	return domain.GatewaySession{
		SessionID:      sessionID,
		GatewayHost:    s.gatewayHost,
		ConnectCommand: fmt.Sprintf("ssh -p 3022 %s@%s", sessionID, s.gatewayHost),
		Status:         "active",
	}, nil
}

func (s *GatewayService) TerminateSession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	return nil
}

func (s *GatewayService) SessionStatus(sessionID string) (domain.GatewaySession, error) {
	if sessionID == "" {
		return domain.GatewaySession{}, fmt.Errorf("session_id is required")
	}
	return domain.GatewaySession{
		SessionID:      sessionID,
		GatewayHost:    s.gatewayHost,
		ConnectCommand: fmt.Sprintf("ssh -p 3022 %s@%s", sessionID, s.gatewayHost),
		Status:         "active",
	}, nil
}
