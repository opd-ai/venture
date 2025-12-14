package engine

import (
	"testing"
)

func TestEventDecorationComponent_Type(t *testing.T) {
	comp := NewEventDecorationComponent(12345)
	if comp.Type() != "event_decoration" {
		t.Errorf("Expected type 'event_decoration', got '%s'", comp.Type())
	}
}

func TestEventDecorationComponent_NewEventDecorationComponent(t *testing.T) {
	comp := NewEventDecorationComponent(12345)

	if comp.Seed != 12345 {
		t.Errorf("Expected Seed 12345, got %d", comp.Seed)
	}
	if comp.ActiveTheme != DecorationThemeNone {
		t.Errorf("Expected DecorationThemeNone, got '%s'", comp.ActiveTheme)
	}
	if comp.DecorationLevel != 0.0 {
		t.Errorf("Expected DecorationLevel 0.0, got %f", comp.DecorationLevel)
	}
	if len(comp.Elements) != 0 {
		t.Errorf("Expected empty Elements, got %d", len(comp.Elements))
	}
	if len(comp.ParticleEffects) != 0 {
		t.Errorf("Expected empty ParticleEffects, got %d", len(comp.ParticleEffects))
	}
}

func TestEventDecorationComponent_HasDecorations(t *testing.T) {
	tests := []struct {
		name     string
		theme    DecorationTheme
		level    float64
		expected bool
	}{
		{"no theme", DecorationThemeNone, 0.5, false},
		{"zero level", DecorationThemeSpring, 0.0, false},
		{"has decorations", DecorationThemeSpring, 0.5, true},
		{"full decorations", DecorationThemeWinter, 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewEventDecorationComponent(12345)
			comp.ActiveTheme = tt.theme
			comp.DecorationLevel = tt.level
			result := comp.HasDecorations()
			if result != tt.expected {
				t.Errorf("HasDecorations() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestEventDecorationComponent_GenerateDecorations(t *testing.T) {
	comp := NewEventDecorationComponent(12345)
	comp.GenerateDecorations(DecorationThemeSpring, 0.8)

	if comp.ActiveTheme != DecorationThemeSpring {
		t.Errorf("Expected theme spring, got '%s'", comp.ActiveTheme)
	}
	if comp.DecorationLevel != 0.8 {
		t.Errorf("Expected level 0.8, got %f", comp.DecorationLevel)
	}
	if len(comp.Elements) == 0 {
		t.Error("Expected some decoration elements")
	}
	if len(comp.ParticleEffects) == 0 {
		t.Error("Expected some particle effects at level 0.8")
	}
	if comp.CostumeVariant < 1 || comp.CostumeVariant > 4 {
		t.Errorf("Expected costume variant 1-4, got %d", comp.CostumeVariant)
	}
}

func TestEventDecorationComponent_GenerateDecorations_Determinism(t *testing.T) {
	seed := int64(54321)

	// Generate twice with same seed
	comp1 := NewEventDecorationComponent(seed)
	comp1.GenerateDecorations(DecorationThemeSummer, 0.7)

	comp2 := NewEventDecorationComponent(seed)
	comp2.GenerateDecorations(DecorationThemeSummer, 0.7)

	if len(comp1.Elements) != len(comp2.Elements) {
		t.Errorf("Element counts differ: %d vs %d", len(comp1.Elements), len(comp2.Elements))
	}

	for i := range comp1.Elements {
		if comp1.Elements[i].Type != comp2.Elements[i].Type {
			t.Errorf("Element %d type mismatch: %s vs %s", i, comp1.Elements[i].Type, comp2.Elements[i].Type)
		}
		if comp1.Elements[i].ColorHue != comp2.Elements[i].ColorHue {
			t.Errorf("Element %d hue mismatch: %d vs %d", i, comp1.Elements[i].ColorHue, comp2.Elements[i].ColorHue)
		}
	}

	if comp1.CostumeVariant != comp2.CostumeVariant {
		t.Errorf("Costume variants differ: %d vs %d", comp1.CostumeVariant, comp2.CostumeVariant)
	}
}

func TestEventDecorationComponent_GetVisibleElements(t *testing.T) {
	comp := NewEventDecorationComponent(12345)
	comp.GenerateDecorations(DecorationThemeAutumn, 1.0)

	// Full visibility when not transitioning
	visible := comp.GetVisibleElements()
	if len(visible) != len(comp.Elements) {
		t.Errorf("Expected %d visible elements, got %d", len(comp.Elements), len(visible))
	}

	// Partial visibility during transition
	comp.IsTransitioning = true
	comp.TransitionProgress = 0.5
	visible = comp.GetVisibleElements()
	expected := int(float64(len(comp.Elements)) * 0.5)
	if len(visible) != expected {
		t.Errorf("Expected %d visible elements at 50%% transition, got %d", expected, len(visible))
	}

	// No visibility at start of transition
	comp.TransitionProgress = 0.0
	visible = comp.GetVisibleElements()
	if visible != nil && len(visible) != 0 {
		t.Errorf("Expected no visible elements at 0%% transition, got %d", len(visible))
	}
}

func TestEventDecorationComponent_GetActiveParticleEffects(t *testing.T) {
	comp := NewEventDecorationComponent(12345)
	comp.GenerateDecorations(DecorationThemeWinter, 1.0)
	comp.TransitionProgress = 1.0

	// Should have effects when fully decorated
	effects := comp.GetActiveParticleEffects()
	if len(effects) == 0 {
		t.Error("Expected particle effects when fully decorated")
	}

	// No effects at low transition progress
	comp.TransitionProgress = 0.3
	effects = comp.GetActiveParticleEffects()
	if effects != nil {
		t.Error("Expected no particle effects at low transition progress")
	}

	// No effects when no decorations
	comp.ClearDecorations()
	effects = comp.GetActiveParticleEffects()
	if effects != nil {
		t.Error("Expected no particle effects after clearing")
	}
}

func TestEventDecorationComponent_ClearDecorations(t *testing.T) {
	comp := NewEventDecorationComponent(12345)
	comp.GenerateDecorations(DecorationThemeSpring, 0.8)
	comp.EventID = "test_event"
	comp.TransitionProgress = 1.0

	comp.ClearDecorations()

	if comp.ActiveTheme != DecorationThemeNone {
		t.Errorf("Expected no theme after clear, got '%s'", comp.ActiveTheme)
	}
	if comp.EventID != "" {
		t.Errorf("Expected empty event ID after clear, got '%s'", comp.EventID)
	}
	if comp.DecorationLevel != 0.0 {
		t.Errorf("Expected level 0 after clear, got %f", comp.DecorationLevel)
	}
	if len(comp.Elements) != 0 {
		t.Errorf("Expected empty elements after clear, got %d", len(comp.Elements))
	}
	if comp.CostumeVariant != 0 {
		t.Errorf("Expected no costume variant after clear, got %d", comp.CostumeVariant)
	}
}

func TestEventDecorationComponent_Serialize_Deserialize(t *testing.T) {
	original := NewEventDecorationComponent(12345)
	original.GenerateDecorations(DecorationThemeWinter, 0.9)
	original.EventID = "winter_celebration"
	original.TransitionProgress = 1.0

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Serialize returned empty data")
	}

	// Deserialize
	restored := &EventDecorationComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify values
	if restored.Seed != original.Seed {
		t.Errorf("Seed mismatch: %d vs %d", restored.Seed, original.Seed)
	}
	if restored.ActiveTheme != original.ActiveTheme {
		t.Errorf("Theme mismatch: %s vs %s", restored.ActiveTheme, original.ActiveTheme)
	}
	if restored.EventID != original.EventID {
		t.Errorf("EventID mismatch: %s vs %s", restored.EventID, original.EventID)
	}
	if len(restored.Elements) != len(original.Elements) {
		t.Errorf("Elements count mismatch: %d vs %d", len(restored.Elements), len(original.Elements))
	}
	if restored.CostumeVariant != original.CostumeVariant {
		t.Errorf("CostumeVariant mismatch: %d vs %d", restored.CostumeVariant, original.CostumeVariant)
	}
}

func TestThemeFromEventTheme(t *testing.T) {
	tests := []struct {
		eventTheme EventTheme
		expected   DecorationTheme
	}{
		{EventThemeSpring, DecorationThemeSpring},
		{EventThemeSummer, DecorationThemeSummer},
		{EventThemeAutumn, DecorationThemeAutumn},
		{EventThemeWinter, DecorationThemeWinter},
		{EventTheme("unknown"), DecorationThemeNone},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventTheme), func(t *testing.T) {
			result := ThemeFromEventTheme(tt.eventTheme)
			if result != tt.expected {
				t.Errorf("ThemeFromEventTheme(%s) = %s, expected %s", tt.eventTheme, result, tt.expected)
			}
		})
	}
}

func TestDecorationTheme_Constants(t *testing.T) {
	themes := []DecorationTheme{
		DecorationThemeSpring,
		DecorationThemeSummer,
		DecorationThemeAutumn,
		DecorationThemeWinter,
	}

	for _, theme := range themes {
		if theme == "" {
			t.Error("Decoration theme constant is empty (should only be empty for None)")
		}
	}

	if DecorationThemeNone != "" {
		t.Error("DecorationThemeNone should be empty string")
	}
}

func TestEventDecorationComponent_GetCostumeOffset(t *testing.T) {
	tests := []struct {
		name      string
		theme     DecorationTheme
		variant   int
		hasDecor  bool
		expectedY int
	}{
		{"no costume", DecorationThemeNone, 0, false, 0},
		{"spring variant 1", DecorationThemeSpring, 1, true, 4},
		{"spring variant 2", DecorationThemeSpring, 2, true, 5},
		{"summer variant 3", DecorationThemeSummer, 3, true, 10},
		{"autumn variant 4", DecorationThemeAutumn, 4, true, 15},
		{"winter variant 1", DecorationThemeWinter, 1, true, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewEventDecorationComponent(12345)
			comp.ActiveTheme = tt.theme
			comp.CostumeVariant = tt.variant
			if tt.hasDecor {
				comp.DecorationLevel = 1.0
			}

			_, offsetY := comp.GetCostumeOffset()
			if offsetY != tt.expectedY {
				t.Errorf("GetCostumeOffset() Y = %d, expected %d", offsetY, tt.expectedY)
			}
		})
	}
}

