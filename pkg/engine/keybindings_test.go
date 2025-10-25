package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestAction_String tests action name formatting
func TestAction_String(t *testing.T) {
	tests := []struct {
		action Action
		want   string
	}{
		{ActionMoveUp, "Move Up"},
		{ActionMoveDown, "Move Down"},
		{ActionMoveLeft, "Move Left"},
		{ActionMoveRight, "Move Right"},
		{ActionAttack, "Attack"},
		{ActionUseItem, "Use Item"},
		{ActionSecondary, "Secondary Action"},
		{ActionCastSpell1, "Cast Spell 1"},
		{ActionCastSpell5, "Cast Spell 5"},
		{ActionInventory, "Inventory"},
		{ActionCharacter, "Character"},
		{ActionSkills, "Skills"},
		{ActionQuests, "Quests"},
		{ActionMap, "Map"},
		{ActionHelp, "Help/Menu"},
		{ActionQuickSave, "Quick Save"},
		{ActionQuickLoad, "Quick Load"},
		{ActionCycleTargets, "Cycle Targets"},
		{Action(9999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.action.String()
			if got != tt.want {
				t.Errorf("Action.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNewKeyBindingRegistry tests registry creation with defaults
func TestNewKeyBindingRegistry(t *testing.T) {
	registry := NewKeyBindingRegistry()

	if registry == nil {
		t.Fatal("NewKeyBindingRegistry() returned nil")
	}

	if len(registry.bindings) == 0 {
		t.Error("Registry should have default bindings loaded")
	}

	// Verify some default bindings
	tests := []struct {
		action Action
		want   ebiten.Key
	}{
		{ActionMoveUp, ebiten.KeyW},
		{ActionMoveDown, ebiten.KeyS},
		{ActionMoveLeft, ebiten.KeyA},
		{ActionMoveRight, ebiten.KeyD},
		{ActionAttack, ebiten.KeySpace},
		{ActionUseItem, ebiten.KeyE},
		{ActionCastSpell1, ebiten.Key1},
		{ActionInventory, ebiten.KeyI},
		{ActionHelp, ebiten.KeyEscape},
		{ActionQuickSave, ebiten.KeyF5},
	}

	for _, tt := range tests {
		got := registry.GetKey(tt.action)
		if got != tt.want {
			t.Errorf("GetKey(%s) = %v, want %v", tt.action.String(), got, tt.want)
		}
	}
}

// TestKeyBindingRegistry_GetKey tests key retrieval
func TestKeyBindingRegistry_GetKey(t *testing.T) {
	registry := NewKeyBindingRegistry()

	// Valid action
	key := registry.GetKey(ActionAttack)
	if key != ebiten.KeySpace {
		t.Errorf("GetKey(ActionAttack) = %v, want %v", key, ebiten.KeySpace)
	}

	// Invalid action (out of range)
	invalidKey := registry.GetKey(Action(9999))
	if invalidKey != ebiten.KeyMax {
		t.Errorf("GetKey(invalid) should return KeyMax, got %v", invalidKey)
	}
}

// TestKeyBindingRegistry_SetKey tests key binding changes
func TestKeyBindingRegistry_SetKey(t *testing.T) {
	registry := NewKeyBindingRegistry()

	// Change a binding
	err := registry.SetKey(ActionAttack, ebiten.KeyR)
	if err != nil {
		t.Errorf("SetKey() unexpected error: %v", err)
	}

	key := registry.GetKey(ActionAttack)
	if key != ebiten.KeyR {
		t.Errorf("After SetKey, GetKey(ActionAttack) = %v, want %v", key, ebiten.KeyR)
	}
}

// TestKeyBindingRegistry_SetKey_Conflict tests conflict detection
func TestKeyBindingRegistry_SetKey_Conflict(t *testing.T) {
	registry := NewKeyBindingRegistry()

	// Try to bind a key that's already used
	err := registry.SetKey(ActionAttack, ebiten.KeyW) // W is bound to MoveUp
	if err == nil {
		t.Error("SetKey should return error when key is already bound")
	}

	// Verify original binding unchanged
	if registry.GetKey(ActionAttack) != ebiten.KeySpace {
		t.Error("Failed SetKey should not change binding")
	}
}

// TestKeyBindingRegistry_SetKey_SameAction tests rebinding same action
func TestKeyBindingRegistry_SetKey_SameAction(t *testing.T) {
	registry := NewKeyBindingRegistry()

	// Change ActionAttack from Space to R
	err := registry.SetKey(ActionAttack, ebiten.KeyR)
	if err != nil {
		t.Fatalf("SetKey() unexpected error: %v", err)
	}

	// Change ActionAttack again from R to T (should work - no conflict with itself)
	err = registry.SetKey(ActionAttack, ebiten.KeyT)
	if err != nil {
		t.Errorf("SetKey() should allow rebinding same action: %v", err)
	}

	if registry.GetKey(ActionAttack) != ebiten.KeyT {
		t.Error("Second SetKey should update binding")
	}
}

// TestKeyBindingRegistry_GetKeyLabel tests UI label generation
func TestKeyBindingRegistry_GetKeyLabel(t *testing.T) {
	registry := NewKeyBindingRegistry()

	tests := []struct {
		action Action
		want   string
	}{
		{ActionMoveUp, "W"},
		{ActionAttack, "Space"},
		{ActionHelp, "ESC"},
		{ActionQuickSave, "F5"},
		{ActionCastSpell1, "1"},
	}

	for _, tt := range tests {
		got := registry.GetKeyLabel(tt.action)
		if got != tt.want {
			t.Errorf("GetKeyLabel(%s) = %q, want %q", tt.action.String(), got, tt.want)
		}
	}
}

// TestKeyBindingRegistry_GetActionLabel tests formatted action labels
func TestKeyBindingRegistry_GetActionLabel(t *testing.T) {
	registry := NewKeyBindingRegistry()

	tests := []struct {
		action Action
		want   string
	}{
		{ActionMoveUp, "Move Up [W]"},
		{ActionAttack, "Attack [Space]"},
		{ActionInventory, "Inventory [I]"},
		{ActionHelp, "Help/Menu [ESC]"},
	}

	for _, tt := range tests {
		got := registry.GetActionLabel(tt.action)
		if got != tt.want {
			t.Errorf("GetActionLabel(%s) = %q, want %q", tt.action.String(), got, tt.want)
		}
	}
}

// TestKeyBindingRegistry_GetAllBindings tests binding enumeration
func TestKeyBindingRegistry_GetAllBindings(t *testing.T) {
	registry := NewKeyBindingRegistry()

	bindings := registry.GetAllBindings()

	if len(bindings) == 0 {
		t.Error("GetAllBindings() returned empty map")
	}

	// Verify it's a copy (mutation doesn't affect original)
	bindings[ActionAttack] = ebiten.KeyZ
	if registry.GetKey(ActionAttack) == ebiten.KeyZ {
		t.Error("GetAllBindings() should return a copy, not a reference")
	}
}

// TestKeyBindingRegistry_ResetToDefaults tests reset functionality
func TestKeyBindingRegistry_ResetToDefaults(t *testing.T) {
	registry := NewKeyBindingRegistry()

	// Change some bindings
	registry.SetKey(ActionAttack, ebiten.KeyR)
	registry.SetKey(ActionInventory, ebiten.KeyY)

	// Verify changes took effect
	if registry.GetKey(ActionAttack) != ebiten.KeyR {
		t.Error("Setup failed: binding not changed")
	}

	// Reset to defaults
	registry.ResetToDefaults()

	// Verify defaults restored
	if registry.GetKey(ActionAttack) != ebiten.KeySpace {
		t.Error("ResetToDefaults() did not restore ActionAttack")
	}
	if registry.GetKey(ActionInventory) != ebiten.KeyI {
		t.Error("ResetToDefaults() did not restore ActionInventory")
	}
}

// TestKeyName tests key name formatting
func TestKeyName(t *testing.T) {
	tests := []struct {
		key  ebiten.Key
		want string
	}{
		// Letters
		{ebiten.KeyA, "A"},
		{ebiten.KeyZ, "Z"},
		{ebiten.KeyW, "W"},

		// Numbers
		{ebiten.Key0, "0"},
		{ebiten.Key5, "5"},
		{ebiten.Key9, "9"},

		// Special keys
		{ebiten.KeySpace, "Space"},
		{ebiten.KeyEnter, "Enter"},
		{ebiten.KeyEscape, "ESC"},
		{ebiten.KeyTab, "Tab"},
		{ebiten.KeyBackspace, "Backspace"},
		{ebiten.KeyShiftLeft, "Shift"},
		{ebiten.KeyShiftRight, "Shift"},
		{ebiten.KeyControlLeft, "Ctrl"},
		{ebiten.KeyAltLeft, "Alt"},

		// Arrow keys
		{ebiten.KeyUp, "↑"},
		{ebiten.KeyDown, "↓"},
		{ebiten.KeyLeft, "←"},
		{ebiten.KeyRight, "→"},

		// Function keys
		{ebiten.KeyF1, "F1"},
		{ebiten.KeyF5, "F5"},
		{ebiten.KeyF9, "F9"},
		{ebiten.KeyF12, "F12"},

		// Unknown key
		{ebiten.KeyMax, "?"},
	}

	for _, tt := range tests {
		got := KeyName(tt.key)
		if got != tt.want {
			t.Errorf("KeyName(%v) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// TestKeyBindingRegistry_AllActionsHaveBindings tests completeness
func TestKeyBindingRegistry_AllActionsHaveBindings(t *testing.T) {
	registry := NewKeyBindingRegistry()

	// Test all actions up to ActionCount
	for action := Action(0); action < ActionCount; action++ {
		key := registry.GetKey(action)
		if key == ebiten.KeyMax {
			t.Errorf("Action %s (%d) has no binding", action.String(), action)
		}
	}
}

// TestGlobalKeyBindings tests the global instance
func TestGlobalKeyBindings(t *testing.T) {
	if GlobalKeyBindings == nil {
		t.Fatal("GlobalKeyBindings should be initialized")
	}

	// Verify it has default bindings
	if len(GlobalKeyBindings.bindings) == 0 {
		t.Error("GlobalKeyBindings should have default bindings")
	}

	// Verify a few key bindings
	if GlobalKeyBindings.GetKey(ActionMoveUp) != ebiten.KeyW {
		t.Error("GlobalKeyBindings has incorrect default for ActionMoveUp")
	}
}

// TestKeyBindingRegistry_MultipleSetKey tests multiple rebindings
func TestKeyBindingRegistry_MultipleSetKey(t *testing.T) {
	registry := NewKeyBindingRegistry()

	// Swap two bindings (requires multiple operations)
	// Save original keys
	originalAttack := registry.GetKey(ActionAttack)       // Space
	originalInventory := registry.GetKey(ActionInventory) // I

	// Rebind Attack to a free key first
	if err := registry.SetKey(ActionAttack, ebiten.KeyR); err != nil {
		t.Fatalf("Failed to rebind ActionAttack: %v", err)
	}

	// Now Inventory can use Space
	if err := registry.SetKey(ActionInventory, originalAttack); err != nil {
		t.Fatalf("Failed to rebind ActionInventory to Space: %v", err)
	}

	// Verify final state
	if registry.GetKey(ActionAttack) != ebiten.KeyR {
		t.Error("ActionAttack should be bound to R")
	}
	if registry.GetKey(ActionInventory) != ebiten.KeySpace {
		t.Error("ActionInventory should be bound to Space")
	}

	// Restore for other tests
	registry.SetKey(ActionAttack, originalAttack)
	registry.SetKey(ActionInventory, originalInventory)
}
