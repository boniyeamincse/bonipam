package domain

import "time"

type SecretRecord struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Ciphertext string            `json:"-"`
	WrappedDEK string            `json:"-"`
	Nonce      string            `json:"-"`
	KEKVersion string            `json:"kek_version"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type CreateSecretRequest struct {
	Name     string            `json:"name" binding:"required"`
	Value    string            `json:"value" binding:"required"`
	Metadata map[string]string `json:"metadata"`
}

type CreateSecretResponse struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	KEKVersion string            `json:"kek_version"`
	CreatedAt  time.Time         `json:"created_at"`
}

type GetSecretResponse struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Value      string            `json:"value"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	KEKVersion string            `json:"kek_version"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type IssueCredentialRequest struct {
	TargetType string            `json:"target_type" binding:"required"`
	TargetID   string            `json:"target_id" binding:"required"`
	Role       string            `json:"role" binding:"required"`
	TTLSeconds int               `json:"ttl_seconds"`
	Metadata   map[string]string `json:"metadata"`
}

type IssueCredentialResponse struct {
	LeaseID      string            `json:"lease_id"`
	TargetType   string            `json:"target_type"`
	TargetID     string            `json:"target_id"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	Role         string            `json:"role"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	IssuedAt     time.Time         `json:"issued_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Revoked      bool              `json:"revoked"`
	LeaseSeconds int               `json:"lease_seconds"`
}

type CredentialLeaseRecord struct {
	LeaseID      string            `json:"lease_id"`
	TargetType   string            `json:"target_type"`
	TargetID     string            `json:"target_id"`
	Username     string            `json:"username"`
	Password     string            `json:"-"`
	Role         string            `json:"role"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	IssuedAt     time.Time         `json:"issued_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Revoked      bool              `json:"revoked"`
	RevokedAt    *time.Time        `json:"revoked_at,omitempty"`
	RevokeReason string            `json:"revoke_reason,omitempty"`
	LeaseSeconds int               `json:"lease_seconds"`
}

type RevokeLeaseRequest struct {
	Reason string `json:"reason"`
}

type RevokeByTargetRequest struct {
	TargetID string `json:"target_id" binding:"required"`
	Reason   string `json:"reason"`
}

type LeaseStatusResponse struct {
	LeaseID      string            `json:"lease_id"`
	TargetType   string            `json:"target_type"`
	TargetID     string            `json:"target_id"`
	Username     string            `json:"username"`
	Role         string            `json:"role"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	IssuedAt     time.Time         `json:"issued_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Status       string            `json:"status"` // active | expired | revoked
	Revoked      bool              `json:"revoked"`
	RevokedAt    *time.Time        `json:"revoked_at,omitempty"`
	RevokeReason string            `json:"revoke_reason,omitempty"`
	LeaseSeconds int               `json:"lease_seconds"`
}

type BulkRevokeResult struct {
	Revoked  int    `json:"revoked"`
	TargetID string `json:"target_id"`
}

// RotationPolicy defines periodic rotation rules for a target.
type RotationPolicy struct {
	PolicyID        string            `json:"policy_id"`
	TargetType      string            `json:"target_type"`
	TargetID        string            `json:"target_id"`
	Role            string            `json:"role"`
	IntervalSeconds int               `json:"interval_seconds"`
	TTLSeconds      int               `json:"ttl_seconds"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Enabled         bool              `json:"enabled"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	LastRotatedAt   *time.Time        `json:"last_rotated_at,omitempty"`
	NextRotationAt  time.Time         `json:"next_rotation_at"`
}

type CreateRotationPolicyRequest struct {
	TargetType      string            `json:"target_type" binding:"required"`
	TargetID        string            `json:"target_id" binding:"required"`
	Role            string            `json:"role" binding:"required"`
	IntervalSeconds int               `json:"interval_seconds" binding:"required"`
	TTLSeconds      int               `json:"ttl_seconds"`
	Metadata        map[string]string `json:"metadata"`
}

type RotationPolicyResponse struct {
	PolicyID        string            `json:"policy_id"`
	TargetType      string            `json:"target_type"`
	TargetID        string            `json:"target_id"`
	Role            string            `json:"role"`
	IntervalSeconds int               `json:"interval_seconds"`
	TTLSeconds      int               `json:"ttl_seconds"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Enabled         bool              `json:"enabled"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	LastRotatedAt   *time.Time        `json:"last_rotated_at,omitempty"`
	NextRotationAt  time.Time         `json:"next_rotation_at"`
}

type RotationResult struct {
	PolicyID  string    `json:"policy_id"`
	LeaseID   string    `json:"lease_id"`
	Username  string    `json:"username"`
	RotatedAt time.Time `json:"rotated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// KMSAdapterInfo describes the active KMS/HSM provider configuration.
type KMSAdapterInfo struct {
	Provider   string `json:"provider"` // "local" | "aws-kms" | "hashicorp-vault"
	KeyID      string `json:"key_id"`
	KEKVersion string `json:"kek_version"`
	Endpoint   string `json:"endpoint,omitempty"`
}
