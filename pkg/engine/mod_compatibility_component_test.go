package engine

import (
	"testing"
)

func TestNewModCompatibilityComponent(t *testing.T) {
	comp := NewModCompatibilityComponent()

	if comp == nil {
		t.Fatal("expected non-nil component")
	}
	if comp.Conflicts == nil {
		t.Error("expected non-nil Conflicts")
	}
	if comp.Dependencies == nil {
		t.Error("expected non-nil Dependencies")
	}
	if comp.LoadOrder == nil {
		t.Error("expected non-nil LoadOrder")
	}
	if comp.Warnings == nil {
		t.Error("expected non-nil Warnings")
	}
	if comp.ModVersions == nil {
		t.Error("expected non-nil ModVersions")
	}
	if comp.Configurations == nil {
		t.Error("expected non-nil Configurations")
	}
}

func TestModCompatibilityComponent_Type(t *testing.T) {
	comp := NewModCompatibilityComponent()
	if comp.Type() != "mod_compatibility" {
		t.Errorf("expected type 'mod_compatibility', got %s", comp.Type())
	}
}

func TestModCompatibilityComponent_GameVersion(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test initial state
	if comp.GetGameVersion() != "" {
		t.Error("expected empty game version initially")
	}

	// Test setting version
	comp.SetGameVersion("16.0.0")
	if comp.GetGameVersion() != "16.0.0" {
		t.Errorf("expected game version '16.0.0', got %s", comp.GetGameVersion())
	}
}

func TestModCompatibilityComponent_ModVersions(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test GetModVersion on empty
	_, exists := comp.GetModVersion("mod1")
	if exists {
		t.Error("expected mod version not to exist initially")
	}

	// Test SetModVersion
	comp.SetModVersion("mod1", "1.0.0")
	version, exists := comp.GetModVersion("mod1")
	if !exists {
		t.Error("expected mod version to exist after set")
	}
	if version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", version)
	}

	// Test RemoveModVersion
	comp.RemoveModVersion("mod1")
	_, exists = comp.GetModVersion("mod1")
	if exists {
		t.Error("expected mod version not to exist after remove")
	}
}

func TestModCompatibilityComponent_Dependencies(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test GetDependencies on empty
	deps := comp.GetDependencies("mod1")
	if deps != nil {
		t.Error("expected nil dependencies initially")
	}

	// Test SetDependencies
	comp.SetDependencies("mod1", []string{"base-mod", "lib-mod"})
	deps = comp.GetDependencies("mod1")
	if len(deps) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(deps))
	}
	if deps[0] != "base-mod" || deps[1] != "lib-mod" {
		t.Error("unexpected dependency values")
	}

	// Test SetDependencies with nil
	comp.SetDependencies("mod2", nil)
	deps = comp.GetDependencies("mod2")
	if deps == nil {
		t.Error("expected empty slice, not nil")
	}
	if len(deps) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(deps))
	}

	// Test RemoveDependencies
	comp.RemoveDependencies("mod1")
	deps = comp.GetDependencies("mod1")
	if deps != nil {
		t.Error("expected nil dependencies after remove")
	}
}

func TestModCompatibilityComponent_Conflicts(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test GetConflicts on empty
	conflicts := comp.GetConflicts()
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts initially, got %d", len(conflicts))
	}

	// Test AddConflict
	conflict1 := ModConflict{
		Mod1:         "mod1",
		Mod2:         "mod2",
		ConflictType: ConflictTypeRule,
		Description:  "Both modify combat.damage_multiplier",
		Severity:     ConflictSeverityWarning,
		Suggestion:   "Disable one mod or adjust load order",
		AffectedArea: "combat",
	}
	comp.AddConflict(conflict1)

	conflicts = comp.GetConflicts()
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}

	// Test duplicate prevention
	comp.AddConflict(conflict1)
	conflicts = comp.GetConflicts()
	if len(conflicts) != 1 {
		t.Errorf("expected duplicate to be ignored, got %d conflicts", len(conflicts))
	}

	// Test different conflict type is added
	conflict2 := ModConflict{
		Mod1:         "mod1",
		Mod2:         "mod2",
		ConflictType: ConflictTypeEvent,
		Description:  "Both handle player_death event",
		Severity:     ConflictSeverityInfo,
	}
	comp.AddConflict(conflict2)
	conflicts = comp.GetConflicts()
	if len(conflicts) != 2 {
		t.Errorf("expected 2 conflicts, got %d", len(conflicts))
	}

	// Test GetConflictCount
	if comp.GetConflictCount() != 2 {
		t.Errorf("expected conflict count 2, got %d", comp.GetConflictCount())
	}
}

