package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/files"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// VersionCommand handles file versioning operations
type VersionCommand struct {
	BaseCommand
	versionManager *files.VersionManager
}

// NewVersionCommand creates a new version command
func NewVersionCommand(client *client.Client, formatter *formatter.Formatter, versionManager *files.VersionManager) *VersionCommand {
	return &VersionCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		versionManager: versionManager,
	}
}

// Execute executes the version command
func (c *VersionCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: version <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createVersion(subArgs)
	case "list":
		return c.listVersions(subArgs)
	case "current":
		return c.getCurrentVersion(subArgs)
	case "restore":
		return c.restoreVersion(subArgs)
	case "delete":
		return c.deleteVersion(subArgs)
	case "compare":
		return c.compareVersions(subArgs)
	case "all":
		return c.listAllVersions()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *VersionCommand) createVersion(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: version create <file_id> <description> <created_by> [tags...]")
	}

	fileID := args[0]
	description := args[1]
	createdBy := args[2]
	tags := args[3:]

	version, err := c.versionManager.CreateVersion(fileID, description, createdBy, tags)
	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	c.formatter.PrintSuccess("Version created successfully")
	c.formatter.PrintInfo("Version ID: " + version.ID)
	c.formatter.PrintInfo("Version: " + strconv.Itoa(version.Version))
	c.formatter.PrintInfo("Size: " + c.formatter.FormatBytes(version.Size))
	c.formatter.PrintInfo("Hash: " + version.Hash)
	c.formatter.PrintInfo("Created: " + version.CreatedAt.Format(time.RFC3339))
	c.formatter.PrintInfo("Created by: " + version.CreatedBy)
	c.formatter.PrintInfo("Description: " + version.Description)
	if len(version.Tags) > 0 {
		c.formatter.PrintInfo("Tags: " + strings.Join(version.Tags, ", "))
	}

	return nil
}

func (c *VersionCommand) listVersions(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: version list <file_id>")
	}

	fileID := args[0]
	versions, err := c.versionManager.GetVersions(fileID)
	if err != nil {
		return fmt.Errorf("failed to get versions: %w", err)
	}

	if len(versions) == 0 {
		c.formatter.PrintInfo("No versions found for file: " + fileID)
		return nil
	}

	c.formatter.PrintInfo("Versions for file: " + fileID)
	c.formatter.PrintTable([]string{"Version", "ID", "Size", "Hash", "Created", "Created By", "Description", "Current"},
		func() [][]string {
			var rows [][]string
			for _, version := range versions {
				current := "No"
				if version.IsCurrent {
					current = "Yes"
				}
				rows = append(rows, []string{
					strconv.Itoa(version.Version),
					version.ID[:8] + "...",
					c.formatter.FormatBytes(version.Size),
					version.Hash[:16] + "...",
					version.CreatedAt.Format("2006-01-02 15:04:05"),
					version.CreatedBy,
					version.Description,
					current,
				})
			}
			return rows
		}())

	return nil
}

func (c *VersionCommand) getCurrentVersion(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: version current <file_id>")
	}

	fileID := args[0]
	version, err := c.versionManager.GetCurrentVersion(fileID)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	c.formatter.PrintInfo("Current version for file: " + fileID)
	c.formatter.PrintInfo("Version: " + strconv.Itoa(version.Version))
	c.formatter.PrintInfo("ID: " + version.ID)
	c.formatter.PrintInfo("Size: " + c.formatter.FormatBytes(version.Size))
	c.formatter.PrintInfo("Hash: " + version.Hash)
	c.formatter.PrintInfo("Created: " + version.CreatedAt.Format(time.RFC3339))
	c.formatter.PrintInfo("Created by: " + version.CreatedBy)
	c.formatter.PrintInfo("Description: " + version.Description)

	return nil
}

func (c *VersionCommand) restoreVersion(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: version restore <file_id> <version_number>")
	}

	fileID := args[0]
	versionNumber, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid version number: %w", err)
	}

	err = c.versionManager.RestoreVersion(fileID, versionNumber)
	if err != nil {
		return fmt.Errorf("failed to restore version: %w", err)
	}

	c.formatter.PrintSuccess("Version restored successfully")
	c.formatter.PrintInfo("File: " + fileID)
	c.formatter.PrintInfo("Version: " + strconv.Itoa(versionNumber))

	return nil
}

func (c *VersionCommand) deleteVersion(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: version delete <file_id> <version_number>")
	}

	fileID := args[0]
	versionNumber, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid version number: %w", err)
	}

	err = c.versionManager.DeleteVersion(fileID, versionNumber)
	if err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	c.formatter.PrintSuccess("Version deleted successfully")
	c.formatter.PrintInfo("File: " + fileID)
	c.formatter.PrintInfo("Version: " + strconv.Itoa(versionNumber))

	return nil
}

