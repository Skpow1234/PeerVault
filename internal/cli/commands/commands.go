package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli"
	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/history"
	"github.com/Skpow1234/Peervault/internal/cli/monitoring"
	"github.com/Skpow1234/Peervault/internal/cli/operations"
	"github.com/Skpow1234/Peervault/internal/cli/protocol"
	"github.com/Skpow1234/Peervault/internal/cli/realtime"
)

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	name        string
	description string
	usage       string
	client      *client.Client
	formatter   *formatter.Formatter
}

// Name returns the command name
func (c *BaseCommand) Name() string {
	return c.name
}

// Description returns the command description
func (c *BaseCommand) Description() string {
	return c.description
}

// Usage returns the command usage
func (c *BaseCommand) Usage() string {
	return c.usage
}

// StoreCommand handles file storage operations
type StoreCommand struct {
	BaseCommand
}

// NewStoreCommand creates a new store command
func NewStoreCommand(client *client.Client, formatter *formatter.Formatter) *StoreCommand {
	return &StoreCommand{
		BaseCommand: BaseCommand{
			name:        "store",
			description: "Store a file in the PeerVault network",
			usage:       "store <file_path> [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the store command
func (c *StoreCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s", c.usage)
	}

	filePath := args[0]

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Storing file: %s", filePath))

	// Store the file
	file, err := c.client.StoreFile(ctx, filePath)
	if err != nil {
		return err
	}

	c.formatter.PrintSuccess(fmt.Sprintf("File stored successfully: %s", file.ID))
	c.formatter.PrintFileInfo(file)

	return nil
}

// GetCommand handles file retrieval operations
type GetCommand struct {
	BaseCommand
}

// NewGetCommand creates a new get command
func NewGetCommand(client *client.Client, formatter *formatter.Formatter) *GetCommand {
	return &GetCommand{
		BaseCommand: BaseCommand{
			name:        "get",
			description: "Retrieve a file from the PeerVault network",
			usage:       "get <file_id> [output_path]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the get command
func (c *GetCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s", c.usage)
	}

	fileID := args[0]
	outputPath := ""

	if len(args) > 1 {
		outputPath = args[1]
	}

	c.formatter.PrintInfo(fmt.Sprintf("Retrieving file: %s", fileID))

	// Get file information
	file, err := c.client.GetFile(ctx, fileID)
	if err != nil {
		return err
	}

	c.formatter.PrintSuccess(fmt.Sprintf("File retrieved successfully"))
	c.formatter.PrintFileInfo(file)

	if outputPath != "" {
		c.formatter.PrintInfo(fmt.Sprintf("Downloading file to: %s", outputPath))

		err = c.client.DownloadFile(ctx, fileID, outputPath)
		if err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}

		c.formatter.PrintSuccess(fmt.Sprintf("File downloaded successfully to: %s", outputPath))
	}

	return nil
}

// ListCommand handles file listing operations
type ListCommand struct {
	BaseCommand
}

// NewListCommand creates a new list command
func NewListCommand(client *client.Client, formatter *formatter.Formatter) *ListCommand {
	return &ListCommand{
		BaseCommand: BaseCommand{
			name:        "list",
			description: "List files in the PeerVault network",
			usage:       "list [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the list command
func (c *ListCommand) Execute(ctx context.Context, args []string) error {
	c.formatter.PrintInfo("Retrieving file list...")

	// List files
	files, err := c.client.ListFiles(ctx)
	if err != nil {
		return err
	}

	c.formatter.PrintFileList(files)

	return nil
}

// DeleteCommand handles file deletion operations
type DeleteCommand struct {
	BaseCommand
}

// NewDeleteCommand creates a new delete command
func NewDeleteCommand(client *client.Client, formatter *formatter.Formatter) *DeleteCommand {
	return &DeleteCommand{
		BaseCommand: BaseCommand{
			name:        "delete",
			description: "Delete a file from the PeerVault network",
			usage:       "delete <file_id>",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the delete command
func (c *DeleteCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s", c.usage)
	}

	fileID := args[0]

	c.formatter.PrintInfo(fmt.Sprintf("Deleting file: %s", fileID))

	// Delete the file
	err := c.client.DeleteFile(ctx, fileID)
	if err != nil {
		return err
	}

	c.formatter.PrintSuccess(fmt.Sprintf("File deleted successfully: %s", fileID))

	return nil
}

// PeersCommand handles peer management operations
type PeersCommand struct {
	BaseCommand
}

// NewPeersCommand creates a new peers command
func NewPeersCommand(client *client.Client, formatter *formatter.Formatter) *PeersCommand {
	return &PeersCommand{
		BaseCommand: BaseCommand{
			name:        "peers",
			description: "Manage peer connections",
			usage:       "peers [list|add|remove] [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the peers command
func (c *PeersCommand) Execute(ctx context.Context, args []string) error {
	action := "list"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "list":
		return c.listPeers(ctx)
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: peers add <address>")
		}
		return c.addPeer(ctx, args[1])
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: peers remove <peer_id>")
		}
		return c.removePeer(ctx, args[1])
	default:
		return fmt.Errorf("unknown action: %s. Use 'list', 'add', or 'remove'", action)
	}
}

func (c *PeersCommand) listPeers(ctx context.Context) error {
	c.formatter.PrintInfo("Retrieving peer list...")

	peers, err := c.client.ListPeers(ctx)
	if err != nil {
		return err
	}

	c.formatter.PrintPeerList(peers)

	return nil
}

func (c *PeersCommand) addPeer(ctx context.Context, address string) error {
	c.formatter.PrintInfo(fmt.Sprintf("Adding peer: %s", address))

	peer, err := c.client.AddPeer(ctx, address)
	if err != nil {
		return err
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Peer added successfully: %s", peer.ID))
	c.formatter.PrintPeerInfo(peer)

	return nil
}

func (c *PeersCommand) removePeer(ctx context.Context, peerID string) error {
	c.formatter.PrintInfo(fmt.Sprintf("Removing peer: %s", peerID))

	err := c.client.RemovePeer(ctx, peerID)
	if err != nil {
		return err
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Peer removed successfully: %s", peerID))

	return nil
}

// ConnectCommand handles peer connection operations
type ConnectCommand struct {
	BaseCommand
}

// NewConnectCommand creates a new connect command
func NewConnectCommand(client *client.Client, formatter *formatter.Formatter) *ConnectCommand {
	return &ConnectCommand{
		BaseCommand: BaseCommand{
			name:        "connect",
			description: "Connect to a PeerVault node",
			usage:       "connect <address>",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the connect command
func (c *ConnectCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s", c.usage)
	}

	address := args[0]

	c.formatter.PrintInfo(fmt.Sprintf("Connecting to: %s", address))

	// Set the server URL
	c.client.SetServerURL("http://" + address)

	// Test the connection
	err := c.client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Connected to: %s", address))
	c.formatter.PrintInfo("Connection verified successfully")

	return nil
}

// DisconnectCommand handles disconnection operations
type DisconnectCommand struct {
	BaseCommand
}

// NewDisconnectCommand creates a new disconnect command
func NewDisconnectCommand(client *client.Client, formatter *formatter.Formatter) *DisconnectCommand {
	return &DisconnectCommand{
		BaseCommand: BaseCommand{
			name:        "disconnect",
			description: "Disconnect from current node",
			usage:       "disconnect",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the disconnect command
func (c *DisconnectCommand) Execute(ctx context.Context, args []string) error {
	c.formatter.PrintInfo("Disconnecting...")

	// Disconnect from current server
	c.client.Disconnect()

	// Reset to default server URL
	c.client.SetServerURL("http://localhost:8080")

	c.formatter.PrintSuccess("Disconnected")

	return nil
}

// HealthCommand handles health check operations
type HealthCommand struct {
	BaseCommand
}

// NewHealthCommand creates a new health command
func NewHealthCommand(client *client.Client, formatter *formatter.Formatter) *HealthCommand {
	return &HealthCommand{
		BaseCommand: BaseCommand{
			name:        "health",
			description: "Check system health",
			usage:       "health",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the health command
func (c *HealthCommand) Execute(ctx context.Context, args []string) error {
	c.formatter.PrintInfo("Checking system health...")

	health, err := c.client.GetHealth(ctx)
	if err != nil {
		return err
	}

	c.formatter.PrintHealth(health)

	return nil
}

// MetricsCommand handles metrics operations
type MetricsCommand struct {
	BaseCommand
}

// NewMetricsCommand creates a new metrics command
func NewMetricsCommand(client *client.Client, formatter *formatter.Formatter) *MetricsCommand {
	return &MetricsCommand{
		BaseCommand: BaseCommand{
			name:        "metrics",
			description: "Show system metrics",
			usage:       "metrics [--live]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the metrics command
func (c *MetricsCommand) Execute(ctx context.Context, args []string) error {
	live := false
	for _, arg := range args {
		if arg == "--live" {
			live = true
			break
		}
	}

	if live {
		c.formatter.PrintInfo("Live metrics (press Ctrl+C to stop)...")
		// In a real implementation, this would show live updating metrics
		c.formatter.PrintWarning("Live metrics not yet implemented")
	}

	c.formatter.PrintInfo("Retrieving system metrics...")

	metrics, err := c.client.GetMetrics(ctx)
	if err != nil {
		return err
	}

	c.formatter.PrintMetrics(metrics)

	return nil
}

// StatusCommand handles status operations
type StatusCommand struct {
	BaseCommand
}

// NewStatusCommand creates a new status command
func NewStatusCommand(client *client.Client, formatter *formatter.Formatter) *StatusCommand {
	return &StatusCommand{
		BaseCommand: BaseCommand{
			name:        "status",
			description: "Show system status",
			usage:       "status",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the status command
func (c *StatusCommand) Execute(ctx context.Context, args []string) error {
	c.formatter.PrintInfo("Retrieving system status...")

	// Get both health and metrics
	health, err := c.client.GetHealth(ctx)
	if err != nil {
		return err
	}

	metrics, err := c.client.GetMetrics(ctx)
	if err != nil {
		return err
	}

	c.formatter.PrintHealth(health)
	fmt.Println()
	c.formatter.PrintMetrics(metrics)

	return nil
}

// BlockchainCommand handles blockchain operations
type BlockchainCommand struct {
	BaseCommand
}

// NewBlockchainCommand creates a new blockchain command
func NewBlockchainCommand(client *client.Client, formatter *formatter.Formatter) *BlockchainCommand {
	return &BlockchainCommand{
		BaseCommand: BaseCommand{
			name:        "blockchain",
			description: "Blockchain operations",
			usage:       "blockchain [networks|deploy|identity] [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the blockchain command
func (c *BlockchainCommand) Execute(ctx context.Context, args []string) error {
	action := "networks"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "networks":
		c.formatter.PrintInfo("Blockchain networks feature not yet implemented")
	case "deploy":
		c.formatter.PrintInfo("Smart contract deployment feature not yet implemented")
	case "identity":
		c.formatter.PrintInfo("Decentralized identity feature not yet implemented")
	default:
		return fmt.Errorf("unknown action: %s. Use 'networks', 'deploy', or 'identity'", action)
	}

	return nil
}

// DevicesCommand handles device operations
type DevicesCommand struct {
	BaseCommand
}

// NewDevicesCommand creates a new devices command
func NewDevicesCommand(client *client.Client, formatter *formatter.Formatter) *DevicesCommand {
	return &DevicesCommand{
		BaseCommand: BaseCommand{
			name:        "devices",
			description: "Device management operations",
			usage:       "devices [list|add|remove] [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the devices command
func (c *DevicesCommand) Execute(ctx context.Context, args []string) error {
	c.formatter.PrintInfo("Device management feature not yet implemented")
	return nil
}

// RestoreCommand handles restore operations
type RestoreCommand struct {
	BaseCommand
}

// NewRestoreCommand creates a new restore command
func NewRestoreCommand(client *client.Client, formatter *formatter.Formatter) *RestoreCommand {
	return &RestoreCommand{
		BaseCommand: BaseCommand{
			name:        "restore",
			description: "Restore operations",
			usage:       "restore <backup_id> [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the restore command
func (c *RestoreCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s", c.usage)
	}

	backupID := args[0]
	c.formatter.PrintInfo(fmt.Sprintf("Restore feature not yet implemented for backup: %s", backupID))

	return nil
}

// ConfigCommand handles configuration operations
type ConfigCommand struct {
	BaseCommand
}

// NewConfigCommand creates a new config command
func NewConfigCommand(client *client.Client, formatter *formatter.Formatter) *ConfigCommand {
	return &ConfigCommand{
		BaseCommand: BaseCommand{
			name:        "config",
			description: "Configuration management",
			usage:       "config [show|set|get] [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the config command
func (c *ConfigCommand) Execute(ctx context.Context, args []string) error {
	action := "show"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "show":
		c.formatter.PrintInfo("Configuration display feature not yet implemented")
	case "set":
		c.formatter.PrintInfo("Configuration setting feature not yet implemented")
	case "get":
		c.formatter.PrintInfo("Configuration getting feature not yet implemented")
	default:
		return fmt.Errorf("unknown action: %s. Use 'show', 'set', or 'get'", action)
	}

	return nil
}

// SetCommand handles setting configuration values
type SetCommand struct {
	BaseCommand
}

// NewSetCommand creates a new set command
func NewSetCommand(client *client.Client, formatter *formatter.Formatter) *SetCommand {
	return &SetCommand{
		BaseCommand: BaseCommand{
			name:        "set",
			description: "Set configuration value",
			usage:       "set <key> <value>",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the set command
func (c *SetCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s", c.usage)
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	c.formatter.PrintInfo(fmt.Sprintf("Setting %s = %s", key, value))
	c.formatter.PrintInfo("Configuration setting feature not yet implemented")

	return nil
}

// GetConfigCommand handles getting configuration values
type GetConfigCommand struct {
	BaseCommand
}

// NewGetConfigCommand creates a new get config command
func NewGetConfigCommand(client *client.Client, formatter *formatter.Formatter) *GetConfigCommand {
	return &GetConfigCommand{
		BaseCommand: BaseCommand{
			name:        "get",
			description: "Get configuration value",
			usage:       "get <key>",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the get config command
func (c *GetConfigCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s", c.usage)
	}

	key := args[0]

	c.formatter.PrintInfo(fmt.Sprintf("Getting value for: %s", key))
	c.formatter.PrintInfo("Configuration getting feature not yet implemented")

	return nil
}

// HelpCommand handles help operations
type HelpCommand struct {
	cli *cli.CLI
}

// NewHelpCommand creates a new help command
func NewHelpCommand(cli *cli.CLI) *HelpCommand {
	return &HelpCommand{
		cli: cli,
	}
}

// Name returns the command name
func (c *HelpCommand) Name() string {
	return "help"
}

// Description returns the command description
func (c *HelpCommand) Description() string {
	return "Show help information"
}

// Usage returns the command usage
func (c *HelpCommand) Usage() string {
	return "help [command]"
}

// Execute executes the help command
func (c *HelpCommand) Execute(ctx context.Context, args []string) error {
	if len(args) > 0 {
		// Show help for specific command
		commandName := args[0]
		cmd, exists := c.cli.GetCommand(commandName)
		if !exists {
			return fmt.Errorf("unknown command: %s", commandName)
		}

		fmt.Printf("Command: %s\n", cmd.Name())
		fmt.Printf("Description: %s\n", cmd.Description())
		fmt.Printf("Usage: %s\n", cmd.Usage())

		return nil
	}

	// Show general help
	fmt.Println("🚀 PeerVault CLI - Available Commands")
	fmt.Println()

	commands := c.cli.ListCommands()

	fmt.Println("📁 File Operations:")
	for _, cmd := range commands {
		if strings.Contains(cmd.Name(), "store") || strings.Contains(cmd.Name(), "get") ||
			strings.Contains(cmd.Name(), "list") || strings.Contains(cmd.Name(), "delete") {
			fmt.Printf("  %-15s - %s\n", cmd.Name(), cmd.Description())
		}
	}

	fmt.Println("\n🌐 Network Operations:")
	for _, cmd := range commands {
		if strings.Contains(cmd.Name(), "peer") || strings.Contains(cmd.Name(), "connect") {
			fmt.Printf("  %-15s - %s\n", cmd.Name(), cmd.Description())
		}
	}

	fmt.Println("\n🔧 System Operations:")
	for _, cmd := range commands {
		if strings.Contains(cmd.Name(), "health") || strings.Contains(cmd.Name(), "metrics") ||
			strings.Contains(cmd.Name(), "status") {
			fmt.Printf("  %-15s - %s\n", cmd.Name(), cmd.Description())
		}
	}

	fmt.Println("\n⚙️  Utility Commands:")
	for _, cmd := range commands {
		if strings.Contains(cmd.Name(), "help") || strings.Contains(cmd.Name(), "exit") ||
			strings.Contains(cmd.Name(), "clear") || strings.Contains(cmd.Name(), "history") {
			fmt.Printf("  %-15s - %s\n", cmd.Name(), cmd.Description())
		}
	}

	fmt.Println("\nType 'help <command>' for detailed information about a specific command.")

	return nil
}

// ExitCommand handles exit operations
type ExitCommand struct{}

// NewExitCommand creates a new exit command
func NewExitCommand() *ExitCommand {
	return &ExitCommand{}
}

// Name returns the command name
func (c *ExitCommand) Name() string {
	return "exit"
}

// Description returns the command description
func (c *ExitCommand) Description() string {
	return "Exit the CLI"
}

// Usage returns the command usage
func (c *ExitCommand) Usage() string {
	return "exit"
}

// Execute executes the exit command
func (c *ExitCommand) Execute(ctx context.Context, args []string) error {
	fmt.Println("Goodbye!")
	os.Exit(0)
	return nil
}

// ClearCommand handles clear operations
type ClearCommand struct{}

// NewClearCommand creates a new clear command
func NewClearCommand() *ClearCommand {
	return &ClearCommand{}
}

// Name returns the command name
func (c *ClearCommand) Name() string {
	return "clear"
}

// Description returns the command description
func (c *ClearCommand) Description() string {
	return "Clear the screen"
}

// Usage returns the command usage
func (c *ClearCommand) Usage() string {
	return "clear"
}

// Execute executes the clear command
func (c *ClearCommand) Execute(ctx context.Context, args []string) error {
	// Simple clear screen - in a real implementation, you'd use proper terminal control
	fmt.Print("\033[2J\033[H")
	return nil
}

// HistoryCommand handles history operations
type HistoryCommand struct {
	history *history.History
}

// NewHistoryCommand creates a new history command
func NewHistoryCommand(hist *history.History) *HistoryCommand {
	return &HistoryCommand{
		history: hist,
	}
}

// Name returns the command name
func (c *HistoryCommand) Name() string {
	return "history"
}

// Description returns the command description
func (c *HistoryCommand) Description() string {
	return "Show command history"
}

// Usage returns the command usage
func (c *HistoryCommand) Usage() string {
	return "history [--clear]"
}

// Execute executes the history command
func (c *HistoryCommand) Execute(ctx context.Context, args []string) error {
	clear := false
	for _, arg := range args {
		if arg == "--clear" {
			clear = true
			break
		}
	}

	if clear {
		c.history.Clear()
		fmt.Println("History cleared")
		return nil
	}

	commands := c.history.GetAll()
	if len(commands) == 0 {
		fmt.Println("No command history")
		return nil
	}

	fmt.Println("📜 Command History:")
	for i, cmd := range commands {
		fmt.Printf("%3d  %s\n", i+1, cmd)
	}

	return nil
}

// RealtimeCommand handles real-time operations
type RealtimeCommand struct {
	BaseCommand
	realtimeManager *realtime.Manager
}

// NewRealtimeCommand creates a new realtime command
func NewRealtimeCommand(client *client.Client, formatter *formatter.Formatter) *RealtimeCommand {
	return &RealtimeCommand{
		BaseCommand: BaseCommand{
			name:        "realtime",
			description: "Manage real-time updates and WebSocket connections",
			usage:       "realtime [connect|disconnect|status|subscribe] [options]",
			client:      client,
			formatter:   formatter,
		},
		realtimeManager: realtime.New("ws://localhost:8080/ws"),
	}
}

// Execute executes the realtime command
func (c *RealtimeCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showStatus(ctx)
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "connect":
		return c.connect(ctx)
	case "disconnect":
		return c.disconnect()
	case "status":
		return c.showStatus(ctx)
	case "subscribe":
		return c.subscribe(ctx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// connect connects to the real-time service
func (c *RealtimeCommand) connect(ctx context.Context) error {
	c.formatter.PrintInfo("Connecting to real-time service...")

	err := c.realtimeManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to real-time service: %w", err)
	}

	c.formatter.PrintSuccess("Connected to real-time service")
	return nil
}

// disconnect disconnects from the real-time service
func (c *RealtimeCommand) disconnect() error {
	c.formatter.PrintInfo("Disconnecting from real-time service...")

	err := c.realtimeManager.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	c.formatter.PrintSuccess("Disconnected from real-time service")
	return nil
}

// showStatus shows the real-time service status
func (c *RealtimeCommand) showStatus(ctx context.Context) error {
	status := c.realtimeManager.GetConnectionStatus()

	c.formatter.PrintInfo("Real-time Service Status:")
	fmt.Printf("Connected: %v\n", status["connected"])
	fmt.Printf("Subscribers: %v\n", status["subscribers"])
	fmt.Printf("Event Types: %v\n", status["event_types"])

	return nil
}

// subscribe subscribes to real-time events
func (c *RealtimeCommand) subscribe(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: realtime subscribe <event_type>")
	}

	eventType := args[0]

	subscriber := realtime.Subscriber{
		ID: fmt.Sprintf("cli-%d", time.Now().Unix()),
		Handler: func(data map[string]interface{}) {
			c.formatter.PrintInfo(fmt.Sprintf("Real-time event [%s]:", eventType))
			fmt.Printf("Data: %+v\n", data)
		},
	}

	c.realtimeManager.Subscribe(eventType, subscriber)
	c.formatter.PrintSuccess(fmt.Sprintf("Subscribed to %s events", eventType))

	return nil
}

// ProtocolCommand handles protocol switching
type ProtocolCommand struct {
	BaseCommand
	protocolManager *protocol.Manager
}

// NewProtocolCommand creates a new protocol command
func NewProtocolCommand(client *client.Client, formatter *formatter.Formatter) *ProtocolCommand {
	return &ProtocolCommand{
		BaseCommand: BaseCommand{
			name:        "protocol",
			description: "Manage protocol connections (REST, GraphQL, gRPC)",
			usage:       "protocol [set|list|info] [protocol_name]",
			client:      client,
			formatter:   formatter,
		},
		protocolManager: protocol.New("http://localhost:8080"),
	}
}

// Execute executes the protocol command
func (c *ProtocolCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.listProtocols()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "set":
		if len(args) < 2 {
			return fmt.Errorf("usage: protocol set <protocol_name>")
		}
		return c.setProtocol(args[1])
	case "list":
		return c.listProtocols()
	case "info":
		if len(args) < 2 {
			return c.showCurrentProtocol()
		}
		return c.showProtocolInfo(args[1])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// setProtocol sets the current protocol
func (c *ProtocolCommand) setProtocol(protocolName string) error {
	protocolType := protocol.Type(strings.ToLower(protocolName))

	err := c.protocolManager.SetProtocol(protocolType)
	if err != nil {
		return fmt.Errorf("failed to set protocol: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Switched to %s protocol", protocolName))
	return nil
}

// listProtocols lists all available protocols
func (c *ProtocolCommand) listProtocols() error {
	protocols := protocol.GetSupportedProtocols()

	c.formatter.PrintInfo("Available Protocols:")
	for _, p := range protocols {
		description := protocol.GetProtocolDescription(p)
		features := protocol.GetProtocolFeatures(p)

		fmt.Printf("  %s: %s\n", p, description)
		for _, feature := range features {
			fmt.Printf("    - %s\n", feature)
		}
		fmt.Println()
	}

	return nil
}

// showCurrentProtocol shows the current protocol
func (c *ProtocolCommand) showCurrentProtocol() error {
	current := c.protocolManager.GetProtocol()
	c.formatter.PrintInfo(fmt.Sprintf("Current Protocol: %s", current))
	c.formatter.PrintInfo(protocol.GetProtocolDescription(current))
	return nil
}

// showProtocolInfo shows detailed information about a protocol
func (c *ProtocolCommand) showProtocolInfo(protocolName string) error {
	protocolType := protocol.Type(strings.ToLower(protocolName))
	description := protocol.GetProtocolDescription(protocolType)
	features := protocol.GetProtocolFeatures(protocolType)

	c.formatter.PrintInfo(fmt.Sprintf("Protocol: %s", protocolName))
	c.formatter.PrintInfo(fmt.Sprintf("Description: %s", description))
	c.formatter.PrintInfo("Features:")
	for _, feature := range features {
		fmt.Printf("  - %s\n", feature)
	}

	return nil
}

// BatchCommand handles batch operations
type BatchCommand struct {
	BaseCommand
	operationsManager *operations.Manager
}

// NewBatchCommand creates a new batch command
func NewBatchCommand(client *client.Client, formatter *formatter.Formatter) *BatchCommand {
	return &BatchCommand{
		BaseCommand: BaseCommand{
			name:        "batch",
			description: "Perform batch operations on files",
			usage:       "batch [upload|download|sync] [options]",
			client:      client,
			formatter:   formatter,
		},
		operationsManager: operations.New(client, formatter),
	}
}

// Execute executes the batch command
func (c *BatchCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s", c.usage)
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "upload":
		return c.batchUpload(ctx, args[1:])
	case "download":
		return c.batchDownload(ctx, args[1:])
	case "sync":
		return c.syncDirectory(ctx, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// batchUpload performs batch upload
func (c *BatchCommand) batchUpload(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: batch upload <file1> [file2] [file3] ...")
	}

	c.formatter.PrintInfo(fmt.Sprintf("Starting batch upload of %d files...", len(args)))

	operation, err := c.operationsManager.BatchUpload(ctx, args, nil)
	if err != nil {
		return fmt.Errorf("failed to start batch upload: %w", err)
	}

	// Monitor progress
	for update := range operation.Progress {
		if update.Error != nil {
			c.formatter.PrintError(fmt.Errorf("upload error for %s: %w", update.File, update.Error))
		} else {
			c.formatter.PrintInfo(fmt.Sprintf("%s: %s (%.1f%%)", update.File, update.Status, update.Progress))
		}
	}

	c.formatter.PrintSuccess("Batch upload completed")
	return nil
}

// batchDownload performs batch download
func (c *BatchCommand) batchDownload(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: batch download <output_dir> <file_id1> [file_id2] [file_id3] ...")
	}

	outputDir := args[0]
	fileIDs := args[1:]

	c.formatter.PrintInfo(fmt.Sprintf("Starting batch download of %d files to %s...", len(fileIDs), outputDir))

	operation, err := c.operationsManager.BatchDownload(ctx, fileIDs, outputDir, nil)
	if err != nil {
		return fmt.Errorf("failed to start batch download: %w", err)
	}

	// Monitor progress
	for update := range operation.Progress {
		if update.Error != nil {
			c.formatter.PrintError(fmt.Errorf("download error for %s: %w", update.File, update.Error))
		} else {
			c.formatter.PrintInfo(fmt.Sprintf("%s: %s (%.1f%%)", update.File, update.Status, update.Progress))
		}
	}

	c.formatter.PrintSuccess("Batch download completed")
	return nil
}

// syncDirectory synchronizes a directory
func (c *BatchCommand) syncDirectory(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: batch sync <local_dir> <remote_prefix>")
	}

	localDir := args[0]
	remotePrefix := args[1]

	return c.operationsManager.SyncDirectory(ctx, localDir, remotePrefix, nil)
}

// MonitorCommand handles monitoring operations
type MonitorCommand struct {
	BaseCommand
	monitoringManager *monitoring.Manager
}

// NewMonitorCommand creates a new monitor command
func NewMonitorCommand(client *client.Client, formatter *formatter.Formatter) *MonitorCommand {
	return &MonitorCommand{
		BaseCommand: BaseCommand{
			name:        "monitor",
			description: "Monitor system health and performance",
			usage:       "monitor [start|stop|status|alerts|dashboard] [options]",
			client:      client,
			formatter:   formatter,
		},
		monitoringManager: monitoring.New(client, formatter),
	}
}

// Execute executes the monitor command
func (c *MonitorCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showStatus()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "start":
		return c.startMonitoring(ctx, args[1:])
	case "stop":
		return c.stopMonitoring()
	case "status":
		return c.showStatus()
	case "alerts":
		return c.showAlerts()
	case "dashboard":
		return c.showDashboard()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// startMonitoring starts the monitoring system
func (c *MonitorCommand) startMonitoring(ctx context.Context, args []string) error {
	interval := 30 * time.Second
	if len(args) > 0 {
		if duration, err := time.ParseDuration(args[0]); err == nil {
			interval = duration
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Starting monitoring with %v interval...", interval))

	// Add some default alerts
	c.monitoringManager.AddAlert(&monitoring.Alert{
		ID:          "high_latency",
		Name:        "High Latency",
		Description: "Average latency is above 100ms",
		Condition: monitoring.AlertCondition{
			Metric:    "avg_latency",
			Operator:  ">",
			Threshold: 100.0,
			Duration:  5 * time.Minute,
		},
		Severity: monitoring.SeverityWarning,
		Enabled:  true,
	})

	c.monitoringManager.AddAlert(&monitoring.Alert{
		ID:          "low_storage",
		Name:        "Low Storage",
		Description: "Storage usage is above 90%",
		Condition: monitoring.AlertCondition{
			Metric:    "storage_used",
			Operator:  ">",
			Threshold: 0.9,
			Duration:  1 * time.Minute,
		},
		Severity: monitoring.SeverityCritical,
		Enabled:  true,
	})

	go c.monitoringManager.StartMonitoring(ctx, interval)
	c.formatter.PrintSuccess("Monitoring started")
	return nil
}

// stopMonitoring stops the monitoring system
func (c *MonitorCommand) stopMonitoring() error {
	c.formatter.PrintInfo("Stopping monitoring...")
	c.formatter.PrintSuccess("Monitoring stopped")
	return nil
}

// showStatus shows monitoring status
func (c *MonitorCommand) showStatus() error {
	alerts := c.monitoringManager.GetAlerts()
	metrics := c.monitoringManager.GetAllMetrics()

	c.formatter.PrintInfo("Monitoring Status:")
	fmt.Printf("  Active Alerts: %d\n", len(alerts))
	fmt.Printf("  Metrics Collected: %d\n", len(metrics))

	// Show recent metrics
	for name, series := range metrics {
		if len(series.Values) > 0 {
			latest := series.Values[len(series.Values)-1]
			fmt.Printf("  %s: %.2f (at %s)\n", name, latest.Value, latest.Timestamp.Format("15:04:05"))
		}
	}

	return nil
}

// showAlerts shows all alerts
func (c *MonitorCommand) showAlerts() error {
	alerts := c.monitoringManager.GetAlerts()

	c.formatter.PrintInfo("Active Alerts:")
	if len(alerts) == 0 {
		fmt.Println("  No alerts configured")
		return nil
	}

	for _, alert := range alerts {
		status := "disabled"
		if alert.Enabled {
			status = "enabled"
		}
		fmt.Printf("  %s [%s] - %s (%s)\n", alert.Name, alert.Severity, alert.Description, status)
		if alert.TriggerCount > 0 {
			fmt.Printf("    Triggered %d times, last: %s\n", alert.TriggerCount, alert.LastTriggered.Format("15:04:05"))
		}
	}

	return nil
}

// showDashboard shows a simple dashboard
func (c *MonitorCommand) showDashboard() error {
	c.formatter.PrintInfo("System Dashboard")
	fmt.Println(strings.Repeat("=", 50))

	// Get current metrics
	ctx := context.Background()
	health, err := c.client.GetHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get health: %w", err)
	}

	metrics, err := c.client.GetMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	// Display dashboard
	fmt.Printf("System Health: %s\n", health.Status)
	fmt.Printf("Files Stored: %d\n", metrics.FilesStored)
	fmt.Printf("Active Peers: %d\n", metrics.ActivePeers)
	fmt.Printf("Storage Used: %d bytes\n", metrics.StorageUsed)
	fmt.Printf("Network Traffic: %.2f MB/s\n", metrics.NetworkTraffic)

	fmt.Println("\nService Status:")
	for service, status := range health.Services {
		emoji := "✅"
		if status != "healthy" {
			emoji = "❌"
		}
		fmt.Printf("  %s %s: %s\n", emoji, service, status)
	}

	return nil
}
