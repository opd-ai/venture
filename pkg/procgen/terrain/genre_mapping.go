package terrain

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TerrainPreference defines genre-specific terrain generation preferences.
type TerrainPreference struct {
	// Generators is a list of preferred generator types for this genre
	Generators []string

	// TileThemes maps tile types to genre-specific theme names
	TileThemes map[TileType]string

	// WaterChance is the probability (0.0-1.0) of including water features
	WaterChance float64

	// TreeType is the genre-specific tree description (empty if no trees)
	TreeType string

	// TreeDensity is the default tree density for forest generation
	TreeDensity float64

	// BuildingDensity is the default building density for city generation
	BuildingDensity float64

	// RoomChance is the default chance of rooms in maze generation
	RoomChance float64
}

// GenreTerrainPreferences maps genre IDs to their terrain preferences.
var GenreTerrainPreferences = map[string]TerrainPreference{
	"fantasy": {
		Generators: []string{"bsp", "cellular", "forest"},
		TileThemes: map[TileType]string{
			TileWall:         "stone_wall",
			TileFloor:        "cobblestone",
			TileCorridor:     "stone_corridor",
			TileDoor:         "wooden_door",
			TileWaterShallow: "clear_water",
			TileWaterDeep:    "deep_water",
			TileTree:         "ancient_oak",
			TileStairsUp:     "stone_stairs_up",
			TileStairsDown:   "stone_stairs_down",
			TileTrapDoor:     "concealed_trapdoor",
			TileSecretDoor:   "hidden_door",
			TileBridge:       "wooden_bridge",
			TileStructure:    "castle_ruins",
		},
		WaterChance:     0.3,
		TreeType:        "oak/pine",
		TreeDensity:     0.3,
		BuildingDensity: 0.7,
		RoomChance:      0.1,
	},
	"scifi": {
		Generators: []string{"city", "maze", "bsp"},
		TileThemes: map[TileType]string{
			TileWall:         "metal_panel",
			TileFloor:        "deck_plating",
			TileCorridor:     "corridor_plating",
			TileDoor:         "airlock",
			TileWaterShallow: "coolant_leak",
			TileWaterDeep:    "coolant_pool",
			TileTree:         "tech_pillar",
			TileStairsUp:     "elevator_up",
			TileStairsDown:   "elevator_down",
			TileTrapDoor:     "maintenance_hatch",
			TileSecretDoor:   "hidden_panel",
			TileBridge:       "catwalk",
			TileStructure:    "tech_building",
		},
		WaterChance:     0.0,
		TreeType:        "",
		TreeDensity:     0.0,
		BuildingDensity: 0.8,
		RoomChance:      0.05,
	},
	"horror": {
		Generators: []string{"cellular", "maze", "forest"},
		TileThemes: map[TileType]string{
			TileWall:         "flesh_wall",
			TileFloor:        "bloodstained_floor",
			TileCorridor:     "narrow_passage",
			TileDoor:         "rusty_door",
			TileWaterShallow: "murky_water",
			TileWaterDeep:    "blood_pool",
			TileTree:         "dead_tree",
			TileStairsUp:     "creaking_stairs_up",
			TileStairsDown:   "creaking_stairs_down",
			TileTrapDoor:     "hidden_trapdoor",
			TileSecretDoor:   "concealed_door",
			TileBridge:       "rickety_bridge",
			TileStructure:    "abandoned_building",
		},
		WaterChance:     0.5,
		TreeType:        "dead_tree/withered",
		TreeDensity:     0.4,
		BuildingDensity: 0.5,
		RoomChance:      0.15,
	},
	"cyberpunk": {
		Generators: []string{"city", "maze", "cellular"},
		TileThemes: map[TileType]string{
			TileWall:         "neon_wall",
			TileFloor:        "wet_pavement",
			TileCorridor:     "alley",
			TileDoor:         "security_door",
			TileWaterShallow: "puddle",
			TileWaterDeep:    "flooded_area",
			TileTree:         "neon_sign",
			TileStairsUp:     "fire_escape_up",
			TileStairsDown:   "fire_escape_down",
			TileTrapDoor:     "sewer_entrance",
			TileSecretDoor:   "hidden_entrance",
			TileBridge:       "overpass",
			TileStructure:    "mega_building",
		},
		WaterChance:     0.2,
		TreeType:        "",
		TreeDensity:     0.0,
		BuildingDensity: 0.9,
		RoomChance:      0.08,
	},
	"postapoc": {
		Generators: []string{"cellular", "city", "forest"},
		TileThemes: map[TileType]string{
			TileWall:         "rubble_wall",
			TileFloor:        "cracked_floor",
			TileCorridor:     "collapsed_corridor",
			TileDoor:         "broken_door",
			TileWaterShallow: "irradiated_water",
			TileWaterDeep:    "toxic_pool",
			TileTree:         "mutated_tree",
			TileStairsUp:     "debris_stairs_up",
			TileStairsDown:   "debris_stairs_down",
			TileTrapDoor:     "bunker_entrance",
			TileSecretDoor:   "hidden_bunker",
			TileBridge:       "makeshift_bridge",
			TileStructure:    "ruined_building",
		},
		WaterChance:     0.4,
		TreeType:        "mutated/dead",
		TreeDensity:     0.2,
		BuildingDensity: 0.4,
		RoomChance:      0.12,
	},
}

