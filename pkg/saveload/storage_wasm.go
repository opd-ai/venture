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

// SaveGame saves the game state to browser localStorage.
// Compatible with SaveManager.SaveGame API.
func (m *SaveManager) SaveGame(name string, save *GameSave) error {
	if save == nil {
		return fmt.Errorf("save cannot be nil")
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
	if m.useInMemory {
		// Load from in-memory fallback
		save, ok := m.memoryStore[name]
		if !ok {
			return nil, fmt.Errorf("save not found: %s", name)
		}
		js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Loaded from memory: %s", name))
		return save, nil
	}

	// Load from localStorage
	key := localStoragePrefix + name
	dataJS := m.localStorage.Call("getItem", key)
	if dataJS.IsNull() {
		return nil, fmt.Errorf("save not found: %s", name)
	}

	data := dataJS.String()
	var save GameSave
	if err := json.Unmarshal([]byte(data), &save); err != nil {
		return nil, fmt.Errorf("failed to unmarshal save: %w", err)
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("[Venture] Loaded from localStorage: %s", name))
	return &save, nil
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
// This mirrors the desktop implementation for security.
func (m *SaveManager) validateSaveName(name string) error {
	if name == "" {
		return fmt.Errorf("save name cannot be empty")
	}

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
