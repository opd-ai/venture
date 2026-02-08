package terrain

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestTerrainGeneratorTypes verifies that all terrain generator types can be instantiated
// and produce valid terrain output. This ensures the server's terrain type selection works.
func TestTerrainGeneratorTypes(t *testing.T) {
	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  100,
			"height": 80,
		},
	}

	tests := []struct {
		name          string
		createGen     func() procgen.Generator
		expectGraph   bool // true if generator returns DungeonGraph
		expectTerrain bool // true if generator returns Terrain
	}{
		{
			name:          "BSP generator",
			createGen:     func() procgen.Generator { return NewBSPGenerator() },
			expectTerrain: true,
		},
		{
			name:          "Cellular generator",
			createGen:     func() procgen.Generator { return NewCellularGenerator() },
			expectTerrain: true,
		},
		{
			name:          "City generator",
			createGen:     func() procgen.Generator { return NewCityGenerator() },
			expectTerrain: true,
		},
		{
			name:          "Forest generator",
			createGen:     func() procgen.Generator { return NewForestGenerator() },
			expectTerrain: true,
		},
		{
			name:          "Composite generator",
			createGen:     func() procgen.Generator { return NewCompositeGenerator() },
			expectTerrain: true,
		},
		{
			name: "Grammar generator",
			createGen: func() procgen.Generator {
				config := GetFantasyConfig(seed)
				return NewGraphGrammarGenerator(config)
			},
			expectGraph: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := tt.createGen()
			result, err := gen.Generate(seed, params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			if result == nil {
				t.Fatal("Generate() returned nil result")
			}

			// Verify result type
			if tt.expectGraph {
				graph, ok := result.(*DungeonGraph)
				if !ok {
					t.Errorf("expected *DungeonGraph, got %T", result)
				}
				if graph != nil && len(graph.Rooms) == 0 {
					t.Error("DungeonGraph has no rooms")
				}

				// Test conversion to Terrain
				terrain := GraphToTerrain(graph)
				if terrain == nil {
					t.Error("GraphToTerrain returned nil")
				}
				if terrain.Width != 100 || terrain.Height != 80 {
					t.Errorf("unexpected terrain dimensions: %dx%d", terrain.Width, terrain.Height)
				}
			}

			if tt.expectTerrain {
				terrain, ok := result.(*Terrain)
				if !ok {
					t.Errorf("expected *Terrain, got %T", result)
				}
				if terrain != nil {
					if terrain.Width != 100 || terrain.Height != 80 {
						t.Errorf("unexpected terrain dimensions: %dx%d", terrain.Width, terrain.Height)
					}
					// Note: Some generators (cellular, composite) may not produce Room structs
					// They generate terrain differently (caves, multi-biome areas)
				}
			}

			// Validate the result
			if err := gen.Validate(result); err != nil {
				t.Errorf("Validate() failed: %v", err)
			}
		})
	}
}

// TestGraphGrammarGenreConfigs verifies that all genre-specific L-system configs exist
// and can be used to generate valid dungeons.
func TestGraphGrammarGenreConfigs(t *testing.T) {
	seed := int64(99999)
	genres := []string{"fantasy", "scifi", "sci-fi", "horror", "cyberpunk", "post-apocalyptic", "postapocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    genre,
				Custom: map[string]interface{}{
					"width":  100,
					"height": 80,
				},
			}

			// Get config based on genre
			var config LSystemConfig
			switch genre {
			case "fantasy":
				config = GetFantasyConfig(seed)
			case "scifi", "sci-fi":
				config = GetSciFiConfig(seed)
			case "horror":
				config = GetHorrorConfig(seed)
			case "cyberpunk":
				config = GetCyberpunkConfig(seed)
			case "post-apocalyptic", "postapocalyptic":
				config = GetPostApocalypticConfig(seed)
			default:
				config = GetFantasyConfig(seed)
			}

			gen := NewGraphGrammarGenerator(config)
			result, err := gen.Generate(seed, params)
			if err != nil {
				t.Fatalf("Generate() failed for genre %s: %v", genre, err)
			}

			graph, ok := result.(*DungeonGraph)
			if !ok {
				t.Fatalf("expected *DungeonGraph, got %T", result)
			}

			if graph.StartRoom == nil {
				t.Error("generated graph has no start room")
			}

			if len(graph.Rooms) < 2 {
				t.Errorf("expected at least 2 rooms, got %d", len(graph.Rooms))
			}

			// Convert to terrain to ensure it works
			terrain := GraphToTerrain(graph)
			if terrain == nil {
				t.Error("GraphToTerrain returned nil")
			}
		})
	}
}
