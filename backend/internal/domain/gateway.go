package domain

import "time"

// SupportedGatewayProtocols lists the protocols handled by the SSH proxy core.
var SupportedGatewayProtocols = map[string]bool{
	"ssh":  true,
	"rdp":  true,
	"http": true,
}

type GatewaySession struct {
	SessionID       string            `json:"session_id"`
	UserID          string            `json:"user_id"`
	TargetAssetID   string            `json:"target_asset_id"`
	Protocol        string            `json:"protocol"` // ssh | rdp | http
	Status          string            `json:"status"`   // pending | active | terminated
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	ActivatedAt     *time.Time        `json:"activated_at,omitempty"`
	TerminatedAt    *time.Time        `json:"terminated_at,omitempty"`
	TerminateReason string            `json:"terminate_reason,omitempty"`
}

type InitiateSessionRequest struct {
	UserID        string            `json:"user_id" binding:"required"`
	TargetAssetID string            `json:"target_asset_id" binding:"required"`
	Protocol      string            `json:"protocol" binding:"required"`
	Metadata      map[string]string `json:"metadata"`
}

type GatewaySessionResponse struct {
	SessionID       string            `json:"session_id"`
	UserID          string            `json:"user_id"`
	TargetAssetID   string            `json:"target_asset_id"`
	Protocol        string            `json:"protocol"`
	Status          string            `json:"status"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	ActivatedAt     *time.Time        `json:"activated_at,omitempty"`
	TerminatedAt    *time.Time        `json:"terminated_at,omitempty"`
	TerminateReason string            `json:"terminate_reason,omitempty"`
}

type TerminateSessionRequest struct {
	Reason string `json:"reason"`
}

// StartSessionRequest is kept for backward compatibility with earlier stubs.
type StartSessionRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	AssetID    string `json:"asset_id" binding:"required"`
	JITGrantID string `json:"jit_grant_id" binding:"required"`
	Protocol   string `json:"protocol" binding:"required"`
}
