//go:build !test
// +build !test

// Package engine provides player input handling.
// This file implements InputSystem which processes keyboard and mouse input
// for player-controlled entities and game controls.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

// InputSystem processes keyboard and mouse input and updates input components.
type InputSystem struct {
	// Movement speed multiplier
	MoveSpeed float64

	// Key bindings
	KeyUp        ebiten.Key
	KeyDown      ebiten.Key
	KeyLeft      ebiten.Key
	KeyRight     ebiten.Key
	KeyAction    ebiten.Key
	KeyUseItem   ebiten.Key
	KeyHelp      ebiten.Key // ESC key for help menu
	KeyQuickSave ebiten.Key // F5 key for quick save
	KeyQuickLoad ebiten.Key // F9 key for quick load

	// References to game systems for special key handling
	helpSystem     *HelpSystem
	tutorialSystem *TutorialSystem

	// Callbacks for save/load operations
	onQuickSave func() error
	onQuickLoad func() error
}

// NewInputSystem creates a new input system with default key bindings.
func NewInputSystem() *InputSystem {
	return &InputSystem{
		MoveSpeed:    100.0, // pixels per second
		KeyUp:        ebiten.KeyW,
		KeyDown:      ebiten.KeyS,
		KeyLeft:      ebiten.KeyA,
		KeyRight:     ebiten.KeyD,
		KeyAction:    ebiten.KeySpace,
		KeyUseItem:   ebiten.KeyE,
		KeyHelp:      ebiten.KeyEscape,
		KeyQuickSave: ebiten.KeyF5,
		KeyQuickLoad: ebiten.KeyF9,
	}
}

// Update processes input for all entities with input components.
func (s *InputSystem) Update(entities []*Entity, deltaTime float64) {
	// Handle global keys first (help menu, save/load, etc.)
	// ESC key handling - context-aware: tutorial takes priority over help menu
	if inpututil.IsKeyJustPressed(s.KeyHelp) {
		// Check if tutorial is active and should handle the ESC key
		if s.tutorialSystem != nil && s.tutorialSystem.Enabled && s.tutorialSystem.ShowUI {
			// Skip current tutorial step
			s.tutorialSystem.Skip()
		} else if s.helpSystem != nil {
			// Otherwise toggle help menu
			s.helpSystem.Toggle()
		}
	}

	// Handle quick save (F5)
	if inpututil.IsKeyJustPressed(s.KeyQuickSave) && s.onQuickSave != nil {
		if err := s.onQuickSave(); err != nil {
			// Error is logged by the callback
		}
	}

	// Handle quick load (F9)
	if inpututil.IsKeyJustPressed(s.KeyQuickLoad) && s.onQuickLoad != nil {
		if err := s.onQuickLoad(); err != nil {
			// Error is logged by the callback
		}
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

	// Process keyboard movement
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
