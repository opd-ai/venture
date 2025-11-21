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
