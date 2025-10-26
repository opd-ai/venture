// Package palette provides procedural color palette generation.
// This file implements palette generators for genre-specific color schemes
// with harmonious color relationships.
package palette

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/genre"
	"github.com/sirupsen/logrus"
)

// Generator creates color palettes based on genre and seed.
type Generator struct {
	registry *genre.Registry
	seedGen  *procgen.SeedGenerator
	logger   *logrus.Entry
}

// NewGenerator creates a new palette generator.
func NewGenerator() *Generator {
	return NewGeneratorWithLogger(nil)
}

// NewGeneratorWithLogger creates a new palette generator with a logger.
func NewGeneratorWithLogger(logger *logrus.Logger) *Generator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator": "palette",
		})
	}
	return &Generator{
		registry: genre.DefaultRegistry(),
		seedGen:  procgen.NewSeedGenerator(0),
		logger:   logEntry,
	}
}

// Generate creates a palette for the given genre ID and seed.
func (g *Generator) Generate(genreID string, seed int64) (*Palette, error) {
	return g.GenerateWithOptions(genreID, seed, DefaultOptions())
}

// GenerateWithOptions creates a palette with specific generation options.
func (g *Generator) GenerateWithOptions(genreID string, seed int64, opts GenerationOptions) (*Palette, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"genreID": genreID,
			"seed":    seed,
		}).Debug("generating color palette")
	}

	genre, err := g.registry.Get(genreID)
	if err != nil {
		if g.logger != nil {
			g.logger.WithError(err).WithField("genreID", genreID).Error("genre not found")
		}
		return nil, err
	}

	g.seedGen = procgen.NewSeedGenerator(seed)
	paletteSeed := g.seedGen.GetSeed("palette", 0)
	rng := rand.New(rand.NewSource(paletteSeed))

	scheme := g.getSchemeForGenre(genre)
	palette := g.generateFromScheme(scheme, rng, opts)

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"genreID": genreID,
			"seed":    seed,
		}).Info("color palette generated")
	}

	return palette, nil
}

// getSchemeForGenre returns the appropriate color scheme for a genre.
func (g *Generator) getSchemeForGenre(genre *genre.Genre) ColorScheme {
	switch genre.ID {
	case "fantasy":
		return ColorScheme{
			BaseHue:             30, // Warm earthy tones
			Saturation:          0.6,
			Lightness:           0.5,
			HueVariation:        60,
			SaturationVariation: 0.2,
			LightnessVariation:  0.2,
		}
	case "scifi":
		return ColorScheme{
			BaseHue:             210, // Cool blues and cyans
			Saturation:          0.7,
			Lightness:           0.5,
			HueVariation:        40,
			SaturationVariation: 0.15,
			LightnessVariation:  0.25,
		}
	case "horror":
		return ColorScheme{
			BaseHue:             0, // Desaturated reds and grays
			Saturation:          0.3,
			Lightness:           0.3,
			HueVariation:        20,
			SaturationVariation: 0.1,
			LightnessVariation:  0.15,
		}
	case "cyberpunk":
		return ColorScheme{
			BaseHue:             300, // Neon purples and magentas
			Saturation:          0.9,
			Lightness:           0.5,
			HueVariation:        80,
			SaturationVariation: 0.1,
			LightnessVariation:  0.3,
		}
	case "postapoc":
		return ColorScheme{
			BaseHue:             45, // Dusty browns and oranges
			Saturation:          0.4,
			Lightness:           0.4,
			HueVariation:        30,
			SaturationVariation: 0.15,
			LightnessVariation:  0.2,
		}
	default:
		// Default to fantasy scheme
		return ColorScheme{
			BaseHue:             30,
			Saturation:          0.6,
			Lightness:           0.5,
			HueVariation:        60,
			SaturationVariation: 0.2,
			LightnessVariation:  0.2,
		}
	}
}

