//go:build js
// +build js

// BUG FIX: Phase 7 - WASM has no save/load persistence (file I/O doesn't work in browsers)
// Resolution: Implemented localStorage backend for browser storage with 5MB fallback awareness
// Platform: WASM (all browsers)
// Priority: P1 (Gameplay Blocker - saves don't persist across sessions)

package saveload

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"syscall/js"
	"time"
)

const (
	// localStoragePrefix is prepended to all save keys to avoid conflicts
	localStoragePrefix = "venture_save_"
	// localStorageMetaKey stores metadata about available saves
	localStorageMetaKey = "venture_save_metadata"
	// maxLocalStorageSize is the typical localStorage limit (5MB)
	maxLocalStorageSize = 5 * 1024 * 1024
)

// SaveManager is the WASM implementation using localStorage.
// On WASM, SaveManager is type-aliased to the localStorage implementation.
// This allows the same API across desktop and WASM platforms.
type SaveManager struct {
	useInMemory  bool
	memoryStore  map[string]*GameSave
	localStorage js.Value
	// migrator is stored for API parity but not used on WASM (migration not supported)
	migrator Migrator
}

// NewSaveManager creates a save manager for WASM/browser environment.
// This overrides the desktop version and uses localStorage instead of files.
func NewSaveManager(saveDir string) (*SaveManager, error) {
	mgr := &SaveManager{
		memoryStore: make(map[string]*GameSave),
	}

	// Check if localStorage is available
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() || localStorage.IsNull() {
		// Fallback to in-memory storage (won't persist across page reloads)
		mgr.useInMemory = true
		js.Global().Get("console").Call("warn", "[Venture] localStorage unavailable, using in-memory storage (saves won't persist)")
		return mgr, nil
	}

	mgr.localStorage = localStorage

	// Test if localStorage is accessible (some browsers block in private mode)
	testKey := localStoragePrefix + "test"
	defer func() {
		if r := recover(); r != nil {
			mgr.useInMemory = true
			js.Global().Get("console").Call("warn", "[Venture] localStorage access denied, using in-memory storage")
		}
	}()

	mgr.localStorage.Call("setItem", testKey, "test")
	mgr.localStorage.Call("removeItem", testKey)

	js.Global().Get("console").Call("log", "[Venture] WASM save manager initialized with localStorage")
	return mgr, nil
}

// NewSaveManagerWithLogger creates a new save manager with a logger.
// On WASM, the logger parameter is ignored (browser console is used instead).
// This method exists for API parity with the desktop implementation.
func NewSaveManagerWithLogger(saveDir string, logger interface{}) (*SaveManager, error) {
	js.Global().Get("console").Call("log", "[Venture] WASM save manager: logger parameter ignored, using browser console")
	return NewSaveManager(saveDir)
}

// NewSaveManagerWithMigrator creates a new save manager with a logger and migrator.
// On WASM, migration is not supported. The migrator is stored for API parity but
// incompatible save versions will be rejected rather than migrated.
// This method exists for API parity with the desktop implementation.
func NewSaveManagerWithMigrator(saveDir string, logger interface{}, migrator Migrator) (*SaveManager, error) {
	mgr, err := NewSaveManager(saveDir)
	if err != nil {
		return nil, err
	}
	if migrator != nil {
		js.Global().Get("console").Call("warn", "[Venture] WASM save manager: migration not supported, migrator will be ignored")
	}
	mgr.migrator = migrator
	return mgr, nil
}

// SetMigrator sets the migrator for handling older save file versions.
// On WASM, migration is not supported. This method exists for API parity
// with the desktop implementation. The migrator is stored but not used;
// incompatible save versions will be rejected.
func (m *SaveManager) SetMigrator(migrator Migrator) {
	if migrator != nil {
		js.Global().Get("console").Call("warn", "[Venture] WASM save manager: migration not supported, migrator will be ignored")
	}
	m.migrator = migrator
}

