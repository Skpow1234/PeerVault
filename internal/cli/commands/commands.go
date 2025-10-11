package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli"
	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
	"github.com/Skpow1234/Peervault/internal/cli/history"
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
		c.formatter.PrintInfo(fmt.Sprintf("File would be saved to: %s", outputPath))
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

	c.formatter.PrintSuccess(fmt.Sprintf("Connected to: %s", address))

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

// IoTCommand handles IoT operations
type IoTCommand struct {
	BaseCommand
}

// NewIoTCommand creates a new IoT command
func NewIoTCommand(client *client.Client, formatter *formatter.Formatter) *IoTCommand {
	return &IoTCommand{
		BaseCommand: BaseCommand{
			name:        "iot",
			description: "IoT device operations",
			usage:       "iot [devices|sensors|commands] [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the IoT command
func (c *IoTCommand) Execute(ctx context.Context, args []string) error {
	action := "devices"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "devices":
		c.formatter.PrintInfo("IoT devices feature not yet implemented")
	case "sensors":
		c.formatter.PrintInfo("IoT sensors feature not yet implemented")
	case "commands":
		c.formatter.PrintInfo("IoT commands feature not yet implemented")
	default:
		return fmt.Errorf("unknown action: %s. Use 'devices', 'sensors', or 'commands'", action)
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

// BackupCommand handles backup operations
type BackupCommand struct {
	BaseCommand
}

// NewBackupCommand creates a new backup command
func NewBackupCommand(client *client.Client, formatter *formatter.Formatter) *BackupCommand {
	return &BackupCommand{
		BaseCommand: BaseCommand{
			name:        "backup",
			description: "Backup operations",
			usage:       "backup [create|list|restore] [options]",
			client:      client,
			formatter:   formatter,
		},
	}
}

// Execute executes the backup command
func (c *BackupCommand) Execute(ctx context.Context, args []string) error {
	action := "list"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "create":
		c.formatter.PrintInfo("Backup creation feature not yet implemented")
	case "list":
		c.formatter.PrintInfo("Backup listing feature not yet implemented")
	case "restore":
		c.formatter.PrintInfo("Backup restore feature not yet implemented")
	default:
		return fmt.Errorf("unknown action: %s. Use 'create', 'list', or 'restore'", action)
	}

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
