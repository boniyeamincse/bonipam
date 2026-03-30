package service

import (
	"boni-pam/internal/domain"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type VaultService struct {
	mu               sync.RWMutex
	secrets          map[string]domain.SecretRecord
	leases           map[string]domain.CredentialLeaseRecord
	rotationPolicies map[string]domain.RotationPolicy
	kmsProvider      KMSProvider
	kekVersion       string
}

func NewVaultService(masterKey string) (*VaultService, error) {
	trimmed := strings.TrimSpace(masterKey)
	if trimmed == "" {
		return nil, fmt.Errorf("vault master key is required")
	}
	if len(trimmed) < 16 {
		return nil, fmt.Errorf("vault master key must be at least 16 characters")
	}

	kms, err := NewLocalKMSProvider(trimmed)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize KMS provider: %w", err)
	}

	return &VaultService{
		secrets:          make(map[string]domain.SecretRecord),
		leases:           make(map[string]domain.CredentialLeaseRecord),
		rotationPolicies: make(map[string]domain.RotationPolicy),
		kmsProvider:      kms,
		kekVersion:       "v1",
	}, nil
}

// KMSInfo returns the active KMS adapter description.
func (s *VaultService) KMSInfo() domain.KMSAdapterInfo {
	return s.kmsProvider.Describe()
}

func (s *VaultService) IssueCredential(req domain.IssueCredentialRequest) (domain.IssueCredentialResponse, error) {
	targetType := strings.ToLower(strings.TrimSpace(req.TargetType))
	targetID := strings.TrimSpace(req.TargetID)
	role := strings.TrimSpace(req.Role)
	if targetType == "" || targetID == "" || role == "" {
		return domain.IssueCredentialResponse{}, fmt.Errorf("target_type, target_id, and role are required")
	}
	if !isSupportedTargetType(targetType) {
		return domain.IssueCredentialResponse{}, fmt.Errorf("unsupported target type")
	}

	ttlSeconds := req.TTLSeconds
	if ttlSeconds == 0 {
		ttlSeconds = 900
	}
	if ttlSeconds < 60 || ttlSeconds > 86400 {
		return domain.IssueCredentialResponse{}, fmt.Errorf("ttl_seconds must be between 60 and 86400")
	}

	username := fmt.Sprintf("jit-%s-%s", normalizeForCredential(targetType), shortToken(8))
	password, err := randomPassword(24)
	if err != nil {
		return domain.IssueCredentialResponse{}, fmt.Errorf("failed to generate credential secret")
	}

	now := time.Now().UTC()
	lease := domain.CredentialLeaseRecord{
		LeaseID:      "lease-" + uuid.NewString(),
		TargetType:   targetType,
		TargetID:     targetID,
		Username:     username,
		Password:     password,
		Role:         role,
		Metadata:     req.Metadata,
		IssuedAt:     now,
		ExpiresAt:    now.Add(time.Duration(ttlSeconds) * time.Second),
		Revoked:      false,
		LeaseSeconds: ttlSeconds,
	}

	s.mu.Lock()
	s.leases[lease.LeaseID] = lease
	s.mu.Unlock()

	return domain.IssueCredentialResponse{
		LeaseID:      lease.LeaseID,
		TargetType:   lease.TargetType,
		TargetID:     lease.TargetID,
		Username:     lease.Username,
		Password:     lease.Password,
		Role:         lease.Role,
		Metadata:     lease.Metadata,
		IssuedAt:     lease.IssuedAt,
		ExpiresAt:    lease.ExpiresAt,
		Revoked:      lease.Revoked,
		LeaseSeconds: lease.LeaseSeconds,
	}, nil
}

// GetLeaseStatus returns the current status of a credential lease.
func (s *VaultService) GetLeaseStatus(leaseID string) (domain.LeaseStatusResponse, error) {
	s.mu.RLock()
	lease, ok := s.leases[leaseID]
	s.mu.RUnlock()
	if !ok {
		return domain.LeaseStatusResponse{}, fmt.Errorf("lease not found")
	}
	return toLeaseStatusResponse(lease), nil
}

