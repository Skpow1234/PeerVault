package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/files"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// CompressionCommand handles compression operations
type CompressionCommand struct {
	BaseCommand
	compressionManager *files.CompressionManager
}

// NewCompressionCommand creates a new compression command
func NewCompressionCommand(client *client.Client, formatter *formatter.Formatter, compressionManager *files.CompressionManager) *CompressionCommand {
	return &CompressionCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		compressionManager: compressionManager,
	}
}

// Execute executes the compression command
func (c *CompressionCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: compress <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "file":
		return c.compressFile(subArgs)
	case "decompress":
		return c.decompressFile(subArgs)
	case "settings":
		return c.getSettings(subArgs)
	case "update-settings":
		return c.updateSettings(subArgs)
	case "stats":
		return c.getStats()
	case "reset-stats":
		return c.resetStats()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *CompressionCommand) compressFile(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: compress file <file_id> <algorithm> <level>")
	}

	fileID := args[0]
	algorithm := args[1]
	level, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid compression level: %w", err)
	}

	if level < 1 || level > 9 {
		return fmt.Errorf("compression level must be between 1 and 9")
	}

	result, err := c.compressionManager.CompressFile(fileID, algorithm, level)
	if err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	if result.Success {
		c.formatter.PrintSuccess("File compressed successfully")
		c.formatter.PrintInfo("Original size: " + c.formatter.FormatBytes(result.OriginalSize))
		c.formatter.PrintInfo("Compressed size: " + c.formatter.FormatBytes(result.CompressedSize))
		c.formatter.PrintInfo("Compression ratio: " + fmt.Sprintf("%.2f%%", result.CompressionRatio*100))
		c.formatter.PrintInfo("Algorithm: " + result.Algorithm)
		c.formatter.PrintInfo("Level: " + strconv.Itoa(result.Level))
		c.formatter.PrintInfo("Time taken: " + result.TimeTaken.String())
	} else {
		c.formatter.PrintError(fmt.Errorf("compression failed: %s", result.Error))
	}

	return nil
}

func (c *CompressionCommand) decompressFile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: compress decompress <file_id> <algorithm>")
	}

	fileID := args[0]
	algorithm := args[1]

	data, err := c.compressionManager.DecompressFile(fileID, algorithm)
	if err != nil {
		return fmt.Errorf("failed to decompress file: %w", err)
	}

	c.formatter.PrintSuccess("File decompressed successfully")
	c.formatter.PrintInfo("Decompressed size: " + c.formatter.FormatBytes(int64(len(data))))
	c.formatter.PrintInfo("Algorithm: " + algorithm)

	return nil
}

func (c *CompressionCommand) getSettings(args []string) error {
	settings := c.compressionManager.GetCompressionSettings()

	c.formatter.PrintInfo("Compression settings:")
	c.formatter.PrintInfo("Enabled: " + strconv.FormatBool(settings.Enabled))
	c.formatter.PrintInfo("Algorithm: " + settings.Algorithm)
	c.formatter.PrintInfo("Level: " + strconv.Itoa(settings.Level))
	c.formatter.PrintInfo("Min size: " + c.formatter.FormatBytes(settings.MinSize))
	c.formatter.PrintInfo("Max size: " + c.formatter.FormatBytes(settings.MaxSize))
	c.formatter.PrintInfo("Auto compress: " + strconv.FormatBool(settings.AutoCompress))
	c.formatter.PrintInfo("Compress on upload: " + strconv.FormatBool(settings.CompressOnUpload))

	return nil
}

func (c *CompressionCommand) updateSettings(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: compress update-settings <setting> <value>")
	}

	setting := args[0]
	value := args[1]

	settings := c.compressionManager.GetCompressionSettings()

	switch setting {
	case "enabled":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %w", err)
		}
		settings.Enabled = enabled
	case "algorithm":
		if value != "gzip" && value != "zlib" && value != "none" {
			return fmt.Errorf("invalid algorithm: %s (must be gzip, zlib, or none)", value)
		}
		settings.Algorithm = value
	case "level":
		level, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid level: %w", err)
		}
		if level < 1 || level > 9 {
			return fmt.Errorf("level must be between 1 and 9")
		}
		settings.Level = level
	case "min_size":
		size, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid size: %w", err)
		}
		settings.MinSize = size
	case "max_size":
		size, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid size: %w", err)
		}
		settings.MaxSize = size
	case "auto_compress":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %w", err)
		}
		settings.AutoCompress = enabled
	case "compress_on_upload":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %w", err)
		}
		settings.CompressOnUpload = enabled
	default:
		return fmt.Errorf("unknown setting: %s", setting)
	}

	err := c.compressionManager.UpdateCompressionSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	c.formatter.PrintSuccess("Settings updated successfully")
	c.formatter.PrintInfo(setting + " = " + value)

	return nil
}

