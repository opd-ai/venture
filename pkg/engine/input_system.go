//go:build !test
// +build !test

// Package engine provides player input handling.
// This file implements InputSystem which processes keyboard, mouse, and touch input
// for player-controlled entities and game controls.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/mobile"
)

// InputComponent stores the current input state for an entity.
// This is typically only used for player-controlled entities.
type InputComponent struct {
	// Movement input (-1.0 to 1.0 for each axis)
	MoveX, MoveY float64

	// Action buttons
	ActionPressed   bool
	SecondaryAction bool
	UseItemPressed  bool

	// GAP-002 REPAIR: Spell casting input flags (keys 1-5)
	Spell1Pressed bool
	Spell2Pressed bool
	Spell3Pressed bool
	Spell4Pressed bool
	Spell5Pressed bool

	// Mouse state
	MouseX, MouseY int
	MousePressed   bool
}

// Type returns the component type identifier.
func (i *InputComponent) Type() string {
	return "input"
}

// InputSystem processes keyboard, mouse, and touch input and updates input components.
type InputSystem struct {
	// Movement speed multiplier
	MoveSpeed float64

	// Key bindings - Movement
	KeyUp    ebiten.Key
	KeyDown  ebiten.Key
	KeyLeft  ebiten.Key
	KeyRight ebiten.Key

	// Key bindings - Actions
	KeyAction  ebiten.Key
	KeyUseItem ebiten.Key

	// GAP-002 REPAIR: Spell casting key bindings (keys 1-5)
	KeySpell1 ebiten.Key
	KeySpell2 ebiten.Key
	KeySpell3 ebiten.Key
	KeySpell4 ebiten.Key
	KeySpell5 ebiten.Key

	// Key bindings - UI
	KeyInventory ebiten.Key // I key for inventory
	KeyCharacter ebiten.Key // C key for character screen
	KeySkills    ebiten.Key // K key for skills screen
	KeyQuests    ebiten.Key // J key for quest log
	KeyMap       ebiten.Key // M key for map

	// Key bindings - System
	KeyHelp         ebiten.Key // ESC key for help menu
	KeyQuickSave    ebiten.Key // F5 key for quick save
	KeyQuickLoad    ebiten.Key // F9 key for quick load
	KeyCycleTargets ebiten.Key // Tab key for cycling targets

	// References to game systems for special key handling
	helpSystem     *HelpSystem
	tutorialSystem *TutorialSystem
	menuSystem     *MenuSystem

	// Mobile input support
	touchHandler    *mobile.TouchInputHandler
	virtualControls *mobile.VirtualControlsLayout
	mobileEnabled   bool
	useTouchInput   bool // Auto-detected or manually set

	// Mouse state tracking for delta calculation (BUG-010 fix)
	lastMouseX, lastMouseY   int
	mouseDeltaX, mouseDeltaY int

	// Callbacks for UI and save/load operations
	onQuickSave     func() error
	onQuickLoad     func() error
	onInventoryOpen func()
	onCharacterOpen func()
	onSkillsOpen    func()
	onQuestsOpen    func()
	onMapOpen       func()
	onCycleTargets  func()
	onMenuToggle    func() // Callback for ESC menu toggle
}

// NewInputSystem creates a new input system with default key bindings.
func NewInputSystem() *InputSystem {
	return &InputSystem{
		MoveSpeed: 100.0, // pixels per second

		// Movement keys
		KeyUp:    ebiten.KeyW,
		KeyDown:  ebiten.KeyS,
		KeyLeft:  ebiten.KeyA,
		KeyRight: ebiten.KeyD,

		// Action keys
		KeyAction:  ebiten.KeySpace,
		KeyUseItem: ebiten.KeyE,

		// GAP-002 REPAIR: Spell casting keys (1-5)
		KeySpell1: ebiten.Key1,
		KeySpell2: ebiten.Key2,
		KeySpell3: ebiten.Key3,
		KeySpell4: ebiten.Key4,
		KeySpell5: ebiten.Key5,

		// UI keys
		KeyInventory: ebiten.KeyI,
		KeyCharacter: ebiten.KeyC,
		KeySkills:    ebiten.KeyK,
		KeyQuests:    ebiten.KeyJ,
		KeyMap:       ebiten.KeyM,

		// System keys
		KeyHelp:         ebiten.KeyEscape,
		KeyQuickSave:    ebiten.KeyF5,
		KeyQuickLoad:    ebiten.KeyF9,
		KeyCycleTargets: ebiten.KeyTab,

		// Mobile input
		touchHandler:  mobile.NewTouchInputHandler(),
		mobileEnabled: mobile.IsMobilePlatform(),
		useTouchInput: mobile.IsMobilePlatform(),
	}
}

