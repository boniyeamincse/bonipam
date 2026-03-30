package service

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"boni-pam/internal/domain"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schemas/*.json
var schemaFiles embed.FS

type PolicyService struct {
	mu       sync.RWMutex
	policies map[uuid.UUID]domain.Policy
}

func NewPolicyService() *PolicyService {
	return &PolicyService{
		policies: make(map[uuid.UUID]domain.Policy),
	}
}

func (s *PolicyService) CreatePolicy(ctx context.Context, req domain.CreatePolicyRequest) (domain.Policy, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.Policy{}, fmt.Errorf("name is required")
	}

	if err := s.ValidatePolicy(ctx, req.Definition); err != nil {
		return domain.Policy{}, err
	}

	createdBy := uuid.Nil
	if strings.TrimSpace(req.CreatedBy) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(req.CreatedBy))
		if err != nil {
			return domain.Policy{}, fmt.Errorf("created_by must be a valid uuid")
		}
		createdBy = parsed
	}

	now := time.Now().UTC()
	policy := domain.Policy{
		ID:         uuid.New(),
		Name:       name,
		Version:    1,
		Status:     domain.PolicyStatusDraft,
		Definition: req.Definition,
		CreatedBy:  createdBy,
		CreatedAt:  now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.policies {
		if strings.EqualFold(existing.Name, policy.Name) {
			return domain.Policy{}, fmt.Errorf("policy name already exists")
		}
	}

	s.policies[policy.ID] = policy
	return policy, nil
}

func (s *PolicyService) ListPolicies() []domain.Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Policy, 0, len(s.policies))
	for _, policy := range s.policies {
		result = append(result, policy)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result
}

func (s *PolicyService) GetPolicy(policyID string) (domain.Policy, error) {
	id, err := uuid.Parse(strings.TrimSpace(policyID))
	if err != nil {
		return domain.Policy{}, fmt.Errorf("invalid policy id")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, ok := s.policies[id]
	if !ok {
		return domain.Policy{}, fmt.Errorf("policy not found")
	}

	return policy, nil
}

func (s *PolicyService) UpdatePolicy(ctx context.Context, policyID string, req domain.UpdatePolicyRequest) (domain.Policy, error) {
	id, err := uuid.Parse(strings.TrimSpace(policyID))
	if err != nil {
		return domain.Policy{}, fmt.Errorf("invalid policy id")
	}

	s.mu.Lock()
	policy, ok := s.policies[id]
	if !ok {
		s.mu.Unlock()
		return domain.Policy{}, fmt.Errorf("policy not found")
	}

	if req.Name != "" {
		name := strings.TrimSpace(req.Name)
		if name == "" {
			s.mu.Unlock()
			return domain.Policy{}, fmt.Errorf("name cannot be empty")
		}
		for otherID, existing := range s.policies {
			if otherID != id && strings.EqualFold(existing.Name, name) {
				s.mu.Unlock()
				return domain.Policy{}, fmt.Errorf("policy name already exists")
			}
		}
		policy.Name = name
	}

	if req.Definition != nil {
		if err := s.ValidatePolicy(ctx, *req.Definition); err != nil {
			s.mu.Unlock()
			return domain.Policy{}, err
		}
		policy.Definition = *req.Definition
		policy.Version++
	}

	s.policies[id] = policy
	s.mu.Unlock()

	return policy, nil
}

// ValidatePolicy checks if a policy definition conforms to the JSON schema
func (s *PolicyService) ValidatePolicy(ctx context.Context, definition domain.PolicyDefinition) error {
	// 1. Determine schema file based on schema_version
	var schemaPath string
	switch definition.SchemaVersion {
	case "1.0":
		schemaPath = "schemas/policy_v1.json"
	default:
		return fmt.Errorf("unsupported schema version: %s", definition.SchemaVersion)
	}

	// 2. Load schema from embedded FS
	schemaData, err := schemaFiles.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to load schema file %s: %w", schemaPath, err)
	}

	// 3. Convert policy definition to JSON for validation
	definitionJSON, err := json.Marshal(definition)
	if err != nil {
		return fmt.Errorf("failed to marshal policy definition: %w", err)
	}

	// 4. Perform validation
	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	documentLoader := gojsonschema.NewBytesLoader(definitionJSON)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation execution failed: %w", err)
	}

	if !result.Valid() {
		var report string
		for _, desc := range result.Errors() {
			report += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("policy definition is invalid:\n%s", report)
	}

	return nil
}
