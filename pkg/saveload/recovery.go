//go:build !js
// +build !js

package saveload

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/errors"
	"github.com/sirupsen/logrus"
)

// checksumFile computes SHA256 checksum of a file.
func checksumFile(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", errors.FileSystemWrap(err, "failed to open file for checksum").
			WithContext("filepath", filepath)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", errors.FileSystemWrap(err, "failed to compute checksum").
			WithContext("filepath", filepath)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// createBackup creates a backup copy of a save file.
// Returns the backup file path or error.
func (m *SaveManager) createBackup(name string) (backupPath string, err error) {
	sourcePath := m.getFilePath(name)

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// No file to backup, not an error
		return "", nil
	}

	// Create backup filename (.bak extension)
	backupPath = sourcePath + ".bak"

	// Copy file to backup
	source, err := os.Open(sourcePath)
	if err != nil {
		return "", errors.FileSystemWrap(err, "failed to open source file for backup").
			WithContext("sourcePath", sourcePath)
	}
	defer source.Close()

	backup, err := os.Create(backupPath)
	if err != nil {
		return "", errors.FileSystemWrap(err, "failed to create backup file").
			WithContext("backupPath", backupPath)
	}
	defer func() {
		if closeErr := backup.Close(); closeErr != nil && err == nil {
			err = errors.FileSystemWrap(closeErr, "failed to close backup file").
				WithContext("backupPath", backupPath)
		}
	}()

	if _, err := io.Copy(backup, source); err != nil {
		return "", errors.FileSystemWrap(err, "failed to copy to backup").
			WithContext("sourcePath", sourcePath).
			WithContext("backupPath", backupPath)
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
		return errors.FileSystemWrap(err, "failed to write checksum file").
			WithContext("checksumPath", checksumPath)
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
		return false, errors.FileSystemWrap(err, "failed to read backup").
			WithContext("backupPath", backupPath)
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
		return false, errors.FileSystemWrap(err, "failed to restore from backup").
			WithContext("savePath", savePath).
			WithContext("backupPath", backupPath)
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

	if err := m.validateSaveRequest(name, save); err != nil {
		return err
	}

	backupPath, err := m.attemptBackupCreation(name)
	if err != nil {
		m.logWarn("failed to create backup", err, logrus.Fields{"name": name})
	}

	if err := m.performSaveOperation(name, save); err != nil {
		m.restoreBackupIfPossible(name, backupPath, err)
		return err
	}

	m.saveChecksumBestEffort(name)
	m.logSaveSuccess(name, save)
	return nil
}

// validateSaveRequest validates the save name and data.
func (m *SaveManager) validateSaveRequest(name string, save *GameSave) error {
	if save == nil {
		return errors.Validation("save cannot be nil")
	}

	if err := m.validateSaveName(name); err != nil {
		m.logWarn("invalid save name", err, logrus.Fields{"name": name})
		return err
	}

	return nil
}

// attemptBackupCreation creates a backup file, returning its path.
func (m *SaveManager) attemptBackupCreation(name string) (string, error) {
	backupPath, err := m.createBackup(name)
	return backupPath, err
}

// performSaveOperation marshals and writes save data to disk.
func (m *SaveManager) performSaveOperation(name string, save *GameSave) error {
	save.Version = SaveVersion
	save.Timestamp = time.Now()

	data, err := m.marshalSave(save, name)
	if err != nil {
		return err
	}

	return m.writeSaveFile(name, data)
}

// restoreBackupIfPossible attempts to restore backup after failed save.
func (m *SaveManager) restoreBackupIfPossible(name, backupPath string, originalErr error) {
	if backupPath == "" {
		return
	}

	m.logWarn("save failed, restoring backup", originalErr, logrus.Fields{
		"name":   name,
		"backup": backupPath,
	})

	backupData, readErr := os.ReadFile(backupPath)
	if readErr != nil {
		return
	}

	if restoreErr := m.writeSaveFile(name, backupData); restoreErr != nil {
		m.logWarn("backup restoration also failed", restoreErr, logrus.Fields{
			"name":   name,
			"backup": backupPath,
		})
	}
}

// saveChecksumBestEffort saves checksum without failing on error.
func (m *SaveManager) saveChecksumBestEffort(name string) {
	if err := m.saveChecksum(name); err != nil {
		m.logWarn("failed to save checksum", err, logrus.Fields{"name": name})
	}
}

// logSaveSuccess logs successful save completion.
func (m *SaveManager) logSaveSuccess(name string, save *GameSave) {
	data, _ := m.marshalSave(save, name)
	m.logInfo("game saved successfully with backup", logrus.Fields{
		"name":      name,
		"size":      len(data),
		"timestamp": save.Timestamp,
	})
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
		return nil, errors.SerializationWrap(err, "failed to validate/migrate save").
			WithContext("name", name)
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
		return errors.FileSystem("save file not found").
			WithContext("name", name)
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
		return errors.FileSystemWrap(recErr, "failed to recover from backup").
			WithContext("name", name)
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
		return nil, errors.FileSystemWrap(recErr, "failed to recover from backup").
			WithContext("name", name)
	}
	if !recovered {
		return nil, errors.SerializationWrap(originalErr, "save corrupted and no valid backup available").
			WithContext("name", name)
	}

	data, err := m.readSaveFile(name)
	if err != nil {
		return nil, errors.FileSystemWrap(err, "failed to load after recovery").
			WithContext("name", name)
	}

	save, err := m.unmarshalSave(data, name)
	if err != nil {
		return nil, errors.SerializationWrap(err, "failed to parse after recovery").
			WithContext("name", name)
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
			return errors.FileSystemWrap(err, "failed to remove backup").
				WithContext("backupPath", backupPath)
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
			return errors.FileSystemWrap(err, "failed to remove checksum").
				WithContext("checksumPath", checksumPath)
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
		return nil, errors.FileSystemWrap(err, "failed to read save directory").
			WithContext("saveDir", m.saveDir)
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