// RevokeLease marks a single lease as revoked immediately.
func (s *VaultService) RevokeLease(leaseID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	lease, ok := s.leases[leaseID]
	if !ok {
		return fmt.Errorf("lease not found")
	}
	if lease.Revoked {
		return fmt.Errorf("lease already revoked")
	}
	now := time.Now().UTC()
	lease.Revoked = true
	lease.RevokedAt = &now
	lease.RevokeReason = strings.TrimSpace(reason)
	s.leases[leaseID] = lease
	return nil
}

// RevokeLeasesByTarget revokes all active leases for a given target asset.
func (s *VaultService) RevokeLeasesByTarget(targetID, reason string) (domain.BulkRevokeResult, error) {
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		return domain.BulkRevokeResult{}, fmt.Errorf("target_id is required")
	}

	now := time.Now().UTC()
	trimmedReason := strings.TrimSpace(reason)

	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, lease := range s.leases {
		if strings.EqualFold(lease.TargetID, targetID) && !lease.Revoked {
			lease.Revoked = true
			lease.RevokedAt = &now
			lease.RevokeReason = trimmedReason
			s.leases[id] = lease
			count++
		}
	}

	return domain.BulkRevokeResult{Revoked: count, TargetID: targetID}, nil
}

func toLeaseStatusResponse(l domain.CredentialLeaseRecord) domain.LeaseStatusResponse {
	status := "active"
	if l.Revoked {
		status = "revoked"
	} else if time.Now().UTC().After(l.ExpiresAt) {
		status = "expired"
	}
	return domain.LeaseStatusResponse{
		LeaseID:      l.LeaseID,
		TargetType:   l.TargetType,
		TargetID:     l.TargetID,
		Username:     l.Username,
		Role:         l.Role,
		Metadata:     l.Metadata,
		IssuedAt:     l.IssuedAt,
		ExpiresAt:    l.ExpiresAt,
		Status:       status,
		Revoked:      l.Revoked,
		RevokedAt:    l.RevokedAt,
		RevokeReason: l.RevokeReason,
		LeaseSeconds: l.LeaseSeconds,
	}
}

// CreateRotationPolicy registers a new rotation policy for a target.
func (s *VaultService) CreateRotationPolicy(req domain.CreateRotationPolicyRequest) (domain.RotationPolicyResponse, error) {
	targetType := strings.ToLower(strings.TrimSpace(req.TargetType))
	targetID := strings.TrimSpace(req.TargetID)
	role := strings.TrimSpace(req.Role)
	if targetType == "" || targetID == "" || role == "" {
		return domain.RotationPolicyResponse{}, fmt.Errorf("target_type, target_id, and role are required")
	}
	if !isSupportedTargetType(targetType) {
		return domain.RotationPolicyResponse{}, fmt.Errorf("unsupported target type")
	}
	if req.IntervalSeconds < 60 {
		return domain.RotationPolicyResponse{}, fmt.Errorf("interval_seconds must be at least 60")
	}
	ttl := req.TTLSeconds
	if ttl == 0 {
		ttl = req.IntervalSeconds
	}
	if ttl < 60 || ttl > 86400 {
		return domain.RotationPolicyResponse{}, fmt.Errorf("ttl_seconds must be between 60 and 86400")
	}

	now := time.Now().UTC()
	policy := domain.RotationPolicy{
		PolicyID:        "rp-" + uuid.NewString(),
		TargetType:      targetType,
		TargetID:        targetID,
		Role:            role,
		IntervalSeconds: req.IntervalSeconds,
		TTLSeconds:      ttl,
		Metadata:        req.Metadata,
		Enabled:         true,
		CreatedAt:       now,
		UpdatedAt:       now,
		LastRotatedAt:   nil,
		NextRotationAt:  now.Add(time.Duration(req.IntervalSeconds) * time.Second),
	}

	s.mu.Lock()
	for _, existing := range s.rotationPolicies {
		if strings.EqualFold(existing.TargetID, targetID) &&
			strings.EqualFold(existing.TargetType, targetType) &&
			strings.EqualFold(existing.Role, role) {
			s.mu.Unlock()
			return domain.RotationPolicyResponse{}, fmt.Errorf("rotation policy already exists for this target and role")
		}
	}
	s.rotationPolicies[policy.PolicyID] = policy
	s.mu.Unlock()

	return toRotationPolicyResponse(policy), nil
}