// InitializeVirtualControls sets up virtual controls for mobile platforms.
// Should be called after screen size is known.
func (s *InputSystem) InitializeVirtualControls(screenWidth, screenHeight int) {
	if s.mobileEnabled {
		s.virtualControls = mobile.NewVirtualControlsLayout(screenWidth, screenHeight)
	}
}

// SetMobileEnabled manually enables or disables mobile input support.
func (s *InputSystem) SetMobileEnabled(enabled bool) {
	s.mobileEnabled = enabled
	s.useTouchInput = enabled
}

// IsMobileEnabled returns true if mobile input support is active.
func (s *InputSystem) IsMobileEnabled() bool {
	return s.mobileEnabled
}

// Update processes input for all entities with input components.
func (s *InputSystem) Update(entities []*Entity, deltaTime float64) {
	// BUG-023 fix: Validate mobile input initialization
	if s.mobileEnabled && s.virtualControls == nil {
		// Auto-initialize with default screen size if not explicitly initialized
		// This prevents silent input failure on mobile platforms
		s.InitializeVirtualControls(800, 600)
	}

	// BUG-010 fix: Track mouse position for delta calculation
	currentMouseX, currentMouseY := ebiten.CursorPosition()
	s.mouseDeltaX = currentMouseX - s.lastMouseX
	s.mouseDeltaY = currentMouseY - s.lastMouseY
	s.lastMouseX = currentMouseX
	s.lastMouseY = currentMouseY

	// Update mobile touch input
	if s.mobileEnabled && s.touchHandler != nil {
		s.touchHandler.Update()

		// Update virtual controls
		if s.virtualControls != nil {
			s.virtualControls.Update()

			// Handle menu button on virtual controls
			if s.virtualControls.IsMenuPressed() && s.onMenuToggle != nil {
				s.onMenuToggle()
			}
		}
	}

	// Handle global keys first (help menu, save/load, etc.)
	// ESC key handling - context-aware priority: tutorial > help > pause menu
	if inpututil.IsKeyJustPressed(s.KeyHelp) {
		// Priority 1: Check if tutorial is active and should handle the ESC key
		if s.tutorialSystem != nil && s.tutorialSystem.Enabled && s.tutorialSystem.ShowUI {
			// Skip current tutorial step
			s.tutorialSystem.Skip()
		} else if s.helpSystem != nil && s.helpSystem.Visible {
			// Priority 2: If help system is visible, close it
			s.helpSystem.Toggle()
		} else if s.onMenuToggle != nil {
			// Priority 3: Otherwise toggle pause menu
			s.onMenuToggle()
		}
	}

	// Handle quick save (F5)
	if inpututil.IsKeyJustPressed(s.KeyQuickSave) && s.onQuickSave != nil {
		if err := s.onQuickSave(); err != nil {
			// Show error notification
			if s.tutorialSystem != nil {
				s.tutorialSystem.ShowNotification("Save Failed: "+err.Error(), 3.0)
			}
		} else {
			// Show success notification
			if s.tutorialSystem != nil {
				s.tutorialSystem.ShowNotification("Game Saved!", 2.0)
			}
		}
	}

	// Handle quick load (F9)
	if inpututil.IsKeyJustPressed(s.KeyQuickLoad) && s.onQuickLoad != nil {
		if err := s.onQuickLoad(); err != nil {
			// Show error notification
			if s.tutorialSystem != nil {
				s.tutorialSystem.ShowNotification("Load Failed: "+err.Error(), 3.0)
			}
		} else {
			// Show success notification
			if s.tutorialSystem != nil {
				s.tutorialSystem.ShowNotification("Game Loaded!", 2.0)
			}
		}
	}

	// Handle UI shortcuts
	if inpututil.IsKeyJustPressed(s.KeyInventory) && s.onInventoryOpen != nil {
		s.onInventoryOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyCharacter) && s.onCharacterOpen != nil {
		s.onCharacterOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeySkills) && s.onSkillsOpen != nil {
		s.onSkillsOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyQuests) && s.onQuestsOpen != nil {
		s.onQuestsOpen()
	}
	if inpututil.IsKeyJustPressed(s.KeyMap) && s.onMapOpen != nil {
		s.onMapOpen()
	}

	// Handle target cycling
	if inpututil.IsKeyJustPressed(s.KeyCycleTargets) && s.onCycleTargets != nil {
		s.onCycleTargets()
	}

	// Handle help topic switching with number keys 1-6 (when help is visible)
	if s.helpSystem != nil && s.helpSystem.Visible {
		topicKeys := []ebiten.Key{
			ebiten.Key1, ebiten.Key2, ebiten.Key3,
			ebiten.Key4, ebiten.Key5, ebiten.Key6,
		}
		topicIDs := []string{
			"controls", "combat", "inventory",
			"progression", "world", "multiplayer",
		}

		for i, key := range topicKeys {
			if inpututil.IsKeyJustPressed(key) {
				s.helpSystem.ShowTopic(topicIDs[i])
				break
			}
		}
	}

	for _, entity := range entities {
		inputComp, ok := entity.GetComponent("input")
		if !ok {
			continue
		}

		input := inputComp.(*InputComponent)
		s.processInput(entity, input, deltaTime)
	}
}

