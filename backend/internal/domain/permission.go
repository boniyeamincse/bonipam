package domain

import "time"

type Permission struct {
	ID          string            `json:"id"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
	Constraints map[string]string `json:"constraints,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type CreatePermissionRequest struct {
	Resource    string            `json:"resource" binding:"required"`
	Action      string            `json:"action" binding:"required"`
	Constraints map[string]string `json:"constraints"`
}

type UpdatePermissionRequest struct {
	Constraints map[string]string `json:"constraints"`
}

type SetRolePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}
