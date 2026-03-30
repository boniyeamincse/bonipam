package service

import (
	"boni-pam/internal/domain"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type mfaChallenge struct {
	ChallengeID       string
	UserID            string
	Method            string
	Code              string
	WebAuthnChallenge string
	ExpiresAt         time.Time
}

type refreshTokenRecord struct {
	UserID    string
	SessionID string
	ExpiresAt time.Time
}

type sessionRecord struct {
	UserID           string
	RefreshTokenHash string
	IssuedAt         time.Time
	RevokedAt        *time.Time
}

type AuthTokenConfig struct {
	Issuer           string
	SigningKey       string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	RequireStrongKey bool
}

type AuthService struct {
	mu            sync.RWMutex
	challenges    map[string]mfaChallenge
	refreshTokens map[string]refreshTokenRecord
	sessions      map[string]sessionRecord
	tokenConfig   AuthTokenConfig
}

func NewAuthService(cfg AuthTokenConfig) (*AuthService, error) {
	if cfg.Issuer == "" {
		cfg.Issuer = "boni-pam-auth"
	}
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = 24 * time.Hour
	}
	if cfg.SigningKey == "" {
		cfg.SigningKey = os.Getenv("JWT_SIGNING_KEY")
	}
	if cfg.SigningKey == "" {
		return nil, fmt.Errorf("missing JWT signing key")
	}
	if cfg.RequireStrongKey && len(cfg.SigningKey) < 32 {
		return nil, fmt.Errorf("JWT signing key must be at least 32 characters")
	}

	return &AuthService{
		challenges:    make(map[string]mfaChallenge),
		refreshTokens: make(map[string]refreshTokenRecord),
		sessions:      make(map[string]sessionRecord),
		tokenConfig:   cfg,
	}, nil
}

func (s *AuthService) ExchangeOIDCCode(req domain.OIDCCallbackRequest) (domain.AuthSession, error) {
	if req.Code == "" || req.State == "" {
		return domain.AuthSession{}, fmt.Errorf("invalid OIDC callback payload")
	}

	return s.issueSession("user-" + uuid.NewString())
}

func (s *AuthService) CreateMFAChallenge(req domain.MFAChallengeRequest) (domain.MFAChallengeResponse, error) {
	if req.UserID == "" || req.Method == "" {
		return domain.MFAChallengeResponse{}, fmt.Errorf("invalid MFA challenge payload")
	}

	method := strings.ToLower(req.Method)
	if method != "totp" && method != "webauthn" {
		return domain.MFAChallengeResponse{}, fmt.Errorf("unsupported MFA method: %s", req.Method)
	}

	challengeID := "mfa-" + uuid.NewString()
	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	challenge := mfaChallenge{
		ChallengeID: challengeID,
		UserID:      req.UserID,
		Method:      method,
		ExpiresAt:   expiresAt,
	}

	response := domain.MFAChallengeResponse{
		ChallengeID: challengeID,
		Method:      method,
		ExpiresAt:   expiresAt,
	}

	if method == "totp" {
		code, err := randomDigits(6)
		if err != nil {
			return domain.MFAChallengeResponse{}, fmt.Errorf("failed to generate TOTP challenge")
		}
		challenge.Code = code
		// For scaffold/testing only. A production system should never return this.
		response.TestCode = code
	} else {
		webauthnChallenge, err := randomToken(32)
		if err != nil {
			return domain.MFAChallengeResponse{}, fmt.Errorf("failed to generate WebAuthn challenge")
		}
		challenge.WebAuthnChallenge = webauthnChallenge
		response.WebAuthnChallenge = webauthnChallenge
	}

	s.mu.Lock()
	s.challenges[challengeID] = challenge
	s.mu.Unlock()

	return response, nil
}

func (s *AuthService) VerifyMFA(req domain.MFAVerifyRequest) (domain.AuthSession, error) {
	if req.ChallengeID == "" || req.Method == "" {
		return domain.AuthSession{}, fmt.Errorf("invalid MFA payload")
	}

	method := strings.ToLower(req.Method)

	s.mu.RLock()
	challenge, ok := s.challenges[req.ChallengeID]
	s.mu.RUnlock()
	if !ok {
		return domain.AuthSession{}, fmt.Errorf("invalid or expired challenge")
	}

	if challenge.Method != method {
		return domain.AuthSession{}, fmt.Errorf("mfa method mismatch")
	}

	if time.Now().UTC().After(challenge.ExpiresAt) {
		s.mu.Lock()
		delete(s.challenges, req.ChallengeID)
		s.mu.Unlock()
		return domain.AuthSession{}, fmt.Errorf("challenge expired")
	}

	switch method {
	case "totp":
		if req.Code == "" || req.Code != challenge.Code {
			return domain.AuthSession{}, fmt.Errorf("invalid totp code")
		}
	case "webauthn":
		if req.Assertion == "" || req.Assertion != challenge.WebAuthnChallenge {
			return domain.AuthSession{}, fmt.Errorf("invalid webauthn assertion")
		}
	default:
		return domain.AuthSession{}, fmt.Errorf("unsupported MFA method: %s", req.Method)
	}

	s.mu.Lock()
	delete(s.challenges, req.ChallengeID)
	s.mu.Unlock()

	return s.issueSession(challenge.UserID)
}

