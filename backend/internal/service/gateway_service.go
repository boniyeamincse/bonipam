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
	mu        sync.RWMutex
	sessions  map[string]domain.GatewaySession
	jitGrants map[string]domain.JITGrant
}

func NewGatewayService() *GatewayService {
	return &GatewayService{
		sessions:  make(map[string]domain.GatewaySession),
		jitGrants: make(map[string]domain.JITGrant),
	}
}

// InitiateSession creates a new gateway session in active state.
// If JITGrantID is set in the request, a valid approved grant is required.
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

	grantID := strings.TrimSpace(req.JITGrantID)
	if grantID != "" {
		if err := s.validateJITGrant(grantID, userID, assetID, protocol); err != nil {
			return domain.GatewaySessionResponse{}, err
		}
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
		JITGrantID:    req.JITGrantID,
	})
}

// RequestJITGrant creates a JIT grant in pending state awaiting approval.
func (s *GatewayService) RequestJITGrant(req domain.RequestJITGrantRequest) (domain.JITGrantResponse, error) {
	userID := strings.TrimSpace(req.UserID)
	assetID := strings.TrimSpace(req.AssetID)
	protocol := strings.ToLower(strings.TrimSpace(req.Protocol))
	reason := strings.TrimSpace(req.Reason)
	if userID == "" || assetID == "" || protocol == "" || reason == "" {
		return domain.JITGrantResponse{}, fmt.Errorf("user_id, asset_id, protocol, and reason are required")
	}
	if !domain.SupportedGatewayProtocols[protocol] {
		return domain.JITGrantResponse{}, fmt.Errorf("unsupported protocol: %s", protocol)
	}
	ttl := req.TTLSeconds
	if ttl == 0 {
		ttl = 3600
	}
	if ttl < 60 || ttl > 86400 {
		return domain.JITGrantResponse{}, fmt.Errorf("ttl_seconds must be between 60 and 86400")
	}

	now := time.Now().UTC()
	grant := domain.JITGrant{
		GrantID:    "jit-" + uuid.NewString(),
		UserID:     userID,
		AssetID:    assetID,
		Protocol:   protocol,
		Reason:     reason,
		Status:     "pending",
		TTLSeconds: ttl,
		CreatedAt:  now,
		ExpiresAt:  now.Add(time.Duration(ttl) * time.Second),
	}

	s.mu.Lock()
	s.jitGrants[grant.GrantID] = grant
	s.mu.Unlock()

	return toJITGrantResponse(grant), nil
}

// GetJITGrant returns a JIT grant by ID.
func (s *GatewayService) GetJITGrant(grantID string) (domain.JITGrantResponse, error) {
	s.mu.RLock()
	grant, ok := s.jitGrants[grantID]
	s.mu.RUnlock()
	if !ok {
		return domain.JITGrantResponse{}, fmt.Errorf("jit grant not found")
	}
	return toJITGrantResponse(grant), nil
}

// ApproveJITGrant marks a pending grant as approved.
func (s *GatewayService) ApproveJITGrant(grantID, approverID string) (domain.JITGrantResponse, error) {
	approverID = strings.TrimSpace(approverID)
	if approverID == "" {
		return domain.JITGrantResponse{}, fmt.Errorf("approver_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	grant, ok := s.jitGrants[grantID]
	if !ok {
		return domain.JITGrantResponse{}, fmt.Errorf("jit grant not found")
	}
	if grant.Status != "pending" {
		return domain.JITGrantResponse{}, fmt.Errorf("grant is not in pending state")
	}
	if time.Now().UTC().After(grant.ExpiresAt) {
		return domain.JITGrantResponse{}, fmt.Errorf("grant has expired")
	}
	now := time.Now().UTC()
	grant.Status = "approved"
	grant.ApprovedAt = &now
	grant.ApproverID = approverID
	s.jitGrants[grantID] = grant
	return toJITGrantResponse(grant), nil
}

// DenyJITGrant marks a pending grant as denied.
func (s *GatewayService) DenyJITGrant(grantID, approverID, reason string) (domain.JITGrantResponse, error) {
	approverID = strings.TrimSpace(approverID)
	if approverID == "" {
		return domain.JITGrantResponse{}, fmt.Errorf("approver_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	grant, ok := s.jitGrants[grantID]
	if !ok {
		return domain.JITGrantResponse{}, fmt.Errorf("jit grant not found")
	}
	if grant.Status != "pending" {
		return domain.JITGrantResponse{}, fmt.Errorf("grant is not in pending state")
	}
	grant.Status = "denied"
	grant.ApproverID = approverID
	grant.DenyReason = strings.TrimSpace(reason)
	s.jitGrants[grantID] = grant
	return toJITGrantResponse(grant), nil
}

// validateJITGrant checks that a grant is approved, not expired, and matches session parameters.
func (s *GatewayService) validateJITGrant(grantID, userID, assetID, protocol string) error {
	s.mu.RLock()
	grant, ok := s.jitGrants[grantID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("jit grant not found")
	}
	if grant.Status != "approved" {
		return fmt.Errorf("jit grant is not approved (status: %s)", grant.Status)
	}
	if time.Now().UTC().After(grant.ExpiresAt) {
		return fmt.Errorf("jit grant has expired")
	}
	if !strings.EqualFold(grant.UserID, userID) {
		return fmt.Errorf("jit grant user mismatch")
	}
	if !strings.EqualFold(grant.AssetID, assetID) {
		return fmt.Errorf("jit grant asset mismatch")
	}
	if !strings.EqualFold(grant.Protocol, protocol) {
		return fmt.Errorf("jit grant protocol mismatch")
	}
	return nil
}

func toJITGrantResponse(g domain.JITGrant) domain.JITGrantResponse {
	return domain.JITGrantResponse{
		GrantID:    g.GrantID,
		UserID:     g.UserID,
		AssetID:    g.AssetID,
		Protocol:   g.Protocol,
		Reason:     g.Reason,
		Status:     g.Status,
		TTLSeconds: g.TTLSeconds,
		CreatedAt:  g.CreatedAt,
		ExpiresAt:  g.ExpiresAt,
		ApprovedAt: g.ApprovedAt,
		ApproverID: g.ApproverID,
		DenyReason: g.DenyReason,
	}
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
