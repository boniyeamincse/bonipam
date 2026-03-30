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
	JITGrantID    string            `json:"jit_grant_id"`
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

// ---- JIT Grant types ----

// JITGrant represents a Just-in-Time access grant request with approval lifecycle.
type JITGrant struct {
	GrantID     string     `json:"grant_id"`
	UserID      string     `json:"user_id"`
	AssetID     string     `json:"asset_id"`
	Protocol    string     `json:"protocol"`
	Reason      string     `json:"reason"`
	Status      string     `json:"status"`       // pending | approved | denied | revoked
	TTLSeconds  int        `json:"ttl_seconds"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	ApproverID  string     `json:"approver_id,omitempty"`
	DenyReason  string     `json:"deny_reason,omitempty"`
}

type RequestJITGrantRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	AssetID    string `json:"asset_id" binding:"required"`
	Protocol   string `json:"protocol" binding:"required"`
	Reason     string `json:"reason" binding:"required"`
	TTLSeconds int    `json:"ttl_seconds"`
}

type ApproveJITGrantRequest struct {
	ApproverID string `json:"approver_id" binding:"required"`
}

type DenyJITGrantRequest struct {
	ApproverID string `json:"approver_id" binding:"required"`
	Reason     string `json:"reason"`
}

type JITGrantResponse struct {
	GrantID    string     `json:"grant_id"`
	UserID     string     `json:"user_id"`
	AssetID    string     `json:"asset_id"`
	Protocol   string     `json:"protocol"`
	Reason     string     `json:"reason"`
	Status     string     `json:"status"`
	TTLSeconds int        `json:"ttl_seconds"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	ApproverID string     `json:"approver_id,omitempty"`
	DenyReason string     `json:"deny_reason,omitempty"`
}
