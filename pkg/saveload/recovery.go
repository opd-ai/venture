//go:build !js
// +build !js

package saveload

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// checksumFile computes SHA256 checksum of a file.
func checksumFile(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// createBackup creates a backup copy of a save file.
// Returns the backup file path or error.
func (m *SaveManager) createBackup(name string) (string, error) {
	sourcePath := m.getFilePath(name)

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// No file to backup, not an error
		return "", nil
	}

	// Create backup filename (.bak extension)
	backupPath := sourcePath + ".bak"

	// Copy file to backup
	source, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file for backup: %w", err)
	}
	defer source.Close()

	backup, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backup.Close()

	if _, err := io.Copy(backup, source); err != nil {
		return "", fmt.Errorf("failed to copy to backup: %w", err)
	}

	m.logDebug("created backup", logrus.Fields{
		"name":   name,
		"backup": backupPath,
	})

	return backupPath, nil
}

// validateChecksum validates a save file's checksum.
// Returns (valid, hasChecksum) where:
//   - valid is true if checksum exists and matches
//   - hasChecksum is true if a checksum file exists (regardless of match)
//
// This allows callers to distinguish between "no checksum" and "checksum mismatch".
func (m *SaveManager) validateChecksum(name string) (bool, bool) {
	savePath := m.getFilePath(name)
	checksumPath := savePath + ".sha256"

	// Check if checksum file exists
	if _, err := os.Stat(checksumPath); os.IsNotExist(err) {
		// No checksum file - not an error, just no checksum available
		return false, false
	}

	// Read stored checksum
	storedChecksum, err := os.ReadFile(checksumPath)
	if err != nil {
		m.logWarn("failed to read checksum file", err, logrus.Fields{"name": name})
		return false, true // Has checksum file but can't read it - treat as mismatch
	}

	// Compute current checksum
	currentChecksum, err := checksumFile(savePath)
	if err != nil {
		m.logWarn("failed to compute checksum", err, logrus.Fields{"name": name})
		return false, true // Has checksum file but can't compute - treat as mismatch
	}

	// Compare checksums
	return strings.TrimSpace(string(storedChecksum)) == currentChecksum, true
}

// saveChecksum computes and saves checksum for a save file.
func (m *SaveManager) saveChecksum(name string) error {
	savePath := m.getFilePath(name)
	checksumPath := savePath + ".sha256"

	// Compute checksum
	checksum, err := checksumFile(savePath)
	if err != nil {
		return err
	}

	// Write checksum to file
	if err := os.WriteFile(checksumPath, []byte(checksum), 0o644); err != nil {
		return fmt.Errorf("failed to write checksum file: %w", err)
	}

	m.logDebug("saved checksum", logrus.Fields{
		"name":     name,
		"checksum": checksum,
	})

	return nil
}

// recoverFromBackup attempts to recover a corrupted save from its backup.
// Returns true if recovery was successful, false otherwise.
func (m *SaveManager) recoverFromBackup(name string) (bool, error) {
	savePath := m.getFilePath(name)
	backupPath := savePath + ".bak"

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		m.logWarn("no backup found for recovery", nil, logrus.Fields{
			"name": name,
		})
		return false, nil
	}

	m.logInfo("attempting recovery from backup", logrus.Fields{
		"name":   name,
		"backup": backupPath,
	})

	// Validate backup by trying to load it
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return false, fmt.Errorf("failed to read backup: %w", err)
	}

	// Try to unmarshal backup to validate it
	_, err = m.unmarshalSave(backupData, name)
	if err != nil {
		m.logWarn("backup is also corrupted", err, logrus.Fields{
			"name": name,
		})
		return false, nil
	}

	// Backup is valid, restore it
	if err := os.WriteFile(savePath, backupData, 0o644); err != nil {
		return false, fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Update checksum for restored file
	if err := m.saveChecksum(name); err != nil {
		m.logWarn("failed to save checksum after recovery", err, logrus.Fields{
			"name": name,
		})
		// Not a critical error, continue
	}

	m.logInfo("successfully recovered from backup", logrus.Fields{
		"name": name,
	})

	return true, nil
}

// SaveGameWithBackup saves the game state with automatic backup and checksum validation.
// This is the recommended method for production use.
func (m *SaveManager) SaveGameWithBackup(name string, save *GameSave) error {
	m.logDebug("saving game with backup", logrus.Fields{"name": name})

	if save == nil {
		return fmt.Errorf("save cannot be nil")
	}

	if err := m.validateSaveName(name); err != nil {
		m.logWarn("invalid save name", err, logrus.Fields{"name": name})
		return err
	}

	// Create backup before overwriting (if file exists)
	backupPath, err := m.createBackup(name)
	if err != nil {
		m.logWarn("failed to create backup", err, logrus.Fields{"name": name})
		// Continue with save anyway, backup failure shouldn't block saving
	}

	// Perform the save
	save.Version = SaveVersion
	save.Timestamp = time.Now()

	data, err := m.marshalSave(save, name)
	if err != nil {
		return err
	}

	if err := m.writeSaveFile(name, data); err != nil {
		// If save failed and we have a backup, restore it
		if backupPath != "" {
			m.logWarn("save failed, restoring backup", err, logrus.Fields{
				"name":   name,
				"backup": backupPath,
			})
			// Attempt to restore backup (best effort)
			if backupData, readErr := os.ReadFile(backupPath); readErr == nil {
				if restoreErr := m.writeSaveFile(name, backupData); restoreErr != nil {
					m.logWarn("backup restoration also failed", restoreErr, logrus.Fields{
						"name":   name,
						"backup": backupPath,
					})
				}
			}
		}
		return err
	}

	// Save checksum
	if err := m.saveChecksum(name); err != nil {
		m.logWarn("failed to save checksum", err, logrus.Fields{"name": name})
		// Not critical, continue
	}

	m.logInfo("game saved successfully with backup", logrus.Fields{
		"name":      name,
		"size":      len(data),
		"timestamp": save.Timestamp,
		"backup":    backupPath != "",
	})

	return nil
}

