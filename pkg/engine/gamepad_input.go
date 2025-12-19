// Package engine provides gamepad input handling for Desktop and WASM platforms.
// This file implements GamepadInputHandler which processes controller input
// for player-controlled entities and game controls.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// GamepadInputHandler manages gamepad/controller input for Desktop platforms.
// Provides unified gamepad support with configurable button mappings.
// Platform support: Desktop (Windows, Linux, macOS) and WASM.
type GamepadInputHandler struct {
	// Active gamepad (first connected controller)
	activeGamepadID ebiten.GamepadID
	hasGamepad      bool

	// Analog stick dead zones
	LeftStickDeadZone  float64
	RightStickDeadZone float64

	// Button mappings (Xbox-style by default)
	ButtonAttack      ebiten.StandardGamepadButton
	ButtonSecondary   ebiten.StandardGamepadButton
	ButtonUseItem     ebiten.StandardGamepadButton
	ButtonMenu        ebiten.StandardGamepadButton
	ButtonInventory   ebiten.StandardGamepadButton
	ButtonMap         ebiten.StandardGamepadButton
	ButtonCycleTarget ebiten.StandardGamepadButton
	ButtonInteract    ebiten.StandardGamepadButton

	// Platform parity fix: UI shortcut button mappings for complete accessibility
	// L3 (left stick click): Character, R3 (right stick click): Skills
	// Note: Quest Log accessible via Menu -> Quest Log submenu (no dedicated button)
	ButtonCharacter ebiten.StandardGamepadButton
	ButtonSkills    ebiten.StandardGamepadButton

	// Spell casting using D-pad + trigger combinations
	ButtonSpell1 ebiten.StandardGamepadButton
	ButtonSpell2 ebiten.StandardGamepadButton
	ButtonSpell3 ebiten.StandardGamepadButton
	ButtonSpell4 ebiten.StandardGamepadButton

	// Trigger thresholds
	TriggerThreshold float64

	// State tracking for edge detection
	lastButtonStates map[ebiten.StandardGamepadButton]bool
}

// NewGamepadInputHandler creates a new gamepad input handler with default mappings.
// Default mappings follow Xbox controller conventions:
// - Left stick: Movement
// - Right stick: Aim
// - A: Attack, B: Secondary, X: Use Item, Y: Interact
// - Start: Menu, Back: Map
// - LB: Cycle Targets, RB: Inventory
// - L3: Character, R3: Skills
// - D-pad: Spell casting (Up=1, Right=2, Down=3, Left=4)
func NewGamepadInputHandler() *GamepadInputHandler {
	return &GamepadInputHandler{
		activeGamepadID:    -1,
		hasGamepad:         false,
		LeftStickDeadZone:  0.15,
		RightStickDeadZone: 0.15,
		TriggerThreshold:   0.5,

		// Xbox-style button mappings
		ButtonAttack:      ebiten.StandardGamepadButtonRightBottom, // A
		ButtonSecondary:   ebiten.StandardGamepadButtonRightRight,  // B
		ButtonUseItem:     ebiten.StandardGamepadButtonRightLeft,   // X
		ButtonInteract:    ebiten.StandardGamepadButtonRightTop,    // Y
		ButtonMenu:        ebiten.StandardGamepadButtonCenterRight, // Start
		ButtonMap:         ebiten.StandardGamepadButtonCenterLeft,  // Back/Select
		ButtonCycleTarget: ebiten.StandardGamepadButtonFrontTopLeft,  // LB
		ButtonInventory:   ebiten.StandardGamepadButtonFrontTopRight, // RB

		// Platform parity fix: L3/R3 for UI shortcuts
		ButtonCharacter: ebiten.StandardGamepadButtonLeftStick,  // L3 (left stick click)
		ButtonSkills:    ebiten.StandardGamepadButtonRightStick, // R3 (right stick click)
		// Note: Quest log accessible via Menu -> Quest Log submenu

		// D-pad for spell casting
		ButtonSpell1: ebiten.StandardGamepadButtonLeftTop,    // D-pad Up
		ButtonSpell2: ebiten.StandardGamepadButtonLeftRight,  // D-pad Right
		ButtonSpell3: ebiten.StandardGamepadButtonLeftBottom, // D-pad Down
		ButtonSpell4: ebiten.StandardGamepadButtonLeftLeft,   // D-pad Left

		lastButtonStates: make(map[ebiten.StandardGamepadButton]bool),
	}
}

