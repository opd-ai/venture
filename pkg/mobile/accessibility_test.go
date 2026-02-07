package mobile

import (
	"testing"
)

func TestNewAccessibilityHint(t *testing.T) {
	label := "Health Bar"
	hint := "Shows current player health"
	traits := []AccessibilityTrait{TraitAdjustable, TraitUpdatesFrequently}

	h := NewAccessibilityHint(label, hint, traits)

	if h.Label != label {
		t.Errorf("expected label %q, got %q", label, h.Label)
	}
	if h.Hint != hint {
		t.Errorf("expected hint %q, got %q", hint, h.Hint)
	}
	if len(h.Traits) != len(traits) {
		t.Errorf("expected %d traits, got %d", len(traits), len(h.Traits))
	}
	if !h.IsEnabled {
		t.Error("expected hint to be enabled by default")
	}
}

func TestAccessibilityHint_SetValue(t *testing.T) {
	h := NewAccessibilityHint("Health", "Player health", []AccessibilityTrait{TraitAdjustable})

	h.SetValue("75%")

	if h.Value != "75%" {
		t.Errorf("expected value %q, got %q", "75%", h.Value)
	}
}

func TestAccessibilityHint_SetEnabled(t *testing.T) {
	h := NewAccessibilityHint("Button", "Test button", []AccessibilityTrait{TraitButton})

	if !h.IsEnabled {
		t.Error("expected hint to be enabled by default")
	}

	h.SetEnabled(false)

	if h.IsEnabled {
		t.Error("expected hint to be disabled after SetEnabled(false)")
	}

	h.SetEnabled(true)

	if !h.IsEnabled {
		t.Error("expected hint to be enabled after SetEnabled(true)")
	}
}

func TestAccessibilityHint_GetARIAAttributes_Button(t *testing.T) {
	h := NewAccessibilityHint("Attack", "Perform melee attack", []AccessibilityTrait{TraitButton})

	attrs := h.GetARIAAttributes()

	if attrs["aria-label"] != "Attack" {
		t.Errorf("expected aria-label %q, got %q", "Attack", attrs["aria-label"])
	}
	if attrs["aria-description"] != "Perform melee attack" {
		t.Errorf("expected aria-description %q, got %q", "Perform melee attack", attrs["aria-description"])
	}
	if attrs["role"] != "button" {
		t.Errorf("expected role %q, got %q", "button", attrs["role"])
	}
	if _, exists := attrs["aria-disabled"]; exists {
		t.Error("expected no aria-disabled attribute for enabled button")
	}
}

func TestAccessibilityHint_GetARIAAttributes_DisabledButton(t *testing.T) {
	h := NewAccessibilityHint("Disabled", "Cannot use", []AccessibilityTrait{TraitButton})
	h.SetEnabled(false)

	attrs := h.GetARIAAttributes()

	if attrs["aria-disabled"] != "true" {
		t.Errorf("expected aria-disabled %q, got %q", "true", attrs["aria-disabled"])
	}
}

func TestAccessibilityHint_GetARIAAttributes_Adjustable(t *testing.T) {
	h := NewAccessibilityHint("Health", "Player health", []AccessibilityTrait{TraitAdjustable})
	h.SetValue("75%")

	attrs := h.GetARIAAttributes()

	if attrs["role"] != "slider" {
		t.Errorf("expected role %q, got %q", "slider", attrs["role"])
	}
	if attrs["aria-valuenow"] != "75%" {
		t.Errorf("expected aria-valuenow %q, got %q", "75%", attrs["aria-valuenow"])
	}
}

func TestAccessibilityHint_GetARIAAttributes_UpdatesFrequently(t *testing.T) {
	h := NewAccessibilityHint("Notification", "Status messages", []AccessibilityTrait{TraitUpdatesFrequently})

	attrs := h.GetARIAAttributes()

	if attrs["aria-live"] != "polite" {
		t.Errorf("expected aria-live %q, got %q", "polite", attrs["aria-live"])
	}
}

func TestAccessibilityHint_GetARIAAttributes_Selected(t *testing.T) {
	h := NewAccessibilityHint("Menu Item", "Selected option", []AccessibilityTrait{TraitButton, TraitSelected})

	attrs := h.GetARIAAttributes()

	if attrs["aria-selected"] != "true" {
		t.Errorf("expected aria-selected %q, got %q", "true", attrs["aria-selected"])
	}
}

func TestAccessibilityHint_GetARIAAttributes_MultipleTraits(t *testing.T) {
	// Test that multiple traits are handled correctly
	// Button trait should take precedence for role
	h := NewAccessibilityHint(
		"Dynamic Button",
		"Updates frequently",
		[]AccessibilityTrait{TraitButton, TraitUpdatesFrequently},
	)

	attrs := h.GetARIAAttributes()

	if attrs["role"] != "button" {
		t.Errorf("expected role %q, got %q", "button", attrs["role"])
	}
	if attrs["aria-live"] != "polite" {
		t.Errorf("expected aria-live %q, got %q", "polite", attrs["aria-live"])
	}
}

func TestAccessibilityHint_ContainsTrait(t *testing.T) {
	h := NewAccessibilityHint(
		"Test",
		"Test element",
		[]AccessibilityTrait{TraitButton, TraitUpdatesFrequently},
	)

	tests := []struct {
		trait    AccessibilityTrait
		expected bool
	}{
		{TraitButton, true},
		{TraitUpdatesFrequently, true},
		{TraitImage, false},
		{TraitHeader, false},
		{TraitStaticText, false},
	}

	for _, tt := range tests {
		got := h.containsTrait(tt.trait)
		if got != tt.expected {
			t.Errorf("containsTrait(%q) = %v, want %v", tt.trait, got, tt.expected)
		}
	}
}

