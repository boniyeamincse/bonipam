package domain

import "time"

type User struct {
	ID          string     `json:"id"`
	ExternalID  string     `json:"external_id,omitempty"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	Groups      []string   `json:"groups,omitempty"`
	Roles       []string   `json:"roles,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateUserRequest struct {
	ExternalID  string   `json:"external_id"`
	Email       string   `json:"email" binding:"required"`
	DisplayName string   `json:"display_name" binding:"required"`
	Groups      []string `json:"groups"`
}

type UpdateUserRequest struct {
	DisplayName string   `json:"display_name"`
	Groups      []string `json:"groups"`
	Status      string   `json:"status"`
}
