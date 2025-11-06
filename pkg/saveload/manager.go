// Package saveload provides save/load manager for game state persistence.
// This file implements SaveLoadManager which handles saving and loading
// game state to/from disk using JSON serialization.
package saveload

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SaveManager handles save/load operations for game state.
type SaveManager struct {
	// Directory where save files are stored
	saveDir string
	// Logger for save/load operations
	logger *logrus.Entry
}

// NewSaveManager creates a new save manager with the specified save directory.
// The directory will be created if it doesn't exist.
func NewSaveManager(saveDir string) (*SaveManager, error) {
	return NewSaveManagerWithLogger(saveDir, nil)
}

// NewSaveManagerWithLogger creates a new save manager with a logger.
func NewSaveManagerWithLogger(saveDir string, logger *logrus.Logger) (*SaveManager, error) {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"component": "saveload",
			"saveDir":   saveDir,
		})
	}

	// Create save directory if it doesn't exist
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		if logEntry != nil {
			logEntry.WithError(err).Error("failed to create save directory")
		}
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}

	if logEntry != nil {
		logEntry.Info("save manager initialized")
	}

	return &SaveManager{
		saveDir: saveDir,
		logger:  logEntry,
	}, nil
}

// SaveGame saves the game state to a file with the specified name.
// The .sav extension is added automatically if not present.
func (m *SaveManager) SaveGame(name string, save *GameSave) error {
	m.logDebug("saving game", logrus.Fields{"name": name})

	if save == nil {
		return fmt.Errorf("save cannot be nil")
	}

	if err := m.validateSaveName(name); err != nil {
		m.logWarn("invalid save name", err, logrus.Fields{"name": name})
		return err
	}

	save.Version = SaveVersion
	save.Timestamp = time.Now()

	data, err := m.marshalSave(save, name)
	if err != nil {
		return err
	}

	if err := m.writeSaveFile(name, data); err != nil {
		return err
	}

	m.logInfo("game saved successfully", logrus.Fields{
		"name":      name,
		"size":      len(data),
		"timestamp": save.Timestamp,
	})

	return nil
}

// LoadGame loads the game state from a file with the specified name.
func (m *SaveManager) LoadGame(name string) (*GameSave, error) {
	m.logDebug("loading game", logrus.Fields{"name": name})

	if err := m.validateSaveName(name); err != nil {
		m.logWarn("invalid save name", err, logrus.Fields{"name": name})
		return nil, err
	}

	data, err := m.readSaveFile(name)
	if err != nil {
		return nil, err
	}

	save, err := m.unmarshalSave(data, name)
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

// DeleteSave deletes a save file.
func (m *SaveManager) DeleteSave(name string) error {
	// Validate save name
	if err := m.validateSaveName(name); err != nil {
		return err
	}

	filename := m.getFilePath(name)
	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("save file not found: %s", name)
		}
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	return nil
}

// ListSaves returns metadata for all save files in the save directory.
func (m *SaveManager) ListSaves() ([]*SaveMetadata, error) {
	// Read directory
	entries, err := os.ReadDir(m.saveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}

	// Collect metadata for each save file
	var saves []*SaveMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .sav files
		if !strings.HasSuffix(entry.Name(), ".sav") {
			continue
		}

		// Get metadata
		metadata, err := m.GetSaveMetadata(strings.TrimSuffix(entry.Name(), ".sav"))
		if err != nil {
			// Skip files that can't be read
			continue
		}

		saves = append(saves, metadata)
	}

	// Sort by timestamp (newest first)
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].Timestamp.After(saves[j].Timestamp)
	})

	return saves, nil
}

// GetSaveMetadata reads metadata from a save file without loading the entire save.
func (m *SaveManager) GetSaveMetadata(name string) (*SaveMetadata, error) {
	// Validate save name
	if err := m.validateSaveName(name); err != nil {
		return nil, err
	}

	filename := m.getFilePath(name)

	// Get file info
	fileInfo, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("save file not found: %s", name)
		}
		return nil, fmt.Errorf("failed to stat save file: %w", err)
	}

	// Open and read file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open save file: %w", err)
	}
	defer file.Close()

	// Decode just enough to get metadata
	var save GameSave
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&save); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty save file: %s", name)
		}
		return nil, fmt.Errorf("failed to decode save file: %w", err)
	}

	// Build metadata
	metadata := &SaveMetadata{
		Name:      name,
		Version:   save.Version,
		Timestamp: save.Timestamp,
		FileSize:  fileInfo.Size(),
	}

	// Add player and world info if available
	if save.PlayerState != nil {
		metadata.PlayerLevel = save.PlayerState.Level
	}
	if save.WorldState != nil {
		metadata.GenreID = save.WorldState.GenreID
		metadata.GameTime = save.WorldState.GameTime
	}

	return metadata, nil
}

