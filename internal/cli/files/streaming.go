package files

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// StreamingManager manages file streaming
type StreamingManager struct {
	client    *client.Client
	configDir string
	streams   map[string]*StreamInfo
	settings  *StreamingSettings
	mu        sync.RWMutex
}

// StreamInfo represents information about an active stream
type StreamInfo struct {
	ID               string                 `json:"id"`
	FileID           string                 `json:"file_id"`
	UserID           string                 `json:"user_id"`
	StreamType       string                 `json:"stream_type"` // "upload", "download"
	Status           string                 `json:"status"`      // "active", "paused", "completed", "error"
	Progress         float64                `json:"progress"`    // 0.0 to 1.0
	BytesTransferred int64                  `json:"bytes_transferred"`
	TotalBytes       int64                  `json:"total_bytes"`
	StartTime        time.Time              `json:"start_time"`
	LastUpdate       time.Time              `json:"last_update"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	Error            string                 `json:"error,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// StreamingSettings represents streaming configuration
type StreamingSettings struct {
	ChunkSize         int64         `json:"chunk_size"`         // Size of each chunk
	BufferSize        int64         `json:"buffer_size"`        // Buffer size for streaming
	MaxConcurrent     int           `json:"max_concurrent"`     // Max concurrent streams
	Timeout           time.Duration `json:"timeout"`            // Stream timeout
	RetryAttempts     int           `json:"retry_attempts"`     // Number of retry attempts
	RetryDelay        time.Duration `json:"retry_delay"`        // Delay between retries
	EnableCompression bool          `json:"enable_compression"` // Enable streaming compression
	EnableEncryption  bool          `json:"enable_encryption"`  // Enable streaming encryption
}

// StreamingStats represents streaming statistics
type StreamingStats struct {
	TotalStreams       int64     `json:"total_streams"`
	ActiveStreams      int64     `json:"active_streams"`
	CompletedStreams   int64     `json:"completed_streams"`
	FailedStreams      int64     `json:"failed_streams"`
	TotalBytesStreamed int64     `json:"total_bytes_streamed"`
	AverageSpeed       float64   `json:"average_speed"` // bytes per second
	LastUpdated        time.Time `json:"last_updated"`
}

// StreamProgress represents progress information for a stream
type StreamProgress struct {
	StreamID         string        `json:"stream_id"`
	Progress         float64       `json:"progress"`
	BytesTransferred int64         `json:"bytes_transferred"`
	TotalBytes       int64         `json:"total_bytes"`
	Speed            float64       `json:"speed"` // bytes per second
	ETA              time.Duration `json:"eta"`   // estimated time remaining
	Status           string        `json:"status"`
}

// NewStreamingManager creates a new streaming manager
func NewStreamingManager(client *client.Client, configDir string) *StreamingManager {
	sm := &StreamingManager{
		client:    client,
		configDir: configDir,
		streams:   make(map[string]*StreamInfo),
		settings:  getDefaultStreamingSettings(),
	}

	sm.loadSettings()
	sm.loadStreams()
	return sm
}

// StartUploadStream starts streaming upload of a file
func (sm *StreamingManager) StartUploadStream(fileID, userID string, reader io.Reader, totalSize int64) (*StreamInfo, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check concurrent stream limit
	if len(sm.streams) >= sm.settings.MaxConcurrent {
		return nil, fmt.Errorf("maximum concurrent streams reached")
	}

	streamID := sm.generateStreamID()
	stream := &StreamInfo{
		ID:               streamID,
		FileID:           fileID,
		UserID:           userID,
		StreamType:       "upload",
		Status:           "active",
		Progress:         0.0,
		BytesTransferred: 0,
		TotalBytes:       totalSize,
		StartTime:        time.Now(),
		LastUpdate:       time.Now(),
		Metadata:         make(map[string]interface{}),
	}

	sm.streams[streamID] = stream
	sm.saveStreams()

	// Start streaming in goroutine
	go sm.uploadStream(stream, reader)

	return stream, nil
}

// StartDownloadStream starts streaming download of a file
func (sm *StreamingManager) StartDownloadStream(fileID, userID string, writer io.Writer) (*StreamInfo, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check concurrent stream limit
	if len(sm.streams) >= sm.settings.MaxConcurrent {
		return nil, fmt.Errorf("maximum concurrent streams reached")
	}

	// Get file info to determine total size
	fileInfo, err := sm.client.GetFile(context.Background(), fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	streamID := sm.generateStreamID()
	stream := &StreamInfo{
		ID:               streamID,
		FileID:           fileID,
		UserID:           userID,
		StreamType:       "download",
		Status:           "active",
		Progress:         0.0,
		BytesTransferred: 0,
		TotalBytes:       fileInfo.Size,
		StartTime:        time.Now(),
		LastUpdate:       time.Now(),
		Metadata:         make(map[string]interface{}),
	}

	sm.streams[streamID] = stream
	sm.saveStreams()

	// Start streaming in goroutine
	go sm.downloadStream(stream, writer)

	return stream, nil
}

// GetStream returns information about a stream
func (sm *StreamingManager) GetStream(streamID string) (*StreamInfo, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stream, exists := sm.streams[streamID]
	if !exists {
		return nil, fmt.Errorf("stream not found: %s", streamID)
	}

	// Return a copy
	info := *stream
	return &info, nil
}

// GetStreamProgress returns progress information for a stream
func (sm *StreamingManager) GetStreamProgress(streamID string) (*StreamProgress, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stream, exists := sm.streams[streamID]
	if !exists {
		return nil, fmt.Errorf("stream not found: %s", streamID)
	}

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

	progress := &StreamProgress{
		StreamID:         streamID,
		Progress:         stream.Progress,
		BytesTransferred: stream.BytesTransferred,
		TotalBytes:       stream.TotalBytes,
		Speed:            speed,
		ETA:              eta,
		Status:           stream.Status,
	}

	return progress, nil
}

// PauseStream pauses a stream
func (sm *StreamingManager) PauseStream(streamID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stream, exists := sm.streams[streamID]
	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	if stream.Status != "active" {
		return fmt.Errorf("stream is not active: %s", stream.Status)
	}

	stream.Status = "paused"
	stream.LastUpdate = time.Now()
	sm.saveStreams()

	return nil
}

// ResumeStream resumes a paused stream
func (sm *StreamingManager) ResumeStream(streamID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stream, exists := sm.streams[streamID]
	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	if stream.Status != "paused" {
		return fmt.Errorf("stream is not paused: %s", stream.Status)
	}

	stream.Status = "active"
	stream.LastUpdate = time.Now()
	sm.saveStreams()

	return nil
}

// CancelStream cancels a stream
func (sm *StreamingManager) CancelStream(streamID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stream, exists := sm.streams[streamID]
	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	stream.Status = "cancelled"
	stream.LastUpdate = time.Now()
	now := time.Now()
	stream.EndTime = &now
	sm.saveStreams()

	return nil
}

// ListStreams returns all active streams
func (sm *StreamingManager) ListStreams() []*StreamInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var streams []*StreamInfo
	for _, stream := range sm.streams {
		if stream.Status == "active" || stream.Status == "paused" {
			// Return a copy
			info := *stream
			streams = append(streams, &info)
		}
	}

	return streams
}

// GetStreamingSettings returns current streaming settings
func (sm *StreamingManager) GetStreamingSettings() *StreamingSettings {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy
	settings := *sm.settings
	return &settings
}

// UpdateStreamingSettings updates streaming settings
func (sm *StreamingManager) UpdateStreamingSettings(settings *StreamingSettings) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate settings
	if settings.ChunkSize <= 0 {
		return fmt.Errorf("chunk size must be positive")
	}

	if settings.BufferSize <= 0 {
		return fmt.Errorf("buffer size must be positive")
	}

	if settings.MaxConcurrent <= 0 {
		return fmt.Errorf("max concurrent streams must be positive")
	}

	sm.settings = settings
	sm.saveSettings()

	return nil
}