// generateFromScheme creates a palette from a color scheme and RNG.
func (g *Generator) generateFromScheme(scheme ColorScheme, rng *rand.Rand, opts GenerationOptions) *Palette {
	// Adjust scheme based on mood
	scheme = g.applyMood(scheme, opts.Mood)

	// Adjust scheme based on rarity
	scheme = g.applyRarity(scheme, opts.Rarity)

	palette := &Palette{
		Colors: make([]color.Color, max(opts.MinColors, 12)),
	}

	// Generate colors based on harmony type
	baseHue := scheme.BaseHue
	harmonyHues := g.getHarmonyHues(baseHue, opts.Harmony)

	// GAP-019 REPAIR: Add hue variation to primary color for entity diversity
	// Use HueVariation to create different colored entities within the genre theme
	primaryHueOffset := (rng.Float64()*2 - 1) * scheme.HueVariation // -HueVariation to +HueVariation
	primaryHue := math.Mod(baseHue+primaryHueOffset, 360)
	if primaryHue < 0 {
		primaryHue += 360
	}

	// Generate primary color with variation
	palette.Primary = hslToColor(
		primaryHue,
		clamp(scheme.Saturation+rng.Float64()*scheme.SaturationVariation-scheme.SaturationVariation/2, 0, 1),
		clamp(scheme.Lightness+rng.Float64()*scheme.LightnessVariation-scheme.LightnessVariation/2, 0, 1),
	)

	// Generate secondary color based on harmony
	secondaryHue := harmonyHues[min(1, len(harmonyHues)-1)]
	palette.Secondary = hslToColor(
		secondaryHue,
		clamp(scheme.Saturation+rng.Float64()*scheme.SaturationVariation-scheme.SaturationVariation/2, 0, 1),
		clamp(scheme.Lightness+rng.Float64()*scheme.LightnessVariation-scheme.LightnessVariation/2, 0, 1),
	)

	// Generate background (darker version of base)
	palette.Background = hslToColor(
		baseHue,
		scheme.Saturation*0.3,
		scheme.Lightness*0.2,
	)

	// Generate text color (high contrast with background)
	if scheme.Lightness < 0.5 {
		palette.Text = color.RGBA{R: 240, G: 240, B: 240, A: 255}
	} else {
		palette.Text = color.RGBA{R: 20, G: 20, B: 20, A: 255}
	}

	// Generate accent colors from harmony hues
	palette.Accent1 = hslToColor(
		harmonyHues[min(1, len(harmonyHues)-1)],
		scheme.Saturation*0.8,
		scheme.Lightness*1.1,
	)

	palette.Accent2 = hslToColor(
		harmonyHues[min(2, len(harmonyHues)-1)],
		scheme.Saturation*0.9,
		scheme.Lightness*0.9,
	)

	palette.Accent3 = hslToColor(
		harmonyHues[min(3, len(harmonyHues)-1)],
		scheme.Saturation*0.85,
		scheme.Lightness*1.05,
	)

	// Generate highlight colors (brighter versions)
	palette.Highlight1 = hslToColor(
		baseHue,
		scheme.Saturation*0.6,
		clamp(scheme.Lightness*1.4, 0, 0.9),
	)

	palette.Highlight2 = hslToColor(
		harmonyHues[min(1, len(harmonyHues)-1)],
		scheme.Saturation*0.5,
		clamp(scheme.Lightness*1.5, 0, 0.95),
	)

	// Generate shadow colors (darker versions)
	palette.Shadow1 = hslToColor(
		baseHue,
		scheme.Saturation*0.4,
		scheme.Lightness*0.3,
	)

	palette.Shadow2 = hslToColor(
		harmonyHues[min(1, len(harmonyHues)-1)],
		scheme.Saturation*0.3,
		scheme.Lightness*0.25,
	)

	// Generate neutral color (desaturated)
	palette.Neutral = hslToColor(
		baseHue,
		scheme.Saturation*0.15,
		scheme.Lightness*0.6,
	)

	// Generate UI feedback colors
	palette.Danger = hslToColor(0, 0.8, 0.5)
	palette.Success = hslToColor(120, 0.7, 0.5)
	palette.Warning = hslToColor(45, 0.9, 0.55)
	palette.Info = hslToColor(200, 0.75, 0.5)

	// Generate additional colors for variety using harmony hues
	minColors := max(opts.MinColors, 12)
	for i := 0; i < minColors; i++ {
		hueIdx := i % len(harmonyHues)
		hue := harmonyHues[hueIdx]

		// Add slight variation to each color
		hueVariation := rng.Float64()*scheme.HueVariation - scheme.HueVariation/2
		hue = math.Mod(hue+hueVariation, 360)

		sat := clamp(scheme.Saturation+rng.Float64()*scheme.SaturationVariation-scheme.SaturationVariation/2, 0, 1)
		light := clamp(scheme.Lightness+rng.Float64()*scheme.LightnessVariation-scheme.LightnessVariation/2, 0.2, 0.8)
		palette.Colors[i] = hslToColor(hue, sat, light)
	}

	return palette
}

// hslToColor converts HSL color space to RGB color.
// h: 0-360, s: 0-1, l: 0-1
func hslToColor(h, s, l float64) color.Color {
	h = math.Mod(h, 360) / 360

	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return color.RGBA{
		R: uint8(clamp(r*255, 0, 255)),
		G: uint8(clamp(g*255, 0, 255)),
		B: uint8(clamp(b*255, 0, 255)),
		A: 255,
	}
}

