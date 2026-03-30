package domain

import "time"

type AuthSession struct {
	UserID       string    `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type MFAVerifyRequest struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type OIDCCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}
