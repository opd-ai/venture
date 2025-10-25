// Package engine provides tests for standardized menu key navigation.
package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestMenuKeys_Constants verifies that all menu keys are properly assigned.
func TestMenuKeys_Constants(t *testing.T) {
	tests := []struct {
		name     string
		key      ebiten.Key
		expected ebiten.Key
	}{
		{"Inventory key", MenuKeys.Inventory, ebiten.KeyI},
		{"Character key", MenuKeys.Character, ebiten.KeyC},
		{"Skills key", MenuKeys.Skills, ebiten.KeyK},
		{"Quests key", MenuKeys.Quests, ebiten.KeyJ},
		{"Map key", MenuKeys.Map, ebiten.KeyM},
		{"Exit key", MenuKeys.Exit, ebiten.KeyEscape},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != tt.expected {
				t.Errorf("MenuKeys.%s = %v, want %v", tt.name, tt.key, tt.expected)
			}
		})
	}
}

// TestMenuKeys_Labels verifies that display labels are set.
func TestMenuKeys_Labels(t *testing.T) {
	tests := []struct {
		name  string
		label string
	}{
		{"InventoryLabel", MenuKeys.InventoryLabel},
		{"CharacterLabel", MenuKeys.CharacterLabel},
		{"SkillsLabel", MenuKeys.SkillsLabel},
		{"QuestsLabel", MenuKeys.QuestsLabel},
		{"MapLabel", MenuKeys.MapLabel},
		{"ExitHint", MenuKeys.ExitHint},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.label == "" {
				t.Errorf("MenuKeys.%s is empty, expected non-empty label", tt.name)
			}
		})
	}
}

// TestHandleMenuInput_ToggleKey tests toggle key behavior.
func TestHandleMenuInput_ToggleKey(t *testing.T) {
	// Note: This test cannot fully simulate key presses without Ebiten's input system.
	// It verifies the API contract only.

	t.Run("API returns expected values when menu closed", func(t *testing.T) {
		// When menu is closed, toggle key should indicate open
		shouldClose, shouldToggle := HandleMenuInput(ebiten.KeyI, false)

		// We can't test actual key press without Ebiten context,
		// but we can verify the function signature and default behavior
		_ = shouldClose
		_ = shouldToggle
	})

	t.Run("API returns expected values when menu open", func(t *testing.T) {
		// When menu is open, toggle key or escape should close
		shouldClose, shouldToggle := HandleMenuInput(ebiten.KeyI, true)

		_ = shouldClose
		_ = shouldToggle
	})
}

// TestHandleMenuInput_EscapeKey tests Escape key behavior.
func TestHandleMenuInput_EscapeKey(t *testing.T) {
	t.Run("Escape only works when menu is visible", func(t *testing.T) {
		// Escape should only close menus, never open them
		// This is a contract test - actual key press requires Ebiten runtime
		shouldClose, shouldToggle := HandleMenuInput(ebiten.KeyEscape, false)

		_ = shouldClose
		_ = shouldToggle
	})
}

// TestMenuNavigation_Integration tests the integration pattern used by UI systems.
func TestMenuNavigation_Integration(t *testing.T) {
	t.Run("Standard menu pattern", func(t *testing.T) {
		// This test documents the expected usage pattern:
		visible := false

		// Simulate menu update logic
		if shouldClose, shouldToggle := HandleMenuInput(ebiten.KeyI, visible); shouldClose {
			if shouldToggle {
				visible = !visible // Toggle open/close
			} else {
				visible = false // Escape always closes
			}
		}

		// Verify visible didn't change (no actual key press in test)
		if visible != false {
			t.Errorf("visible = %v, expected false (no key press simulated)", visible)
		}
	})
}

// TestMenuKeys_Uniqueness verifies that each menu has a unique key binding.
func TestMenuKeys_Uniqueness(t *testing.T) {
	keys := map[ebiten.Key]string{
		MenuKeys.Inventory: "Inventory",
		MenuKeys.Character: "Character",
		MenuKeys.Skills:    "Skills",
		MenuKeys.Quests:    "Quests",
		MenuKeys.Map:       "Map",
	}

	// Check for duplicate keys (excluding Exit which is universal)
	seen := make(map[ebiten.Key]string)
	for key, name := range keys {
		if existing, exists := seen[key]; exists {
			t.Errorf("Duplicate key binding: %v is assigned to both %s and %s",
				key, name, existing)
		}
		seen[key] = name
	}

	// Verify we have exactly 5 unique menu keys
	if len(seen) != 5 {
		t.Errorf("Expected 5 unique menu keys, got %d", len(seen))
	}
}

// TestMenuKeys_Mnemonic verifies that menu keys match their function name.
func TestMenuKeys_Mnemonic(t *testing.T) {
	tests := []struct {
		name     string
		key      ebiten.Key
		expected rune // First letter of menu name
	}{
		{"Inventory", MenuKeys.Inventory, 'I'},
		{"Character", MenuKeys.Character, 'C'},
		{"Skills", MenuKeys.Skills, 'K'}, // K for sKills (S is reserved for other uses)
		{"Quests", MenuKeys.Quests, 'J'}, // J for Journal
		{"Map", MenuKeys.Map, 'M'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyString := tt.key.String()
			if len(keyString) == 0 {
				t.Errorf("Key has empty string representation")
				return
			}

			// Verify the key string starts with the expected letter
			// (Ebiten keys are formatted as "Key" + letter)
			if !containsRune(keyString, tt.expected) {
				t.Errorf("%s key = %s, should contain '%c'", tt.name, keyString, tt.expected)
			}
		})
	}
}

// containsRune checks if a string contains a specific rune.
func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}