func (c *CompressionCommand) getStats() error {
	stats := c.compressionManager.GetCompressionStats()

	c.formatter.PrintInfo("Compression statistics:")
	c.formatter.PrintInfo("Total files: " + strconv.FormatInt(stats.TotalFiles, 10))
	c.formatter.PrintInfo("Compressed files: " + strconv.FormatInt(stats.CompressedFiles, 10))
	c.formatter.PrintInfo("Uncompressed files: " + strconv.FormatInt(stats.UncompressedFiles, 10))
	c.formatter.PrintInfo("Total original size: " + c.formatter.FormatBytes(stats.TotalOriginalSize))
	c.formatter.PrintInfo("Total compressed size: " + c.formatter.FormatBytes(stats.TotalCompressedSize))
	c.formatter.PrintInfo("Compression ratio: " + fmt.Sprintf("%.2f%%", stats.CompressionRatio*100))
	c.formatter.PrintInfo("Space saved: " + c.formatter.FormatBytes(stats.SpaceSaved))
	c.formatter.PrintInfo("Last updated: " + stats.LastUpdated.Format("2006-01-02 15:04:05"))

	return nil
}

func (c *CompressionCommand) resetStats() error {
	c.compressionManager.ResetStats()
	c.formatter.PrintSuccess("Compression statistics reset successfully")
	return nil
}

// DeduplicationCommand handles deduplication operations
type DeduplicationCommand struct {
	BaseCommand
	deduplicationManager *files.DeduplicationManager
}

// NewDeduplicationCommand creates a new deduplication command
func NewDeduplicationCommand(client *client.Client, formatter *formatter.Formatter, deduplicationManager *files.DeduplicationManager) *DeduplicationCommand {
	return &DeduplicationCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		deduplicationManager: deduplicationManager,
	}
}

// Execute executes the deduplication command
func (c *DeduplicationCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dedup <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "file":
		return c.deduplicateFile(subArgs)
	case "reconstruct":
		return c.reconstructFile(subArgs)
	case "chunk":
		return c.getChunkInfo(subArgs)
	case "list-chunks":
		return c.listChunks()
	case "remove-file":
		return c.removeFileChunks(subArgs)
	case "cleanup":
		return c.cleanupUnusedChunks()
	case "stats":
		return c.getStats()
	case "reset-stats":
		return c.resetStats()
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *DeduplicationCommand) deduplicateFile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: dedup file <file_id> <chunk_size>")
	}

	fileID := args[0]
	chunkSizeStr := args[1]

	var chunkSize files.ChunkSize
	switch chunkSizeStr {
	case "small":
		chunkSize = files.ChunkSizeSmall
	case "medium":
		chunkSize = files.ChunkSizeMedium
	case "large":
		chunkSize = files.ChunkSizeLarge
	default:
		size, err := strconv.ParseInt(chunkSizeStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chunk size: %w", err)
		}
		chunkSize = files.ChunkSize(size)
	}

	result, err := c.deduplicationManager.DeduplicateFile(fileID, chunkSize)
	if err != nil {
		return fmt.Errorf("failed to deduplicate file: %w", err)
	}

	if result.Success {
		c.formatter.PrintSuccess("File deduplicated successfully")
		c.formatter.PrintInfo("File ID: " + result.FileID)
		c.formatter.PrintInfo("Original size: " + c.formatter.FormatBytes(result.OriginalSize))
		c.formatter.PrintInfo("Deduplicated size: " + c.formatter.FormatBytes(result.DeduplicatedSize))
		c.formatter.PrintInfo("Chunks created: " + strconv.Itoa(result.ChunksCreated))
		c.formatter.PrintInfo("Chunks reused: " + strconv.Itoa(result.ChunksReused))
		c.formatter.PrintInfo("Space saved: " + c.formatter.FormatBytes(result.SpaceSaved))
		c.formatter.PrintInfo("Deduplication ratio: " + fmt.Sprintf("%.2f%%", result.DeduplicationRatio*100))
		c.formatter.PrintInfo("Time taken: " + result.TimeTaken.String())
	} else {
		c.formatter.PrintError(fmt.Errorf("deduplication failed: %s", result.Error))
	}

	return nil
}

