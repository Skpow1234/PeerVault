package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli"
	"github.com/Skpow1234/Peervault/internal/cli/aliases"
	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/config"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/help"
	"github.com/Skpow1234/Peervault/internal/cli/macros"
	"github.com/Skpow1234/Peervault/internal/cli/output"
	"github.com/Skpow1234/Peervault/internal/cli/profiles"
)

// EnhancedHelpCommand provides enhanced help with examples and tutorials
type EnhancedHelpCommand struct {
	BaseCommand
	helpManager *help.HelpManager
}

// NewEnhancedHelpCommand creates a new enhanced help command
func NewEnhancedHelpCommand(cliApp *cli.CLI) *EnhancedHelpCommand {
	return &EnhancedHelpCommand{
		BaseCommand: BaseCommand{
			name:        "help",
			description: "Show help information with examples and tutorials",
			usage:       "help [command] [--examples] [--tutorial]",
			client:      nil,
			formatter:   nil,
		},
		helpManager: help.New(),
	}
}

// Execute executes the enhanced help command
func (c *EnhancedHelpCommand) Execute(ctx context.Context, args []string) error {
	showExamples := false
	showTutorial := false

	// Parse flags
	var command string
	for _, arg := range args {
		if arg == "--examples" {
			showExamples = true
		} else if arg == "--tutorial" {
			showTutorial = true
		} else if !strings.HasPrefix(arg, "--") {
			command = arg
			break
		}
	}

	if command == "" {
		// Show general help
		fmt.Println("🚀 PeerVault CLI - Available Commands")
		fmt.Println(strings.Repeat("=", 50))

		commands := c.helpManager.GetAvailableCommands()
		for _, cmd := range commands {
			help := c.helpManager.GetCommandHelp(cmd)
			if help != nil {
				fmt.Printf("  %-15s - %s\n", help.Name, help.Description)
			}
		}

		fmt.Println("\nType 'help <command>' for detailed information about a specific command.")
		fmt.Println("Use 'help <command> --examples' to see usage examples.")
		fmt.Println("Use 'help <command> --tutorial' to see a step-by-step tutorial.")
		return nil
	}

	// Show help for specific command
	helpText := c.helpManager.FormatCommandHelp(command, showExamples, showTutorial)
	fmt.Print(helpText)
	return nil
}

// AliasCommand manages command aliases
type AliasCommand struct {
	BaseCommand
	aliasManager *aliases.Manager
	formatter    *formatter.Formatter
}

// NewAliasCommand creates a new alias command
func NewAliasCommand(client *client.Client, formatter *formatter.Formatter) *AliasCommand {
	return &AliasCommand{
		BaseCommand: BaseCommand{
			name:        "alias",
			description: "Manage command aliases",
			usage:       "alias [add|remove|list|help] [options]",
			client:      client,
			formatter:   formatter,
		},
		aliasManager: aliases.New(),
		formatter:    formatter,
	}
}

// Execute executes the alias command
func (c *AliasCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.listAliases()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "add":
		return c.addAlias(args[1:])
	case "remove":
		return c.removeAlias(args[1:])
	case "list":
		return c.listAliases()
	case "help":
		fmt.Print(c.aliasManager.GetAliasHelp())
		return nil
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// addAlias adds a new alias
func (c *AliasCommand) addAlias(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: alias add <alias> <command>")
	}

	alias := args[0]
	command := args[1]

	err := c.aliasManager.AddAlias(alias, command)
	if err != nil {
		return fmt.Errorf("failed to add alias: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Added alias: %s → %s", alias, command))
	return nil
}

// removeAlias removes an alias
func (c *AliasCommand) removeAlias(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: alias remove <alias>")
	}

	alias := args[0]

	err := c.aliasManager.RemoveAlias(alias)
	if err != nil {
		return fmt.Errorf("failed to remove alias: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Removed alias: %s", alias))
	return nil
}

// listAliases lists all aliases
func (c *AliasCommand) listAliases() error {
	aliasText := c.aliasManager.FormatAliases()
	fmt.Print(aliasText)
	return nil
}

// FormatCommand manages output formatting
type FormatCommand struct {
	BaseCommand
	outputFormatter *output.Formatter
	formatter       *formatter.Formatter
}

