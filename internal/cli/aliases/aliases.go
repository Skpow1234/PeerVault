package aliases

import (
	"fmt"
	"strings"
	"sync"
)

// Manager manages command aliases
type Manager struct {
	aliases map[string]string
	mu      sync.RWMutex
}

// New creates a new alias manager
func New() *Manager {
	am := &Manager{
		aliases: make(map[string]string),
	}

	am.initializeDefaultAliases()
	return am
}

// initializeDefaultAliases sets up default aliases
func (am *Manager) initializeDefaultAliases() {
	am.mu.Lock()
	defer am.mu.Unlock()

	// File operations
	am.aliases["s"] = "store"
	am.aliases["g"] = "get"
	am.aliases["l"] = "list"
	am.aliases["d"] = "delete"

	// Network operations
	am.aliases["c"] = "connect"
	am.aliases["dc"] = "disconnect"
	am.aliases["p"] = "peers"

	// System operations
	am.aliases["h"] = "health"
	am.aliases["m"] = "metrics"
	am.aliases["st"] = "status"

	// Utility operations
	am.aliases["q"] = "exit"
	am.aliases["quit"] = "exit"
	am.aliases["cls"] = "clear"
	am.aliases["hist"] = "history"

	// Advanced operations
	am.aliases["b"] = "batch"
	am.aliases["mon"] = "monitor"
	am.aliases["rt"] = "realtime"
	am.aliases["proto"] = "protocol"

	// Help
	am.aliases["?"] = "help"
	am.aliases["man"] = "help"
}

// AddAlias adds a new alias
func (am *Manager) AddAlias(alias, command string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if alias conflicts with existing command
	if am.isBuiltinCommand(alias) {
		return fmt.Errorf("cannot create alias '%s': conflicts with builtin command", alias)
	}

	// Check if command exists
	if !am.isValidCommand(command) {
		return fmt.Errorf("cannot create alias '%s': command '%s' does not exist", alias, command)
	}

	am.aliases[alias] = command
	return nil
}

// RemoveAlias removes an alias
func (am *Manager) RemoveAlias(alias string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.aliases[alias]; !exists {
		return fmt.Errorf("alias '%s' does not exist", alias)
	}

	delete(am.aliases, alias)
	return nil
}

// GetAlias returns the command for an alias
func (am *Manager) GetAlias(alias string) (string, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	command, exists := am.aliases[alias]
	return command, exists
}

// ExpandAliases expands aliases in a command line
func (am *Manager) ExpandAliases(input string) string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return input
	}

	// Check if first part is an alias
	if command, exists := am.GetAlias(parts[0]); exists {
		parts[0] = command
		return strings.Join(parts, " ")
	}

	return input
}

// ListAliases returns all aliases
func (am *Manager) ListAliases() map[string]string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Return a copy
	result := make(map[string]string)
	for alias, command := range am.aliases {
		result[alias] = command
	}
	return result
}

// GetAliasesForCommand returns all aliases for a specific command
func (am *Manager) GetAliasesForCommand(command string) []string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var aliases []string
	for alias, cmd := range am.aliases {
		if cmd == command {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

// isBuiltinCommand checks if a string is a builtin command
func (am *Manager) isBuiltinCommand(command string) bool {
	builtins := []string{
		"store", "get", "list", "delete", "connect", "disconnect", "peers",
		"health", "metrics", "status", "help", "exit", "clear", "history",
		"batch", "monitor", "realtime", "protocol", "config", "set",
	}

	for _, builtin := range builtins {
		if command == builtin {
			return true
		}
	}
	return false
}

// isValidCommand checks if a command is valid
func (am *Manager) isValidCommand(command string) bool {
	// For now, we'll consider all builtin commands as valid
	// In a real implementation, this would check against the actual command registry
	return am.isBuiltinCommand(command)
}

// FormatAliases formats aliases for display
func (am *Manager) FormatAliases() string {
	aliases := am.ListAliases()

	if len(aliases) == 0 {
		return "No aliases defined"
	}

	var result strings.Builder
	result.WriteString("📝 Command Aliases:\n")
	result.WriteString(strings.Repeat("=", 30) + "\n")

	// Group aliases by category
	categories := map[string][]string{
		"File Operations":     {"s", "g", "l", "d"},
		"Network Operations":  {"c", "dc", "p"},
		"System Operations":   {"h", "m", "st"},
		"Utility Operations":  {"q", "quit", "cls", "hist", "?", "man"},
		"Advanced Operations": {"b", "mon", "rt", "proto"},
	}

	for category, aliasList := range categories {
		result.WriteString(fmt.Sprintf("\n%s:\n", category))
		for _, alias := range aliasList {
			if command, exists := aliases[alias]; exists {
				result.WriteString(fmt.Sprintf("  %-4s → %s\n", alias, command))
			}
		}
	}

	// Show any custom aliases not in categories
	result.WriteString("\nCustom Aliases:\n")
	for alias, command := range aliases {
		found := false
		for _, aliasList := range categories {
			for _, catAlias := range aliasList {
				if alias == catAlias {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			result.WriteString(fmt.Sprintf("  %-4s → %s\n", alias, command))
		}
	}

	return result.String()
}

// GetAliasHelp returns help text for aliases
func (am *Manager) GetAliasHelp() string {
	return `📝 Command Aliases

Aliases provide shortcuts for commonly used commands. You can use them just like regular commands.

Examples:
  s document.pdf          # Same as: store document.pdf
  g abc123def456          # Same as: get abc123def456
  l --format=json         # Same as: list --format=json
  h                       # Same as: health
  ? store                 # Same as: help store

Managing Aliases:
  alias add <alias> <command>     # Add a new alias
  alias remove <alias>            # Remove an alias
  alias list                      # List all aliases
  alias help                      # Show this help

Note: Aliases cannot override builtin commands.`
}
