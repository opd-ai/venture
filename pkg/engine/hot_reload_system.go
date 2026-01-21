// Package engine provides the hot reload system for live mod updates.
// HotReloadSystem monitors mod files for changes and applies updates without restart.
package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// HotReloadSystem manages live mod reloading.
type HotReloadSystem struct {
	world *World

	// File watching
	fileWatcher FileWatcher
	lastCheck   time.Time
	checkMutex  sync.Mutex

	// Callbacks for mod operations
	reloadCallback   ModReloadCallback
	rollbackCallback ModRollbackCallback
	hashCallback     ModHashCallback

	// State migration
	migrationHandler StateMigrationHandler

	mu sync.RWMutex
}

// ModReloadCallback is called to reload a mod with new data.
type ModReloadCallback func(modID string, modData []byte) error

// ModRollbackCallback is called to rollback a mod to previous state.
type ModRollbackCallback func(modID string, state *ModState) error

// ModHashCallback is called to compute hash for mod files.
type ModHashCallback func(modID string) (string, error)

// NewHotReloadSystem creates a new hot reload system.
func NewHotReloadSystem(world *World) *HotReloadSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "hot_reload",
	}).Debug("Creating hot reload system")

	return &HotReloadSystem{
		world:     world,
		lastCheck: time.Now(),
	}
}

// SetFileWatcher sets the file watcher implementation.
func (s *HotReloadSystem) SetFileWatcher(watcher FileWatcher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fileWatcher = watcher
}

// SetReloadCallback sets the callback for mod reloading.
func (s *HotReloadSystem) SetReloadCallback(callback ModReloadCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reloadCallback = callback
}

// SetRollbackCallback sets the callback for mod rollback.
func (s *HotReloadSystem) SetRollbackCallback(callback ModRollbackCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rollbackCallback = callback
}

// SetHashCallback sets the callback for computing mod hashes.
func (s *HotReloadSystem) SetHashCallback(callback ModHashCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hashCallback = callback
}

// SetMigrationHandler sets the state migration handler.
func (s *HotReloadSystem) SetMigrationHandler(handler StateMigrationHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.migrationHandler = handler
}

// Update checks for mod file changes and processes pending reloads.
func (s *HotReloadSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("hot_reload") {
			continue
		}

		comp, ok := entity.GetComponent("hot_reload")
		if !ok || comp == nil {
			continue
		}
		reloadComp := comp.(*HotReloadComponent)

		if !reloadComp.IsEnabled() {
			continue
		}

		// Check for file changes
		s.checkForChanges(reloadComp)

		// Process auto-reloads if enabled
		if reloadComp.IsAutoReloadEnabled() && reloadComp.HasPendingUpdates() {
			s.processAutoReloads(reloadComp)
		}
	}
}

// checkForChanges checks watched mods for file changes.
func (s *HotReloadSystem) checkForChanges(comp *HotReloadComponent) {
	s.checkMutex.Lock()
	defer s.checkMutex.Unlock()

	// Throttle checks based on watch interval
	interval := comp.GetWatchInterval()
	if time.Since(s.lastCheck) < interval {
		return
	}
	s.lastCheck = time.Now()

	s.mu.RLock()
	watcher := s.fileWatcher
	hashCb := s.hashCallback
	s.mu.RUnlock()

	if watcher == nil && hashCb == nil {
		return
	}

	// Check each watched mod
	for _, modID := range comp.GetWatchedModIDs() {
		var newHash string
		var err error

		if hashCb != nil {
			newHash, err = hashCb(modID)
		} else if watcher != nil {
			newHash, err = watcher.GetFileHash(modID)
		}

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"mod_id": modID,
			}).WithError(err).Debug("hot reload: failed to get file hash")
			continue
		}

		if comp.DetectChange(modID, newHash) {
			hashDisplay := newHash
			if len(hashDisplay) > 8 {
				hashDisplay = newHash[:8]
			}
			logrus.WithFields(logrus.Fields{
				"mod_id":   modID,
				"new_hash": hashDisplay,
			}).Info("hot reload: detected change in mod")
		}
	}
}

// processAutoReloads automatically reloads mods with pending changes.
func (s *HotReloadSystem) processAutoReloads(comp *HotReloadComponent) {
	pending := comp.GetPendingUpdates()

	for _, modID := range pending {
		err := s.ReloadMod(comp, modID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"mod_id": modID,
			}).WithError(err).Error("hot reload: auto-reload failed")
		}
	}
}