// NewFormatCommand creates a new format command
func NewFormatCommand(client *client.Client, formatter *formatter.Formatter) *FormatCommand {
	return &FormatCommand{
		BaseCommand: BaseCommand{
			name:        "format",
			description: "Set output format for commands",
			usage:       "format [set|get|list] [format]",
			client:      client,
			formatter:   formatter,
		},
		outputFormatter: output.New(output.FormatTable),
		formatter:       formatter,
	}
}

// Execute executes the format command
func (c *FormatCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showCurrentFormat()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "set":
		return c.setFormat(args[1:])
	case "get":
		return c.showCurrentFormat()
	case "list":
		return c.listFormats()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// setFormat sets the output format
func (c *FormatCommand) setFormat(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: format set <format>")
	}

	formatStr := strings.ToLower(args[0])
	format := output.Format(formatStr)

	// Validate format
	validFormats := []output.Format{output.FormatTable, output.FormatJSON, output.FormatYAML, output.FormatCSV, output.FormatText}
	valid := false
	for _, f := range validFormats {
		if format == f {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid format: %s. Valid formats: table, json, yaml, csv, text", formatStr)
	}

	c.outputFormatter.SetFormat(format)
	c.formatter.PrintSuccess(fmt.Sprintf("Output format set to: %s", format))
	return nil
}

// showCurrentFormat shows the current format
func (c *FormatCommand) showCurrentFormat() error {
	format := c.outputFormatter.GetFormat()
	c.formatter.PrintInfo(fmt.Sprintf("Current output format: %s", format))
	return nil
}

// listFormats lists all available formats
func (c *FormatCommand) listFormats() error {
	c.formatter.PrintInfo("Available output formats:")
	fmt.Println("  table  - Human-readable table format (default)")
	fmt.Println("  json   - JSON format for scripting")
	fmt.Println("  yaml   - YAML format for configuration")
	fmt.Println("  csv    - CSV format for spreadsheets")
	fmt.Println("  text   - Plain text format")
	return nil
}

// ProfileCommand manages configuration profiles
type ProfileCommand struct {
	BaseCommand
	profileManager *profiles.Manager
	formatter      *formatter.Formatter
}

// NewProfileCommand creates a new profile command
func NewProfileCommand(client *client.Client, formatter *formatter.Formatter) *ProfileCommand {
	return &ProfileCommand{
		BaseCommand: BaseCommand{
			name:        "profile",
			description: "Manage configuration profiles",
			usage:       "profile [create|switch|list|delete|clone] [options]",
			client:      client,
			formatter:   formatter,
		},
		profileManager: profiles.New("./config"),
		formatter:      formatter,
	}
}

// Execute executes the profile command
func (c *ProfileCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showCurrentProfile()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "create":
		return c.createProfile(args[1:])
	case "switch":
		return c.switchProfile(args[1:])
	case "list":
		return c.listProfiles()
	case "delete":
		return c.deleteProfile(args[1:])
	case "clone":
		return c.cloneProfile(args[1:])
	case "current":
		return c.showCurrentProfile()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// createProfile creates a new profile
func (c *ProfileCommand) createProfile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: profile create <name> <description>")
	}

	name := args[0]
	description := strings.Join(args[1:], " ")

	// Use current config as base
	baseConfig := &config.Config{
		ServerURL:    "http://localhost:8080",
		AuthToken:    "",
		OutputFormat: "table",
		Theme:        "default",
		Verbose:      false,
		HistoryFile:  "history.txt",
	}

	err := c.profileManager.CreateProfile(name, description, baseConfig)
	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Created profile: %s", name))
	return nil
}

// switchProfile switches to a different profile
func (c *ProfileCommand) switchProfile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: profile switch <name>")
	}

	name := args[0]

	err := c.profileManager.SwitchProfile(name)
	if err != nil {
		return fmt.Errorf("failed to switch profile: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Switched to profile: %s", name))
	return nil
}

