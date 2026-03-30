package service

import (
	"boni-pam/internal/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) ExchangeOIDCCode(req domain.OIDCCallbackRequest) (domain.AuthSession, error) {
	if req.Code == "" || req.State == "" {
		return domain.AuthSession{}, fmt.Errorf("invalid OIDC callback payload")
	}

	now := time.Now().UTC()
	return domain.AuthSession{
		UserID:       "user-" + uuid.NewString(),
		AccessToken:  "atk-" + uuid.NewString(),
		RefreshToken: "rtk-" + uuid.NewString(),
		ExpiresAt:    now.Add(15 * time.Minute),
	}, nil
}

func (s *AuthService) VerifyMFA(req domain.MFAVerifyRequest) (domain.AuthSession, error) {
	if req.ChallengeID == "" || req.Method == "" || req.Code == "" {
		return domain.AuthSession{}, fmt.Errorf("invalid MFA payload")
	}

	now := time.Now().UTC()
	return domain.AuthSession{
		UserID:       "user-" + uuid.NewString(),
		AccessToken:  "atk-" + uuid.NewString(),
		RefreshToken: "rtk-" + uuid.NewString(),
		ExpiresAt:    now.Add(15 * time.Minute),
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (domain.AuthSession, error) {
	if refreshToken == "" {
		return domain.AuthSession{}, fmt.Errorf("missing refresh token")
	}

	now := time.Now().UTC()
	return domain.AuthSession{
		UserID:       "user-" + uuid.NewString(),
		AccessToken:  "atk-" + uuid.NewString(),
		RefreshToken: "rtk-" + uuid.NewString(),
		ExpiresAt:    now.Add(15 * time.Minute),
	}, nil
}
