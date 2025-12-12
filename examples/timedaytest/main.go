// Command timedaytest demonstrates time-of-day color palette modulation.
// Phase 17.3: Time-of-Day Color Shifts
//
// This tool generates palettes for different times of day and visualizes
// the color transitions. It supports all genres and can export to PNG.
//
// Usage:
//
//	go run ./cmd/timedaytest -genre fantasy -seed 12345
//	go run ./cmd/timedaytest -genre scifi -time night -output scifi-night.png
//	go run ./cmd/timedaytest -genre horror -transition 0.5 -verbose
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
	genreID    = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	seed       = flag.Int64("seed", 12345, "Random seed for generation")
	timeOfDay  = flag.String("time", "all", "Time of day (dawn, day, dusk, night, all)")
	transition = flag.Float64("transition", 0.0, "Transition progress (0.0-1.0)")
	intensity  = flag.Float64("intensity", 1.0, "Effect intensity (0.0-1.0)")
	output     = flag.String("output", "", "Output PNG file (optional)")
	verbose    = flag.Bool("verbose", false, "Verbose logging")
)

func main() {
	flag.Parse()

	// Setup logging
	log := logrus.New()
	if *verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.WithFields(logrus.Fields{
		"genre": *genreID,
		"seed":  *seed,
		"time":  *timeOfDay,
	}).Info("Time-of-Day Color Shifts Test")

	// Create generator
	gen := palette.NewGeneratorWithLogger(log)

	// Generate palettes for requested time(s)
	if *timeOfDay == "all" {
		generateAllTimes(gen, log)
	} else {
		generateSingleTime(gen, log)
	}

	// Generate visualization if output requested
	if *output != "" {
		if err := generateVisualization(gen, log); err != nil {
			log.WithError(err).Fatal("Failed to generate visualization")
		}
		log.WithField("file", *output).Info("Visualization saved")
	}
}

func generateAllTimes(gen *palette.Generator, log *logrus.Logger) {
	times := []palette.TimeOfDay{
		palette.TimeOfDayDawn,
		palette.TimeOfDayDay,
		palette.TimeOfDayDusk,
		palette.TimeOfDayNight,
	}

	fmt.Printf("\nTime-of-Day Palette Comparison for %s (seed: %d)\n", *genreID, *seed)
	fmt.Println("============================================================")

	for _, time := range times {
		config := palette.TimeConfig{
			CurrentTime:         time,
			TransitionProgress:  *transition,
			IntensityMultiplier: *intensity,
		}

		pal, err := gen.GenerateWithTime(*genreID, *seed, config)
		if err != nil {
			log.WithError(err).WithField("time", time.String()).Error("Generation failed")
			continue
		}

		printPalette(time.String(), pal)
	}
}

func generateSingleTime(gen *palette.Generator, log *logrus.Logger) {
	time := parseTimeOfDay(*timeOfDay)
	if time < 0 {
		log.WithField("time", *timeOfDay).Fatal("Invalid time of day")
	}

	config := palette.TimeConfig{
		CurrentTime:         time,
		TransitionProgress:  *transition,
		IntensityMultiplier: *intensity,
	}

	pal, err := gen.GenerateWithTime(*genreID, *seed, config)
	if err != nil {
		log.WithError(err).Fatal("Generation failed")
	}

	fmt.Printf("\nPalette: %s at %s (seed: %d)\n", *genreID, time.String(), *seed)
	if *transition > 0 {
		fmt.Printf("Transition: %.1f%% to next time state\n", *transition*100)
	}
	if *intensity != 1.0 {
		fmt.Printf("Intensity: %.1f%%\n", *intensity*100)
	}
	fmt.Println("============================================================")
	printPalette("", pal)
}

func printPalette(label string, pal *palette.Palette) {
	if label != "" {
		fmt.Printf("\n%s:\n", label)
	}

	printColor("Primary", pal.Primary)
	printColor("Secondary", pal.Secondary)
	printColor("Background", pal.Background)
	printColor("Text", pal.Text)
	printColor("Accent1", pal.Accent1)
	printColor("Accent2", pal.Accent2)
	printColor("Accent3", pal.Accent3)

	fmt.Printf("  Additional Colors: %d\n", len(pal.Colors))
}

func printColor(name string, c color.Color) {
	if c == nil {
		fmt.Printf("  %-12s: nil\n", name)
		return
	}

	r, g, b, a := c.RGBA()
	fmt.Printf("  %-12s: RGB(%3d, %3d, %3d) A=%3d  [%s]\n",
		name,
		r>>8, g>>8, b>>8, a>>8,
		colorBar(r>>8, g>>8, b>>8))
}

func colorBar(r, g, b uint32) string {
	// Simple luminance calculation
	lum := (299*r + 587*g + 114*b) / 1000

	if lum > 200 {
		return "████████ LIGHT"
	} else if lum > 150 {
		return "██████   MID-LIGHT"
	} else if lum > 100 {
		return "████     MEDIUM"
	} else if lum > 50 {
		return "██       MID-DARK"
	} else {
		return "█        DARK"
	}
}

func parseTimeOfDay(s string) palette.TimeOfDay {
	switch s {
	case "dawn":
		return palette.TimeOfDayDawn
	case "day":
		return palette.TimeOfDayDay
	case "dusk":
		return palette.TimeOfDayDusk
	case "night":
		return palette.TimeOfDayNight
	default:
		return -1
	}
}

func generateVisualization(gen *palette.Generator, log *logrus.Logger) error {
	times := []palette.TimeOfDay{
		palette.TimeOfDayDawn,
		palette.TimeOfDayDay,
		palette.TimeOfDayDusk,
		palette.TimeOfDayNight,
	}

	// Create image: 800x600 (200x600 per time state)
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Fill background
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, color.RGBA{30, 30, 30, 255})
		}
	}

	// Generate and draw each time state
	for i, time := range times {
		config := palette.TimeConfig{
			CurrentTime:         time,
			TransitionProgress:  *transition,
			IntensityMultiplier: *intensity,
		}

		pal, err := gen.GenerateWithTime(*genreID, *seed, config)
		if err != nil {
			return fmt.Errorf("failed to generate palette for %s: %w", time.String(), err)
		}

		drawPaletteSection(img, i*200, pal, time.String())
	}

	// Save to file
	f, err := os.Create(*output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

func drawPaletteSection(img *image.RGBA, xOffset int, pal *palette.Palette, label string) {
	colors := []color.Color{
		pal.Primary,
		pal.Secondary,
		pal.Background,
		pal.Text,
		pal.Accent1,
		pal.Accent2,
		pal.Accent3,
		pal.Highlight1,
		pal.Highlight2,
		pal.Shadow1,
	}

	// Draw color swatches (10 colors, 60 pixels each)
	for i, c := range colors {
		if c == nil {
			continue
		}

		r, g, b, a := c.RGBA()
		rgba := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}

		yStart := i * 60
		yEnd := (i + 1) * 60

		for y := yStart; y < yEnd; y++ {
			for x := xOffset; x < xOffset+200; x++ {
				img.Set(x, y, rgba)
			}
		}
	}

	// Draw label on first swatch
	// (In a real implementation, we'd draw text. For simplicity, we'll skip that.)
}
