// Package engine provides the mod compatibility component for conflict detection.
// ModCompatibilityComponent tracks mod conflicts, dependencies, and load order
// to ensure mods work together without breaking the game.
package engine

import (
	"encoding/json"
	"sort"
	"sync"
)

// ModCompatibilityComponent tracks mod conflicts, dependencies, and load order.
type ModCompatibilityComponent struct {
	Conflicts      []ModConflict          // detected conflicts between mods
	Dependencies   map[string][]string    // modID -> required mod IDs
	LoadOrder      []string               // calculated optimal load order
	Warnings       []CompatibilityWarning // compatibility warnings
	ModVersions    map[string]string      // modID -> version string
	GameVersion    string                 // current game version
	Configurations map[string]ModConfig2  // saved mod configurations
	ActiveConfigID string                 // currently active configuration

	mu sync.RWMutex
}

// ModConflict represents a conflict between two mods.
type ModConflict struct {
	Mod1         string           `json:"mod1"`          // first mod ID
	Mod2         string           `json:"mod2"`          // second mod ID
	ConflictType ConflictType     `json:"conflict_type"` // type of conflict
	Description  string           `json:"description"`   // human-readable description
	Severity     ConflictSeverity `json:"severity"`      // error, warning, info
	Suggestion   string           `json:"suggestion"`    // suggested resolution
	AffectedArea string           `json:"affected_area"` // what game area is affected
}

// ConflictType represents the type of mod conflict.
type ConflictType string

const (
	ConflictTypeRule     ConflictType = "rule"     // both mods modify same rule
	ConflictTypeEvent    ConflictType = "event"    // both mods handle same event
	ConflictTypeResource ConflictType = "resource" // both mods modify same resource
	ConflictTypeOverride ConflictType = "override" // one mod overrides another
	ConflictTypeVersion  ConflictType = "version"  // version incompatibility
)

// ConflictSeverity represents how serious a conflict is.
type ConflictSeverity string

const (
	ConflictSeverityError   ConflictSeverity = "error"   // mods cannot work together
	ConflictSeverityWarning ConflictSeverity = "warning" // mods may have issues
	ConflictSeverityInfo    ConflictSeverity = "info"    // informational only
)

// CompatibilityWarning represents a non-blocking compatibility issue.
type CompatibilityWarning struct {
	ModID       string `json:"mod_id"`       // affected mod
	WarningType string `json:"warning_type"` // type of warning
	Message     string `json:"message"`      // warning message
	Suggestion  string `json:"suggestion"`   // how to resolve
}

// ModConfig2 represents a saved mod configuration (using ModConfig2 to avoid collision).
type ModConfig2 struct {
	ID          string   `json:"id"`           // configuration ID
	Name        string   `json:"name"`         // human-readable name
	Description string   `json:"description"`  // configuration description
	EnabledMods []string `json:"enabled_mods"` // mods in this configuration
	LoadOrder   []string `json:"load_order"`   // custom load order
	CreatedAt   int64    `json:"created_at"`   // unix timestamp
	UpdatedAt   int64    `json:"updated_at"`   // unix timestamp
}

// NewModCompatibilityComponent creates a new mod compatibility component.
func NewModCompatibilityComponent() *ModCompatibilityComponent {
	return &ModCompatibilityComponent{
		Conflicts:      make([]ModConflict, 0),
		Dependencies:   make(map[string][]string),
		LoadOrder:      make([]string, 0),
		Warnings:       make([]CompatibilityWarning, 0),
		ModVersions:    make(map[string]string),
		Configurations: make(map[string]ModConfig2),
	}
}

// Type returns the component type identifier.
func (m *ModCompatibilityComponent) Type() string {
	return "mod_compatibility"
}

// SetGameVersion sets the current game version for compatibility checks.
func (m *ModCompatibilityComponent) SetGameVersion(version string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GameVersion = version
}

// GetGameVersion returns the current game version.
func (m *ModCompatibilityComponent) GetGameVersion() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.GameVersion
}

// SetModVersion records the version of an installed mod.
func (m *ModCompatibilityComponent) SetModVersion(modID, version string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ModVersions[modID] = version
}

// GetModVersion returns the version of an installed mod.
func (m *ModCompatibilityComponent) GetModVersion(modID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	version, exists := m.ModVersions[modID]
	return version, exists
}

