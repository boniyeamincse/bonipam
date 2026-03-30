package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PolicyHandler struct {
	policyService *service.PolicyService
}

func NewPolicyHandler(policyService *service.PolicyService) *PolicyHandler {
	return &PolicyHandler{policyService: policyService}
}

func (h *PolicyHandler) RegisterRoutes(group *gin.RouterGroup) {
	policies := group.Group("/policies")
	policies.POST("", h.CreatePolicy)
	policies.GET("", h.ListPolicies)
	policies.GET("/:policyId", h.GetPolicy)
	policies.PUT("/:policyId", h.UpdatePolicy)
	policies.POST("/:policyId/evaluate", h.EvaluatePolicy)
}

func (h *PolicyHandler) CreatePolicy(c *gin.Context) {
	var req domain.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid create policy request")
		return
	}

	policy, err := h.policyService.CreatePolicy(c.Request.Context(), req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "POLICY_CREATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusCreated, policy)
}

func (h *PolicyHandler) ListPolicies(c *gin.Context) {
	RespondOK(c, http.StatusOK, h.policyService.ListPolicies())
}

func (h *PolicyHandler) GetPolicy(c *gin.Context) {
	policy, err := h.policyService.GetPolicy(c.Param("policyId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "POLICY_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, policy)
}

func (h *PolicyHandler) UpdatePolicy(c *gin.Context) {
	var req domain.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid update policy request")
		return
	}

	policy, err := h.policyService.UpdatePolicy(c.Request.Context(), c.Param("policyId"), req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "POLICY_UPDATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, policy)
}

func (h *PolicyHandler) EvaluatePolicy(c *gin.Context) {
	var req domain.PolicyEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid policy evaluation request")
		return
	}

	result, err := h.policyService.EvaluatePolicy(c.Request.Context(), c.Param("policyId"), req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "POLICY_EVALUATION_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}
