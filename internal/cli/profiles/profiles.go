package profiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Skpow1234/Peervault/internal/cli/config"
)

// Manager manages configuration profiles
type Manager struct {
	profiles  map[string]*Profile
	current   string
	configDir string
	mu        sync.RWMutex
}

// Profile represents a configuration profile
type Profile struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      *config.Config         `json:"config"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// New creates a new profile manager
func New(configDir string) *Manager {
	pm := &Manager{
		profiles:  make(map[string]*Profile),
		configDir: configDir,
	}

	pm.initializeDefaultProfiles()
	_ = pm.loadProfiles()
	return pm
}

// initializeDefaultProfiles creates default profiles
func (pm *Manager) initializeDefaultProfiles() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Development profile
	pm.profiles["development"] = &Profile{
		Name:        "development",
		Description: "Development environment with local server",
		Config: &config.Config{
			ServerURL:    "http://localhost:8080",
			AuthToken:    "",
			OutputFormat: "table",
			Theme:        "default",
			Verbose:      true,
			HistoryFile:  "history_dev.txt",
		},
		Metadata: map[string]interface{}{
			"environment": "development",
			"debug":       true,
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	// Production profile
	pm.profiles["production"] = &Profile{
		Name:        "production",
		Description: "Production environment with secure settings",
		Config: &config.Config{
			ServerURL:    "https://api.peervault.com",
			AuthToken:    "",
			OutputFormat: "json",
			Theme:        "minimal",
			Verbose:      false,
			HistoryFile:  "history_prod.txt",
		},
		Metadata: map[string]interface{}{
			"environment": "production",
			"debug":       false,
			"security":    "high",
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	// Testing profile
	pm.profiles["testing"] = &Profile{
		Name:        "testing",
		Description: "Testing environment with mock server",
		Config: &config.Config{
			ServerURL:    "http://localhost:3000",
			AuthToken:    "test-token",
			OutputFormat: "yaml",
			Theme:        "test",
			Verbose:      true,
			HistoryFile:  "history_test.txt",
		},
		Metadata: map[string]interface{}{
			"environment": "testing",
			"debug":       true,
			"mock":        true,
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	// Set default current profile
	pm.current = "development"
}

// CreateProfile creates a new profile
func (pm *Manager) CreateProfile(name, description string, baseConfig *config.Config) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.profiles[name]; exists {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	profile := &Profile{
		Name:        name,
		Description: description,
		Config:      baseConfig,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
	}

	pm.profiles[name] = profile
	return pm.saveProfile(profile)
}

// DeleteProfile deletes a profile
func (pm *Manager) DeleteProfile(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.profiles[name]; !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Don't allow deleting the current profile
	if pm.current == name {
		return fmt.Errorf("cannot delete current profile '%s'", name)
	}

	delete(pm.profiles, name)
	return pm.deleteProfileFile(name)
}

// SwitchProfile switches to a different profile
func (pm *Manager) SwitchProfile(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.profiles[name]; !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	pm.current = name
	return pm.saveCurrentProfile()
}

// GetCurrentProfile returns the current profile
func (pm *Manager) GetCurrentProfile() *Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.profiles[pm.current]
}

// GetCurrentProfileName returns the current profile name
func (pm *Manager) GetCurrentProfileName() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.current
}

// GetProfile returns a specific profile
func (pm *Manager) GetProfile(name string) *Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.profiles[name]
}

// ListProfiles returns all profiles
func (pm *Manager) ListProfiles() map[string]*Profile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy
	result := make(map[string]*Profile)
	for name, profile := range pm.profiles {
		result[name] = profile
	}
	return result
}

// UpdateProfile updates a profile
func (pm *Manager) UpdateProfile(name string, updates map[string]interface{}) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	profile, exists := pm.profiles[name]
	if !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Update config fields
	if configUpdates, ok := updates["config"].(map[string]interface{}); ok {
		if serverURL, ok := configUpdates["server_url"].(string); ok {
			profile.Config.ServerURL = serverURL
		}
		if authToken, ok := configUpdates["auth_token"].(string); ok {
			profile.Config.AuthToken = authToken
		}
		if outputFormat, ok := configUpdates["output_format"].(string); ok {
			profile.Config.OutputFormat = outputFormat
		}
		if theme, ok := configUpdates["theme"].(string); ok {
			profile.Config.Theme = theme
		}
		if verbose, ok := configUpdates["verbose"].(bool); ok {
			profile.Config.Verbose = verbose
		}
	}

	// Update metadata
	if metadata, ok := updates["metadata"].(map[string]interface{}); ok {
		for key, value := range metadata {
			profile.Metadata[key] = value
		}
	}

	// Update description
	if description, ok := updates["description"].(string); ok {
		profile.Description = description
	}

	profile.UpdatedAt = getCurrentTimestamp()
	return pm.saveProfile(profile)
}

// CloneProfile clones an existing profile
func (pm *Manager) CloneProfile(sourceName, newName, description string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	sourceProfile, exists := pm.profiles[sourceName]
	if !exists {
		return fmt.Errorf("source profile '%s' does not exist", sourceName)
	}

	if _, exists := pm.profiles[newName]; exists {
		return fmt.Errorf("profile '%s' already exists", newName)
	}

	// Deep copy the config
	newConfig := &config.Config{
		ServerURL:    sourceProfile.Config.ServerURL,
		AuthToken:    sourceProfile.Config.AuthToken,
		OutputFormat: sourceProfile.Config.OutputFormat,
		Theme:        sourceProfile.Config.Theme,
		Verbose:      sourceProfile.Config.Verbose,
		HistoryFile:  sourceProfile.Config.HistoryFile,
	}

	// Deep copy metadata
	newMetadata := make(map[string]interface{})
	for key, value := range sourceProfile.Metadata {
		newMetadata[key] = value
	}

	newProfile := &Profile{
		Name:        newName,
		Description: description,
		Config:      newConfig,
		Metadata:    newMetadata,
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
	}

	pm.profiles[newName] = newProfile
	return pm.saveProfile(newProfile)
}

// ExportProfile exports a profile to a file
func (pm *Manager) ExportProfile(name, filePath string) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	profile, exists := pm.profiles[name]
	if !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// ImportProfile imports a profile from a file
func (pm *Manager) ImportProfile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.profiles[profile.Name] = &profile
	return pm.saveProfile(&profile)
}

// loadProfiles loads profiles from disk
func (pm *Manager) loadProfiles() error {
	profilesDir := filepath.Join(pm.configDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Load current profile setting
	currentFile := filepath.Join(profilesDir, "current.json")
	if data, err := os.ReadFile(currentFile); err == nil {
		var current struct {
			Current string `json:"current"`
		}
		if err := json.Unmarshal(data, &current); err == nil {
			pm.current = current.Current
		}
	}

	// Load individual profiles
	files, err := os.ReadDir(profilesDir)
	if err != nil {
		return fmt.Errorf("failed to read profiles directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == "current.json" {
			continue
		}

		profilePath := filepath.Join(profilesDir, file.Name())
		data, err := os.ReadFile(profilePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var profile Profile
		if err := json.Unmarshal(data, &profile); err != nil {
			continue // Skip files that can't be parsed
		}

		pm.profiles[profile.Name] = &profile
	}

	return nil
}

// saveProfile saves a profile to disk
func (pm *Manager) saveProfile(profile *Profile) error {
	profilesDir := filepath.Join(pm.configDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	profileFile := filepath.Join(profilesDir, profile.Name+".json")
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	return os.WriteFile(profileFile, data, 0644)
}

// saveCurrentProfile saves the current profile setting
func (pm *Manager) saveCurrentProfile() error {
	profilesDir := filepath.Join(pm.configDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	currentFile := filepath.Join(profilesDir, "current.json")
	data, err := json.MarshalIndent(map[string]string{"current": pm.current}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal current profile: %w", err)
	}

	return os.WriteFile(currentFile, data, 0644)
}

// deleteProfileFile deletes a profile file from disk
func (pm *Manager) deleteProfileFile(name string) error {
	profilesDir := filepath.Join(pm.configDir, "profiles")
	profileFile := filepath.Join(profilesDir, name+".json")
	return os.Remove(profileFile)
}

// getCurrentTimestamp returns the current timestamp
func getCurrentTimestamp() string {
	return "2024-01-01T00:00:00Z" // In a real implementation, use time.Now().Format(time.RFC3339)
}