// hueToRGB is a helper function for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// clamp restricts a value to a given range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getHarmonyHues returns hues based on harmony type.
func (g *Generator) getHarmonyHues(baseHue float64, harmony HarmonyType) []float64 {
	switch harmony {
	case HarmonyComplementary:
		// Base + opposite (180°)
		return []float64{
			baseHue,
			math.Mod(baseHue+180, 360),
		}
	case HarmonyAnalogous:
		// Base + adjacent hues (±30°)
		return []float64{
			baseHue,
			math.Mod(baseHue+30, 360),
			math.Mod(baseHue-30+360, 360),
		}
	case HarmonyTriadic:
		// Three evenly spaced hues (120° apart)
		return []float64{
			baseHue,
			math.Mod(baseHue+120, 360),
			math.Mod(baseHue+240, 360),
		}
	case HarmonyTetradic:
		// Four evenly spaced hues (90° apart)
		return []float64{
			baseHue,
			math.Mod(baseHue+90, 360),
			math.Mod(baseHue+180, 360),
			math.Mod(baseHue+270, 360),
		}
	case HarmonySplitComplementary:
		// Base + two adjacent to complement (150° and 210°)
		return []float64{
			baseHue,
			math.Mod(baseHue+150, 360),
			math.Mod(baseHue+210, 360),
		}
	case HarmonyMonochromatic:
		// Single hue with variations in saturation/lightness
		return []float64{baseHue}
	default:
		return []float64{baseHue}
	}
}

// applyMood adjusts color scheme based on mood type.
func (g *Generator) applyMood(scheme ColorScheme, mood MoodType) ColorScheme {
	adjusted := scheme

	switch mood {
	case MoodBright:
		// Increase lightness for cheerful tone
		adjusted.Lightness = clamp(adjusted.Lightness*1.3, 0, 0.9)
		adjusted.Saturation = clamp(adjusted.Saturation*1.1, 0, 1)
	case MoodDark:
		// Decrease lightness for somber tone
		adjusted.Lightness = clamp(adjusted.Lightness*0.6, 0.1, 1)
		adjusted.Saturation = clamp(adjusted.Saturation*0.8, 0, 1)
	case MoodSaturated:
		// Increase saturation for intense colors
		adjusted.Saturation = clamp(adjusted.Saturation*1.4, 0, 1)
	case MoodMuted:
		// Decrease saturation for subdued colors
		adjusted.Saturation = clamp(adjusted.Saturation*0.5, 0, 1)
	case MoodVibrant:
		// Maximize saturation and optimize lightness
		adjusted.Saturation = clamp(adjusted.Saturation*1.5, 0, 1)
		adjusted.Lightness = clamp(adjusted.Lightness*1.2, 0, 0.7)
	case MoodPastel:
		// High lightness with low saturation
		adjusted.Lightness = clamp(adjusted.Lightness*1.5, 0.7, 0.95)
		adjusted.Saturation = clamp(adjusted.Saturation*0.4, 0, 0.5)
	case MoodNormal:
		// No adjustments
	}

	return adjusted
}

// applyRarity adjusts color scheme based on rarity tier.
func (g *Generator) applyRarity(scheme ColorScheme, rarity Rarity) ColorScheme {
	adjusted := scheme

	switch rarity {
	case RarityCommon:
		// Muted, standard colors
		adjusted.Saturation = clamp(adjusted.Saturation*0.8, 0, 1)
	case RarityUncommon:
		// Slightly enhanced colors
		adjusted.Saturation = clamp(adjusted.Saturation*1.0, 0, 1)
		adjusted.Lightness = clamp(adjusted.Lightness*1.05, 0, 1)
	case RarityRare:
		// Vibrant, saturated colors
		adjusted.Saturation = clamp(adjusted.Saturation*1.3, 0, 1)
		adjusted.Lightness = clamp(adjusted.Lightness*1.1, 0, 1)
	case RarityEpic:
		// Intense colors with higher contrast
		adjusted.Saturation = clamp(adjusted.Saturation*1.5, 0, 1)
		adjusted.Lightness = clamp(adjusted.Lightness*1.15, 0, 0.85)
		adjusted.LightnessVariation *= 1.3
	case RarityLegendary:
		// Extraordinary colors with maximum impact
		adjusted.Saturation = clamp(adjusted.Saturation*1.7, 0, 1)
		adjusted.Lightness = clamp(adjusted.Lightness*1.2, 0, 0.8)
		adjusted.HueVariation *= 1.5
		adjusted.SaturationVariation *= 1.4
		adjusted.LightnessVariation *= 1.5
	}

	return adjusted
}
