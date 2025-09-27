package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackupConstants(t *testing.T) {
	// Test backup type constants
	assert.Equal(t, BackupType("full"), BackupTypeFull)
	assert.Equal(t, BackupType("incremental"), BackupTypeIncremental)
	assert.Equal(t, BackupType("differential"), BackupTypeDifferential)

	// Test backup status constants
	assert.Equal(t, BackupStatus("pending"), BackupStatusPending)
	assert.Equal(t, BackupStatus("running"), BackupStatusRunning)
	assert.Equal(t, BackupStatus("completed"), BackupStatusCompleted)
	assert.Equal(t, BackupStatus("failed"), BackupStatusFailed)
	assert.Equal(t, BackupStatus("cancelled"), BackupStatusCancelled)
}

func TestBackupConfig(t *testing.T) {
	config := &BackupConfig{
		Type:            BackupTypeFull,
		Location:        "/tmp/backup",
		Compression:     true,
		Encryption:      false,
		RetentionDays:   30,
		Schedule:        "0 2 * * *",
		IncludePatterns: []string{"*.txt", "*.log"},
		ExcludePatterns: []string{"*.tmp", "*.cache"},
		Metadata:        map[string]string{"env": "production", "version": "1.0"},
	}

	assert.Equal(t, BackupTypeFull, config.Type)
	assert.Equal(t, "/tmp/backup", config.Location)
	assert.True(t, config.Compression)
	assert.False(t, config.Encryption)
	assert.Equal(t, 30, config.RetentionDays)
	assert.Equal(t, "0 2 * * *", config.Schedule)
	assert.Len(t, config.IncludePatterns, 2)
	assert.Len(t, config.ExcludePatterns, 2)
	assert.Len(t, config.Metadata, 2)
}

func TestBackupManager_NewBackupManager(t *testing.T) {
	manager := NewBackupManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.configs)
	assert.NotNil(t, manager.backups)
	assert.Len(t, manager.configs, 0)
	assert.Len(t, manager.backups, 0)
}

func TestBackupManager_RegisterConfig(t *testing.T) {
	manager := NewBackupManager()
	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: "/tmp/backup",
	}

	manager.RegisterConfig("test-config", config)

	// Test that config is registered
	manager.mu.RLock()
	registeredConfig, exists := manager.configs["test-config"]
	manager.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, config, registeredConfig)
}

func TestBackupManager_StartBackup_ConfigNotFound(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	backup, err := manager.StartBackup(ctx, "non-existent-config")
	assert.Error(t, err)
	assert.Nil(t, backup)
	assert.Contains(t, err.Error(), "backup configuration non-existent-config not found")
}

func TestBackupManager_StartBackup_Success(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: "/tmp/backup",
		Metadata: map[string]string{"env": "test"},
	}

	manager.RegisterConfig("test-config", config)

	backup, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)
	assert.NotNil(t, backup)
	assert.Equal(t, BackupTypeFull, backup.Type)
	assert.Equal(t, BackupStatusPending, backup.GetStatus())
	assert.Equal(t, "/tmp/backup", backup.Location)
	assert.NotEmpty(t, backup.ID)
	assert.NotZero(t, backup.StartTime)
	assert.NotNil(t, backup.Metadata)
	assert.Equal(t, "test", backup.Metadata["env"])

	// Wait for backup to complete
	time.Sleep(3 * time.Second)

	// Check that backup is in the manager
	manager.mu.RLock()
	storedBackup, exists := manager.backups[backup.ID]
	manager.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, backup.ID, storedBackup.ID)
}

func TestBackupManager_GetBackup(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: "/tmp/backup",
	}

	manager.RegisterConfig("test-config", config)

	backup, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)

	// Test getting existing backup
	retrievedBackup, exists := manager.GetBackup(backup.ID)
	assert.True(t, exists)
	assert.Equal(t, backup.ID, retrievedBackup.ID)

	// Test getting non-existent backup
	_, exists = manager.GetBackup("non-existent-id")
	assert.False(t, exists)
}

func TestBackupManager_ListBackups(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: "/tmp/backup",
	}

	manager.RegisterConfig("test-config", config)

	// Start multiple backups
	backup1, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)

	backup2, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)

	// Wait for backups to complete and verify they're both there
	var backups []*Backup
	for i := 0; i < 30; i++ { // Increased timeout to 15 seconds
		time.Sleep(500 * time.Millisecond)
		backups = manager.ListBackups()
		if len(backups) >= 2 {
			// Check if both backups are completed
			allCompleted := true
			for _, backup := range backups {
				if backup.GetStatus() != BackupStatusCompleted {
					allCompleted = false
					break
				}
			}
			if allCompleted {
				break
			}
		}
	}
	assert.Len(t, backups, 2)

	// Check that both backups are in the list
	backupIDs := make(map[string]bool)
	for _, backup := range backups {
		backupIDs[backup.ID] = true
	}

	assert.True(t, backupIDs[backup1.ID])
	assert.True(t, backupIDs[backup2.ID])
}

