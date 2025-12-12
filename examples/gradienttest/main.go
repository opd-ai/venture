// Package main provides a CLI tool for testing gradient generation.
// Phase 19.2: Dynamic Palette System - Gradient Test Tool
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

var (
	width           = flag.Int("width", 800, "Width of gradient image")
	height          = flag.Int("height", 600, "Height of gradient image")
	gradientType    = flag.String("type", "linear", "Gradient type: linear, radial, angular, diamond, spiral, conic")
	angle           = flag.Float64("angle", 0, "Angle for linear/angular/conic gradients (degrees)")
	centerX         = flag.Float64("centerx", 0.5, "Center X position (0.0-1.0)")
	centerY         = flag.Float64("centery", 0.5, "Center Y position (0.0-1.0)")
	radius          = flag.Float64("radius", 0.5, "Radius for radial gradients (0.0-1.0)")
	spiralRotations = flag.Float64("rotations", 2.0, "Number of rotations for spiral gradients")
	reverse         = flag.Bool("reverse", false, "Reverse gradient direction")
	smoothness      = flag.Float64("smoothness", 0.5, "Smoothness factor (0.0-1.0)")
	output          = flag.String("output", "gradient.png", "Output PNG file path")
	verbose         = flag.Bool("verbose", false, "Enable verbose logging")
	genre           = flag.String("genre", "fantasy", "Genre for palette colors: fantasy, scifi, horror, cyberpunk, postapoc")
	mood            = flag.String("mood", "normal", "Mood for palette: normal, bright, dark, calm, tense, victorious, etc.")
	colorSteps      = flag.Int("steps", 5, "Number of color steps in gradient")
)

func main() {
	flag.Parse()

	// Setup logging
	logger := logrus.New()
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.WithFields(logrus.Fields{
		"type":   *gradientType,
		"width":  *width,
		"height": *height,
		"genre":  *genre,
		"mood":   *mood,
		"output": *output,
	}).Info("Starting gradient generation")

	// Parse gradient type
	var gType palette.GradientType
	switch *gradientType {
	case "linear":
		gType = palette.GradientLinear
	case "radial":
		gType = palette.GradientRadial
	case "angular":
		gType = palette.GradientAngular
	case "diamond":
		gType = palette.GradientDiamond
	case "spiral":
		gType = palette.GradientSpiral
	case "conic":
		gType = palette.GradientConic
	default:
		logger.Fatalf("Unknown gradient type: %s", *gradientType)
	}

	// Generate palette colors based on genre and mood
	colors := generatePaletteColors(*genre, *mood, *colorSteps)

	// Create gradient configuration
	config := palette.GradientConfig{
		Type:            gType,
		Colors:          colors,
		Angle:           *angle,
		CenterX:         *centerX,
		CenterY:         *centerY,
		Radius:          *radius,
		SpiralRotations: *spiralRotations,
		Reverse:         *reverse,
		Smoothness:      *smoothness,
	}

	logger.WithFields(logrus.Fields{
		"colors": len(config.Colors),
		"angle":  config.Angle,
	}).Debug("Gradient configuration created")

	// Generate gradient
	img := palette.GenerateGradient(*width, *height, config)

	logger.WithFields(logrus.Fields{
		"width":  img.Bounds().Dx(),
		"height": img.Bounds().Dy(),
	}).Info("Gradient generated")

	// Save to PNG
	if err := savePNG(img, *output); err != nil {
		logger.Fatalf("Failed to save PNG: %v", err)
	}

	logger.WithField("file", *output).Info("Gradient saved successfully")

	// Print statistics
	printStatistics(img, logger)
}

// generatePaletteColors creates a color array based on genre and mood.
func generatePaletteColors(genreID, moodStr string, steps int) []color.Color {
	gen := palette.NewGenerator()

	// Parse mood
	var moodType palette.MoodType
	switch moodStr {
	case "normal":
		moodType = palette.MoodNormal
	case "bright":
		moodType = palette.MoodBright
	case "dark":
		moodType = palette.MoodDark
	case "saturated":
		moodType = palette.MoodSaturated
	case "muted":
		moodType = palette.MoodMuted
	case "vibrant":
		moodType = palette.MoodVibrant
	case "pastel":
		moodType = palette.MoodPastel
	case "tense":
		moodType = palette.MoodTense
	case "calm":
		moodType = palette.MoodCalm
	case "victorious":
		moodType = palette.MoodVictorious
	case "melancholic":
		moodType = palette.MoodMelancholic
	case "energetic":
		moodType = palette.MoodEnergetic
	case "mystical":
		moodType = palette.MoodMystical
	case "ominous":
		moodType = palette.MoodOminous
	case "serene":
		moodType = palette.MoodSerene
	case "aggressive":
		moodType = palette.MoodAggressive
	case "playful":
		moodType = palette.MoodPlayful
	case "somber":
		moodType = palette.MoodSomber
	case "ethereal":
		moodType = palette.MoodEthereal
	case "dangerous":
		moodType = palette.MoodDangerous
	case "peaceful":
		moodType = palette.MoodPeaceful
	case "chaotic":
		moodType = palette.MoodChaotic
	case "regal":
		moodType = palette.MoodRegal
	case "desolate":
		moodType = palette.MoodDesolate
	default:
		moodType = palette.MoodNormal
	}

	// Generate palette
	opts := palette.DefaultOptions()
	opts.Mood = moodType
	opts.MinColors = steps

	pal, err := gen.GenerateWithOptions(genreID, 12345, opts)
	if err != nil {
		logrus.Fatalf("Failed to generate palette: %v", err)
	}

	// Extract colors from palette
	colors := make([]color.Color, 0, steps)
	colors = append(colors, pal.Primary)
	colors = append(colors, pal.Secondary)

	// Add additional colors from palette.Colors
	for i := 0; i < steps-2 && i < len(pal.Colors); i++ {
		colors = append(colors, pal.Colors[i])
	}

	return colors
}

// savePNG saves an image to a PNG file.
func savePNG(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}

	return nil
}

// printStatistics prints statistics about the generated gradient.
func printStatistics(img *image.RGBA, logger *logrus.Logger) {
	bounds := img.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()

	// Sample colors from different positions
	samples := []struct {
		name string
		x, y int
	}{
		{"Top-Left", bounds.Min.X, bounds.Min.Y},
		{"Top-Right", bounds.Max.X - 1, bounds.Min.Y},
		{"Bottom-Left", bounds.Min.X, bounds.Max.Y - 1},
		{"Bottom-Right", bounds.Max.X - 1, bounds.Max.Y - 1},
		{"Center", bounds.Dx() / 2, bounds.Dy() / 2},
	}

	logger.Info("Gradient color samples:")
	for _, s := range samples {
		c := img.At(s.x, s.y)
		r, g, b, a := c.RGBA()
		logger.WithFields(logrus.Fields{
			"position": s.name,
			"r":        r >> 8,
			"g":        g >> 8,
			"b":        b >> 8,
			"a":        a >> 8,
		}).Info("Sample color")
	}

	logger.WithField("total_pixels", totalPixels).Info("Gradient statistics")
}