// Streaming implementation
func (sm *StreamingManager) uploadStream(stream *StreamInfo, reader io.Reader) {
	buffer := make([]byte, sm.settings.ChunkSize)

	for {
		// Check if stream is paused or cancelled
		sm.mu.RLock()
		status := stream.Status
		sm.mu.RUnlock()

		if status == "paused" {
			time.Sleep(time.Second)
			continue
		}

		if status == "cancelled" {
			break
		}

		// Read chunk
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// Upload complete
				sm.mu.Lock()
				stream.Status = "completed"
				stream.Progress = 1.0
				stream.BytesTransferred = stream.TotalBytes
				stream.LastUpdate = time.Now()
				now := time.Now()
				stream.EndTime = &now
				sm.mu.Unlock()
				break
			} else {
				// Upload error
				sm.mu.Lock()
				stream.Status = "error"
				stream.Error = err.Error()
				stream.LastUpdate = time.Now()
				now := time.Now()
				stream.EndTime = &now
				sm.mu.Unlock()
				break
			}
		}

		// Update progress
		sm.mu.Lock()
		stream.BytesTransferred += int64(n)
		if stream.TotalBytes > 0 {
			stream.Progress = float64(stream.BytesTransferred) / float64(stream.TotalBytes)
		}
		stream.LastUpdate = time.Now()
		sm.mu.Unlock()

		// Simulate upload delay
		time.Sleep(time.Millisecond * 10)
	}

	sm.saveStreams()
}

