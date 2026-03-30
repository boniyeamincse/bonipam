package domain

import "time"

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