// SaveGame saves the game state to browser localStorage.
// Compatible with SaveManager.SaveGame API.
func (m *SaveManager) SaveGame(name string, save *GameSave) error {
	if save == nil {
		return fmt.Errorf("save cannot be nil")
	}

	// Validate save name for security (path traversal prevention)
	if err := m.validateSaveName(name); err != nil {
		js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Invalid save name: %s - %v", name, err))
		return err
	}

	save.Version = SaveVersion
	save.Timestamp = time.Now()

	// Serialize save to JSON
	data, err := json.Marshal(save)
	if err != nil {
		return fmt.Errorf("failed to marshal save: %w", err)
	}

	// Check size limit (localStorage typically 5MB per origin)
	if len(data) > maxLocalStorageSize {
		return fmt.Errorf("save data too large: %d bytes exceeds localStorage limit of %d bytes", len(data), maxLocalStorageSize)
	}

	if m.useInMemory {
		// In-memory fallback
		m.memoryStore[name] = save
		js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Saved to memory: %s (%d bytes)", name, len(data)))
		return nil
	}

	// Save to localStorage
	key := localStoragePrefix + name
	defer func() {
		if r := recover(); r != nil {
			// Quota exceeded or other localStorage error
			js.Global().Get("console").Call("error", fmt.Sprintf("[Venture] localStorage.setItem failed: %v", r))
			// Fallback to in-memory
			m.useInMemory = true
			m.memoryStore[name] = save
		}
	}()

	m.localStorage.Call("setItem", key, string(data))

	// Update metadata
	m.updateMetadata(name, save)

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Saved to localStorage: %s (%d bytes)", name, len(data)))
	return nil
}

// LoadGame loads the game state from browser localStorage.
// Compatible with SaveManager.LoadGame API.
func (m *SaveManager) LoadGame(name string) (*GameSave, error) {
	// Validate save name for security (path traversal prevention)
	if err := m.validateSaveName(name); err != nil {
		js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Invalid save name: %s - %v", name, err))
		return nil, err
	}

	var save *GameSave

	if m.useInMemory {
		// Load from in-memory fallback
		s, ok := m.memoryStore[name]
		if !ok {
			return nil, fmt.Errorf("save not found: %s", name)
		}
		save = s
		js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Loaded from memory: %s", name))
	} else {
		// Load from localStorage
		key := localStoragePrefix + name
		dataJS := m.localStorage.Call("getItem", key)
		if dataJS.IsNull() {
			return nil, fmt.Errorf("save not found: %s", name)
		}

		data := dataJS.String()
		var s GameSave
		if err := json.Unmarshal([]byte(data), &s); err != nil {
			return nil, fmt.Errorf("failed to unmarshal save: %w", err)
		}
		save = &s
		js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Loaded from localStorage: %s", name))
	}

	// Validate and check version compatibility
	if err := m.validateSave(save); err != nil {
		js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Save validation failed: %s - %v", name, err))
		return nil, fmt.Errorf("failed to validate save: %w", err)
	}

	return save, nil
}

// DeleteSave removes a save from browser localStorage.
// Compatible with SaveManager.DeleteSave API.
func (m *SaveManager) DeleteSave(name string) error {
	if m.useInMemory {
		delete(m.memoryStore, name)
		js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Deleted from memory: %s", name))
		return nil
	}

	key := localStoragePrefix + name
	m.localStorage.Call("removeItem", key)

	// Update metadata
	m.removeFromMetadata(name)

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Deleted from localStorage: %s", name))
	return nil
}

