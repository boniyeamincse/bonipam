package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type syncedUser struct {
	UserID       string
	ExternalID   string
	Email        string
	DisplayName  string
	Groups       []string
	LastSyncedAt time.Time
}

// UserSyncService simulates IdP profile/group synchronization for scaffold usage.
// Replace this with a real IdP connector when integrating OIDC/SAML providers.
type UserSyncService struct {
	mu    sync.RWMutex
	users map[string]syncedUser
}

func NewUserSyncService() *UserSyncService {
	return &UserSyncService{users: make(map[string]syncedUser)}
}

func (s *UserSyncService) SyncUserOnLogin(code, state string) (string, error) {
	if code == "" || state == "" {
		return "", fmt.Errorf("missing OIDC exchange payload")
	}

	externalID := stableID(code + ":" + state)
	userID := "user-" + externalID
	now := time.Now().UTC()

	profile := syncedUser{
		UserID:       userID,
		ExternalID:   externalID,
		Email:        externalID + "@idp.example.com",
		DisplayName:  "User " + externalID[:8],
		Groups:       []string{"pam-users", "engineering"},
		LastSyncedAt: now,
	}

	s.mu.Lock()
	s.users[externalID] = profile
	s.mu.Unlock()

	return userID, nil
}

func (s *UserSyncService) SyncAllUsers() (int, error) {
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.users) == 0 {
		seed := []syncedUser{
			{
				UserID:       "user-idp-seed001",
				ExternalID:   "idp-seed001",
				Email:        "seed001@idp.example.com",
				DisplayName:  "Seed User 001",
				Groups:       []string{"pam-users"},
				LastSyncedAt: now,
			},
			{
				UserID:       "user-idp-seed002",
				ExternalID:   "idp-seed002",
				Email:        "seed002@idp.example.com",
				DisplayName:  "Seed User 002",
				Groups:       []string{"pam-users", "security"},
				LastSyncedAt: now,
			},
		}
		for _, user := range seed {
			s.users[user.ExternalID] = user
		}
		return len(seed), nil
	}

	count := 0
	for externalID, user := range s.users {
		user.LastSyncedAt = now
		s.users[externalID] = user
		count++
	}

	return count, nil
}

func stableID(value string) string {
	h := sha256.Sum256([]byte(value))
	return "idp-" + hex.EncodeToString(h[:])[:20]
}
