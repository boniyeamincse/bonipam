package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionService *service.PermissionService
}

func NewPermissionHandler(permissionService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

func (h *PermissionHandler) RegisterRoutes(group *gin.RouterGroup) {
	permissions := group.Group("/permissions")
	permissions.POST("", h.CreatePermission)
	permissions.GET("", h.ListPermissions)
	permissions.GET("/:permissionId", h.GetPermission)
	permissions.PUT("/:permissionId", h.UpdatePermission)
	permissions.DELETE("/:permissionId", h.DeletePermission)
}

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req domain.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid create permission request")
		return
	}

	permission, err := h.permissionService.CreatePermission(req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "PERMISSION_CREATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusCreated, permission)
}

func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	RespondOK(c, http.StatusOK, h.permissionService.ListPermissions())
}

func (h *PermissionHandler) GetPermission(c *gin.Context) {
	permission, err := h.permissionService.GetPermission(c.Param("permissionId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "PERMISSION_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, permission)
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	var req domain.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid update permission request")
		return
	}

	permission, err := h.permissionService.UpdatePermission(c.Param("permissionId"), req)
	if err != nil {
		RespondError(c, http.StatusNotFound, "PERMISSION_UPDATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, permission)
}

func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	if err := h.permissionService.DeletePermission(c.Param("permissionId")); err != nil {
		RespondError(c, http.StatusNotFound, "PERMISSION_DELETE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, gin.H{"message": "permission deleted"})
}
