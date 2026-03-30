package domain

type GroupRoleMapping struct {
	GroupID string   `json:"group_id"`
	RoleIDs []string `json:"role_ids"`
}

type ReconcileGroupRequest struct {
	MemberUserIDs []string `json:"member_user_ids" binding:"required"`
}

type GroupRoleActionResponse struct {
	GroupID      string   `json:"group_id"`
	RoleIDs      []string `json:"role_ids"`
	UpdatedUsers int      `json:"updated_users"`
}
