package files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// ShareManager manages file sharing
type ShareManager struct {
	client    *client.Client
	configDir string
	shares    map[string]*FileShare
	mu        sync.RWMutex
}

// FileShare represents a shared file
type FileShare struct {
	ID           string                 `json:"id"`
	FileID       string                 `json:"file_id"`
	SharedBy     string                 `json:"shared_by"`
	SharedWith   []string               `json:"shared_with"`
	Permissions  []string               `json:"permissions"` // "read", "write", "delete"
	ExpiresAt    time.Time              `json:"expires_at"`
	CreatedAt    time.Time              `json:"created_at"`
	AccessCount  int                    `json:"access_count"`
	LastAccessed time.Time              `json:"last_accessed"`
	IsPublic     bool                   `json:"is_public"`
	PublicURL    string                 `json:"public_url,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SharePermission represents sharing permissions
type SharePermission struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Permissions []string  `json:"permissions"`
	GrantedAt   time.Time `json:"granted_at"`
	GrantedBy   string    `json:"granted_by"`
}

// ShareStats represents sharing statistics
type ShareStats struct {
	TotalShares   int `json:"total_shares"`
	PublicShares  int `json:"public_shares"`
	PrivateShares int `json:"private_shares"`
	ExpiredShares int `json:"expired_shares"`
	ActiveShares  int `json:"active_shares"`
	TotalAccesses int `json:"total_accesses"`
}

// NewShareManager creates a new share manager
func NewShareManager(client *client.Client, configDir string) *ShareManager {
	sm := &ShareManager{
		client:    client,
		configDir: configDir,
		shares:    make(map[string]*FileShare),
	}

	sm.loadShares()
	return sm
}

// ShareFile shares a file with specific users
func (sm *ShareManager) ShareFile(fileID, sharedBy string, sharedWith []string, permissions []string, expiresIn time.Duration) (*FileShare, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate permissions
	validPermissions := []string{"read", "write", "delete"}
	for _, perm := range permissions {
		valid := false
		for _, validPerm := range validPermissions {
			if perm == validPerm {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("invalid permission: %s", perm)
		}
	}

	// Create share
	share := &FileShare{
		ID:          sm.generateShareID(),
		FileID:      fileID,
		SharedBy:    sharedBy,
		SharedWith:  sharedWith,
		Permissions: permissions,
		ExpiresAt:   time.Now().Add(expiresIn),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		IsPublic:    false,
		Metadata:    make(map[string]interface{}),
	}

	sm.shares[share.ID] = share
	sm.saveShares()

	return share, nil
}

// ShareFilePublicly shares a file publicly
func (sm *ShareManager) ShareFilePublicly(fileID, sharedBy string, permissions []string, expiresIn time.Duration) (*FileShare, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Create public share
	share := &FileShare{
		ID:          sm.generateShareID(),
		FileID:      fileID,
		SharedBy:    sharedBy,
		SharedWith:  []string{}, // Empty for public shares
		Permissions: permissions,
		ExpiresAt:   time.Now().Add(expiresIn),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		IsPublic:    true,
		PublicURL:   sm.generatePublicURL(),
		Metadata:    make(map[string]interface{}),
	}

	sm.shares[share.ID] = share
	sm.saveShares()

	return share, nil
}

// GetShare returns a share by ID
func (sm *ShareManager) GetShare(shareID string) (*FileShare, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	share, exists := sm.shares[shareID]
	if !exists {
		return nil, fmt.Errorf("share not found: %s", shareID)
	}

	// Check if expired
	if time.Now().After(share.ExpiresAt) {
		return nil, fmt.Errorf("share has expired")
	}

	return share, nil
}

// GetSharesForFile returns all shares for a file
func (sm *ShareManager) GetSharesForFile(fileID string) ([]*FileShare, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var shares []*FileShare
	for _, share := range sm.shares {
		if share.FileID == fileID {
			shares = append(shares, share)
		}
	}

	return shares, nil
}

// GetSharesByUser returns all shares created by a user
func (sm *ShareManager) GetSharesByUser(userID string) ([]*FileShare, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var shares []*FileShare
	for _, share := range sm.shares {
		if share.SharedBy == userID {
			shares = append(shares, share)
		}
	}

	return shares, nil
}

// GetSharedWithUser returns all shares shared with a user
func (sm *ShareManager) GetSharedWithUser(userID string) ([]*FileShare, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var shares []*FileShare
	for _, share := range sm.shares {
		// Check if user is in shared_with list
		for _, sharedUser := range share.SharedWith {
			if sharedUser == userID {
				shares = append(shares, share)
				break
			}
		}
	}

	return shares, nil
}

// AccessShare records access to a shared file
func (sm *ShareManager) AccessShare(shareID, userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	share, exists := sm.shares[shareID]
	if !exists {
		return fmt.Errorf("share not found: %s", shareID)
	}

	// Check if expired
	if time.Now().After(share.ExpiresAt) {
		return fmt.Errorf("share has expired")
	}

	// Check permissions for private shares
	if !share.IsPublic {
		hasAccess := false
		for _, sharedUser := range share.SharedWith {
			if sharedUser == userID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			return fmt.Errorf("access denied for user: %s", userID)
		}
	}

	// Update access statistics
	share.AccessCount++
	share.LastAccessed = time.Now()
	sm.saveShares()

	return nil
}

// UpdateSharePermissions updates permissions for a share
func (sm *ShareManager) UpdateSharePermissions(shareID string, permissions []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	share, exists := sm.shares[shareID]
	if !exists {
		return fmt.Errorf("share not found: %s", shareID)
	}

	// Validate permissions
	validPermissions := []string{"read", "write", "delete"}
	for _, perm := range permissions {
		valid := false
		for _, validPerm := range validPermissions {
			if perm == validPerm {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	share.Permissions = permissions
	sm.saveShares()

	return nil
}

// AddUserToShare adds a user to an existing share
func (sm *ShareManager) AddUserToShare(shareID, userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	share, exists := sm.shares[shareID]
	if !exists {
		return fmt.Errorf("share not found: %s", shareID)
	}

	// Check if user is already in the list
	for _, existingUser := range share.SharedWith {
		if existingUser == userID {
			return fmt.Errorf("user already has access to this share")
		}
	}

	share.SharedWith = append(share.SharedWith, userID)
	sm.saveShares()

	return nil
}

// RemoveUserFromShare removes a user from a share
func (sm *ShareManager) RemoveUserFromShare(shareID, userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	share, exists := sm.shares[shareID]
	if !exists {
		return fmt.Errorf("share not found: %s", shareID)
	}

	// Remove user from list
	var newSharedWith []string
	for _, existingUser := range share.SharedWith {
		if existingUser != userID {
			newSharedWith = append(newSharedWith, existingUser)
		}
	}

	share.SharedWith = newSharedWith
	sm.saveShares()

	return nil
}

// RevokeShare revokes a share
func (sm *ShareManager) RevokeShare(shareID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	share, exists := sm.shares[shareID]
	if !exists {
		return fmt.Errorf("share not found: %s", shareID)
	}

	// Mark as expired
	share.ExpiresAt = time.Now().Add(-time.Hour)
	sm.saveShares()

	return nil
}

// DeleteShare permanently deletes a share
func (sm *ShareManager) DeleteShare(shareID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, exists := sm.shares[shareID]
	if !exists {
		return fmt.Errorf("share not found: %s", shareID)
	}

	delete(sm.shares, shareID)
	sm.saveShares()

	return nil
}

// GetShareStats returns sharing statistics
func (sm *ShareManager) GetShareStats() *ShareStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := &ShareStats{}
	now := time.Now()

	for _, share := range sm.shares {
		stats.TotalShares++
		stats.TotalAccesses += share.AccessCount

		if share.IsPublic {
			stats.PublicShares++
		} else {
			stats.PrivateShares++
		}

		if now.After(share.ExpiresAt) {
			stats.ExpiredShares++
		} else {
			stats.ActiveShares++
		}
	}

	return stats
}

// CleanupExpiredShares removes expired shares
func (sm *ShareManager) CleanupExpiredShares() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	var toDelete []string

	for shareID, share := range sm.shares {
		if now.After(share.ExpiresAt) {
			toDelete = append(toDelete, shareID)
		}
	}

	for _, shareID := range toDelete {
		delete(sm.shares, shareID)
	}

	if len(toDelete) > 0 {
		sm.saveShares()
	}

	return nil
}

// Utility functions
func (sm *ShareManager) generateShareID() string {
	return fmt.Sprintf("share_%d", time.Now().UnixNano())
}

func (sm *ShareManager) generatePublicURL() string {
	return fmt.Sprintf("https://peervault.com/share/%s", sm.generateShareID())
}

func (sm *ShareManager) loadShares() error {
	sharesFile := filepath.Join(sm.configDir, "shares.json")
	if _, err := os.Stat(sharesFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty shares
	}

	data, err := os.ReadFile(sharesFile)
	if err != nil {
		return fmt.Errorf("failed to read shares file: %w", err)
	}

	var shares map[string]*FileShare
	if err := json.Unmarshal(data, &shares); err != nil {
		return fmt.Errorf("failed to unmarshal shares: %w", err)
	}

	sm.shares = shares
	return nil
}

func (sm *ShareManager) saveShares() error {
	sharesFile := filepath.Join(sm.configDir, "shares.json")

	data, err := json.MarshalIndent(sm.shares, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal shares: %w", err)
	}

	return os.WriteFile(sharesFile, data, 0644)
}
