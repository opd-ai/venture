// Package saveload provides save/load manager for game state persistence.
// This file implements MemorySaveManager, an in-memory fallback implementation
// for when file system storage is unavailable.
// Available on all platforms (including WASM) as a fallback save backend.
package saveload

import (
	"sort"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/errors"
	"github.com/sirupsen/logrus"
)

// MemorySaveManager provides in-memory save storage when file-based
// storage is unavailable. Data is lost when the application exits.
// This is used as a fallback to prevent nil pointer dereferences when
// the normal SaveManager fails to initialize.
type MemorySaveManager struct {
	saves  map[string]*GameSave
	mu     sync.RWMutex
	logger *logrus.Entry
}

// NewMemorySaveManager creates a new in-memory save manager.
func NewMemorySaveManager() *MemorySaveManager {
	return NewMemorySaveManagerWithLogger(nil)
}

// NewMemorySaveManagerWithLogger creates a new in-memory save manager with logging.
func NewMemorySaveManagerWithLogger(logger *logrus.Logger) *MemorySaveManager {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"component": "saveload",
			"storage":   "memory",
		})
		logEntry.Warn("using in-memory save storage - saves will not persist after exit")
	}

	return &MemorySaveManager{
		saves:  make(map[string]*GameSave),
		logger: logEntry,
	}
}

// SaveGame saves the game state to memory.
func (m *MemorySaveManager) SaveGame(name string, save *GameSave) error {
	if save == nil {
		return errors.Validation("save cannot be nil")
	}

	if name == "" {
		return errors.Validation("save name cannot be empty")
	}

	save.Version = SaveVersion
	save.Timestamp = time.Now()

	m.mu.Lock()
	m.saves[name] = save
	m.mu.Unlock()

	if m.logger != nil {
		m.logger.WithFields(logrus.Fields{
			"name":      name,
			"timestamp": save.Timestamp,
		}).Info("game saved to memory")
	}

	return nil
}

// LoadGame loads the game state from memory.
func (m *MemorySaveManager) LoadGame(name string) (*GameSave, error) {
	if name == "" {
		return nil, errors.Validation("save name cannot be empty")
	}

	m.mu.RLock()
	save, exists := m.saves[name]
	m.mu.RUnlock()

	if !exists {
		return nil, errors.FileSystem("save not found").WithContext("name", name)
	}

	if m.logger != nil {
		m.logger.WithFields(logrus.Fields{
			"name":      name,
			"version":   save.Version,
			"timestamp": save.Timestamp,
		}).Info("game loaded from memory")
	}

	return save, nil
}

// DeleteSave deletes a save from memory.
func (m *MemorySaveManager) DeleteSave(name string) error {
	if name == "" {
		return errors.Validation("save name cannot be empty")
	}

	m.mu.Lock()
	_, exists := m.saves[name]
	if exists {
		delete(m.saves, name)
	}
	m.mu.Unlock()

	if !exists {
		return errors.FileSystem("save not found").WithContext("name", name)
	}

	if m.logger != nil {
		m.logger.WithField("name", name).Info("save deleted from memory")
	}

	return nil
}

// ListSaves returns metadata for all saves in memory.
func (m *MemorySaveManager) ListSaves() ([]*SaveMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var metadata []*SaveMetadata
	for name, save := range m.saves {
		meta := &SaveMetadata{
			Name:      name,
			Version:   save.Version,
			Timestamp: save.Timestamp,
			GameTime:  0,
		}

		if save.PlayerState != nil {
			meta.PlayerLevel = save.PlayerState.Level
		}
		if save.WorldState != nil {
			meta.GenreID = save.WorldState.GenreID
			meta.GameTime = save.WorldState.GameTime
		}

		metadata = append(metadata, meta)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Timestamp.After(metadata[j].Timestamp)
	})

	return metadata, nil
}

// GetSaveMetadata returns metadata for a specific save.
func (m *MemorySaveManager) GetSaveMetadata(name string) (*SaveMetadata, error) {
	if name == "" {
		return nil, errors.Validation("save name cannot be empty")
	}

	m.mu.RLock()
	save, exists := m.saves[name]
	m.mu.RUnlock()

	if !exists {
		return nil, errors.FileSystem("save not found").WithContext("name", name)
	}

	meta := &SaveMetadata{
		Name:      name,
		Version:   save.Version,
		Timestamp: save.Timestamp,
	}

	if save.PlayerState != nil {
		meta.PlayerLevel = save.PlayerState.Level
	}
	if save.WorldState != nil {
		meta.GenreID = save.WorldState.GenreID
		meta.GameTime = save.WorldState.GameTime
	}

	return meta, nil
}

// SaveExists checks if a save exists in memory.
func (m *MemorySaveManager) SaveExists(name string) bool {
	m.mu.RLock()
	_, exists := m.saves[name]
	m.mu.RUnlock()
	return exists
}

// SetMigrator is a no-op for memory manager since no migration is needed.
func (m *MemorySaveManager) SetMigrator(_ Migrator) {
	logrus.Warn("MemorySaveManager.SetMigrator: in-memory storage does not support migration, migrator ignored")
}