func TestStandardHints_HealthBar(t *testing.T) {
	h := StandardHints.HealthBar

	if h.Label != "Health" {
		t.Errorf("expected label %q, got %q", "Health", h.Label)
	}
	if !h.containsTrait(TraitAdjustable) {
		t.Error("expected HealthBar to have TraitAdjustable")
	}
	if !h.containsTrait(TraitUpdatesFrequently) {
		t.Error("expected HealthBar to have TraitUpdatesFrequently")
	}
}

func TestStandardHints_ManaBar(t *testing.T) {
	h := StandardHints.ManaBar

	if h.Label != "Mana" {
		t.Errorf("expected label %q, got %q", "Mana", h.Label)
	}
	if !h.containsTrait(TraitAdjustable) {
		t.Error("expected ManaBar to have TraitAdjustable")
	}
}

func TestStandardHints_InventoryButton(t *testing.T) {
	h := StandardHints.InventoryButton

	if h.Label != "Inventory" {
		t.Errorf("expected label %q, got %q", "Inventory", h.Label)
	}
	if !h.containsTrait(TraitButton) {
		t.Error("expected InventoryButton to have TraitButton")
	}

	attrs := h.GetARIAAttributes()
	if attrs["role"] != "button" {
		t.Errorf("expected role %q, got %q", "button", attrs["role"])
	}
}

func TestStandardHints_AllButtonsHaveButtonTrait(t *testing.T) {
	buttons := []*AccessibilityHint{
		StandardHints.InventoryButton,
		StandardHints.MapButton,
		StandardHints.MenuButton,
		StandardHints.AttackButton,
		StandardHints.InteractButton,
	}

	for _, btn := range buttons {
		if !btn.containsTrait(TraitButton) {
			t.Errorf("%q should have TraitButton", btn.Label)
		}
	}
}

func TestStandardHints_AllBarsHaveAdjustableTrait(t *testing.T) {
	bars := []*AccessibilityHint{
		StandardHints.HealthBar,
		StandardHints.ManaBar,
		StandardHints.ExperienceBar,
	}

	for _, bar := range bars {
		if !bar.containsTrait(TraitAdjustable) {
			t.Errorf("%q should have TraitAdjustable", bar.Label)
		}
		if !bar.containsTrait(TraitUpdatesFrequently) {
			t.Errorf("%q should have TraitUpdatesFrequently", bar.Label)
		}
	}
}

func TestStandardHints_MinimapConfiguration(t *testing.T) {
	h := StandardHints.Minimap

	if !h.containsTrait(TraitImage) {
		t.Error("expected Minimap to have TraitImage")
	}
	if !h.containsTrait(TraitUpdatesFrequently) {
		t.Error("expected Minimap to have TraitUpdatesFrequently")
	}

	attrs := h.GetARIAAttributes()
	if attrs["role"] != "img" {
		t.Errorf("expected role %q, got %q", "img", attrs["role"])
	}
	if attrs["aria-live"] != "polite" {
		t.Errorf("expected aria-live %q, got %q", "polite", attrs["aria-live"])
	}
}

func TestAccessibilityHint_GetARIAAttributes_EmptyHint(t *testing.T) {
	// Test that hints work without a description
	h := NewAccessibilityHint("Button", "", []AccessibilityTrait{TraitButton})

	attrs := h.GetARIAAttributes()

	if attrs["aria-label"] != "Button" {
		t.Errorf("expected aria-label %q, got %q", "Button", attrs["aria-label"])
	}
	if _, exists := attrs["aria-description"]; exists {
		t.Error("expected no aria-description when hint is empty")
	}
}

func TestAccessibilityHint_GetARIAAttributes_NoValue(t *testing.T) {
	// Test that adjustable elements work without a value set
	h := NewAccessibilityHint("Slider", "Test slider", []AccessibilityTrait{TraitAdjustable})

	attrs := h.GetARIAAttributes()

	if _, exists := attrs["aria-valuenow"]; exists {
		t.Error("expected no aria-valuenow when value is not set")
	}
}

func TestAccessibilityHint_GetARIAAttributes_AllRoles(t *testing.T) {
	// Test each trait maps to correct ARIA role
	tests := []struct {
		trait        AccessibilityTrait
		expectedRole string
	}{
		{TraitButton, "button"},
		{TraitImage, "img"},
		{TraitHeader, "heading"},
		{TraitLink, "link"},
		{TraitAdjustable, "slider"},
		{TraitStaticText, "text"},
	}

	for _, tt := range tests {
		h := NewAccessibilityHint("Test", "Test element", []AccessibilityTrait{tt.trait})
		attrs := h.GetARIAAttributes()

		if attrs["role"] != tt.expectedRole {
			t.Errorf("trait %q: expected role %q, got %q", tt.trait, tt.expectedRole, attrs["role"])
		}
	}
}

func BenchmarkGetARIAAttributes(b *testing.B) {
	h := NewAccessibilityHint(
		"Health",
		"Current player health percentage",
		[]AccessibilityTrait{TraitAdjustable, TraitUpdatesFrequently},
	)
	h.SetValue("75%")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.GetARIAAttributes()
	}
}

func BenchmarkContainsTrait(b *testing.B) {
	h := NewAccessibilityHint(
		"Test",
		"Test element",
		[]AccessibilityTrait{TraitButton, TraitUpdatesFrequently, TraitSelected},
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.containsTrait(TraitSelected)
	}
}