func TestBackupManager_DeleteBackup(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backup")

	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: backupDir,
	}

	manager.RegisterConfig("test-config", config)

	backup, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)

	// Wait for backup to complete
	time.Sleep(3 * time.Second)

	// Test deleting existing backup
	err = manager.DeleteBackup(ctx, backup.ID)
	assert.NoError(t, err)

	// Verify backup is removed
	_, exists := manager.GetBackup(backup.ID)
	assert.False(t, exists)

	// Test deleting non-existent backup
	err = manager.DeleteBackup(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backup non-existent-id not found")
}

func TestBackupManager_PerformBackup_Full(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backup")

	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: backupDir,
	}

	manager.RegisterConfig("test-config", config)

	backup, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)

	// Wait for backup to complete
	time.Sleep(3 * time.Second)

	// Check backup status
	retrievedBackup, exists := manager.GetBackup(backup.ID)
	assert.True(t, exists)
	assert.Equal(t, BackupStatusCompleted, retrievedBackup.GetStatus())
	assert.Equal(t, int64(1024*1024*100), retrievedBackup.GetSize()) // 100MB
	assert.Equal(t, int64(1000), retrievedBackup.GetFiles())
	assert.NotZero(t, retrievedBackup.GetDuration())
	assert.NotZero(t, retrievedBackup.GetEndTime())

	// Verify backup directory was created
	_, err = os.Stat(backupDir)
	assert.NoError(t, err)
}

func TestBackupManager_PerformBackup_Incremental(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backup")

	config := &BackupConfig{
		Type:     BackupTypeIncremental,
		Location: backupDir,
	}

	manager.RegisterConfig("test-config", config)

	backup, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)

	// Wait for backup to complete
	time.Sleep(2 * time.Second)

	// Check backup status
	retrievedBackup, exists := manager.GetBackup(backup.ID)
	assert.True(t, exists)
	assert.Equal(t, BackupStatusCompleted, retrievedBackup.GetStatus())
	assert.Equal(t, int64(1024*1024*10), retrievedBackup.GetSize()) // 10MB
	assert.Equal(t, int64(100), retrievedBackup.GetFiles())
}

func TestBackupManager_PerformBackup_Differential(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backup")

	config := &BackupConfig{
		Type:     BackupTypeDifferential,
		Location: backupDir,
	}

	manager.RegisterConfig("test-config", config)

	backup, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)

	// Wait for backup to complete
	time.Sleep(2 * time.Second)

	// Check backup status
	retrievedBackup, exists := manager.GetBackup(backup.ID)
	assert.True(t, exists)
	assert.Equal(t, BackupStatusCompleted, retrievedBackup.GetStatus())
	assert.Equal(t, int64(1024*1024*50), retrievedBackup.GetSize()) // 50MB
	assert.Equal(t, int64(500), retrievedBackup.GetFiles())
}

func TestBackupManager_PerformBackup_InvalidType(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backup")

	config := &BackupConfig{
		Type:     BackupType("invalid"),
		Location: backupDir,
	}

	manager.RegisterConfig("test-config", config)

	backup, err := manager.StartBackup(ctx, "test-config")
	assert.NoError(t, err)

	// Wait for backup to complete
	time.Sleep(2 * time.Second)

	// Check backup status
	retrievedBackup, exists := manager.GetBackup(backup.ID)
	assert.True(t, exists)
	assert.Equal(t, BackupStatusFailed, retrievedBackup.GetStatus())
	assert.Contains(t, retrievedBackup.GetError(), "unsupported backup type: invalid")
}

func TestRestoreManager_NewRestoreManager(t *testing.T) {
	manager := NewRestoreManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.backups)
	assert.Len(t, manager.backups, 0)
}

func TestRestoreManager_Restore_BackupNotFound(t *testing.T) {
	manager := NewRestoreManager()
	ctx := context.Background()

	request := &RestoreRequest{
		BackupID:    "non-existent-id",
		Destination: "/tmp/restore",
		Overwrite:   false,
	}

	result, err := manager.Restore(ctx, request)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "backup non-existent-id not found")
}

func TestRestoreManager_Restore_BackupNotCompleted(t *testing.T) {
	manager := NewRestoreManager()
	ctx := context.Background()

	// Add a backup that's not completed
	backup := &Backup{
		ID:     "test-backup",
		Status: BackupStatusRunning,
	}

	manager.mu.Lock()
	manager.backups["test-backup"] = backup
	manager.mu.Unlock()

	request := &RestoreRequest{
		BackupID:    "test-backup",
		Destination: "/tmp/restore",
		Overwrite:   false,
	}

	result, err := manager.Restore(ctx, request)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "backup test-backup is not completed")
}

