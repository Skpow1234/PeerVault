package completion

import (
	"os"
	"path/filepath"
	"strings"
)

// Completer provides command and argument completion
type Completer struct {
	commands []string
	// filePaths  []string // TODO: implement file path completion
	addresses  []string
	configKeys []string
}

// New creates a new completer
func New() *Completer {
	return &Completer{
		commands: []string{
			"store", "get", "list", "delete", "connect", "disconnect",
			"peers", "health", "metrics", "status", "help", "exit",
			"clear", "history", "config", "set", "blockchain", "iot",
			"backup", "restore", "ml", "edge", "cache", "compression",
			"deduplication", "encryption", "ipfs", "mqtt", "coap",
			"ls", "bc", "devices", "quit",
		},
		addresses: []string{
			"localhost:3000", "localhost:8080", "node1:3000", "node2:7000",
			"127.0.0.1:3000", "127.0.0.1:8080", "peer1:3000", "peer2:7000",
		},
		configKeys: []string{
			"server_url", "auth_token", "output_format", "theme", "verbose",
			"auto_complete", "history_size", "timeout", "retry_count",
		},
	}
}

// CompleteCommand completes command names
func (c *Completer) CompleteCommand(prefix string) []string {
	var completions []string
	for _, cmd := range c.commands {
		if strings.HasPrefix(cmd, prefix) {
			completions = append(completions, cmd)
		}
	}
	return completions
}

// CompleteArgument completes command arguments based on the command
func (c *Completer) CompleteArgument(command, prefix string) []string {
	switch command {
	case "store":
		return c.CompleteFilePath(prefix)
	case "get":
		// Check if this is a config get or file get
		if strings.HasPrefix(prefix, "config.") || strings.Contains(prefix, "_") {
			return c.CompleteConfigKey(prefix)
		}
		return c.CompleteFilePath(prefix)
	case "connect", "peers":
		return c.CompleteAddress(prefix)
	case "set":
		return c.CompleteConfigKey(prefix)
	case "delete":
		return c.CompleteFileID(prefix)
	case "backup", "restore":
		return c.CompleteBackupID(prefix)
	case "iot", "devices":
		return c.CompleteDeviceID(prefix)
	case "blockchain", "bc":
		return c.CompleteBlockchainCommand(prefix)
	default:
		return nil
	}
}

// CompleteFilePath completes file paths
func (c *Completer) CompleteFilePath(prefix string) []string {
	var completions []string

	// If prefix is empty, show current directory
	if prefix == "" {
		prefix = "."
	}

	// Get directory and file parts
	dir := filepath.Dir(prefix)
	file := filepath.Base(prefix)

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return completions
	}

	// Find matching entries
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, file) {
			fullPath := filepath.Join(dir, name)
			if entry.IsDir() {
				fullPath += "/"
			}
			completions = append(completions, fullPath)
		}
	}

	return completions
}

// CompleteAddress completes network addresses
func (c *Completer) CompleteAddress(prefix string) []string {
	var completions []string
	for _, addr := range c.addresses {
		if strings.HasPrefix(addr, prefix) {
			completions = append(completions, addr)
		}
	}
	return completions
}

// CompleteConfigKey completes configuration keys
func (c *Completer) CompleteConfigKey(prefix string) []string {
	var completions []string
	for _, key := range c.configKeys {
		if strings.HasPrefix(key, prefix) {
			completions = append(completions, key)
		}
	}
	return completions
}

// CompleteFileID completes file IDs (simplified)
func (c *Completer) CompleteFileID(prefix string) []string {
	// In a real implementation, you'd query the API for file IDs
	return []string{"file-123", "file-456", "file-789"}
}

// CompleteBackupID completes backup IDs (simplified)
func (c *Completer) CompleteBackupID(prefix string) []string {
	// In a real implementation, you'd query the API for backup IDs
	return []string{"backup-123", "backup-456", "backup-789"}
}

// CompleteDeviceID completes device IDs (simplified)
func (c *Completer) CompleteDeviceID(prefix string) []string {
	// In a real implementation, you'd query the API for device IDs
	return []string{"device-123", "device-456", "device-789"}
}

// CompleteBlockchainCommand completes blockchain subcommands
func (c *Completer) CompleteBlockchainCommand(prefix string) []string {
	commands := []string{"wallet", "contract", "tx", "transaction", "stats", "help"}
	var completions []string
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, prefix) {
			completions = append(completions, cmd)
		}
	}
	return completions
}

// CompleteOption completes command options (flags)
func (c *Completer) CompleteOption(command, prefix string) []string {
	options := map[string][]string{
		"store":   {"--encrypt", "--compress", "--backup"},
		"get":     {"--output", "--format", "--verify"},
		"list":    {"--format", "--filter", "--sort"},
		"peers":   {"--status", "--format", "--filter"},
		"health":  {"--format", "--verbose"},
		"metrics": {"--format", "--live", "--interval"},
		"backup":  {"--type", "--schedule", "--retention"},
		"restore": {"--backup-id", "--target", "--verify"},
		"config":  {"--format", "--export", "--import"},
	}

	cmdOptions, exists := options[command]
	if !exists {
		return nil
	}

	var completions []string
	for _, opt := range cmdOptions {
		if strings.HasPrefix(opt, prefix) {
			completions = append(completions, opt)
		}
	}
	return completions
}

// CompleteValue completes option values
func (c *Completer) CompleteValue(option, prefix string) []string {
	switch option {
	case "--format":
		return []string{"table", "json", "yaml"}
	case "--output":
		return c.CompleteFilePath(prefix)
	case "--type":
		return []string{"full", "incremental", "differential"}
	case "--status":
		return []string{"healthy", "degraded", "unhealthy", "all"}
	case "--filter":
		return []string{"active", "inactive", "recent", "large"}
	case "--sort":
		return []string{"name", "size", "date", "type"}
	case "--theme":
		return []string{"default", "dark", "light", "colorful"}
	case "--interval":
		return []string{"1s", "5s", "10s", "30s", "1m", "5m"}
	default:
		return nil
	}
}

// GetCompletions returns all possible completions for a given input
func (c *Completer) GetCompletions(input string) []string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return c.CompleteCommand("")
	}

	// Check if we're completing an option value
	if len(parts) >= 2 {
		lastPart := parts[len(parts)-1]
		secondLastPart := parts[len(parts)-2]

		// If second last part is an option, complete the value
		if strings.HasPrefix(secondLastPart, "--") {
			return c.CompleteValue(secondLastPart, lastPart)
		}
	}

	// Check if we're completing an option
	lastPart := parts[len(parts)-1]
	if strings.HasPrefix(lastPart, "--") {
		return c.CompleteOption(parts[0], lastPart)
	}

	// Complete arguments
	if len(parts) == 1 {
		return c.CompleteCommand(lastPart)
	}

	return c.CompleteArgument(parts[0], lastPart)
}