// ReloadMod reloads a mod with updated files.
func (s *HotReloadSystem) ReloadMod(comp *HotReloadComponent, modID string) error {
	if err := s.validateReloadRequest(comp, modID); err != nil {
		return err
	}

	startTime := time.Now()
	oldVersion := s.getOldModVersion(comp, modID)
	s.logReloadStart(modID, oldVersion)

	s.saveStateForRollback(comp, modID, oldVersion)

	watcher, reloadCb, err := s.getReloadDependencies()
	if err != nil {
		return s.recordReloadFailure(comp, modID, oldVersion, startTime, err.Error())
	}

	modData, newVersion, newHash, err := s.fetchModData(watcher, modID)
	if err != nil {
		return s.recordReloadFailure(comp, modID, oldVersion, startTime, err.Error())
	}

	if err := s.performReload(comp, modID, oldVersion, startTime, reloadCb, modData); err != nil {
		return err
	}

	s.restoreModState(comp, modID)
	s.recordReloadSuccess(comp, modID, oldVersion, newVersion, newHash, startTime)
	s.logReloadSuccess(modID, newVersion, time.Since(startTime).Milliseconds())

	return nil
}

// validateReloadRequest checks if the reload request is valid.
func (s *HotReloadSystem) validateReloadRequest(comp *HotReloadComponent, modID string) error {
	if comp == nil {
		return fmt.Errorf("hot reload component is nil")
	}
	if !comp.IsWatching(modID) {
		return fmt.Errorf("mod %s is not being watched", modID)
	}
	return nil
}

// getOldModVersion retrieves the current version of the mod.
func (s *HotReloadSystem) getOldModVersion(comp *HotReloadComponent, modID string) string {
	oldMod, _ := comp.GetWatchedMod(modID)
	if oldMod != nil {
		return oldMod.Version
	}
	return ""
}

// logReloadStart logs the beginning of mod reload.
func (s *HotReloadSystem) logReloadStart(modID, oldVersion string) {
	logrus.WithFields(logrus.Fields{
		"mod_id":      modID,
		"old_version": oldVersion,
	}).Info("hot reload: starting mod reload")
}

// saveStateForRollback saves current mod state for potential rollback.
func (s *HotReloadSystem) saveStateForRollback(comp *HotReloadComponent, modID, oldVersion string) {
	s.mu.RLock()
	migrationHandler := s.migrationHandler
	s.mu.RUnlock()

	if migrationHandler != nil {
		scripts, variables, err := migrationHandler.SaveState(modID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"mod_id": modID,
			}).WithError(err).Warn("hot reload: failed to save state for rollback")
		} else {
			comp.SaveStateForRollback(modID, oldVersion, scripts, variables)
		}
	}
}

// getReloadDependencies retrieves file watcher and reload callback.
func (s *HotReloadSystem) getReloadDependencies() (FileWatcher, ModReloadCallback, error) {
	s.mu.RLock()
	watcher := s.fileWatcher
	reloadCb := s.reloadCallback
	s.mu.RUnlock()

	if watcher == nil || reloadCb == nil {
		return nil, nil, fmt.Errorf("no file watcher or reload callback configured")
	}
	return watcher, reloadCb, nil
}

// fetchModData retrieves new mod data, version, and hash.
func (s *HotReloadSystem) fetchModData(watcher FileWatcher, modID string) ([]byte, string, string, error) {
	modData, err := watcher.GetModData(modID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get mod data: %v", err)
	}

	newVersion, err := watcher.GetModVersion(modID)
	if err != nil {
		newVersion = "unknown"
	}

	newHash, _ := watcher.GetFileHash(modID)
	return modData, newVersion, newHash, nil
}

// performReload executes the reload and handles rollback on failure.
func (s *HotReloadSystem) performReload(comp *HotReloadComponent, modID, oldVersion string, startTime time.Time, reloadCb ModReloadCallback, modData []byte) error {
	err := reloadCb(modID, modData)
	if err != nil {
		rollbackErr := s.RollbackMod(comp, modID)
		if rollbackErr != nil {
			logrus.WithFields(logrus.Fields{
				"mod_id": modID,
			}).WithError(rollbackErr).Error("hot reload: rollback also failed")
		}
		return s.recordReloadFailure(comp, modID, oldVersion, startTime, fmt.Sprintf("reload failed: %v", err))
	}
	return nil
}

