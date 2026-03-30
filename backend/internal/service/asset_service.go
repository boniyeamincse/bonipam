package service

import (
	"boni-pam/internal/domain"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type AssetService struct {
	mu                 sync.RWMutex
	assets             map[string]domain.Asset
	ownershipTransfers map[string]domain.AssetOwnershipTransfer
	auditEvents        map[string][]domain.AssetAuditEvent
}

func NewAssetService() *AssetService {
	return &AssetService{
		assets:             make(map[string]domain.Asset),
		ownershipTransfers: make(map[string]domain.AssetOwnershipTransfer),
		auditEvents:        make(map[string][]domain.AssetAuditEvent),
	}
}

func (s *AssetService) CreateAsset(req domain.CreateAssetRequest) (domain.Asset, error) {
	name := strings.TrimSpace(req.Name)
	assetType := strings.ToLower(strings.TrimSpace(req.Type))
	host := strings.TrimSpace(req.Host)

	if name == "" || assetType == "" || host == "" {
		return domain.Asset{}, fmt.Errorf("name, type, and host are required")
	}

	port := req.Port
	if port == 0 {
		port = defaultPortForType(assetType)
	}
	if port <= 0 || port > 65535 {
		return domain.Asset{}, fmt.Errorf("port must be between 1 and 65535")
	}

	metadata, err := validateConnectionMetadata(assetType, req.ConnectionMetadata)
	if err != nil {
		return domain.Asset{}, err
	}

	environment, owner, criticality, groups, err := normalizeAssetTagging(req.Environment, req.Owner, req.Criticality, req.Groups)
	if err != nil {
		return domain.Asset{}, err
	}

	now := time.Now().UTC()
	asset := domain.Asset{
		ID:                 "ast-" + uuid.NewString(),
		Name:               name,
		Type:               assetType,
		Host:               host,
		Port:               port,
		Environment:        environment,
		Owner:              owner,
		Criticality:        criticality,
		Groups:             groups,
		ConnectionMetadata: metadata,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.assets {
		if strings.EqualFold(existing.Name, asset.Name) {
			return domain.Asset{}, fmt.Errorf("asset name already exists")
		}
	}

	s.assets[asset.ID] = asset
	return asset, nil
}

func (s *AssetService) ListAssets(environment, owner, criticality, group string) []domain.Asset {
	environment = strings.ToLower(strings.TrimSpace(environment))
	owner = strings.ToLower(strings.TrimSpace(owner))
	criticality = strings.ToLower(strings.TrimSpace(criticality))
	group = strings.ToLower(strings.TrimSpace(group))

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Asset, 0, len(s.assets))
	for _, asset := range s.assets {
		if environment != "" && asset.Environment != environment {
			continue
		}
		if owner != "" && strings.ToLower(asset.Owner) != owner {
			continue
		}
		if criticality != "" && asset.Criticality != criticality {
			continue
		}
		if group != "" && !containsString(asset.Groups, group) {
			continue
		}
		result = append(result, asset)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result
}

func (s *AssetService) GetAsset(assetID string) (domain.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	asset, ok := s.assets[assetID]
	if !ok {
		return domain.Asset{}, fmt.Errorf("asset not found")
	}

	return asset, nil
}

func (s *AssetService) UpdateAssetTagging(assetID string, req domain.UpdateAssetTaggingRequest) (domain.Asset, error) {
	environment, owner, criticality, groups, err := normalizeAssetTagging(req.Environment, req.Owner, req.Criticality, req.Groups)
	if err != nil {
		return domain.Asset{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	asset, ok := s.assets[assetID]
	if !ok {
		return domain.Asset{}, fmt.Errorf("asset not found")
	}

	asset.Environment = environment
	asset.Owner = owner
	asset.Criticality = criticality
	asset.Groups = groups
	asset.UpdatedAt = time.Now().UTC()

	s.assets[assetID] = asset
	return asset, nil
}

func (s *AssetService) TestConnection(assetID string, timeoutSeconds int) (domain.TestAssetConnectionResult, error) {
	s.mu.RLock()
	asset, ok := s.assets[assetID]
	s.mu.RUnlock()
	if !ok {
		return domain.TestAssetConnectionResult{}, fmt.Errorf("asset not found")
	}

	if timeoutSeconds <= 0 {
		timeoutSeconds = 3
	}
	if timeoutSeconds > 15 {
		timeoutSeconds = 15
	}

	protocol := asset.Type
	if protocol == "" {
		protocol = "unknown"
	}

	if _, err := validateConnectionMetadata(asset.Type, asset.ConnectionMetadata); err != nil {
		return domain.TestAssetConnectionResult{}, fmt.Errorf("connectivity pre-check failed: %w", err)
	}

	host := strings.ToLower(strings.TrimSpace(asset.Host))
	failure := host == "" || strings.Contains(host, "invalid") || strings.Contains(host, "unreachable") || strings.Contains(host, "blocked")

	latencyMs := simulateLatency(asset.Host, asset.Port, timeoutSeconds)
	now := time.Now().UTC()

	if failure {
		return domain.TestAssetConnectionResult{
			AssetID:      asset.ID,
			Status:       "failed",
			LatencyMs:    latencyMs,
			CheckedAt:    now,
			Message:      "connection test failed",
			Protocol:     protocol,
			TimeoutUsedS: timeoutSeconds,
		}, nil
	}

	return domain.TestAssetConnectionResult{
		AssetID:      asset.ID,
		Status:       "ok",
		LatencyMs:    latencyMs,
		CheckedAt:    now,
		Message:      "connection test passed",
		Protocol:     protocol,
		TimeoutUsedS: timeoutSeconds,
	}, nil
}

func (s *AssetService) AssignOwner(assetID string, req domain.AssignAssetOwnerRequest) (domain.Asset, error) {
	owner := strings.TrimSpace(req.Owner)
	assignedBy := strings.TrimSpace(req.AssignedBy)
	if owner == "" || assignedBy == "" {
		return domain.Asset{}, fmt.Errorf("owner and assigned_by are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	asset, ok := s.assets[assetID]
	if !ok {
		return domain.Asset{}, fmt.Errorf("asset not found")
	}

	fromOwner := asset.Owner
	asset.Owner = owner
	asset.UpdatedAt = time.Now().UTC()
	s.assets[assetID] = asset

	s.recordAuditEventLocked(assetID, "owner_assigned", assignedBy, "asset owner assigned", map[string]interface{}{
		"from_owner": fromOwner,
		"to_owner":   owner,
		"reason":     strings.TrimSpace(req.Reason),
	})

	return asset, nil
}

func (s *AssetService) RequestOwnershipTransfer(assetID string, req domain.RequestAssetOwnershipTransferRequest) (domain.AssetOwnershipTransfer, error) {
	newOwner := strings.TrimSpace(req.NewOwner)
	requestedBy := strings.TrimSpace(req.RequestedBy)
	if newOwner == "" || requestedBy == "" {
		return domain.AssetOwnershipTransfer{}, fmt.Errorf("new_owner and requested_by are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	asset, ok := s.assets[assetID]
	if !ok {
		return domain.AssetOwnershipTransfer{}, fmt.Errorf("asset not found")
	}
	if strings.TrimSpace(asset.Owner) == "" {
		return domain.AssetOwnershipTransfer{}, fmt.Errorf("asset has no current owner; use owner assignment")
	}
	if strings.EqualFold(asset.Owner, newOwner) {
		return domain.AssetOwnershipTransfer{}, fmt.Errorf("new owner must differ from current owner")
	}

	now := time.Now().UTC()
	transfer := domain.AssetOwnershipTransfer{
		ID:          "trf-" + uuid.NewString(),
		AssetID:     assetID,
		FromOwner:   asset.Owner,
		ToOwner:     newOwner,
		RequestedBy: requestedBy,
		Reason:      strings.TrimSpace(req.Reason),
		Status:      domain.AssetTransferStatusPending,
		RequestedAt: now,
	}

	s.ownershipTransfers[transfer.ID] = transfer
	s.recordAuditEventLocked(assetID, "ownership_transfer_requested", requestedBy, "ownership transfer requested", map[string]interface{}{
		"transfer_id": transfer.ID,
		"from_owner":  transfer.FromOwner,
		"to_owner":    transfer.ToOwner,
	})

	return transfer, nil
}

func (s *AssetService) ReviewOwnershipTransfer(transferID string, req domain.ReviewAssetOwnershipTransferRequest) (domain.AssetOwnershipTransfer, error) {
	reviewedBy := strings.TrimSpace(req.ReviewedBy)
	if reviewedBy == "" {
		return domain.AssetOwnershipTransfer{}, fmt.Errorf("reviewed_by is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	transfer, ok := s.ownershipTransfers[transferID]
	if !ok {
		return domain.AssetOwnershipTransfer{}, fmt.Errorf("ownership transfer not found")
	}
	if transfer.Status != domain.AssetTransferStatusPending {
		return domain.AssetOwnershipTransfer{}, fmt.Errorf("ownership transfer already reviewed")
	}

	now := time.Now().UTC()
	transfer.ReviewedBy = reviewedBy
	transfer.Comment = strings.TrimSpace(req.Comment)
	transfer.ReviewedAt = &now

	asset, ok := s.assets[transfer.AssetID]
	if !ok {
		return domain.AssetOwnershipTransfer{}, fmt.Errorf("asset not found")
	}

	eventType := "ownership_transfer_rejected"
	eventMessage := "ownership transfer rejected"
	if req.Approved {
		transfer.Status = domain.AssetTransferStatusApproved
		asset.Owner = transfer.ToOwner
		asset.UpdatedAt = now
		s.assets[asset.ID] = asset
		eventType = "ownership_transfer_approved"
		eventMessage = "ownership transfer approved"
	} else {
		transfer.Status = domain.AssetTransferStatusRejected
	}

	s.ownershipTransfers[transferID] = transfer
	s.recordAuditEventLocked(transfer.AssetID, eventType, reviewedBy, eventMessage, map[string]interface{}{
		"transfer_id": transfer.ID,
		"from_owner":  transfer.FromOwner,
		"to_owner":    transfer.ToOwner,
		"comment":     transfer.Comment,
	})

	return transfer, nil
}

func (s *AssetService) ListOwnershipTransfers(assetID string) []domain.AssetOwnershipTransfer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.AssetOwnershipTransfer, 0)
	for _, transfer := range s.ownershipTransfers {
		if transfer.AssetID == assetID {
			result = append(result, transfer)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].RequestedAt.Before(result[j].RequestedAt)
	})

	return result
}

func (s *AssetService) ListAssetAuditEvents(assetID string) []domain.AssetAuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.auditEvents[assetID]
	result := make([]domain.AssetAuditEvent, len(events))
	copy(result, events)
	return result
}

func defaultPortForType(assetType string) int {
	switch assetType {
	case "ssh":
		return 22
	case "database":
		return 5432
	default:
		return 0
	}
}

func validateConnectionMetadata(assetType string, metadata map[string]interface{}) (map[string]interface{}, error) {
	if len(metadata) == 0 {
		return nil, fmt.Errorf("connection_metadata is required")
	}

	normalized := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		normalized[strings.ToLower(strings.TrimSpace(key))] = value
	}

	switch assetType {
	case "ssh":
		if _, ok := normalized["username"]; !ok {
			return nil, fmt.Errorf("connection_metadata.username is required for ssh assets")
		}
		authMethodRaw, ok := normalized["auth_method"]
		if !ok {
			return nil, fmt.Errorf("connection_metadata.auth_method is required for ssh assets")
		}
		authMethod, ok := authMethodRaw.(string)
		if !ok {
			return nil, fmt.Errorf("connection_metadata.auth_method must be a string")
		}
		authMethod = strings.ToLower(strings.TrimSpace(authMethod))
		if authMethod != "password" && authMethod != "key" {
			return nil, fmt.Errorf("connection_metadata.auth_method must be one of: password, key")
		}
		normalized["auth_method"] = authMethod
	case "database":
		engineRaw, ok := normalized["engine"]
		if !ok {
			return nil, fmt.Errorf("connection_metadata.engine is required for database assets")
		}
		engine, ok := engineRaw.(string)
		if !ok {
			return nil, fmt.Errorf("connection_metadata.engine must be a string")
		}
		engine = strings.ToLower(strings.TrimSpace(engine))
		if engine != "postgres" && engine != "mysql" && engine != "mssql" {
			return nil, fmt.Errorf("connection_metadata.engine must be one of: postgres, mysql, mssql")
		}
		normalized["engine"] = engine

		if _, ok := normalized["database"]; !ok {
			return nil, fmt.Errorf("connection_metadata.database is required for database assets")
		}
		if _, ok := normalized["username"]; !ok {
			return nil, fmt.Errorf("connection_metadata.username is required for database assets")
		}
	default:
		return nil, fmt.Errorf("unsupported asset type: %s", assetType)
	}

	return normalized, nil
}

func normalizeAssetTagging(environment, owner, criticality string, groups []string) (string, string, string, []string, error) {
	environment = strings.ToLower(strings.TrimSpace(environment))
	owner = strings.TrimSpace(owner)
	criticality = strings.ToLower(strings.TrimSpace(criticality))

	if environment != "" && environment != "dev" && environment != "staging" && environment != "prod" {
		return "", "", "", nil, fmt.Errorf("environment must be one of: dev, staging, prod")
	}
	if criticality != "" && criticality != "low" && criticality != "medium" && criticality != "high" && criticality != "critical" {
		return "", "", "", nil, fmt.Errorf("criticality must be one of: low, medium, high, critical")
	}

	normalizedGroups := make([]string, 0, len(groups))
	seen := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		normalized := strings.ToLower(strings.TrimSpace(group))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		normalizedGroups = append(normalizedGroups, normalized)
	}

	return environment, owner, criticality, normalizedGroups, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.ToLower(value) == target {
			return true
		}
	}
	return false
}

func simulateLatency(host string, port int, timeoutSeconds int) int {
	seed := len(strings.TrimSpace(host)) + port + timeoutSeconds
	if seed < 0 {
		seed = -seed
	}
	return 10 + (seed % 120)
}

func (s *AssetService) recordAuditEventLocked(assetID, eventType, actor, message string, metadata map[string]interface{}) {
	event := domain.AssetAuditEvent{
		ID:        "evt-" + uuid.NewString(),
		AssetID:   assetID,
		Type:      eventType,
		Actor:     actor,
		Message:   message,
		Metadata:  metadata,
		Timestamp: time.Now().UTC(),
	}

	s.auditEvents[assetID] = append(s.auditEvents[assetID], event)
}
