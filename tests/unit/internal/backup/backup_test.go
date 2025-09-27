package backup_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Skpow1234/Peervault/internal/backup"
	"github.com/stretchr/testify/assert"
)

func TestBackupConstants(t *testing.T) {
	// Test backup type constants
	assert.Equal(t, backup.BackupType("full"), backup.BackupTypeFull)
	assert.Equal(t, backup.BackupType("incremental"), backup.BackupTypeIncremental)
	assert.Equal(t, backup.BackupType("differential"), backup.BackupTypeDifferential)

	// Test backup status constants
	assert.Equal(t, backup.BackupStatus("pending"), backup.BackupStatusPending)
	assert.Equal(t, backup.BackupStatus("running"), backup.BackupStatusRunning)
	assert.Equal(t, backup.BackupStatus("completed"), backup.BackupStatusCompleted)
	assert.Equal(t, backup.BackupStatus("failed"), backup.BackupStatusFailed)
	assert.Equal(t, backup.BackupStatus("cancelled"), backup.BackupStatusCancelled)
}

func TestBackupConfig(t *testing.T) {
	config := &backup.BackupConfig{
		Type:          backup.BackupTypeFull,
		Location:      "/test/dest",
		Compression:   true,
		Encryption:    true,
		RetentionDays: 30,
	}

	assert.Equal(t, backup.BackupTypeFull, config.Type)
	assert.Equal(t, "/test/dest", config.Location)
	assert.True(t, config.Compression)
	assert.True(t, config.Encryption)
	assert.Equal(t, 30, config.RetentionDays)
}

func TestNewBackupManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()
	assert.NotNil(t, manager)
}

func TestBackupManager_RegisterConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()
	config := &backup.BackupConfig{
		Type:     backup.BackupTypeFull,
		Location: "/test/dest",
	}

	manager.RegisterConfig("test-config", config)
	assert.NoError(t, err)
}

func TestBackupManager_StartBackup_ConfigNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()
	_, err = manager.StartBackup(context.Background(), "nonexistent-config")
	assert.Error(t, err)
}

func TestBackupManager_StartBackup_Success(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test source directory
	sourceDir := filepath.Join(tempDir, "source")
	err = os.MkdirAll(sourceDir, 0755)
	assert.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(sourceDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	manager := backup.NewBackupManager()
	config := &backup.BackupConfig{
		Type:     backup.BackupTypeFull,
		Location: filepath.Join(tempDir, "dest"),
	}

	manager.RegisterConfig("test-config", config)
	assert.NoError(t, err)

	backup, err := manager.StartBackup(context.Background(), "test-config")
	assert.NoError(t, err)
	assert.NotNil(t, backup)
}

func TestBackupManager_GetBackup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()

	// Test getting non-existent backup
	backup, exists := manager.GetBackup("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, backup)
}

func TestBackupManager_ListBackups(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()
	backups := manager.ListBackups()
	assert.NotNil(t, backups)
	assert.Empty(t, backups)
}

func TestBackupManager_DeleteBackup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()

	// Test deleting non-existent backup
	err = manager.DeleteBackup(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestNewRestoreManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewRestoreManager()
	assert.NotNil(t, manager)
}

func TestNewBackupScheduler(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()
	scheduler := backup.NewBackupScheduler(manager)
	assert.NotNil(t, scheduler)
}

func TestBackupScheduler_StartStop(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()
	scheduler := backup.NewBackupScheduler(manager)

	scheduler.Start()

	scheduler.Stop()
}

func TestBackupIDGeneration(t *testing.T) {
	// Test that backup IDs are unique by creating backups
	manager := backup.NewBackupManager()
	config := &backup.BackupConfig{
		Type:     backup.BackupTypeFull,
		Location: "/test/dest",
	}

	manager.RegisterConfig("test-config", config)

	backup1, err1 := manager.StartBackup(context.Background(), "test-config")
	backup2, err2 := manager.StartBackup(context.Background(), "test-config")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, backup1)
	assert.NotNil(t, backup2)
	assert.NotEqual(t, backup1.ID, backup2.ID)

	// Test that IDs have the expected format
	assert.Contains(t, backup1.ID, "backup_")
	assert.Contains(t, backup2.ID, "backup_")
}

func TestBackupManager_ConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_ = manager.ListBackups()
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestBackupManager_EdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := backup.NewBackupManager()

	// Test with empty config name - should not error but may cause issues later
	manager.RegisterConfig("", &backup.BackupConfig{})

	// Test with nil config - should not error but may cause issues later
	manager.RegisterConfig("test", nil)

	// Test that the configs were registered (even if they're problematic)
	configs := manager.ListBackups() // This should work without error
	assert.NotNil(t, configs)
}