// LoadGameWithRecovery loads the game state with automatic corruption detection and recovery.
// This is the recommended method for production use.
func (m *SaveManager) LoadGameWithRecovery(name string) (*GameSave, error) {
	m.logDebug("loading game with recovery", logrus.Fields{"name": name})

	if err := m.validateSaveName(name); err != nil {
		m.logWarn("invalid save name", err, logrus.Fields{"name": name})
		return nil, err
	}

	if err := m.verifySaveFileExists(name); err != nil {
		return nil, err
	}

	if err := m.handleChecksumValidation(name); err != nil {
		return nil, err
	}

	save, err := m.loadAndParseSave(name)
	if err != nil {
		return nil, err
	}

	if err := m.validateAndMigrate(save); err != nil {
		m.logError("failed to validate/migrate save", err, logrus.Fields{"name": name})
		return nil, fmt.Errorf("failed to validate/migrate save: %w", err)
	}

	m.logInfo("game loaded successfully", logrus.Fields{
		"name":      name,
		"version":   save.Version,
		"timestamp": save.Timestamp,
	})

	return save, nil
}

// verifySaveFileExists checks if the save file exists before attempting to load.
func (m *SaveManager) verifySaveFileExists(name string) error {
	savePath := m.getFilePath(name)
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return fmt.Errorf("save file not found: %s", name)
	}
	return nil
}

// handleChecksumValidation validates checksum and attempts recovery if corrupted.
func (m *SaveManager) handleChecksumValidation(name string) error {
	valid, hasChecksum := m.validateChecksum(name)
	if !hasChecksum || valid {
		return nil
	}

	m.logWarn("checksum validation failed, save may be corrupted", nil, logrus.Fields{"name": name})
	recovered, recErr := m.recoverFromBackup(name)
	if recErr != nil {
		return fmt.Errorf("failed to recover from backup: %w", recErr)
	}
	if !recovered {
		m.logWarn("recovery failed, attempting to load corrupted file", nil, logrus.Fields{"name": name})
	}
	return nil
}

// loadAndParseSave loads save data and attempts recovery if parsing fails.
func (m *SaveManager) loadAndParseSave(name string) (*GameSave, error) {
	data, err := m.readSaveFile(name)
	if err != nil {
		return nil, err
	}

	save, err := m.unmarshalSave(data, name)
	if err != nil {
		return m.recoverAndRetryLoad(name, err)
	}

	return save, nil
}

// recoverAndRetryLoad attempts recovery from backup and retries save loading.
func (m *SaveManager) recoverAndRetryLoad(name string, originalErr error) (*GameSave, error) {
	m.logError("failed to parse save, attempting recovery", originalErr, logrus.Fields{"name": name})

	recovered, recErr := m.recoverFromBackup(name)
	if recErr != nil {
		return nil, fmt.Errorf("failed to recover from backup: %w", recErr)
	}
	if !recovered {
		return nil, fmt.Errorf("save corrupted and no valid backup available: %w", originalErr)
	}

	data, err := m.readSaveFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load after recovery: %w", err)
	}

	save, err := m.unmarshalSave(data, name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse after recovery: %w", err)
	}

	m.logInfo("successfully loaded after recovery", logrus.Fields{"name": name})
	return save, nil
}

// CleanupBackups removes old backup files for a save.
func (m *SaveManager) CleanupBackups(name string) error {
	if err := m.validateSaveName(name); err != nil {
		return err
	}

	savePath := m.getFilePath(name)
	backupPath := savePath + ".bak"

	// Remove backup file if it exists
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Remove(backupPath); err != nil {
			return fmt.Errorf("failed to remove backup: %w", err)
		}
		m.logDebug("removed backup", logrus.Fields{
			"name":   name,
			"backup": backupPath,
		})
	}

	// Remove checksum file if it exists
	checksumPath := savePath + ".sha256"
	if _, err := os.Stat(checksumPath); err == nil {
		if err := os.Remove(checksumPath); err != nil {
			return fmt.Errorf("failed to remove checksum: %w", err)
		}
		m.logDebug("removed checksum", logrus.Fields{
			"name":     name,
			"checksum": checksumPath,
		})
	}

	return nil
}

// ListBackups returns paths to all backup files in the save directory.
func (m *SaveManager) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(m.saveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".sav.bak") {
			backupName := strings.TrimSuffix(strings.TrimSuffix(entry.Name(), ".bak"), ".sav")
			backups = append(backups, backupName)
		}
	}

	return backups, nil
}

// GetBackupPath returns the path to the backup file for a save.
func (m *SaveManager) GetBackupPath(name string) string {
	return m.getFilePath(name) + ".bak"
}

// BackupExists checks if a backup file exists for a save.
func (m *SaveManager) BackupExists(name string) bool {
	if err := m.validateSaveName(name); err != nil {
		return false
	}

	backupPath := m.GetBackupPath(name)
	_, err := os.Stat(backupPath)
	return err == nil
}
