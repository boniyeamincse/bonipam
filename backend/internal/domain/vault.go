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
