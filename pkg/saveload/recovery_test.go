//go:build !js
// +build !js

package saveload

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSaveManager_SaveGameWithBackup tests saving with backup creation.
func TestSaveManager_SaveGameWithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create initial save
	save := NewGameSave()
	save.PlayerState.Level = 10
	save.PlayerState.Experience = 5000

	err = manager.SaveGameWithBackup("test", save)
	if err != nil {
		t.Fatalf("SaveGameWithBackup failed: %v", err)
	}

	// Verify save file exists
	savePath := manager.getFilePath("test")
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Save file was not created")
	}

	// Verify checksum file exists
	checksumPath := savePath + ".sha256"
	if _, err := os.Stat(checksumPath); os.IsNotExist(err) {
		t.Error("Checksum file was not created")
	}

	// Update save (should create backup)
	save.PlayerState.Level = 20
	save.PlayerState.Experience = 15000

	err = manager.SaveGameWithBackup("test", save)
	if err != nil {
		t.Fatalf("Second SaveGameWithBackup failed: %v", err)
	}

	// Verify backup file exists
	backupPath := savePath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Load and verify it has the new values
	loaded, err := manager.LoadGame("test")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	if loaded.PlayerState.Level != 20 {
		t.Errorf("Expected Level 20, got %d", loaded.PlayerState.Level)
	}
	if loaded.PlayerState.Experience != 15000 {
		t.Errorf("Expected Experience 15000, got %d", loaded.PlayerState.Experience)
	}
}

// TestSaveManager_LoadGameWithRecovery tests loading with corruption detection.
func TestSaveManager_LoadGameWithRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create and save a valid game
	save := NewGameSave()
	save.PlayerState.Level = 15
	save.PlayerState.X = 100.0

	err = manager.SaveGameWithBackup("recovery_test", save)
	if err != nil {
		t.Fatalf("SaveGameWithBackup failed: %v", err)
	}

	// Load with recovery (should succeed normally)
	loaded, err := manager.LoadGameWithRecovery("recovery_test")
	if err != nil {
		t.Fatalf("LoadGameWithRecovery failed: %v", err)
	}

	if loaded.PlayerState.Level != 15 {
		t.Errorf("Expected Level 15, got %d", loaded.PlayerState.Level)
	}
}

// TestSaveManager_RecoverFromBackup tests recovery from corrupted save.
func TestSaveManager_RecoverFromBackup(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create initial save
	save := NewGameSave()
	save.PlayerState.Level = 10
	save.PlayerState.Experience = 5000

	err = manager.SaveGameWithBackup("corrupt_test", save)
	if err != nil {
		t.Fatalf("SaveGameWithBackup failed: %v", err)
	}

	// Update save to create a backup
	save.PlayerState.Level = 20
	err = manager.SaveGameWithBackup("corrupt_test", save)
	if err != nil {
		t.Fatalf("Second SaveGameWithBackup failed: %v", err)
	}

	// Corrupt the save file
	savePath := manager.getFilePath("corrupt_test")
	err = os.WriteFile(savePath, []byte("corrupted data {{{"), 0o644)
	if err != nil {
		t.Fatalf("Failed to corrupt save file: %v", err)
	}

	// Attempt to load with recovery (should recover from backup)
	loaded, err := manager.LoadGameWithRecovery("corrupt_test")
	if err != nil {
		t.Fatalf("LoadGameWithRecovery failed: %v", err)
	}

	// Should have recovered the backup (level 10, before corruption)
	// Note: Backup contains the previous version before the second save
	if loaded.PlayerState.Level != 10 {
		t.Errorf("Expected Level 10 (from backup), got %d", loaded.PlayerState.Level)
	}
}

