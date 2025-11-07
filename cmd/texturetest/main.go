package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/rendering/patterns"
	"github.com/sirupsen/logrus"
)

var (
	textureType = flag.String("texture", "stone", "Texture type (stone, wood, metal, organic)")
	genreID     = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapocalyptic)")
	seed        = flag.Int64("seed", 12345, "Seed for deterministic generation")
	width       = flag.Int("width", 128, "Texture width in pixels")
	height      = flag.Int("height", 128, "Texture height in pixels")
	detail      = flag.Float64("detail", 0.5, "Detail level (0.0-1.0)")
	scale       = flag.Float64("scale", 0.1, "Scale factor for noise patterns")
	output      = flag.String("output", "texture.png", "Output PNG file path")
	verbose     = flag.Bool("verbose", false, "Show detailed generation information")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("texturetest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{
		"texture": *textureType,
		"genre":   *genreID,
		"seed":    *seed,
		"size":    fmt.Sprintf("%dx%d", *width, *height),
	}).Info("Texture Test Tool started")

	// Parse texture type
	var texType patterns.TextureType
	switch *textureType {
	case "stone":
		texType = patterns.TextureStone
	case "wood":
		texType = patterns.TextureWood
	case "metal":
		texType = patterns.TextureMetal
	case "organic":
		texType = patterns.TextureOrganic
	default:
		logger.WithField("texture", *textureType).Fatal("unknown texture type")
	}

	// Generate colors based on genre
	color1, color2 := getGenreColors(*genreID, texType)

	// Create generator
	gen := patterns.NewGeneratorWithLogger(logger)

	// Configure texture generation
	config := patterns.TextureConfig{
		Texture:     texType,
		Width:       *width,
		Height:      *height,
		GenreID:     *genreID,
		Seed:        *seed,
		Color1:      color1,
		Color2:      color2,
		DetailLevel: *detail,
		Scale:       *scale,
	}

	logger.Info("generating texture pattern")

	// Generate texture
	img, err := gen.Generate(config)
	if err != nil {
		logger.WithError(err).Fatal("texture generation failed")
	}

	logger.Info("texture generated successfully")

	// Validate texture
	if err := gen.Validate(img); err != nil {
		logger.WithError(err).Warn("texture validation warning")
	}

	// Save to PNG file
	if err := saveTexture(img, *output); err != nil {
		logger.WithError(err).Fatal("failed to save texture")
	}

	logger.WithField("path", *output).Info("texture saved successfully")

	// Display information
	displayInfo(config, img)
}

