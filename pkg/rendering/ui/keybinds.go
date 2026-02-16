// Package ui provides keybind customization for Phase 60.1.
package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// KeyAction represents an in-game action that can be bound to a key
type KeyAction string

const (
	// Movement actions
	ActionMoveUp    KeyAction = "move_up"
	ActionMoveDown  KeyAction = "move_down"
	ActionMoveLeft  KeyAction = "move_left"
	ActionMoveRight KeyAction = "move_right"
	ActionSprint    KeyAction = "sprint"

	// Combat actions
	ActionAttack   KeyAction = "attack"
	ActionBlock    KeyAction = "block"
	ActionDodge    KeyAction = "dodge"
	ActionAbility1 KeyAction = "ability_1"
	ActionAbility2 KeyAction = "ability_2"
	ActionAbility3 KeyAction = "ability_3"
	ActionAbility4 KeyAction = "ability_4"
	ActionUltimate KeyAction = "ultimate"

	// Inventory actions
	ActionInventory  KeyAction = "inventory"
	ActionPickup     KeyAction = "pickup"
	ActionDrop       KeyAction = "drop"
	ActionUseItem    KeyAction = "use_item"
	ActionQuickSlot1 KeyAction = "quick_slot_1"
	ActionQuickSlot2 KeyAction = "quick_slot_2"
	ActionQuickSlot3 KeyAction = "quick_slot_3"
	ActionQuickSlot4 KeyAction = "quick_slot_4"

	// UI actions
	ActionMap       KeyAction = "map"
	ActionCharacter KeyAction = "character"
	ActionSkills    KeyAction = "skills"
	ActionQuests    KeyAction = "quests"
	ActionCrafting  KeyAction = "crafting"
	ActionSettings  KeyAction = "settings"
	ActionEscape    KeyAction = "escape"
	ActionChat      KeyAction = "chat"
	ActionEmote     KeyAction = "emote"

	// Social actions
	ActionTrade   KeyAction = "trade"
	ActionParty   KeyAction = "party"
	ActionGuild   KeyAction = "guild"
	ActionFriend  KeyAction = "friend"
	ActionWhisper KeyAction = "whisper"

	// Vehicle actions
	ActionVehicleMount    KeyAction = "vehicle_mount"
	ActionVehicleDismount KeyAction = "vehicle_dismount"
	ActionVehicleBoost    KeyAction = "vehicle_boost"
	ActionVehicleBrake    KeyAction = "vehicle_brake"

	// Companion actions
	ActionCompanionFollow KeyAction = "companion_follow"
	ActionCompanionStay   KeyAction = "companion_stay"
	ActionCompanionAttack KeyAction = "companion_attack"

	// Housing actions
	ActionHousing        KeyAction = "housing"
	ActionPlaceFurniture KeyAction = "place_furniture"
	ActionEditMode       KeyAction = "edit_mode"

	// Misc actions
	ActionScreenshot  KeyAction = "screenshot"
	ActionQuickTravel KeyAction = "quick_travel"
	ActionAutoRun     KeyAction = "auto_run"
	ActionTarget      KeyAction = "target"
	ActionInteract    KeyAction = "interact"
)

// Key represents keyboard/mouse inputs
type Key string

const (
	// Letter keys
	KeyA Key = "A"
	KeyB Key = "B"
	KeyC Key = "C"
	KeyD Key = "D"
	KeyE Key = "E"
	KeyF Key = "F"
	KeyG Key = "G"
	KeyH Key = "H"
	KeyI Key = "I"
	KeyJ Key = "J"
	KeyK Key = "K"
	KeyL Key = "L"
	KeyM Key = "M"
	KeyN Key = "N"
	KeyO Key = "O"
	KeyP Key = "P"
	KeyQ Key = "Q"
	KeyR Key = "R"
	KeyS Key = "S"
	KeyT Key = "T"
	KeyU Key = "U"
	KeyV Key = "V"
	KeyW Key = "W"
	KeyX Key = "X"
	KeyY Key = "Y"
	KeyZ Key = "Z"

	// Number keys
	Key0 Key = "0"
	Key1 Key = "1"
	Key2 Key = "2"
	Key3 Key = "3"
	Key4 Key = "4"
	Key5 Key = "5"
	Key6 Key = "6"
	Key7 Key = "7"
	Key8 Key = "8"
	Key9 Key = "9"

	// Function keys
	KeyF1  Key = "F1"
	KeyF2  Key = "F2"
	KeyF3  Key = "F3"
	KeyF4  Key = "F4"
	KeyF5  Key = "F5"
	KeyF6  Key = "F6"
	KeyF7  Key = "F7"
	KeyF8  Key = "F8"
	KeyF9  Key = "F9"
	KeyF10 Key = "F10"
	KeyF11 Key = "F11"
	KeyF12 Key = "F12"

	// Special keys
	KeySpace     Key = "Space"
	KeyEnter     Key = "Enter"
	KeyEscape    Key = "Escape"
	KeyTab       Key = "Tab"
	KeyShift     Key = "Shift"
	KeyControl   Key = "Control"
	KeyAlt       Key = "Alt"
	KeyBackspace Key = "Backspace"

	// Arrow keys
	KeyUp    Key = "Up"
	KeyDown  Key = "Down"
	KeyLeft  Key = "Left"
	KeyRight Key = "Right"

	// Mouse buttons
	MouseLeft   Key = "MouseLeft"
	MouseRight  Key = "MouseRight"
	MouseMiddle Key = "MouseMiddle"
	Mouse4      Key = "Mouse4"
	Mouse5      Key = "Mouse5"

	// Numpad keys
	NumPad0 Key = "NumPad0"
	NumPad1 Key = "NumPad1"
	NumPad2 Key = "NumPad2"
	NumPad3 Key = "NumPad3"
	NumPad4 Key = "NumPad4"
	NumPad5 Key = "NumPad5"
	NumPad6 Key = "NumPad6"
	NumPad7 Key = "NumPad7"
	NumPad8 Key = "NumPad8"
	NumPad9 Key = "NumPad9"
)

