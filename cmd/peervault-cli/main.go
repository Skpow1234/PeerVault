package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli"
	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/commands"
	"github.com/Skpow1234/Peervault/internal/cli/config"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/history"
	"github.com/Skpow1234/Peervault/internal/cli/prompt"
)

func main() {
	// Initialize CLI
	cliApp := cli.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v\n", err)
		cfg = config.Default()
	}

	// Initialize client
	client := client.New(cfg)

	// Initialize formatter
	formatter := formatter.New()

	// Initialize history
	hist := history.New(cfg.HistoryFile)

	// Initialize prompt
	prompt := prompt.New(cfg, hist)

	// Register commands
	registerCommands(cliApp, client, formatter, hist)

	// Start interactive mode
	runInteractiveMode(cliApp, client, formatter, prompt, cfg, hist)
}

func registerCommands(cliApp *cli.CLI, client *client.Client, formatter *formatter.Formatter, hist *history.History) {
	// File operations
	cliApp.RegisterCommand("store", commands.NewStoreCommand(client, formatter))
	cliApp.RegisterCommand("get", commands.NewGetCommand(client, formatter))
	cliApp.RegisterCommand("list", commands.NewListCommand(client, formatter))
	cliApp.RegisterCommand("delete", commands.NewDeleteCommand(client, formatter))
	cliApp.RegisterCommand("ls", commands.NewListCommand(client, formatter)) // Alias

	// Peer operations
	cliApp.RegisterCommand("peers", commands.NewPeersCommand(client, formatter))
	cliApp.RegisterCommand("connect", commands.NewConnectCommand(client, formatter))
	cliApp.RegisterCommand("disconnect", commands.NewDisconnectCommand(client, formatter))

	// System operations
	cliApp.RegisterCommand("health", commands.NewHealthCommand(client, formatter))
	cliApp.RegisterCommand("metrics", commands.NewMetricsCommand(client, formatter))
	cliApp.RegisterCommand("status", commands.NewStatusCommand(client, formatter))

	// Blockchain operations
	cliApp.RegisterCommand("blockchain", commands.NewBlockchainCommand(client, formatter))
	cliApp.RegisterCommand("bc", commands.NewBlockchainCommand(client, formatter)) // Alias

	// IoT operations
	cliApp.RegisterCommand("iot", commands.NewIoTCommand(client, formatter))
	cliApp.RegisterCommand("devices", commands.NewDevicesCommand(client, formatter))

	// Backup operations
	cliApp.RegisterCommand("backup", commands.NewBackupCommand(client, formatter))
	cliApp.RegisterCommand("restore", commands.NewRestoreCommand(client, formatter))

	// Configuration
	cliApp.RegisterCommand("config", commands.NewConfigCommand(client, formatter))
	cliApp.RegisterCommand("set", commands.NewSetCommand(client, formatter))
	cliApp.RegisterCommand("get", commands.NewGetConfigCommand(client, formatter))

	// Utility commands
	cliApp.RegisterCommand("help", commands.NewHelpCommand(cliApp))
	cliApp.RegisterCommand("exit", commands.NewExitCommand())
	cliApp.RegisterCommand("quit", commands.NewExitCommand()) // Alias
	cliApp.RegisterCommand("clear", commands.NewClearCommand())
	cliApp.RegisterCommand("history", commands.NewHistoryCommand(hist))
}

func runInteractiveMode(cliApp *cli.CLI, client *client.Client, formatter *formatter.Formatter, prompt *prompt.Prompt, cfg *config.Config, hist *history.History) {
	fmt.Println("🚀 PeerVault CLI - Interactive Mode")
	fmt.Println("Type 'help' for available commands or 'exit' to quit")
	fmt.Println()

	ctx := context.Background()

	for {
		// Get user input with prompt
		input, err := prompt.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\nGoodbye!")
				break
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		// Skip empty input
		if strings.TrimSpace(input) == "" {
			continue
		}

		// Add to history
		prompt.AddToHistory(input)

		// Parse and execute command
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		// Execute command
		err = cliApp.Execute(ctx, command, args)
		if err != nil {
			formatter.PrintError(err)
		}

		fmt.Println() // Add spacing between commands
	}
}