// restoreModState restores saved state after successful reload.
func (s *HotReloadSystem) restoreModState(comp *HotReloadComponent, modID string) {
	s.mu.RLock()
	migrationHandler := s.migrationHandler
	s.mu.RUnlock()

	if migrationHandler != nil {
		if state, exists := comp.GetRollbackState(modID); exists {
			err := migrationHandler.RestoreState(modID, state.Scripts, state.Variables)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"mod_id": modID,
				}).WithError(err).Warn("hot reload: failed to restore state after reload")
			}
		}
	}
}

// recordReloadSuccess records successful reload with updated metadata.
func (s *HotReloadSystem) recordReloadSuccess(comp *HotReloadComponent, modID, oldVersion, newVersion, newHash string, startTime time.Time) {
	duration := time.Since(startTime).Milliseconds()
	comp.AddReloadEntry(ReloadEntry{
		ModID:       modID,
		Timestamp:   time.Now().Unix(),
		OldVersion:  oldVersion,
		NewVersion:  newVersion,
		Success:     true,
		Duration:    duration,
		AutoReload:  comp.IsAutoReloadEnabled(),
		StateChange: "reloaded",
	})

	comp.UpdateModVersion(modID, newVersion, newHash)
	comp.ClearPendingUpdate(modID)
	comp.ClearRollbackState(modID)
}

// logReloadSuccess logs successful mod reload completion.
func (s *HotReloadSystem) logReloadSuccess(modID, newVersion string, durationMs int64) {
	logrus.WithFields(logrus.Fields{
		"mod_id":      modID,
		"new_version": newVersion,
		"duration_ms": durationMs,
	}).Info("hot reload: mod reloaded successfully")
}

// recordReloadFailure records a failed reload attempt.
func (s *HotReloadSystem) recordReloadFailure(comp *HotReloadComponent, modID, oldVersion string, startTime time.Time, errMsg string) error {
	duration := time.Since(startTime).Milliseconds()

	comp.AddReloadEntry(ReloadEntry{
		ModID:       modID,
		Timestamp:   time.Now().Unix(),
		OldVersion:  oldVersion,
		NewVersion:  oldVersion,
		Success:     false,
		Error:       errMsg,
		Duration:    duration,
		AutoReload:  comp.IsAutoReloadEnabled(),
		StateChange: "failed",
	})

	return fmt.Errorf("hot reload failed for mod %s: %s", modID, errMsg)
}

// RollbackMod rolls back a mod to its previous state.
func (s *HotReloadSystem) RollbackMod(comp *HotReloadComponent, modID string) error {
	if comp == nil {
		return fmt.Errorf("hot reload component is nil")
	}

	state, exists := comp.GetRollbackState(modID)
	if !exists {
		return fmt.Errorf("no rollback state available for mod %s", modID)
	}

	s.mu.RLock()
	rollbackCb := s.rollbackCallback
	s.mu.RUnlock()

	if rollbackCb == nil {
		return fmt.Errorf("no rollback callback configured")
	}

	logrus.WithFields(logrus.Fields{
		"mod_id":  modID,
		"version": state.Version,
	}).Info("hot reload: rolling back mod")

	err := rollbackCb(modID, state)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Record rollback
	comp.AddReloadEntry(ReloadEntry{
		ModID:       modID,
		Timestamp:   time.Now().Unix(),
		OldVersion:  "",
		NewVersion:  state.Version,
		Success:     true,
		RolledBack:  true,
		StateChange: "rolled_back",
	})

	comp.ClearRollbackState(modID)

	logrus.WithFields(logrus.Fields{
		"mod_id":  modID,
		"version": state.Version,
	}).Info("hot reload: rollback completed")

	return nil
}

// StartWatchingMod begins monitoring a mod for changes.
func (s *HotReloadSystem) StartWatchingMod(comp *HotReloadComponent, modID string) error {
	if comp == nil {
		return fmt.Errorf("hot reload component is nil")
	}

	if comp.IsWatching(modID) {
		return nil // Already watching
	}

	s.mu.RLock()
	watcher := s.fileWatcher
	hashCb := s.hashCallback
	s.mu.RUnlock()

	var fileHash string
	var version string
	var err error

	if hashCb != nil {
		fileHash, err = hashCb(modID)
		if err != nil {
			fileHash = ""
		}
	} else if watcher != nil {
		fileHash, err = watcher.GetFileHash(modID)
		if err != nil {
			fileHash = ""
		}
		version, _ = watcher.GetModVersion(modID)
	}

	comp.StartWatching(modID, version, fileHash)

	logrus.WithFields(logrus.Fields{
		"mod_id":  modID,
		"version": version,
	}).Debug("hot reload: started watching mod")

	return nil
}

