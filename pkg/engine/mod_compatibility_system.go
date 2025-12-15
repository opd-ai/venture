// Package engine provides the mod compatibility system for conflict detection.
// ModCompatibilitySystem validates mod combinations, detects conflicts, and
// calculates optimal load order based on dependencies.
package engine

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// ModCompatibilitySystem manages mod compatibility validation and load order.
type ModCompatibilitySystem struct {
	world            *World
	ruleRegistry     map[string][]string     // ruleName -> modIDs that modify it
	eventRegistry    map[string][]string     // eventName -> modIDs that handle it
	resourceRegistry map[string][]string     // resourceName -> modIDs that modify it
	modMetadata      map[string]*ModMetadata // modID -> metadata

	mu sync.RWMutex
}

// ModMetadata stores metadata about a mod for compatibility analysis.
type ModMetadata struct {
	ID             string
	Version        string
	MinGameVersion string
	MaxGameVersion string
	Rules          []string // rules this mod modifies
	Events         []string // events this mod handles
	Resources      []string // resources this mod modifies
	Dependencies   []string
	Conflicts      []string // known incompatible mods
	Enabled        bool
}

// NewModCompatibilitySystem creates a new mod compatibility system.
func NewModCompatibilitySystem(world *World) *ModCompatibilitySystem {
	log.WithFields(log.Fields{
		"system_name": "mod_compatibility",
	}).Debug("Creating mod compatibility system")

	return &ModCompatibilitySystem{
		world:            world,
		ruleRegistry:     make(map[string][]string),
		eventRegistry:    make(map[string][]string),
		resourceRegistry: make(map[string][]string),
		modMetadata:      make(map[string]*ModMetadata),
	}
}

// Update processes compatibility checks for entities with ModCompatibilityComponent.
func (s *ModCompatibilitySystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("mod_compatibility") {
			continue
		}
		// System is event-driven; no per-frame updates needed
	}
}

// RegisterMod registers a mod and its metadata for compatibility tracking.
func (s *ModCompatibilitySystem) RegisterMod(metadata *ModMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}
	if metadata.ID == "" {
		return fmt.Errorf("mod ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.modMetadata[metadata.ID] = metadata

	// Register rules
	for _, rule := range metadata.Rules {
		s.ruleRegistry[rule] = append(s.ruleRegistry[rule], metadata.ID)
	}

	// Register events
	for _, event := range metadata.Events {
		s.eventRegistry[event] = append(s.eventRegistry[event], metadata.ID)
	}

	// Register resources
	for _, resource := range metadata.Resources {
		s.resourceRegistry[resource] = append(s.resourceRegistry[resource], metadata.ID)
	}

	log.WithFields(log.Fields{
		"system_name": "mod_compatibility",
		"mod_id":      metadata.ID,
		"version":     metadata.Version,
		"rules":       len(metadata.Rules),
		"events":      len(metadata.Events),
	}).Debug("Registered mod for compatibility tracking")

	return nil
}

// UnregisterMod removes a mod from compatibility tracking.
func (s *ModCompatibilitySystem) UnregisterMod(modID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metadata, exists := s.modMetadata[modID]
	if !exists {
		return
	}

	// Remove from rule registry
	for _, rule := range metadata.Rules {
		s.removeFromSlice(s.ruleRegistry, rule, modID)
	}

	// Remove from event registry
	for _, event := range metadata.Events {
		s.removeFromSlice(s.eventRegistry, event, modID)
	}

	// Remove from resource registry
	for _, resource := range metadata.Resources {
		s.removeFromSlice(s.resourceRegistry, resource, modID)
	}

	delete(s.modMetadata, modID)

	log.WithFields(log.Fields{
		"system_name": "mod_compatibility",
		"mod_id":      modID,
	}).Debug("Unregistered mod from compatibility tracking")
}

// removeFromSlice removes a value from a registry slice.
func (s *ModCompatibilitySystem) removeFromSlice(registry map[string][]string, key, value string) {
	slice := registry[key]
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == value {
			registry[key] = append(slice[:i], slice[i+1:]...)
			break
		}
	}
}

