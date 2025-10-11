package cli

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Command represents a CLI command
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(ctx context.Context, args []string) error
}

// CLI represents the main CLI application
type CLI struct {
	commands map[string]Command
	mu       sync.RWMutex
}

// New creates a new CLI instance
func New() *CLI {
	return &CLI{
		commands: make(map[string]Command),
	}
}

// RegisterCommand registers a new command
func (c *CLI) RegisterCommand(name string, cmd Command) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.commands[name] = cmd
}

// Execute executes a command with the given arguments
func (c *CLI) Execute(ctx context.Context, command string, args []string) error {
	c.mu.RLock()
	cmd, exists := c.commands[command]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown command: %s. Type 'help' for available commands", command)
	}

	return cmd.Execute(ctx, args)
}

// GetCommand returns a command by name
func (c *CLI) GetCommand(name string) (Command, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cmd, exists := c.commands[name]
	return cmd, exists
}

// ListCommands returns all registered commands
func (c *CLI) ListCommands() []Command {
	c.mu.RLock()
	defer c.mu.RUnlock()

	commands := make([]Command, 0, len(c.commands))
	for _, cmd := range c.commands {
		commands = append(commands, cmd)
	}

	return commands
}

// GetCommandNames returns all command names for completion
func (c *CLI) GetCommandNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.commands))
	for name := range c.commands {
		names = append(names, name)
	}

	return names
}

// GetCompletions returns possible completions for a given input
func (c *CLI) GetCompletions(input string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var completions []string
	input = strings.ToLower(input)

	for name := range c.commands {
		if strings.HasPrefix(strings.ToLower(name), input) {
			completions = append(completions, name)
		}
	}

	return completions
}