// ListSaves returns a list of available saves with metadata.
// Compatible with SaveManager.ListSaves API (returns []*SaveMetadata).
func (m *SaveManager) ListSaves() ([]*SaveMetadata, error) {
	if m.useInMemory {
		// List from in-memory fallback
		var saves []*SaveMetadata
		for name, save := range m.memoryStore {
			saves = append(saves, &SaveMetadata{
				Name:      name,
				Timestamp: save.Timestamp,
				Version:   save.Version,
			})
		}

		// Sort by timestamp (newest first)
		sort.Slice(saves, func(i, j int) bool {
			return saves[i].Timestamp.After(saves[j].Timestamp)
		})

		return saves, nil
	}

	// Load metadata from localStorage
	metaJS := m.localStorage.Call("getItem", localStorageMetaKey)
	if metaJS.IsNull() {
		// No metadata yet, scan localStorage keys
		return m.scanLocalStorageKeys()
	}

	var saves []*SaveMetadata
	if err := json.Unmarshal([]byte(metaJS.String()), &saves); err != nil {
		// Corrupted metadata, rebuild by scanning
		return m.scanLocalStorageKeys()
	}

	// Sort by timestamp (newest first)
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].Timestamp.After(saves[j].Timestamp)
	})

	return saves, nil
}

// GetSaveMetadata reads metadata from a save without loading the entire save.
// Compatible with SaveManager.GetSaveMetadata API.
func (m *SaveManager) GetSaveMetadata(name string) (*SaveMetadata, error) {
	// For WASM, we need to load the entire save to get metadata
	// (localStorage doesn't support partial reads like file seeks)
	save, err := m.LoadGame(name)
	if err != nil {
		return nil, err
	}

	metadata := &SaveMetadata{
		Name:      name,
		Version:   save.Version,
		Timestamp: save.Timestamp,
	}

	if save.PlayerState != nil {
		metadata.PlayerLevel = save.PlayerState.Level
	}

	if save.WorldState != nil {
		metadata.GenreID = save.WorldState.GenreID
		metadata.GameTime = save.WorldState.GameTime
	}

	return metadata, nil
}

// updateMetadata updates the save metadata index in localStorage.
func (m *SaveManager) updateMetadata(name string, save *GameSave) {
	saves, _ := m.ListSaves()

	// Build metadata entry with nil-safe field access
	metadata := &SaveMetadata{
		Name:      name,
		Timestamp: save.Timestamp,
		Version:   save.Version,
	}
	if save.PlayerState != nil {
		metadata.PlayerLevel = save.PlayerState.Level
	}
	if save.WorldState != nil {
		metadata.GenreID = save.WorldState.GenreID
		metadata.GameTime = save.WorldState.GameTime
	}

	// Update or add metadata entry
	found := false
	for i, s := range saves {
		if s.Name == name {
			saves[i] = metadata
			found = true
			break
		}
	}

	if !found {
		saves = append(saves, metadata)
	}

	// Save updated metadata
	data, _ := json.Marshal(saves)
	m.localStorage.Call("setItem", localStorageMetaKey, string(data))
}

// removeFromMetadata removes a save from the metadata index.
func (m *SaveManager) removeFromMetadata(name string) {
	saves, _ := m.ListSaves()

	// Remove metadata entry
	var updated []*SaveMetadata
	for _, s := range saves {
		if s.Name != name {
			updated = append(updated, s)
		}
	}

	// Save updated metadata
	data, _ := json.Marshal(updated)
	m.localStorage.Call("setItem", localStorageMetaKey, string(data))
}

// scanLocalStorageKeys scans localStorage for venture save keys and rebuilds metadata.
func (m *SaveManager) scanLocalStorageKeys() ([]*SaveMetadata, error) {
	var saves []*SaveMetadata

	length := m.localStorage.Get("length").Int()
	for i := 0; i < length; i++ {
		key := m.localStorage.Call("key", i).String()
		if len(key) > len(localStoragePrefix) && key[:len(localStoragePrefix)] == localStoragePrefix {
			name := key[len(localStoragePrefix):]

			// Try to load save to get metadata
			if save, err := m.LoadGame(name); err == nil {
				saves = append(saves, &SaveMetadata{
					Name:      name,
					Timestamp: save.Timestamp,
					Version:   save.Version,
				})
			}
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].Timestamp.After(saves[j].Timestamp)
	})

	return saves, nil
}