// processInput handles input processing for a single entity.
func (s *InputSystem) processInput(entity *Entity, input *InputComponent, deltaTime float64) {
	// Reset input state
	input.MoveX = 0
	input.MoveY = 0
	input.ActionPressed = false
	input.UseItemPressed = false
	// GAP-002 REPAIR: Reset spell input flags
	input.Spell1Pressed = false
	input.Spell2Pressed = false
	input.Spell3Pressed = false
	input.Spell4Pressed = false
	input.Spell5Pressed = false

	// Auto-detect input method: if touch input is detected, switch to touch mode
	if s.mobileEnabled && len(ebiten.TouchIDs()) > 0 {
		s.useTouchInput = true
	} else if !s.mobileEnabled && len(ebiten.TouchIDs()) == 0 {
		// Allow falling back to keyboard if no touches (e.g., tablet with keyboard)
		s.useTouchInput = false
	}

	// Process touch input (priority on mobile)
	if s.useTouchInput && s.virtualControls != nil {
		// Get movement from virtual D-pad
		moveX, moveY := s.virtualControls.GetMovementInput()
		input.MoveX = moveX
		input.MoveY = moveY

		// Get action button presses
		if s.virtualControls.IsActionPressed() {
			input.ActionPressed = true
		}
		if s.virtualControls.IsSecondaryPressed() {
			input.UseItemPressed = true
		}

		// Use first touch outside controls as "mouse" position
		if s.touchHandler != nil {
			touches := s.touchHandler.GetActiveTouches()
			for _, touch := range touches {
				// Check if touch is outside virtual controls
				// (simple heuristic: use center-screen touches)
				screenW, _ := ebiten.WindowSize()
				if touch.X > 200 && touch.X < screenW-200 {
					input.MouseX = touch.X
					input.MouseY = touch.Y
					input.MousePressed = true
					break
				}
			}
		}
	} else {
		// Process keyboard movement (desktop mode)
		if ebiten.IsKeyPressed(s.KeyUp) {
			input.MoveY = -1.0
		}
		if ebiten.IsKeyPressed(s.KeyDown) {
			input.MoveY = 1.0
		}
		if ebiten.IsKeyPressed(s.KeyLeft) {
			input.MoveX = -1.0
		}
		if ebiten.IsKeyPressed(s.KeyRight) {
			input.MoveX = 1.0
		}

		// Normalize diagonal movement
		if input.MoveX != 0 && input.MoveY != 0 {
			// Divide by sqrt(2) to maintain constant speed in all directions
			input.MoveX *= 0.707
			input.MoveY *= 0.707
		}

		// Process action keys
		if inpututil.IsKeyJustPressed(s.KeyAction) {
			input.ActionPressed = true
		}
		if inpututil.IsKeyJustPressed(s.KeyUseItem) {
			input.UseItemPressed = true
		}

		// GAP-002 REPAIR: Process spell casting keys (1-5)
		if inpututil.IsKeyJustPressed(s.KeySpell1) {
			input.Spell1Pressed = true
		}
		if inpututil.IsKeyJustPressed(s.KeySpell2) {
			input.Spell2Pressed = true
		}
		if inpututil.IsKeyJustPressed(s.KeySpell3) {
			input.Spell3Pressed = true
		}
		if inpututil.IsKeyJustPressed(s.KeySpell4) {
			input.Spell4Pressed = true
		}
		if inpututil.IsKeyJustPressed(s.KeySpell5) {
			input.Spell5Pressed = true
		}

		// Process mouse input
		input.MouseX, input.MouseY = ebiten.CursorPosition()
		input.MousePressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	}

	// Apply movement to velocity component if it exists
	if velComp, ok := entity.GetComponent("velocity"); ok {
		velocity := velComp.(*VelocityComponent)
		velocity.VX = input.MoveX * s.MoveSpeed
		velocity.VY = input.MoveY * s.MoveSpeed
	}
}

