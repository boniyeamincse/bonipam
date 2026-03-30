package service

import (
	"boni-pam/internal/domain"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

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

type AuthService struct {
	mu         sync.RWMutex
	challenges map[string]mfaChallenge
}

func NewAuthService() *AuthService {
	return &AuthService{
		challenges: make(map[string]mfaChallenge),
	}
}

func (s *AuthService) ExchangeOIDCCode(req domain.OIDCCallbackRequest) (domain.AuthSession, error) {
	if req.Code == "" || req.State == "" {
		return domain.AuthSession{}, fmt.Errorf("invalid OIDC callback payload")
	}

	return s.issueSession("user-" + uuid.NewString()), nil
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

	return s.issueSession(challenge.UserID), nil
}

func (s *AuthService) RefreshToken(refreshToken string) (domain.AuthSession, error) {
	if refreshToken == "" {
		return domain.AuthSession{}, fmt.Errorf("missing refresh token")
	}

	return s.issueSession("user-" + uuid.NewString()), nil
}

func (s *AuthService) issueSession(userID string) domain.AuthSession {
	now := time.Now().UTC()
	return domain.AuthSession{
		UserID:       userID,
		AccessToken:  "atk-" + uuid.NewString(),
		RefreshToken: "rtk-" + uuid.NewString(),
		ExpiresAt:    now.Add(15 * time.Minute),
	}
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
