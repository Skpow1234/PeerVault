package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/files"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// StreamingCommand handles streaming operations
type StreamingCommand struct {
	BaseCommand
	streamingManager *files.StreamingManager
}

// NewStreamingCommand creates a new streaming command
func NewStreamingCommand(client *client.Client, formatter *formatter.Formatter, streamingManager *files.StreamingManager) *StreamingCommand {
	return &StreamingCommand{
		BaseCommand: BaseCommand{
			client:    client,
			formatter: formatter,
		},
		streamingManager: streamingManager,
	}
}

// Execute executes the streaming command
func (c *StreamingCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stream <subcommand> [args...]")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "upload":
		return c.startUploadStream(subArgs)
	case "download":
		return c.startDownloadStream(subArgs)
	case "get":
		return c.getStream(subArgs)
	case "progress":
		return c.getStreamProgress(subArgs)
	case "pause":
		return c.pauseStream(subArgs)
	case "resume":
		return c.resumeStream(subArgs)
	case "cancel":
		return c.cancelStream(subArgs)
	case "list":
		return c.listStreams()
	case "settings":
		return c.getSettings(subArgs)
	case "update-settings":
		return c.updateSettings(subArgs)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func (c *StreamingCommand) startUploadStream(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: stream upload <file_id> <user_id> <file_path>")
	}

	fileID := args[0]
	userID := args[1]
	filePath := args[2]

	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	stream, err := c.streamingManager.StartUploadStream(fileID, userID, file, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to start upload stream: %w", err)
	}

	c.formatter.PrintSuccess("Upload stream started")
	c.formatter.PrintInfo("Stream ID: " + stream.ID)
	c.formatter.PrintInfo("File ID: " + stream.FileID)
	c.formatter.PrintInfo("User ID: " + stream.UserID)
	c.formatter.PrintInfo("Total size: " + c.formatter.FormatBytes(stream.TotalBytes))
	c.formatter.PrintInfo("Status: " + stream.Status)

	return nil
}

func (c *StreamingCommand) startDownloadStream(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: stream download <file_id> <user_id> <output_path>")
	}

	fileID := args[0]
	userID := args[1]
	outputPath := args[2]

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	stream, err := c.streamingManager.StartDownloadStream(fileID, userID, file)
	if err != nil {
		return fmt.Errorf("failed to start download stream: %w", err)
	}

	c.formatter.PrintSuccess("Download stream started")
	c.formatter.PrintInfo("Stream ID: " + stream.ID)
	c.formatter.PrintInfo("File ID: " + stream.FileID)
	c.formatter.PrintInfo("User ID: " + stream.UserID)
	c.formatter.PrintInfo("Total size: " + c.formatter.FormatBytes(stream.TotalBytes))
	c.formatter.PrintInfo("Status: " + stream.Status)
	c.formatter.PrintInfo("Output path: " + outputPath)

	return nil
}

func (c *StreamingCommand) getStream(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stream get <stream_id>")
	}

	streamID := args[0]
	stream, err := c.streamingManager.GetStream(streamID)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	c.formatter.PrintInfo("Stream details:")
	c.formatter.PrintInfo("ID: " + stream.ID)
	c.formatter.PrintInfo("File ID: " + stream.FileID)
	c.formatter.PrintInfo("User ID: " + stream.UserID)
	c.formatter.PrintInfo("Type: " + stream.StreamType)
	c.formatter.PrintInfo("Status: " + stream.Status)
	c.formatter.PrintInfo("Progress: " + fmt.Sprintf("%.2f%%", stream.Progress*100))
	c.formatter.PrintInfo("Bytes transferred: " + c.formatter.FormatBytes(stream.BytesTransferred))
	c.formatter.PrintInfo("Total bytes: " + c.formatter.FormatBytes(stream.TotalBytes))
	c.formatter.PrintInfo("Start time: " + stream.StartTime.Format(time.RFC3339))
	c.formatter.PrintInfo("Last update: " + stream.LastUpdate.Format(time.RFC3339))
	if stream.EndTime != nil {
		c.formatter.PrintInfo("End time: " + stream.EndTime.Format(time.RFC3339))
	}
	if stream.Error != "" {
		c.formatter.PrintInfo("Error: " + stream.Error)
	}

	return nil
}

func (c *StreamingCommand) getStreamProgress(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stream progress <stream_id>")
	}

	streamID := args[0]
	progress, err := c.streamingManager.GetStreamProgress(streamID)
	if err != nil {
		return fmt.Errorf("failed to get stream progress: %w", err)
	}

	c.formatter.PrintInfo("Stream progress:")
	c.formatter.PrintInfo("Stream ID: " + progress.StreamID)
	c.formatter.PrintInfo("Progress: " + fmt.Sprintf("%.2f%%", progress.Progress*100))
	c.formatter.PrintInfo("Bytes transferred: " + c.formatter.FormatBytes(progress.BytesTransferred))
	c.formatter.PrintInfo("Total bytes: " + c.formatter.FormatBytes(progress.TotalBytes))
	c.formatter.PrintInfo("Speed: " + c.formatter.FormatBytes(int64(progress.Speed)) + "/s")
	c.formatter.PrintInfo("ETA: " + progress.ETA.String())
	c.formatter.PrintInfo("Status: " + progress.Status)

	return nil
}

