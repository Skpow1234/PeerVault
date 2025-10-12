package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli/backup"
	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// BackupCommand manages backup and recovery operations
type BackupCommand struct {
	BaseCommand
	backupManager *backup.Manager
}

// NewBackupCommand creates a new backup command
func NewBackupCommand(client *client.Client, formatter *formatter.Formatter) *BackupCommand {
	return &BackupCommand{
		BaseCommand: BaseCommand{
			name:        "backup",
			description: "Manage backup and recovery operations",
			usage:       "backup [create|list|restore|delete|schedule|cleanup] [options]",
			client:      client,
			formatter:   formatter,
		},
		backupManager: backup.New(client, "./config"),
	}
}

// Execute executes the backup command
func (c *BackupCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showBackupHelp()
	}

	subcommand := strings.ToLower(args[0])
	switch subcommand {
	case "create":
		return c.createBackup(args[1:])
	case "list":
		return c.listBackups()
	case "restore":
		return c.restoreBackup(args[1:])
	case "delete":
		return c.deleteBackup(args[1:])
	case "schedule":
		return c.scheduleBackup(args[1:])
	case "cleanup":
		return c.cleanupBackups(args[1:])
	case "status":
		return c.showBackupStatus(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// createBackup creates a new backup
func (c *BackupCommand) createBackup(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: backup create <name> <type> [options]")
	}

	name := args[0]
	backupType := args[1]

	// Validate backup type
	validTypes := []string{"full", "incremental", "differential"}
	if !contains(validTypes, backupType) {
		return fmt.Errorf("invalid backup type: %s. Valid types: %s", backupType, strings.Join(validTypes, ", "))
	}

	// Parse options
	config := &backup.BackupConfig{
		Name:        name,
		Type:        backupType,
		Compression: true,
		Encryption:  false,
		Destination: "./backups",
		Include:     []string{},
		Exclude:     []string{},
		Options:     make(map[string]string),
	}

	// Parse additional options
	for i := 2; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}

		option := args[i]
		value := args[i+1]

		switch option {
		case "--compression":
			config.Compression = value == "true"
		case "--encryption":
			config.Encryption = value == "true"
		case "--destination":
			config.Destination = value
		case "--include":
			config.Include = strings.Split(value, ",")
		case "--exclude":
			config.Exclude = strings.Split(value, ",")
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Creating %s backup: %s", backupType, name))

	// Create backup
	backupObj, err := c.backupManager.CreateBackup(config)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Start backup
	err = c.backupManager.StartBackup(backupObj.ID, config)
	if err != nil {
		return fmt.Errorf("failed to start backup: %w", err)
	}

	c.formatter.PrintSuccess(fmt.Sprintf("Backup created and started: %s", backupObj.ID))
	c.formatter.PrintInfo(fmt.Sprintf("Backup file: %s", backupObj.FilePath))

	return nil
}

// listBackups lists all backups
func (c *BackupCommand) listBackups() error {
	backups := c.backupManager.ListBackups()

	if len(backups) == 0 {
		c.formatter.PrintInfo("No backups found")
		return nil
	}

	c.formatter.PrintInfo("Backups:")
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-15s %-20s %-12s %-12s %-20s %-10s %-15s %-10s\n", "ID", "Name", "Type", "Status", "Created", "Size", "File Count", "Duration")
	fmt.Println(strings.Repeat("-", 120))

	for _, backup := range backups {
		duration := ""
		if !backup.StartedAt.IsZero() && !backup.CompletedAt.IsZero() {
			duration = backup.CompletedAt.Sub(backup.StartedAt).String()
		} else if !backup.StartedAt.IsZero() {
			duration = "Running..."
		}

		size := formatBytes(backup.Size)

		fmt.Printf("%-15s %-20s %-12s %-12s %-20s %-10s %-15d %-10s\n",
			backup.ID[:12],
			backup.Name,
			backup.Type,
			backup.Status,
			backup.CreatedAt.Format("2006-01-02 15:04:05"),
			size,
			backup.FileCount,
			duration,
		)
	}

	return nil
}

// restoreBackup restores from a backup
func (c *BackupCommand) restoreBackup(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: backup restore <backup_id> <destination>")
	}

	backupID := args[0]
	destination := args[1]

	c.formatter.PrintInfo(fmt.Sprintf("Restoring backup %s to %s", backupID, destination))

	// Get backup info
	backupObj, err := c.backupManager.GetBackup(backupID)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	if backupObj.Status != "completed" {
		return fmt.Errorf("backup is not completed: %s", backupObj.Status)
	}

	// Restore backup
	err = c.backupManager.RestoreBackup(backupID, destination)
	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	c.formatter.PrintSuccess("Backup restored successfully")
	c.formatter.PrintInfo(fmt.Sprintf("Restored %d files to %s", backupObj.FileCount, destination))

	return nil
}

// deleteBackup deletes a backup
func (c *BackupCommand) deleteBackup(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: backup delete <backup_id>")
	}

	backupID := args[0]

	c.formatter.PrintInfo(fmt.Sprintf("Deleting backup: %s", backupID))

	err := c.backupManager.DeleteBackup(backupID)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	c.formatter.PrintSuccess("Backup deleted successfully")

	return nil
}

