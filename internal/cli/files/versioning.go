package files

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// VersionManager manages file versioning
type VersionManager struct {
	client    *client.Client
	configDir string
	versions  map[string][]*FileVersion
	mu        sync.RWMutex
}

// FileVersion represents a version of a file
type FileVersion struct {
	ID          string                 `json:"id"`
	FileID      string                 `json:"file_id"`
	Version     int                    `json:"version"`
	Size        int64                  `json:"size"`
	Hash        string                 `json:"hash"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsCurrent   bool                   `json:"is_current"`
}

// VersionInfo represents version information for a file
type VersionInfo struct {
	FileID         string         `json:"file_id"`
	CurrentVersion int            `json:"current_version"`
	TotalVersions  int            `json:"total_versions"`
	Versions       []*FileVersion `json:"versions"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// NewVersionManager creates a new version manager
func NewVersionManager(client *client.Client, configDir string) *VersionManager {
	vm := &VersionManager{
		client:    client,
		configDir: configDir,
		versions:  make(map[string][]*FileVersion),
	}

	_ = vm.loadVersions() // Ignore error for initialization
	return vm
}

// CreateVersion creates a new version of a file
func (vm *VersionManager) CreateVersion(fileID, description, createdBy string, tags []string) (*FileVersion, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Get current file info
	fileInfo, err := vm.client.GetFile(context.Background(), fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Get existing versions
	versions := vm.versions[fileID]
	nextVersion := 1
	if len(versions) > 0 {
		nextVersion = versions[len(versions)-1].Version + 1
	}

	// Create new version
	version := &FileVersion{
		ID:          vm.generateVersionID(),
		FileID:      fileID,
		Version:     nextVersion,
		Size:        fileInfo.Size,
		Hash:        fileInfo.Hash,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Description: description,
		Tags:        tags,
		Metadata:    make(map[string]interface{}),
		IsCurrent:   true,
	}

	// Mark previous versions as not current
	for _, v := range versions {
		v.IsCurrent = false
	}

	// Add new version
	vm.versions[fileID] = append(versions, version)
	_ = vm.saveVersions() // Ignore error for demo purposes

	return version, nil
}

// GetVersions returns all versions for a file
func (vm *VersionManager) GetVersions(fileID string) ([]*FileVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[fileID]
	if !exists {
		return []*FileVersion{}, nil
	}

	// Return a copy
	result := make([]*FileVersion, len(versions))
	copy(result, versions)
	return result, nil
}

// GetCurrentVersion returns the current version of a file
func (vm *VersionManager) GetCurrentVersion(fileID string) (*FileVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[fileID]
	if !exists || len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for file: %s", fileID)
	}

	// Find current version
	for _, version := range versions {
		if version.IsCurrent {
			return version, nil
		}
	}

	// If no current version marked, return the latest
	return versions[len(versions)-1], nil
}

// RestoreVersion restores a specific version of a file
func (vm *VersionManager) RestoreVersion(fileID string, versionNumber int) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[fileID]
	if !exists {
		return fmt.Errorf("no versions found for file: %s", fileID)
	}

	// Find the version to restore
	var targetVersion *FileVersion
	for _, version := range versions {
		if version.Version == versionNumber {
			targetVersion = version
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version %d not found for file: %s", versionNumber, fileID)
	}

	// Mark all versions as not current
	for _, version := range versions {
		version.IsCurrent = false
	}

	// Mark target version as current
	targetVersion.IsCurrent = true
	_ = vm.saveVersions() // Ignore error for demo purposes

	return nil
}

// DeleteVersion deletes a specific version
func (vm *VersionManager) DeleteVersion(fileID string, versionNumber int) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	versions, exists := vm.versions[fileID]
	if !exists {
		return fmt.Errorf("no versions found for file: %s", fileID)
	}

	// Find and remove the version
	var newVersions []*FileVersion
	for _, version := range versions {
		if version.Version != versionNumber {
			newVersions = append(newVersions, version)
		}
	}

	if len(newVersions) == len(versions) {
		return fmt.Errorf("version %d not found for file: %s", versionNumber, fileID)
	}

	vm.versions[fileID] = newVersions
	_ = vm.saveVersions() // Ignore error for demo purposes

	return nil
}