// Keybind maps an action to a key
type Keybind struct {
	Action       KeyAction
	PrimaryKey   Key
	SecondaryKey Key // optional alternate binding
	Description  string
}

// KeybindManager manages all keybindings
type KeybindManager struct {
	mu       sync.RWMutex
	bindings map[KeyAction]*Keybind
	keyMap   map[Key]KeyAction // reverse lookup
}

// NewKeybindManager creates a new keybind manager with defaults
func NewKeybindManager() *KeybindManager {
	km := &KeybindManager{
		bindings: make(map[KeyAction]*Keybind),
		keyMap:   make(map[Key]KeyAction),
	}
	km.registerDefaultBindings()
	return km
}

// registerDefaultBindings sets up default keybindings
func (km *KeybindManager) registerDefaultBindings() {
	defaults := []Keybind{
		// Movement (WASD)
		{ActionMoveUp, KeyW, KeyUp, "Move up"},
		{ActionMoveDown, KeyS, KeyDown, "Move down"},
		{ActionMoveLeft, KeyA, KeyLeft, "Move left"},
		{ActionMoveRight, KeyD, KeyRight, "Move right"},
		{ActionSprint, KeyShift, "", "Sprint"},

		// Combat
		{ActionAttack, MouseLeft, "", "Attack"},
		{ActionBlock, MouseRight, "", "Block/Defend"},
		{ActionDodge, KeySpace, "", "Dodge roll"},
		{ActionAbility1, Key1, "", "Ability slot 1"},
		{ActionAbility2, Key2, "", "Ability slot 2"},
		{ActionAbility3, Key3, "", "Ability slot 3"},
		{ActionAbility4, Key4, "", "Ability slot 4"},
		{ActionUltimate, KeyR, "", "Ultimate ability"},

		// Inventory
		{ActionInventory, KeyI, "", "Open inventory"},
		{ActionPickup, KeyE, "", "Pick up item"},
		{ActionDrop, KeyG, "", "Drop item"},
		{ActionUseItem, KeyF, "", "Use item"},
		{ActionQuickSlot1, Key5, "", "Quick item slot 1"},
		{ActionQuickSlot2, Key6, "", "Quick item slot 2"},
		{ActionQuickSlot3, Key7, "", "Quick item slot 3"},
		{ActionQuickSlot4, Key8, "", "Quick item slot 4"},

		// UI
		{ActionMap, KeyM, "", "Open map"},
		{ActionCharacter, KeyC, "", "Character sheet"},
		{ActionSkills, KeyK, "", "Skills menu"},
		{ActionQuests, KeyJ, "", "Quest log"},
		{ActionCrafting, KeyN, "", "Crafting menu"},
		{ActionSettings, KeyEscape, "", "Settings menu"},
		{ActionChat, KeyEnter, "", "Chat"},
		{ActionEmote, KeyY, "", "Emote menu"},

		// Social
		{ActionTrade, KeyT, "", "Trade with player"},
		{ActionParty, KeyP, "", "Party menu"},
		{ActionGuild, KeyO, "", "Guild menu"},
		{ActionFriend, KeyF11, "", "Friends list"},
		{ActionWhisper, KeyU, "", "Whisper player"},

		// Vehicle (consolidated bindings)
		{ActionVehicleMount, KeyV, "", "Mount/dismount vehicle"},
		{ActionVehicleDismount, NumPad5, "", "Dismount vehicle"},
		{ActionVehicleBoost, KeyControl, "", "Vehicle boost"},
		{ActionVehicleBrake, KeyAlt, "", "Vehicle brake"},

		// Companion
		{ActionCompanionFollow, KeyF1, "", "Companion: Follow"},
		{ActionCompanionStay, KeyF2, "", "Companion: Stay"},
		{ActionCompanionAttack, KeyF3, "", "Companion: Attack"},

		// Housing
		{ActionHousing, KeyH, "", "Housing menu"},
		{ActionPlaceFurniture, KeyB, "", "Place furniture"},
		{ActionEditMode, KeyL, "", "Edit mode"},

		// Misc
		{ActionScreenshot, KeyF12, "", "Take screenshot"},
		{ActionQuickTravel, KeyF10, "", "Quick travel"},
		{ActionAutoRun, NumPad0, "", "Auto-run"},
		{ActionTarget, KeyTab, "", "Target nearest"},
		{ActionInteract, KeyZ, "", "Interact"},
	}

	for _, kb := range defaults {
		km.bindings[kb.Action] = &kb
		km.keyMap[kb.PrimaryKey] = kb.Action
		if kb.SecondaryKey != "" {
			km.keyMap[kb.SecondaryKey] = kb.Action
		}
	}
}