func TestModCompatibilityComponent_RemoveConflict(t *testing.T) {
	comp := NewModCompatibilityComponent()

	comp.AddConflict(ModConflict{
		Mod1:         "mod1",
		Mod2:         "mod2",
		ConflictType: ConflictTypeRule,
		Severity:     ConflictSeverityWarning,
	})
	comp.AddConflict(ModConflict{
		Mod1:         "mod3",
		Mod2:         "mod4",
		ConflictType: ConflictTypeResource,
		Severity:     ConflictSeverityError,
	})

	// Test RemoveConflict
	comp.RemoveConflict("mod1", "mod2")
	conflicts := comp.GetConflicts()
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict after removal, got %d", len(conflicts))
	}
	if conflicts[0].Mod1 != "mod3" {
		t.Error("wrong conflict removed")
	}

	// Test RemoveConflict with reversed order
	comp.AddConflict(ModConflict{Mod1: "modA", Mod2: "modB", ConflictType: ConflictTypeEvent})
	comp.RemoveConflict("modB", "modA") // Reversed order
	conflictsForModA := comp.GetConflictsForMod("modA")
	if len(conflictsForModA) != 0 {
		t.Error("expected conflict to be removed with reversed mod order")
	}
}

func TestModCompatibilityComponent_GetConflictsForMod(t *testing.T) {
	comp := NewModCompatibilityComponent()

	comp.AddConflict(ModConflict{Mod1: "mod1", Mod2: "mod2", ConflictType: ConflictTypeRule})
	comp.AddConflict(ModConflict{Mod1: "mod2", Mod2: "mod3", ConflictType: ConflictTypeEvent})
	comp.AddConflict(ModConflict{Mod1: "mod4", Mod2: "mod5", ConflictType: ConflictTypeResource})

	// Test GetConflictsForMod
	conflicts := comp.GetConflictsForMod("mod2")
	if len(conflicts) != 2 {
		t.Errorf("expected 2 conflicts for mod2, got %d", len(conflicts))
	}

	conflicts = comp.GetConflictsForMod("mod4")
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict for mod4, got %d", len(conflicts))
	}

	conflicts = comp.GetConflictsForMod("nonexistent")
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts for nonexistent mod, got %d", len(conflicts))
	}
}

func TestModCompatibilityComponent_GetConflictsBySeverity(t *testing.T) {
	comp := NewModCompatibilityComponent()

	comp.AddConflict(ModConflict{Mod1: "mod1", Mod2: "mod2", ConflictType: ConflictTypeRule, Severity: ConflictSeverityError})
	comp.AddConflict(ModConflict{Mod1: "mod3", Mod2: "mod4", ConflictType: ConflictTypeEvent, Severity: ConflictSeverityWarning})
	comp.AddConflict(ModConflict{Mod1: "mod5", Mod2: "mod6", ConflictType: ConflictTypeOverride, Severity: ConflictSeverityError})
	comp.AddConflict(ModConflict{Mod1: "mod7", Mod2: "mod8", ConflictType: ConflictTypeVersion, Severity: ConflictSeverityInfo})

	errors := comp.GetConflictsBySeverity(ConflictSeverityError)
	if len(errors) != 2 {
		t.Errorf("expected 2 error conflicts, got %d", len(errors))
	}

	warnings := comp.GetConflictsBySeverity(ConflictSeverityWarning)
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning conflict, got %d", len(warnings))
	}

	infos := comp.GetConflictsBySeverity(ConflictSeverityInfo)
	if len(infos) != 1 {
		t.Errorf("expected 1 info conflict, got %d", len(infos))
	}
}

func TestModCompatibilityComponent_HasBlockingConflicts(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// No conflicts
	if comp.HasBlockingConflicts() {
		t.Error("expected no blocking conflicts initially")
	}

	// Add warning-level conflict
	comp.AddConflict(ModConflict{
		Mod1:     "mod1",
		Mod2:     "mod2",
		Severity: ConflictSeverityWarning,
	})
	if comp.HasBlockingConflicts() {
		t.Error("expected no blocking conflicts with only warnings")
	}

	// Add error-level conflict
	comp.AddConflict(ModConflict{
		Mod1:         "mod3",
		Mod2:         "mod4",
		ConflictType: ConflictTypeVersion,
		Severity:     ConflictSeverityError,
	})
	if !comp.HasBlockingConflicts() {
		t.Error("expected blocking conflicts with error present")
	}
}

func TestModCompatibilityComponent_ClearConflicts(t *testing.T) {
	comp := NewModCompatibilityComponent()

	comp.AddConflict(ModConflict{Mod1: "mod1", Mod2: "mod2"})
	comp.AddConflict(ModConflict{Mod1: "mod3", Mod2: "mod4"})

	comp.ClearConflicts()

	if comp.GetConflictCount() != 0 {
		t.Errorf("expected 0 conflicts after clear, got %d", comp.GetConflictCount())
	}
}

