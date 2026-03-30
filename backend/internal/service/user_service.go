package service

import (
	"boni-pam/internal/domain"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	mu    sync.RWMutex
	users map[string]domain.User
}

func NewUserService() *UserService {
	return &UserService{users: make(map[string]domain.User)}
}

func (s *UserService) CreateUser(req domain.CreateUserRequest) (domain.User, error) {
	if req.Email == "" || req.DisplayName == "" {
		return domain.User{}, fmt.Errorf("email and display_name are required")
	}

	now := time.Now().UTC()
	user := domain.User{
		ID:          "usr-" + uuid.NewString(),
		ExternalID:  req.ExternalID,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Groups:      req.Groups,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.users {
		if existing.Email == user.Email && existing.DeletedAt == nil {
			return domain.User{}, fmt.Errorf("user with email already exists")
		}
	}

	s.users[user.ID] = user
	return user, nil
}

func (s *UserService) ListUsers(includeDeleted bool) []domain.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		if !includeDeleted && user.DeletedAt != nil {
			continue
		}
		result = append(result, user)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result
}

func (s *UserService) GetUser(userID string, includeDeleted bool) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.User{}, fmt.Errorf("user not found")
	}
	if !includeDeleted && user.DeletedAt != nil {
		return domain.User{}, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *UserService) UpdateUser(userID string, req domain.UpdateUserRequest) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok || user.DeletedAt != nil {
		return domain.User{}, fmt.Errorf("user not found")
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Groups != nil {
		user.Groups = req.Groups
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	user.UpdatedAt = time.Now().UTC()

	s.users[userID] = user
	return user, nil
}

func (s *UserService) SoftDeleteUser(userID string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.User{}, fmt.Errorf("user not found")
	}
	if user.DeletedAt != nil {
		return domain.User{}, fmt.Errorf("user already deleted")
	}

	now := time.Now().UTC()
	user.DeletedAt = &now
	user.Status = "deleted"
	user.UpdatedAt = now
	s.users[userID] = user

	return user, nil
}

func (s *UserService) RestoreUser(userID string) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.User{}, fmt.Errorf("user not found")
	}
	if user.DeletedAt == nil {
		return domain.User{}, fmt.Errorf("user is not deleted")
	}

	user.DeletedAt = nil
	user.Status = "active"
	user.UpdatedAt = time.Now().UTC()
	s.users[userID] = user

	return user, nil
}
