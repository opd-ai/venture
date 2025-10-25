package terrain

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestGenreTerrainPreferences_AllGenresExist(t *testing.T) {
	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range expectedGenres {
		t.Run(genre, func(t *testing.T) {
			prefs, ok := GenreTerrainPreferences[genre]
			if !ok {
				t.Errorf("Genre %s not found in GenreTerrainPreferences", genre)
				return
			}

			// Check that preferences have required fields
			if len(prefs.Generators) == 0 {
				t.Errorf("Genre %s has no generators", genre)
			}
			if len(prefs.TileThemes) == 0 {
				t.Errorf("Genre %s has no tile themes", genre)
			}
			if prefs.WaterChance < 0.0 || prefs.WaterChance > 1.0 {
				t.Errorf("Genre %s has invalid WaterChance: %f (must be 0.0-1.0)", genre, prefs.WaterChance)
			}
		})
	}
}

func TestGetGeneratorForGenre_DepthSelection(t *testing.T) {
	tests := []struct {
		name     string
		genre    string
		depth    int
		wantType string
	}{
		{"fantasy_early", "fantasy", 1, "*terrain.BSPGenerator"},
		{"fantasy_mid", "fantasy", 5, "*terrain.CellularGenerator"},
		{"fantasy_late", "fantasy", 8, "*terrain.ForestGenerator"},
		{"fantasy_deep", "fantasy", 10, "*terrain.CompositeGenerator"},
		{"scifi_early", "scifi", 2, "*terrain.CityGenerator"},
		{"scifi_mid", "scifi", 5, "*terrain.MazeGenerator"},
		{"scifi_late", "scifi", 8, "*terrain.BSPGenerator"},
		{"scifi_deep", "scifi", 12, "*terrain.CompositeGenerator"},
		{"horror_early", "horror", 1, "*terrain.CellularGenerator"},
		{"horror_mid", "horror", 6, "*terrain.MazeGenerator"},
		{"horror_late", "horror", 9, "*terrain.ForestGenerator"},
		{"cyberpunk_early", "cyberpunk", 2, "*terrain.CityGenerator"},
		{"postapoc_early", "postapoc", 1, "*terrain.CellularGenerator"},
		{"unknown_genre", "unknown", 1, "*terrain.BSPGenerator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(12345))
			gen := GetGeneratorForGenre(tt.genre, tt.depth, rng)

			gotType := GetGeneratorName(gen)
			expectedName := ""
			switch tt.wantType {
			case "*terrain.BSPGenerator":
				expectedName = "BSP Dungeon"
			case "*terrain.CellularGenerator":
				expectedName = "Cellular Cave"
			case "*terrain.MazeGenerator":
				expectedName = "Maze"
			case "*terrain.ForestGenerator":
				expectedName = "Forest"
			case "*terrain.CityGenerator":
				expectedName = "City"
			case "*terrain.CompositeGenerator":
				expectedName = "Composite"
			}

			if gotType != expectedName {
				t.Errorf("GetGeneratorForGenre(%s, %d) = %s, want %s", tt.genre, tt.depth, gotType, expectedName)
			}
		})
	}
}

func TestGetGeneratorForGenre_Determinism(t *testing.T) {
	genre := "fantasy"
	depth := 5

	rng1 := rand.New(rand.NewSource(12345))
	gen1 := GetGeneratorForGenre(genre, depth, rng1)

	rng2 := rand.New(rand.NewSource(12345))
	gen2 := GetGeneratorForGenre(genre, depth, rng2)

	name1 := GetGeneratorName(gen1)
	name2 := GetGeneratorName(gen2)

	if name1 != name2 {
		t.Errorf("GetGeneratorForGenre not deterministic: %s != %s", name1, name2)
	}
}

func TestGetTileTheme(t *testing.T) {
	tests := []struct {
		name      string
		genre     string
		tile      TileType
		wantTheme string
	}{
		{"fantasy_wall", "fantasy", TileWall, "stone_wall"},
		{"fantasy_floor", "fantasy", TileFloor, "cobblestone"},
		{"fantasy_tree", "fantasy", TileTree, "ancient_oak"},
		{"scifi_wall", "scifi", TileWall, "metal_panel"},
		{"scifi_floor", "scifi", TileFloor, "deck_plating"},
		{"scifi_structure", "scifi", TileStructure, "tech_building"},
		{"horror_wall", "horror", TileWall, "flesh_wall"},
		{"horror_water", "horror", TileWaterDeep, "blood_pool"},
		{"cyberpunk_wall", "cyberpunk", TileWall, "neon_wall"},
		{"cyberpunk_structure", "cyberpunk", TileStructure, "mega_building"},
		{"postapoc_wall", "postapoc", TileWall, "rubble_wall"},
		{"postapoc_tree", "postapoc", TileTree, "mutated_tree"},
		{"unknown_genre", "unknown", TileWall, "unknown"},
		{"fantasy_unknown_tile", "fantasy", TileType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetTileTheme(tt.genre, tt.tile)
			if theme != tt.wantTheme {
				t.Errorf("GetTileTheme(%s, %v) = %s, want %s", tt.genre, tt.tile, theme, tt.wantTheme)
			}
		})
	}
}