func (s *AuthService) RefreshToken(refreshToken string) (domain.AuthSession, error) {
	if refreshToken == "" {
		return domain.AuthSession{}, fmt.Errorf("missing refresh token")
	}

	hash := hashToken(refreshToken)

	s.mu.Lock()
	record, ok := s.refreshTokens[hash]
	if !ok {
		s.mu.Unlock()
		return domain.AuthSession{}, fmt.Errorf("invalid refresh token")
	}

	if time.Now().UTC().After(record.ExpiresAt) {
		delete(s.refreshTokens, hash)
		if session, ok := s.sessions[record.SessionID]; ok {
			now := time.Now().UTC()
			session.RevokedAt = &now
			s.sessions[record.SessionID] = session
		}
		s.mu.Unlock()
		return domain.AuthSession{}, fmt.Errorf("refresh token expired")
	}

	// Rotation: old refresh token is invalidated before issuing a new token.
	delete(s.refreshTokens, hash)
	if session, ok := s.sessions[record.SessionID]; ok {
		session.RefreshTokenHash = ""
		s.sessions[record.SessionID] = session
	}
	s.mu.Unlock()

	return s.issueSession(record.UserID)
}

func (s *AuthService) RevokeSessions(req domain.RevokeSessionsRequest) (domain.RevokeSessionsResponse, error) {
	if req.UserID == "" {
		return domain.RevokeSessionsResponse{}, fmt.Errorf("user_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	revokedIDs := make([]string, 0)
	now := time.Now().UTC()

	if req.SessionID != "" {
		session, ok := s.sessions[req.SessionID]
		if !ok {
			return domain.RevokeSessionsResponse{}, fmt.Errorf("session not found")
		}
		if session.UserID != req.UserID {
			return domain.RevokeSessionsResponse{}, fmt.Errorf("session does not belong to user")
		}
		if session.RevokedAt != nil {
			return domain.RevokeSessionsResponse{RevokedCount: 0, RevokedIDs: []string{}}, nil
		}

		if session.RefreshTokenHash != "" {
			delete(s.refreshTokens, session.RefreshTokenHash)
		}
		session.RevokedAt = &now
		session.RefreshTokenHash = ""
		s.sessions[req.SessionID] = session
		revokedIDs = append(revokedIDs, req.SessionID)

		return domain.RevokeSessionsResponse{RevokedCount: 1, RevokedIDs: revokedIDs}, nil
	}

	for sessionID, session := range s.sessions {
		if session.UserID != req.UserID || session.RevokedAt != nil {
			continue
		}
		if session.RefreshTokenHash != "" {
			delete(s.refreshTokens, session.RefreshTokenHash)
		}
		session.RevokedAt = &now
		session.RefreshTokenHash = ""
		s.sessions[sessionID] = session
		revokedIDs = append(revokedIDs, sessionID)
	}

	return domain.RevokeSessionsResponse{RevokedCount: len(revokedIDs), RevokedIDs: revokedIDs}, nil
}

func (s *AuthService) issueSession(userID string) (domain.AuthSession, error) {
	now := time.Now().UTC()
	accessExp := now.Add(s.tokenConfig.AccessTokenTTL)
	sessionID := uuid.NewString()

	claims := jwt.MapClaims{
		"sub": userID,
		"iss": s.tokenConfig.Issuer,
		"iat": now.Unix(),
		"exp": accessExp.Unix(),
		"jti": sessionID,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.tokenConfig.SigningKey))
	if err != nil {
		return domain.AuthSession{}, fmt.Errorf("failed to sign access token")
	}

	refreshToken, err := randomToken(48)
	if err != nil {
		return domain.AuthSession{}, fmt.Errorf("failed to generate refresh token")
	}

	refreshExp := now.Add(s.tokenConfig.RefreshTokenTTL)
	refreshHash := hashToken(refreshToken)

	s.mu.Lock()
	s.refreshTokens[refreshHash] = refreshTokenRecord{
		UserID:    userID,
		SessionID: sessionID,
		ExpiresAt: refreshExp,
	}
	s.sessions[sessionID] = sessionRecord{
		UserID:           userID,
		RefreshTokenHash: refreshHash,
		IssuedAt:         now,
	}
	s.mu.Unlock()

	return domain.AuthSession{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func randomDigits(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid code length")
	}

	max := big.NewInt(10)
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = byte('0' + n.Int64())
	}
	return string(buf), nil
}

func randomToken(byteLen int) (string, error) {
	if byteLen <= 0 {
		return "", fmt.Errorf("invalid token length")
	}

	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