func (c *DeduplicationCommand) reconstructFile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dedup reconstruct <file_id>")
	}

	fileID := args[0]
	data, err := c.deduplicationManager.ReconstructFile(fileID)
	if err != nil {
		return fmt.Errorf("failed to reconstruct file: %w", err)
	}

	c.formatter.PrintSuccess("File reconstructed successfully")
	c.formatter.PrintInfo("File ID: " + fileID)
	c.formatter.PrintInfo("Reconstructed size: " + c.formatter.FormatBytes(int64(len(data))))

	return nil
}

func (c *DeduplicationCommand) getChunkInfo(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dedup chunk <chunk_hash>")
	}

	chunkHash := args[0]
	chunkInfo, err := c.deduplicationManager.GetChunkInfo(chunkHash)
	if err != nil {
		return fmt.Errorf("failed to get chunk info: %w", err)
	}

	c.formatter.PrintInfo("Chunk information:")
	c.formatter.PrintInfo("Hash: " + chunkInfo.Hash)
	c.formatter.PrintInfo("Size: " + c.formatter.FormatBytes(chunkInfo.Size))
	c.formatter.PrintInfo("Reference count: " + strconv.Itoa(chunkInfo.RefCount))
	c.formatter.PrintInfo("Created: " + chunkInfo.CreatedAt.Format("2006-01-02 15:04:05"))
	c.formatter.PrintInfo("Last accessed: " + chunkInfo.LastAccessed.Format("2006-01-02 15:04:05"))
	c.formatter.PrintInfo("File IDs: " + strings.Join(chunkInfo.FileIDs, ", "))

	return nil
}

func (c *DeduplicationCommand) listChunks() error {
	chunks := c.deduplicationManager.ListChunks()

	if len(chunks) == 0 {
		c.formatter.PrintInfo("No chunks found")
		return nil
	}

	c.formatter.PrintInfo("All chunks:")
	c.formatter.PrintTable([]string{"Hash", "Size", "Ref Count", "Created", "Last Accessed", "File Count"},
		func() [][]string {
			var rows [][]string
			for _, chunk := range chunks {
				rows = append(rows, []string{
					chunk.Hash[:16] + "...",
					c.formatter.FormatBytes(chunk.Size),
					strconv.Itoa(chunk.RefCount),
					chunk.CreatedAt.Format("2006-01-02 15:04:05"),
					chunk.LastAccessed.Format("2006-01-02 15:04:05"),
					strconv.Itoa(len(chunk.FileIDs)),
				})
			}
			return rows
		}())

	return nil
}

func (c *DeduplicationCommand) removeFileChunks(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dedup remove-file <file_id>")
	}

	fileID := args[0]
	err := c.deduplicationManager.RemoveFileChunks(fileID)
	if err != nil {
		return fmt.Errorf("failed to remove file chunks: %w", err)
	}

	c.formatter.PrintSuccess("File chunks removed successfully")
	c.formatter.PrintInfo("File ID: " + fileID)

	return nil
}

func (c *DeduplicationCommand) cleanupUnusedChunks() error {
	err := c.deduplicationManager.CleanupUnusedChunks()
	if err != nil {
		return fmt.Errorf("failed to cleanup unused chunks: %w", err)
	}

	c.formatter.PrintSuccess("Unused chunks cleaned up successfully")
	return nil
}

func (c *DeduplicationCommand) getStats() error {
	stats := c.deduplicationManager.GetDeduplicationStats()

	c.formatter.PrintInfo("Deduplication statistics:")
	c.formatter.PrintInfo("Total files: " + strconv.FormatInt(stats.TotalFiles, 10))
	c.formatter.PrintInfo("Unique chunks: " + strconv.FormatInt(stats.UniqueChunks, 10))
	c.formatter.PrintInfo("Duplicate chunks: " + strconv.FormatInt(stats.DuplicateChunks, 10))
	c.formatter.PrintInfo("Total size: " + c.formatter.FormatBytes(stats.TotalSize))
	c.formatter.PrintInfo("Deduplicated size: " + c.formatter.FormatBytes(stats.DeduplicatedSize))
	c.formatter.PrintInfo("Space saved: " + c.formatter.FormatBytes(stats.SpaceSaved))
	c.formatter.PrintInfo("Deduplication ratio: " + fmt.Sprintf("%.2f%%", stats.DeduplicationRatio*100))
	c.formatter.PrintInfo("Last updated: " + stats.LastUpdated.Format("2006-01-02 15:04:05"))

	return nil
}

func (c *DeduplicationCommand) resetStats() error {
	c.deduplicationManager.ResetStats()
	c.formatter.PrintSuccess("Deduplication statistics reset successfully")
	return nil
}
