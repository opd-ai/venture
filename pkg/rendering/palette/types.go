package palette

import (
	"image/color"
)

// Palette represents a cohesive color scheme for visual theming.
type Palette struct {
	// Primary color used for main elements
	Primary color.Color

	// Secondary color for accents
	Secondary color.Color

	// Background color
	Background color.Color

	// Text color for UI elements
	Text color.Color

	// Accent colors for variety
	Accent1 color.Color
	Accent2 color.Color

	// Danger and success colors for UI feedback
	Danger  color.Color
	Success color.Color

	// Additional theme colors for variation
	Colors []color.Color
}

// ColorScheme defines the base colors for a genre.
type ColorScheme struct {
	// Base hue range (0-360)
	BaseHue    float64
	Saturation float64 // 0.0-1.0
	Lightness  float64 // 0.0-1.0

	// Variation parameters
	HueVariation        float64
	SaturationVariation float64
	LightnessVariation  float64
}