// TestSaveManager_ChecksumValidation tests checksum validation.
func TestSaveManager_ChecksumValidation(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create save with checksum
	save := NewGameSave()
	save.PlayerState.Level = 10

	err = manager.SaveGameWithBackup("checksum_test", save)
	if err != nil {
		t.Fatalf("SaveGameWithBackup failed: %v", err)
	}

	// Validate checksum (should be valid)
	valid, err := manager.validateChecksum("checksum_test")
	if err != nil {
		t.Fatalf("validateChecksum failed: %v", err)
	}
	if !valid {
		t.Error("Checksum should be valid")
	}

	// Corrupt the save file
	savePath := manager.getFilePath("checksum_test")
	data, _ := os.ReadFile(savePath)
	data[0] = '^' // Corrupt first byte
	os.WriteFile(savePath, data, 0o644)

	// Validate checksum (should be invalid)
	valid, err = manager.validateChecksum("checksum_test")
	if err != nil {
		t.Fatalf("validateChecksum failed: %v", err)
	}
	if valid {
		t.Error("Checksum should be invalid after corruption")
	}
}

// TestSaveManager_BackupExists tests backup existence check.
func TestSaveManager_BackupExists(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// No backup initially
	if manager.BackupExists("backup_test") {
		t.Error("Backup should not exist initially")
	}

	// Create initial save (no backup yet)
	save := NewGameSave()
	err = manager.SaveGameWithBackup("backup_test", save)
	if err != nil {
		t.Fatalf("SaveGameWithBackup failed: %v", err)
	}

	// Still no backup after first save
	if manager.BackupExists("backup_test") {
		t.Error("Backup should not exist after first save")
	}

	// Update save (should create backup)
	save.PlayerState.Level = 20
	err = manager.SaveGameWithBackup("backup_test", save)
	if err != nil {
		t.Fatalf("Second SaveGameWithBackup failed: %v", err)
	}

	// Backup should exist now
	if !manager.BackupExists("backup_test") {
		t.Error("Backup should exist after second save")
	}
}

// TestSaveManager_CleanupBackups tests backup cleanup.
func TestSaveManager_CleanupBackups(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create save with backup
	save := NewGameSave()
	err = manager.SaveGameWithBackup("cleanup_test", save)
	if err != nil {
		t.Fatalf("SaveGameWithBackup failed: %v", err)
	}

	save.PlayerState.Level = 20
	err = manager.SaveGameWithBackup("cleanup_test", save)
	if err != nil {
		t.Fatalf("Second SaveGameWithBackup failed: %v", err)
	}

	// Verify backup and checksum exist
	if !manager.BackupExists("cleanup_test") {
		t.Error("Backup should exist before cleanup")
	}

	savePath := manager.getFilePath("cleanup_test")
	checksumPath := savePath + ".sha256"
	if _, err := os.Stat(checksumPath); os.IsNotExist(err) {
		t.Error("Checksum should exist before cleanup")
	}

	// Cleanup backups
	err = manager.CleanupBackups("cleanup_test")
	if err != nil {
		t.Fatalf("CleanupBackups failed: %v", err)
	}

	// Verify backup and checksum are gone
	if manager.BackupExists("cleanup_test") {
		t.Error("Backup should not exist after cleanup")
	}

	if _, err := os.Stat(checksumPath); !os.IsNotExist(err) {
		t.Error("Checksum should not exist after cleanup")
	}

	// Save file should still exist
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("Save file should still exist after cleanup")
	}
}

// TestSaveManager_ListBackups tests listing backup files.
func TestSaveManager_ListBackups(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create multiple saves with backups
	saveNames := []string{"list_test_1", "list_test_2", "list_test_3"}
	for _, name := range saveNames {
		save := NewGameSave()
		
		// First save
		err = manager.SaveGameWithBackup(name, save)
		if err != nil {
			t.Fatalf("SaveGameWithBackup failed for %s: %v", name, err)
		}

		// Second save to create backup
		save.PlayerState.Level = 10
		err = manager.SaveGameWithBackup(name, save)
		if err != nil {
			t.Fatalf("Second SaveGameWithBackup failed for %s: %v", name, err)
		}
	}

	// List backups
	backups, err := manager.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("Expected 3 backups, got %d", len(backups))
	}
}

