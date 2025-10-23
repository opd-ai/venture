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
