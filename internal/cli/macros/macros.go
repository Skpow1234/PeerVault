package macros

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Manager manages command macros
type Manager struct {
	macros    map[string]*Macro
	configDir string
	mu        sync.RWMutex
}

// Macro represents a command macro
type Macro struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Commands    []string  `json:"commands"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UsageCount  int       `json:"usage_count"`
	LastUsed    time.Time `json:"last_used"`
}

// New creates a new macro manager
func New(configDir string) *Manager {
	mm := &Manager{
		macros:    make(map[string]*Macro),
		configDir: configDir,
	}

	mm.initializeDefaultMacros()
	_ = mm.loadMacros()
	return mm
}

// initializeDefaultMacros creates default macros
func (mm *Manager) initializeDefaultMacros() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Daily check macro
	mm.macros["daily-check"] = &Macro{
		Name:        "daily-check",
		Description: "Perform daily system health checks",
		Commands: []string{
			"health",
			"metrics",
			"peers",
			"status",
		},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UsageCount: 0,
	}

	// File backup macro
	mm.macros["backup-files"] = &Macro{
		Name:        "backup-files",
		Description: "Backup important files",
		Commands: []string{
			"list --format=json",
			"batch download ./backup/ $(list --format=json | jq -r '.files[].id')",
		},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UsageCount: 0,
	}

	// System maintenance macro
	mm.macros["maintenance"] = &Macro{
		Name:        "maintenance",
		Description: "Perform system maintenance tasks",
		Commands: []string{
			"monitor start 60s",
			"health",
			"metrics",
			"peers",
			"monitor dashboard",
		},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UsageCount: 0,
	}

	// Development setup macro
	mm.macros["dev-setup"] = &Macro{
		Name:        "dev-setup",
		Description: "Setup development environment",
		Commands: []string{
			"protocol set rest",
			"connect localhost:8080",
			"config set verbose true",
			"config set output_format table",
		},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UsageCount: 0,
	}
}

// CreateMacro creates a new macro
func (mm *Manager) CreateMacro(name, description string, commands []string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.macros[name]; exists {
		return fmt.Errorf("macro '%s' already exists", name)
	}

	macro := &Macro{
		Name:        name,
		Description: description,
		Commands:    commands,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UsageCount:  0,
	}

	mm.macros[name] = macro
	return mm.saveMacro(macro)
}

// DeleteMacro deletes a macro
func (mm *Manager) DeleteMacro(name string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.macros[name]; !exists {
		return fmt.Errorf("macro '%s' does not exist", name)
	}

	delete(mm.macros, name)
	return mm.deleteMacroFile(name)
}

// GetMacro returns a macro by name
func (mm *Manager) GetMacro(name string) *Macro {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return mm.macros[name]
}

// ListMacros returns all macros
func (mm *Manager) ListMacros() map[string]*Macro {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	// Return a copy
	result := make(map[string]*Macro)
	for name, macro := range mm.macros {
		result[name] = macro
	}
	return result
}

// ExecuteMacro executes a macro and returns the commands
func (mm *Manager) ExecuteMacro(name string) ([]string, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	macro, exists := mm.macros[name]
	if !exists {
		return nil, fmt.Errorf("macro '%s' does not exist", name)
	}

	// Update usage statistics
	macro.UsageCount++
	macro.LastUsed = time.Now()
	_ = mm.saveMacro(macro)

	// Return a copy of the commands
	commands := make([]string, len(macro.Commands))
	copy(commands, macro.Commands)

	return commands, nil
}

// UpdateMacro updates an existing macro
func (mm *Manager) UpdateMacro(name, description string, commands []string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	macro, exists := mm.macros[name]
	if !exists {
		return fmt.Errorf("macro '%s' does not exist", name)
	}

	macro.Description = description
	macro.Commands = commands
	macro.UpdatedAt = time.Now()

	return mm.saveMacro(macro)
}

// CloneMacro clones an existing macro
func (mm *Manager) CloneMacro(sourceName, newName, description string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	sourceMacro, exists := mm.macros[sourceName]
	if !exists {
		return fmt.Errorf("source macro '%s' does not exist", sourceName)
	}

	if _, exists := mm.macros[newName]; exists {
		return fmt.Errorf("macro '%s' already exists", newName)
	}

	// Deep copy commands
	newCommands := make([]string, len(sourceMacro.Commands))
	copy(newCommands, sourceMacro.Commands)

	newMacro := &Macro{
		Name:        newName,
		Description: description,
		Commands:    newCommands,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UsageCount:  0,
	}

	mm.macros[newName] = newMacro
	return mm.saveMacro(newMacro)
}

// GetMacroStats returns usage statistics for macros
func (mm *Manager) GetMacroStats() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_macros":  len(mm.macros),
		"total_usage":   0,
		"most_used":     "",
		"recently_used": []string{},
	}

	var mostUsedCount int
	var recentlyUsed []string

	for name, macro := range mm.macros {
		stats["total_usage"] = stats["total_usage"].(int) + macro.UsageCount

		if macro.UsageCount > mostUsedCount {
			mostUsedCount = macro.UsageCount
			stats["most_used"] = name
		}

		// Check if used in last 24 hours
		if time.Since(macro.LastUsed) < 24*time.Hour {
			recentlyUsed = append(recentlyUsed, name)
		}
	}

	stats["recently_used"] = recentlyUsed
	return stats
}

// SearchMacros searches for macros by name or description
func (mm *Manager) SearchMacros(query string) []string {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var matches []string
	query = strings.ToLower(query)

	for name, macro := range mm.macros {
		if strings.Contains(strings.ToLower(name), query) ||
			strings.Contains(strings.ToLower(macro.Description), query) {
			matches = append(matches, name)
		}
	}

	return matches
}

// ExportMacro exports a macro to a file
func (mm *Manager) ExportMacro(name, filePath string) error {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	macro, exists := mm.macros[name]
	if !exists {
		return fmt.Errorf("macro '%s' does not exist", name)
	}

	data, err := json.MarshalIndent(macro, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal macro: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// ImportMacro imports a macro from a file
func (mm *Manager) ImportMacro(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read macro file: %w", err)
	}

	var macro Macro
	if err := json.Unmarshal(data, &macro); err != nil {
		return fmt.Errorf("failed to unmarshal macro: %w", err)
	}

	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.macros[macro.Name] = &macro
	return mm.saveMacro(&macro)
}

// loadMacros loads macros from disk
func (mm *Manager) loadMacros() error {
	macrosDir := filepath.Join(mm.configDir, "macros")
	if err := os.MkdirAll(macrosDir, 0755); err != nil {
		return fmt.Errorf("failed to create macros directory: %w", err)
	}

	files, err := os.ReadDir(macrosDir)
	if err != nil {
		return fmt.Errorf("failed to read macros directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		macroPath := filepath.Join(macrosDir, file.Name())
		data, err := os.ReadFile(macroPath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var macro Macro
		if err := json.Unmarshal(data, &macro); err != nil {
			continue // Skip files that can't be parsed
		}

		mm.macros[macro.Name] = &macro
	}

	return nil
}

// saveMacro saves a macro to disk
func (mm *Manager) saveMacro(macro *Macro) error {
	macrosDir := filepath.Join(mm.configDir, "macros")
	if err := os.MkdirAll(macrosDir, 0755); err != nil {
		return fmt.Errorf("failed to create macros directory: %w", err)
	}

	macroFile := filepath.Join(macrosDir, macro.Name+".json")
	data, err := json.MarshalIndent(macro, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal macro: %w", err)
	}

	return os.WriteFile(macroFile, data, 0644)
}

// deleteMacroFile deletes a macro file from disk
func (mm *Manager) deleteMacroFile(name string) error {
	macrosDir := filepath.Join(mm.configDir, "macros")
	macroFile := filepath.Join(macrosDir, name+".json")
	return os.Remove(macroFile)
}

// FormatMacroList formats the macro list for display
func (mm *Manager) FormatMacroList() string {
	macros := mm.ListMacros()

	if len(macros) == 0 {
		return "No macros defined"
	}

	var result strings.Builder
	result.WriteString("📝 Command Macros:\n")
	result.WriteString(strings.Repeat("=", 50) + "\n")

	for name, macro := range macros {
		result.WriteString(fmt.Sprintf("📋 %s\n", name))
		result.WriteString(fmt.Sprintf("   %s\n", macro.Description))
		result.WriteString(fmt.Sprintf("   Commands: %d | Used: %d times\n", len(macro.Commands), macro.UsageCount))
		if !macro.LastUsed.IsZero() {
			result.WriteString(fmt.Sprintf("   Last used: %s\n", macro.LastUsed.Format("2006-01-02 15:04:05")))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// GetMacroHelp returns help text for macros
func (mm *Manager) GetMacroHelp() string {
	return `📝 Command Macros

Macros allow you to save and replay sequences of commands. They're perfect for automating repetitive tasks.

Examples:
  macro run daily-check          # Run the daily-check macro
  macro run backup-files         # Run the backup-files macro
  macro run maintenance          # Run the maintenance macro

Managing Macros:
  macro create <name> <description> <command1> [command2] ...  # Create a new macro
  macro delete <name>                                          # Delete a macro
  macro list                                                    # List all macros
  macro show <name>                                            # Show macro details
  macro clone <source> <new_name> <description>               # Clone a macro
  macro search <query>                                         # Search macros
  macro stats                                                  # Show usage statistics

Built-in Macros:
  daily-check    - Perform daily system health checks
  backup-files   - Backup important files
  maintenance    - Perform system maintenance tasks
  dev-setup      - Setup development environment

Note: Macros are saved automatically and persist between sessions.`
}