// CompareVersions compares two versions of a file
func (vm *VersionManager) CompareVersions(fileID string, version1, version2 int) (*VersionComparison, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[fileID]
	if !exists {
		return nil, fmt.Errorf("no versions found for file: %s", fileID)
	}

	var v1, v2 *FileVersion
	for _, version := range versions {
		if version.Version == version1 {
			v1 = version
		}
		if version.Version == version2 {
			v2 = version
		}
	}

	if v1 == nil {
		return nil, fmt.Errorf("version %d not found", version1)
	}
	if v2 == nil {
		return nil, fmt.Errorf("version %d not found", version2)
	}

	comparison := &VersionComparison{
		FileID:   fileID,
		Version1: v1,
		Version2: v2,
		SizeDiff: v2.Size - v1.Size,
		HashDiff: v1.Hash != v2.Hash,
		TimeDiff: v2.CreatedAt.Sub(v1.CreatedAt),
		TagDiff:  vm.compareTags(v1.Tags, v2.Tags),
	}

	return comparison, nil
}

// ListAllVersions returns version info for all files
func (vm *VersionManager) ListAllVersions() []*VersionInfo {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	var result []*VersionInfo
	for fileID, versions := range vm.versions {
		if len(versions) == 0 {
			continue
		}

		currentVersion := 0
		for _, version := range versions {
			if version.IsCurrent {
				currentVersion = version.Version
				break
			}
		}

		info := &VersionInfo{
			FileID:         fileID,
			CurrentVersion: currentVersion,
			TotalVersions:  len(versions),
			Versions:       versions,
			CreatedAt:      versions[0].CreatedAt,
			UpdatedAt:      versions[len(versions)-1].CreatedAt,
		}

		result = append(result, info)
	}

	return result
}

// Utility functions
func (vm *VersionManager) generateVersionID() string {
	return fmt.Sprintf("ver_%d", time.Now().UnixNano())
}

func (vm *VersionManager) compareTags(tags1, tags2 []string) []string {
	var diff []string

	// Find tags in v1 but not in v2
	for _, tag1 := range tags1 {
		found := false
		for _, tag2 := range tags2 {
			if tag1 == tag2 {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, "-"+tag1)
		}
	}

	// Find tags in v2 but not in v1
	for _, tag2 := range tags2 {
		found := false
		for _, tag1 := range tags1 {
			if tag1 == tag2 {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, "+"+tag2)
		}
	}

	return diff
}

func (vm *VersionManager) loadVersions() error {
	versionsFile := filepath.Join(vm.configDir, "versions.json")
	if _, err := os.Stat(versionsFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty versions
	}

	data, err := os.ReadFile(versionsFile)
	if err != nil {
		return fmt.Errorf("failed to read versions file: %w", err)
	}

	var versions map[string][]*FileVersion
	if err := json.Unmarshal(data, &versions); err != nil {
		return fmt.Errorf("failed to unmarshal versions: %w", err)
	}

	vm.versions = versions
	return nil
}

func (vm *VersionManager) saveVersions() error {
	versionsFile := filepath.Join(vm.configDir, "versions.json")

	data, err := json.MarshalIndent(vm.versions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal versions: %w", err)
	}

	return os.WriteFile(versionsFile, data, 0644)
}

// VersionComparison represents a comparison between two versions
type VersionComparison struct {
	FileID   string        `json:"file_id"`
	Version1 *FileVersion  `json:"version1"`
	Version2 *FileVersion  `json:"version2"`
	SizeDiff int64         `json:"size_diff"`
	HashDiff bool          `json:"hash_diff"`
	TimeDiff time.Duration `json:"time_diff"`
	TagDiff  []string      `json:"tag_diff"`
}
