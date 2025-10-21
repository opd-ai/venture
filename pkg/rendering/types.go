package rendering

import (
	"image/color"
)

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
