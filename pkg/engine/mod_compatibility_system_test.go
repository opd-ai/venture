package engine

import (
	"fmt"
	"testing"
)

func TestNewModCompatibilitySystem(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.world != world {
		t.Error("expected world to be set")
	}
	if sys.ruleRegistry == nil {
		t.Error("expected non-nil ruleRegistry")
	}
	if sys.eventRegistry == nil {
		t.Error("expected non-nil eventRegistry")
	}
	if sys.resourceRegistry == nil {
		t.Error("expected non-nil resourceRegistry")
	}
	if sys.modMetadata == nil {
		t.Error("expected non-nil modMetadata")
	}
}

func TestModCompatibilitySystem_RegisterMod(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)

	// Test nil metadata
	err := sys.RegisterMod(nil)
	if err == nil {
		t.Error("expected error for nil metadata")
	}

	// Test empty ID
	err = sys.RegisterMod(&ModMetadata{})
	if err == nil {
		t.Error("expected error for empty ID")
	}

	// Test valid registration
	metadata := &ModMetadata{
		ID:           "test-mod",
		Version:      "1.0.0",
		Rules:        []string{"combat.damage", "combat.speed"},
		Events:       []string{"player_death"},
		Resources:    []string{"sprites/player"},
		Dependencies: []string{"base-mod"},
	}
	err = sys.RegisterMod(metadata)
	if err != nil {
		t.Errorf("unexpected error registering mod: %v", err)
	}

	// Verify registration
	if sys.GetRegisteredModCount() != 1 {
		t.Errorf("expected 1 registered mod, got %d", sys.GetRegisteredModCount())
	}

	// Verify metadata retrieval
	retrieved, exists := sys.GetModMetadata("test-mod")
	if !exists {
		t.Error("expected mod metadata to exist")
	}
	if retrieved.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", retrieved.Version)
	}
}

func TestModCompatibilitySystem_UnregisterMod(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)

	// Register mod
	metadata := &ModMetadata{
		ID:        "test-mod",
		Rules:     []string{"combat.damage"},
		Events:    []string{"player_death"},
		Resources: []string{"sprites/player"},
	}
	sys.RegisterMod(metadata)

	// Unregister
	sys.UnregisterMod("test-mod")

	// Verify removal
	if sys.GetRegisteredModCount() != 0 {
		t.Errorf("expected 0 registered mods, got %d", sys.GetRegisteredModCount())
	}

	_, exists := sys.GetModMetadata("test-mod")
	if exists {
		t.Error("expected mod metadata to be removed")
	}

	// Unregister nonexistent (should not panic)
	sys.UnregisterMod("nonexistent")
}

func TestModCompatibilitySystem_ValidateMods_KnownConflicts(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register two mods that declare incompatibility
	sys.RegisterMod(&ModMetadata{
		ID:        "mod-a",
		Conflicts: []string{"mod-b"},
	})
	sys.RegisterMod(&ModMetadata{
		ID: "mod-b",
	})

	// Validate
	sys.ValidateMods(comp, []string{"mod-a", "mod-b"})

	// Should have error-level conflict
	if !comp.HasBlockingConflicts() {
		t.Error("expected blocking conflict for declared incompatibility")
	}

	conflicts := comp.GetConflicts()
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].ConflictType != ConflictTypeOverride {
		t.Errorf("expected ConflictTypeOverride, got %s", conflicts[0].ConflictType)
	}
}

func TestModCompatibilitySystem_ValidateMods_RuleConflicts(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register two mods that modify the same rule
	sys.RegisterMod(&ModMetadata{
		ID:    "mod-a",
		Rules: []string{"combat.damage_multiplier"},
	})
	sys.RegisterMod(&ModMetadata{
		ID:    "mod-b",
		Rules: []string{"combat.damage_multiplier"},
	})

	// Validate
	sys.ValidateMods(comp, []string{"mod-a", "mod-b"})

	// Should have warning-level conflict
	conflicts := comp.GetConflictsBySeverity(ConflictSeverityWarning)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 warning conflict, got %d", len(conflicts))
	}
	if conflicts[0].ConflictType != ConflictTypeRule {
		t.Errorf("expected ConflictTypeRule, got %s", conflicts[0].ConflictType)
	}
	if conflicts[0].AffectedArea != "combat.damage_multiplier" {
		t.Errorf("expected affected area 'combat.damage_multiplier', got %s", conflicts[0].AffectedArea)
	}
}