// StopWatchingMod stops monitoring a mod for changes.
func (s *HotReloadSystem) StopWatchingMod(comp *HotReloadComponent, modID string) {
	if comp == nil {
		return
	}

	comp.StopWatching(modID)
	comp.ClearRollbackState(modID)

	logrus.WithFields(logrus.Fields{
		"mod_id": modID,
	}).Debug("hot reload: stopped watching mod")
}

// ForceReloadAll reloads all watched mods.
func (s *HotReloadSystem) ForceReloadAll(comp *HotReloadComponent) []error {
	if comp == nil {
		return []error{fmt.Errorf("hot reload component is nil")}
	}

	var errors []error
	for _, modID := range comp.GetWatchedModIDs() {
		if err := s.ReloadMod(comp, modID); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// GetReloadStatistics returns reload statistics.
func (s *HotReloadSystem) GetReloadStatistics(comp *HotReloadComponent) (total, successful, failed int) {
	if comp == nil {
		return 0, 0, 0
	}

	total = comp.GetTotalReloads()
	successful = comp.GetSuccessfulReloads()
	failed = comp.GetFailedReloads()
	return total, successful, failed
}

// InMemoryFileWatcher provides a simple in-memory file watcher for testing.
type InMemoryFileWatcher struct {
	mods map[string]*inMemoryMod
	mu   sync.RWMutex
}

type inMemoryMod struct {
	data    []byte
	version string
	hash    string
}

// NewInMemoryFileWatcher creates a new in-memory file watcher.
func NewInMemoryFileWatcher() *InMemoryFileWatcher {
	return &InMemoryFileWatcher{
		mods: make(map[string]*inMemoryMod),
	}
}

// AddMod adds a mod to the watcher.
func (w *InMemoryFileWatcher) AddMod(modID string, data []byte, version string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	hash := ComputeHash(data)
	w.mods[modID] = &inMemoryMod{
		data:    data,
		version: version,
		hash:    hash,
	}
}

// UpdateMod updates a mod's data (simulates file change).
func (w *InMemoryFileWatcher) UpdateMod(modID string, data []byte, version string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	hash := ComputeHash(data)
	if mod, exists := w.mods[modID]; exists {
		mod.data = data
		mod.version = version
		mod.hash = hash
	}
}

// GetFileHash implements FileWatcher.
func (w *InMemoryFileWatcher) GetFileHash(modID string) (string, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	mod, exists := w.mods[modID]
	if !exists {
		return "", fmt.Errorf("mod %s not found", modID)
	}

	return mod.hash, nil
}

// GetModData implements FileWatcher.
func (w *InMemoryFileWatcher) GetModData(modID string) ([]byte, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	mod, exists := w.mods[modID]
	if !exists {
		return nil, fmt.Errorf("mod %s not found", modID)
	}

	return mod.data, nil
}

// GetModVersion implements FileWatcher.
func (w *InMemoryFileWatcher) GetModVersion(modID string) (string, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	mod, exists := w.mods[modID]
	if !exists {
		return "", fmt.Errorf("mod %s not found", modID)
	}

	return mod.version, nil
}

// ComputeHash computes SHA256 hash of data.
func ComputeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// InMemoryStateMigrationHandler provides state migration for testing.
type InMemoryStateMigrationHandler struct {
	states map[string]*savedState
	mu     sync.RWMutex
}

type savedState struct {
	scripts   map[string]any
	variables map[string]any
}

// NewInMemoryStateMigrationHandler creates a new in-memory migration handler.
func NewInMemoryStateMigrationHandler() *InMemoryStateMigrationHandler {
	return &InMemoryStateMigrationHandler{
		states: make(map[string]*savedState),
	}
}

// SetState sets the state that will be returned by SaveState.
func (h *InMemoryStateMigrationHandler) SetState(modID string, scripts, variables map[string]any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.states[modID] = &savedState{
		scripts:   scripts,
		variables: variables,
	}
}

// SaveState implements StateMigrationHandler.
func (h *InMemoryStateMigrationHandler) SaveState(modID string) (map[string]any, map[string]any, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	state, exists := h.states[modID]
	if !exists {
		return make(map[string]any), make(map[string]any), nil
	}

	return state.scripts, state.variables, nil
}

// RestoreState implements StateMigrationHandler.
func (h *InMemoryStateMigrationHandler) RestoreState(modID string, scripts, variables map[string]any) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.states[modID] = &savedState{
		scripts:   scripts,
		variables: variables,
	}

	return nil
}