// SetKeyBindings allows customizing key bindings.
func (s *InputSystem) SetKeyBindings(up, down, left, right, action, useItem ebiten.Key) {
	s.KeyUp = up
	s.KeyDown = down
	s.KeyLeft = left
	s.KeyRight = right
	s.KeyAction = action
	s.KeyUseItem = useItem
}

// SetHelpSystem connects the help system for ESC key toggling.
func (s *InputSystem) SetHelpSystem(helpSystem *HelpSystem) {
	s.helpSystem = helpSystem
}

// SetTutorialSystem connects the tutorial system for ESC key handling.
func (s *InputSystem) SetTutorialSystem(tutorialSystem *TutorialSystem) {
	s.tutorialSystem = tutorialSystem
}

// SetQuickSaveCallback sets the callback function for quick save (F5).
func (s *InputSystem) SetQuickSaveCallback(callback func() error) {
	s.onQuickSave = callback
}

// SetQuickLoadCallback sets the callback function for quick load (F9).
func (s *InputSystem) SetQuickLoadCallback(callback func() error) {
	s.onQuickLoad = callback
}

// SetInventoryCallback sets the callback function for opening inventory (I key).
func (s *InputSystem) SetInventoryCallback(callback func()) {
	s.onInventoryOpen = callback
}

// SetCharacterCallback sets the callback function for opening character screen (C key).
func (s *InputSystem) SetCharacterCallback(callback func()) {
	s.onCharacterOpen = callback
}

// SetSkillsCallback sets the callback function for opening skills screen (K key).
func (s *InputSystem) SetSkillsCallback(callback func()) {
	s.onSkillsOpen = callback
}

// SetQuestsCallback sets the callback function for opening quest log (J key).
func (s *InputSystem) SetQuestsCallback(callback func()) {
	s.onQuestsOpen = callback
}

// SetMapCallback sets the callback function for opening map (M key).
func (s *InputSystem) SetMapCallback(callback func()) {
	s.onMapOpen = callback
}

// SetCycleTargetsCallback sets the callback function for cycling targets (Tab key).
func (s *InputSystem) SetCycleTargetsCallback(callback func()) {
	s.onCycleTargets = callback
}

// SetMenuToggleCallback sets the callback function for toggling the pause menu (ESC key).
// This is called when ESC is pressed and neither tutorial nor help system consume the event.
func (s *InputSystem) SetMenuToggleCallback(callback func()) {
	s.onMenuToggle = callback
}

// SetMenuSystem connects the menu system for ESC key toggling.
// Deprecated: Use SetMenuToggleCallback instead for better decoupling.
func (s *InputSystem) SetMenuSystem(menuSystem *MenuSystem) {
	s.menuSystem = menuSystem
}

// ===== KEYBOARD INPUT METHODS =====

// IsKeyPressed returns true if the specified key is currently held down.
// Wraps Ebiten's continuous state check for consistency.
func (s *InputSystem) IsKeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

// IsKeyJustPressed returns true only on the frame when the key was first pressed.
// This is edge-triggered: returns true once per key press, not continuously while held.
// Explicitly exposed for use in game code.
func (s *InputSystem) IsKeyJustPressed(key ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(key)
}