// getGenreColors returns appropriate colors for the genre and texture type.
func getGenreColors(genreID string, texType patterns.TextureType) (color.Color, color.Color) {
	switch genreID {
	case "fantasy":
		switch texType {
		case patterns.TextureStone:
			return color.RGBA{R: 80, G: 70, B: 60, A: 255}, color.RGBA{R: 140, G: 130, B: 120, A: 255}
		case patterns.TextureWood:
			return color.RGBA{R: 90, G: 60, B: 30, A: 255}, color.RGBA{R: 140, G: 100, B: 60, A: 255}
		case patterns.TextureMetal:
			return color.RGBA{R: 150, G: 150, B: 150, A: 255}, color.RGBA{R: 200, G: 200, B: 200, A: 255}
		case patterns.TextureOrganic:
			return color.RGBA{R: 60, G: 80, B: 50, A: 255}, color.RGBA{R: 100, G: 140, B: 90, A: 255}
		}
	case "scifi":
		switch texType {
		case patterns.TextureStone:
			return color.RGBA{R: 70, G: 80, B: 90, A: 255}, color.RGBA{R: 120, G: 140, B: 160, A: 255}
		case patterns.TextureWood:
			return color.RGBA{R: 60, G: 70, B: 80, A: 255}, color.RGBA{R: 110, G: 130, B: 150, A: 255}
		case patterns.TextureMetal:
			return color.RGBA{R: 100, G: 120, B: 140, A: 255}, color.RGBA{R: 180, G: 200, B: 220, A: 255}
		case patterns.TextureOrganic:
			return color.RGBA{R: 50, G: 100, B: 120, A: 255}, color.RGBA{R: 90, G: 160, B: 190, A: 255}
		}
	case "horror":
		switch texType {
		case patterns.TextureStone:
			return color.RGBA{R: 40, G: 30, B: 30, A: 255}, color.RGBA{R: 80, G: 60, B: 60, A: 255}
		case patterns.TextureWood:
			return color.RGBA{R: 50, G: 30, B: 20, A: 255}, color.RGBA{R: 90, G: 60, B: 40, A: 255}
		case patterns.TextureMetal:
			return color.RGBA{R: 60, G: 50, B: 50, A: 255}, color.RGBA{R: 100, G: 80, B: 80, A: 255}
		case patterns.TextureOrganic:
			return color.RGBA{R: 80, G: 40, B: 40, A: 255}, color.RGBA{R: 130, G: 70, B: 70, A: 255}
		}
	case "cyberpunk":
		switch texType {
		case patterns.TextureStone:
			return color.RGBA{R: 20, G: 20, B: 30, A: 255}, color.RGBA{R: 60, G: 60, B: 90, A: 255}
		case patterns.TextureWood:
			return color.RGBA{R: 30, G: 30, B: 40, A: 255}, color.RGBA{R: 70, G: 70, B: 100, A: 255}
		case patterns.TextureMetal:
			return color.RGBA{R: 80, G: 90, B: 110, A: 255}, color.RGBA{R: 140, G: 160, B: 200, A: 255}
		case patterns.TextureOrganic:
			return color.RGBA{R: 100, G: 40, B: 120, A: 255}, color.RGBA{R: 160, G: 80, B: 200, A: 255}
		}
	case "postapocalyptic":
		switch texType {
		case patterns.TextureStone:
			return color.RGBA{R: 70, G: 60, B: 50, A: 255}, color.RGBA{R: 110, G: 100, B: 90, A: 255}
		case patterns.TextureWood:
			return color.RGBA{R: 80, G: 60, B: 40, A: 255}, color.RGBA{R: 120, G: 90, B: 60, A: 255}
		case patterns.TextureMetal:
			return color.RGBA{R: 100, G: 80, B: 60, A: 255}, color.RGBA{R: 150, G: 120, B: 90, A: 255}
		case patterns.TextureOrganic:
			return color.RGBA{R: 70, G: 80, B: 50, A: 255}, color.RGBA{R: 110, G: 130, B: 80, A: 255}
		}
	}

	// Default fallback
	return color.RGBA{R: 100, G: 100, B: 100, A: 255}, color.RGBA{R: 150, G: 150, B: 150, A: 255}
}

// saveTexture saves the texture to a PNG file.
func saveTexture(img *image.RGBA, path string) error {
	// Create output file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write PNG
	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// displayInfo shows information about the generated texture.
func displayInfo(config patterns.TextureConfig, img *image.RGBA) {
	fmt.Println("\n=== Texture Generation Info ===")
	fmt.Printf("Texture Type:  %s\n", config.Texture.String())
	fmt.Printf("Genre:         %s\n", config.GenreID)
	fmt.Printf("Dimensions:    %dx%d pixels\n", config.Width, config.Height)
	fmt.Printf("Seed:          %d\n", config.Seed)
	fmt.Printf("Detail Level:  %.2f\n", config.DetailLevel)
	fmt.Printf("Scale Factor:  %.2f\n", config.Scale)
	fmt.Printf("Output Size:   %dx%d pixels\n", img.Bounds().Dx(), img.Bounds().Dy())
	fmt.Println("\n=== Usage Examples ===")
	fmt.Println("  Generate different texture types:")
	fmt.Println("    ./texturetest -texture stone -genre fantasy -output stone_fantasy.png")
	fmt.Println("    ./texturetest -texture wood -genre scifi -output wood_scifi.png")
	fmt.Println("    ./texturetest -texture metal -genre cyberpunk -output metal_cyber.png")
	fmt.Println("    ./texturetest -texture organic -genre horror -output organic_horror.png")
	fmt.Println("\n  Adjust generation parameters:")
	fmt.Println("    ./texturetest -detail 0.8 -scale 0.05 -seed 99999")
	fmt.Println("    ./texturetest -width 256 -height 256 -detail 1.0")
	fmt.Println()
}
