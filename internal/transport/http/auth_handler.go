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
	auth.POST("/mfa/verify", h.VerifyMFA)
	auth.POST("/token/refresh", h.RefreshToken)
	auth.POST("/logout", h.Logout)
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

func (h *AuthHandler) Logout(c *gin.Context) {
	RespondOK(c, http.StatusOK, gin.H{"message": "logout successful"})
}