// IsKeyReleased returns true only on the frame when the key was released.
// BUG-001 fix: Missing method for detecting key release events.
// Essential for charge attacks, aim mechanics, and jump height control.
func (s *InputSystem) IsKeyReleased(key ebiten.Key) bool {
	return inpututil.IsKeyJustReleased(key)
}

// IsKeyJustReleased is an alias for IsKeyReleased for API consistency.
// BUG-002 fix: Matches naming convention of IsKeyJustPressed.
func (s *InputSystem) IsKeyJustReleased(key ebiten.Key) bool {
	return s.IsKeyReleased(key)
}

// GetPressedKeys returns a slice of all keys currently pressed.
// BUG-003 fix: Needed for key binding UI and "press any key" prompts.
func (s *InputSystem) GetPressedKeys() []ebiten.Key {
	keys := make([]ebiten.Key, 0, 10) // Pre-allocate for common case
	return inpututil.AppendPressedKeys(keys)
}

// IsAnyKeyPressed returns true if any keyboard key is currently pressed.
// BUG-021 fix: Common pattern for "press any key to continue" scenarios.
func (s *InputSystem) IsAnyKeyPressed() bool {
	return len(inpututil.AppendPressedKeys(nil)) > 0
}

// GetAnyPressedKey returns the first pressed key found, or (0, false) if none.
// BUG-021 fix: Useful for key binding configuration UI.
func (s *InputSystem) GetAnyPressedKey() (ebiten.Key, bool) {
	keys := inpututil.AppendPressedKeys(nil)
	if len(keys) > 0 {
		return keys[0], true
	}
	return 0, false
}

// IsShiftPressed returns true if either left or right Shift key is pressed.
// BUG-022 fix: Convenience method for modifier keys.
func (s *InputSystem) IsShiftPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)
}

// IsControlPressed returns true if either left or right Control key is pressed.
// BUG-022 fix: Convenience method for modifier keys.
func (s *InputSystem) IsControlPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)
}

// IsAltPressed returns true if either left or right Alt key is pressed.
// BUG-022 fix: Convenience method for modifier keys.
func (s *InputSystem) IsAltPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight)
}

// IsSuperPressed returns true if either left or right Super/Meta key is pressed.
// BUG-022 fix: Super key is Windows/Command/Meta depending on OS.
func (s *InputSystem) IsSuperPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyMetaLeft) || ebiten.IsKeyPressed(ebiten.KeyMetaRight)
}

// ===== MOUSE INPUT METHODS =====

// IsMouseButtonPressed returns true if the mouse button is currently held down.
// This is continuous state - returns true every frame while held.
func (s *InputSystem) IsMouseButtonPressed(button ebiten.MouseButton) bool {
	return ebiten.IsMouseButtonPressed(button)
}

// IsMouseButtonJustPressed returns true only on the frame when the button was first pressed.
// BUG-004 fix: Edge-triggered click detection for single-action clicks.
// Critical for UI interactions that should fire once per click, not continuously.
func (s *InputSystem) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(button)
}

// IsMouseButtonReleased returns true only on the frame when the button was released.
// BUG-005 fix: Essential for drag-and-drop and context menus.
func (s *InputSystem) IsMouseButtonReleased(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustReleased(button)
}

// IsMouseButtonJustReleased is an alias for IsMouseButtonReleased for consistency.
// BUG-006 fix: Matches naming convention of IsMouseButtonJustPressed.
func (s *InputSystem) IsMouseButtonJustReleased(button ebiten.MouseButton) bool {
	return s.IsMouseButtonReleased(button)
}

// GetMousePosition returns the current mouse cursor position.
// Alias for GetCursorPosition for API clarity.
func (s *InputSystem) GetMousePosition() (x, y int) {
	return ebiten.CursorPosition()
}

// GetCursorPosition returns the current mouse cursor position.
func (s *InputSystem) GetCursorPosition() (x, y int) {
	return ebiten.CursorPosition()
}

// GetMouseDelta returns the mouse movement since the last frame.
// BUG-008 fix: Essential for first-person camera control and aiming.
// Returns the change in X and Y coordinates from the previous frame.
func (s *InputSystem) GetMouseDelta() (dx, dy int) {
	return s.mouseDeltaX, s.mouseDeltaY
}