// TestSaveManager_RecoveryWithoutBackup tests recovery when no backup exists.
func TestSaveManager_RecoveryWithoutBackup(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create corrupted save without backup
	savePath := manager.getFilePath("no_backup")
	err = os.WriteFile(savePath, []byte("corrupted{{{"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Attempt to load with recovery (should fail)
	_, err = manager.LoadGameWithRecovery("no_backup")
	if err == nil {
		t.Error("LoadGameWithRecovery should fail for corrupted save without backup")
	}
}

// TestSaveManager_RecoveryWithCorruptedBackup tests recovery when backup is also corrupted.
func TestSaveManager_RecoveryWithCorruptedBackup(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	// Create corrupted save
	savePath := manager.getFilePath("bad_backup")
	err = os.WriteFile(savePath, []byte("corrupted{{{"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Create corrupted backup
	backupPath := savePath + ".bak"
	err = os.WriteFile(backupPath, []byte("also corrupted{{{"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create corrupted backup: %v", err)
	}

	// Attempt to load with recovery (should fail)
	_, err = manager.LoadGameWithRecovery("bad_backup")
	if err == nil {
		t.Error("LoadGameWithRecovery should fail when both save and backup are corrupted")
	}
}

// TestSaveManager_GetBackupPath tests getting backup file path.
func TestSaveManager_GetBackupPath(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	backupPath := manager.GetBackupPath("test")
	expectedPath := filepath.Join(tmpDir, "test.sav.bak")

	if backupPath != expectedPath {
		t.Errorf("Expected backup path %s, got %s", expectedPath, backupPath)
	}
}

// TestSaveManager_SaveGameWithBackupNil tests saving nil save.
func TestSaveManager_SaveGameWithBackupNil(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	err = manager.SaveGameWithBackup("nil_test", nil)
	if err == nil {
		t.Error("SaveGameWithBackup should fail for nil save")
	}
}

// TestSaveManager_SaveGameWithBackupInvalidName tests saving with invalid name.
func TestSaveManager_SaveGameWithBackupInvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	save := NewGameSave()
	err = manager.SaveGameWithBackup("../invalid", save)
	if err == nil {
		t.Error("SaveGameWithBackup should fail for invalid name")
	}
}

// TestSaveManager_LoadGameWithRecoveryNonexistent tests loading nonexistent file.
func TestSaveManager_LoadGameWithRecoveryNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewSaveManager(tmpDir)
	if err != nil {
		t.Fatalf("NewSaveManager failed: %v", err)
	}

	_, err = manager.LoadGameWithRecovery("nonexistent")
	if err == nil {
		t.Error("LoadGameWithRecovery should fail for nonexistent file")
	}
}

// TestSaveManager_ChecksumFile tests checksum computation.
func TestSaveManager_ChecksumFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	content := []byte("test content")
	err := os.WriteFile(testFile, content, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Compute checksum
	checksum1, err := checksumFile(testFile)
	if err != nil {
		t.Fatalf("checksumFile failed: %v", err)
	}

	if checksum1 == "" {
		t.Error("Checksum should not be empty")
	}

	// Compute again (should be same)
	checksum2, err := checksumFile(testFile)
	if err != nil {
		t.Fatalf("checksumFile failed: %v", err)
	}

	if checksum1 != checksum2 {
		t.Error("Checksums should be identical for same content")
	}

	// Change content
	err = os.WriteFile(testFile, []byte("different content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	// Checksum should be different
	checksum3, err := checksumFile(testFile)
	if err != nil {
		t.Fatalf("checksumFile failed: %v", err)
	}

	if checksum1 == checksum3 {
		t.Error("Checksums should differ for different content")
	}
}

// TestSaveManager_ChecksumFileNonexistent tests checksum of nonexistent file.
func TestSaveManager_ChecksumFileNonexistent(t *testing.T) {
	_, err := checksumFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("checksumFile should fail for nonexistent file")
	}
}
