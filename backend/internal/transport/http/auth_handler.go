package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterRoutes(group *gin.RouterGroup) {
	auth := group.Group("/auth")
	auth.POST("/sso/callback", h.SSOCallback)
	auth.POST("/mfa/challenge", h.CreateMFAChallenge)
	auth.POST("/mfa/verify", h.VerifyMFA)
	auth.POST("/token/refresh", h.RefreshToken)
	auth.POST("/sessions/revoke", h.RevokeSessions)
	auth.POST("/logout", h.Logout)
}

func (h *AuthHandler) CreateMFAChallenge(c *gin.Context) {
	var req domain.MFAChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid mfa challenge request")
		return
	}

	challenge, err := h.authService.CreateMFAChallenge(req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "MFA_CHALLENGE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, challenge)
}

func (h *AuthHandler) SSOCallback(c *gin.Context) {
	var req domain.OIDCCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid sso callback request")
		return
	}

	session, err := h.authService.ExchangeOIDCCode(req)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, gin.H{
		"user_id":       session.UserID,
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"expires_in":    900,
	})
}

func (h *AuthHandler) VerifyMFA(c *gin.Context) {
	var req domain.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid mfa request")
		return
	}

	session, err := h.authService.VerifyMFA(req)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "MFA_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, gin.H{
		"user_id":       session.UserID,
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"expires_in":    900,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "refresh_token is required")
		return
	}

	session, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "TOKEN_REFRESH_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, gin.H{
		"user_id":       session.UserID,
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"expires_in":    900,
	})
}

func (h *AuthHandler) RevokeSessions(c *gin.Context) {
	var req domain.RevokeSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid session revocation request")
		return
	}

	result, err := h.authService.RevokeSessions(req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "SESSION_REVOCATION_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	RespondOK(c, http.StatusOK, gin.H{"message": "logout successful"})
}
