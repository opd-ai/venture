// Package engine provides tests for gamepad input handling.
package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewGamepadInputHandler verifies correct initialization of GamepadInputHandler.
func TestNewGamepadInputHandler(t *testing.T) {
	handler := NewGamepadInputHandler()

	if handler == nil {
		t.Fatal("NewGamepadInputHandler() returned nil")
	}

	// Verify default dead zones
	if handler.LeftStickDeadZone != 0.15 {
		t.Errorf("LeftStickDeadZone = %v, want 0.15", handler.LeftStickDeadZone)
	}
	if handler.RightStickDeadZone != 0.15 {
		t.Errorf("RightStickDeadZone = %v, want 0.15", handler.RightStickDeadZone)
	}

	// Verify default trigger threshold
	if handler.TriggerThreshold != 0.5 {
		t.Errorf("TriggerThreshold = %v, want 0.5", handler.TriggerThreshold)
	}

	// Verify no gamepad connected initially
	if handler.HasGamepad() {
		t.Error("HasGamepad() should return false when no gamepad connected")
	}
}

// TestGamepadInputHandler_DefaultButtonMappings verifies Xbox-style button mappings.
func TestGamepadInputHandler_DefaultButtonMappings(t *testing.T) {
	handler := NewGamepadInputHandler()

	tests := []struct {
		name     string
		button   ebiten.StandardGamepadButton
		expected ebiten.StandardGamepadButton
	}{
		{"Attack", handler.ButtonAttack, ebiten.StandardGamepadButtonRightBottom},     // A
		{"Secondary", handler.ButtonSecondary, ebiten.StandardGamepadButtonRightRight}, // B
		{"UseItem", handler.ButtonUseItem, ebiten.StandardGamepadButtonRightLeft},      // X
		{"Interact", handler.ButtonInteract, ebiten.StandardGamepadButtonRightTop},     // Y
		{"Menu", handler.ButtonMenu, ebiten.StandardGamepadButtonCenterRight},          // Start
		{"Map", handler.ButtonMap, ebiten.StandardGamepadButtonCenterLeft},             // Back
		{"CycleTarget", handler.ButtonCycleTarget, ebiten.StandardGamepadButtonFrontTopLeft},  // LB
		{"Inventory", handler.ButtonInventory, ebiten.StandardGamepadButtonFrontTopRight},     // RB
		{"Spell1", handler.ButtonSpell1, ebiten.StandardGamepadButtonLeftTop},          // D-pad Up
		{"Spell2", handler.ButtonSpell2, ebiten.StandardGamepadButtonLeftRight},        // D-pad Right
		{"Spell3", handler.ButtonSpell3, ebiten.StandardGamepadButtonLeftBottom},       // D-pad Down
		{"Spell4", handler.ButtonSpell4, ebiten.StandardGamepadButtonLeftLeft},         // D-pad Left
		// Platform parity fix: UI shortcut buttons
		{"Character", handler.ButtonCharacter, ebiten.StandardGamepadButtonLeftStick},  // L3
		{"Skills", handler.ButtonSkills, ebiten.StandardGamepadButtonRightStick},       // R3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.button != tt.expected {
				t.Errorf("Button%s = %v, want %v", tt.name, tt.button, tt.expected)
			}
		})
	}
}

// TestGamepadInputHandler_ApplyDeadZone verifies dead zone calculation.
func TestGamepadInputHandler_ApplyDeadZone(t *testing.T) {
	handler := NewGamepadInputHandler()
	handler.LeftStickDeadZone = 0.2

	tests := []struct {
		name     string
		value    float64
		deadZone float64
		expected float64
	}{
		{"zero", 0, 0.2, 0},
		{"inside_positive", 0.15, 0.2, 0},
		{"inside_negative", -0.15, 0.2, 0},
		{"at_deadzone", 0.2, 0.2, 0},
		{"outside_positive", 0.6, 0.2, 0.5},
		{"outside_negative", -0.6, 0.2, -0.5},
		{"max_positive", 1.0, 0.2, 1.0},
		{"max_negative", -1.0, 0.2, -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.applyDeadZone(tt.value, tt.deadZone)
			// Allow small floating point variance
			if diff := result - tt.expected; diff > 0.01 || diff < -0.01 {
				t.Errorf("applyDeadZone(%v, %v) = %v, want %v", tt.value, tt.deadZone, result, tt.expected)
			}
		})
	}
}

// TestGamepadInputHandler_GetMovementInput_NoGamepad verifies zero movement when no gamepad.
func TestGamepadInputHandler_GetMovementInput_NoGamepad(t *testing.T) {
	handler := NewGamepadInputHandler()

	x, y := handler.GetMovementInput()
	if x != 0 || y != 0 {
		t.Errorf("GetMovementInput() without gamepad = (%v, %v), want (0, 0)", x, y)
	}
}

