// Package engine provides centralized menu key configuration.
// This file defines standard menu navigation keys used throughout the UI systems.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// MenuKeys defines the standard key bindings for all in-game menus.
// Each menu is assigned a unique, mnemonic letter key for opening/toggling.
// All menus support TWO exit mechanisms:
//   - Toggle key: The same key that opened the menu (e.g., pressing "I" again closes Inventory)
//   - Universal exit: The Escape key closes any open menu
//
// This standardization ensures consistent, predictable navigation across all UI screens.
var MenuKeys = struct {
	// Menu activation keys (mnemonic letter assignments)
	Inventory ebiten.Key // I - Inventory management
	Character ebiten.Key // C - Character stats and equipment
	Skills    ebiten.Key // K - Skill tree
	Quests    ebiten.Key // J - Quest log (J for "Journal")
	Map       ebiten.Key // M - World map
	Shop      ebiten.Key // F - Shop/Merchant interaction (F for "interFact" with merchants)
	Crafting  ebiten.Key // R - Crafting recipes

	// Universal exit key (works for all menus)
	Exit ebiten.Key // Escape - Universal menu closer

	// Menu navigation description strings for UI display
	InventoryLabel string
	CharacterLabel string
	SkillsLabel    string
	QuestsLabel    string
	MapLabel       string
	ShopLabel      string
	CraftingLabel  string
	ExitHint       string // Standard exit hint text
}{
	// Key assignments
	Inventory: ebiten.KeyI,
	Character: ebiten.KeyC,
	Skills:    ebiten.KeyK,
	Quests:    ebiten.KeyJ,
	Map:       ebiten.KeyM,
	Shop:      ebiten.KeyF,
	Crafting:  ebiten.KeyR,
	Exit:      ebiten.KeyEscape,

	// Display labels
	InventoryLabel: "[I] Inventory",
	CharacterLabel: "[C] Character",
	SkillsLabel:    "[K] Skills",
	QuestsLabel:    "[J] Quests",
	MapLabel:       "[M] Map",
	ShopLabel:      "[F] Shop",
	CraftingLabel:  "[R] Crafting",
	ExitHint:       "Press [KEY] or [ESC] to close",
}

// HandleMenuInput provides standardized dual-exit menu input handling.
// Returns true if the menu should be closed (either by toggle key or Escape).
//
// Parameters:
//   - toggleKey: The menu's assigned toggle key (e.g., ebiten.KeyI for Inventory)
//   - isVisible: Current visibility state of the menu
//
// Returns:
//   - shouldClose: true if menu should close, false otherwise
//   - shouldToggle: true if toggle key was pressed (open/close), false if Escape was pressed (close only)
//
// Usage:
//
//	if shouldClose, _ := HandleMenuInput(ebiten.KeyI, ui.visible); shouldClose {
//	    ui.visible = false
//	    return
//	}
func HandleMenuInput(toggleKey ebiten.Key, isVisible bool) (shouldClose, shouldToggle bool) {
	// Check for toggle key (works whether menu is open or closed)
	if IsKeyJustPressed(toggleKey) {
		return true, true // Close if open, open if closed
	}

	// Check for Escape key (only works when menu is open)
	if isVisible && IsKeyJustPressed(MenuKeys.Exit) {
		return true, false // Always close, never open
	}

	return false, false
}

// IsKeyJustPressed is a convenience wrapper for inpututil.IsKeyJustPressed.
// Used internally by HandleMenuInput to detect single key press events.
func IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}

// GetExitHint returns a standardized exit hint string for the given menu key.
// This centralizes exit hint formatting to ensure consistency across all menus.
//
// Parameters:
//   - menuKey: The menu's toggle key (e.g., MenuKeys.Inventory)
//
// Returns:
//   - A formatted string like "Press [I] or [ESC] to close"
//
// Usage:
//
//	exitHint := GetExitHint(MenuKeys.Inventory)
//	ebitenutil.DebugPrintAt(screen, exitHint, x, y)
func GetExitHint(menuKey ebiten.Key) string {
	keyName := getKeyName(menuKey)
	return "Press [" + keyName + "] or [ESC] to close"
}

// getKeyName converts an ebiten.Key to its display name.
// Returns uppercase letter for standard keys, or full name for special keys.
func getKeyName(key ebiten.Key) string {
	switch key {
	case ebiten.KeyI:
		return "I"
	case ebiten.KeyC:
		return "C"
	case ebiten.KeyK:
		return "K"
	case ebiten.KeyJ:
		return "J"
	case ebiten.KeyM:
		return "M"
	case ebiten.KeyS:
		return "S"
	case ebiten.KeyR:
		return "R"
	case ebiten.KeyEscape:
		return "ESC"
	default:
		return "KEY"
	}
}
