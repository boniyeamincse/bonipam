package service

import (
	"boni-pam/internal/domain"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GatewayService struct {
	mu       sync.RWMutex
	sessions map[string]domain.GatewaySession
}

func NewGatewayService() *GatewayService {
	return &GatewayService{
		sessions: make(map[string]domain.GatewaySession),
	}
}

// InitiateSession creates a new gateway session in active state.
func (s *GatewayService) InitiateSession(req domain.InitiateSessionRequest) (domain.GatewaySessionResponse, error) {
	userID := strings.TrimSpace(req.UserID)
	assetID := strings.TrimSpace(req.TargetAssetID)
	protocol := strings.ToLower(strings.TrimSpace(req.Protocol))
	if userID == "" || assetID == "" || protocol == "" {
		return domain.GatewaySessionResponse{}, fmt.Errorf("user_id, target_asset_id, and protocol are required")
	}
	if !domain.SupportedGatewayProtocols[protocol] {
		return domain.GatewaySessionResponse{}, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	now := time.Now().UTC()
	session := domain.GatewaySession{
		SessionID:     "gw-" + uuid.NewString(),
		UserID:        userID,
		TargetAssetID: assetID,
		Protocol:      protocol,
		Status:        "active",
		Metadata:      req.Metadata,
		CreatedAt:     now,
		ActivatedAt:   &now,
	}

	s.mu.Lock()
	s.sessions[session.SessionID] = session
	s.mu.Unlock()

	return toGatewaySessionResponse(session), nil
}

// GetSession returns the current state of a gateway session.
func (s *GatewayService) GetSession(sessionID string) (domain.GatewaySessionResponse, error) {
	s.mu.RLock()
	session, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	if !ok {
		return domain.GatewaySessionResponse{}, fmt.Errorf("session not found")
	}
	return toGatewaySessionResponse(session), nil
}

// TerminateSession marks a session as terminated.
func (s *GatewayService) TerminateSession(sessionID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}
	if session.Status == "terminated" {
		return fmt.Errorf("session already terminated")
	}
	now := time.Now().UTC()
	session.Status = "terminated"
	session.TerminatedAt = &now
	session.TerminateReason = strings.TrimSpace(reason)
	s.sessions[sessionID] = session
	return nil
}

// StartSession kept for backward compatibility with earlier stubs.
func (s *GatewayService) StartSession(req domain.StartSessionRequest) (domain.GatewaySessionResponse, error) {
	return s.InitiateSession(domain.InitiateSessionRequest{
		UserID:        req.UserID,
		TargetAssetID: req.AssetID,
		Protocol:      req.Protocol,
	})
}

func toGatewaySessionResponse(gs domain.GatewaySession) domain.GatewaySessionResponse {
	return domain.GatewaySessionResponse{
		SessionID:       gs.SessionID,
		UserID:          gs.UserID,
		TargetAssetID:   gs.TargetAssetID,
		Protocol:        gs.Protocol,
		Status:          gs.Status,
		Metadata:        gs.Metadata,
		CreatedAt:       gs.CreatedAt,
		ActivatedAt:     gs.ActivatedAt,
		TerminatedAt:    gs.TerminatedAt,
		TerminateReason: gs.TerminateReason,
	}
}
