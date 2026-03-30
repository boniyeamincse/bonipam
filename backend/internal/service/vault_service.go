package service

import (
	"boni-pam/internal/domain"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type VaultService struct {
	mu          sync.RWMutex
	secrets     map[string]domain.SecretRecord
	masterKey   []byte
	kekVersion  string
	masterAlias string
}

func NewVaultService(masterKey string) (*VaultService, error) {
	trimmed := strings.TrimSpace(masterKey)
	if trimmed == "" {
		return nil, fmt.Errorf("vault master key is required")
	}
	if len(trimmed) < 16 {
		return nil, fmt.Errorf("vault master key must be at least 16 characters")
	}

	return &VaultService{
		secrets:     make(map[string]domain.SecretRecord),
		masterKey:   []byte(trimmed),
		kekVersion:  "v1",
		masterAlias: "local-kek",
	}, nil
}

func (s *VaultService) StoreSecret(req domain.CreateSecretRequest) (domain.CreateSecretResponse, error) {
	name := strings.TrimSpace(req.Name)
	value := strings.TrimSpace(req.Value)
	if name == "" || value == "" {
		return domain.CreateSecretResponse{}, fmt.Errorf("name and value are required")
	}

	ciphertext, wrappedDEK, nonce, err := s.encryptWithEnvelope([]byte(value))
	if err != nil {
		return domain.CreateSecretResponse{}, err
	}

	now := time.Now().UTC()
	record := domain.SecretRecord{
		ID:         "sec-" + uuid.NewString(),
		Name:       name,
		Metadata:   req.Metadata,
		Ciphertext: ciphertext,
		WrappedDEK: wrappedDEK,
		Nonce:      nonce,
		KEKVersion: s.kekVersion,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.secrets {
		if strings.EqualFold(existing.Name, record.Name) {
			return domain.CreateSecretResponse{}, fmt.Errorf("secret name already exists")
		}
	}
	s.secrets[record.ID] = record

	return domain.CreateSecretResponse{
		ID:         record.ID,
		Name:       record.Name,
		Metadata:   record.Metadata,
		KEKVersion: record.KEKVersion,
		CreatedAt:  record.CreatedAt,
	}, nil
}

func (s *VaultService) GetSecret(secretID string) (domain.GetSecretResponse, error) {
	s.mu.RLock()
	record, ok := s.secrets[secretID]
	s.mu.RUnlock()
	if !ok {
		return domain.GetSecretResponse{}, fmt.Errorf("secret not found")
	}

	plaintext, err := s.decryptWithEnvelope(record.Ciphertext, record.WrappedDEK, record.Nonce)
	if err != nil {
		return domain.GetSecretResponse{}, err
	}

	return domain.GetSecretResponse{
		ID:         record.ID,
		Name:       record.Name,
		Value:      string(plaintext),
		Metadata:   record.Metadata,
		KEKVersion: record.KEKVersion,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}, nil
}

func (s *VaultService) encryptWithEnvelope(plaintext []byte) (string, string, string, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return "", "", "", fmt.Errorf("failed to generate data key: %w", err)
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	wrapped := s.wrapDEK(dek)

	return base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(wrapped),
		base64.StdEncoding.EncodeToString(nonce), nil
}

func (s *VaultService) decryptWithEnvelope(ciphertextB64, wrappedDEKB64, nonceB64 string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	wrappedDEK, err := base64.StdEncoding.DecodeString(wrappedDEKB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode wrapped dek: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	dek := s.unwrapDEK(wrappedDEK)
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret")
	}

	return plaintext, nil
}

func (s *VaultService) wrapDEK(dek []byte) []byte {
	kek := sha256.Sum256(s.masterKey)
	wrapped := make([]byte, len(dek))
	for i := 0; i < len(dek); i++ {
		wrapped[i] = dek[i] ^ kek[i%len(kek)]
	}
	return wrapped
}

func (s *VaultService) unwrapDEK(wrapped []byte) []byte {
	kek := sha256.Sum256(s.masterKey)
	dek := make([]byte, len(wrapped))
	for i := 0; i < len(wrapped); i++ {
		dek[i] = wrapped[i] ^ kek[i%len(kek)]
	}
	return dek
}
