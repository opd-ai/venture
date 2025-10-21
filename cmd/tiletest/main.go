package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"

	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

var (
	tileType  = flag.String("type", "floor", "Tile type (floor, wall, door, corridor, water, lava, trap, stairs)")
	width     = flag.Int("width", 32, "Tile width in pixels")
	height    = flag.Int("height", 32, "Tile height in pixels")
	genre     = flag.String("genre", "fantasy", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	seed      = flag.Int64("seed", 12345, "Random seed for generation")
	variant   = flag.Float64("variant", 0.5, "Visual variant (0.0-1.0)")
	count     = flag.Int("count", 10, "Number of tiles to generate")
	output    = flag.String("output", "", "Output file for tile image (PNG format)")
	verbose   = flag.Bool("verbose", false, "Show verbose output")
)

func main() {
	flag.Parse()

	// Parse tile type
	var tileTypeEnum tiles.TileType
	switch *tileType {
	case "floor":
		tileTypeEnum = tiles.TileFloor
	case "wall":
		tileTypeEnum = tiles.TileWall
	case "door":
		tileTypeEnum = tiles.TileDoor
	case "corridor":
		tileTypeEnum = tiles.TileCorridor
	case "water":
		tileTypeEnum = tiles.TileWater
	case "lava":
		tileTypeEnum = tiles.TileLava
	case "trap":
		tileTypeEnum = tiles.TileTrap
	case "stairs":
		tileTypeEnum = tiles.TileStairs
	default:
		log.Fatalf("Unknown tile type: %s", *tileType)
	}

	// Create generator
	gen := tiles.NewGenerator()

	fmt.Printf("=== Tile Generator Test ===\n\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Type:    %s\n", *tileType)
	fmt.Printf("  Size:    %dx%d\n", *width, *height)
	fmt.Printf("  Genre:   %s\n", *genre)
	fmt.Printf("  Seed:    %d\n", *seed)
	fmt.Printf("  Variant: %.2f\n", *variant)
	fmt.Printf("  Count:   %d\n\n", *count)

	// Generate tiles
	for i := 0; i < *count; i++ {
		config := tiles.Config{
			Type:    tileTypeEnum,
			Width:   *width,
			Height:  *height,
			GenreID: *genre,
			Seed:    *seed + int64(i),
			Variant: *variant,
		}

		img, err := gen.Generate(config)
		if err != nil {
			log.Fatalf("Failed to generate tile %d: %v", i+1, err)
		}

		if *verbose {
			fmt.Printf("Tile %d:\n", i+1)
			fmt.Printf("  Seed:       %d\n", config.Seed)
			fmt.Printf("  Dimensions: %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
			fmt.Printf("  Generated:  ✓\n\n")
		}

		// Save first tile if output is specified
		if *output != "" && i == 0 {
			f, err := os.Create(*output)
			if err != nil {
				log.Fatalf("Failed to create output file: %v", err)
			}
			defer f.Close()

			err = png.Encode(f, img)
			if err != nil {
				log.Fatalf("Failed to encode PNG: %v", err)
			}

			fmt.Printf("Saved tile to: %s\n", *output)
		}
	}

	fmt.Printf("Successfully generated %d tiles\n", *count)
}
