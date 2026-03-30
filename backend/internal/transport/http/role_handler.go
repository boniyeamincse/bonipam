package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleService *service.RoleService
}

func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) RegisterRoutes(group *gin.RouterGroup) {
	roles := group.Group("/roles")
	roles.POST("", h.CreateRole)
	roles.GET("", h.ListRoles)
	roles.GET("/:roleId", h.GetRole)
	roles.PUT("/:roleId", h.UpdateRole)
	roles.DELETE("/:roleId", h.DeleteRole)
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req domain.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid create role request")
		return
	}

	role, err := h.roleService.CreateRole(req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ROLE_CREATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusCreated, role)
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	RespondOK(c, http.StatusOK, h.roleService.ListRoles())
}

func (h *RoleHandler) GetRole(c *gin.Context) {
	role, err := h.roleService.GetRole(c.Param("roleId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "ROLE_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, role)
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	var req domain.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid update role request")
		return
	}

	role, err := h.roleService.UpdateRole(c.Param("roleId"), req)
	if err != nil {
		RespondError(c, http.StatusNotFound, "ROLE_UPDATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, role)
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	if err := h.roleService.DeleteRole(c.Param("roleId")); err != nil {
		RespondError(c, http.StatusBadRequest, "ROLE_DELETE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, gin.H{"message": "role deleted"})
}
