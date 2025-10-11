package operations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
	"github.com/Skpow1234/Peervault/internal/cli/formatter"
)

// Manager manages advanced file operations
type Manager struct {
	client    *client.Client
	formatter *formatter.Formatter
}

// New creates a new operations manager
func New(client *client.Client, formatter *formatter.Formatter) *Manager {
	return &Manager{
		client:    client,
		formatter: formatter,
	}
}

// BatchOperation represents a batch operation
type BatchOperation struct {
	Type     string
	Files    []string
	Options  map[string]interface{}
	Progress chan ProgressUpdate
	Cancel   context.CancelFunc
}

// ProgressUpdate represents a progress update
type ProgressUpdate struct {
	File     string
	Progress float64
	Status   string
	Error    error
}

// BatchUpload uploads multiple files in parallel
func (m *Manager) BatchUpload(ctx context.Context, files []string, options map[string]interface{}) (*BatchOperation, error) {
	ctx, cancel := context.WithCancel(ctx)

	operation := &BatchOperation{
		Type:     "upload",
		Files:    files,
		Options:  options,
		Progress: make(chan ProgressUpdate, len(files)),
		Cancel:   cancel,
	}

	go m.performBatchUpload(ctx, operation)
	return operation, nil
}

// performBatchUpload performs the actual batch upload
func (m *Manager) performBatchUpload(ctx context.Context, operation *BatchOperation) {
	defer close(operation.Progress)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent uploads

	for i, file := range operation.Files {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)
		go func(index int, filePath string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			progress := ProgressUpdate{
				File:   filePath,
				Status: "starting",
			}
			operation.Progress <- progress

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				progress.Status = "error"
				progress.Error = fmt.Errorf("file not found: %s", filePath)
				operation.Progress <- progress
				return
			}

			progress.Status = "uploading"
			operation.Progress <- progress

			// Upload file
			_, err := m.client.StoreFile(ctx, filePath)
			if err != nil {
				progress.Status = "error"
				progress.Error = err
			} else {
				progress.Status = "completed"
				progress.Progress = 100.0
			}

			operation.Progress <- progress
		}(i, file)
	}

	wg.Wait()
}

// BatchDownload downloads multiple files in parallel
func (m *Manager) BatchDownload(ctx context.Context, fileIDs []string, outputDir string, options map[string]interface{}) (*BatchOperation, error) {
	ctx, cancel := context.WithCancel(ctx)

	operation := &BatchOperation{
		Type:     "download",
		Files:    fileIDs,
		Options:  options,
		Progress: make(chan ProgressUpdate, len(fileIDs)),
		Cancel:   cancel,
	}

	go m.performBatchDownload(ctx, operation, outputDir)
	return operation, nil
}

// performBatchDownload performs the actual batch download
func (m *Manager) performBatchDownload(ctx context.Context, operation *BatchOperation, outputDir string) {
	defer close(operation.Progress)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent downloads

	for i, fileID := range operation.Files {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)
		go func(index int, id string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			progress := ProgressUpdate{
				File:   id,
				Status: "starting",
			}
			operation.Progress <- progress

			// Get file info first
			fileInfo, err := m.client.GetFile(ctx, id)
			if err != nil {
				progress.Status = "error"
				progress.Error = fmt.Errorf("failed to get file info: %w", err)
				operation.Progress <- progress
				return
			}

			progress.Status = "downloading"
			operation.Progress <- progress

			// Create output path
			outputPath := filepath.Join(outputDir, fileInfo.Key)
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				progress.Status = "error"
				progress.Error = fmt.Errorf("failed to create output directory: %w", err)
				operation.Progress <- progress
				return
			}

			// Download file
			err = m.client.DownloadFile(ctx, id, outputPath)
			if err != nil {
				progress.Status = "error"
				progress.Error = err
			} else {
				progress.Status = "completed"
				progress.Progress = 100.0
			}

			operation.Progress <- progress
		}(i, fileID)
	}

	wg.Wait()
}

