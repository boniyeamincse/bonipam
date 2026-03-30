package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) RegisterRoutes(group *gin.RouterGroup) {
	users := group.Group("/users")
	users.POST("", h.CreateUser)
	users.GET("", h.ListUsers)
	users.GET("/:userId", h.GetUser)
	users.PUT("/:userId", h.UpdateUser)
	users.DELETE("/:userId", h.DeleteUser)
	users.POST("/:userId/restore", h.RestoreUser)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid create user request")
		return
	}

	user, err := h.userService.CreateUser(req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "USER_CREATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusCreated, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	includeDeleted, _ := strconv.ParseBool(c.DefaultQuery("include_deleted", "false"))
	users := h.userService.ListUsers(includeDeleted)
	RespondOK(c, http.StatusOK, users)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	includeDeleted, _ := strconv.ParseBool(c.DefaultQuery("include_deleted", "false"))
	user, err := h.userService.GetUser(c.Param("userId"), includeDeleted)
	if err != nil {
		RespondError(c, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid update user request")
		return
	}

	user, err := h.userService.UpdateUser(c.Param("userId"), req)
	if err != nil {
		RespondError(c, http.StatusNotFound, "USER_UPDATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	user, err := h.userService.SoftDeleteUser(c.Param("userId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "USER_DELETE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, user)
}

func (h *UserHandler) RestoreUser(c *gin.Context) {
	user, err := h.userService.RestoreUser(c.Param("userId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "USER_RESTORE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, user)
}