// RemoveModVersion removes version tracking for a mod.
func (m *ModCompatibilityComponent) RemoveModVersion(modID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.ModVersions, modID)
}

// SetDependencies sets the dependencies for a mod.
func (m *ModCompatibilityComponent) SetDependencies(modID string, deps []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if deps == nil {
		deps = make([]string, 0)
	}
	m.Dependencies[modID] = deps
}

// GetDependencies returns the dependencies for a mod.
func (m *ModCompatibilityComponent) GetDependencies(modID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deps, exists := m.Dependencies[modID]
	if !exists {
		return nil
	}
	// Return a copy
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// RemoveDependencies removes dependency tracking for a mod.
func (m *ModCompatibilityComponent) RemoveDependencies(modID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Dependencies, modID)
}

// AddConflict adds a detected conflict.
func (m *ModCompatibilityComponent) AddConflict(conflict ModConflict) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate
	for _, existing := range m.Conflicts {
		if existing.Mod1 == conflict.Mod1 && existing.Mod2 == conflict.Mod2 &&
			existing.ConflictType == conflict.ConflictType {
			return // Already exists
		}
	}

	m.Conflicts = append(m.Conflicts, conflict)
}

// RemoveConflict removes a conflict by mod IDs.
func (m *ModCompatibilityComponent) RemoveConflict(mod1, mod2 string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := len(m.Conflicts) - 1; i >= 0; i-- {
		c := m.Conflicts[i]
		if (c.Mod1 == mod1 && c.Mod2 == mod2) || (c.Mod1 == mod2 && c.Mod2 == mod1) {
			m.Conflicts = append(m.Conflicts[:i], m.Conflicts[i+1:]...)
		}
	}
}

// GetConflicts returns all detected conflicts.
func (m *ModCompatibilityComponent) GetConflicts() []ModConflict {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]ModConflict, len(m.Conflicts))
	copy(result, m.Conflicts)
	return result
}

// GetConflictsForMod returns all conflicts involving a specific mod.
func (m *ModCompatibilityComponent) GetConflictsForMod(modID string) []ModConflict {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []ModConflict
	for _, c := range m.Conflicts {
		if c.Mod1 == modID || c.Mod2 == modID {
			result = append(result, c)
		}
	}
	return result
}

// GetConflictsBySeverity returns conflicts filtered by severity.
func (m *ModCompatibilityComponent) GetConflictsBySeverity(severity ConflictSeverity) []ModConflict {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []ModConflict
	for _, c := range m.Conflicts {
		if c.Severity == severity {
			result = append(result, c)
		}
	}
	return result
}

// ClearConflicts removes all conflicts.
func (m *ModCompatibilityComponent) ClearConflicts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Conflicts = make([]ModConflict, 0)
}

// HasBlockingConflicts returns true if there are any error-level conflicts.
func (m *ModCompatibilityComponent) HasBlockingConflicts() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.Conflicts {
		if c.Severity == ConflictSeverityError {
			return true
		}
	}
	return false
}

// AddWarning adds a compatibility warning.
func (m *ModCompatibilityComponent) AddWarning(warning CompatibilityWarning) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Warnings = append(m.Warnings, warning)
}

// GetWarnings returns all warnings.
func (m *ModCompatibilityComponent) GetWarnings() []CompatibilityWarning {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]CompatibilityWarning, len(m.Warnings))
	copy(result, m.Warnings)
	return result
}

// GetWarningsForMod returns warnings for a specific mod.
func (m *ModCompatibilityComponent) GetWarningsForMod(modID string) []CompatibilityWarning {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []CompatibilityWarning
	for _, w := range m.Warnings {
		if w.ModID == modID {
			result = append(result, w)
		}
	}
	return result
}

// ClearWarnings removes all warnings.
func (m *ModCompatibilityComponent) ClearWarnings() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Warnings = make([]CompatibilityWarning, 0)
}

// SetLoadOrder sets the calculated load order.
func (m *ModCompatibilityComponent) SetLoadOrder(order []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LoadOrder = make([]string, len(order))
	copy(m.LoadOrder, order)
}

// GetLoadOrder returns the calculated load order.
func (m *ModCompatibilityComponent) GetLoadOrder() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, len(m.LoadOrder))
	copy(result, m.LoadOrder)
	return result
}

