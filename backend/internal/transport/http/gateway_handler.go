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
	gateway.POST("", h.InitiateSession)
	gateway.POST("/start", h.InitiateSession) // legacy alias
	gateway.GET("/:sessionId", h.GetSession)
	gateway.GET("/:sessionId/status", h.GetSession) // legacy alias
	gateway.POST("/:sessionId/terminate", h.TerminateSession)
}

func (h *GatewayHandler) InitiateSession(c *gin.Context) {
	var req domain.InitiateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid session initiate payload")
		return
	}
	session, err := h.gatewayService.InitiateSession(req)
	if err != nil {
		status := http.StatusBadRequest
		code := "SESSION_INITIATE_FAILED"
		if err.Error()[:19] == "unsupported protocol" {
			code = "UNSUPPORTED_PROTOCOL"
		}
		RespondError(c, status, code, err.Error())
		return
	}
	RespondOK(c, http.StatusCreated, session)
}

func (h *GatewayHandler) GetSession(c *gin.Context) {
	result, err := h.gatewayService.GetSession(c.Param("sessionId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "SESSION_NOT_FOUND", err.Error())
		return
	}
	RespondOK(c, http.StatusOK, result)
}

func (h *GatewayHandler) TerminateSession(c *gin.Context) {
	var req domain.TerminateSessionRequest
	_ = c.ShouldBindJSON(&req) // reason is optional
	err := h.gatewayService.TerminateSession(c.Param("sessionId"), req.Reason)
	if err != nil {
		status := http.StatusBadRequest
		code := "SESSION_TERMINATE_FAILED"
		if err.Error() == "session not found" {
			status = http.StatusNotFound
			code = "SESSION_NOT_FOUND"
		}
		RespondError(c, status, code, err.Error())
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"terminated": true})
}
