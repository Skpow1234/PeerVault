package backup

import (
	"archive/tar"
	"compress/gzip"
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

// Manager manages backup and recovery operations
type Manager struct {
	client    *client.Client
	configDir string
	backups   map[string]*Backup
	mu        sync.RWMutex
}

// Backup represents a backup operation
type Backup struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`   // "full", "incremental", "differential"
	Status      string                 `json:"status"` // "pending", "running", "completed", "failed"
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at,omitempty"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
	Size        int64                  `json:"size"`
	FileCount   int                    `json:"file_count"`
	FilePath    string                 `json:"file_path"`
	Metadata    map[string]interface{} `json:"metadata"`
	Error       string                 `json:"error,omitempty"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Schedule    string            `json:"schedule"`  // cron expression
	Retention   int               `json:"retention"` // days
	Compression bool              `json:"compression"`
	Encryption  bool              `json:"encryption"`
	Include     []string          `json:"include"`
	Exclude     []string          `json:"exclude"`
	Destination string            `json:"destination"`
	Options     map[string]string `json:"options"`
}

// RestorePoint represents a restore point
type RestorePoint struct {
	ID          string    `json:"id"`
	BackupID    string    `json:"backup_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	FileCount   int       `json:"file_count"`
	Description string    `json:"description"`
}

// New creates a new backup manager
func New(client *client.Client, configDir string) *Manager {
	bm := &Manager{
		client:    client,
		configDir: configDir,
		backups:   make(map[string]*Backup),
	}

	bm.loadBackups()
	return bm
}

// CreateBackup creates a new backup
func (bm *Manager) CreateBackup(config *BackupConfig) (*Backup, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	backupID := bm.generateID()

	backup := &Backup{
		ID:        backupID,
		Name:      config.Name,
		Type:      config.Type,
		Status:    "pending",
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Set file path
	backup.FilePath = filepath.Join(config.Destination, fmt.Sprintf("%s_%s.tar.gz", config.Name, backupID))

	bm.backups[backupID] = backup
	bm.saveBackups()

	return backup, nil
}

// StartBackup starts a backup operation
func (bm *Manager) StartBackup(backupID string, config *BackupConfig) error {
	bm.mu.Lock()
	backup, exists := bm.backups[backupID]
	if !exists {
		bm.mu.Unlock()
		return fmt.Errorf("backup not found: %s", backupID)
	}
	bm.mu.Unlock()

	// Update status
	backup.Status = "running"
	backup.StartedAt = time.Now()
	bm.saveBackups()

	// Start backup in goroutine
	go bm.performBackup(backup, config)

	return nil
}

// performBackup performs the actual backup operation
func (bm *Manager) performBackup(backup *Backup, config *BackupConfig) {
	defer func() {
		backup.CompletedAt = time.Now()
		bm.saveBackups()
	}()

	// Create backup file
	backupFile, err := os.Create(backup.FilePath)
	if err != nil {
		backup.Status = "failed"
		backup.Error = fmt.Sprintf("failed to create backup file: %v", err)
		return
	}
	defer backupFile.Close()

	var writer io.Writer = backupFile

	// Add compression if enabled
	if config.Compression {
		gzipWriter := gzip.NewWriter(backupFile)
		defer gzipWriter.Close()
		writer = gzipWriter
	}

	// Create tar writer
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	// Get files to backup
	files, err := bm.getFilesToBackup(config)
	if err != nil {
		backup.Status = "failed"
		backup.Error = fmt.Sprintf("failed to get files: %v", err)
		return
	}

	// Backup files
	fileCount := 0
	totalSize := int64(0)

	for _, file := range files {
		if err := bm.addFileToBackup(tarWriter, file); err != nil {
			backup.Status = "failed"
			backup.Error = fmt.Sprintf("failed to backup file %s: %v", file, err)
			return
		}

		// Get file size
		if stat, err := os.Stat(file); err == nil {
			totalSize += stat.Size()
		}
		fileCount++
	}

	// Update backup metadata
	backup.Status = "completed"
	backup.Size = totalSize
	backup.FileCount = fileCount
	backup.Metadata["compression"] = config.Compression
	backup.Metadata["encryption"] = config.Encryption
	backup.Metadata["file_count"] = fileCount
	backup.Metadata["total_size"] = totalSize
}

// getFilesToBackup gets the list of files to backup
func (bm *Manager) getFilesToBackup(config *BackupConfig) ([]string, error) {
	var files []string

	// If no include patterns specified, backup all files
	if len(config.Include) == 0 {
		// Get all files from the client
		fileList, err := bm.client.ListFiles(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, file := range fileList.Files {
			files = append(files, file.Key)
		}
	} else {
		// Use include patterns
		for _, pattern := range config.Include {
			matches, err := bm.matchFiles(pattern)
			if err != nil {
				return nil, fmt.Errorf("failed to match pattern %s: %w", pattern, err)
			}
			files = append(files, matches...)
		}
	}

	// Apply exclude patterns
	if len(config.Exclude) > 0 {
		var filteredFiles []string
		for _, file := range files {
			excluded := false
			for _, pattern := range config.Exclude {
				if matched, _ := filepath.Match(pattern, file); matched {
					excluded = true
					break
				}
			}
			if !excluded {
				filteredFiles = append(filteredFiles, file)
			}
		}
		files = filteredFiles
	}

	return files, nil
}

// matchFiles matches files against a pattern
func (bm *Manager) matchFiles(pattern string) ([]string, error) {
	// Simple pattern matching - in a real implementation, this would be more sophisticated
	fileList, err := bm.client.ListFiles(context.Background())
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, file := range fileList.Files {
		if matched, _ := filepath.Match(pattern, file.Key); matched {
			matches = append(matches, file.Key)
		}
	}

	return matches, nil
}

// addFileToBackup adds a file to the backup archive
func (bm *Manager) addFileToBackup(tarWriter *tar.Writer, filePath string) error {
	// Get file info from client
	fileInfo, err := bm.client.GetFile(context.Background(), filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create tar header
	header := &tar.Header{
		Name:    filePath,
		Size:    fileInfo.Size,
		Mode:    0644,
		ModTime: fileInfo.CreatedAt,
	}

	// Write header
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	// Download file content
	// In a real implementation, you would stream the file content
	// For now, we'll just write the file info as content
	content := fmt.Sprintf("File: %s\nSize: %d\nCreated: %s\nHash: %s\n",
		fileInfo.Key, fileInfo.Size, fileInfo.CreatedAt.Format(time.RFC3339), fileInfo.Hash)

	if _, err := tarWriter.Write([]byte(content)); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

// ListBackups returns all backups
func (bm *Manager) ListBackups() []*Backup {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var backups []*Backup
	for _, backup := range bm.backups {
		backups = append(backups, backup)
	}
	return backups
}

// GetBackup returns a backup by ID
func (bm *Manager) GetBackup(id string) (*Backup, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	backup, exists := bm.backups[id]
	if !exists {
		return nil, fmt.Errorf("backup not found: %s", id)
	}

	return backup, nil
}

// DeleteBackup deletes a backup
func (bm *Manager) DeleteBackup(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	backup, exists := bm.backups[id]
	if !exists {
		return fmt.Errorf("backup not found: %s", id)
	}

	// Delete backup file
	if err := os.Remove(backup.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	// Remove from memory
	delete(bm.backups, id)
	bm.saveBackups()

	return nil
}

// RestoreBackup restores from a backup
func (bm *Manager) RestoreBackup(backupID, destination string) error {
	backup, err := bm.GetBackup(backupID)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	if backup.Status != "completed" {
		return fmt.Errorf("backup is not completed: %s", backup.Status)
	}

	// Open backup file
	backupFile, err := os.Open(backup.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer backupFile.Close()

	var reader io.Reader = backupFile

	// Handle compression
	if backup.Metadata["compression"] == true {
		gzipReader, err := gzip.NewReader(backupFile)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Create tar reader
	tarReader := tar.NewReader(reader)

	// Restore files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Restore file
		if err := bm.restoreFile(tarReader, header, destination); err != nil {
			return fmt.Errorf("failed to restore file %s: %w", header.Name, err)
		}
	}

	return nil
}

// restoreFile restores a single file from backup
func (bm *Manager) restoreFile(tarReader *tar.Reader, header *tar.Header, destination string) error {
	// Create destination path
	destPath := filepath.Join(destination, header.Name)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	if _, err := io.Copy(file, tarReader); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// ScheduleBackup schedules a backup
func (bm *Manager) ScheduleBackup(config *BackupConfig) error {
	// In a real implementation, this would integrate with a cron scheduler
	// For now, we'll just store the configuration
	configFile := filepath.Join(bm.configDir, "backup_schedules.json")

	// Load existing schedules
	var schedules []*BackupConfig
	if data, err := os.ReadFile(configFile); err == nil {
		json.Unmarshal(data, &schedules)
	}

	// Add new schedule
	schedules = append(schedules, config)

	// Save schedules
	data, err := json.MarshalIndent(schedules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schedules: %w", err)
	}

	return os.WriteFile(configFile, data, 0644)
}

// ListScheduledBackups returns all scheduled backups
func (bm *Manager) ListScheduledBackups() ([]*BackupConfig, error) {
	configFile := filepath.Join(bm.configDir, "backup_schedules.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return []*BackupConfig{}, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read schedules file: %w", err)
	}

	var schedules []*BackupConfig
	if err := json.Unmarshal(data, &schedules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedules: %w", err)
	}

	return schedules, nil
}

// CleanupOldBackups removes old backups based on retention policy
func (bm *Manager) CleanupOldBackups(retentionDays int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	for id, backup := range bm.backups {
		if backup.CreatedAt.Before(cutoff) {
			// Delete backup file
			if err := os.Remove(backup.FilePath); err != nil && !os.IsNotExist(err) {
				continue // Skip if file deletion fails
			}

			// Remove from memory
			delete(bm.backups, id)
		}
	}

	bm.saveBackups()
	return nil
}

// Utility functions
func (bm *Manager) generateID() string {
	return fmt.Sprintf("backup_%d", time.Now().UnixNano())
}

func (bm *Manager) loadBackups() error {
	backupsFile := filepath.Join(bm.configDir, "backups.json")
	if _, err := os.Stat(backupsFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty backups
	}

	data, err := os.ReadFile(backupsFile)
	if err != nil {
		return fmt.Errorf("failed to read backups file: %w", err)
	}

	var backups []*Backup
	if err := json.Unmarshal(data, &backups); err != nil {
		return fmt.Errorf("failed to unmarshal backups: %w", err)
	}

	bm.backups = make(map[string]*Backup)
	for _, backup := range backups {
		bm.backups[backup.ID] = backup
	}

	return nil
}

func (bm *Manager) saveBackups() error {
	backupsFile := filepath.Join(bm.configDir, "backups.json")

	var backups []*Backup
	for _, backup := range bm.backups {
		backups = append(backups, backup)
	}

	data, err := json.MarshalIndent(backups, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backups: %w", err)
	}

	return os.WriteFile(backupsFile, data, 0644)
}