func TestGetWaterChance(t *testing.T) {
	tests := []struct {
		name       string
		genre      string
		wantChance float64
	}{
		{"fantasy", "fantasy", 0.3},
		{"scifi", "scifi", 0.0},
		{"horror", "horror", 0.5},
		{"cyberpunk", "cyberpunk", 0.2},
		{"postapoc", "postapoc", 0.4},
		{"unknown", "unknown", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chance := GetWaterChance(tt.genre)
			if chance != tt.wantChance {
				t.Errorf("GetWaterChance(%s) = %f, want %f", tt.genre, chance, tt.wantChance)
			}
		})
	}
}

func TestGetTreeType(t *testing.T) {
	tests := []struct {
		name     string
		genre    string
		wantType string
	}{
		{"fantasy", "fantasy", "oak/pine"},
		{"scifi", "scifi", ""},
		{"horror", "horror", "dead_tree/withered"},
		{"cyberpunk", "cyberpunk", ""},
		{"postapoc", "postapoc", "mutated/dead"},
		{"unknown", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			treeType := GetTreeType(tt.genre)
			if treeType != tt.wantType {
				t.Errorf("GetTreeType(%s) = %s, want %s", tt.genre, treeType, tt.wantType)
			}
		})
	}
}

func TestGetTreeDensity(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		wantDensity float64
	}{
		{"fantasy", "fantasy", 0.3},
		{"scifi", "scifi", 0.0},
		{"horror", "horror", 0.4},
		{"cyberpunk", "cyberpunk", 0.0},
		{"postapoc", "postapoc", 0.2},
		{"unknown", "unknown", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			density := GetTreeDensity(tt.genre)
			if density != tt.wantDensity {
				t.Errorf("GetTreeDensity(%s) = %f, want %f", tt.genre, density, tt.wantDensity)
			}
		})
	}
}

func TestGetBuildingDensity(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		wantDensity float64
	}{
		{"fantasy", "fantasy", 0.7},
		{"scifi", "scifi", 0.8},
		{"horror", "horror", 0.5},
		{"cyberpunk", "cyberpunk", 0.9},
		{"postapoc", "postapoc", 0.4},
		{"unknown", "unknown", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			density := GetBuildingDensity(tt.genre)
			if density != tt.wantDensity {
				t.Errorf("GetBuildingDensity(%s) = %f, want %f", tt.genre, density, tt.wantDensity)
			}
		})
	}
}

func TestGetRoomChance(t *testing.T) {
	tests := []struct {
		name       string
		genre      string
		wantChance float64
	}{
		{"fantasy", "fantasy", 0.1},
		{"scifi", "scifi", 0.05},
		{"horror", "horror", 0.15},
		{"cyberpunk", "cyberpunk", 0.08},
		{"postapoc", "postapoc", 0.12},
		{"unknown", "unknown", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chance := GetRoomChance(tt.genre)
			if chance != tt.wantChance {
				t.Errorf("GetRoomChance(%s) = %f, want %f", tt.genre, chance, tt.wantChance)
			}
		})
	}
}

func TestApplyGenreDefaults(t *testing.T) {
	tests := []struct {
		name         string
		genre        string
		customParams map[string]interface{}
		checkKey     string
		wantValue    interface{}
	}{
		{
			name:         "fantasy_tree_density",
			genre:        "fantasy",
			customParams: map[string]interface{}{},
			checkKey:     "treeDensity",
			wantValue:    0.3,
		},
		{
			name:         "scifi_building_density",
			genre:        "scifi",
			customParams: map[string]interface{}{},
			checkKey:     "buildingDensity",
			wantValue:    0.8,
		},
		{
			name:         "horror_room_chance",
			genre:        "horror",
			customParams: map[string]interface{}{},
			checkKey:     "roomChance",
			wantValue:    0.15,
		},
		{
			name:         "custom_overrides_default",
			genre:        "fantasy",
			customParams: map[string]interface{}{"treeDensity": 0.9},
			checkKey:     "treeDensity",
			wantValue:    0.9,
		},
		{
			name:         "water_included",
			genre:        "fantasy",
			customParams: map[string]interface{}{},
			checkKey:     "includeWater",
			wantValue:    true,
		},
		{
			name:         "scifi_no_water",
			genre:        "scifi",
			customParams: map[string]interface{}{},
			checkKey:     "includeWater",
			wantValue:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				GenreID: tt.genre,
				Custom:  tt.customParams,
			}

			ApplyGenreDefaults(&params)

			gotValue, ok := params.Custom[tt.checkKey]
			if !ok {
				t.Errorf("ApplyGenreDefaults did not set %s", tt.checkKey)
				return
			}

			if gotValue != tt.wantValue {
				t.Errorf("ApplyGenreDefaults set %s = %v, want %v", tt.checkKey, gotValue, tt.wantValue)
			}
		})
	}
}