// SyncDirectory synchronizes a local directory with the remote storage
func (m *Manager) SyncDirectory(ctx context.Context, localDir, remotePrefix string, options map[string]interface{}) error {
	m.formatter.PrintInfo(fmt.Sprintf("Synchronizing directory: %s", localDir))

	// Get local files
	localFiles, err := m.getLocalFiles(localDir)
	if err != nil {
		return fmt.Errorf("failed to get local files: %w", err)
	}

	// Get remote files
	remoteFiles, err := m.client.ListFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get remote files: %w", err)
	}

	// Create remote file map
	remoteMap := make(map[string]*client.FileInfo)
	for _, file := range remoteFiles.Files {
		if strings.HasPrefix(file.Key, remotePrefix) {
			remoteMap[file.Key] = &file
		}
	}

	// Find files to upload
	var toUpload []string
	for _, localFile := range localFiles {
		remoteKey := filepath.Join(remotePrefix, localFile)
		if remoteFile, exists := remoteMap[remoteKey]; !exists {
			toUpload = append(toUpload, filepath.Join(localDir, localFile))
		} else {
			// Check if file needs update (simplified - just check size)
			localPath := filepath.Join(localDir, localFile)
			if stat, err := os.Stat(localPath); err == nil {
				if stat.Size() != remoteFile.Size {
					toUpload = append(toUpload, localPath)
				}
			}
		}
	}

	if len(toUpload) > 0 {
		m.formatter.PrintInfo(fmt.Sprintf("Uploading %d files...", len(toUpload)))
		operation, err := m.BatchUpload(ctx, toUpload, options)
		if err != nil {
			return fmt.Errorf("failed to start batch upload: %w", err)
		}

		// Monitor progress
		for update := range operation.Progress {
			if update.Error != nil {
				m.formatter.PrintError(fmt.Errorf("upload error for %s: %w", update.File, update.Error))
			} else {
				m.formatter.PrintInfo(fmt.Sprintf("%s: %s (%.1f%%)", update.File, update.Status, update.Progress))
			}
		}
	} else {
		m.formatter.PrintInfo("Directory is already synchronized")
	}

	return nil
}

// getLocalFiles gets all files in a directory recursively
func (m *Manager) getLocalFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}

// SearchFiles searches for files by pattern
func (m *Manager) SearchFiles(ctx context.Context, pattern string, options map[string]interface{}) ([]*client.FileInfo, error) {
	m.formatter.PrintInfo(fmt.Sprintf("Searching for files matching: %s", pattern))

	// Get all files
	files, err := m.client.ListFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Filter by pattern
	var matches []*client.FileInfo
	for _, file := range files.Files {
		if matched, _ := filepath.Match(pattern, file.Key); matched {
			matches = append(matches, &file)
		}
	}

	m.formatter.PrintInfo(fmt.Sprintf("Found %d matching files", len(matches)))
	return matches, nil
}

// GetFileStats gets statistics about files
func (m *Manager) GetFileStats(ctx context.Context) (map[string]interface{}, error) {
	files, err := m.client.ListFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	stats := map[string]interface{}{
		"total_files": len(files.Files),
		"total_size":  int64(0),
		"file_types":  make(map[string]int),
		"owners":      make(map[string]int),
	}

	for _, file := range files.Files {
		stats["total_size"] = stats["total_size"].(int64) + file.Size

		// Count file types
		ext := filepath.Ext(file.Key)
		if ext == "" {
			ext = "no_extension"
		}
		stats["file_types"].(map[string]int)[ext]++

		// Count owners
		stats["owners"].(map[string]int)[file.Owner]++
	}

	return stats, nil
}

// CleanupOldFiles removes old files based on criteria
func (m *Manager) CleanupOldFiles(ctx context.Context, olderThan time.Duration, dryRun bool) error {
	m.formatter.PrintInfo(fmt.Sprintf("Cleaning up files older than %v", olderThan))

	files, err := m.client.ListFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	cutoff := time.Now().Add(-olderThan)
	var toDelete []string

	for _, file := range files.Files {
		if file.CreatedAt.Before(cutoff) {
			toDelete = append(toDelete, file.ID)
		}
	}

	if len(toDelete) == 0 {
		m.formatter.PrintInfo("No old files found")
		return nil
	}

	if dryRun {
		m.formatter.PrintInfo(fmt.Sprintf("Would delete %d files (dry run)", len(toDelete)))
		return nil
	}

	m.formatter.PrintInfo(fmt.Sprintf("Deleting %d old files...", len(toDelete)))

	// Delete files in parallel
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5)

	for _, fileID := range toDelete {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			err := m.client.DeleteFile(ctx, id)
			if err != nil {
				m.formatter.PrintError(fmt.Errorf("failed to delete file %s: %w", id, err))
			}
		}(fileID)
	}

	wg.Wait()
	m.formatter.PrintSuccess("Cleanup completed")

	return nil
}
