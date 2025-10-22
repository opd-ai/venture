package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

var (
	algorithm = flag.String("algorithm", "bsp", "Generation algorithm: bsp or cellular")
	width     = flag.Int("width", 80, "Map width")
	height    = flag.Int("height", 50, "Map height")
	seed      = flag.Int64("seed", 12345, "Generation seed")
	output    = flag.String("output", "", "Output file (leave empty for console)")
)

func main() {
	flag.Parse()

	log.Printf("Generating terrain using %s algorithm", *algorithm)
	log.Printf("Size: %dx%d, Seed: %d", *width, *height, *seed)

	// Create generator based on algorithm choice
	var gen procgen.Generator
	switch *algorithm {
	case "bsp":
		gen = terrain.NewBSPGenerator()
	case "cellular":
		gen = terrain.NewCellularGenerator()
	default:
		log.Fatalf("Unknown algorithm: %s (use 'bsp' or 'cellular')", *algorithm)
	}

	// Set up generation parameters
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  *width,
			"height": *height,
		},
	}

	// Generate terrain
	result, err := gen.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	terr, ok := result.(*terrain.Terrain)
	if !ok {
		log.Fatal("Result is not a Terrain")
	}

	// Validate
	if err := gen.Validate(terr); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	log.Printf("Generated %d rooms", len(terr.Rooms))

	// Render to string
	rendered := renderTerrain(terr)

	// Output to file or console
	if *output != "" {
		if err := os.WriteFile(*output, []byte(rendered), 0o644); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		log.Printf("Terrain saved to %s", *output)
	} else {
		fmt.Println(rendered)
	}
}

// renderTerrain converts terrain to ASCII art
func renderTerrain(terr *terrain.Terrain) string {
	var builder strings.Builder

	// Add header
	builder.WriteString(fmt.Sprintf("Terrain %dx%d (Seed: %d)\n", terr.Width, terr.Height, terr.Seed))
	builder.WriteString(fmt.Sprintf("Rooms: %d\n\n", len(terr.Rooms)))

	// Render each row
	for y := 0; y < terr.Height; y++ {
		for x := 0; x < terr.Width; x++ {
			tile := terr.GetTile(x, y)
			switch tile {
			case terrain.TileWall:
				builder.WriteString("#")
			case terrain.TileFloor:
				builder.WriteString(".")
			case terrain.TileCorridor:
				builder.WriteString(":")
			case terrain.TileDoor:
				builder.WriteString("+")
			default:
				builder.WriteString("?")
			}
		}
		builder.WriteString("\n")
	}

	// Add stats
	walkable := 0
	for y := 0; y < terr.Height; y++ {
		for x := 0; x < terr.Width; x++ {
			if terr.IsWalkable(x, y) {
				walkable++
			}
		}
	}

	totalTiles := terr.Width * terr.Height
	builder.WriteString(fmt.Sprintf("\nWalkable tiles: %d/%d (%.1f%%)\n",
		walkable, totalTiles, float64(walkable)/float64(totalTiles)*100))

	return builder.String()
}