// TestGamepadInputHandler_GetAimInput_NoGamepad verifies zero aim when no gamepad.
func TestGamepadInputHandler_GetAimInput_NoGamepad(t *testing.T) {
	handler := NewGamepadInputHandler()

	x, y := handler.GetAimInput()
	if x != 0 || y != 0 {
		t.Errorf("GetAimInput() without gamepad = (%v, %v), want (0, 0)", x, y)
	}
}

// TestGamepadInputHandler_IsButtonPressed_NoGamepad verifies false when no gamepad.
func TestGamepadInputHandler_IsButtonPressed_NoGamepad(t *testing.T) {
	handler := NewGamepadInputHandler()

	if handler.IsButtonPressed(ebiten.StandardGamepadButtonRightBottom) {
		t.Error("IsButtonPressed() without gamepad should return false")
	}
	if handler.IsAttackPressed() {
		t.Error("IsAttackPressed() without gamepad should return false")
	}
	if handler.IsSecondaryPressed() {
		t.Error("IsSecondaryPressed() without gamepad should return false")
	}
	if handler.IsUseItemPressed() {
		t.Error("IsUseItemPressed() without gamepad should return false")
	}
}

// TestGamepadInputHandler_UIShortcutHelpers_NoGamepad verifies UI shortcut button helper methods.
// Platform parity fix: Ensures Character and Skills buttons work correctly
func TestGamepadInputHandler_UIShortcutHelpers_NoGamepad(t *testing.T) {
	handler := NewGamepadInputHandler()

	// Without gamepad, all UI shortcut inputs should be false
	if handler.IsCharacterJustPressed() {
		t.Error("IsCharacterJustPressed() without gamepad should return false")
	}
	if handler.IsSkillsJustPressed() {
		t.Error("IsSkillsJustPressed() without gamepad should return false")
	}
}

// TestGamepadInputHandler_SpellInputHelpers verifies spell button helper methods.
func TestGamepadInputHandler_SpellInputHelpers(t *testing.T) {
	handler := NewGamepadInputHandler()

	// Without gamepad, all spell inputs should be false
	for slot := 1; slot <= 4; slot++ {
		if handler.IsSpellPressed(slot) {
			t.Errorf("IsSpellPressed(%d) without gamepad should return false", slot)
		}
		if handler.IsSpellJustPressed(slot) {
			t.Errorf("IsSpellJustPressed(%d) without gamepad should return false", slot)
		}
	}

	// Invalid slot should return false
	if handler.IsSpellPressed(0) {
		t.Error("IsSpellPressed(0) should return false for invalid slot")
	}
	if handler.IsSpellPressed(5) {
		t.Error("IsSpellPressed(5) should return false for invalid slot (D-pad only has 4)")
	}
}

// TestGamepadInputHandler_TriggerHelpers verifies trigger input methods.
func TestGamepadInputHandler_TriggerHelpers(t *testing.T) {
	handler := NewGamepadInputHandler()

	// Without gamepad, triggers should return 0
	if handler.GetLeftTrigger() != 0 {
		t.Error("GetLeftTrigger() without gamepad should return 0")
	}
	if handler.GetRightTrigger() != 0 {
		t.Error("GetRightTrigger() without gamepad should return 0")
	}
	if handler.IsLeftTriggerPressed() {
		t.Error("IsLeftTriggerPressed() without gamepad should return false")
	}
	if handler.IsRightTriggerPressed() {
		t.Error("IsRightTriggerPressed() without gamepad should return false")
	}
}

// TestGamepadInputHandler_GetGamepadName verifies gamepad name retrieval.
func TestGamepadInputHandler_GetGamepadName(t *testing.T) {
	handler := NewGamepadInputHandler()

	name := handler.GetGamepadName()
	if name != "" {
		t.Errorf("GetGamepadName() without gamepad should return empty string, got %q", name)
	}
}

// TestGamepadInputHandler_GetGamepadID verifies gamepad ID retrieval.
func TestGamepadInputHandler_GetGamepadID(t *testing.T) {
	handler := NewGamepadInputHandler()

	id := handler.GetGamepadID()
	if id != -1 {
		t.Errorf("GetGamepadID() without gamepad should return -1, got %v", id)
	}
}

// TestGamepadInputHandler_Update_NoGamepad verifies Update doesn't panic without gamepad.
func TestGamepadInputHandler_Update_NoGamepad(t *testing.T) {
	handler := NewGamepadInputHandler()

	// Should not panic
	handler.Update()

	if handler.HasGamepad() {
		t.Error("HasGamepad() should still be false after Update() with no gamepad")
	}
}