func TestApplyGenreDefaults_NilCustom(t *testing.T) {
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom:  nil,
	}

	ApplyGenreDefaults(&params)

	if params.Custom == nil {
		t.Error("ApplyGenreDefaults did not initialize Custom map")
		return
	}

	if _, ok := params.Custom["treeDensity"]; !ok {
		t.Error("ApplyGenreDefaults did not set treeDensity")
	}
}

func TestApplyGenreDefaults_EmptyGenre(t *testing.T) {
	params := procgen.GenerationParams{
		GenreID: "",
		Custom:  map[string]interface{}{},
	}

	ApplyGenreDefaults(&params)

	// Should default to fantasy
	treeDensity, ok := params.Custom["treeDensity"]
	if !ok {
		t.Error("ApplyGenreDefaults did not set treeDensity for empty genre")
		return
	}

	if treeDensity != 0.3 {
		t.Errorf("ApplyGenreDefaults with empty genre set treeDensity = %v, want 0.3 (fantasy default)", treeDensity)
	}
}

func TestGetGeneratorName(t *testing.T) {
	tests := []struct {
		name     string
		gen      procgen.Generator
		wantName string
	}{
		{"bsp", NewBSPGenerator(), "BSP Dungeon"},
		{"cellular", NewCellularGenerator(), "Cellular Cave"},
		{"maze", NewMazeGenerator(), "Maze"},
		{"forest", NewForestGenerator(), "Forest"},
		{"city", NewCityGenerator(), "City"},
		{"composite", NewCompositeGenerator(), "Composite"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := GetGeneratorName(tt.gen)
			if name != tt.wantName {
				t.Errorf("GetGeneratorName(%T) = %s, want %s", tt.gen, name, tt.wantName)
			}
		})
	}
}

func TestGenreTerrainPreferences_TileThemes(t *testing.T) {
	// Verify that all genres have themes for critical tile types
	criticalTiles := []TileType{
		TileWall,
		TileFloor,
		TileCorridor,
		TileDoor,
	}

	for genre, prefs := range GenreTerrainPreferences {
		for _, tile := range criticalTiles {
			t.Run(genre+"_"+tile.String(), func(t *testing.T) {
				theme, ok := prefs.TileThemes[tile]
				if !ok {
					t.Errorf("Genre %s missing theme for critical tile %s", genre, tile.String())
					return
				}
				if theme == "" {
					t.Errorf("Genre %s has empty theme for tile %s", genre, tile.String())
				}
			})
		}
	}
}

func TestGenreTerrainPreferences_GeneratorValidity(t *testing.T) {
	validGenerators := map[string]bool{
		"bsp":       true,
		"cellular":  true,
		"maze":      true,
		"forest":    true,
		"city":      true,
		"composite": true,
	}

	for genre, prefs := range GenreTerrainPreferences {
		for _, genName := range prefs.Generators {
			t.Run(genre+"_"+genName, func(t *testing.T) {
				if !validGenerators[genName] {
					t.Errorf("Genre %s has invalid generator name: %s", genre, genName)
				}
			})
		}
	}
}

func BenchmarkGetGeneratorForGenre(b *testing.B) {
	rng := rand.New(rand.NewSource(12345))
	for i := 0; i < b.N; i++ {
		GetGeneratorForGenre("fantasy", 5, rng)
	}
}

func BenchmarkGetTileTheme(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetTileTheme("fantasy", TileWall)
	}
}

func BenchmarkApplyGenreDefaults(b *testing.B) {
	for i := 0; i < b.N; i++ {
		params := procgen.GenerationParams{
			GenreID: "fantasy",
			Custom:  map[string]interface{}{},
		}
		ApplyGenreDefaults(&params)
	}
}
