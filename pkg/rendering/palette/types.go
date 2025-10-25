// Package palette provides color palette type definitions.
// This file defines palette data structures and color relationships
// used by the palette generator.
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
	Accent3 color.Color

	// Highlight colors for emphasis
	Highlight1 color.Color
	Highlight2 color.Color

	// Shadow colors for depth
	Shadow1 color.Color
	Shadow2 color.Color

	// Neutral color for UI elements
	Neutral color.Color

	// Danger and success colors for UI feedback
	Danger  color.Color
	Success color.Color
	Warning color.Color
	Info    color.Color

	// Additional theme colors for variation (minimum 12)
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

// HarmonyType defines color harmony relationships.
type HarmonyType int

const (
	// HarmonyComplementary uses opposite hues (180° apart)
	HarmonyComplementary HarmonyType = iota
	// HarmonyAnalogous uses adjacent hues (30° apart)
	HarmonyAnalogous
	// HarmonyTriadic uses three evenly spaced hues (120° apart)
	HarmonyTriadic
	// HarmonyTetradic uses four evenly spaced hues (90° apart)
	HarmonyTetradic
	// HarmonySplitComplementary uses base hue plus two adjacent to complement
	HarmonySplitComplementary
	// HarmonyMonochromatic uses single hue with varying saturation/lightness
	HarmonyMonochromatic
)

// String returns the string representation of HarmonyType.
func (h HarmonyType) String() string {
	switch h {
	case HarmonyComplementary:
		return "Complementary"
	case HarmonyAnalogous:
		return "Analogous"
	case HarmonyTriadic:
		return "Triadic"
	case HarmonyTetradic:
		return "Tetradic"
	case HarmonySplitComplementary:
		return "SplitComplementary"
	case HarmonyMonochromatic:
		return "Monochromatic"
	default:
		return "Unknown"
	}
}

// MoodType defines emotional color adjustments.
type MoodType int

const (
	// MoodNormal uses standard color values
	MoodNormal MoodType = iota
	// MoodBright increases lightness for cheerful tone
	MoodBright
	// MoodDark decreases lightness for somber tone
	MoodDark
	// MoodSaturated increases saturation for intense colors
	MoodSaturated
	// MoodMuted decreases saturation for subdued colors
	MoodMuted
	// MoodVibrant maximizes saturation and lightness
	MoodVibrant
	// MoodPastel uses high lightness with low saturation
	MoodPastel
)

// String returns the string representation of MoodType.
func (m MoodType) String() string {
	switch m {
	case MoodNormal:
		return "Normal"
	case MoodBright:
		return "Bright"
	case MoodDark:
		return "Dark"
	case MoodSaturated:
		return "Saturated"
	case MoodMuted:
		return "Muted"
	case MoodVibrant:
		return "Vibrant"
	case MoodPastel:
		return "Pastel"
	default:
		return "Unknown"
	}
}

// Rarity defines item rarity tiers affecting color intensity.
type Rarity int

const (
	// RarityCommon uses muted, standard colors
	RarityCommon Rarity = iota
	// RarityUncommon uses slightly enhanced colors
	RarityUncommon
	// RarityRare uses vibrant, saturated colors
	RarityRare
	// RarityEpic uses intense colors with metallic hints
	RarityEpic
	// RarityLegendary uses extraordinary colors with special effects
	RarityLegendary
)

// String returns the string representation of Rarity.
func (r Rarity) String() string {
	switch r {
	case RarityCommon:
		return "Common"
	case RarityUncommon:
		return "Uncommon"
	case RarityRare:
		return "Rare"
	case RarityEpic:
		return "Epic"
	case RarityLegendary:
		return "Legendary"
	default:
		return "Unknown"
	}
}

// GenerationOptions configures palette generation.
type GenerationOptions struct {
	// Harmony type for color relationships
	Harmony HarmonyType
	// Mood for emotional tone
	Mood MoodType
	// Rarity tier for color intensity
	Rarity Rarity
	// MinColors minimum number of colors to generate (default: 12)
	MinColors int
}

// DefaultOptions returns default generation options.
func DefaultOptions() GenerationOptions {
	return GenerationOptions{
		Harmony:   HarmonyComplementary,
		Mood:      MoodNormal,
		Rarity:    RarityCommon,
		MinColors: 12,
	}
}
