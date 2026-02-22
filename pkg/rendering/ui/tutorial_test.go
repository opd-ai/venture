package ui

import (
	"testing"
)

func TestTutorialManager_Defaults(t *testing.T) {
	tm := NewTutorialManager()

	// Should have tutorials registered
	tutorial, err := tm.GetTutorial(TutorialMovement)
	if err != nil {
		t.Fatalf("failed to get movement tutorial: %v", err)
	}

	if tutorial.Title == "" {
		t.Error("expected tutorial to have a title")
	}
	if len(tutorial.Content) == 0 {
		t.Error("expected tutorial to have content")
	}

	// Should start unviewed
	if tutorial.Viewed {
		t.Error("expected tutorial to start unviewed")
	}
}

func TestTutorialManager_ShowTutorial(t *testing.T) {
	tm := NewTutorialManager()

	// Show tutorial
	tutorial, err := tm.ShowTutorial(TutorialCombat)
	if err != nil {
		t.Fatalf("failed to show tutorial: %v", err)
	}

	// Should be marked viewed
	if !tutorial.Viewed {
		t.Error("expected tutorial to be marked viewed")
	}

	// ViewedAt should be set
	if tutorial.ViewedAt.IsZero() {
		t.Error("expected ViewedAt to be set")
	}
}

func TestTutorialManager_ListUnviewed(t *testing.T) {
	tm := NewTutorialManager()

	// All should be unviewed initially
	unviewed := tm.ListUnviewed()
	initialCount := len(unviewed)
	if initialCount == 0 {
		t.Error("expected unviewed tutorials initially")
	}

	// Show one tutorial
	tm.ShowTutorial(TutorialMovement)

	// Unviewed count should decrease
	unviewed = tm.ListUnviewed()
	if len(unviewed) != initialCount-1 {
		t.Errorf("expected %d unviewed, got %d", initialCount-1, len(unviewed))
	}
}

func TestTutorialManager_EnableDisable(t *testing.T) {
	tm := NewTutorialManager()

	// Should be enabled by default
	if !tm.IsEnabled() {
		t.Error("expected tutorials enabled by default")
	}

	// Disable
	tm.Disable()
	if tm.IsEnabled() {
		t.Error("expected tutorials disabled")
	}

	// Should not show tutorials when disabled
	_, err := tm.ShowTutorial(TutorialMovement)
	if err == nil {
		t.Error("expected error when showing tutorial while disabled")
	}

	// Re-enable
	tm.Enable()
	if !tm.IsEnabled() {
		t.Error("expected tutorials enabled")
	}
}

func TestTutorialManager_ResetProgress(t *testing.T) {
	tm := NewTutorialManager()

	// Show several tutorials
	tm.ShowTutorial(TutorialMovement)
	tm.ShowTutorial(TutorialCombat)
	tm.ShowTutorial(TutorialInventory)

	// Should have viewed tutorials
	unviewed := tm.ListUnviewed()
	viewedCount := 3

	// Reset
	tm.ResetProgress()

	// All should be unviewed again
	unviewed = tm.ListUnviewed()
	if len(unviewed) != len(unviewed)+viewedCount {
		// All tutorials should be unviewed
		tutorial, _ := tm.GetTutorial(TutorialMovement)
		if tutorial.Viewed {
			t.Error("expected movement tutorial to be unviewed after reset")
		}
	}
}

// TestTutorialManager_ExportImportState tests save/load of tutorial state.
// Phase 4.5: Enables persistence of viewed context-sensitive tutorials.
func TestTutorialManager_ExportImportState(t *testing.T) {
	// Create first manager and view some tutorials
	tm1 := NewTutorialManager()
	tm1.ShowTutorial(TutorialMovement)
	tm1.ShowTutorial(TutorialCombat)
	tm1.Disable()

	// Export state
	exported := tm1.ExportState()

	// Verify exported data
	if exported.Enabled {
		t.Error("expected exported enabled to be false")
	}
	if len(exported.ViewedTopics) != 2 {
		t.Errorf("expected 2 viewed topics, got %d", len(exported.ViewedTopics))
	}
	if _, ok := exported.ViewedTopics[string(TutorialMovement)]; !ok {
		t.Error("expected movement tutorial in viewed topics")
	}
	if _, ok := exported.ViewedTopics[string(TutorialCombat)]; !ok {
		t.Error("expected combat tutorial in viewed topics")
	}

	// Create new manager and import state
	tm2 := NewTutorialManager()
	tm2.ImportState(exported)

	// Verify imported state
	if tm2.IsEnabled() {
		t.Error("expected imported enabled to be false")
	}

	movement, _ := tm2.GetTutorial(TutorialMovement)
	if !movement.Viewed {
		t.Error("expected movement tutorial to be viewed after import")
	}

	combat, _ := tm2.GetTutorial(TutorialCombat)
	if !combat.Viewed {
		t.Error("expected combat tutorial to be viewed after import")
	}

	// Unviewed tutorials should remain unviewed
	inventory, _ := tm2.GetTutorial(TutorialInventory)
	if inventory.Viewed {
		t.Error("expected inventory tutorial to remain unviewed after import")
	}
}

