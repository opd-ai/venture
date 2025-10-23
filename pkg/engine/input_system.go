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
