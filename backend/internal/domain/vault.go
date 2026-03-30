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
	LeaseSeconds int               `json:"lease_seconds"`
}