// GetRotationPolicy returns a rotation policy by ID.
func (s *VaultService) GetRotationPolicy(policyID string) (domain.RotationPolicyResponse, error) {
	s.mu.RLock()
	policy, ok := s.rotationPolicies[policyID]
	s.mu.RUnlock()
	if !ok {
		return domain.RotationPolicyResponse{}, fmt.Errorf("rotation policy not found")
	}
	return toRotationPolicyResponse(policy), nil
}

// TriggerRotation immediately rotates the credential for a policy.
func (s *VaultService) TriggerRotation(policyID string) (domain.RotationResult, error) {
	s.mu.Lock()
	policy, ok := s.rotationPolicies[policyID]
	if !ok {
		s.mu.Unlock()
		return domain.RotationResult{}, fmt.Errorf("rotation policy not found")
	}
	if !policy.Enabled {
		s.mu.Unlock()
		return domain.RotationResult{}, fmt.Errorf("rotation policy is disabled")
	}
	s.mu.Unlock()

	issued, err := s.IssueCredential(domain.IssueCredentialRequest{
		TargetType: policy.TargetType,
		TargetID:   policy.TargetID,
		Role:       policy.Role,
		TTLSeconds: policy.TTLSeconds,
		Metadata:   policy.Metadata,
	})
	if err != nil {
		return domain.RotationResult{}, fmt.Errorf("failed to rotate credential: %w", err)
	}

	now := time.Now().UTC()
	s.mu.Lock()
	policy.LastRotatedAt = &now
	policy.NextRotationAt = now.Add(time.Duration(policy.IntervalSeconds) * time.Second)
	policy.UpdatedAt = now
	s.rotationPolicies[policyID] = policy
	s.mu.Unlock()

	return domain.RotationResult{
		PolicyID:  policyID,
		LeaseID:   issued.LeaseID,
		Username:  issued.Username,
		RotatedAt: now,
		ExpiresAt: issued.ExpiresAt,
	}, nil
}

// RunDueRotations checks all enabled policies and rotates any that are past their next rotation time.
func (s *VaultService) RunDueRotations() []domain.RotationResult {
	s.mu.RLock()
	var dueIDs []string
	for id, policy := range s.rotationPolicies {
		if policy.Enabled && time.Now().UTC().After(policy.NextRotationAt) {
			dueIDs = append(dueIDs, id)
		}
	}
	s.mu.RUnlock()

	var results []domain.RotationResult
	for _, id := range dueIDs {
		result, err := s.TriggerRotation(id)
		if err == nil {
			results = append(results, result)
		}
	}
	return results
}

func toRotationPolicyResponse(p domain.RotationPolicy) domain.RotationPolicyResponse {
	return domain.RotationPolicyResponse{
		PolicyID:        p.PolicyID,
		TargetType:      p.TargetType,
		TargetID:        p.TargetID,
		Role:            p.Role,
		IntervalSeconds: p.IntervalSeconds,
		TTLSeconds:      p.TTLSeconds,
		Metadata:        p.Metadata,
		Enabled:         p.Enabled,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		LastRotatedAt:   p.LastRotatedAt,
		NextRotationAt:  p.NextRotationAt,
	}
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
	wrapped, err := s.kmsProvider.WrapKey(dek)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to wrap data key: %w", err)
	}

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

	dek, err := s.kmsProvider.UnwrapKey(wrappedDEK)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap data key: %w", err)
	}
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

func isSupportedTargetType(targetType string) bool {
	switch targetType {
	case "database", "ssh", "api":
		return true
	default:
		return false
	}
}

func normalizeForCredential(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	if normalized == "" {
		return "target"
	}
	return normalized
}

func shortToken(length int) string {
	token, err := randomString(length, "abcdefghijklmnopqrstuvwxyz0123456789")
	if err != nil {
		return uuid.NewString()[:length]
	}
	return token
}

func randomPassword(length int) (string, error) {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_-+=<>?"
	return randomString(length, charset)
}

func randomString(length int, charset string) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid random string length")
	}
	if charset == "" {
		return "", fmt.Errorf("charset is required")
	}

	buf := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = charset[n.Int64()]
	}

	return string(buf), nil
}