// SetBinding updates a keybinding
func (km *KeybindManager) SetBinding(action KeyAction, primary, secondary Key) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Check for conflicts
	if conflict := km.findConflict(primary, action); conflict != "" {
		return fmt.Errorf("key %s already bound to %s", primary, conflict)
	}
	if secondary != "" {
		if conflict := km.findConflict(secondary, action); conflict != "" {
			return fmt.Errorf("key %s already bound to %s", secondary, conflict)
		}
	}

	// Clear old key mappings
	if kb, exists := km.bindings[action]; exists {
		delete(km.keyMap, kb.PrimaryKey)
		if kb.SecondaryKey != "" {
			delete(km.keyMap, kb.SecondaryKey)
		}
	}

	// Update binding
	kb, exists := km.bindings[action]
	if !exists {
		kb = &Keybind{Action: action}
		km.bindings[action] = kb
	}

	kb.PrimaryKey = primary
	kb.SecondaryKey = secondary

	// Update reverse map
	km.keyMap[primary] = action
	if secondary != "" {
		km.keyMap[secondary] = action
	}

	return nil
}

// findConflict checks if a key is already bound (excluding given action)
func (km *KeybindManager) findConflict(key Key, exclude KeyAction) KeyAction {
	if action, exists := km.keyMap[key]; exists && action != exclude {
		return action
	}
	return ""
}

// GetBinding retrieves a keybinding for an action
func (km *KeybindManager) GetBinding(action KeyAction) (*Keybind, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	kb, exists := km.bindings[action]
	if !exists {
		return nil, fmt.Errorf("no binding for action: %s", action)
	}
	return kb, nil
}

// GetActionForKey returns the action bound to a key
func (km *KeybindManager) GetActionForKey(key Key) KeyAction {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.keyMap[key]
}

// ListAllBindings returns all keybindings
func (km *KeybindManager) ListAllBindings() []*Keybind {
	km.mu.RLock()
	defer km.mu.RUnlock()

	result := make([]*Keybind, 0, len(km.bindings))
	for _, kb := range km.bindings {
		result = append(result, kb)
	}
	return result
}

// DetectConflicts returns all conflicting keybindings
func (km *KeybindManager) DetectConflicts() []string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	conflicts := make([]string, 0)
	seen := make(map[Key][]KeyAction)

	for action, kb := range km.bindings {
		seen[kb.PrimaryKey] = append(seen[kb.PrimaryKey], action)
		if kb.SecondaryKey != "" {
			seen[kb.SecondaryKey] = append(seen[kb.SecondaryKey], action)
		}
	}

	for key, actions := range seen {
		if len(actions) > 1 {
			conflicts = append(conflicts, fmt.Sprintf("Key %s bound to multiple actions: %v", key, actions))
		}
	}

	return conflicts
}

// ResetToDefaults resets all keybindings to defaults
func (km *KeybindManager) ResetToDefaults() {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.bindings = make(map[KeyAction]*Keybind)
	km.keyMap = make(map[Key]KeyAction)
	km.registerDefaultBindings()
}

// Save persists keybindings to a file
func (km *KeybindManager) Save(filename string) error {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// Create bindings data
	data := make(map[string]map[string]string)
	for action, kb := range km.bindings {
		data[string(action)] = map[string]string{
			"primary":     string(kb.PrimaryKey),
			"secondary":   string(kb.SecondaryKey),
			"description": kb.Description,
		}
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keybindings: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write keybindings file: %w", err)
	}

	return nil
}

// Load reads keybindings from a file
func (km *KeybindManager) Load(filename string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Read file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read keybindings file: %w", err)
	}

	// Unmarshal JSON
	var data map[string]map[string]string
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal keybindings: %w", err)
	}

	// Clear existing bindings
	km.bindings = make(map[KeyAction]*Keybind)
	km.keyMap = make(map[Key]KeyAction)

	// Apply bindings
	for actionStr, keys := range data {
		action := KeyAction(actionStr)
		primary := Key(keys["primary"])
		secondary := Key(keys["secondary"])

		kb := &Keybind{
			Action:       action,
			PrimaryKey:   primary,
			SecondaryKey: secondary,
			Description:  keys["description"],
		}

		km.bindings[action] = kb
		km.keyMap[primary] = action
		if secondary != "" {
			km.keyMap[secondary] = action
		}
	}

	return nil
}