// ValidateMods validates a set of mods and updates the compatibility component.
func (s *ModCompatibilitySystem) ValidateMods(comp *ModCompatibilityComponent, modIDs []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	comp.ClearConflicts()
	comp.ClearWarnings()

	// Check for known incompatibilities
	s.detectKnownConflicts(comp, modIDs)

	// Check for rule conflicts
	s.detectRuleConflicts(comp, modIDs)

	// Check for event conflicts
	s.detectEventConflicts(comp, modIDs)

	// Check for resource conflicts
	s.detectResourceConflicts(comp, modIDs)

	// Check version compatibility
	s.detectVersionConflicts(comp, modIDs)

	// Check for missing dependencies
	s.detectMissingDependencies(comp, modIDs)

	log.WithFields(log.Fields{
		"system_name": "mod_compatibility",
		"mods":        len(modIDs),
		"conflicts":   comp.GetConflictCount(),
		"warnings":    comp.GetWarningCount(),
	}).Debug("Completed mod validation")
}

// detectKnownConflicts checks for explicitly declared mod incompatibilities.
func (s *ModCompatibilitySystem) detectKnownConflicts(comp *ModCompatibilityComponent, modIDs []string) {
	modSet := make(map[string]bool)
	for _, id := range modIDs {
		modSet[id] = true
	}

	for _, modID := range modIDs {
		metadata, exists := s.modMetadata[modID]
		if !exists {
			continue
		}

		for _, conflictID := range metadata.Conflicts {
			if modSet[conflictID] {
				comp.AddConflict(ModConflict{
					Mod1:         modID,
					Mod2:         conflictID,
					ConflictType: ConflictTypeOverride,
					Description:  fmt.Sprintf("%s declares incompatibility with %s", modID, conflictID),
					Severity:     ConflictSeverityError,
					Suggestion:   fmt.Sprintf("Disable either %s or %s", modID, conflictID),
				})
			}
		}
	}
}

// detectRuleConflicts checks for mods modifying the same rules.
func (s *ModCompatibilitySystem) detectRuleConflicts(comp *ModCompatibilityComponent, modIDs []string) {
	modSet := make(map[string]bool)
	for _, id := range modIDs {
		modSet[id] = true
	}

	for rule, mods := range s.ruleRegistry {
		// Filter to only enabled mods
		var enabledMods []string
		for _, modID := range mods {
			if modSet[modID] {
				enabledMods = append(enabledMods, modID)
			}
		}

		// If multiple mods modify same rule, create conflict
		if len(enabledMods) > 1 {
			for i := 0; i < len(enabledMods)-1; i++ {
				for j := i + 1; j < len(enabledMods); j++ {
					comp.AddConflict(ModConflict{
						Mod1:         enabledMods[i],
						Mod2:         enabledMods[j],
						ConflictType: ConflictTypeRule,
						Description:  fmt.Sprintf("Both mods modify rule '%s'", rule),
						Severity:     ConflictSeverityWarning,
						Suggestion:   "Later mod in load order will override",
						AffectedArea: rule,
					})
				}
			}
		}
	}
}

// detectEventConflicts checks for mods handling the same events.
func (s *ModCompatibilitySystem) detectEventConflicts(comp *ModCompatibilityComponent, modIDs []string) {
	modSet := make(map[string]bool)
	for _, id := range modIDs {
		modSet[id] = true
	}

	for event, mods := range s.eventRegistry {
		var enabledMods []string
		for _, modID := range mods {
			if modSet[modID] {
				enabledMods = append(enabledMods, modID)
			}
		}

		// Multiple mods handling same event is usually OK, just informational
		if len(enabledMods) > 1 {
			for i := 0; i < len(enabledMods)-1; i++ {
				for j := i + 1; j < len(enabledMods); j++ {
					comp.AddConflict(ModConflict{
						Mod1:         enabledMods[i],
						Mod2:         enabledMods[j],
						ConflictType: ConflictTypeEvent,
						Description:  fmt.Sprintf("Both mods handle event '%s'", event),
						Severity:     ConflictSeverityInfo,
						Suggestion:   "Both handlers will execute; check load order",
						AffectedArea: event,
					})
				}
			}
		}
	}
}

// detectResourceConflicts checks for mods modifying the same resources.
func (s *ModCompatibilitySystem) detectResourceConflicts(comp *ModCompatibilityComponent, modIDs []string) {
	modSet := make(map[string]bool)
	for _, id := range modIDs {
		modSet[id] = true
	}

	for resource, mods := range s.resourceRegistry {
		var enabledMods []string
		for _, modID := range mods {
			if modSet[modID] {
				enabledMods = append(enabledMods, modID)
			}
		}

		if len(enabledMods) > 1 {
			for i := 0; i < len(enabledMods)-1; i++ {
				for j := i + 1; j < len(enabledMods); j++ {
					comp.AddConflict(ModConflict{
						Mod1:         enabledMods[i],
						Mod2:         enabledMods[j],
						ConflictType: ConflictTypeResource,
						Description:  fmt.Sprintf("Both mods modify resource '%s'", resource),
						Severity:     ConflictSeverityWarning,
						Suggestion:   "Only one mod's changes will take effect",
						AffectedArea: resource,
					})
				}
			}
		}
	}
}