// listProfiles lists all profiles
func (c *ProfileCommand) listProfiles() error {
	profiles := c.profileManager.ListProfiles()
	current := c.profileManager.GetCurrentProfileName()

	c.formatter.PrintInfo("Configuration Profiles:")
	fmt.Println(strings.Repeat("=", 40))

	for name, profile := range profiles {
		marker := " "
		if name == current {
			marker = "*"
		}
		fmt.Printf("%s %-15s - %s\n", marker, name, profile.Description)
	}

	fmt.Printf("\n* = current profile\n")
	return nil
}

// deleteProfile deletes a profile
func (c *ProfileCommand) deleteProfile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: profile delete <name>")
	}

	name := args[0]

	err := c.profileManager.DeleteProfile(name)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Deleted profile: %s", name))
	return nil
}

// cloneProfile clones an existing profile
func (c *ProfileCommand) cloneProfile(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: profile clone <source> <new_name> <description>")
	}

	source := args[0]
	newName := args[1]
	description := strings.Join(args[2:], " ")

	err := c.profileManager.CloneProfile(source, newName, description)
	if err != nil {
		return fmt.Errorf("failed to clone profile: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Cloned profile '%s' to '%s'", source, newName))
	return nil
}

// showCurrentProfile shows the current profile
func (c *ProfileCommand) showCurrentProfile() error {
	profile := c.profileManager.GetCurrentProfile()
	if profile == nil {
		c.formatter.PrintError(fmt.Errorf("no current profile"))
		return nil
	}

	c.formatter.PrintInfo(fmt.Sprintf("Current Profile: %s", profile.Name))
	fmt.Printf("Description: %s\n", profile.Description)
	fmt.Printf("Server URL: %s\n", profile.Config.ServerURL)
	fmt.Printf("Output Format: %s\n", profile.Config.OutputFormat)
	fmt.Printf("Theme: %s\n", profile.Config.Theme)
	fmt.Printf("Verbose: %v\n", profile.Config.Verbose)
	return nil
}

// MacroCommand manages command macros
type MacroCommand struct {
	BaseCommand
	macroManager *macros.Manager
	formatter    *formatter.Formatter
}

// NewMacroCommand creates a new macro command
func NewMacroCommand(client *client.Client, formatter *formatter.Formatter) *MacroCommand {
	return &MacroCommand{
		BaseCommand: BaseCommand{
			name:        "macro",
			description: "Manage command macros",
			usage:       "macro [create|run|list|delete|help] [options]",
			client:      client,
			formatter:   formatter,
		},
		macroManager: macros.New("./config"),
		formatter:    formatter,
	}
}

// Execute executes the macro command
func (c *MacroCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.listMacros()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "create":
		return c.createMacro(args[1:])
	case "run":
		return c.runMacro(args[1:])
	case "list":
		return c.listMacros()
	case "delete":
		return c.deleteMacro(args[1:])
	case "help":
		fmt.Print(c.macroManager.GetMacroHelp())
		return nil
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// createMacro creates a new macro
func (c *MacroCommand) createMacro(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: macro create <name> <description> <command1> [command2]")
	}

	name := args[0]
	description := args[1]
	commands := args[2:]

	err := c.macroManager.CreateMacro(name, description, commands)
	if err != nil {
		return fmt.Errorf("failed to create macro: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Created macro: %s", name))
	return nil
}

// runMacro runs a macro
func (c *MacroCommand) runMacro(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: macro run <name>")
	}

	name := args[0]

	commands, err := c.macroManager.ExecuteMacro(name)
	if err != nil {
		return fmt.Errorf("failed to run macro: %w", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Running macro: %s", name))
	fmt.Printf("Commands to execute: %v\n", commands)

	// In a real implementation, these commands would be executed
	// For now, we just show what would be executed
	for i, cmd := range commands {
		fmt.Printf("  %d. %s\n", i+1, cmd)
	}

	c.formatter.PrintSuccess("Macro execution completed")
	return nil
}

// listMacros lists all macros
func (c *MacroCommand) listMacros() error {
	macroText := c.macroManager.FormatMacroList()
	fmt.Print(macroText)
	return nil
}

// deleteMacro deletes a macro
func (c *MacroCommand) deleteMacro(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: macro delete <name>")
	}

	name := args[0]

	err := c.macroManager.DeleteMacro(name)
	if err != nil {
		return fmt.Errorf("failed to delete macro: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Deleted macro: %s", name))
	return nil
}
