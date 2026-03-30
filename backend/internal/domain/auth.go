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
	Code        string `json:"code"`
	Assertion   string `json:"assertion"`
}

type MFAChallengeRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Method string `json:"method" binding:"required"`
}

type MFAChallengeResponse struct {
	ChallengeID       string    `json:"challenge_id"`
	Method            string    `json:"method"`
	ExpiresAt         time.Time `json:"expires_at"`
	WebAuthnChallenge string    `json:"webauthn_challenge,omitempty"`
	TestCode          string    `json:"test_code,omitempty"`
}

type RevokeSessionsRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	SessionID string `json:"session_id"`
}

type RevokeSessionsResponse struct {
	RevokedCount int      `json:"revoked_count"`
	RevokedIDs   []string `json:"revoked_ids"`
}

type OIDCCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}