// detectVersionConflicts checks for game version incompatibilities.
func (s *ModCompatibilitySystem) detectVersionConflicts(comp *ModCompatibilityComponent, modIDs []string) {
	gameVersion := comp.GetGameVersion()
	if gameVersion == "" {
		return
	}

	for _, modID := range modIDs {
		metadata, exists := s.modMetadata[modID]
		if !exists {
			continue
		}

		// Check minimum version
		if metadata.MinGameVersion != "" && compareModVersions(gameVersion, metadata.MinGameVersion) < 0 {
			comp.AddConflict(ModConflict{
				Mod1:         modID,
				Mod2:         "game",
				ConflictType: ConflictTypeVersion,
				Description:  fmt.Sprintf("%s requires game version %s or higher", modID, metadata.MinGameVersion),
				Severity:     ConflictSeverityError,
				Suggestion:   fmt.Sprintf("Update game to version %s or higher, or find older mod version", metadata.MinGameVersion),
			})
		}

		// Check maximum version
		if metadata.MaxGameVersion != "" && compareModVersions(gameVersion, metadata.MaxGameVersion) > 0 {
			comp.AddWarning(CompatibilityWarning{
				ModID:       modID,
				WarningType: "version",
				Message:     fmt.Sprintf("Mod was made for game version %s or lower", metadata.MaxGameVersion),
				Suggestion:  "Mod may not work correctly; check for updates",
			})
		}
	}
}

// detectMissingDependencies checks for unmet mod dependencies.
func (s *ModCompatibilitySystem) detectMissingDependencies(comp *ModCompatibilityComponent, modIDs []string) {
	modSet := make(map[string]bool)
	for _, id := range modIDs {
		modSet[id] = true
	}

	for _, modID := range modIDs {
		metadata, exists := s.modMetadata[modID]
		if !exists {
			continue
		}

		for _, depID := range metadata.Dependencies {
			if !modSet[depID] {
				comp.AddConflict(ModConflict{
					Mod1:         modID,
					Mod2:         depID,
					ConflictType: ConflictTypeVersion,
					Description:  fmt.Sprintf("%s requires %s", modID, depID),
					Severity:     ConflictSeverityError,
					Suggestion:   fmt.Sprintf("Install and enable %s", depID),
				})
			}
		}
	}
}

// CalculateLoadOrder computes the optimal load order based on dependencies.
func (s *ModCompatibilitySystem) CalculateLoadOrder(comp *ModCompatibilityComponent, modIDs []string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Build dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, modID := range modIDs {
		graph[modID] = []string{}
		inDegree[modID] = 0
	}

	// Add edges for dependencies
	for _, modID := range modIDs {
		metadata, exists := s.modMetadata[modID]
		if !exists {
			continue
		}

		for _, depID := range metadata.Dependencies {
			// Only add edge if dependency is in our mod list
			if _, inList := inDegree[depID]; inList {
				graph[depID] = append(graph[depID], modID)
				inDegree[modID]++
			}
		}
	}

	// Topological sort (Kahn's algorithm)
	var queue []string
	for modID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, modID)
		}
	}

	// Sort queue for deterministic ordering
	sort.Strings(queue)

	var result []string
	for len(queue) > 0 {
		// Take first (alphabetically first for determinism)
		modID := queue[0]
		queue = queue[1:]
		result = append(result, modID)

		// Process dependents
		dependents := graph[modID]
		sort.Strings(dependents) // Deterministic
		for _, depID := range dependents {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				queue = append(queue, depID)
				sort.Strings(queue)
			}
		}
	}

	// Check for cycles
	if len(result) != len(modIDs) {
		// Find mods that couldn't be sorted (cycle)
		var cycledMods []string
		for modID, degree := range inDegree {
			if degree > 0 {
				cycledMods = append(cycledMods, modID)
			}
		}
		return nil, fmt.Errorf("circular dependency detected involving: %v", cycledMods)
	}

	comp.SetLoadOrder(result)

	log.WithFields(log.Fields{
		"system_name": "mod_compatibility",
		"mods":        len(result),
		"order":       result,
	}).Debug("Calculated load order")

	return result, nil
}

