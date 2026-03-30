package service

import (
	"boni-pam/internal/domain"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type GroupMappingService struct {
	mu          sync.RWMutex
	mappings    map[string]map[string]struct{}
	roleService *RoleService
	userService *UserService
}

func NewGroupMappingService(roleService *RoleService, userService *UserService) *GroupMappingService {
	return &GroupMappingService{
		mappings:    make(map[string]map[string]struct{}),
		roleService: roleService,
		userService: userService,
	}
}

func (s *GroupMappingService) AssignRole(groupID, roleID string) (domain.GroupRoleActionResponse, error) {
	groupID = strings.TrimSpace(groupID)
	roleID = strings.TrimSpace(roleID)
	if groupID == "" || roleID == "" {
		return domain.GroupRoleActionResponse{}, fmt.Errorf("group_id and role_id are required")
	}

	if _, err := s.roleService.GetRole(roleID); err != nil {
		return domain.GroupRoleActionResponse{}, err
	}

	s.mu.Lock()
	if s.mappings[groupID] == nil {
		s.mappings[groupID] = make(map[string]struct{})
	}
	s.mappings[groupID][roleID] = struct{}{}
	roleIDs := mapKeys(s.mappings[groupID])
	s.mu.Unlock()

	updatedUsers := s.userService.ApplyRoleMapping(groupID, s.roleNamesForIDs(roleIDs))

	return domain.GroupRoleActionResponse{
		GroupID:      groupID,
		RoleIDs:      roleIDs,
		UpdatedUsers: updatedUsers,
	}, nil
}

func (s *GroupMappingService) UnassignRole(groupID, roleID string) (domain.GroupRoleActionResponse, error) {
	groupID = strings.TrimSpace(groupID)
	roleID = strings.TrimSpace(roleID)
	if groupID == "" || roleID == "" {
		return domain.GroupRoleActionResponse{}, fmt.Errorf("group_id and role_id are required")
	}

	s.mu.Lock()
	if s.mappings[groupID] == nil {
		s.mu.Unlock()
		return domain.GroupRoleActionResponse{}, fmt.Errorf("group mapping not found")
	}
	delete(s.mappings[groupID], roleID)
	if len(s.mappings[groupID]) == 0 {
		delete(s.mappings, groupID)
	}
	roleIDs := mapKeys(s.mappings[groupID])
	s.mu.Unlock()

	updatedUsers := s.userService.ApplyRoleMapping(groupID, s.roleNamesForIDs(roleIDs))
	return domain.GroupRoleActionResponse{
		GroupID:      groupID,
		RoleIDs:      roleIDs,
		UpdatedUsers: updatedUsers,
	}, nil
}

func (s *GroupMappingService) GetGroupRoles(groupID string) (domain.GroupRoleMapping, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return domain.GroupRoleMapping{}, fmt.Errorf("group_id is required")
	}

	s.mu.RLock()
	roleIDs := mapKeys(s.mappings[groupID])
	s.mu.RUnlock()

	if len(roleIDs) == 0 {
		return domain.GroupRoleMapping{}, fmt.Errorf("group mapping not found")
	}

	return domain.GroupRoleMapping{GroupID: groupID, RoleIDs: roleIDs}, nil
}

func (s *GroupMappingService) ReconcileGroupMembers(groupID string, memberUserIDs []string) (domain.GroupRoleActionResponse, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return domain.GroupRoleActionResponse{}, fmt.Errorf("group_id is required")
	}

	if _, err := s.userService.ReconcileGroupMembership(groupID, memberUserIDs); err != nil {
		return domain.GroupRoleActionResponse{}, err
	}

	s.mu.RLock()
	roleIDs := mapKeys(s.mappings[groupID])
	s.mu.RUnlock()

	updatedUsers := s.userService.ApplyRoleMapping(groupID, s.roleNamesForIDs(roleIDs))
	return domain.GroupRoleActionResponse{
		GroupID:      groupID,
		RoleIDs:      roleIDs,
		UpdatedUsers: updatedUsers,
	}, nil
}

func (s *GroupMappingService) roleNamesForIDs(roleIDs []string) []string {
	roleNames := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		role, err := s.roleService.GetRole(roleID)
		if err == nil {
			roleNames = append(roleNames, role.Name)
		}
	}
	return roleNames
}

func mapKeys(input map[string]struct{}) []string {
	if len(input) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(input))
	for key := range input {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
