package service

import (
	"boni-pam/internal/domain"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

// KMSProvider abstracts key wrapping and unwrapping operations,
// allowing the vault to delegate to a local KEK or an external KMS/HSM.
type KMSProvider interface {
	WrapKey(dek []byte) ([]byte, error)
	UnwrapKey(wrapped []byte) ([]byte, error)
	Describe() domain.KMSAdapterInfo
}

// LocalKMSProvider implements KMSProvider using a local master key.
// It uses XOR with a SHA-256 digest of the master key as the KEK.
type LocalKMSProvider struct {
	masterKey  []byte
	kekVersion string
	keyID      string
}

func NewLocalKMSProvider(masterKey string) (*LocalKMSProvider, error) {
	trimmed := strings.TrimSpace(masterKey)
	if trimmed == "" {
		return nil, fmt.Errorf("master key is required")
	}
	if len(trimmed) < 16 {
		return nil, fmt.Errorf("master key must be at least 16 characters")
	}
	return &LocalKMSProvider{
		masterKey:  []byte(trimmed),
		kekVersion: "v1",
		keyID:      "local-kek-v1",
	}, nil
}

func (p *LocalKMSProvider) WrapKey(dek []byte) ([]byte, error) {
	if len(dek) == 0 {
		return nil, fmt.Errorf("dek must not be empty")
	}
	kek := sha256.Sum256(p.masterKey)
	wrapped := make([]byte, len(dek))
	for i := 0; i < len(dek); i++ {
		wrapped[i] = dek[i] ^ kek[i%len(kek)]
	}
	return wrapped, nil
}

func (p *LocalKMSProvider) UnwrapKey(wrapped []byte) ([]byte, error) {
	if len(wrapped) == 0 {
		return nil, fmt.Errorf("wrapped key must not be empty")
	}
	kek := sha256.Sum256(p.masterKey)
	dek := make([]byte, len(wrapped))
	for i := 0; i < len(wrapped); i++ {
		dek[i] = wrapped[i] ^ kek[i%len(kek)]
	}
	return dek, nil
}

func (p *LocalKMSProvider) Describe() domain.KMSAdapterInfo {
	return domain.KMSAdapterInfo{
		Provider:   "local",
		KeyID:      p.keyID,
		KEKVersion: p.kekVersion,
	}
}

// RemoteKMSProvider is a stub for external KMS/HSM integration
// (e.g. AWS KMS, HashiCorp Vault Transit). All operations return
// ErrNotImplemented until a real adapter is wired.
type RemoteKMSProvider struct {
	provider string
	endpoint string
	keyID    string
}

var ErrKMSNotImplemented = errors.New("remote KMS adapter not yet configured")

func NewRemoteKMSProvider(provider, endpoint, keyID string) *RemoteKMSProvider {
	return &RemoteKMSProvider{
		provider: strings.ToLower(strings.TrimSpace(provider)),
		endpoint: strings.TrimSpace(endpoint),
		keyID:    strings.TrimSpace(keyID),
	}
}

func (p *RemoteKMSProvider) WrapKey(_ []byte) ([]byte, error) {
	return nil, fmt.Errorf("%w: provider=%s", ErrKMSNotImplemented, p.provider)
}

func (p *RemoteKMSProvider) UnwrapKey(_ []byte) ([]byte, error) {
	return nil, fmt.Errorf("%w: provider=%s", ErrKMSNotImplemented, p.provider)
}

func (p *RemoteKMSProvider) Describe() domain.KMSAdapterInfo {
	return domain.KMSAdapterInfo{
		Provider:   p.provider,
		KeyID:      p.keyID,
		KEKVersion: "n/a",
		Endpoint:   p.endpoint,
	}
}