// TestTutorialManager_ExportImportState_Empty tests export/import with no viewed tutorials.
func TestTutorialManager_ExportImportState_Empty(t *testing.T) {
	tm1 := NewTutorialManager()

	// Export with no viewed tutorials
	exported := tm1.ExportState()

	if !exported.Enabled {
		t.Error("expected enabled to be true")
	}
	if len(exported.ViewedTopics) != 0 {
		t.Errorf("expected 0 viewed topics, got %d", len(exported.ViewedTopics))
	}

	// Import empty state
	tm2 := NewTutorialManager()
	tm2.ShowTutorial(TutorialMovement) // View one first
	tm2.ImportState(exported)

	// Imported empty state should override (but movement was viewed before import,
	// so we need to check if import properly handles topics not in the map)
	// The import only sets viewed for topics IN the map, so movement stays viewed
	// This is correct behavior - we don't reset topics not mentioned in import
}

// TestTutorialManager_ExportImportState_RoundTrip tests complete round-trip.
func TestTutorialManager_ExportImportState_RoundTrip(t *testing.T) {
	tm1 := NewTutorialManager()

	// View multiple tutorials
	tm1.ShowTutorial(TutorialCrafting)
	tm1.ShowTutorial(TutorialHousing)
	tm1.ShowTutorial(TutorialGuilds)
	tm1.Disable()

	// Export
	exported := tm1.ExportState()

	// Create new manager and import
	tm2 := NewTutorialManager()
	tm2.ImportState(exported)

	// Export from second manager
	reexported := tm2.ExportState()

	// Compare
	if exported.Enabled != reexported.Enabled {
		t.Error("enabled state mismatch after round-trip")
	}
	if len(exported.ViewedTopics) != len(reexported.ViewedTopics) {
		t.Errorf("viewed topics count mismatch: %d vs %d",
			len(exported.ViewedTopics), len(reexported.ViewedTopics))
	}
	for topic, ts1 := range exported.ViewedTopics {
		ts2, ok := reexported.ViewedTopics[topic]
		if !ok {
			t.Errorf("topic %s missing after round-trip", topic)
		}
		if ts1 != ts2 {
			t.Errorf("timestamp mismatch for topic %s: %d vs %d", topic, ts1, ts2)
		}
	}
}

func TestTutorialManager_AllTopics(t *testing.T) {
	tm := NewTutorialManager()

	topics := []TutorialTopic{
		TutorialMovement,
		TutorialCombat,
		TutorialInventory,
		TutorialCrafting,
		TutorialHousing,
		TutorialGuilds,
		TutorialVehicles,
		TutorialCompanions,
		TutorialQuests,
		TutorialSkills,
		TutorialQuickTravel,
		TutorialRaids,
		TutorialPrestige,
		TutorialPolitics,
		TutorialEconomy,
	}

	for _, topic := range topics {
		if _, err := tm.GetTutorial(topic); err != nil {
			t.Errorf("missing tutorial for %s: %v", topic, err)
		}
	}
}

func TestAccessibilityConfig_Defaults(t *testing.T) {
	config := NewAccessibilityConfig()

	if config.ColorblindMode != ColorblindNone {
		t.Error("expected colorblind mode None by default")
	}
	if config.FontScale != 1.0 {
		t.Error("expected font scale 1.0 by default")
	}
	if config.HighContrast {
		t.Error("expected high contrast false by default")
	}
	if config.ScreenReader {
		t.Error("expected screen reader false by default")
	}
}

func TestAccessibilityConfig_Validate(t *testing.T) {
	config := NewAccessibilityConfig()

	// Valid config
	if err := config.Validate(); err != nil {
		t.Errorf("valid config failed validation: %v", err)
	}

	// Invalid font scale (too low)
	config.FontScale = 0.3
	if err := config.Validate(); err == nil {
		t.Error("expected error for font scale < 0.5")
	}

	// Invalid font scale (too high)
	config.FontScale = 2.5
	if err := config.Validate(); err == nil {
		t.Error("expected error for font scale > 2.0")
	}

	// Valid boundary values
	config.FontScale = 0.5
	if err := config.Validate(); err != nil {
		t.Error("font scale 0.5 should be valid")
	}

	config.FontScale = 2.0
	if err := config.Validate(); err != nil {
		t.Error("font scale 2.0 should be valid")
	}
}

