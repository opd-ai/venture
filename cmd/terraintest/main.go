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
	algorithm  = flag.String("algorithm", "bsp", "Generation algorithm: bsp, cellular, maze, forest, city, composite, or multilevel")
	width      = flag.Int("width", 80, "Map width")
	height     = flag.Int("height", 50, "Map height")
	seed       = flag.Int64("seed", 12345, "Generation seed")
	output     = flag.String("output", "", "Output file (leave empty for console)")
	numLevels  = flag.Int("levels", 1, "Number of levels for multilevel generation")
	showAll    = flag.Bool("showAll", false, "Show all levels (multilevel only)")
	biomeCount = flag.Int("biomes", 3, "Number of biomes for composite generation (2-4)")
	genre      = flag.String("genre", "fantasy", "Genre theme: fantasy, scifi, horror, cyberpunk, or postapoc")
	visualize  = flag.String("visualize", "ascii", "Visualization mode: ascii, color, or stats")
)

func main() {
	flag.Parse()

	log.Printf("Generating terrain using %s algorithm", *algorithm)
	log.Printf("Size: %dx%d, Seed: %d", *width, *height, *seed)
	log.Printf("Genre: %s", *genre)
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
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"width":  *width,
			"height": *height,
		},
	}

	// Apply genre-specific defaults
	terrain.ApplyGenreDefaults(&params)

	// Add biome count for composite generation
	if *algorithm == "composite" {
		params.Custom["biomeCount"] = *biomeCount
		log.Printf("Biomes: %d", *biomeCount)
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

	// Render to string based on visualization mode
	var rendered string
	switch *visualize {
	case "color":
		rendered = renderTerrainColor(terr, *genre)
	case "stats":
		rendered = renderStats(terr)
	case "ascii":
		fallthrough
	default:
		rendered = renderTerrain(terr)
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

// generateMultiLevel handles multi-level dungeon generation
func generateMultiLevel() {
	// Create level generator
	gen := terrain.NewLevelGenerator()

	// Set up generation parameters
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"width":  *width,
			"height": *height,
		},
	}

	// Apply genre-specific defaults
	terrain.ApplyGenreDefaults(&params)

	// Generate all levels
	levels, err := gen.GenerateMultiLevel(*numLevels, *seed, params)
	if err != nil {
		log.Fatalf("Multi-level generation failed: %v", err)
	}

	log.Printf("Generated %d levels", len(levels))

	// Render based on showAll flag and visualization mode
	var rendered string
	if *showAll {
		// Show all levels
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("Multi-Level Dungeon: %d levels\n", len(levels)))
		builder.WriteString(fmt.Sprintf("Size: %dx%d per level, Seed: %d\n\n", *width, *height, *seed))

		for i, level := range levels {
			builder.WriteString(fmt.Sprintf("=== LEVEL %d ===\n", i))

			// Render based on visualization mode
			switch *visualize {
			case "color":
				builder.WriteString(renderTerrainColor(level, *genre))
			case "stats":
				builder.WriteString(renderStats(level))
			default:
				builder.WriteString(renderTerrain(level))
			}

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

		// Render based on visualization mode
		switch *visualize {
		case "color":
			rendered = renderTerrainColor(levels[0], *genre)
		case "stats":
			rendered = renderStats(levels[0])
		default:
			rendered = renderTerrain(levels[0])
		}
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

// renderTerrainColor converts terrain to colored ASCII art using ANSI codes
func renderTerrainColor(terr *terrain.Terrain, genreID string) string {
	var builder strings.Builder

	// Add header (no color)
	builder.WriteString(fmt.Sprintf("Terrain %dx%d (Seed: %d, Level: %d)\n", terr.Width, terr.Height, terr.Seed, terr.Level))
	builder.WriteString(fmt.Sprintf("Genre: %s, Rooms: %d, Stairs Up: %d, Stairs Down: %d\n\n", genreID, len(terr.Rooms), len(terr.StairsUp), len(terr.StairsDown)))

	// Render each row with colors
	for y := 0; y < terr.Height; y++ {
		for x := 0; x < terr.Width; x++ {
			tile := terr.GetTile(x, y)
			color := getTileColor(tile, genreID)
			char := getTileChar(tile)
			builder.WriteString(fmt.Sprintf("%s%s\033[0m", color, char))
		}
		builder.WriteString("\n")
	}

	// Add stats (no color)
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

	// Add color legend
	builder.WriteString("\nLegend (with colors):\n")
	builder.WriteString(fmt.Sprintf("  %s#\033[0m = Wall      %s.\033[0m = Floor     %s:\033[0m = Corridor  %s+\033[0m = Door\n",
		getTileColor(terrain.TileWall, genreID),
		getTileColor(terrain.TileFloor, genreID),
		getTileColor(terrain.TileCorridor, genreID),
		getTileColor(terrain.TileDoor, genreID)))
	builder.WriteString(fmt.Sprintf("  %sW\033[0m = Shallow   %s~\033[0m = Deep      %sT\033[0m = Tree      %s^\033[0m = Stairs Up\n",
		getTileColor(terrain.TileWaterShallow, genreID),
		getTileColor(terrain.TileWaterDeep, genreID),
		getTileColor(terrain.TileTree, genreID),
		getTileColor(terrain.TileStairsUp, genreID)))
	builder.WriteString(fmt.Sprintf("  %sv\033[0m = Stairs Dn %s[\033[0m = Trap Door %s?\033[0m = Secret    %s=\033[0m = Bridge\n",
		getTileColor(terrain.TileStairsDown, genreID),
		getTileColor(terrain.TileTrapDoor, genreID),
		getTileColor(terrain.TileSecretDoor, genreID),
		getTileColor(terrain.TileBridge, genreID)))
	builder.WriteString(fmt.Sprintf("  %s@\033[0m = Structure\n",
		getTileColor(terrain.TileStructure, genreID)))

	return builder.String()
}

// getTileColor returns ANSI color code for a tile type based on genre
func getTileColor(tile terrain.TileType, genreID string) string {
	// Genre-specific color schemes
	switch genreID {
	case "scifi":
		return getTileColorSciFi(tile)
	case "horror":
		return getTileColorHorror(tile)
	case "cyberpunk":
		return getTileColorCyberpunk(tile)
	case "postapoc":
		return getTileColorPostApoc(tile)
	default: // fantasy
		return getTileColorFantasy(tile)
	}
}

// getTileColorFantasy returns fantasy-themed colors
func getTileColorFantasy(tile terrain.TileType) string {
	switch tile {
	case terrain.TileWall:
		return "\033[90m" // Dark gray (stone)
	case terrain.TileFloor:
		return "\033[37m" // White (cobblestone)
	case terrain.TileCorridor:
		return "\033[37m" // White
	case terrain.TileDoor:
		return "\033[33m" // Yellow (wooden door)
	case terrain.TileWaterShallow:
		return "\033[96m" // Cyan (clear water)
	case terrain.TileWaterDeep:
		return "\033[34m" // Blue (deep water)
	case terrain.TileTree:
		return "\033[32m" // Green (oak tree)
	case terrain.TileStairsUp, terrain.TileStairsDown:
		return "\033[37m" // White (stone stairs)
	case terrain.TileTrapDoor:
		return "\033[33m" // Yellow
	case terrain.TileSecretDoor:
		return "\033[90m" // Dark gray (hidden)
	case terrain.TileBridge:
		return "\033[33m" // Yellow (wooden bridge)
	case terrain.TileStructure:
		return "\033[90m" // Dark gray (castle ruins)
	default:
		return "\033[37m" // White
	}
}

// getTileColorSciFi returns sci-fi themed colors
func getTileColorSciFi(tile terrain.TileType) string {
	switch tile {
	case terrain.TileWall:
		return "\033[90m" // Dark gray (metal panel)
	case terrain.TileFloor:
		return "\033[37m" // White (deck plating)
	case terrain.TileCorridor:
		return "\033[37m" // White
	case terrain.TileDoor:
		return "\033[96m" // Cyan (airlock)
	case terrain.TileWaterShallow:
		return "\033[92m" // Light green (coolant leak)
	case terrain.TileWaterDeep:
		return "\033[32m" // Green (coolant pool)
	case terrain.TileTree:
		return "\033[90m" // Dark gray (tech pillar)
	case terrain.TileStairsUp, terrain.TileStairsDown:
		return "\033[93m" // Light yellow (elevator)
	case terrain.TileTrapDoor:
		return "\033[90m" // Dark gray (hatch)
	case terrain.TileSecretDoor:
		return "\033[90m" // Dark gray (hidden panel)
	case terrain.TileBridge:
		return "\033[37m" // White (catwalk)
	case terrain.TileStructure:
		return "\033[90m" // Dark gray (tech building)
	default:
		return "\033[37m" // White
	}
}

// getTileColorHorror returns horror-themed colors
func getTileColorHorror(tile terrain.TileType) string {
	switch tile {
	case terrain.TileWall:
		return "\033[91m" // Light red (flesh wall)
	case terrain.TileFloor:
		return "\033[31m" // Red (bloodstained)
	case terrain.TileCorridor:
		return "\033[90m" // Dark gray (narrow passage)
	case terrain.TileDoor:
		return "\033[33m" // Yellow (rusty door)
	case terrain.TileWaterShallow:
		return "\033[35m" // Magenta (murky water)
	case terrain.TileWaterDeep:
		return "\033[31m" // Red (blood pool)
	case terrain.TileTree:
		return "\033[90m" // Dark gray (dead tree)
	case terrain.TileStairsUp, terrain.TileStairsDown:
		return "\033[33m" // Yellow (creaking stairs)
	case terrain.TileTrapDoor:
		return "\033[90m" // Dark gray
	case terrain.TileSecretDoor:
		return "\033[90m" // Dark gray
	case terrain.TileBridge:
		return "\033[33m" // Yellow (rickety bridge)
	case terrain.TileStructure:
		return "\033[90m" // Dark gray (abandoned building)
	default:
		return "\033[37m" // White
	}
}

// getTileColorCyberpunk returns cyberpunk-themed colors
func getTileColorCyberpunk(tile terrain.TileType) string {
	switch tile {
	case terrain.TileWall:
		return "\033[95m" // Light magenta (neon wall)
	case terrain.TileFloor:
		return "\033[36m" // Cyan (wet pavement)
	case terrain.TileCorridor:
		return "\033[90m" // Dark gray (alley)
	case terrain.TileDoor:
		return "\033[93m" // Light yellow (security door)
	case terrain.TileWaterShallow:
		return "\033[96m" // Light cyan (puddle)
	case terrain.TileWaterDeep:
		return "\033[34m" // Blue (flooded area)
	case terrain.TileTree:
		return "\033[95m" // Light magenta (neon sign)
	case terrain.TileStairsUp, terrain.TileStairsDown:
		return "\033[33m" // Yellow (fire escape)
	case terrain.TileTrapDoor:
		return "\033[90m" // Dark gray (sewer entrance)
	case terrain.TileSecretDoor:
		return "\033[90m" // Dark gray
	case terrain.TileBridge:
		return "\033[37m" // White (overpass)
	case terrain.TileStructure:
		return "\033[95m" // Light magenta (mega building)
	default:
		return "\033[37m" // White
	}
}

// getTileColorPostApoc returns post-apocalyptic themed colors
func getTileColorPostApoc(tile terrain.TileType) string {
	switch tile {
	case terrain.TileWall:
		return "\033[33m" // Yellow (rubble wall)
	case terrain.TileFloor:
		return "\033[90m" // Dark gray (cracked floor)
	case terrain.TileCorridor:
		return "\033[90m" // Dark gray (collapsed corridor)
	case terrain.TileDoor:
		return "\033[33m" // Yellow (broken door)
	case terrain.TileWaterShallow:
		return "\033[92m" // Light green (irradiated water)
	case terrain.TileWaterDeep:
		return "\033[32m" // Green (toxic pool)
	case terrain.TileTree:
		return "\033[33m" // Yellow (mutated tree)
	case terrain.TileStairsUp, terrain.TileStairsDown:
		return "\033[33m" // Yellow (debris stairs)
	case terrain.TileTrapDoor:
		return "\033[90m" // Dark gray (bunker entrance)
	case terrain.TileSecretDoor:
		return "\033[90m" // Dark gray
	case terrain.TileBridge:
		return "\033[33m" // Yellow (makeshift bridge)
	case terrain.TileStructure:
		return "\033[90m" // Dark gray (ruined building)
	default:
		return "\033[37m" // White
	}
}

// renderStats generates detailed statistics about the terrain
func renderStats(terr *terrain.Terrain) string {
	var builder strings.Builder

	// Header
	builder.WriteString("=== TERRAIN STATISTICS ===\n\n")
	builder.WriteString(fmt.Sprintf("Dimensions: %dx%d (%d total tiles)\n", terr.Width, terr.Height, terr.Width*terr.Height))
	builder.WriteString(fmt.Sprintf("Seed: %d, Level: %d\n\n", terr.Seed, terr.Level))

	// Tile type distribution
	tileCounts := make(map[terrain.TileType]int)
	for y := 0; y < terr.Height; y++ {
		for x := 0; x < terr.Width; x++ {
			tile := terr.GetTile(x, y)
			tileCounts[tile]++
		}
	}

	builder.WriteString("Tile Distribution:\n")
	totalTiles := terr.Width * terr.Height
	tileTypes := []terrain.TileType{
		terrain.TileWall,
		terrain.TileFloor,
		terrain.TileCorridor,
		terrain.TileDoor,
		terrain.TileWaterShallow,
		terrain.TileWaterDeep,
		terrain.TileTree,
		terrain.TileStairsUp,
		terrain.TileStairsDown,
		terrain.TileTrapDoor,
		terrain.TileSecretDoor,
		terrain.TileBridge,
		terrain.TileStructure,
	}

	for _, tileType := range tileTypes {
		count := tileCounts[tileType]
		if count > 0 {
			pct := float64(count) / float64(totalTiles) * 100
			builder.WriteString(fmt.Sprintf("  %-15s: %5d tiles (%.1f%%)\n", tileType.String(), count, pct))
		}
	}

	// Walkability metrics
	walkable := 0
	for y := 0; y < terr.Height; y++ {
		for x := 0; x < terr.Width; x++ {
			if terr.IsWalkable(x, y) {
				walkable++
			}
		}
	}

	builder.WriteString(fmt.Sprintf("\nWalkability:\n"))
	builder.WriteString(fmt.Sprintf("  Walkable tiles: %d/%d (%.1f%%)\n", walkable, totalTiles, float64(walkable)/float64(totalTiles)*100))
	builder.WriteString(fmt.Sprintf("  Non-walkable:   %d/%d (%.1f%%)\n", totalTiles-walkable, totalTiles, float64(totalTiles-walkable)/float64(totalTiles)*100))

	// Room information
	builder.WriteString(fmt.Sprintf("\nRooms: %d\n", len(terr.Rooms)))
	if len(terr.Rooms) > 0 {
		roomTypeCounts := make(map[terrain.RoomType]int)
		totalRoomArea := 0
		minRoomArea := terr.Width * terr.Height
		maxRoomArea := 0

		for _, room := range terr.Rooms {
			roomTypeCounts[room.Type]++
			area := room.Width * room.Height
			totalRoomArea += area
			if area < minRoomArea {
				minRoomArea = area
			}
			if area > maxRoomArea {
				maxRoomArea = area
			}
		}

		// Room types
		builder.WriteString("  Room Types:\n")
		roomTypes := []terrain.RoomType{
			terrain.RoomSpawn,
			terrain.RoomNormal,
			terrain.RoomTreasure,
			terrain.RoomBoss,
			terrain.RoomTrap,
			terrain.RoomExit,
		}
		for _, roomType := range roomTypes {
			count := roomTypeCounts[roomType]
			if count > 0 {
				builder.WriteString(fmt.Sprintf("    %-10s: %d\n", roomType.String(), count))
			}
		}

		avgRoomArea := float64(totalRoomArea) / float64(len(terr.Rooms))
		builder.WriteString(fmt.Sprintf("  Room Size: min=%d, max=%d, avg=%.1f tiles\n", minRoomArea, maxRoomArea, avgRoomArea))
	}

	// Stairs information
	builder.WriteString(fmt.Sprintf("\nStairs:\n"))
	builder.WriteString(fmt.Sprintf("  Up:   %d\n", len(terr.StairsUp)))
	builder.WriteString(fmt.Sprintf("  Down: %d\n", len(terr.StairsDown)))

	// Water coverage
	waterShallow := tileCounts[terrain.TileWaterShallow]
	waterDeep := tileCounts[terrain.TileWaterDeep]
	totalWater := waterShallow + waterDeep
	if totalWater > 0 {
		builder.WriteString(fmt.Sprintf("\nWater Features:\n"))
		builder.WriteString(fmt.Sprintf("  Total water: %d tiles (%.1f%%)\n", totalWater, float64(totalWater)/float64(totalTiles)*100))
		builder.WriteString(fmt.Sprintf("  Shallow:     %d tiles\n", waterShallow))
		builder.WriteString(fmt.Sprintf("  Deep:        %d tiles\n", waterDeep))
		if tileCounts[terrain.TileBridge] > 0 {
			builder.WriteString(fmt.Sprintf("  Bridges:     %d\n", tileCounts[terrain.TileBridge]))
		}
	}

	// Natural features
	trees := tileCounts[terrain.TileTree]
	if trees > 0 {
		builder.WriteString(fmt.Sprintf("\nNatural Features:\n"))
		builder.WriteString(fmt.Sprintf("  Trees: %d (%.1f%%)\n", trees, float64(trees)/float64(totalTiles)*100))
	}

	// Urban features
	structures := tileCounts[terrain.TileStructure]
	if structures > 0 {
		builder.WriteString(fmt.Sprintf("\nUrban Features:\n"))
		builder.WriteString(fmt.Sprintf("  Structures: %d (%.1f%%)\n", structures, float64(structures)/float64(totalTiles)*100))
	}

	// Special tiles
	trapDoors := tileCounts[terrain.TileTrapDoor]
	secretDoors := tileCounts[terrain.TileSecretDoor]
	if trapDoors > 0 || secretDoors > 0 {
		builder.WriteString(fmt.Sprintf("\nSpecial Tiles:\n"))
		if trapDoors > 0 {
			builder.WriteString(fmt.Sprintf("  Trap doors:   %d\n", trapDoors))
		}
		if secretDoors > 0 {
			builder.WriteString(fmt.Sprintf("  Secret doors: %d\n", secretDoors))
		}
	}

	return builder.String()
}