// GetLoadPosition returns the position of a mod in the load order (-1 if not found).
func (m *ModCompatibilityComponent) GetLoadPosition(modID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i, id := range m.LoadOrder {
		if id == modID {
			return i
		}
	}
	return -1
}

// SaveConfiguration saves the current mod state as a named configuration.
func (m *ModCompatibilityComponent) SaveConfiguration(config ModConfig2) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Configurations[config.ID] = config
}

// GetConfiguration returns a saved configuration by ID.
func (m *ModCompatibilityComponent) GetConfiguration(configID string) (ModConfig2, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, exists := m.Configurations[configID]
	return config, exists
}

// DeleteConfiguration removes a saved configuration.
func (m *ModCompatibilityComponent) DeleteConfiguration(configID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Configurations, configID)
}

// ListConfigurations returns all saved configurations.
func (m *ModCompatibilityComponent) ListConfigurations() []ModConfig2 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configs := make([]ModConfig2, 0, len(m.Configurations))
	for _, config := range m.Configurations {
		configs = append(configs, config)
	}

	// Sort by name for consistent ordering
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name < configs[j].Name
	})

	return configs
}

// SetActiveConfiguration sets the currently active configuration.
func (m *ModCompatibilityComponent) SetActiveConfiguration(configID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveConfigID = configID
}

// GetActiveConfiguration returns the currently active configuration ID.
func (m *ModCompatibilityComponent) GetActiveConfiguration() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ActiveConfigID
}

// GetConflictCount returns the total number of conflicts.
func (m *ModCompatibilityComponent) GetConflictCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Conflicts)
}

// GetWarningCount returns the total number of warnings.
func (m *ModCompatibilityComponent) GetWarningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Warnings)
}

// GetErrorCount returns the number of error-level conflicts.
func (m *ModCompatibilityComponent) GetErrorCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, c := range m.Conflicts {
		if c.Severity == ConflictSeverityError {
			count++
		}
	}
	return count
}

// ModCompatibilityData represents serializable compatibility data.
type ModCompatibilityData struct {
	Conflicts      []ModConflict          `json:"conflicts"`
	Dependencies   map[string][]string    `json:"dependencies"`
	LoadOrder      []string               `json:"load_order"`
	Warnings       []CompatibilityWarning `json:"warnings"`
	ModVersions    map[string]string      `json:"mod_versions"`
	GameVersion    string                 `json:"game_version"`
	Configurations map[string]ModConfig2  `json:"configurations"`
	ActiveConfigID string                 `json:"active_config_id"`
}

// Serialize converts the component to JSON bytes.
func (m *ModCompatibilityComponent) Serialize() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := ModCompatibilityData{
		Conflicts:      m.Conflicts,
		Dependencies:   m.Dependencies,
		LoadOrder:      m.LoadOrder,
		Warnings:       m.Warnings,
		ModVersions:    m.ModVersions,
		GameVersion:    m.GameVersion,
		Configurations: m.Configurations,
		ActiveConfigID: m.ActiveConfigID,
	}

	return json.Marshal(data)
}

// Deserialize loads the component from JSON bytes.
func (m *ModCompatibilityComponent) Deserialize(data []byte) error {
	var compatData ModCompatibilityData
	if err := json.Unmarshal(data, &compatData); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if compatData.Conflicts != nil {
		m.Conflicts = compatData.Conflicts
	} else {
		m.Conflicts = make([]ModConflict, 0)
	}

	if compatData.Dependencies != nil {
		m.Dependencies = compatData.Dependencies
	} else {
		m.Dependencies = make(map[string][]string)
	}

	if compatData.LoadOrder != nil {
		m.LoadOrder = compatData.LoadOrder
	} else {
		m.LoadOrder = make([]string, 0)
	}

	if compatData.Warnings != nil {
		m.Warnings = compatData.Warnings
	} else {
		m.Warnings = make([]CompatibilityWarning, 0)
	}

	if compatData.ModVersions != nil {
		m.ModVersions = compatData.ModVersions
	} else {
		m.ModVersions = make(map[string]string)
	}

	m.GameVersion = compatData.GameVersion

	if compatData.Configurations != nil {
		m.Configurations = compatData.Configurations
	} else {
		m.Configurations = make(map[string]ModConfig2)
	}

	m.ActiveConfigID = compatData.ActiveConfigID

	return nil
}