func (c *StreamingCommand) pauseStream(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stream pause <stream_id>")
	}

	streamID := args[0]
	err := c.streamingManager.PauseStream(streamID)
	if err != nil {
		return fmt.Errorf("failed to pause stream: %w", err)
	}

	c.formatter.PrintSuccess("Stream paused successfully")
	c.formatter.PrintInfo("Stream ID: " + streamID)

	return nil
}

func (c *StreamingCommand) resumeStream(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stream resume <stream_id>")
	}

	streamID := args[0]
	err := c.streamingManager.ResumeStream(streamID)
	if err != nil {
		return fmt.Errorf("failed to resume stream: %w", err)
	}

	c.formatter.PrintSuccess("Stream resumed successfully")
	c.formatter.PrintInfo("Stream ID: " + streamID)

	return nil
}

func (c *StreamingCommand) cancelStream(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stream cancel <stream_id>")
	}

	streamID := args[0]
	err := c.streamingManager.CancelStream(streamID)
	if err != nil {
		return fmt.Errorf("failed to cancel stream: %w", err)
	}

	c.formatter.PrintSuccess("Stream cancelled successfully")
	c.formatter.PrintInfo("Stream ID: " + streamID)

	return nil
}

func (c *StreamingCommand) listStreams() error {
	streams := c.streamingManager.ListStreams()

	if len(streams) == 0 {
		c.formatter.PrintInfo("No active streams found")
		return nil
	}

	c.formatter.PrintInfo("Active streams:")
	c.formatter.PrintTable([]string{"Stream ID", "File ID", "User ID", "Type", "Status", "Progress", "Speed", "ETA"},
		func() [][]string {
			var rows [][]string
			for _, stream := range streams {
				// Calculate speed and ETA
				speed := float64(0)
				eta := time.Duration(0)

				if stream.TotalBytes > 0 && stream.BytesTransferred > 0 {
					elapsed := time.Since(stream.StartTime)
					if elapsed > 0 {
						speed = float64(stream.BytesTransferred) / elapsed.Seconds()
						remainingBytes := stream.TotalBytes - stream.BytesTransferred
						if speed > 0 {
							eta = time.Duration(float64(remainingBytes)/speed) * time.Second
						}
					}
				}

				rows = append(rows, []string{
					stream.ID[:8] + "...",
					stream.FileID[:8] + "...",
					stream.UserID,
					stream.StreamType,
					stream.Status,
					fmt.Sprintf("%.2f%%", stream.Progress*100),
					c.formatter.FormatBytes(int64(speed)) + "/s",
					eta.String(),
				})
			}
			return rows
		}())

	return nil
}

func (c *StreamingCommand) getSettings(args []string) error {
	settings := c.streamingManager.GetStreamingSettings()

	c.formatter.PrintInfo("Streaming settings:")
	c.formatter.PrintInfo("Chunk size: " + c.formatter.FormatBytes(settings.ChunkSize))
	c.formatter.PrintInfo("Buffer size: " + c.formatter.FormatBytes(settings.BufferSize))
	c.formatter.PrintInfo("Max concurrent: " + strconv.Itoa(settings.MaxConcurrent))
	c.formatter.PrintInfo("Timeout: " + settings.Timeout.String())
	c.formatter.PrintInfo("Retry attempts: " + strconv.Itoa(settings.RetryAttempts))
	c.formatter.PrintInfo("Retry delay: " + settings.RetryDelay.String())
	c.formatter.PrintInfo("Enable compression: " + strconv.FormatBool(settings.EnableCompression))
	c.formatter.PrintInfo("Enable encryption: " + strconv.FormatBool(settings.EnableEncryption))

	return nil
}

func (c *StreamingCommand) updateSettings(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: stream update-settings <setting> <value>")
	}

	setting := args[0]
	value := args[1]

	settings := c.streamingManager.GetStreamingSettings()

	switch setting {
	case "chunk_size":
		size, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid size: %w", err)
		}
		settings.ChunkSize = size
	case "buffer_size":
		size, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid size: %w", err)
		}
		settings.BufferSize = size
	case "max_concurrent":
		max, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max concurrent: %w", err)
		}
		settings.MaxConcurrent = max
	case "timeout":
		timeout, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		settings.Timeout = timeout
	case "retry_attempts":
		attempts, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid retry attempts: %w", err)
		}
		settings.RetryAttempts = attempts
	case "retry_delay":
		delay, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid retry delay: %w", err)
		}
		settings.RetryDelay = delay
	case "enable_compression":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %w", err)
		}
		settings.EnableCompression = enabled
	case "enable_encryption":
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %w", err)
		}
		settings.EnableEncryption = enabled
	default:
		return fmt.Errorf("unknown setting: %s", setting)
	}

	err := c.streamingManager.UpdateStreamingSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	c.formatter.PrintSuccess("Settings updated successfully")
	c.formatter.PrintInfo(setting + " = " + value)

	return nil
}