// Update checks for connected gamepads and updates active gamepad state.
// Should be called every frame before reading input.
func (g *GamepadInputHandler) Update() {
	// Check for newly connected gamepads
	gamepadIDs := inpututil.AppendJustConnectedGamepadIDs(nil)
	if len(gamepadIDs) > 0 && !g.hasGamepad {
		g.activeGamepadID = gamepadIDs[0]
		g.hasGamepad = true
	}

	// Check if active gamepad disconnected
	if g.hasGamepad {
		connected := false
		for _, id := range ebiten.GamepadIDs() {
			if id == g.activeGamepadID {
				connected = true
				break
			}
		}
		if !connected {
			g.hasGamepad = false
			g.activeGamepadID = -1
		}
	}

	// Try to find any available gamepad if we don't have one
	if !g.hasGamepad {
		gamepadIDs := ebiten.GamepadIDs()
		if len(gamepadIDs) > 0 {
			g.activeGamepadID = gamepadIDs[0]
			g.hasGamepad = true
		}
	}
}

// HasGamepad returns true if a gamepad is connected and active.
func (g *GamepadInputHandler) HasGamepad() bool {
	return g.hasGamepad
}

// GetGamepadID returns the active gamepad ID.
func (g *GamepadInputHandler) GetGamepadID() ebiten.GamepadID {
	return g.activeGamepadID
}

// GetMovementInput returns the left analog stick input for movement.
// Returns normalized values (-1.0 to 1.0) with dead zone applied.
func (g *GamepadInputHandler) GetMovementInput() (x, y float64) {
	if !g.hasGamepad {
		return 0, 0
	}

	// Check if gamepad supports standard layout
	if !ebiten.IsStandardGamepadLayoutAvailable(g.activeGamepadID) {
		return 0, 0
	}

	x = ebiten.StandardGamepadAxisValue(g.activeGamepadID, ebiten.StandardGamepadAxisLeftStickHorizontal)
	y = ebiten.StandardGamepadAxisValue(g.activeGamepadID, ebiten.StandardGamepadAxisLeftStickVertical)

	// Apply dead zone
	x = g.applyDeadZone(x, g.LeftStickDeadZone)
	y = g.applyDeadZone(y, g.LeftStickDeadZone)

	return x, y
}

// GetAimInput returns the right analog stick input for aiming.
// Returns normalized values (-1.0 to 1.0) with dead zone applied.
func (g *GamepadInputHandler) GetAimInput() (x, y float64) {
	if !g.hasGamepad {
		return 0, 0
	}

	if !ebiten.IsStandardGamepadLayoutAvailable(g.activeGamepadID) {
		return 0, 0
	}

	x = ebiten.StandardGamepadAxisValue(g.activeGamepadID, ebiten.StandardGamepadAxisRightStickHorizontal)
	y = ebiten.StandardGamepadAxisValue(g.activeGamepadID, ebiten.StandardGamepadAxisRightStickVertical)

	// Apply dead zone
	x = g.applyDeadZone(x, g.RightStickDeadZone)
	y = g.applyDeadZone(y, g.RightStickDeadZone)

	return x, y
}

// applyDeadZone applies dead zone to an analog value.
func (g *GamepadInputHandler) applyDeadZone(value, deadZone float64) float64 {
	if value > -deadZone && value < deadZone {
		return 0
	}
	// Re-normalize to 0-1 range outside dead zone
	if value > 0 {
		return (value - deadZone) / (1.0 - deadZone)
	}
	return (value + deadZone) / (1.0 - deadZone)
}

// IsButtonPressed returns true if the specified button is currently held.
func (g *GamepadInputHandler) IsButtonPressed(button ebiten.StandardGamepadButton) bool {
	if !g.hasGamepad {
		return false
	}
	if !ebiten.IsStandardGamepadLayoutAvailable(g.activeGamepadID) {
		return false
	}
	return ebiten.IsStandardGamepadButtonPressed(g.activeGamepadID, button)
}

// IsButtonJustPressed returns true only on the frame when button was first pressed.
func (g *GamepadInputHandler) IsButtonJustPressed(button ebiten.StandardGamepadButton) bool {
	if !g.hasGamepad {
		return false
	}
	if !ebiten.IsStandardGamepadLayoutAvailable(g.activeGamepadID) {
		return false
	}
	return inpututil.IsStandardGamepadButtonJustPressed(g.activeGamepadID, button)
}

// IsButtonJustReleased returns true only on the frame when button was released.
func (g *GamepadInputHandler) IsButtonJustReleased(button ebiten.StandardGamepadButton) bool {
	if !g.hasGamepad {
		return false
	}
	if !ebiten.IsStandardGamepadLayoutAvailable(g.activeGamepadID) {
		return false
	}
	return inpututil.IsStandardGamepadButtonJustReleased(g.activeGamepadID, button)
}

// Action input helpers

// IsAttackPressed returns true if attack button is currently pressed.
func (g *GamepadInputHandler) IsAttackPressed() bool {
	return g.IsButtonPressed(g.ButtonAttack)
}

// IsAttackJustPressed returns true on the frame when attack button is first pressed.
func (g *GamepadInputHandler) IsAttackJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonAttack)
}

