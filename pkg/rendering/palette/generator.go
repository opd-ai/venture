package palette

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/genre"
)

// Generator creates color palettes based on genre and seed.
type Generator struct {
	registry *genre.Registry
	seedGen  *procgen.SeedGenerator
}

// NewGenerator creates a new palette generator.
func NewGenerator() *Generator {
	return &Generator{
		registry: genre.DefaultRegistry(),
		seedGen:  procgen.NewSeedGenerator(0),
	}
}

// Generate creates a palette for the given genre ID and seed.
func (g *Generator) Generate(genreID string, seed int64) (*Palette, error) {
	genre, err := g.registry.Get(genreID)
	if err != nil {
		return nil, err
	}

	g.seedGen = procgen.NewSeedGenerator(seed)
	paletteSeed := g.seedGen.GetSeed("palette", 0)
	rng := rand.New(rand.NewSource(paletteSeed))

	scheme := g.getSchemeForGenre(genre)
	return g.generateFromScheme(scheme, rng), nil
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
func (g *Generator) generateFromScheme(scheme ColorScheme, rng *rand.Rand) *Palette {
	palette := &Palette{
		Colors: make([]color.Color, 8),
	}

	// Generate primary color
	palette.Primary = hslToColor(
		scheme.BaseHue,
		scheme.Saturation,
		scheme.Lightness,
	)

	// Generate secondary color (complementary hue)
	palette.Secondary = hslToColor(
		math.Mod(scheme.BaseHue+180+rng.Float64()*scheme.HueVariation-scheme.HueVariation/2, 360),
		clamp(scheme.Saturation+rng.Float64()*scheme.SaturationVariation-scheme.SaturationVariation/2, 0, 1),
		clamp(scheme.Lightness+rng.Float64()*scheme.LightnessVariation-scheme.LightnessVariation/2, 0, 1),
	)

	// Generate background (darker version of base)
	palette.Background = hslToColor(
		scheme.BaseHue,
		scheme.Saturation*0.3,
		scheme.Lightness*0.2,
	)

	// Generate text color (high contrast with background)
	if scheme.Lightness < 0.5 {
		palette.Text = color.RGBA{R: 240, G: 240, B: 240, A: 255}
	} else {
		palette.Text = color.RGBA{R: 20, G: 20, B: 20, A: 255}
	}

	// Generate accent colors
	palette.Accent1 = hslToColor(
		math.Mod(scheme.BaseHue+120, 360),
		scheme.Saturation*0.8,
		scheme.Lightness*1.1,
	)

	palette.Accent2 = hslToColor(
		math.Mod(scheme.BaseHue+240, 360),
		scheme.Saturation*0.9,
		scheme.Lightness*0.9,
	)

	// Generate danger color (red tint)
	palette.Danger = hslToColor(0, 0.8, 0.5)

	// Generate success color (green tint)
	palette.Success = hslToColor(120, 0.7, 0.5)

	// Generate additional colors for variety
	for i := 0; i < 8; i++ {
		hue := math.Mod(scheme.BaseHue+float64(i)*45+rng.Float64()*scheme.HueVariation, 360)
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