func TestModCompatibilityComponent_Warnings(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test GetWarnings on empty
	warnings := comp.GetWarnings()
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings initially, got %d", len(warnings))
	}

	// Test AddWarning
	warning := CompatibilityWarning{
		ModID:       "mod1",
		WarningType: "deprecated_api",
		Message:     "Mod uses deprecated API",
		Suggestion:  "Update to latest version",
	}
	comp.AddWarning(warning)

	warnings = comp.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].ModID != "mod1" {
		t.Error("unexpected warning values")
	}

	// Test GetWarningCount
	if comp.GetWarningCount() != 1 {
		t.Errorf("expected warning count 1, got %d", comp.GetWarningCount())
	}

	// Test GetWarningsForMod
	comp.AddWarning(CompatibilityWarning{ModID: "mod2", WarningType: "performance"})
	comp.AddWarning(CompatibilityWarning{ModID: "mod1", WarningType: "compatibility"})

	mod1Warnings := comp.GetWarningsForMod("mod1")
	if len(mod1Warnings) != 2 {
		t.Errorf("expected 2 warnings for mod1, got %d", len(mod1Warnings))
	}

	// Test ClearWarnings
	comp.ClearWarnings()
	if comp.GetWarningCount() != 0 {
		t.Errorf("expected 0 warnings after clear, got %d", comp.GetWarningCount())
	}
}

func TestModCompatibilityComponent_LoadOrder(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test GetLoadOrder on empty
	order := comp.GetLoadOrder()
	if len(order) != 0 {
		t.Errorf("expected empty load order initially, got %d", len(order))
	}

	// Test SetLoadOrder
	comp.SetLoadOrder([]string{"base-mod", "lib-mod", "content-mod"})
	order = comp.GetLoadOrder()
	if len(order) != 3 {
		t.Errorf("expected 3 mods in load order, got %d", len(order))
	}
	if order[0] != "base-mod" || order[1] != "lib-mod" || order[2] != "content-mod" {
		t.Error("unexpected load order values")
	}

	// Test GetLoadPosition
	pos := comp.GetLoadPosition("lib-mod")
	if pos != 1 {
		t.Errorf("expected position 1 for lib-mod, got %d", pos)
	}

	pos = comp.GetLoadPosition("nonexistent")
	if pos != -1 {
		t.Errorf("expected position -1 for nonexistent mod, got %d", pos)
	}
}

func TestModCompatibilityComponent_Configurations(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test ListConfigurations on empty
	configs := comp.ListConfigurations()
	if len(configs) != 0 {
		t.Errorf("expected 0 configurations initially, got %d", len(configs))
	}

	// Test SaveConfiguration
	config1 := ModConfig2{
		ID:          "config1",
		Name:        "Vanilla Plus",
		Description: "Vanilla with QoL mods",
		EnabledMods: []string{"qol-mod", "ui-mod"},
		LoadOrder:   []string{"qol-mod", "ui-mod"},
		CreatedAt:   1000,
		UpdatedAt:   1000,
	}
	comp.SaveConfiguration(config1)

	// Test GetConfiguration
	config, exists := comp.GetConfiguration("config1")
	if !exists {
		t.Error("expected configuration to exist")
	}
	if config.Name != "Vanilla Plus" {
		t.Errorf("expected name 'Vanilla Plus', got %s", config.Name)
	}
	if len(config.EnabledMods) != 2 {
		t.Errorf("expected 2 enabled mods, got %d", len(config.EnabledMods))
	}

	// Test GetConfiguration for nonexistent
	_, exists = comp.GetConfiguration("nonexistent")
	if exists {
		t.Error("expected nonexistent configuration not to exist")
	}

	// Test ListConfigurations sorting
	config2 := ModConfig2{ID: "config2", Name: "Hardcore"}
	comp.SaveConfiguration(config2)
	configs = comp.ListConfigurations()
	if len(configs) != 2 {
		t.Errorf("expected 2 configurations, got %d", len(configs))
	}
	if configs[0].Name != "Hardcore" {
		t.Error("expected configurations sorted by name")
	}

	// Test DeleteConfiguration
	comp.DeleteConfiguration("config1")
	_, exists = comp.GetConfiguration("config1")
	if exists {
		t.Error("expected configuration to be deleted")
	}
}

func TestModCompatibilityComponent_ActiveConfiguration(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test GetActiveConfiguration on empty
	if comp.GetActiveConfiguration() != "" {
		t.Error("expected empty active configuration initially")
	}

	// Test SetActiveConfiguration
	comp.SetActiveConfiguration("config1")
	if comp.GetActiveConfiguration() != "config1" {
		t.Errorf("expected active configuration 'config1', got %s", comp.GetActiveConfiguration())
	}
}

