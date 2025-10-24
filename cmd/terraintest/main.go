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
	algorithm = flag.String("algorithm", "bsp", "Generation algorithm: bsp, cellular, or maze")
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
	case "maze":
		gen = terrain.NewMazeGenerator()
	default:
		log.Fatalf("Unknown algorithm: %s (use 'bsp', 'cellular', or 'maze')", *algorithm)
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
	builder.WriteString(fmt.Sprintf("Terrain %dx%d (Seed: %d, Level: %d)\n", terr.Width, terr.Height, terr.Seed, terr.Level))
	builder.WriteString(fmt.Sprintf("Rooms: %d, Stairs Up: %d, Stairs Down: %d\n\n", len(terr.Rooms), len(terr.StairsUp), len(terr.StairsDown)))

	// Render each row
	for y := 0; y < terr.Height; y++ {
		for x := 0; x < terr.Width; x++ {
			tile := terr.GetTile(x, y)
			builder.WriteString(getTileChar(tile))
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

	// Add legend
	builder.WriteString("\nLegend:\n")
	builder.WriteString("  # = Wall      . = Floor     : = Corridor  + = Door\n")
	builder.WriteString("  W = Shallow   ~ = Deep      T = Tree      ^ = Stairs Up\n")
	builder.WriteString("  v = Stairs Dn [ = Trap Door ? = Secret    = = Bridge\n")
	builder.WriteString("  @ = Structure\n")

	return builder.String()
}

// getTileChar returns the ASCII character for a tile type
func getTileChar(tile terrain.TileType) string {
	switch tile {
	case terrain.TileWall:
		return "#"
	case terrain.TileFloor:
		return "."
	case terrain.TileCorridor:
		return ":"
	case terrain.TileDoor:
		return "+"
	case terrain.TileWaterShallow:
		return "W"
	case terrain.TileWaterDeep:
		return "~"
	case terrain.TileTree:
		return "T"
	case terrain.TileStairsUp:
		return "^"
	case terrain.TileStairsDown:
		return "v"
	case terrain.TileTrapDoor:
		return "["
	case terrain.TileSecretDoor:
		return "?"
	case terrain.TileBridge:
		return "="
	case terrain.TileStructure:
		return "@"
	default:
		return "?"
	}
}
