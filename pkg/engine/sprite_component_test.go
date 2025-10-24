package engine

import "image/color"

// StubSprite is a test sprite component without Ebiten dependencies.
// Implements SpriteProvider interface for testing.
type StubSprite struct {
	Width    float64
	Height   float64
	Color    color.Color
	Rotation float64
	Visible  bool
	Layer    int
}

// Type implements Component interface.
func (s *StubSprite) Type() string {
	return "sprite"
}

// GetImage implements SpriteProvider interface.
// Returns nil for stubs as we don't need actual images in tests.
func (s *StubSprite) GetImage() ImageProvider {
	return nil
}

// GetSize implements SpriteProvider interface.
func (s *StubSprite) GetSize() (width, height float64) {
	return s.Width, s.Height
}

// GetColor implements SpriteProvider interface.
func (s *StubSprite) GetColor() color.Color {
	if s.Color == nil {
		return color.White
	}
	return s.Color
}

// GetRotation implements SpriteProvider interface.
func (s *StubSprite) GetRotation() float64 {
	return s.Rotation
}

// GetLayer implements SpriteProvider interface.
func (s *StubSprite) GetLayer() int {
	return s.Layer
}

// IsVisible implements SpriteProvider interface.
func (s *StubSprite) IsVisible() bool {
	return s.Visible
}

// SetVisible implements SpriteProvider interface.
func (s *StubSprite) SetVisible(visible bool) {
	s.Visible = visible
}

// SetColor implements SpriteProvider interface.
func (s *StubSprite) SetColor(col color.Color) {
	s.Color = col
}

// SetRotation implements SpriteProvider interface.
func (s *StubSprite) SetRotation(rotation float64) {
	s.Rotation = rotation
}

// NewStubSprite creates a new test sprite component.
func NewStubSprite(width, height float64, col color.Color) *StubSprite {
	return &StubSprite{
		Width:   width,
		Height:  height,
		Color:   col,
		Visible: true,
		Layer:   0,
	}
}

// Compile-time interface check
var _ SpriteProvider = (*StubSprite)(nil)
