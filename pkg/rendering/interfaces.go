//go:build !test
// +build !test

package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer is the base interface for all rendering components.
type Renderer interface {
	// Render draws to the provided image at the specified position
	Render(screen *ebiten.Image, x, y float64)
}

// Shape represents a geometric primitive that can be rendered.
type Shape interface {
	// Bounds returns the width and height of the shape
	Bounds() (width, height int)

	// Generate creates the image data for this shape
	Generate() *ebiten.Image
}

// PaletteGenerator generates color palettes based on genre and seed.
type PaletteGenerator interface {
	// Generate creates a palette for the given genre and seed
	Generate(genre string, seed int64) *Palette
}

// SpriteGenerator creates sprite images procedurally.
type SpriteGenerator interface {
	// Generate creates a sprite image from the configuration
	Generate(config SpriteConfig) (*ebiten.Image, error)
}