func (c *VersionCommand) compareVersions(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: version compare <file_id> <version1> <version2>")
	}

	fileID := args[0]
	version1, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid version1 number: %w", err)
	}
	version2, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid version2 number: %w", err)
	}

	comparison, err := c.versionManager.CompareVersions(fileID, version1, version2)
	if err != nil {
		return fmt.Errorf("failed to compare versions: %w", err)
	}

	c.formatter.PrintInfo("Version comparison for file: " + fileID)
	c.formatter.PrintInfo("Version " + strconv.Itoa(version1) + " vs Version " + strconv.Itoa(version2))
	c.formatter.PrintInfo("Size difference: " + c.formatter.FormatBytes(comparison.SizeDiff))
	c.formatter.PrintInfo("Hash different: " + strconv.FormatBool(comparison.HashDiff))
	c.formatter.PrintInfo("Time difference: " + comparison.TimeDiff.String())
	if len(comparison.TagDiff) > 0 {
		c.formatter.PrintInfo("Tag differences: " + strings.Join(comparison.TagDiff, ", "))
	}

	return nil
}

func (c *VersionCommand) listAllVersions() error {
	versionInfos := c.versionManager.ListAllVersions()

	if len(versionInfos) == 0 {
		c.formatter.PrintInfo("No files with versions found")
		return nil
	}

	c.formatter.PrintInfo("All file versions:")
	c.formatter.PrintTable([]string{"File ID", "Current Version", "Total Versions", "Created", "Updated"},
		func() [][]string {
			var rows [][]string
			for _, info := range versionInfos {
				rows = append(rows, []string{
					info.FileID[:8] + "...",
					strconv.Itoa(info.CurrentVersion),
					strconv.Itoa(info.TotalVersions),
					info.CreatedAt.Format("2006-01-02 15:04:05"),
					info.UpdatedAt.Format("2006-01-02 15:04:05"),
				})
			}
			return rows
		}())

	return nil
}

// ShareCommand handles file sharing operations
type ShareCommand struct {
	BaseCommand
	shareManager *files.ShareManager
}

// NewShareCommand creates a new share command
func NewShareCommand(client *client.Client, formatter *formatter.Formatter, shareManager *files.ShareManager) *ShareCommand {
	return &ShareCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		shareManager: shareManager,
	}
}

// Execute executes the share command
func (c *ShareCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: share <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return c.createShare(subArgs)
	case "public":
		return c.createPublicShare(subArgs)
	case "get":
		return c.getShare(subArgs)
	case "list":
		return c.listShares(subArgs)
	case "update":
		return c.updateShare(subArgs)
	case "revoke":
		return c.revokeShare(subArgs)
	case "delete":
		return c.deleteShare(subArgs)
	case "stats":
		return c.getShareStats()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *ShareCommand) createShare(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: share create <file_id> <shared_by> <shared_with> <permissions> [expires_in]")
	}

	fileID := args[0]
	sharedBy := args[1]
	sharedWith := strings.Split(args[2], ",")
	permissions := strings.Split(args[3], ",")

	expiresIn := 24 * time.Hour // Default 24 hours
	if len(args) > 4 {
		duration, err := time.ParseDuration(args[4])
		if err != nil {
			return fmt.Errorf("invalid expires_in duration: %w", err)
		}
		expiresIn = duration
	}

	share, err := c.shareManager.ShareFile(fileID, sharedBy, sharedWith, permissions, expiresIn)
	if err != nil {
		return fmt.Errorf("failed to create share: %w", err)
	}

	c.formatter.PrintSuccess("File shared successfully")
	c.formatter.PrintInfo("Share ID: " + share.ID)
	c.formatter.PrintInfo("File ID: " + share.FileID)
	c.formatter.PrintInfo("Shared by: " + share.SharedBy)
	c.formatter.PrintInfo("Shared with: " + strings.Join(share.SharedWith, ", "))
	c.formatter.PrintInfo("Permissions: " + strings.Join(share.Permissions, ", "))
	c.formatter.PrintInfo("Expires: " + share.ExpiresAt.Format(time.RFC3339))

	return nil
}

func (c *ShareCommand) createPublicShare(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: share public <file_id> <shared_by> <permissions> [expires_in]")
	}

	fileID := args[0]
	sharedBy := args[1]
	permissions := strings.Split(args[2], ",")

	expiresIn := 24 * time.Hour // Default 24 hours
	if len(args) > 3 {
		duration, err := time.ParseDuration(args[3])
		if err != nil {
			return fmt.Errorf("invalid expires_in duration: %w", err)
		}
		expiresIn = duration
	}

	share, err := c.shareManager.ShareFilePublicly(fileID, sharedBy, permissions, expiresIn)
	if err != nil {
		return fmt.Errorf("failed to create public share: %w", err)
	}

	c.formatter.PrintSuccess("File shared publicly")
	c.formatter.PrintInfo("Share ID: " + share.ID)
	c.formatter.PrintInfo("File ID: " + share.FileID)
	c.formatter.PrintInfo("Shared by: " + share.SharedBy)
	c.formatter.PrintInfo("Permissions: " + strings.Join(share.Permissions, ", "))
	c.formatter.PrintInfo("Public URL: " + share.PublicURL)
	c.formatter.PrintInfo("Expires: " + share.ExpiresAt.Format(time.RFC3339))

	return nil
}

