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
	algorithm = flag.String("algorithm", "bsp", "Generation algorithm: bsp, cellular, maze, forest, city, composite, or multilevel")
	width     = flag.Int("width", 80, "Map width")
	height    = flag.Int("height", 50, "Map height")
	seed      = flag.Int64("seed", 12345, "Generation seed")
	output    = flag.String("output", "", "Output file (leave empty for console)")
	numLevels = flag.Int("levels", 1, "Number of levels for multilevel generation")
	showAll   = flag.Bool("showAll", false, "Show all levels (multilevel only)")
	biomeCount = flag.Int("biomes", 3, "Number of biomes for composite generation (2-4)")
)

func main() {
	flag.Parse()

	log.Printf("Generating terrain using %s algorithm", *algorithm)
	log.Printf("Size: %dx%d, Seed: %d", *width, *height, *seed)
	if *algorithm == "multilevel" {
		log.Printf("Levels: %d", *numLevels)
	}

	// Handle multi-level generation separately
	if *algorithm == "multilevel" {
		generateMultiLevel()
		return
	}

	// Create generator based on algorithm choice
	var gen procgen.Generator
	switch *algorithm {
	case "bsp":
		gen = terrain.NewBSPGenerator()
	case "cellular":
		gen = terrain.NewCellularGenerator()
	case "maze":
		gen = terrain.NewMazeGenerator()
	case "forest":
		gen = terrain.NewForestGenerator()
	case "city":
		gen = terrain.NewCityGenerator()
	case "composite":
		gen = terrain.NewCompositeGenerator()
	default:
		log.Fatalf("Unknown algorithm: %s (use 'bsp', 'cellular', 'maze', 'forest', 'city', 'composite', or 'multilevel')", *algorithm)
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

// generateMultiLevel handles multi-level dungeon generation
func generateMultiLevel() {
	// Create level generator
	gen := terrain.NewLevelGenerator()

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

	// Generate all levels
	levels, err := gen.GenerateMultiLevel(*numLevels, *seed, params)
	if err != nil {
		log.Fatalf("Multi-level generation failed: %v", err)
	}

	log.Printf("Generated %d levels", len(levels))

	// Render based on showAll flag
	var rendered string
	if *showAll {
		// Show all levels
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("Multi-Level Dungeon: %d levels\n", len(levels)))
		builder.WriteString(fmt.Sprintf("Size: %dx%d per level, Seed: %d\n\n", *width, *height, *seed))

		for i, level := range levels {
			builder.WriteString(fmt.Sprintf("=== LEVEL %d ===\n", i))
			builder.WriteString(renderTerrain(level))
			builder.WriteString("\n")

			// Show connectivity
			if i < len(levels)-1 {
				builder.WriteString("Connections:\n")
				if len(level.StairsDown) > 0 {
					builder.WriteString(fmt.Sprintf("  Stairs Down: %v\n", level.StairsDown))
				}
				if len(levels[i+1].StairsUp) > 0 {
					builder.WriteString(fmt.Sprintf("  Stairs Up (next level): %v\n", levels[i+1].StairsUp))
				}
				builder.WriteString("\n")
			}
		}

		rendered = builder.String()
	} else {
		// Show only first level
		log.Printf("Showing level 0 (use -showAll to see all levels)")
		rendered = renderTerrain(levels[0])
	}

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
