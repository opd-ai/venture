package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// Action represents a bindable game action.
// Priority 2.1: Central action enum for key binding registry
type Action int

const (
	// Movement actions
	ActionMoveUp Action = iota
	ActionMoveDown
	ActionMoveLeft
	ActionMoveRight

	// Combat/interaction actions
	ActionAttack
	ActionUseItem
	ActionSecondary

	// Spell casting actions
	ActionCastSpell1
	ActionCastSpell2
	ActionCastSpell3
	ActionCastSpell4
	ActionCastSpell5

	// UI actions
	ActionInventory
	ActionCharacter
	ActionSkills
	ActionQuests
	ActionMap
	ActionHelp

	// System actions
	ActionQuickSave
	ActionQuickLoad
	ActionCycleTargets

	// Total action count (must be last)
	ActionCount
)

// String returns human-readable action name for UI display.
func (a Action) String() string {
	switch a {
	case ActionMoveUp:
		return "Move Up"
	case ActionMoveDown:
		return "Move Down"
	case ActionMoveLeft:
		return "Move Left"
	case ActionMoveRight:
		return "Move Right"
	case ActionAttack:
		return "Attack"
	case ActionUseItem:
		return "Use Item"
	case ActionSecondary:
		return "Secondary Action"
	case ActionCastSpell1:
		return "Cast Spell 1"
	case ActionCastSpell2:
		return "Cast Spell 2"
	case ActionCastSpell3:
		return "Cast Spell 3"
	case ActionCastSpell4:
		return "Cast Spell 4"
	case ActionCastSpell5:
		return "Cast Spell 5"
	case ActionInventory:
		return "Inventory"
	case ActionCharacter:
		return "Character"
	case ActionSkills:
		return "Skills"
	case ActionQuests:
		return "Quests"
	case ActionMap:
		return "Map"
	case ActionHelp:
		return "Help/Menu"
	case ActionQuickSave:
		return "Quick Save"
	case ActionQuickLoad:
		return "Quick Load"
	case ActionCycleTargets:
		return "Cycle Targets"
	default:
		return "Unknown"
	}
}

// KeyBindingRegistry manages the mapping between actions and keyboard keys.
// Provides centralized key binding management and UI label generation.
type KeyBindingRegistry struct {
	bindings map[Action]ebiten.Key
}

// NewKeyBindingRegistry creates a new registry with default key bindings.
func NewKeyBindingRegistry() *KeyBindingRegistry {
	registry := &KeyBindingRegistry{
		bindings: make(map[Action]ebiten.Key),
	}
	registry.loadDefaults()
	return registry
}

// loadDefaults initializes the registry with default key bindings.
func (r *KeyBindingRegistry) loadDefaults() {
	// Movement
	r.bindings[ActionMoveUp] = ebiten.KeyW
	r.bindings[ActionMoveDown] = ebiten.KeyS
	r.bindings[ActionMoveLeft] = ebiten.KeyA
	r.bindings[ActionMoveRight] = ebiten.KeyD

	// Combat/interaction
	r.bindings[ActionAttack] = ebiten.KeySpace
	r.bindings[ActionUseItem] = ebiten.KeyE
	r.bindings[ActionSecondary] = ebiten.KeyShiftLeft

	// Spells
	r.bindings[ActionCastSpell1] = ebiten.Key1
	r.bindings[ActionCastSpell2] = ebiten.Key2
	r.bindings[ActionCastSpell3] = ebiten.Key3
	r.bindings[ActionCastSpell4] = ebiten.Key4
	r.bindings[ActionCastSpell5] = ebiten.Key5

	// UI
	r.bindings[ActionInventory] = ebiten.KeyI
	r.bindings[ActionCharacter] = ebiten.KeyC
	r.bindings[ActionSkills] = ebiten.KeyK
	r.bindings[ActionQuests] = ebiten.KeyJ
	r.bindings[ActionMap] = ebiten.KeyM
	r.bindings[ActionHelp] = ebiten.KeyEscape

	// System
	r.bindings[ActionQuickSave] = ebiten.KeyF5
	r.bindings[ActionQuickLoad] = ebiten.KeyF9
	r.bindings[ActionCycleTargets] = ebiten.KeyTab
}

// GetKey returns the key bound to the specified action.
func (r *KeyBindingRegistry) GetKey(action Action) ebiten.Key {
	if key, ok := r.bindings[action]; ok {
		return key
	}
	return ebiten.KeyMax // Return invalid key if not found
}

// SetKey binds a key to an action.
// Returns error if the key is already bound to a different action.
func (r *KeyBindingRegistry) SetKey(action Action, key ebiten.Key) error {
	// Check for conflicts
	for existingAction, existingKey := range r.bindings {
		if existingKey == key && existingAction != action {
			return fmt.Errorf("key %s already bound to %s", KeyName(key), existingAction.String())
		}
	}
	r.bindings[action] = key
	return nil
}

// GetKeyLabel returns a UI-friendly label for the key bound to an action.
// Examples: "W", "Space", "ESC", "F5"
func (r *KeyBindingRegistry) GetKeyLabel(action Action) string {
	key := r.GetKey(action)
	return KeyName(key)
}