func TestModCompatibilitySystem_ValidateMods_EventConflicts(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register two mods that handle the same event
	sys.RegisterMod(&ModMetadata{
		ID:     "mod-a",
		Events: []string{"player_death"},
	})
	sys.RegisterMod(&ModMetadata{
		ID:     "mod-b",
		Events: []string{"player_death"},
	})

	// Validate
	sys.ValidateMods(comp, []string{"mod-a", "mod-b"})

	// Should have info-level conflict (not blocking)
	conflicts := comp.GetConflictsBySeverity(ConflictSeverityInfo)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 info conflict, got %d", len(conflicts))
	}
	if conflicts[0].ConflictType != ConflictTypeEvent {
		t.Errorf("expected ConflictTypeEvent, got %s", conflicts[0].ConflictType)
	}
}

func TestModCompatibilitySystem_ValidateMods_ResourceConflicts(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register two mods that modify the same resource
	sys.RegisterMod(&ModMetadata{
		ID:        "mod-a",
		Resources: []string{"sprites/player"},
	})
	sys.RegisterMod(&ModMetadata{
		ID:        "mod-b",
		Resources: []string{"sprites/player"},
	})

	// Validate
	sys.ValidateMods(comp, []string{"mod-a", "mod-b"})

	conflicts := comp.GetConflictsBySeverity(ConflictSeverityWarning)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 warning conflict, got %d", len(conflicts))
	}
	if conflicts[0].ConflictType != ConflictTypeResource {
		t.Errorf("expected ConflictTypeResource, got %s", conflicts[0].ConflictType)
	}
}

func TestModCompatibilitySystem_ValidateMods_VersionConflicts(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()
	comp.SetGameVersion("15.0.0")

	// Register mod requiring newer game version
	sys.RegisterMod(&ModMetadata{
		ID:             "mod-a",
		MinGameVersion: "16.0.0",
	})

	// Register mod made for older version
	sys.RegisterMod(&ModMetadata{
		ID:             "mod-b",
		MaxGameVersion: "14.0.0",
	})

	// Validate
	sys.ValidateMods(comp, []string{"mod-a", "mod-b"})

	// Should have error for min version
	conflicts := comp.GetConflictsBySeverity(ConflictSeverityError)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 error conflict, got %d", len(conflicts))
	}
	if conflicts[0].ConflictType != ConflictTypeVersion {
		t.Errorf("expected ConflictTypeVersion, got %s", conflicts[0].ConflictType)
	}

	// Should have warning for max version
	warnings := comp.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].WarningType != "version" {
		t.Errorf("expected warning type 'version', got %s", warnings[0].WarningType)
	}
}

func TestModCompatibilitySystem_ValidateMods_MissingDependencies(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register mod with dependency
	sys.RegisterMod(&ModMetadata{
		ID:           "mod-a",
		Dependencies: []string{"base-mod"},
	})

	// Validate without base-mod
	sys.ValidateMods(comp, []string{"mod-a"})

	// Should have error for missing dependency
	conflicts := comp.GetConflictsBySeverity(ConflictSeverityError)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 error conflict, got %d", len(conflicts))
	}
	if conflicts[0].Mod2 != "base-mod" {
		t.Errorf("expected missing dependency 'base-mod', got %s", conflicts[0].Mod2)
	}
}

func TestModCompatibilitySystem_CalculateLoadOrder(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register mods with dependencies
	sys.RegisterMod(&ModMetadata{ID: "base-mod"})
	sys.RegisterMod(&ModMetadata{
		ID:           "mod-a",
		Dependencies: []string{"base-mod"},
	})
	sys.RegisterMod(&ModMetadata{
		ID:           "mod-b",
		Dependencies: []string{"base-mod"},
	})
	sys.RegisterMod(&ModMetadata{
		ID:           "mod-c",
		Dependencies: []string{"mod-a", "mod-b"},
	})

	// Calculate load order
	order, err := sys.CalculateLoadOrder(comp, []string{"mod-c", "mod-a", "mod-b", "base-mod"})
	if err != nil {
		t.Fatalf("unexpected error calculating load order: %v", err)
	}

	// Verify order
	if len(order) != 4 {
		t.Errorf("expected 4 mods in load order, got %d", len(order))
	}

	// base-mod must come first
	if order[0] != "base-mod" {
		t.Errorf("expected base-mod first, got %s", order[0])
	}

	// mod-c must come last
	if order[3] != "mod-c" {
		t.Errorf("expected mod-c last, got %s", order[3])
	}

	// Verify component was updated
	compOrder := comp.GetLoadOrder()
	if len(compOrder) != 4 {
		t.Errorf("expected component load order to be set")
	}
}

func TestModCompatibilitySystem_CalculateLoadOrder_CircularDependency(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register mods with circular dependency
	sys.RegisterMod(&ModMetadata{
		ID:           "mod-a",
		Dependencies: []string{"mod-b"},
	})
	sys.RegisterMod(&ModMetadata{
		ID:           "mod-b",
		Dependencies: []string{"mod-a"},
	})

	// Should return error
	_, err := sys.CalculateLoadOrder(comp, []string{"mod-a", "mod-b"})
	if err == nil {
		t.Error("expected error for circular dependency")
	}
}