func TestModCompatibilityComponent_GetErrorCount(t *testing.T) {
	comp := NewModCompatibilityComponent()

	if comp.GetErrorCount() != 0 {
		t.Errorf("expected 0 errors initially, got %d", comp.GetErrorCount())
	}

	comp.AddConflict(ModConflict{Mod1: "m1", Mod2: "m2", Severity: ConflictSeverityError})
	comp.AddConflict(ModConflict{Mod1: "m3", Mod2: "m4", Severity: ConflictSeverityWarning})
	comp.AddConflict(ModConflict{Mod1: "m5", Mod2: "m6", Severity: ConflictSeverityError})

	if comp.GetErrorCount() != 2 {
		t.Errorf("expected 2 errors, got %d", comp.GetErrorCount())
	}
}

func TestModCompatibilityComponent_Serialization(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Set up state
	comp.SetGameVersion("16.0.0")
	comp.SetModVersion("mod1", "1.0.0")
	comp.SetDependencies("mod1", []string{"base"})
	comp.AddConflict(ModConflict{
		Mod1:         "mod1",
		Mod2:         "mod2",
		ConflictType: ConflictTypeRule,
		Description:  "Test conflict",
		Severity:     ConflictSeverityWarning,
	})
	comp.AddWarning(CompatibilityWarning{
		ModID:   "mod1",
		Message: "Test warning",
	})
	comp.SetLoadOrder([]string{"base", "mod1", "mod2"})
	comp.SaveConfiguration(ModConfig2{
		ID:          "config1",
		Name:        "Test Config",
		EnabledMods: []string{"mod1"},
	})
	comp.SetActiveConfiguration("config1")

	// Serialize
	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	// Create new component and deserialize
	comp2 := NewModCompatibilityComponent()
	err = comp2.Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error deserializing: %v", err)
	}

	// Verify state
	if comp2.GetGameVersion() != "16.0.0" {
		t.Errorf("expected game version '16.0.0', got %s", comp2.GetGameVersion())
	}

	version, exists := comp2.GetModVersion("mod1")
	if !exists || version != "1.0.0" {
		t.Error("mod version not restored correctly")
	}

	deps := comp2.GetDependencies("mod1")
	if len(deps) != 1 || deps[0] != "base" {
		t.Error("dependencies not restored correctly")
	}

	conflicts := comp2.GetConflicts()
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].Description != "Test conflict" {
		t.Error("conflict description not restored")
	}

	warnings := comp2.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}

	order := comp2.GetLoadOrder()
	if len(order) != 3 {
		t.Errorf("expected 3 in load order, got %d", len(order))
	}

	config, exists := comp2.GetConfiguration("config1")
	if !exists || config.Name != "Test Config" {
		t.Error("configuration not restored correctly")
	}

	if comp2.GetActiveConfiguration() != "config1" {
		t.Error("active configuration not restored")
	}
}

func TestModCompatibilityComponent_DeserializeEmpty(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Deserialize empty JSON object
	err := comp.Deserialize([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error deserializing empty: %v", err)
	}

	// Verify defaults
	if comp.Conflicts == nil {
		t.Error("expected non-nil Conflicts after deserialize")
	}
	if comp.Dependencies == nil {
		t.Error("expected non-nil Dependencies after deserialize")
	}
	if comp.LoadOrder == nil {
		t.Error("expected non-nil LoadOrder after deserialize")
	}
}

func TestModCompatibilityComponent_DeserializeInvalid(t *testing.T) {
	comp := NewModCompatibilityComponent()

	err := comp.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("expected error deserializing invalid JSON")
	}
}

func TestConflictTypes(t *testing.T) {
	// Verify conflict type constants
	types := []ConflictType{
		ConflictTypeRule,
		ConflictTypeEvent,
		ConflictTypeResource,
		ConflictTypeOverride,
		ConflictTypeVersion,
	}

	for _, ct := range types {
		if ct == "" {
			t.Error("conflict type should not be empty")
		}
	}
}

func TestConflictSeverities(t *testing.T) {
	// Verify severity constants
	severities := []ConflictSeverity{
		ConflictSeverityError,
		ConflictSeverityWarning,
		ConflictSeverityInfo,
	}

	for _, s := range severities {
		if s == "" {
			t.Error("severity should not be empty")
		}
	}
}

func TestModCompatibilityComponent_ConcurrentAccess(t *testing.T) {
	comp := NewModCompatibilityComponent()

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			comp.AddConflict(ModConflict{
				Mod1: "mod1",
				Mod2: "mod2",
			})
			comp.SetModVersion("mod1", "1.0.0")
			comp.SetDependencies("mod1", []string{"base"})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = comp.GetConflicts()
			_, _ = comp.GetModVersion("mod1")
			_ = comp.GetDependencies("mod1")
		}
		done <- true
	}()

	// Wait for both
	<-done
	<-done
}
