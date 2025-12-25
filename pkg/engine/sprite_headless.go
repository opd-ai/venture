//go:build headless
// +build headless

// Package engine provides headless sprite stubs for server builds.
// This file contains sprite implementations that don't require Ebiten/X11.
package engine

import (
	"image/color"
)

// HeadlessSprite is a lightweight sprite component for headless server builds.
// It tracks sprite metadata without actual image data, suitable for network
// servers that don't render graphics.
type HeadlessSprite struct {
	// Current facing direction for sprite selection
	CurrentDirection int

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

// Type returns the component type identifier (implements Component).
func (s *HeadlessSprite) Type() string {
	return "sprite"
}

// GetSize returns the sprite dimensions.
func (s *HeadlessSprite) GetSize() (width, height float64) {
	return s.Width, s.Height
}

// GetColor returns the sprite color tint.
func (s *HeadlessSprite) GetColor() color.Color {
	if s.Color == nil {
		return color.White
	}
	return s.Color
}

// GetRotation returns the sprite rotation in radians.
func (s *HeadlessSprite) GetRotation() float64 {
	return s.Rotation
}

// GetLayer returns the sprite rendering layer.
func (s *HeadlessSprite) GetLayer() int {
	return s.Layer
}

// IsVisible returns whether the sprite is visible.
func (s *HeadlessSprite) IsVisible() bool {
	return s.Visible
}

// SetVisible sets the sprite visibility.
func (s *HeadlessSprite) SetVisible(visible bool) {
	s.Visible = visible
}

// SetColor sets the sprite color tint.
func (s *HeadlessSprite) SetColor(col color.Color) {
	s.Color = col
}

// SetRotation sets the sprite rotation in radians.
func (s *HeadlessSprite) SetRotation(rotation float64) {
	s.Rotation = rotation
}

// NewSpriteComponent creates a new headless sprite component.
// This is a stub implementation for server builds that don't need actual rendering.
func NewSpriteComponent(width, height float64, color color.Color) *HeadlessSprite {
	return &HeadlessSprite{
		Width:            width,
		Height:           height,
		Color:            color,
		Visible:          true,
		Layer:            0,
		CurrentDirection: 1, // Default to DirDown (1)
	}
}