// SaveExists checks if a save file exists.
// Compatible with SaveManager.SaveExists API.
func (m *SaveManager) SaveExists(name string) bool {
	if err := m.validateSaveName(name); err != nil {
		return false
	}

	if m.useInMemory {
		_, ok := m.memoryStore[name]
		return ok
	}

	key := localStoragePrefix + name
	dataJS := m.localStorage.Call("getItem", key)
	return !dataJS.IsNull()
}

// validateSaveName validates that a save name is acceptable.
// Delegates to the shared ValidateSaveName function for consistency.
func (m *SaveManager) validateSaveName(name string) error {
	return ValidateSaveName(name)
}

// validateSave validates a loaded save file for version compatibility and required fields.
// This mirrors the desktop validateAndMigrate() behavior.
// Note: WASM does not support migration, so incompatible versions are rejected.
func (m *SaveManager) validateSave(save *GameSave) error {
	if save == nil {
		return fmt.Errorf("save cannot be nil")
	}

	// Check version exists
	if save.Version == "" {
		return fmt.Errorf("save file has no version")
	}

	// Check version compatibility
	// WASM does not support migration, so only current version is accepted
	if save.Version != SaveVersion {
		return fmt.Errorf("save file version %s is not supported (current version: %s) - migration not available on WASM", save.Version, SaveVersion)
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

// SaveGameWithBackup saves the game state with automatic backup and checksum.
// This is the recommended method for production use on WASM.
// Backups are stored in localStorage with a .bak suffix on the key.
func (m *SaveManager) SaveGameWithBackup(name string, save *GameSave) error {
	if save == nil {
		return fmt.Errorf("save cannot be nil")
	}

	if err := m.validateSaveName(name); err != nil {
		js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Invalid save name: %s - %v", name, err))
		return err
	}

	// Create backup of existing save before overwriting
	if m.SaveExists(name) {
		if err := m.createBackup(name); err != nil {
			js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Failed to create backup: %v", err))
			// Continue with save anyway, backup failure shouldn't block saving
		}
	}

	// Perform the save
	save.Version = SaveVersion
	save.Timestamp = time.Now()

	data, err := json.Marshal(save)
	if err != nil {
		return fmt.Errorf("failed to marshal save: %w", err)
	}

	if len(data) > maxLocalStorageSize {
		return fmt.Errorf("save data too large: %d bytes exceeds localStorage limit of %d bytes", len(data), maxLocalStorageSize)
	}

	if m.useInMemory {
		m.memoryStore[name] = save
		m.memoryStore[name+".fnv1a"] = &GameSave{Version: m.computeChecksum(data)}
		js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Saved with backup to memory: %s (%d bytes)", name, len(data)))
		return nil
	}

	// Save to localStorage
	key := localStoragePrefix + name
	defer func() {
		if r := recover(); r != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("[Venture] localStorage.setItem failed: %v", r))
			m.useInMemory = true
			m.memoryStore[name] = save
		}
	}()

	m.localStorage.Call("setItem", key, string(data))

	// Save checksum
	checksum := m.computeChecksum(data)
	m.localStorage.Call("setItem", key+".fnv1a", checksum)

	// Update metadata
	m.updateMetadata(name, save)

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Saved with backup to localStorage: %s (%d bytes)", name, len(data)))
	return nil
}

// LoadGameWithRecovery loads the game state with corruption detection and recovery.
// This is the recommended method for production use on WASM.
func (m *SaveManager) LoadGameWithRecovery(name string) (*GameSave, error) {
	if err := m.validateSaveName(name); err != nil {
		js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Invalid save name: %s - %v", name, err))
		return nil, err
	}

	if !m.SaveExists(name) {
		return nil, fmt.Errorf("save file not found: %s", name)
	}

	if err := m.attemptChecksumRecovery(name); err != nil {
		return nil, err
	}

	save, err := m.loadWithFallback(name)
	if err != nil {
		return nil, err
	}

	return save, nil
}

