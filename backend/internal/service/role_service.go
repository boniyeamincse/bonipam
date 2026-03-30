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

type RoleService struct {
	mu    sync.RWMutex
	roles map[string]domain.Role
}

func NewRoleService() *RoleService {
	return &RoleService{roles: make(map[string]domain.Role)}
}

func (s *RoleService) CreateRole(req domain.CreateRoleRequest) (domain.Role, error) {
	name := strings.TrimSpace(strings.ToLower(req.Name))
	if name == "" {
		return domain.Role{}, fmt.Errorf("role name is required")
	}

	now := time.Now().UTC()
	role := domain.Role{
		ID:          "rol-" + uuid.NewString(),
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Permissions: dedupe(req.Permissions),
		IsSystem:    req.IsSystem,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.roles {
		if existing.Name == role.Name {
			return domain.Role{}, fmt.Errorf("role name already exists")
		}
	}

	s.roles[role.ID] = role
	return role, nil
}

func (s *RoleService) ListRoles() []domain.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Role, 0, len(s.roles))
	for _, role := range s.roles {
		result = append(result, role)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result
}

func (s *RoleService) GetRole(roleID string) (domain.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, ok := s.roles[roleID]
	if !ok {
		return domain.Role{}, fmt.Errorf("role not found")
	}

	return role, nil
}

func (s *RoleService) UpdateRole(roleID string, req domain.UpdateRoleRequest) (domain.Role, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, ok := s.roles[roleID]
	if !ok {
		return domain.Role{}, fmt.Errorf("role not found")
	}

	if req.Description != "" {
		role.Description = strings.TrimSpace(req.Description)
	}
	if req.Permissions != nil {
		role.Permissions = dedupe(req.Permissions)
	}
	role.UpdatedAt = time.Now().UTC()

	s.roles[roleID] = role
	return role, nil
}

func (s *RoleService) DeleteRole(roleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, ok := s.roles[roleID]
	if !ok {
		return fmt.Errorf("role not found")
	}
	if role.IsSystem {
		return fmt.Errorf("system role cannot be deleted")
	}

	delete(s.roles, roleID)
	return nil
}

func dedupe(values []string) []string {
	if values == nil {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}