func (c *ShareCommand) getShare(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: share get <share_id>")
	}

	shareID := args[0]
	share, err := c.shareManager.GetShare(shareID)
	if err != nil {
		return fmt.Errorf("failed to get share: %w", err)
	}

	c.formatter.PrintInfo("Share details:")
	c.formatter.PrintInfo("ID: " + share.ID)
	c.formatter.PrintInfo("File ID: " + share.FileID)
	c.formatter.PrintInfo("Shared by: " + share.SharedBy)
	c.formatter.PrintInfo("Shared with: " + strings.Join(share.SharedWith, ", "))
	c.formatter.PrintInfo("Permissions: " + strings.Join(share.Permissions, ", "))
	c.formatter.PrintInfo("Is public: " + strconv.FormatBool(share.IsPublic))
	if share.IsPublic {
		c.formatter.PrintInfo("Public URL: " + share.PublicURL)
	}
	c.formatter.PrintInfo("Access count: " + strconv.Itoa(share.AccessCount))
	c.formatter.PrintInfo("Created: " + share.CreatedAt.Format(time.RFC3339))
	c.formatter.PrintInfo("Expires: " + share.ExpiresAt.Format(time.RFC3339))

	return nil
}

func (c *ShareCommand) listShares(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: share list <file_id>")
	}

	fileID := args[0]
	shares, err := c.shareManager.GetSharesForFile(fileID)
	if err != nil {
		return fmt.Errorf("failed to get shares: %w", err)
	}

	if len(shares) == 0 {
		c.formatter.PrintInfo("No shares found for file: " + fileID)
		return nil
	}

	c.formatter.PrintInfo("Shares for file: " + fileID)
	c.formatter.PrintTable([]string{"Share ID", "Shared By", "Shared With", "Permissions", "Public", "Access Count", "Expires"},
		func() [][]string {
			var rows [][]string
			for _, share := range shares {
				sharedWith := strings.Join(share.SharedWith, ", ")
				if share.IsPublic {
					sharedWith = "Public"
				}
				rows = append(rows, []string{
					share.ID[:8] + "...",
					share.SharedBy,
					sharedWith,
					strings.Join(share.Permissions, ", "),
					strconv.FormatBool(share.IsPublic),
					strconv.Itoa(share.AccessCount),
					share.ExpiresAt.Format("2006-01-02 15:04:05"),
				})
			}
			return rows
		}())

	return nil
}

func (c *ShareCommand) updateShare(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: share update <share_id> <permissions>")
	}

	shareID := args[0]
	permissions := strings.Split(args[1], ",")

	err := c.shareManager.UpdateSharePermissions(shareID, permissions)
	if err != nil {
		return fmt.Errorf("failed to update share: %w", err)
	}

	c.formatter.PrintSuccess("Share updated successfully")
	c.formatter.PrintInfo("Share ID: " + shareID)
	c.formatter.PrintInfo("New permissions: " + strings.Join(permissions, ", "))

	return nil
}

func (c *ShareCommand) revokeShare(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: share revoke <share_id>")
	}

	shareID := args[0]
	err := c.shareManager.RevokeShare(shareID)
	if err != nil {
		return fmt.Errorf("failed to revoke share: %w", err)
	}

	c.formatter.PrintSuccess("Share revoked successfully")
	c.formatter.PrintInfo("Share ID: " + shareID)

	return nil
}

func (c *ShareCommand) deleteShare(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: share delete <share_id>")
	}

	shareID := args[0]
	err := c.shareManager.DeleteShare(shareID)
	if err != nil {
		return fmt.Errorf("failed to delete share: %w", err)
	}

	c.formatter.PrintSuccess("Share deleted successfully")
	c.formatter.PrintInfo("Share ID: " + shareID)

	return nil
}

func (c *ShareCommand) getShareStats() error {
	stats := c.shareManager.GetShareStats()

	c.formatter.PrintInfo("Sharing statistics:")
	c.formatter.PrintInfo("Total shares: " + strconv.Itoa(stats.TotalShares))
	c.formatter.PrintInfo("Public shares: " + strconv.Itoa(stats.PublicShares))
	c.formatter.PrintInfo("Private shares: " + strconv.Itoa(stats.PrivateShares))
	c.formatter.PrintInfo("Active shares: " + strconv.Itoa(stats.ActiveShares))
	c.formatter.PrintInfo("Expired shares: " + strconv.Itoa(stats.ExpiredShares))
	c.formatter.PrintInfo("Total accesses: " + strconv.Itoa(stats.TotalAccesses))

	return nil
}