// attemptChecksumRecovery validates checksum and attempts recovery if needed.
func (m *SaveManager) attemptChecksumRecovery(name string) error {
	valid, hasChecksum := m.validateChecksum(name)
	if !hasChecksum || valid {
		return nil
	}

	js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Checksum validation failed for %s, attempting recovery", name))
	recovered, err := m.recoverFromBackup(name)
	if err != nil {
		return fmt.Errorf("failed to recover from backup: %w", err)
	}
	if !recovered {
		js.Global().Get("console").Call("warn", "[Venture] Recovery failed, attempting to load corrupted file")
	}
	return nil
}

// loadWithFallback attempts to load a save, falling back to recovery if needed.
func (m *SaveManager) loadWithFallback(name string) (*GameSave, error) {
	save, err := m.LoadGame(name)
	if err == nil {
		return save, nil
	}

	js.Global().Get("console").Call("error", fmt.Sprintf("[Venture] Failed to load save, attempting recovery: %v", err))

	if err := m.performRecoveryAndRetry(name, err); err != nil {
		return nil, err
	}

	save, retryErr := m.LoadGame(name)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to load after recovery: %w", retryErr)
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Successfully recovered save: %s", name))
	return save, nil
}

// performRecoveryAndRetry attempts to recover from backup.
func (m *SaveManager) performRecoveryAndRetry(name string, originalErr error) error {
	recovered, recErr := m.recoverFromBackup(name)
	if recErr != nil {
		return fmt.Errorf("failed to recover from backup: %w", recErr)
	}
	if !recovered {
		return fmt.Errorf("save corrupted and no valid backup available: %w", originalErr)
	}
	return nil
}

// BackupExists checks if a backup exists for a save.
func (m *SaveManager) BackupExists(name string) bool {
	if err := m.validateSaveName(name); err != nil {
		return false
	}

	if m.useInMemory {
		_, ok := m.memoryStore[name+".bak"]
		return ok
	}

	key := localStoragePrefix + name + ".bak"
	dataJS := m.localStorage.Call("getItem", key)
	return !dataJS.IsNull()
}

// GetBackupPath returns a string identifier for the backup.
// On WASM, this returns the localStorage key (not a file path).
func (m *SaveManager) GetBackupPath(name string) string {
	return localStoragePrefix + name + ".bak"
}

// ListBackups returns names of all saves that have backups.
func (m *SaveManager) ListBackups() ([]string, error) {
	var backups []string

	if m.useInMemory {
		for key := range m.memoryStore {
			if strings.HasSuffix(key, ".bak") {
				backups = append(backups, strings.TrimSuffix(key, ".bak"))
			}
		}
		return backups, nil
	}

	length := m.localStorage.Get("length").Int()
	for i := 0; i < length; i++ {
		key := m.localStorage.Call("key", i).String()
		if strings.HasPrefix(key, localStoragePrefix) && strings.HasSuffix(key, ".bak") {
			// Extract save name from key
			name := key[len(localStoragePrefix) : len(key)-4] // Remove prefix and .bak suffix
			backups = append(backups, name)
		}
	}

	return backups, nil
}

// CleanupBackups removes backup and checksum files for a save.
func (m *SaveManager) CleanupBackups(name string) error {
	if err := m.validateSaveName(name); err != nil {
		return err
	}

	if m.useInMemory {
		delete(m.memoryStore, name+".bak")
		delete(m.memoryStore, name+".fnv1a")
		js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Cleaned up backups from memory: %s", name))
		return nil
	}

	// Remove backup
	backupKey := localStoragePrefix + name + ".bak"
	m.localStorage.Call("removeItem", backupKey)

	// Remove checksum
	checksumKey := localStoragePrefix + name + ".fnv1a"
	m.localStorage.Call("removeItem", checksumKey)

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Cleaned up backups from localStorage: %s", name))
	return nil
}

// createBackup creates a backup of an existing save in localStorage.
func (m *SaveManager) createBackup(name string) error {
	if m.useInMemory {
		if save, ok := m.memoryStore[name]; ok {
			m.memoryStore[name+".bak"] = save
		}
		return nil
	}

	key := localStoragePrefix + name
	dataJS := m.localStorage.Call("getItem", key)
	if dataJS.IsNull() {
		return nil // No save to backup
	}

	// Copy save data to backup key
	backupKey := key + ".bak"
	m.localStorage.Call("setItem", backupKey, dataJS.String())

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Created backup: %s", name))
	return nil
}

