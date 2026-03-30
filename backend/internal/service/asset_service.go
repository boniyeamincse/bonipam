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
	mu     sync.RWMutex
	assets map[string]domain.Asset
}

func NewAssetService() *AssetService {
	return &AssetService{assets: make(map[string]domain.Asset)}
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

	now := time.Now().UTC()
	asset := domain.Asset{
		ID:                 "ast-" + uuid.NewString(),
		Name:               name,
		Type:               assetType,
		Host:               host,
		Port:               port,
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

func (s *AssetService) ListAssets() []domain.Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Asset, 0, len(s.assets))
	for _, asset := range s.assets {
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
