package service

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"reflect"
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
	mu        sync.RWMutex
	policies  map[uuid.UUID]domain.Policy
	versions  map[uuid.UUID]map[int]domain.PolicyDefinition
	published map[uuid.UUID]int
}

func NewPolicyService() *PolicyService {
	return &PolicyService{
		policies:  make(map[uuid.UUID]domain.Policy),
		versions:  make(map[uuid.UUID]map[int]domain.PolicyDefinition),
		published: make(map[uuid.UUID]int),
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
	s.versions[policy.ID] = map[int]domain.PolicyDefinition{1: policy.Definition}
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
		policy.Status = domain.PolicyStatusDraft
		policy.PublishedAt = nil
		if s.versions[id] == nil {
			s.versions[id] = make(map[int]domain.PolicyDefinition)
		}
		s.versions[id][policy.Version] = policy.Definition
	}

	s.policies[id] = policy
	s.mu.Unlock()

	return policy, nil
}

func (s *PolicyService) PublishPolicy(policyID string) (domain.PublishPolicyResponse, error) {
	id, err := uuid.Parse(strings.TrimSpace(policyID))
	if err != nil {
		return domain.PublishPolicyResponse{}, fmt.Errorf("invalid policy id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[id]
	if !ok {
		return domain.PublishPolicyResponse{}, fmt.Errorf("policy not found")
	}

	now := time.Now().UTC()
	policy.Status = domain.PolicyStatusPublished
	policy.PublishedAt = &now
	s.policies[id] = policy
	s.published[id] = policy.Version

	return domain.PublishPolicyResponse{
		PolicyID:    policy.ID,
		Version:     policy.Version,
		Status:      policy.Status,
		PublishedAt: now,
	}, nil
}

func (s *PolicyService) RollbackPolicy(ctx context.Context, policyID string, targetVersion int) (domain.Policy, error) {
	_ = ctx

	id, err := uuid.Parse(strings.TrimSpace(policyID))
	if err != nil {
		return domain.Policy{}, fmt.Errorf("invalid policy id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[id]
	if !ok {
		return domain.Policy{}, fmt.Errorf("policy not found")
	}

	availableVersions := s.versions[id]
	if len(availableVersions) == 0 {
		return domain.Policy{}, fmt.Errorf("no policy versions available for rollback")
	}

	if targetVersion <= 0 {
		targetVersion = s.previousVersion(policy.ID, policy.Version)
		if targetVersion == 0 {
			return domain.Policy{}, fmt.Errorf("no previous version available for rollback")
		}
	}

	definition, ok := availableVersions[targetVersion]
	if !ok {
		return domain.Policy{}, fmt.Errorf("target version not found")
	}

	policy.Definition = definition
	policy.Version++
	now := time.Now().UTC()
	policy.Status = domain.PolicyStatusPublished
	policy.PublishedAt = &now

	if s.versions[id] == nil {
		s.versions[id] = make(map[int]domain.PolicyDefinition)
	}
	s.versions[id][policy.Version] = policy.Definition
	s.published[id] = policy.Version
	s.policies[id] = policy

	return policy, nil
}

func (s *PolicyService) EvaluatePolicy(ctx context.Context, policyID string, req domain.PolicyEvaluationRequest) (domain.PolicyEvaluationResponse, error) {
	_ = ctx

	policy, err := s.GetPolicy(policyID)
	if err != nil {
		return domain.PolicyEvaluationResponse{}, err
	}

	subject := strings.TrimSpace(req.Subject)
	resource := strings.TrimSpace(req.Resource)
	action := strings.TrimSpace(req.Action)
	if subject == "" || resource == "" || action == "" {
		return domain.PolicyEvaluationResponse{}, fmt.Errorf("subject, resource, and action are required")
	}

	attributes := NewContextResolverService().Resolve(req)

	decision := strings.ToLower(policy.Definition.DefaultEffect)
	if decision != "allow" {
		decision = "deny"
	}

	obligations := make([]string, 0)
	obligationSet := make(map[string]struct{})
	matchedRuleIDs := make([]string, 0)
	hasAllow := false
	hasDeny := false

	for i, rule := range policy.Definition.Rules {
		if !matchesAny(rule.Subjects, subject) {
			continue
		}
		if !matchesAny(rule.Resources, resource) {
			continue
		}
		if !matchesAny(rule.Actions, action) {
			continue
		}
		if !conditionsMatch(rule.Conditions, attributes) {
			continue
		}

		ruleID := strings.TrimSpace(rule.ID)
		if ruleID == "" {
			ruleID = fmt.Sprintf("rule-%d", i+1)
		}
		matchedRuleIDs = append(matchedRuleIDs, ruleID)

		effect := strings.ToLower(strings.TrimSpace(rule.Effect))
		if effect == "deny" {
			hasDeny = true
		}
		if effect == "allow" {
			hasAllow = true
		}

		for _, obligation := range rule.Obligations {
			normalized := strings.TrimSpace(obligation)
			if normalized == "" {
				continue
			}
			if _, seen := obligationSet[normalized]; seen {
				continue
			}
			obligationSet[normalized] = struct{}{}
			obligations = append(obligations, normalized)
		}
	}

	if hasDeny {
		decision = "deny"
	} else if hasAllow {
		decision = "allow"
	}

	return domain.PolicyEvaluationResponse{
		PolicyID:      policy.ID,
		Decision:      decision,
		Obligations:   obligations,
		MatchedRuleID: matchedRuleIDs,
	}, nil
}

// ValidatePolicy checks if a policy definition conforms to the JSON schema
func (s *PolicyService) ValidatePolicy(ctx context.Context, definition domain.PolicyDefinition) error {
	_ = ctx
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

func matchesAny(patterns []string, value string) bool {
	if len(patterns) == 0 {
		return false
	}
	for _, pattern := range patterns {
		if matchesPattern(strings.TrimSpace(pattern), value) {
			return true
		}
	}
	return false
}

func matchesPattern(pattern, value string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(value, prefix)
	}
	return pattern == value
}

func conditionsMatch(conditions []domain.PolicyCondition, attrs map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}
	if attrs == nil {
		return false
	}

	for _, condition := range conditions {
		actual, ok := attrs[condition.Attribute]
		if !ok {
			return false
		}
		if !evaluateCondition(actual, condition.Operator, condition.Value) {
			return false
		}
	}

	return true
}

func evaluateCondition(actual interface{}, operator string, expected interface{}) bool {
	operator = strings.ToLower(strings.TrimSpace(operator))
	switch operator {
	case "eq":
		return looselyEqual(actual, expected)
	case "neq":
		return !looselyEqual(actual, expected)
	case "gt", "gte", "lt", "lte":
		cmp, ok := compareScalars(actual, expected)
		if !ok {
			return false
		}
		switch operator {
		case "gt":
			return cmp > 0
		case "gte":
			return cmp >= 0
		case "lt":
			return cmp < 0
		default:
			return cmp <= 0
		}
	case "in", "nin":
		items, ok := toInterfaceSlice(expected)
		if !ok {
			return false
		}
		found := false
		for _, item := range items {
			if looselyEqual(actual, item) {
				found = true
				break
			}
		}
		if operator == "in" {
			return found
		}
		return !found
	case "between":
		items, ok := toInterfaceSlice(expected)
		if !ok || len(items) != 2 {
			return false
		}
		lowCmp, okLow := compareScalars(actual, items[0])
		highCmp, okHigh := compareScalars(actual, items[1])
		if !okLow || !okHigh {
			return false
		}
		return lowCmp >= 0 && highCmp <= 0
	default:
		return false
	}
}

func compareScalars(a, b interface{}) (int, bool) {
	if af, ok := toFloat64(a); ok {
		bf, ok := toFloat64(b)
		if !ok {
			return 0, false
		}
		switch {
		case af < bf:
			return -1, true
		case af > bf:
			return 1, true
		default:
			return 0, true
		}
	}

	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		switch {
		case as < bs:
			return -1, true
		case as > bs:
			return 1, true
		default:
			return 0, true
		}
	}

	return 0, false
}

func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func looselyEqual(a, b interface{}) bool {
	if af, ok := toFloat64(a); ok {
		bf, ok := toFloat64(b)
		if ok {
			return af == bf
		}
	}
	return reflect.DeepEqual(a, b)
}

func toInterfaceSlice(value interface{}) ([]interface{}, bool) {
	if value == nil {
		return nil, false
	}

	if items, ok := value.([]interface{}); ok {
		return items, true
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}

	result := make([]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result[i] = rv.Index(i).Interface()
	}

	return result, true
}

func (s *PolicyService) previousVersion(policyID uuid.UUID, currentVersion int) int {
	available := s.versions[policyID]
	if len(available) == 0 {
		return 0
	}

	best := 0
	for version := range available {
		if version < currentVersion && version > best {
			best = version
		}
	}

	return best
}
