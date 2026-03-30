package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupMappingHandler struct {
	groupMappingService *service.GroupMappingService
}

func NewGroupMappingHandler(groupMappingService *service.GroupMappingService) *GroupMappingHandler {
	return &GroupMappingHandler{groupMappingService: groupMappingService}
}

func (h *GroupMappingHandler) RegisterRoutes(group *gin.RouterGroup) {
	groups := group.Group("/groups")
	groups.POST("/:groupId/roles/:roleId", h.AssignRole)
	groups.DELETE("/:groupId/roles/:roleId", h.UnassignRole)
	groups.GET("/:groupId/roles", h.GetGroupRoles)
	groups.POST("/:groupId/reconcile", h.Reconcile)
}

func (h *GroupMappingHandler) AssignRole(c *gin.Context) {
	result, err := h.groupMappingService.AssignRole(c.Param("groupId"), c.Param("roleId"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "GROUP_ROLE_ASSIGN_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}

func (h *GroupMappingHandler) UnassignRole(c *gin.Context) {
	result, err := h.groupMappingService.UnassignRole(c.Param("groupId"), c.Param("roleId"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "GROUP_ROLE_UNASSIGN_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}

func (h *GroupMappingHandler) GetGroupRoles(c *gin.Context) {
	mapping, err := h.groupMappingService.GetGroupRoles(c.Param("groupId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "GROUP_MAPPING_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, mapping)
}

func (h *GroupMappingHandler) Reconcile(c *gin.Context) {
	var req domain.ReconcileGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid reconcile request")
		return
	}

	result, err := h.groupMappingService.ReconcileGroupMembers(c.Param("groupId"), req.MemberUserIDs)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "GROUP_RECONCILE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}
