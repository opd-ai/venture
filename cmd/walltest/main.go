// walltest demonstrates Phase 47 enhanced wall rendering.
// This tool generates sample wall tiles with anti-aliasing, corner blending,
// and shadow integration, saving them as PNG files for visual inspection.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/opd-ai/venture/pkg/rendering/tiles"
	"github.com/sirupsen/logrus"
)

var (
	outputDir = flag.String("output", "wall_samples", "Output directory for generated wall tiles")
	size      = flag.Int("size", 64, "Tile size in pixels (width and height)")
	seed      = flag.Int64("seed", 12345, "Random seed for generation")
	genreID   = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	verbose   = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()
	logger := setupLogger()
	logTestInfo(logger)

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		logger.Fatalf("Failed to create output directory: %v", err)
	}

	gen := tiles.NewGeneratorWithLogger(logger)
	testCases := createTestCases()

	successCount := generateTestCases(gen, testCases, logger)
	logger.Infof("Generation complete: %d/%d successful", successCount, len(testCases))

	if *genreID == "fantasy" {
		generateMultiGenreComparison(gen, logger)
	}

	logger.Info("All operations complete!")
}

// setupLogger creates and configures the logger.
func setupLogger() *logrus.Logger {
	logger := logrus.New()
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	return logger
}

// logTestInfo logs test configuration information.
func logTestInfo(logger *logrus.Logger) {
	logger.Info("Phase 47 Wall Rendering Test")
	logger.Infof("Output directory: %s", *outputDir)
	logger.Infof("Tile size: %dx%d", *size, *size)
	logger.Infof("Seed: %d", *seed)
	logger.Infof("Genre: %s", *genreID)
}

// createTestCases creates all test case configurations.
func createTestCases() []struct {
	name      string
	neighbors tiles.WallNeighbors
	antiAlias bool
	shadows   bool
	blendRad  int
} {
	return []struct {
		name      string
		neighbors tiles.WallNeighbors
		antiAlias bool
		shadows   bool
		blendRad  int
	}{
		{
			name:      "basic_noaa",
			neighbors: tiles.WallNeighbors{},
			antiAlias: false,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "basic_aa",
			neighbors: tiles.WallNeighbors{},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "corner_L_ne",
			neighbors: tiles.WallNeighbors{North: true, East: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "corner_L_se",
			neighbors: tiles.WallNeighbors{South: true, East: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "corner_L_sw",
			neighbors: tiles.WallNeighbors{South: true, West: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "corner_L_nw",
			neighbors: tiles.WallNeighbors{North: true, West: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "tjunction_north",
			neighbors: tiles.WallNeighbors{North: true, East: true, West: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "tjunction_south",
			neighbors: tiles.WallNeighbors{South: true, East: true, West: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "tjunction_east",
			neighbors: tiles.WallNeighbors{North: true, South: true, East: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "tjunction_west",
			neighbors: tiles.WallNeighbors{North: true, South: true, West: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "cross_junction",
			neighbors: tiles.WallNeighbors{North: true, South: true, East: true, West: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "corridor_horizontal",
			neighbors: tiles.WallNeighbors{East: true, West: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "corridor_vertical",
			neighbors: tiles.WallNeighbors{North: true, South: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  4,
		},
		{
			name:      "noshadow",
			neighbors: tiles.WallNeighbors{},
			antiAlias: true,
			shadows:   false,
			blendRad:  4,
		},
		{
			name:      "large_blend",
			neighbors: tiles.WallNeighbors{North: true, East: true},
			antiAlias: true,
			shadows:   true,
			blendRad:  8,
		},
	}
}

// generateTestCases generates all test cases and saves to files.
func generateTestCases(gen *tiles.Generator, testCases []struct {
	name      string
	neighbors tiles.WallNeighbors
	antiAlias bool
	shadows   bool
	blendRad  int
}, logger *logrus.Logger,
) int {
	logger.Infof("Generating %d test cases...", len(testCases))
	successCount := 0
	for i, tc := range testCases {
		logger.Infof("[%d/%d] Generating: %s", i+1, len(testCases), tc.name)

		config := tiles.DefaultEnhancedWallConfig()
		config.Config.Width = *size
		config.Config.Height = *size
		config.Config.Seed = *seed
		config.Config.GenreID = *genreID
		config.Neighbors = tc.neighbors
		config.EnableAntialiasing = tc.antiAlias
		config.EnableShadows = tc.shadows
		config.BlendRadius = tc.blendRad

		img, err := gen.GenerateEnhancedWall(config)
		if err != nil {
			logger.Errorf("Failed to generate %s: %v", tc.name, err)
			continue
		}

		filename := filepath.Join(*outputDir, fmt.Sprintf("%s_%s.png", tc.name, *genreID))
		if err := saveImage(filename, img, logger); err != nil {
			continue
		}

		logger.Infof("  ✓ Saved: %s", filename)
		successCount++
	}
	return successCount
}

// generateMultiGenreComparison generates comparison grid for different genres.
func generateMultiGenreComparison(gen *tiles.Generator, logger *logrus.Logger) {
	logger.Info("Generating multi-genre comparison...")
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		logger.Infof("  Genre: %s", genre)

		config := tiles.DefaultEnhancedWallConfig()
		config.Config.Width = *size
		config.Config.Height = *size
		config.Config.Seed = *seed
		config.Config.GenreID = genre
		config.Neighbors = tiles.WallNeighbors{North: true, East: true}
		config.EnableAntialiasing = true
		config.EnableShadows = true

		img, err := gen.GenerateEnhancedWall(config)
		if err != nil {
			logger.Errorf("Failed to generate %s: %v", genre, err)
			continue
		}

		filename := filepath.Join(*outputDir, fmt.Sprintf("genre_%s.png", genre))
		if err := saveImage(filename, img, logger); err != nil {
			continue
		}

		logger.Infof("  ✓ Saved: %s", filename)
	}
}

// saveImage saves an image to a PNG file.
func saveImage(filename string, img image.Image, logger *logrus.Logger) error {
	f, err := os.Create(filename)
	if err != nil {
		logger.Errorf("Failed to create file %s: %v", filename, err)
		return err
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		logger.Errorf("Failed to encode PNG %s: %v", filename, err)
		return err
	}
	return nil
}
