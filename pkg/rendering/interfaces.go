package rendering

import (
	"image/color"

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

// Palette represents a color scheme for consistent theming.
type Palette struct {
	// Primary color used for main elements
	Primary color.Color
	
	// Secondary color for accents
	Secondary color.Color
	
	// Background color
	Background color.Color
	
	// Text color for UI elements
	Text color.Color
	
	// Additional theme colors
	Colors []color.Color
}

// PaletteGenerator generates color palettes based on genre and seed.
type PaletteGenerator interface {
	// Generate creates a palette for the given genre and seed
	Generate(genre string, seed int64) *Palette
}

// SpriteConfig contains parameters for procedural sprite generation.
type SpriteConfig struct {
	// Width and height in pixels
	Width  int
	Height int
	
	// Seed for deterministic generation
	Seed int64
	
	// Palette to use for colors
	Palette *Palette
	
	// Type influences the generation algorithm
	Type string
	
	// Additional custom parameters
	Custom map[string]interface{}
}

// SpriteGenerator creates sprite images procedurally.
type SpriteGenerator interface {
	// Generate creates a sprite image from the configuration
	Generate(config SpriteConfig) (*ebiten.Image, error)
}
