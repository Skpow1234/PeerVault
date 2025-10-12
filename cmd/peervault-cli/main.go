package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli"
	"github.com/Skpow1234/Peervault/internal/cli/aliases"
	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/commands"
	"github.com/Skpow1234/Peervault/internal/cli/config"
	"github.com/Skpow1234/Peervault/internal/cli/edge"
	"github.com/Skpow1234/Peervault/internal/cli/files"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/history"
	"github.com/Skpow1234/Peervault/internal/cli/iot"
	"github.com/Skpow1234/Peervault/internal/cli/network"
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

	// Initialize alias manager
	aliasManager := aliases.New()

	// Initialize file managers
	configDir := "config" // Default config directory
	versionManager := files.NewVersionManager(client, configDir)
	shareManager := files.NewShareManager(client, configDir)
	compressionManager := files.NewCompressionManager(client, configDir)
	deduplicationManager := files.NewDeduplicationManager(client, configDir)
	streamingManager := files.NewStreamingManager(client, configDir)

	// Initialize network managers
	loadBalancer := network.NewLoadBalancer(client, configDir)
	cacheManager := network.NewCacheManager(client, configDir)
	cdnManager := network.NewCDNManager(client, configDir)
	bandwidthManager := network.NewBandwidthManager(client, configDir)

	// Initialize IoT and Edge managers
	deviceManager := iot.NewDeviceManager(client, configDir)
	edgeManager := edge.NewEdgeManager(client, configDir)

	// Register commands
	registerCommands(cliApp, client, formatter, hist, aliasManager, versionManager, shareManager, compressionManager, deduplicationManager, streamingManager, loadBalancer, cacheManager, cdnManager, bandwidthManager, deviceManager, edgeManager)

	// Start interactive mode
	runInteractiveMode(cliApp, client, formatter, prompt, cfg, hist, aliasManager)
}

func registerCommands(cliApp *cli.CLI, client *client.Client, formatter *formatter.Formatter, hist *history.History, aliasManager *aliases.Manager, versionManager *files.VersionManager, shareManager *files.ShareManager, compressionManager *files.CompressionManager, deduplicationManager *files.DeduplicationManager, streamingManager *files.StreamingManager, loadBalancer *network.LoadBalancer, cacheManager *network.CacheManager, cdnManager *network.CDNManager, bandwidthManager *network.BandwidthManager, deviceManager *iot.DeviceManager, edgeManager *edge.EdgeManager) {
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
	cliApp.RegisterCommand("iot", commands.NewIoTCommand(client, formatter, deviceManager))
	cliApp.RegisterCommand("devices", commands.NewIoTCommand(client, formatter, deviceManager)) // Alias

	// Edge Computing operations
	cliApp.RegisterCommand("edge", commands.NewEdgeCommand(client, formatter, edgeManager))

	// Backup operations
	cliApp.RegisterCommand("backup", commands.NewBackupCommand(client, formatter))
	cliApp.RegisterCommand("restore", commands.NewRestoreCommand(client, formatter))

	// Configuration
	cliApp.RegisterCommand("config", commands.NewConfigCommand(client, formatter))
	cliApp.RegisterCommand("set", commands.NewSetCommand(client, formatter))
	cliApp.RegisterCommand("get", commands.NewGetConfigCommand(client, formatter))

	// Real-time commands
	cliApp.RegisterCommand("realtime", commands.NewRealtimeCommand(client, formatter))

	// Advanced commands
	cliApp.RegisterCommand("protocol", commands.NewProtocolCommand(client, formatter))
	cliApp.RegisterCommand("batch", commands.NewBatchCommand(client, formatter))
	cliApp.RegisterCommand("monitor", commands.NewMonitorCommand(client, formatter))

	// Quick Wins commands
	cliApp.RegisterCommand("alias", commands.NewAliasCommand(client, formatter))
	cliApp.RegisterCommand("format", commands.NewFormatCommand(client, formatter))
	cliApp.RegisterCommand("profile", commands.NewProfileCommand(client, formatter))
	cliApp.RegisterCommand("macro", commands.NewMacroCommand(client, formatter))

	// Security commands
	cliApp.RegisterCommand("auth", commands.NewAuthCommand(client, formatter))
	cliApp.RegisterCommand("cert", commands.NewCertCommand(client, formatter))
	cliApp.RegisterCommand("audit", commands.NewAuditCommand(client, formatter))

	// Backup commands
	cliApp.RegisterCommand("backup", commands.NewBackupCommand(client, formatter))

	// Advanced File Operations
	cliApp.RegisterCommand("version", commands.NewVersionCommand(client, formatter, versionManager))
	cliApp.RegisterCommand("share", commands.NewShareCommand(client, formatter, shareManager))
	cliApp.RegisterCommand("compress", commands.NewCompressionCommand(client, formatter, compressionManager))
	cliApp.RegisterCommand("dedup", commands.NewDeduplicationCommand(client, formatter, deduplicationManager))
	cliApp.RegisterCommand("stream", commands.NewStreamingCommand(client, formatter, streamingManager))

	// Network Optimization
	cliApp.RegisterCommand("lb", commands.NewLoadBalancerCommand(client, formatter, loadBalancer))
	cliApp.RegisterCommand("cache", commands.NewCacheCommand(client, formatter, cacheManager))
	cliApp.RegisterCommand("cdn", commands.NewCDNCommand(client, formatter, cdnManager))
	cliApp.RegisterCommand("bandwidth", commands.NewBandwidthCommand(client, formatter, bandwidthManager))

	// Utility commands
	cliApp.RegisterCommand("help", commands.NewEnhancedHelpCommand(cliApp))
	cliApp.RegisterCommand("exit", commands.NewExitCommand())
	cliApp.RegisterCommand("quit", commands.NewExitCommand()) // Alias
	cliApp.RegisterCommand("clear", commands.NewClearCommand())
	cliApp.RegisterCommand("history", commands.NewHistoryCommand(hist))
}

func runInteractiveMode(cliApp *cli.CLI, client *client.Client, formatter *formatter.Formatter, prompt *prompt.Prompt, cfg *config.Config, hist *history.History, aliasManager *aliases.Manager) {
	// Clear screen and show welcome
	formatter.ClearScreen()
	formatter.PrintHeader("🚀 PeerVault CLI - Interactive Mode")
	formatter.PrintInfo("Type 'help' for available commands or 'exit' to quit")
	formatter.PrintInfo("Use Tab for completion, ↑↓ for history, Ctrl+C to exit")
	fmt.Println()

	ctx := context.Background()

	for {
		// Get user input with advanced prompt
		input, err := prompt.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\n👋 Goodbye!")
				break
			}
			formatter.PrintError(fmt.Errorf("error reading input: %v", err))
			continue
		}

		// Skip empty input
		if strings.TrimSpace(input) == "" {
			continue
		}

		// Expand aliases
		expandedInput := aliasManager.ExpandAliases(input)

		// Add to history
		prompt.AddToHistory(expandedInput)

		// Parse and execute command
		parts := strings.Fields(expandedInput)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		// Show spinner for long operations
		spinner := formatter.PrintSpinner("Executing command...")

		// Execute command
		err = cliApp.Execute(ctx, command, args)
		spinner.Stop()

		if err != nil {
			formatter.PrintError(err)
		}

		fmt.Println() // Add spacing between commands
	}
}
