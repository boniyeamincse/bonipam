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
	vault.POST("/credentials/issue", h.IssueCredential)
	vault.GET("/leases/:leaseId", h.GetLeaseStatus)
	vault.POST("/leases/:leaseId/revoke", h.RevokeLease)
	vault.POST("/leases/revoke-by-target", h.RevokeLeasesByTarget)
	vault.POST("/rotation-policies", h.CreateRotationPolicy)
	vault.GET("/rotation-policies/:policyId", h.GetRotationPolicy)
	vault.POST("/rotation-policies/:policyId/trigger", h.TriggerRotation)
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

func (h *VaultHandler) IssueCredential(c *gin.Context) {
	var req domain.IssueCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid issue credential request")
		return
	}

	result, err := h.vaultService.IssueCredential(req)
	if err != nil {
		status := http.StatusBadRequest
		code := "CREDENTIAL_ISSUE_FAILED"
		if err.Error() == "unsupported target type" {
			code = "UNSUPPORTED_TARGET_TYPE"
		}
		RespondError(c, status, code, err.Error())
		return
	}

	RespondOK(c, http.StatusCreated, result)
}

func (h *VaultHandler) GetLeaseStatus(c *gin.Context) {
	result, err := h.vaultService.GetLeaseStatus(c.Param("leaseId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "LEASE_NOT_FOUND", err.Error())
		return
	}
	RespondOK(c, http.StatusOK, result)
}

func (h *VaultHandler) RevokeLease(c *gin.Context) {
	var req domain.RevokeLeaseRequest
	_ = c.ShouldBindJSON(&req) // reason is optional; ignore bind error
	err := h.vaultService.RevokeLease(c.Param("leaseId"), req.Reason)
	if err != nil {
		status := http.StatusBadRequest
		code := "LEASE_REVOKE_FAILED"
		if err.Error() == "lease not found" {
			status = http.StatusNotFound
			code = "LEASE_NOT_FOUND"
		}
		RespondError(c, status, code, err.Error())
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"revoked": true})
}

func (h *VaultHandler) RevokeLeasesByTarget(c *gin.Context) {
	var req domain.RevokeByTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid revoke-by-target request")
		return
	}
	result, err := h.vaultService.RevokeLeasesByTarget(req.TargetID, req.Reason)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "BULK_REVOKE_FAILED", err.Error())
		return
	}
	RespondOK(c, http.StatusOK, result)
}

func (h *VaultHandler) CreateRotationPolicy(c *gin.Context) {
	var req domain.CreateRotationPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid create rotation policy request")
		return
	}

	result, err := h.vaultService.CreateRotationPolicy(req)
	if err != nil {
		status := http.StatusBadRequest
		code := "ROTATION_POLICY_CREATE_FAILED"
		if err.Error() == "unsupported target type" {
			code = "UNSUPPORTED_TARGET_TYPE"
		}
		RespondError(c, status, code, err.Error())
		return
	}

	RespondOK(c, http.StatusCreated, result)
}

func (h *VaultHandler) GetRotationPolicy(c *gin.Context) {
	result, err := h.vaultService.GetRotationPolicy(c.Param("policyId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "ROTATION_POLICY_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}

func (h *VaultHandler) TriggerRotation(c *gin.Context) {
	result, err := h.vaultService.TriggerRotation(c.Param("policyId"))
	if err != nil {
		status := http.StatusBadRequest
		code := "ROTATION_TRIGGER_FAILED"
		if err.Error() == "rotation policy not found" {
			status = http.StatusNotFound
			code = "ROTATION_POLICY_NOT_FOUND"
		}
		RespondError(c, status, code, err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}
