package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

var (
	genreID = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	seed    = flag.Int64("seed", 12345, "Seed for deterministic generation")
	verbose = flag.Bool("verbose", false, "Show detailed color information")
	output  = flag.String("output", "", "Output file path (optional)")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("rendertest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{
		"genre": *genreID,
		"seed":  *seed,
	}).Info("Rendering Test Tool started")

	// Generate palette
	gen := palette.NewGeneratorWithLogger(logger)
	pal, err := gen.Generate(*genreID, *seed)
	if err != nil {
		logger.WithError(err).Fatal("palette generation failed")
	}

	logger.Info("palette generated successfully")

	// Display palette
	displayPalette(pal)

	// Save to file if requested
	if *output != "" {
		if err := savePalette(pal, *output); err != nil {
			logger.WithError(err).Fatal("failed to save palette")
		}
		logger.WithField("path", *output).Info("palette saved")
	}
}

func displayPalette(pal *palette.Palette) {
	fmt.Println("\n=== Color Palette ===")
	fmt.Println()

	displayColor("Primary", pal.Primary)
	displayColor("Secondary", pal.Secondary)
	displayColor("Background", pal.Background)
	displayColor("Text", pal.Text)
	displayColor("Accent1", pal.Accent1)
	displayColor("Accent2", pal.Accent2)
	displayColor("Danger", pal.Danger)
	displayColor("Success", pal.Success)

	if *verbose {
		fmt.Println("\nAdditional Colors:")
		for i, c := range pal.Colors {
			displayColor(fmt.Sprintf("Color[%d]", i), c)
		}
	}

	fmt.Println()
}

func displayColor(name string, c color.Color) {
	r, g, b, a := c.RGBA()
	// Convert from 16-bit to 8-bit
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)
	a8 := uint8(a >> 8)

	fmt.Printf("%-12s: RGB(%3d, %3d, %3d, %3d) Hex: #%02X%02X%02X\n",
		name, r8, g8, b8, a8, r8, g8, b8)

	if *verbose {
		// Show color bar (simplified terminal visualization)
		bar := makeColorBar(r8, g8, b8)
		fmt.Printf("              %s\n", bar)
	}
}

func makeColorBar(r, g, b uint8) string {
	// Simple ASCII color bar representation
	// In a real terminal with color support, this would use ANSI codes
	brightness := (int(r) + int(g) + int(b)) / 3
	chars := " ░▒▓█"
	index := brightness * len(chars) / 256
	if index >= len(chars) {
		index = len(chars) - 1
	}

	bar := ""
	for i := 0; i < 20; i++ {
		bar += string(chars[index])
	}
	return bar
}

func savePalette(pal *palette.Palette, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Venture Color Palette\n")
	fmt.Fprintf(f, "# Genre: %s, Seed: %d\n\n", *genreID, *seed)

	saveColor(f, "primary", pal.Primary)
	saveColor(f, "secondary", pal.Secondary)
	saveColor(f, "background", pal.Background)
	saveColor(f, "text", pal.Text)
	saveColor(f, "accent1", pal.Accent1)
	saveColor(f, "accent2", pal.Accent2)
	saveColor(f, "danger", pal.Danger)
	saveColor(f, "success", pal.Success)

	fmt.Fprintf(f, "\n# Additional Colors\n")
	for i, c := range pal.Colors {
		saveColor(f, fmt.Sprintf("color_%d", i), c)
	}

	return nil
}

func saveColor(f *os.File, name string, c color.Color) {
	r, g, b, _ := c.RGBA()
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)

	fmt.Fprintf(f, "%s=#%02X%02X%02X\n", name, r8, g8, b8)
}
