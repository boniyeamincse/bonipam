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

type PermissionService struct {
	mu          sync.RWMutex
	permissions map[string]domain.Permission
}

func NewPermissionService() *PermissionService {
	return &PermissionService{permissions: make(map[string]domain.Permission)}
}

func (s *PermissionService) CreatePermission(req domain.CreatePermissionRequest) (domain.Permission, error) {
	resource := strings.TrimSpace(strings.ToLower(req.Resource))
	action := strings.TrimSpace(strings.ToLower(req.Action))
	if resource == "" || action == "" {
		return domain.Permission{}, fmt.Errorf("resource and action are required")
	}

	now := time.Now().UTC()
	perm := domain.Permission{
		ID:          "prm-" + uuid.NewString(),
		Resource:    resource,
		Action:      action,
		Constraints: req.Constraints,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.permissions {
		if existing.Resource == resource && existing.Action == action {
			return domain.Permission{}, fmt.Errorf("permission already exists for resource/action")
		}
	}

	s.permissions[perm.ID] = perm
	return perm, nil
}

func (s *PermissionService) ListPermissions() []domain.Permission {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Permission, 0, len(s.permissions))
	for _, permission := range s.permissions {
		result = append(result, permission)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Resource == result[j].Resource {
			return result[i].Action < result[j].Action
		}
		return result[i].Resource < result[j].Resource
	})

	return result
}

func (s *PermissionService) GetPermission(permissionID string) (domain.Permission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	permission, ok := s.permissions[permissionID]
	if !ok {
		return domain.Permission{}, fmt.Errorf("permission not found")
	}
	return permission, nil
}

func (s *PermissionService) UpdatePermission(permissionID string, req domain.UpdatePermissionRequest) (domain.Permission, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	permission, ok := s.permissions[permissionID]
	if !ok {
		return domain.Permission{}, fmt.Errorf("permission not found")
	}

	if req.Constraints != nil {
		permission.Constraints = req.Constraints
	}
	permission.UpdatedAt = time.Now().UTC()
	s.permissions[permissionID] = permission

	return permission, nil
}

func (s *PermissionService) DeletePermission(permissionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.permissions[permissionID]; !ok {
		return fmt.Errorf("permission not found")
	}
	delete(s.permissions, permissionID)
	return nil
}

func (s *PermissionService) Exists(permissionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.permissions[permissionID]
	return ok
}
