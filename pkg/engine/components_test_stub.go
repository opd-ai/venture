//go:build test
// +build test

// Package engine provides test stubs for components that depend on Ebiten.
// This file provides unified stub implementations when building with the test tag,
// allowing unit tests to compile without Ebiten/X11 dependencies.
package engine

import "image/color"

// InputComponent stores the current input state for an entity (test stub).
type InputComponent struct {
	// Movement input (-1.0 to 1.0 for each axis)
	MoveX, MoveY float64

	// Action buttons
	ActionPressed   bool
	SecondaryAction bool
	UseItemPressed  bool
	Action          bool // Alias for ActionPressed (backward compatibility)

	// GAP-001/GAP-002 REPAIR: Frame-persistent detection flags for tutorial/UI
	ActionJustPressed  bool // Set when action key first pressed this frame
	UseItemJustPressed bool // Set when use item key first pressed this frame
	AnyKeyPressed      bool // GAP-005 REPAIR: Set when any key pressed this frame

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

// SpriteComponent holds visual representation data for an entity (test stub).
type SpriteComponent struct {
	// Color tint
	Color color.Color

	// Size (width, height)
	Width, Height float64

	// Rotation in radians
	Rotation float64

	// Visibility flag
	Visible bool

	// Layer for rendering order (higher = drawn on top)
	Layer int
}

// Type returns the component type identifier.
func (s *SpriteComponent) Type() string {
	return "sprite"
}

// NewSpriteComponent creates a new sprite component (test stub).
func NewSpriteComponent(width, height float64, col color.Color) *SpriteComponent {
	return &SpriteComponent{
		Width:   width,
		Height:  height,
		Color:   col,
		Visible: true,
		Layer:   0,
	}
}
