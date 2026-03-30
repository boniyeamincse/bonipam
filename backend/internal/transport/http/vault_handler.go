package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VaultHandler struct {
	vaultService *service.VaultService
}

func NewVaultHandler(vaultService *service.VaultService) *VaultHandler {
	return &VaultHandler{vaultService: vaultService}
}

func (h *VaultHandler) RegisterRoutes(group *gin.RouterGroup) {
	vault := group.Group("/vault")
	vault.POST("/secrets", h.StoreSecret)
	vault.GET("/secrets/:secretId", h.GetSecret)
}

func (h *VaultHandler) StoreSecret(c *gin.Context) {
	var req domain.CreateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid create secret request")
		return
	}

	result, err := h.vaultService.StoreSecret(req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "SECRET_STORE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusCreated, result)
}

func (h *VaultHandler) GetSecret(c *gin.Context) {
	result, err := h.vaultService.GetSecret(c.Param("secretId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "SECRET_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}