func TestModCompatibilitySystem_ExportConfiguration(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	modIDs := []string{"mod-a", "mod-b", "mod-c"}
	config := sys.ExportConfiguration(comp, "Test Config", "For testing", modIDs)

	if config.Name != "Test Config" {
		t.Errorf("expected name 'Test Config', got %s", config.Name)
	}
	if config.Description != "For testing" {
		t.Errorf("expected description 'For testing', got %s", config.Description)
	}
	if len(config.EnabledMods) != 3 {
		t.Errorf("expected 3 enabled mods, got %d", len(config.EnabledMods))
	}
	if config.CreatedAt == 0 {
		t.Error("expected CreatedAt to be set")
	}

	// Verify it was saved
	_, exists := comp.GetConfiguration(config.ID)
	if !exists {
		t.Error("expected configuration to be saved")
	}
}

func TestModCompatibilitySystem_ImportConfiguration(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register compatible mods
	sys.RegisterMod(&ModMetadata{ID: "mod-a"})
	sys.RegisterMod(&ModMetadata{ID: "mod-b"})

	// Save configuration
	config := ModConfig2{
		ID:          "config1",
		Name:        "Test",
		EnabledMods: []string{"mod-a", "mod-b"},
		LoadOrder:   []string{"mod-a", "mod-b"},
	}
	comp.SaveConfiguration(config)

	// Import
	mods, err := sys.ImportConfiguration(comp, "config1")
	if err != nil {
		t.Fatalf("unexpected error importing configuration: %v", err)
	}
	if len(mods) != 2 {
		t.Errorf("expected 2 mods, got %d", len(mods))
	}

	// Verify active configuration was set
	if comp.GetActiveConfiguration() != "config1" {
		t.Error("expected active configuration to be set")
	}
}

func TestModCompatibilitySystem_ImportConfiguration_NotFound(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	_, err := sys.ImportConfiguration(comp, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent configuration")
	}
}

func TestModCompatibilitySystem_ImportConfiguration_WithConflicts(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register conflicting mods
	sys.RegisterMod(&ModMetadata{
		ID:        "mod-a",
		Conflicts: []string{"mod-b"},
	})
	sys.RegisterMod(&ModMetadata{ID: "mod-b"})

	// Save configuration with conflicts
	config := ModConfig2{
		ID:          "config1",
		Name:        "Test",
		EnabledMods: []string{"mod-a", "mod-b"},
	}
	comp.SaveConfiguration(config)

	// Import should fail
	_, err := sys.ImportConfiguration(comp, "config1")
	if err == nil {
		t.Error("expected error for configuration with blocking conflicts")
	}
}

func TestModCompatibilitySystem_CheckModCompatibility(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()
	comp.SetGameVersion("16.0.0")

	// Register mods
	sys.RegisterMod(&ModMetadata{
		ID:             "base-mod",
		MinGameVersion: "15.0.0",
	})
	sys.RegisterMod(&ModMetadata{
		ID:           "dependent-mod",
		Dependencies: []string{"base-mod"},
	})
	sys.RegisterMod(&ModMetadata{
		ID:        "conflicting-mod",
		Conflicts: []string{"base-mod"},
	})

	// Test compatible mod
	ok, issues := sys.CheckModCompatibility(comp, "base-mod", []string{})
	if !ok {
		t.Error("expected base-mod to be compatible")
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}

	// Test mod with missing dependency
	ok, issues = sys.CheckModCompatibility(comp, "dependent-mod", []string{})
	if ok {
		t.Error("expected dependent-mod to be incompatible without base-mod")
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
	}

	// Test mod with dependency present
	ok, issues = sys.CheckModCompatibility(comp, "dependent-mod", []string{"base-mod"})
	if !ok {
		t.Error("expected dependent-mod to be compatible with base-mod")
	}

	// Test conflicting mod
	ok, issues = sys.CheckModCompatibility(comp, "conflicting-mod", []string{"base-mod"})
	if ok {
		t.Error("expected conflicting-mod to be incompatible with base-mod")
	}
}

func TestModCompatibilitySystem_GetRecommendedResolutions(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Add some conflicts with suggestions
	comp.AddConflict(ModConflict{
		Mod1:       "mod-a",
		Mod2:       "mod-b",
		Suggestion: "Disable one of the mods",
	})
	comp.AddConflict(ModConflict{
		Mod1:       "mod-c",
		Mod2:       "mod-d",
		Suggestion: "Update mod-c to latest version",
	})
	comp.AddConflict(ModConflict{
		Mod1:       "mod-e",
		Mod2:       "mod-f",
		Suggestion: "Disable one of the mods", // Duplicate
	})

	suggestions := sys.GetRecommendedResolutions(comp)
	if len(suggestions) != 2 { // Should deduplicate
		t.Errorf("expected 2 unique suggestions, got %d", len(suggestions))
	}
}

func TestCompareModVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.1", "1.0.0", 1},
		{"10.0.0", "9.0.0", 1},
		{"1.10.0", "1.9.0", 1},
		{"1.0", "1.0.0", 0},
		{"16.0.0", "15.0.0", 1},
	}

	for _, tt := range tests {
		result := compareModVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("compareVersions(%s, %s) = %d, expected %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestModCompatibilitySystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)

	// Create entity with compatibility component
	entity := NewEntity(123)
	entity.AddComponent(NewModCompatibilityComponent())

	// Update should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestModCompatibilitySystem_ValidateMods_NoGameVersion(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register mod with version requirements
	sys.RegisterMod(&ModMetadata{
		ID:             "mod-a",
		MinGameVersion: "16.0.0",
	})

	// Validate without game version set
	sys.ValidateMods(comp, []string{"mod-a"})

	// Should not produce version conflicts when game version is unknown
	conflicts := comp.GetConflicts()
	for _, c := range conflicts {
		if c.ConflictType == ConflictTypeVersion && c.Mod2 == "game" {
			t.Error("should not check version when game version is not set")
		}
	}
}

func TestModCompatibilitySystem_ValidateMods_ClearsOldData(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Add some old conflicts and warnings
	comp.AddConflict(ModConflict{Mod1: "old1", Mod2: "old2"})
	comp.AddWarning(CompatibilityWarning{ModID: "old1"})

	// Validate with no mods (should clear)
	sys.ValidateMods(comp, []string{})

	if comp.GetConflictCount() != 0 {
		t.Error("expected conflicts to be cleared")
	}
	if comp.GetWarningCount() != 0 {
		t.Error("expected warnings to be cleared")
	}
}

func TestModCompatibilitySystem_CalculateLoadOrder_NoDependencies(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register mods without dependencies
	sys.RegisterMod(&ModMetadata{ID: "mod-c"})
	sys.RegisterMod(&ModMetadata{ID: "mod-a"})
	sys.RegisterMod(&ModMetadata{ID: "mod-b"})

	// Calculate load order
	order, err := sys.CalculateLoadOrder(comp, []string{"mod-c", "mod-a", "mod-b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be sorted alphabetically for determinism
	if order[0] != "mod-a" || order[1] != "mod-b" || order[2] != "mod-c" {
		t.Errorf("expected alphabetical order, got %v", order)
	}
}

func TestModCompatibilitySystem_ExportConfiguration_WithLoadOrder(t *testing.T) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Set custom load order
	comp.SetLoadOrder([]string{"base", "mod-a", "mod-b"})

	modIDs := []string{"mod-a", "mod-b", "base"}
	config := sys.ExportConfiguration(comp, "Test", "Desc", modIDs)

	// Should use the existing load order
	if config.LoadOrder[0] != "base" {
		t.Error("expected export to use existing load order")
	}
}

func BenchmarkModCompatibilitySystem_ValidateMods(b *testing.B) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()
	comp.SetGameVersion("16.0.0")

	// Register many mods
	for i := 0; i < 100; i++ {
		sys.RegisterMod(&ModMetadata{
			ID:    fmt.Sprintf("mod-%d", i),
			Rules: []string{fmt.Sprintf("rule-%d", i%10)},
		})
	}

	modIDs := make([]string, 50)
	for i := 0; i < 50; i++ {
		modIDs[i] = fmt.Sprintf("mod-%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ValidateMods(comp, modIDs)
	}
}

func BenchmarkModCompatibilitySystem_CalculateLoadOrder(b *testing.B) {
	world := NewWorld()
	sys := NewModCompatibilitySystem(world)
	comp := NewModCompatibilityComponent()

	// Register mods with chain dependencies
	sys.RegisterMod(&ModMetadata{ID: "base"})
	for i := 1; i < 50; i++ {
		sys.RegisterMod(&ModMetadata{
			ID:           fmt.Sprintf("mod-%d", i),
			Dependencies: []string{fmt.Sprintf("mod-%d", i-1)},
		})
	}
	sys.RegisterMod(&ModMetadata{
		ID:           "mod-0",
		Dependencies: []string{"base"},
	})

	modIDs := make([]string, 51)
	modIDs[0] = "base"
	for i := 0; i < 50; i++ {
		modIDs[i+1] = fmt.Sprintf("mod-%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.CalculateLoadOrder(comp, modIDs)
	}
}
