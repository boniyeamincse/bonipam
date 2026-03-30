package domain

type StartSessionRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	AssetID    string `json:"asset_id" binding:"required"`
	JITGrantID string `json:"jit_grant_id" binding:"required"`
	Protocol   string `json:"protocol" binding:"required"`
}

type GatewaySession struct {
	SessionID      string `json:"session_id"`
	GatewayHost    string `json:"gateway_host"`
	ConnectCommand string `json:"connect_command"`
	Status         string `json:"status"`
}