func TestEventDecorationComponent_IsFullyDecorated(t *testing.T) {
	comp := NewEventDecorationComponent(12345)

	// Not decorated
	if comp.IsFullyDecorated() {
		t.Error("Should not be fully decorated when empty")
	}

	// Generating decorations
	comp.GenerateDecorations(DecorationThemeSpring, 0.8)
	comp.TransitionProgress = 0.5
	comp.IsTransitioning = true
	if comp.IsFullyDecorated() {
		t.Error("Should not be fully decorated during transition")
	}

	// Fully decorated
	comp.TransitionProgress = 1.0
	comp.IsTransitioning = false
	if !comp.IsFullyDecorated() {
		t.Error("Should be fully decorated")
	}
}

func TestGenerateDecorationElements_AllThemes(t *testing.T) {
	themes := []DecorationTheme{
		DecorationThemeSpring,
		DecorationThemeSummer,
		DecorationThemeAutumn,
		DecorationThemeWinter,
	}

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			comp := NewEventDecorationComponent(12345)
			comp.GenerateDecorations(theme, 1.0)

			if len(comp.Elements) == 0 {
				t.Errorf("Expected elements for theme %s", theme)
			}

			// Verify all elements have valid types
			validTypes := getDecorationTypesForTheme(theme)
			for _, elem := range comp.Elements {
				found := false
				for _, vt := range validTypes {
					if elem.Type == vt {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Invalid element type '%s' for theme %s", elem.Type, theme)
				}
			}
		})
	}
}

func TestGenerateParticleEffects_LevelThreshold(t *testing.T) {
	// No effects at low level
	comp := NewEventDecorationComponent(12345)
	comp.GenerateDecorations(DecorationThemeSpring, 0.2)
	if len(comp.ParticleEffects) > 0 {
		t.Error("Should have no particle effects at level 0.2")
	}

	// Effects at higher level
	comp.GenerateDecorations(DecorationThemeSpring, 0.5)
	if len(comp.ParticleEffects) == 0 {
		t.Error("Should have particle effects at level 0.5")
	}
}

func TestGenerateDecorations_LevelClamping(t *testing.T) {
	comp := NewEventDecorationComponent(12345)

	// Test level > 1.0 is clamped
	comp.GenerateDecorations(DecorationThemeWinter, 1.5)
	if comp.DecorationLevel != 1.0 {
		t.Errorf("Level should be clamped to 1.0, got %f", comp.DecorationLevel)
	}

	// Test level < 0.0 is clamped
	comp.GenerateDecorations(DecorationThemeWinter, -0.5)
	if comp.DecorationLevel != 0.0 {
		t.Errorf("Level should be clamped to 0.0, got %f", comp.DecorationLevel)
	}
}