// GetGeneratorForGenre returns an appropriate terrain generator based on genre and depth.
// The function considers genre preferences and depth level to select the most suitable generator.
//
// Depth-based selection:
//   - Depth 1-3: First preferred generator (usually structured)
//   - Depth 4-6: Second preferred generator (usually organic)
//   - Depth 7-9: Third preferred generator or maze (confusing)
//   - Depth 10+: Composite (multi-biome)
//
// If the genre is unknown, returns a BSP generator as fallback.
func GetGeneratorForGenre(genreID string, depth int, rng *rand.Rand) procgen.Generator {
	prefs, ok := GenreTerrainPreferences[genreID]
	if !ok {
		// Unknown genre - return default BSP generator
		return NewBSPGenerator()
	}

	// Depth 10+ always uses composite for variety
	if depth >= 10 {
		return NewCompositeGenerator()
	}

	// Select generator based on depth and genre preferences
	var generatorName string
	if depth <= 3 && len(prefs.Generators) > 0 {
		// Early depths: use first preference (structured)
		generatorName = prefs.Generators[0]
	} else if depth <= 6 && len(prefs.Generators) > 1 {
		// Mid depths: use second preference (organic)
		generatorName = prefs.Generators[1]
	} else if depth <= 9 && len(prefs.Generators) > 2 {
		// Late depths: use third preference or maze
		generatorName = prefs.Generators[2]
	} else if len(prefs.Generators) > 0 {
		// Fallback: random selection from preferences
		generatorName = prefs.Generators[rng.Intn(len(prefs.Generators))]
	} else {
		// No preferences: default to BSP
		generatorName = "bsp"
	}

	// Create the appropriate generator
	switch generatorName {
	case "bsp":
		return NewBSPGenerator()
	case "cellular":
		return NewCellularGenerator()
	case "maze":
		return NewMazeGenerator()
	case "forest":
		return NewForestGenerator()
	case "city":
		return NewCityGenerator()
	case "composite":
		return NewCompositeGenerator()
	default:
		return NewBSPGenerator()
	}
}

// GetTileTheme returns the genre-specific theme name for a tile type.
// Returns "unknown" if the genre or tile type has no specific theme.
//
// The theme strings can be used by rendering systems to select appropriate
// visual assets, colors, or descriptions.
func GetTileTheme(genreID string, tile TileType) string {
	prefs, ok := GenreTerrainPreferences[genreID]
	if !ok {
		return "unknown"
	}

	theme, ok := prefs.TileThemes[tile]
	if !ok {
		return "unknown"
	}

	return theme
}

// GetWaterChance returns the probability of including water features for a genre.
// Returns 0.0 if the genre is unknown.
func GetWaterChance(genreID string) float64 {
	prefs, ok := GenreTerrainPreferences[genreID]
	if !ok {
		return 0.0
	}
	return prefs.WaterChance
}

// GetTreeType returns the genre-specific tree type description.
// Returns empty string if the genre has no trees or is unknown.
func GetTreeType(genreID string) string {
	prefs, ok := GenreTerrainPreferences[genreID]
	if !ok {
		return ""
	}
	return prefs.TreeType
}

// GetTreeDensity returns the default tree density for a genre.
// Returns 0.0 if the genre is unknown.
func GetTreeDensity(genreID string) float64 {
	prefs, ok := GenreTerrainPreferences[genreID]
	if !ok {
		return 0.0
	}
	return prefs.TreeDensity
}

// GetBuildingDensity returns the default building density for a genre.
// Returns 0.0 if the genre is unknown.
func GetBuildingDensity(genreID string) float64 {
	prefs, ok := GenreTerrainPreferences[genreID]
	if !ok {
		return 0.0
	}
	return prefs.BuildingDensity
}

// GetRoomChance returns the default room generation chance for a genre.
// Returns 0.0 if the genre is unknown.
func GetRoomChance(genreID string) float64 {
	prefs, ok := GenreTerrainPreferences[genreID]
	if !ok {
		return 0.0
	}
	return prefs.RoomChance
}

// ApplyGenreDefaults modifies generation parameters to include genre-specific defaults
// if they are not already specified in the Custom map.
//
// This function should be called by generators to ensure genre preferences are respected.
func ApplyGenreDefaults(params *procgen.GenerationParams) {
	if params.Custom == nil {
		params.Custom = make(map[string]interface{})
	}

	genreID := params.GenreID
	if genreID == "" {
		genreID = "fantasy" // Default genre
	}

	// Apply tree density if not specified
	if _, ok := params.Custom["treeDensity"]; !ok {
		params.Custom["treeDensity"] = GetTreeDensity(genreID)
	}

	// Apply building density if not specified
	if _, ok := params.Custom["buildingDensity"]; !ok {
		params.Custom["buildingDensity"] = GetBuildingDensity(genreID)
	}

	// Apply room chance if not specified
	if _, ok := params.Custom["roomChance"]; !ok {
		params.Custom["roomChance"] = GetRoomChance(genreID)
	}

	// Apply water features flag if not specified
	if _, ok := params.Custom["includeWater"]; !ok {
		waterChance := GetWaterChance(genreID)
		// For determinism, we'd need to use a seeded RNG here
		// For now, just check if water is allowed for this genre
		params.Custom["includeWater"] = waterChance > 0.0
	}
}

// GetGeneratorName returns a human-readable name for a generator.
// Used for logging and debugging.
func GetGeneratorName(gen procgen.Generator) string {
	switch gen.(type) {
	case *BSPGenerator:
		return "BSP Dungeon"
	case *CellularGenerator:
		return "Cellular Cave"
	case *MazeGenerator:
		return "Maze"
	case *ForestGenerator:
		return "Forest"
	case *CityGenerator:
		return "City"
	case *CompositeGenerator:
		return "Composite"
	default:
		return fmt.Sprintf("Unknown (%T)", gen)
	}
}