// GetMouseWheel returns the mouse wheel scroll delta for the current frame.
// BUG-007 fix: Explicitly documented feature for camera zoom.
// Positive Y = scroll up, negative Y = scroll down.
// X is for horizontal scroll wheels (less common).
func (s *InputSystem) GetMouseWheel() (deltaX, deltaY float64) {
	return ebiten.Wheel()
}

// ===== KEY BINDING MANAGEMENT =====

// SetKeyBinding sets a specific key binding by action name.
// BUG-019 fix: Comprehensive key binding API supporting all 18 keys.
// Valid action names: "up", "down", "left", "right", "action", "useitem",
// "inventory", "character", "skills", "quests", "map",
// "help", "quicksave", "quickload", "cycletargets"
func (s *InputSystem) SetKeyBinding(action string, key ebiten.Key) bool {
	switch action {
	// Movement
	case "up":
		s.KeyUp = key
	case "down":
		s.KeyDown = key
	case "left":
		s.KeyLeft = key
	case "right":
		s.KeyRight = key
	// Actions
	case "action":
		s.KeyAction = key
	case "useitem":
		s.KeyUseItem = key
	// UI
	case "inventory":
		s.KeyInventory = key
	case "character":
		s.KeyCharacter = key
	case "skills":
		s.KeySkills = key
	case "quests":
		s.KeyQuests = key
	case "map":
		s.KeyMap = key
	// System
	case "help":
		s.KeyHelp = key
	case "quicksave":
		s.KeyQuickSave = key
	case "quickload":
		s.KeyQuickLoad = key
	case "cycletargets":
		s.KeyCycleTargets = key
	default:
		return false // Unknown action
	}
	return true
}

// GetKeyBinding returns the current key binding for the specified action.
// BUG-020 fix: Query API for displaying current bindings in settings UI.
// Returns (key, true) if action exists, (0, false) if unknown action.
func (s *InputSystem) GetKeyBinding(action string) (ebiten.Key, bool) {
	switch action {
	// Movement
	case "up":
		return s.KeyUp, true
	case "down":
		return s.KeyDown, true
	case "left":
		return s.KeyLeft, true
	case "right":
		return s.KeyRight, true
	// Actions
	case "action":
		return s.KeyAction, true
	case "useitem":
		return s.KeyUseItem, true
	// UI
	case "inventory":
		return s.KeyInventory, true
	case "character":
		return s.KeyCharacter, true
	case "skills":
		return s.KeySkills, true
	case "quests":
		return s.KeyQuests, true
	case "map":
		return s.KeyMap, true
	// System
	case "help":
		return s.KeyHelp, true
	case "quicksave":
		return s.KeyQuickSave, true
	case "quickload":
		return s.KeyQuickLoad, true
	case "cycletargets":
		return s.KeyCycleTargets, true
	default:
		return 0, false
	}
}

// GetAllKeyBindings returns a map of all current key bindings.
// BUG-020 fix: Comprehensive query for settings UI display.
func (s *InputSystem) GetAllKeyBindings() map[string]ebiten.Key {
	return map[string]ebiten.Key{
		// Movement
		"up":    s.KeyUp,
		"down":  s.KeyDown,
		"left":  s.KeyLeft,
		"right": s.KeyRight,
		// Actions
		"action":  s.KeyAction,
		"useitem": s.KeyUseItem,
		// UI
		"inventory": s.KeyInventory,
		"character": s.KeyCharacter,
		"skills":    s.KeySkills,
		"quests":    s.KeyQuests,
		"map":       s.KeyMap,
		// System
		"help":         s.KeyHelp,
		"quicksave":    s.KeyQuickSave,
		"quickload":    s.KeyQuickLoad,
		"cycletargets": s.KeyCycleTargets,
	}
}

// DrawVirtualControls renders virtual controls on screen (mobile only).
// Should be called during the game's Draw phase.
func (s *InputSystem) DrawVirtualControls(screen *ebiten.Image) {
	if s.mobileEnabled && s.virtualControls != nil {
		s.virtualControls.Draw(screen)
	}
}

// GetTouchHandler returns the touch input handler for advanced touch processing.
func (s *InputSystem) GetTouchHandler() *mobile.TouchInputHandler {
	return s.touchHandler
}

// SetVirtualControlsVisible controls visibility of virtual controls.
func (s *InputSystem) SetVirtualControlsVisible(visible bool) {
	if s.virtualControls != nil {
		s.virtualControls.SetVisible(visible)
	}
}
