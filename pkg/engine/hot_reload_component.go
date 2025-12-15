// Package engine provides the hot reload component for live mod updates.
// HotReloadComponent tracks mod file changes, pending updates, and reload state
// to enable modders to update mods without restarting the game.
package engine

import (
	"encoding/json"
	"sync"
	"time"
)

// HotReloadComponent tracks hot reload state for mods.
type HotReloadComponent struct {
	WatchedMods    map[string]*WatchedMod // modID -> watched mod state
	PendingUpdates []string               // mod IDs with detected changes
	LastReload     int64                  // unix timestamp of last reload
	AutoReload     bool                   // developer mode - auto reload on change
	ReloadHistory  []ReloadEntry          // recent reload history
	Enabled        bool                   // hot reload enabled
	WatchInterval  time.Duration          // how often to check for changes

	// Rollback state
	RollbackAvailable bool                 // rollback data exists
	RollbackData      map[string]*ModState // modID -> previous state for rollback

	mu sync.RWMutex
}

// WatchedMod represents a mod being monitored for changes.
type WatchedMod struct {
	ModID         string `json:"mod_id"`
	Version       string `json:"version"`
	LastModified  int64  `json:"last_modified"` // unix timestamp
	FileHash      string `json:"file_hash"`     // hash of mod files
	WatchStarted  int64  `json:"watch_started"` // when watching started
	ChangeCount   int    `json:"change_count"`  // number of detected changes
	LastChange    int64  `json:"last_change"`   // timestamp of last change
	PendingReload bool   `json:"pending_reload"`
}

// ReloadEntry represents a single reload event.
type ReloadEntry struct {
	ModID       string `json:"mod_id"`
	Timestamp   int64  `json:"timestamp"`   // when reload occurred
	OldVersion  string `json:"old_version"` // version before reload
	NewVersion  string `json:"new_version"` // version after reload
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	Duration    int64  `json:"duration_ms"` // reload duration in milliseconds
	RolledBack  bool   `json:"rolled_back"` // was this reload rolled back
	AutoReload  bool   `json:"auto_reload"` // was this an auto reload
	StateChange string `json:"state_change"`
}

// ModState represents the saved state of a mod for rollback.
type ModState struct {
	ModID     string         `json:"mod_id"`
	Version   string         `json:"version"`
	Scripts   map[string]any `json:"scripts"`   // script state
	Variables map[string]any `json:"variables"` // mod variables
	SavedAt   int64          `json:"saved_at"`  // when state was saved
}

// ReloadStatus represents the current status of a reload operation.
type ReloadStatus string

const (
	ReloadStatusIdle       ReloadStatus = "idle"
	ReloadStatusPending    ReloadStatus = "pending"
	ReloadStatusInProgress ReloadStatus = "in_progress"
	ReloadStatusComplete   ReloadStatus = "complete"
	ReloadStatusFailed     ReloadStatus = "failed"
	ReloadStatusRolledBack ReloadStatus = "rolled_back"
)

// MaxReloadHistory is the maximum number of reload entries to keep.
const MaxReloadHistory = 50

// DefaultWatchInterval is the default interval for checking mod changes.
const DefaultWatchInterval = 500 * time.Millisecond

// NewHotReloadComponent creates a new hot reload component.
func NewHotReloadComponent() *HotReloadComponent {
	return &HotReloadComponent{
		WatchedMods:    make(map[string]*WatchedMod),
		PendingUpdates: make([]string, 0),
		ReloadHistory:  make([]ReloadEntry, 0),
		RollbackData:   make(map[string]*ModState),
		Enabled:        true,
		WatchInterval:  DefaultWatchInterval,
	}
}

// Type returns the component type identifier.
func (h *HotReloadComponent) Type() string {
	return "hot_reload"
}

// StartWatching begins monitoring a mod for changes.
func (h *HotReloadComponent) StartWatching(modID, version, fileHash string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now().Unix()
	h.WatchedMods[modID] = &WatchedMod{
		ModID:        modID,
		Version:      version,
		LastModified: now,
		FileHash:     fileHash,
		WatchStarted: now,
	}
}

// StopWatching stops monitoring a mod for changes.
func (h *HotReloadComponent) StopWatching(modID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.WatchedMods, modID)

	// Remove from pending updates
	for i := len(h.PendingUpdates) - 1; i >= 0; i-- {
		if h.PendingUpdates[i] == modID {
			h.PendingUpdates = append(h.PendingUpdates[:i], h.PendingUpdates[i+1:]...)
		}
	}
}

