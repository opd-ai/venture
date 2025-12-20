// Package engine provides centralized menu key configuration.
// This file defines standard menu navigation keys used throughout the UI systems.
package engine

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/mobile"
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
	Inventory    ebiten.Key // I - Inventory management
	Character    ebiten.Key // C - Character stats and equipment
	Skills       ebiten.Key // K - Skill tree
	Quests       ebiten.Key // J - Quest log (J for "Journal")
	Map          ebiten.Key // M - World map
	Shop         ebiten.Key // F - Shop/Merchant interaction (F for "interFact" with merchants)
	Crafting     ebiten.Key // R - Crafting recipes
	Trade        ebiten.Key // T - Player-to-player trading
	Classes      ebiten.Key // X - Advanced classes (multi-classing, prestige, talents) - Changed from A to avoid WASD conflict
	Territory    ebiten.Key // Y - Territory control and guild warfare
	Dialog       ebiten.Key // D - Dialog with NPCs
	Statistics   ebiten.Key // N - Player statistics (N for "Numbers")
	Achievements ebiten.Key // H - Achievements (H for "Honor")

	// Universal exit key (works for all menus)
	Exit ebiten.Key // Escape - Universal menu closer

	// Menu navigation description strings for UI display
	InventoryLabel    string
	CharacterLabel    string
	SkillsLabel       string
	QuestsLabel       string
	MapLabel          string
	ShopLabel         string
	CraftingLabel     string
	TradeLabel        string
	ClassesLabel      string
	TerritoryLabel    string
	DialogLabel       string
	StatisticsLabel   string
	AchievementsLabel string
	ExitHint          string // Standard exit hint text
}{
	// Key assignments
	Inventory:    ebiten.KeyI,
	Character:    ebiten.KeyC,
	Skills:       ebiten.KeyK,
	Quests:       ebiten.KeyJ,
	Map:          ebiten.KeyM,
	Shop:         ebiten.KeyF,
	Crafting:     ebiten.KeyR,
	Trade:        ebiten.KeyT,
	Classes:      ebiten.KeyX, // Changed from A to avoid WASD movement conflict
	Territory:    ebiten.KeyY,
	Dialog:       ebiten.KeyD,
	Statistics:   ebiten.KeyN,
	Achievements: ebiten.KeyH,
	Exit:         ebiten.KeyEscape,

	// Display labels
	InventoryLabel:    "[I] Inventory",
	CharacterLabel:    "[C] Character",
	SkillsLabel:       "[K] Skills",
	QuestsLabel:       "[J] Quests",
	MapLabel:          "[M] Map",
	ShopLabel:         "[F] Shop",
	CraftingLabel:     "[R] Crafting",
	TradeLabel:        "[T] Trade",
	ClassesLabel:      "[X] Classes", // Changed from A to avoid WASD conflict
	TerritoryLabel:    "[Y] Territory",
	DialogLabel:       "[D] Dialog",
	StatisticsLabel:   "[N] Statistics",
	AchievementsLabel: "[H] Achievements",
	ExitHint:          "Press [KEY] or [ESC] to close",
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
	// BUG FIX: Phase 1.2 - Android back button maps to Escape key in Ebiten
	// This handles both desktop Escape key AND Android back button
	// Platform: Desktop (Escape), Android (back button → Escape), iOS (N/A)
	if isVisible && IsKeyJustPressed(MenuKeys.Exit) {
		return true, false // Always close, never open
	}

	return false, false
}

// HandleMenuInputWithTouch provides mobile-enhanced menu input handling with gesture support.
// BUG FIX: Phase 1.2 - Mobile menu exit gestures missing
// Resolution: Added swipe-down and edge-swipe gestures for iOS/Android menu closing
// Platform: Mobile (Android/iOS touch gestures)
//
// Parameters:
//   - toggleKey: The menu's assigned toggle key (e.g., ebiten.KeyI for Inventory)
//   - isVisible: Current visibility state of the menu
//   - touchHandler: Touch input handler (nil for non-touch platforms)
//
// Returns:
//   - shouldClose: true if menu should close, false otherwise
//   - shouldToggle: true if toggle key was pressed (open/close), false if gesture/Escape (close only)
//
// Usage:
//
//	if shouldClose, _ := HandleMenuInputWithTouch(MenuKeys.Inventory, ui.visible, ui.touchHandler); shouldClose {
//	    ui.Hide()
//	    return
//	}
func HandleMenuInputWithTouch(toggleKey ebiten.Key, isVisible bool, touchHandler *mobile.TouchInputHandler) (shouldClose, shouldToggle bool) {
	// Check keyboard input first (desktop + Android back button)
	if shouldClose, shouldToggle := HandleMenuInput(toggleKey, isVisible); shouldClose {
		return shouldClose, shouldToggle
	}

	// BUG FIX: Mobile gesture support - only check when menu is visible
	if !isVisible || touchHandler == nil {
		return false, false
	}

	// Check for swipe-down gesture (common mobile dismiss pattern)
	direction, distance, detected := touchHandler.GetSwipe()
	if detected {
		// Convert angle to degrees for easier logic
		// direction is in radians: 0=right, π/2=down, π=left, 3π/2=up
		degrees := direction * 180.0 / math.Pi
		if degrees < 0 {
			degrees += 360
		}

		// Swipe down: 45° to 135° (downward arc)
		// This matches iOS "swipe down to dismiss" pattern
		if degrees >= 45 && degrees <= 135 && distance >= 50 {
			return true, false // Close menu via downward swipe
		}

		// Edge swipe (iOS back gesture): swipe from left edge (135° to 225°)
		// Only trigger if swipe starts near screen edge
		touches := touchHandler.GetActiveTouches()
		if len(touches) > 0 {
			touch := touches[0]
			// Check if swipe started near left or right edge (within 50px)
			if touch.StartX < 50 || touch.StartX > (720-50) { // Assuming 720px width
				// Right-to-left swipe (back gesture)
				if degrees >= 135 && degrees <= 225 && distance >= 75 {
					return true, false // Close menu via edge swipe
				}
			}
		}
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
	case ebiten.KeyF:
		return "F"
	case ebiten.KeyR:
		return "R"
	case ebiten.KeyT:
		return "T"
	case ebiten.KeyA:
		return "A"
	case ebiten.KeyP:
		return "P"
	case ebiten.KeyX:
		return "X"
	case ebiten.KeyY:
		return "Y"
	case ebiten.KeyD:
		return "D"
	case ebiten.KeyN:
		return "N"
	case ebiten.KeyH:
		return "H"
	case ebiten.KeyEscape:
		return "ESC"
	default:
		return "KEY"
	}
}