// GetRecommendedResolutions returns suggested fixes for conflicts.
func (s *ModCompatibilitySystem) GetRecommendedResolutions(comp *ModCompatibilityComponent) []string {
	conflicts := comp.GetConflicts()
	var suggestions []string

	seenSuggestions := make(map[string]bool)

	for _, conflict := range conflicts {
		if conflict.Suggestion != "" && !seenSuggestions[conflict.Suggestion] {
			suggestions = append(suggestions, conflict.Suggestion)
			seenSuggestions[conflict.Suggestion] = true
		}
	}

	return suggestions
}

// ExportConfiguration exports the current mod state as a configuration.
func (s *ModCompatibilitySystem) ExportConfiguration(comp *ModCompatibilityComponent, name, description string, modIDs []string) ModConfig2 {
	now := time.Now().Unix()

	// Get load order if available, otherwise use mod IDs
	loadOrder := comp.GetLoadOrder()
	if len(loadOrder) == 0 {
		loadOrder = make([]string, len(modIDs))
		copy(loadOrder, modIDs)
		sort.Strings(loadOrder)
	}

	config := ModConfig2{
		ID:          fmt.Sprintf("config_%d", now),
		Name:        name,
		Description: description,
		EnabledMods: modIDs,
		LoadOrder:   loadOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	comp.SaveConfiguration(config)

	log.WithFields(log.Fields{
		"system_name": "mod_compatibility",
		"config_id":   config.ID,
		"name":        name,
		"mods":        len(modIDs),
	}).Debug("Exported mod configuration")

	return config
}

// ImportConfiguration loads a saved configuration and validates it.
func (s *ModCompatibilitySystem) ImportConfiguration(comp *ModCompatibilityComponent, configID string) ([]string, error) {
	config, exists := comp.GetConfiguration(configID)
	if !exists {
		return nil, fmt.Errorf("configuration %s not found", configID)
	}

	// Validate the mods in this configuration
	s.ValidateMods(comp, config.EnabledMods)

	// Check for blocking conflicts
	if comp.HasBlockingConflicts() {
		return nil, fmt.Errorf("configuration has blocking conflicts")
	}

	// Set load order from configuration
	comp.SetLoadOrder(config.LoadOrder)
	comp.SetActiveConfiguration(configID)

	log.WithFields(log.Fields{
		"system_name": "mod_compatibility",
		"config_id":   configID,
		"mods":        len(config.EnabledMods),
	}).Debug("Imported mod configuration")

	return config.EnabledMods, nil
}

// compareModVersions compares two semantic version strings.
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareModVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Pad to same length
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// CheckModCompatibility performs a quick compatibility check for a single mod.
func (s *ModCompatibilitySystem) CheckModCompatibility(comp *ModCompatibilityComponent, modID string, existingMods []string) (bool, []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata, exists := s.modMetadata[modID]
	if !exists {
		return true, nil // Unknown mod, assume compatible
	}

	var issues []string

	// Check known conflicts
	modSet := make(map[string]bool)
	for _, id := range existingMods {
		modSet[id] = true
	}

	for _, conflictID := range metadata.Conflicts {
		if modSet[conflictID] {
			issues = append(issues, fmt.Sprintf("Incompatible with %s", conflictID))
		}
	}

	// Check dependencies
	for _, depID := range metadata.Dependencies {
		if !modSet[depID] {
			issues = append(issues, fmt.Sprintf("Requires %s", depID))
		}
	}

	// Check game version
	gameVersion := comp.GetGameVersion()
	if gameVersion != "" && metadata.MinGameVersion != "" {
		if compareModVersions(gameVersion, metadata.MinGameVersion) < 0 {
			issues = append(issues, fmt.Sprintf("Requires game version %s+", metadata.MinGameVersion))
		}
	}

	return len(issues) == 0, issues
}

// GetModMetadata returns metadata for a registered mod.
func (s *ModCompatibilitySystem) GetModMetadata(modID string) (*ModMetadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata, exists := s.modMetadata[modID]
	return metadata, exists
}

// GetRegisteredModCount returns the number of registered mods.
func (s *ModCompatibilitySystem) GetRegisteredModCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.modMetadata)
}