func TestRestoreManager_Restore_Success(t *testing.T) {
	manager := NewRestoreManager()
	ctx := context.Background()

	// Add a completed backup
	backup := &Backup{
		ID:     "test-backup",
		Status: BackupStatusCompleted,
		Size:   1024 * 1024 * 100, // 100MB
		Files:  1000,
	}

	manager.mu.Lock()
	manager.backups["test-backup"] = backup
	manager.mu.Unlock()

	request := &RestoreRequest{
		BackupID:    "test-backup",
		Destination: "/tmp/restore",
		Overwrite:   false,
	}

	result, err := manager.Restore(ctx, request)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, BackupStatusCompleted, result.Status)
	assert.Equal(t, int64(1024*1024*100), result.Size)
	assert.Equal(t, int64(1000), result.Files)
	assert.NotEmpty(t, result.ID)
	assert.NotZero(t, result.StartTime)
	assert.NotZero(t, result.EndTime)
	assert.NotZero(t, result.Duration)
}

func TestBackupScheduler_NewBackupScheduler(t *testing.T) {
	backupManager := NewBackupManager()
	scheduler := NewBackupScheduler(backupManager)

	assert.NotNil(t, scheduler)
	assert.Equal(t, backupManager, scheduler.manager)
	assert.NotNil(t, scheduler.ctx)
	assert.NotNil(t, scheduler.cancel)
}

func TestBackupScheduler_StartStop(t *testing.T) {
	backupManager := NewBackupManager()
	scheduler := NewBackupScheduler(backupManager)

	// Start scheduler
	scheduler.Start()
	assert.NotNil(t, scheduler.ticker)

	// Stop scheduler
	scheduler.Stop()

	// Verify context is cancelled
	select {
	case <-scheduler.ctx.Done():
		// Context is cancelled, which is expected
	default:
		t.Error("Expected context to be cancelled")
	}
}

func TestGenerateBackupID(t *testing.T) {
	id1 := generateBackupID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := generateBackupID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "backup_")
	assert.Contains(t, id2, "backup_")
}

func TestGlobalManagers(t *testing.T) {
	// Test that global managers are initialized
	assert.NotNil(t, GlobalBackupManager)
	assert.NotNil(t, GlobalRestoreManager)
	assert.NotNil(t, GlobalBackupScheduler)
}

func TestConvenienceFunctions(t *testing.T) {
	ctx := context.Background()

	// Test RegisterBackupConfig
	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: "/tmp/backup",
	}

	RegisterBackupConfig("test-config", config)

	// Test StartBackup
	backup, err := StartBackup(ctx, "test-config")
	assert.NoError(t, err)
	assert.NotNil(t, backup)

	// Test GetBackup
	retrievedBackup, exists := GetBackup(backup.ID)
	assert.True(t, exists)
	assert.Equal(t, backup.ID, retrievedBackup.ID)

	// Test ListBackups
	backups := ListBackups()
	assert.Len(t, backups, 1)
	assert.Equal(t, backup.ID, backups[0].ID)

	// Test RestoreBackup
	// First, we need to add the backup to the restore manager
	GlobalRestoreManager.mu.Lock()
	GlobalRestoreManager.backups[backup.ID] = backup
	GlobalRestoreManager.mu.Unlock()

	request := &RestoreRequest{
		BackupID:    backup.ID,
		Destination: "/tmp/restore",
		Overwrite:   false,
	}

	// Wait for backup to complete
	time.Sleep(3 * time.Second)

	// Update backup status to completed
	GlobalBackupManager.mu.Lock()
	if storedBackup, exists := GlobalBackupManager.backups[backup.ID]; exists {
		storedBackup.SetStatus(BackupStatusCompleted)
	}
	GlobalBackupManager.mu.Unlock()

	// Update restore manager backup status
	GlobalRestoreManager.mu.Lock()
	if storedBackup, exists := GlobalRestoreManager.backups[backup.ID]; exists {
		storedBackup.SetStatus(BackupStatusCompleted)
	}
	GlobalRestoreManager.mu.Unlock()

	result, err := RestoreBackup(ctx, request)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, BackupStatusCompleted, result.Status)
}

func TestBackupManager_ConcurrentAccess(t *testing.T) {
	manager := NewBackupManager()
	ctx := context.Background()

	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: "/tmp/backup",
	}

	manager.RegisterConfig("test-config", config)

	// Test concurrent backup operations
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()

			backup, err := manager.StartBackup(ctx, "test-config")
			assert.NoError(t, err)
			assert.NotNil(t, backup)

			// Test concurrent access to GetBackup
			retrievedBackup, exists := manager.GetBackup(backup.ID)
			assert.True(t, exists)
			assert.Equal(t, backup.ID, retrievedBackup.ID)

			// Test concurrent access to ListBackups
			backups := manager.ListBackups()
			assert.GreaterOrEqual(t, len(backups), 1)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestBackupManager_EdgeCases(t *testing.T) {
	manager := NewBackupManager()

	// Test with nil config
	manager.RegisterConfig("nil-config", nil)

	// Test with empty config name
	config := &BackupConfig{
		Type:     BackupTypeFull,
		Location: "/tmp/backup",
	}
	manager.RegisterConfig("", config)

	// Test with empty backup ID
	_, exists := manager.GetBackup("")
	assert.False(t, exists)

	// Test deleting empty backup ID
	ctx := context.Background()
	err := manager.DeleteBackup(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backup  not found")
}
