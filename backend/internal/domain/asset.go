package domain

import "time"

type AssetTransferStatus string

const (
	AssetTransferStatusPending  AssetTransferStatus = "pending"
	AssetTransferStatusApproved AssetTransferStatus = "approved"
	AssetTransferStatusRejected AssetTransferStatus = "rejected"
)

type Asset struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Type               string                 `json:"type"`
	Host               string                 `json:"host"`
	Port               int                    `json:"port"`
	Environment        string                 `json:"environment,omitempty"`
	Owner              string                 `json:"owner,omitempty"`
	Criticality        string                 `json:"criticality,omitempty"`
	Groups             []string               `json:"groups,omitempty"`
	ConnectionMetadata map[string]interface{} `json:"connection_metadata"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type CreateAssetRequest struct {
	Name               string                 `json:"name" binding:"required"`
	Type               string                 `json:"type" binding:"required"`
	Host               string                 `json:"host" binding:"required"`
	Port               int                    `json:"port"`
	Environment        string                 `json:"environment"`
	Owner              string                 `json:"owner"`
	Criticality        string                 `json:"criticality"`
	Groups             []string               `json:"groups"`
	ConnectionMetadata map[string]interface{} `json:"connection_metadata" binding:"required"`
}

type UpdateAssetTaggingRequest struct {
	Environment string   `json:"environment"`
	Owner       string   `json:"owner"`
	Criticality string   `json:"criticality"`
	Groups      []string `json:"groups"`
}

type TestAssetConnectionRequest struct {
	TimeoutSeconds int `json:"timeout_seconds"`
}

type TestAssetConnectionResult struct {
	AssetID      string    `json:"asset_id"`
	Status       string    `json:"status"`
	LatencyMs    int       `json:"latency_ms"`
	CheckedAt    time.Time `json:"checked_at"`
	Message      string    `json:"message"`
	Protocol     string    `json:"protocol"`
	TimeoutUsedS int       `json:"timeout_used_seconds"`
}

type AssignAssetOwnerRequest struct {
	Owner      string `json:"owner" binding:"required"`
	AssignedBy string `json:"assigned_by" binding:"required"`
	Reason     string `json:"reason"`
}

type RequestAssetOwnershipTransferRequest struct {
	NewOwner    string `json:"new_owner" binding:"required"`
	RequestedBy string `json:"requested_by" binding:"required"`
	Reason      string `json:"reason"`
}

type ReviewAssetOwnershipTransferRequest struct {
	Approved   bool   `json:"approved"`
	ReviewedBy string `json:"reviewed_by" binding:"required"`
	Comment    string `json:"comment"`
}

type AssetOwnershipTransfer struct {
	ID          string              `json:"id"`
	AssetID     string              `json:"asset_id"`
	FromOwner   string              `json:"from_owner"`
	ToOwner     string              `json:"to_owner"`
	RequestedBy string              `json:"requested_by"`
	ReviewedBy  string              `json:"reviewed_by,omitempty"`
	Reason      string              `json:"reason,omitempty"`
	Comment     string              `json:"comment,omitempty"`
	Status      AssetTransferStatus `json:"status"`
	RequestedAt time.Time           `json:"requested_at"`
	ReviewedAt  *time.Time          `json:"reviewed_at,omitempty"`
}

type AssetAuditEvent struct {
	ID        string                 `json:"id"`
	AssetID   string                 `json:"asset_id"`
	Type      string                 `json:"type"`
	Actor     string                 `json:"actor"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type AssetImportRequest struct {
	Assets  []CreateAssetRequest `json:"assets"`
	CSVData string               `json:"csv_data"`
}

type AssetImportIssue struct {
	Index  int    `json:"index"`
	Name   string `json:"name,omitempty"`
	Reason string `json:"reason"`
}

type AssetImportResult struct {
	TotalRows int                `json:"total_rows"`
	Imported  []Asset            `json:"imported"`
	Skipped   []AssetImportIssue `json:"skipped"`
	Failed    []AssetImportIssue `json:"failed"`
}