// validateChecksum validates a save's checksum.
// Returns (valid, hasChecksum) where hasChecksum indicates if a checksum exists.
func (m *SaveManager) validateChecksum(name string) (bool, bool) {
	var storedChecksum, currentData string

	if m.useInMemory {
		checksumSave, ok := m.memoryStore[name+".fnv1a"]
		if !ok {
			return false, false // No checksum
		}
		storedChecksum = checksumSave.Version

		save, ok := m.memoryStore[name]
		if !ok {
			return false, true // Has checksum but no save
		}
		data, err := json.Marshal(save)
		if err != nil {
			return false, true
		}
		currentData = string(data)
	} else {
		checksumKey := localStoragePrefix + name + ".fnv1a"
		checksumJS := m.localStorage.Call("getItem", checksumKey)
		if checksumJS.IsNull() {
			return false, false // No checksum
		}
		storedChecksum = checksumJS.String()

		key := localStoragePrefix + name
		dataJS := m.localStorage.Call("getItem", key)
		if dataJS.IsNull() {
			return false, true // Has checksum but no save
		}
		currentData = dataJS.String()
	}

	currentChecksum := m.computeChecksum([]byte(currentData))
	return storedChecksum == currentChecksum, true
}

// recoverFromBackup attempts to recover a corrupted save from its backup.
// Returns true if recovery was successful.
func (m *SaveManager) recoverFromBackup(name string) (bool, error) {
	if !m.BackupExists(name) {
		js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] No backup found for recovery: %s", name))
		return false, nil
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Attempting recovery from backup: %s", name))

	if m.useInMemory {
		backup, ok := m.memoryStore[name+".bak"]
		if !ok {
			return false, nil
		}

		// Validate backup
		if err := m.validateSave(backup); err != nil {
			js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Backup is also corrupted: %s - %v", name, err))
			return false, nil
		}

		// Restore from backup
		m.memoryStore[name] = backup

		// Update checksum
		data, _ := json.Marshal(backup)
		m.memoryStore[name+".fnv1a"] = &GameSave{Version: m.computeChecksum(data)}

		js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Successfully recovered from backup: %s", name))
		return true, nil
	}

	// Load backup data
	backupKey := localStoragePrefix + name + ".bak"
	backupDataJS := m.localStorage.Call("getItem", backupKey)
	if backupDataJS.IsNull() {
		return false, nil
	}
	backupData := backupDataJS.String()

	// Validate backup by parsing it
	var backup GameSave
	if err := json.Unmarshal([]byte(backupData), &backup); err != nil {
		js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Backup is corrupted (parse error): %s - %v", name, err))
		return false, nil
	}

	if err := m.validateSave(&backup); err != nil {
		js.Global().Get("console").Call("warn", fmt.Sprintf("[Venture] Backup is corrupted (validation error): %s - %v", name, err))
		return false, nil
	}

	// Restore from backup
	saveKey := localStoragePrefix + name
	m.localStorage.Call("setItem", saveKey, backupData)

	// Update checksum
	checksum := m.computeChecksum([]byte(backupData))
	m.localStorage.Call("setItem", saveKey+".fnv1a", checksum)

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Successfully recovered from backup: %s", name))
	return true, nil
}

// computeChecksum computes a simple checksum for data using FNV-1a.
// Uses FNV-1a for simplicity and detection of accidental corruption.
// Note: This is not cryptographically secure; it is intended for single-player WASM saves only.
func (m *SaveManager) computeChecksum(data []byte) string {
	// Use FNV-1a hash for simplicity and WASM compatibility
	var hash uint64 = 14695981039346656037 // FNV offset basis
	for _, b := range data {
		hash ^= uint64(b)
		hash *= 1099511628211 // FNV prime
	}
	return fmt.Sprintf("%016x", hash)
}