func (sm *StreamingManager) downloadStream(stream *StreamInfo, writer io.Writer) {
	// Simulate download streaming
	buffer := make([]byte, sm.settings.ChunkSize)
	chunkSize := int64(len(buffer))

	for stream.BytesTransferred < stream.TotalBytes {
		// Check if stream is paused or cancelled
		sm.mu.RLock()
		status := stream.Status
		sm.mu.RUnlock()

		if status == "paused" {
			time.Sleep(time.Second)
			continue
		}

		if status == "cancelled" {
			break
		}

		// Calculate chunk size for this iteration
		remaining := stream.TotalBytes - stream.BytesTransferred
		if remaining < chunkSize {
			chunkSize = remaining
		}

		// Simulate writing chunk
		_, err := writer.Write(buffer[:chunkSize])
		if err != nil {
			sm.mu.Lock()
			stream.Status = "error"
			stream.Error = err.Error()
			stream.LastUpdate = time.Now()
			now := time.Now()
			stream.EndTime = &now
			sm.mu.Unlock()
			break
		}

		// Update progress
		sm.mu.Lock()
		stream.BytesTransferred += chunkSize
		stream.Progress = float64(stream.BytesTransferred) / float64(stream.TotalBytes)
		stream.LastUpdate = time.Now()
		sm.mu.Unlock()

		// Simulate download delay
		time.Sleep(time.Millisecond * 10)
	}

	// Mark as completed if not already marked
	sm.mu.Lock()
	if stream.Status == "active" {
		stream.Status = "completed"
		stream.Progress = 1.0
		stream.LastUpdate = time.Now()
		now := time.Now()
		stream.EndTime = &now
	}
	sm.mu.Unlock()

	sm.saveStreams()
}

// Utility functions
func (sm *StreamingManager) generateStreamID() string {
	return fmt.Sprintf("stream_%d", time.Now().UnixNano())
}

func getDefaultStreamingSettings() *StreamingSettings {
	return &StreamingSettings{
		ChunkSize:         64 * 1024,   // 64KB
		BufferSize:        1024 * 1024, // 1MB
		MaxConcurrent:     10,
		Timeout:           30 * time.Minute,
		RetryAttempts:     3,
		RetryDelay:        5 * time.Second,
		EnableCompression: true,
		EnableEncryption:  false,
	}
}

// Data persistence
func (sm *StreamingManager) loadSettings() error {
	settingsFile := filepath.Join(sm.configDir, "streaming.json")
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		return nil // Use default settings
	}

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings StreamingSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	sm.settings = &settings
	return nil
}

func (sm *StreamingManager) saveSettings() error {
	settingsFile := filepath.Join(sm.configDir, "streaming.json")

	data, err := json.MarshalIndent(sm.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	return os.WriteFile(settingsFile, data, 0644)
}

func (sm *StreamingManager) loadStreams() error {
	streamsFile := filepath.Join(sm.configDir, "streams.json")
	if _, err := os.Stat(streamsFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty streams
	}

	data, err := os.ReadFile(streamsFile)
	if err != nil {
		return fmt.Errorf("failed to read streams file: %w", err)
	}

	var streams map[string]*StreamInfo
	if err := json.Unmarshal(data, &streams); err != nil {
		return fmt.Errorf("failed to unmarshal streams: %w", err)
	}

	sm.streams = streams
	return nil
}

func (sm *StreamingManager) saveStreams() error {
	streamsFile := filepath.Join(sm.configDir, "streams.json")

	data, err := json.MarshalIndent(sm.streams, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal streams: %w", err)
	}

	return os.WriteFile(streamsFile, data, 0644)
}