// IsWatching returns whether a mod is being watched.
func (h *HotReloadComponent) IsWatching(modID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.WatchedMods[modID]
	return exists
}

// GetWatchedMod returns the watched state for a mod.
func (h *HotReloadComponent) GetWatchedMod(modID string) (*WatchedMod, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	mod, exists := h.WatchedMods[modID]
	if !exists {
		return nil, false
	}
	// Return a copy
	copy := *mod
	return &copy, true
}

// GetWatchedModIDs returns all watched mod IDs.
func (h *HotReloadComponent) GetWatchedModIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.WatchedMods))
	for id := range h.WatchedMods {
		ids = append(ids, id)
	}
	return ids
}

// DetectChange records that a mod file has changed.
func (h *HotReloadComponent) DetectChange(modID, newHash string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	mod, exists := h.WatchedMods[modID]
	if !exists {
		return false
	}

	// Check if hash actually changed
	if mod.FileHash == newHash {
		return false
	}

	now := time.Now().Unix()
	mod.FileHash = newHash
	mod.LastChange = now
	mod.ChangeCount++
	mod.PendingReload = true

	// Add to pending updates if not already there
	found := false
	for _, id := range h.PendingUpdates {
		if id == modID {
			found = true
			break
		}
	}
	if !found {
		h.PendingUpdates = append(h.PendingUpdates, modID)
	}

	return true
}

// GetPendingUpdates returns mod IDs with pending updates.
func (h *HotReloadComponent) GetPendingUpdates() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, len(h.PendingUpdates))
	copy(result, h.PendingUpdates)
	return result
}

// HasPendingUpdates returns whether there are pending updates.
func (h *HotReloadComponent) HasPendingUpdates() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.PendingUpdates) > 0
}

// ClearPendingUpdate removes a mod from pending updates.
func (h *HotReloadComponent) ClearPendingUpdate(modID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i := len(h.PendingUpdates) - 1; i >= 0; i-- {
		if h.PendingUpdates[i] == modID {
			h.PendingUpdates = append(h.PendingUpdates[:i], h.PendingUpdates[i+1:]...)
		}
	}

	if mod, exists := h.WatchedMods[modID]; exists {
		mod.PendingReload = false
	}
}

// ClearAllPendingUpdates removes all pending updates.
func (h *HotReloadComponent) ClearAllPendingUpdates() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.PendingUpdates = make([]string, 0)
	for _, mod := range h.WatchedMods {
		mod.PendingReload = false
	}
}

// SaveStateForRollback saves mod state for potential rollback.
func (h *HotReloadComponent) SaveStateForRollback(modID, version string, scripts, variables map[string]any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.RollbackData[modID] = &ModState{
		ModID:     modID,
		Version:   version,
		Scripts:   scripts,
		Variables: variables,
		SavedAt:   time.Now().Unix(),
	}
	h.RollbackAvailable = true
}

// GetRollbackState returns the saved state for a mod.
func (h *HotReloadComponent) GetRollbackState(modID string) (*ModState, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	state, exists := h.RollbackData[modID]
	if !exists {
		return nil, false
	}
	// Return a copy
	copy := *state
	return &copy, true
}

// ClearRollbackState removes rollback state for a mod.
func (h *HotReloadComponent) ClearRollbackState(modID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.RollbackData, modID)
	if len(h.RollbackData) == 0 {
		h.RollbackAvailable = false
	}
}

// ClearAllRollbackState removes all rollback state.
func (h *HotReloadComponent) ClearAllRollbackState() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.RollbackData = make(map[string]*ModState)
	h.RollbackAvailable = false
}

// AddReloadEntry adds a reload event to history.
func (h *HotReloadComponent) AddReloadEntry(entry ReloadEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.ReloadHistory = append(h.ReloadHistory, entry)

	// Trim history if too long
	if len(h.ReloadHistory) > MaxReloadHistory {
		h.ReloadHistory = h.ReloadHistory[len(h.ReloadHistory)-MaxReloadHistory:]
	}

	h.LastReload = entry.Timestamp
}

