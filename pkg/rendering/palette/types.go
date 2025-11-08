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
// Phase 19.2: Expanded to 20+ moods for variety
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
	// MoodTense creates anxiety with desaturated, dark colors
	MoodTense
	// MoodCalm creates peace with soft, balanced colors
	MoodCalm
	// MoodVictorious creates triumph with bright, saturated golds
	MoodVictorious
	// MoodMelancholic creates sadness with desaturated blues
	MoodMelancholic
	// MoodEnergetic creates excitement with bright, warm colors
	MoodEnergetic
	// MoodMystical creates wonder with purples and deep blues
	MoodMystical
	// MoodOminous creates dread with dark reds and blacks
	MoodOminous
	// MoodSerene creates tranquility with soft blues and greens
	MoodSerene
	// MoodAggressive creates intensity with high saturation reds
	MoodAggressive
	// MoodPlayful creates fun with varied bright colors
	MoodPlayful
	// MoodSomber creates gravity with dark, desaturated colors
	MoodSomber
	// MoodEthereal creates otherworldly feel with high lightness pastels
	MoodEthereal
	// MoodDangerous creates threat with deep reds and oranges
	MoodDangerous
	// MoodPeaceful creates harmony with balanced greens
	MoodPeaceful
	// MoodChaotic creates disorder with high variation
	MoodChaotic
	// MoodRegal creates majesty with purples and golds
	MoodRegal
	// MoodDesolate creates emptiness with grays and browns
	MoodDesolate
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
	case MoodTense:
		return "Tense"
	case MoodCalm:
		return "Calm"
	case MoodVictorious:
		return "Victorious"
	case MoodMelancholic:
		return "Melancholic"
	case MoodEnergetic:
		return "Energetic"
	case MoodMystical:
		return "Mystical"
	case MoodOminous:
		return "Ominous"
	case MoodSerene:
		return "Serene"
	case MoodAggressive:
		return "Aggressive"
	case MoodPlayful:
		return "Playful"
	case MoodSomber:
		return "Somber"
	case MoodEthereal:
		return "Ethereal"
	case MoodDangerous:
		return "Dangerous"
	case MoodPeaceful:
		return "Peaceful"
	case MoodChaotic:
		return "Chaotic"
	case MoodRegal:
		return "Regal"
	case MoodDesolate:
		return "Desolate"
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

// TimeOfDay defines time periods that affect palette color adjustments.
// Phase 17.3: Time-of-Day Color Shifts
type TimeOfDay int

const (
	// TimeOfDayDawn represents sunrise with warm, soft tones
	TimeOfDayDawn TimeOfDay = iota
	// TimeOfDayDay represents daytime with bright, saturated colors
	TimeOfDayDay
	// TimeOfDayDusk represents sunset with warm, rich tones
	TimeOfDayDusk
	// TimeOfDayNight represents nighttime with cool, desaturated colors
	TimeOfDayNight
)

// String returns the string representation of TimeOfDay.
func (t TimeOfDay) String() string {
	switch t {
	case TimeOfDayDawn:
		return "Dawn"
	case TimeOfDayDay:
		return "Day"
	case TimeOfDayDusk:
		return "Dusk"
	case TimeOfDayNight:
		return "Night"
	default:
		return "Unknown"
	}
}

// TimeConfig defines time-based color modulation parameters.
// Phase 17.3: Time-of-Day Color Shifts
type TimeConfig struct {
	// CurrentTime is the current time of day
	CurrentTime TimeOfDay
	// TransitionProgress is interpolation factor (0.0-1.0) between current and next time state
	// Used for smooth 5-second transitions
	TransitionProgress float64
	// IntensityMultiplier adjusts the strength of time-based effects (0.0-1.0, default 1.0)
	IntensityMultiplier float64
}

// DefaultTimeConfig returns a default time configuration (Day, no transition).
func DefaultTimeConfig() TimeConfig {
	return TimeConfig{
		CurrentTime:         TimeOfDayDay,
		TransitionProgress:  0.0,
		IntensityMultiplier: 1.0,
	}
}

// ColorModulation defines color adjustments for a specific time of day.
// Phase 17.3: Time-of-Day Color Shifts
type ColorModulation struct {
	// HueShift adjusts hue (in degrees, -180 to +180)
	HueShift float64
	// SaturationMultiplier scales saturation (0.0-2.0, 1.0 = no change)
	SaturationMultiplier float64
	// LightnessOffset adds to lightness (-0.3 to +0.3)
	LightnessOffset float64
	// TemperatureShift adjusts color temperature (negative = cooler, positive = warmer)
	// -1.0 = very cool (blue), +1.0 = very warm (orange)
	TemperatureShift float64
}