// scheduleBackup schedules a backup
func (c *BackupCommand) scheduleBackup(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: backup schedule <name> <type> <schedule> [options]")
	}

	name := args[0]
	backupType := args[1]
	schedule := args[2]

	// Validate backup type
	validTypes := []string{"full", "incremental", "differential"}
	if !contains(validTypes, backupType) {
		return fmt.Errorf("invalid backup type: %s. Valid types: %s", backupType, strings.Join(validTypes, ", "))
	}

	config := &backup.BackupConfig{
		Name:        name,
		Type:        backupType,
		Schedule:    schedule,
		Retention:   30, // Default 30 days
		Compression: true,
		Encryption:  false,
		Destination: "./backups",
		Include:     []string{},
		Exclude:     []string{},
		Options:     make(map[string]string),
	}

	// Parse additional options
	for i := 3; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}

		option := args[i]
		value := args[i+1]

		switch option {
		case "--retention":
			if days, err := fmt.Sscanf(value, "%d", &config.Retention); err != nil || days != 1 {
				return fmt.Errorf("invalid retention days: %s", value)
			}
		case "--compression":
			config.Compression = value == "true"
		case "--encryption":
			config.Encryption = value == "true"
		case "--destination":
			config.Destination = value
		case "--include":
			config.Include = strings.Split(value, ",")
		case "--exclude":
			config.Exclude = strings.Split(value, ",")
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Scheduling %s backup: %s", backupType, name))

	err := c.backupManager.ScheduleBackup(config)
	if err != nil {
		return fmt.Errorf("failed to schedule backup: %w", err)
	}

	c.formatter.PrintSuccess("Backup scheduled successfully")
	c.formatter.PrintInfo(fmt.Sprintf("Schedule: %s", schedule))
	c.formatter.PrintInfo(fmt.Sprintf("Retention: %d days", config.Retention))

	return nil
}

// cleanupBackups cleans up old backups
func (c *BackupCommand) cleanupBackups(args []string) error {
	retentionDays := 30 // Default

	if len(args) > 0 {
		if days, err := fmt.Sscanf(args[0], "%d", &retentionDays); err != nil || days != 1 {
			return fmt.Errorf("invalid retention days: %s", args[0])
		}
	}

	c.formatter.PrintInfo(fmt.Sprintf("Cleaning up backups older than %d days", retentionDays))

	err := c.backupManager.CleanupOldBackups(retentionDays)
	if err != nil {
		return fmt.Errorf("failed to cleanup backups: %w", err)
	}

	c.formatter.PrintSuccess("Backup cleanup completed")

	return nil
}

// showBackupStatus shows backup status
func (c *BackupCommand) showBackupStatus(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: backup status <backup_id>")
	}

	backupID := args[0]

	backupObj, err := c.backupManager.GetBackup(backupID)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	c.formatter.PrintInfo(fmt.Sprintf("Backup Status: %s", backupObj.Name))
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("ID: %s\n", backupObj.ID)
	fmt.Printf("Type: %s\n", backupObj.Type)
	fmt.Printf("Status: %s\n", backupObj.Status)
	fmt.Printf("Created: %s\n", backupObj.CreatedAt.Format("2006-01-02 15:04:05"))

	if !backupObj.StartedAt.IsZero() {
		fmt.Printf("Started: %s\n", backupObj.StartedAt.Format("2006-01-02 15:04:05"))
	}

	if !backupObj.CompletedAt.IsZero() {
		fmt.Printf("Completed: %s\n", backupObj.CompletedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Duration: %s\n", backupObj.CompletedAt.Sub(backupObj.StartedAt))
	}

	fmt.Printf("Size: %s\n", formatBytes(backupObj.Size))
	fmt.Printf("File Count: %d\n", backupObj.FileCount)
	fmt.Printf("File Path: %s\n", backupObj.FilePath)

	if backupObj.Error != "" {
		fmt.Printf("Error: %s\n", backupObj.Error)
	}

	// Show metadata
	if len(backupObj.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		for key, value := range backupObj.Metadata {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	return nil
}

// showBackupHelp shows backup help
func (c *BackupCommand) showBackupHelp() error {
	c.formatter.PrintInfo("Backup Commands:")
	fmt.Println("  create <name> <type> [options]     - Create a new backup")
	fmt.Println("  list                                - List all backups")
	fmt.Println("  restore <backup_id> <destination>  - Restore from backup")
	fmt.Println("  delete <backup_id>                  - Delete a backup")
	fmt.Println("  schedule <name> <type> <schedule>   - Schedule a backup")
	fmt.Println("  cleanup [retention_days]            - Cleanup old backups")
	fmt.Println("  status <backup_id>                  - Show backup status")
	fmt.Println("\nBackup Types:")
	fmt.Println("  full        - Complete backup of all files")
	fmt.Println("  incremental - Backup of files changed since last backup")
	fmt.Println("  differential - Backup of files changed since last full backup")
	fmt.Println("\nOptions:")
	fmt.Println("  --compression true/false    - Enable/disable compression")
	fmt.Println("  --encryption true/false     - Enable/disable encryption")
	fmt.Println("  --destination <path>        - Backup destination directory")
	fmt.Println("  --include <patterns>        - Include file patterns (comma-separated)")
	fmt.Println("  --exclude <patterns>        - Exclude file patterns (comma-separated)")
	fmt.Println("  --retention <days>          - Retention period in days")
	return nil
}

// Utility functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