// GetReloadHistory returns recent reload history.
func (h *HotReloadComponent) GetReloadHistory() []ReloadEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]ReloadEntry, len(h.ReloadHistory))
	copy(result, h.ReloadHistory)
	return result
}

// GetReloadHistoryForMod returns reload history for a specific mod.
func (h *HotReloadComponent) GetReloadHistoryForMod(modID string) []ReloadEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []ReloadEntry
	for _, entry := range h.ReloadHistory {
		if entry.ModID == modID {
			result = append(result, entry)
		}
	}
	return result
}

// GetLastReloadTime returns the timestamp of the last reload.
func (h *HotReloadComponent) GetLastReloadTime() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.LastReload
}

// SetAutoReload enables or disables automatic reloading.
func (h *HotReloadComponent) SetAutoReload(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.AutoReload = enabled
}

// IsAutoReloadEnabled returns whether auto reload is enabled.
func (h *HotReloadComponent) IsAutoReloadEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.AutoReload
}

// SetEnabled enables or disables hot reload.
func (h *HotReloadComponent) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Enabled = enabled
}

// IsEnabled returns whether hot reload is enabled.
func (h *HotReloadComponent) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.Enabled
}

// SetWatchInterval sets the interval for checking mod changes.
func (h *HotReloadComponent) SetWatchInterval(interval time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.WatchInterval = interval
}

// GetWatchInterval returns the watch interval.
func (h *HotReloadComponent) GetWatchInterval() time.Duration {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.WatchInterval
}

// UpdateModVersion updates the version of a watched mod after reload.
func (h *HotReloadComponent) UpdateModVersion(modID, version, fileHash string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if mod, exists := h.WatchedMods[modID]; exists {
		mod.Version = version
		mod.FileHash = fileHash
		mod.LastModified = time.Now().Unix()
	}
}

// GetWatchedModCount returns the number of watched mods.
func (h *HotReloadComponent) GetWatchedModCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.WatchedMods)
}

// GetTotalReloads returns the total number of reloads.
func (h *HotReloadComponent) GetTotalReloads() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.ReloadHistory)
}

// GetSuccessfulReloads returns the number of successful reloads.
func (h *HotReloadComponent) GetSuccessfulReloads() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, entry := range h.ReloadHistory {
		if entry.Success {
			count++
		}
	}
	return count
}

// GetFailedReloads returns the number of failed reloads.
func (h *HotReloadComponent) GetFailedReloads() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, entry := range h.ReloadHistory {
		if !entry.Success {
			count++
		}
	}
	return count
}

// HotReloadData represents serializable hot reload data.
type HotReloadData struct {
	WatchedMods   map[string]*WatchedMod `json:"watched_mods"`
	AutoReload    bool                   `json:"auto_reload"`
	ReloadHistory []ReloadEntry          `json:"reload_history"`
	Enabled       bool                   `json:"enabled"`
	WatchInterval int64                  `json:"watch_interval_ms"`
}

// Serialize converts the component to JSON bytes.
func (h *HotReloadComponent) Serialize() ([]byte, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data := HotReloadData{
		WatchedMods:   h.WatchedMods,
		AutoReload:    h.AutoReload,
		ReloadHistory: h.ReloadHistory,
		Enabled:       h.Enabled,
		WatchInterval: h.WatchInterval.Milliseconds(),
	}

	return json.Marshal(data)
}

// Deserialize loads the component from JSON bytes.
func (h *HotReloadComponent) Deserialize(data []byte) error {
	var reloadData HotReloadData
	if err := json.Unmarshal(data, &reloadData); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if reloadData.WatchedMods != nil {
		h.WatchedMods = reloadData.WatchedMods
	} else {
		h.WatchedMods = make(map[string]*WatchedMod)
	}

	h.AutoReload = reloadData.AutoReload
	h.Enabled = reloadData.Enabled

	if reloadData.ReloadHistory != nil {
		h.ReloadHistory = reloadData.ReloadHistory
	} else {
		h.ReloadHistory = make([]ReloadEntry, 0)
	}

	if reloadData.WatchInterval > 0 {
		h.WatchInterval = time.Duration(reloadData.WatchInterval) * time.Millisecond
	} else {
		h.WatchInterval = DefaultWatchInterval
	}

	// Reset transient state
	h.PendingUpdates = make([]string, 0)
	h.RollbackData = make(map[string]*ModState)
	h.RollbackAvailable = false

	return nil
}