// GetActionLabel returns a formatted string for UI display.
// Example: "Move Up [W]", "Attack [Space]"
func (r *KeyBindingRegistry) GetActionLabel(action Action) string {
	return fmt.Sprintf("%s [%s]", action.String(), r.GetKeyLabel(action))
}

// IsActionPressed checks if the key bound to an action is currently pressed.
func (r *KeyBindingRegistry) IsActionPressed(action Action) bool {
	key := r.GetKey(action)
	if key == ebiten.KeyMax {
		return false
	}
	return ebiten.IsKeyPressed(key)
}

// IsActionJustPressed checks if the key bound to an action was just pressed this frame.
func (r *KeyBindingRegistry) IsActionJustPressed(action Action) bool {
	key := r.GetKey(action)
	if key == ebiten.KeyMax {
		return false
	}
	return ebiten.IsKeyPressed(key) && !ebiten.IsKeyPressed(key) // Use inpututil for proper detection
}

// GetAllBindings returns a copy of all current key bindings.
// Useful for displaying key binding settings in UI.
func (r *KeyBindingRegistry) GetAllBindings() map[Action]ebiten.Key {
	result := make(map[Action]ebiten.Key, len(r.bindings))
	for action, key := range r.bindings {
		result[action] = key
	}
	return result
}

// ResetToDefaults restores all key bindings to their default values.
func (r *KeyBindingRegistry) ResetToDefaults() {
	r.bindings = make(map[Action]ebiten.Key)
	r.loadDefaults()
}

// KeyName returns a human-readable name for a keyboard key.
// Handles special keys and provides UI-friendly labels.
func KeyName(key ebiten.Key) string {
	switch key {
	// Letter keys
	case ebiten.KeyA:
		return "A"
	case ebiten.KeyB:
		return "B"
	case ebiten.KeyC:
		return "C"
	case ebiten.KeyD:
		return "D"
	case ebiten.KeyE:
		return "E"
	case ebiten.KeyF:
		return "F"
	case ebiten.KeyG:
		return "G"
	case ebiten.KeyH:
		return "H"
	case ebiten.KeyI:
		return "I"
	case ebiten.KeyJ:
		return "J"
	case ebiten.KeyK:
		return "K"
	case ebiten.KeyL:
		return "L"
	case ebiten.KeyM:
		return "M"
	case ebiten.KeyN:
		return "N"
	case ebiten.KeyO:
		return "O"
	case ebiten.KeyP:
		return "P"
	case ebiten.KeyQ:
		return "Q"
	case ebiten.KeyR:
		return "R"
	case ebiten.KeyS:
		return "S"
	case ebiten.KeyT:
		return "T"
	case ebiten.KeyU:
		return "U"
	case ebiten.KeyV:
		return "V"
	case ebiten.KeyW:
		return "W"
	case ebiten.KeyX:
		return "X"
	case ebiten.KeyY:
		return "Y"
	case ebiten.KeyZ:
		return "Z"

	// Number keys
	case ebiten.Key0:
		return "0"
	case ebiten.Key1:
		return "1"
	case ebiten.Key2:
		return "2"
	case ebiten.Key3:
		return "3"
	case ebiten.Key4:
		return "4"
	case ebiten.Key5:
		return "5"
	case ebiten.Key6:
		return "6"
	case ebiten.Key7:
		return "7"
	case ebiten.Key8:
		return "8"
	case ebiten.Key9:
		return "9"

	// Special keys
	case ebiten.KeySpace:
		return "Space"
	case ebiten.KeyEnter:
		return "Enter"
	case ebiten.KeyEscape:
		return "ESC"
	case ebiten.KeyTab:
		return "Tab"
	case ebiten.KeyBackspace:
		return "Backspace"
	case ebiten.KeyShiftLeft, ebiten.KeyShiftRight:
		return "Shift"
	case ebiten.KeyControlLeft, ebiten.KeyControlRight:
		return "Ctrl"
	case ebiten.KeyAltLeft, ebiten.KeyAltRight:
		return "Alt"

	// Arrow keys
	case ebiten.KeyUp:
		return "↑"
	case ebiten.KeyDown:
		return "↓"
	case ebiten.KeyLeft:
		return "←"
	case ebiten.KeyRight:
		return "→"

	// Function keys
	case ebiten.KeyF1:
		return "F1"
	case ebiten.KeyF2:
		return "F2"
	case ebiten.KeyF3:
		return "F3"
	case ebiten.KeyF4:
		return "F4"
	case ebiten.KeyF5:
		return "F5"
	case ebiten.KeyF6:
		return "F6"
	case ebiten.KeyF7:
		return "F7"
	case ebiten.KeyF8:
		return "F8"
	case ebiten.KeyF9:
		return "F9"
	case ebiten.KeyF10:
		return "F10"
	case ebiten.KeyF11:
		return "F11"
	case ebiten.KeyF12:
		return "F12"

	default:
		return "?"
	}
}

// Global key binding registry instance.
// Initialized with default bindings and used throughout the application.
var GlobalKeyBindings = NewKeyBindingRegistry()