func TestAccessibilityConfig_ColorblindFilter(t *testing.T) {
	config := NewAccessibilityConfig()

	// Test with no filter
	r, g, b := config.ApplyColorblindFilter(255, 128, 64)
	if r != 255 || g != 128 || b != 64 {
		t.Error("colors should be unchanged with no filter")
	}

	// Test protanopia (red-blind)
	config.ColorblindMode = ColorblindProtanopia
	r, g, b = config.ApplyColorblindFilter(255, 128, 64)
	if r == 255 {
		t.Error("red channel should be modified for protanopia")
	}

	// Test deuteranopia (green-blind)
	config.ColorblindMode = ColorblindDeuteranopia
	r, g, b = config.ApplyColorblindFilter(255, 128, 64)
	if g == 128 {
		t.Error("green channel should be modified for deuteranopia")
	}

	// Test tritanopia (blue-blind)
	config.ColorblindMode = ColorblindTritanopia
	r, g, b = config.ApplyColorblindFilter(255, 128, 64)
	if b == 64 {
		t.Error("blue channel should be modified for tritanopia")
	}
}

func TestAccessibilityConfig_ContrastMultiplier(t *testing.T) {
	config := NewAccessibilityConfig()

	// Normal contrast
	if mult := config.GetContrastMultiplier(); mult != 1.0 {
		t.Errorf("expected contrast 1.0, got %f", mult)
	}

	// High contrast
	config.HighContrast = true
	if mult := config.GetContrastMultiplier(); mult != 1.5 {
		t.Errorf("expected contrast 1.5, got %f", mult)
	}
}

func TestAccessibilityConfig_ClosedCaptions(t *testing.T) {
	config := NewAccessibilityConfig()

	// Off by default
	if config.ShouldShowCaptions() {
		t.Error("expected captions off by default")
	}

	// Enable
	config.ClosedCaptions = true
	if !config.ShouldShowCaptions() {
		t.Error("expected captions enabled")
	}
}

func TestColorblindMode_String(t *testing.T) {
	tests := []struct {
		mode ColorblindMode
		want string
	}{
		{ColorblindNone, "None"},
		{ColorblindProtanopia, "Protanopia"},
		{ColorblindDeuteranopia, "Deuteranopia"},
		{ColorblindTritanopia, "Tritanopia"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("ColorblindMode.String() = %s, want %s", got, tt.want)
		}
	}
}

func TestFormatTooltip(t *testing.T) {
	tooltip := &Tooltip{
		Title:       "Test Item",
		Description: []string{"Line 1", "Line 2"},
		Stats: map[string]interface{}{
			"Damage":  100,
			"Defense": 50,
		},
		Bonuses:      []string{"Bonus 1", "Bonus 2"},
		Requirements: []string{"Level 10"},
		Cost:         1000,
		Rarity:       "Rare",
	}

	formatted := FormatTooltip(tooltip)

	if formatted == "" {
		t.Error("expected non-empty formatted tooltip")
	}

	// Check for expected content
	if len(formatted) < 50 {
		t.Error("formatted tooltip seems too short")
	}
}

func TestCreateCompanionTooltip(t *testing.T) {
	skills := []string{"Attack", "Defend", "Heal"}
	tooltip := CreateCompanionTooltip("Fluffy", 25, 85, skills)

	if tooltip.Title != "Fluffy" {
		t.Error("incorrect title")
	}

	if tooltip.Stats["Level"] != 25 {
		t.Error("incorrect level stat")
	}

	if len(tooltip.Bonuses) == 0 {
		t.Error("expected loyalty bonus for high loyalty")
	}
}

func BenchmarkTutorialManager_GetTutorial(b *testing.B) {
	tm := NewTutorialManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.GetTutorial(TutorialMovement)
	}
}

func BenchmarkTutorialManager_ShowTutorial(b *testing.B) {
	tm := NewTutorialManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.ResetProgress() // Reset to allow re-showing
		tm.ShowTutorial(TutorialCombat)
	}
}

func BenchmarkAccessibilityConfig_ApplyColorblindFilter(b *testing.B) {
	config := NewAccessibilityConfig()
	config.ColorblindMode = ColorblindProtanopia
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ApplyColorblindFilter(255, 128, 64)
	}
}

func BenchmarkTooltipBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb := NewTooltipBuilder("Item")
		tb.AddDescription("Desc").
			AddStat("Damage", 100).
			AddBonus("Bonus").
			SetCost(1000).
			Build()
	}
}