// IsSecondaryPressed returns true if secondary action button is pressed.
func (g *GamepadInputHandler) IsSecondaryPressed() bool {
	return g.IsButtonPressed(g.ButtonSecondary)
}

// IsSecondaryJustPressed returns true on the frame when secondary button is first pressed.
func (g *GamepadInputHandler) IsSecondaryJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonSecondary)
}

// IsUseItemPressed returns true if use item button is pressed.
func (g *GamepadInputHandler) IsUseItemPressed() bool {
	return g.IsButtonPressed(g.ButtonUseItem)
}

// IsUseItemJustPressed returns true on the frame when use item button is first pressed.
func (g *GamepadInputHandler) IsUseItemJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonUseItem)
}

// IsInteractJustPressed returns true on the frame when interact button is first pressed.
func (g *GamepadInputHandler) IsInteractJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonInteract)
}

// IsMenuJustPressed returns true on the frame when menu button is first pressed.
func (g *GamepadInputHandler) IsMenuJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonMenu)
}

// IsMapJustPressed returns true on the frame when map button is first pressed.
func (g *GamepadInputHandler) IsMapJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonMap)
}

// IsInventoryJustPressed returns true on the frame when inventory button is first pressed.
func (g *GamepadInputHandler) IsInventoryJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonInventory)
}

// IsCycleTargetJustPressed returns true on the frame when cycle target button is first pressed.
func (g *GamepadInputHandler) IsCycleTargetJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonCycleTarget)
}

// Platform parity fix: UI shortcut button helpers

// IsCharacterJustPressed returns true on the frame when character button is first pressed.
// Platform parity fix: Provides gamepad equivalent of 'C' key on Desktop.
func (g *GamepadInputHandler) IsCharacterJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonCharacter)
}

// IsSkillsJustPressed returns true on the frame when skills button is first pressed.
// Platform parity fix: Provides gamepad equivalent of 'K' key on Desktop.
func (g *GamepadInputHandler) IsSkillsJustPressed() bool {
	return g.IsButtonJustPressed(g.ButtonSkills)
}

// Spell input helpers

// IsSpellPressed returns true if the specified spell button (1-4) is pressed.
func (g *GamepadInputHandler) IsSpellPressed(slot int) bool {
	switch slot {
	case 1:
		return g.IsButtonPressed(g.ButtonSpell1)
	case 2:
		return g.IsButtonPressed(g.ButtonSpell2)
	case 3:
		return g.IsButtonPressed(g.ButtonSpell3)
	case 4:
		return g.IsButtonPressed(g.ButtonSpell4)
	default:
		return false
	}
}

// IsSpellJustPressed returns true on the frame when spell button is first pressed.
func (g *GamepadInputHandler) IsSpellJustPressed(slot int) bool {
	switch slot {
	case 1:
		return g.IsButtonJustPressed(g.ButtonSpell1)
	case 2:
		return g.IsButtonJustPressed(g.ButtonSpell2)
	case 3:
		return g.IsButtonJustPressed(g.ButtonSpell3)
	case 4:
		return g.IsButtonJustPressed(g.ButtonSpell4)
	default:
		return false
	}
}

// GetLeftTrigger returns the left trigger value (0.0 to 1.0).
func (g *GamepadInputHandler) GetLeftTrigger() float64 {
	if !g.hasGamepad || !ebiten.IsStandardGamepadLayoutAvailable(g.activeGamepadID) {
		return 0
	}
	// Triggers are mapped to buttons in standard layout, not axes
	// Check if front bottom left is pressed (LT)
	if ebiten.IsStandardGamepadButtonPressed(g.activeGamepadID, ebiten.StandardGamepadButtonFrontBottomLeft) {
		return 1.0
	}
	return 0
}

// GetRightTrigger returns the right trigger value (0.0 to 1.0).
func (g *GamepadInputHandler) GetRightTrigger() float64 {
	if !g.hasGamepad || !ebiten.IsStandardGamepadLayoutAvailable(g.activeGamepadID) {
		return 0
	}
	// Check if front bottom right is pressed (RT)
	if ebiten.IsStandardGamepadButtonPressed(g.activeGamepadID, ebiten.StandardGamepadButtonFrontBottomRight) {
		return 1.0
	}
	return 0
}

// IsLeftTriggerPressed returns true if left trigger exceeds threshold.
func (g *GamepadInputHandler) IsLeftTriggerPressed() bool {
	return g.GetLeftTrigger() >= g.TriggerThreshold
}

// IsRightTriggerPressed returns true if right trigger exceeds threshold.
func (g *GamepadInputHandler) IsRightTriggerPressed() bool {
	return g.GetRightTrigger() >= g.TriggerThreshold
}

// GetGamepadName returns the name of the connected gamepad, or empty string if none.
func (g *GamepadInputHandler) GetGamepadName() string {
	if !g.hasGamepad {
		return ""
	}
	return ebiten.GamepadName(g.activeGamepadID)
}
