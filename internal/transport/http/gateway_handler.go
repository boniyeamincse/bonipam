package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GatewayHandler struct {
	gatewayService *service.GatewayService
}

func NewGatewayHandler(gatewayService *service.GatewayService) *GatewayHandler {
	return &GatewayHandler{gatewayService: gatewayService}
}

func (h *GatewayHandler) RegisterRoutes(group *gin.RouterGroup) {
	gateway := group.Group("/gateway/sessions")
	gateway.POST("/start", h.StartSession)
	gateway.POST("/:sessionId/terminate", h.TerminateSession)
	gateway.GET("/:sessionId/status", h.Status)
}

func (h *GatewayHandler) StartSession(c *gin.Context) {
	var req domain.StartSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid session start payload")
		return
	}

	session, err := h.gatewayService.StartSession(req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "SESSION_START_FAILED", err.Error())
		return
	}
	RespondOK(c, http.StatusOK, session)
}

func (h *GatewayHandler) TerminateSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if err := h.gatewayService.TerminateSession(sessionID); err != nil {
		RespondError(c, http.StatusBadRequest, "SESSION_TERMINATE_FAILED", err.Error())
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"session_id": sessionID, "status": "terminated"})
}

func (h *GatewayHandler) Status(c *gin.Context) {
	sessionID := c.Param("sessionId")
	session, err := h.gatewayService.SessionStatus(sessionID)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "SESSION_STATUS_FAILED", err.Error())
		return
	}
	RespondOK(c, http.StatusOK, session)
}