// SaveExists checks if a save file exists.
func (m *SaveManager) SaveExists(name string) bool {
	if err := m.validateSaveName(name); err != nil {
		return false
	}

	filename := m.getFilePath(name)
	_, err := os.Stat(filename)
	return err == nil
}

// getFilePath returns the full path to a save file.
func (m *SaveManager) getFilePath(name string) string {
	// Add .sav extension if not present
	if !strings.HasSuffix(name, ".sav") {
		name += ".sav"
	}
	return filepath.Join(m.saveDir, name)
}

// validateSaveName validates that a save name is acceptable.
func (m *SaveManager) validateSaveName(name string) error {
	if name == "" {
		return fmt.Errorf("save name cannot be empty")
	}

	// Remove extension for validation
	name = strings.TrimSuffix(name, ".sav")

	// Check for path separators (security check)
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("save name cannot contain path separators")
	}

	// Check for special characters
	if strings.ContainsAny(name, "<>:\"|?*") {
		return fmt.Errorf("save name contains invalid characters")
	}

	return nil
}

// validateAndMigrate validates a save file and migrates it if necessary.
func (m *SaveManager) validateAndMigrate(save *GameSave) error {
	if save == nil {
		return fmt.Errorf("save cannot be nil")
	}

	// Check version
	if save.Version == "" {
		return fmt.Errorf("save file has no version")
	}

	// For now, we only support the current version
	// Future versions can add migration logic here
	if save.Version != SaveVersion {
		// Example migration logic (not needed yet):
		// if save.Version == "0.9.0" {
		//     migrateSaveFrom090To100(save)
		// }

		// For now, just warn about version mismatch
		// In production, you might want to reject or migrate
		return fmt.Errorf("save file version %s is not supported (current version: %s)", save.Version, SaveVersion)
	}

	// Validate required fields
	if save.PlayerState == nil {
		return fmt.Errorf("save file missing player state")
	}
	if save.WorldState == nil {
		return fmt.Errorf("save file missing world state")
	}
	if save.Settings == nil {
		return fmt.Errorf("save file missing settings")
	}

	return nil
}

// marshalSave marshals save data to JSON.
func (m *SaveManager) marshalSave(save *GameSave, name string) ([]byte, error) {
	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		m.logError("failed to marshal save data", err, logrus.Fields{"name": name})
		return nil, fmt.Errorf("failed to marshal save data: %w", err)
	}
	return data, nil
}

// writeSaveFile writes save data to file.
func (m *SaveManager) writeSaveFile(name string, data []byte) error {
	filename := m.getFilePath(name)
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		m.logError("failed to write save file", err, logrus.Fields{
			"name":     name,
			"filename": filename,
		})
		return fmt.Errorf("failed to write save file: %w", err)
	}
	return nil
}

// readSaveFile reads save data from file.
func (m *SaveManager) readSaveFile(name string) ([]byte, error) {
	filename := m.getFilePath(name)
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			m.logWarn("save file not found", err, logrus.Fields{"name": name})
			return nil, fmt.Errorf("save file not found: %s", name)
		}
		m.logError("failed to read save file", err, logrus.Fields{"name": name})
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}
	return data, nil
}

// unmarshalSave unmarshals save data from JSON.
func (m *SaveManager) unmarshalSave(data []byte, name string) (*GameSave, error) {
	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		m.logError("failed to parse save file", err, logrus.Fields{"name": name})
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}
	return &save, nil
}

// logDebug logs a debug message if logger and level are configured.
func (m *SaveManager) logDebug(msg string, fields logrus.Fields) {
	if m.logger != nil && m.logger.Logger.GetLevel() >= logrus.DebugLevel {
		m.logger.WithFields(fields).Debug(msg)
	}
}

// logInfo logs an info message if logger is configured.
func (m *SaveManager) logInfo(msg string, fields logrus.Fields) {
	if m.logger != nil {
		m.logger.WithFields(fields).Info(msg)
	}
}

// logWarn logs a warning message if logger is configured.
func (m *SaveManager) logWarn(msg string, err error, fields logrus.Fields) {
	if m.logger != nil {
		m.logger.WithError(err).WithFields(fields).Warn(msg)
	}
}

// logError logs an error message if logger is configured.
func (m *SaveManager) logError(msg string, err error, fields logrus.Fields) {
	if m.logger != nil {
		m.logger.WithError(err).WithFields(fields).Error(msg)
	}
}
