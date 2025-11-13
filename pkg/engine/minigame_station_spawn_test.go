// Package engine provides tests for mini-game station spawning (Phase 27.3).
package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestSpawnMiniGameStation(t *testing.T) {
	tests := []struct {
		name        string
		seed        int64
		params      procgen.GenerationParams
		worldNil    bool
		wantErr     bool
		wantCompMin int // Minimum number of components
	}{
		{
			name: "fantasy card station",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			worldNil:    false,
			wantErr:     false,
			wantCompMin: 7, // position, velocity, station, contextAction, collider, team, sprite, animation
		},
		{
			name: "sci-fi hacking station",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      5,
				GenreID:    "sci-fi",
			},
			worldNil:    false,
			wantErr:     false,
			wantCompMin: 7,
		},
		{
			name: "nil world error",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			worldNil: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var world *World
			if !tt.worldNil {
				world = NewWorld()
			}

			entity, err := SpawnMiniGameStation(world, 100.0, 200.0, tt.seed, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if entity == nil {
				t.Fatal("expected entity, got nil")
			}

			// Check position
			posComp, ok := entity.GetComponent("position")
			if !ok {
				t.Error("entity missing position component")
			} else {
				pos := posComp.(*PositionComponent)
				if pos.X != 100.0 || pos.Y != 200.0 {
					t.Errorf("expected position (100, 200), got (%.0f, %.0f)", pos.X, pos.Y)
				}
			}

			// Check velocity
			if _, ok := entity.GetComponent("velocity"); !ok {
				t.Error("entity missing velocity component")
			}

			// Check station component
			stationCompRaw, ok := entity.GetComponent("minigameStation")
			if !ok {
				t.Error("entity missing minigameStation component")
			} else {
				station := stationCompRaw.(*MiniGameStationComponent)
				if station.Difficulty < 0.0 || station.Difficulty > 1.0 {
					t.Errorf("invalid difficulty: %v", station.Difficulty)
				}
				if station.IsOccupied {
					t.Error("new station should not be occupied")
				}
			}

			// Check context action
			ctxCompRaw, ok := entity.GetComponent("contextAction")
			if !ok {
				t.Error("entity missing contextAction component")
			} else {
				ctx := ctxCompRaw.(*ContextActionComponent)
				if ctx.ActionType != ActionPlayGame {
					t.Errorf("expected ActionPlayGame, got %v", ctx.ActionType)
				}
			}

			// Check collider
			if _, ok := entity.GetComponent("collider"); !ok {
				t.Error("entity missing collider component")
			}

			// Check team
			if _, ok := entity.GetComponent("team"); !ok {
				t.Error("entity missing team component")
			}

			// Check sprite
			if _, ok := entity.GetComponent("sprite"); !ok {
				t.Error("entity missing sprite component")
			}

			// Check animation
			if _, ok := entity.GetComponent("animation"); !ok {
				t.Error("entity missing animation component")
			}

			// Verify minimum component count
			// We expect: position, velocity, minigameStation, contextAction, collider, team, sprite, animation
			hasMinRequired := true
			requiredComps := []string{"position", "velocity", "minigameStation", "contextAction", "collider", "team", "sprite", "animation"}
			for _, compType := range requiredComps {
				if _, ok := entity.GetComponent(compType); !ok {
					hasMinRequired = false
					break
				}
			}
			if !hasMinRequired {
				t.Error("entity missing required components")
			}
		})
	}
}

func TestSpawnMultipleStations(t *testing.T) {
	tests := []struct {
		name      string
		count     int
		seed      int64
		params    procgen.GenerationParams
		worldNil  bool
		wantErr   bool
		wantCount int
	}{
		{
			name:  "spawn 3 stations",
			count: 3,
			seed:  12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			worldNil:  false,
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:  "spawn 5 stations",
			count: 5,
			seed:  67890,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      3,
				GenreID:    "sci-fi",
			},
			worldNil:  false,
			wantErr:   false,
			wantCount: 5,
		},
		{
			name:  "zero count returns empty",
			count: 0,
			seed:  12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			worldNil:  false,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:  "negative count returns empty",
			count: -1,
			seed:  12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			worldNil:  false,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:  "nil world error",
			count: 3,
			seed:  12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			worldNil: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var world *World
			if !tt.worldNil {
				world = NewWorld()
			}

			stations, err := SpawnMultipleStations(world, 500.0, 500.0, tt.count, tt.seed, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(stations) != tt.wantCount {
				t.Errorf("expected %d stations, got %d", tt.wantCount, len(stations))
			}

			// Verify all stations are valid
			for i, station := range stations {
				if station == nil {
					t.Errorf("station %d is nil", i)
					continue
				}

				if _, ok := station.GetComponent("minigameStation"); !ok {
					t.Errorf("station %d missing minigameStation component", i)
				}
			}
		})
	}
}

func TestSelectGameType_Determinism(t *testing.T) {
	// Test that same seed produces same game type
	seed := int64(12345)
	genreID := "fantasy"

	rng1 := rand.New(rand.NewSource(seed))
	gameType1 := selectGameType(rng1, genreID)

	rng2 := rand.New(rand.NewSource(seed))
	gameType2 := selectGameType(rng2, genreID)

	if gameType1 != gameType2 {
		t.Errorf("same seed produced different game types: %v vs %v", gameType1, gameType2)
	}
}

func TestSelectGameType_GenreDistribution(t *testing.T) {
	// Test that different genres produce valid game types
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic", "unknown"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			rng := rand.New(rand.NewSource(12345))
			gameType := selectGameType(rng, genreID)

			// Should be a valid game type
			if gameType < 0 || gameType > MiniGameRitual {
				t.Errorf("invalid game type %v for genre %s", gameType, genreID)
			}

			// Should have a valid name
			name := gameType.String()
			if name == "Unknown" {
				t.Errorf("game type %v returned Unknown string", gameType)
			}
		})
	}
}

func TestCalculateStationDifficulty(t *testing.T) {
	tests := []struct {
		name           string
		depth          int
		baseDifficulty float64
		wantMin        float64
		wantMax        float64
	}{
		{
			name:           "depth 1 easy",
			depth:          1,
			baseDifficulty: 0.2,
			wantMin:        0.2,
			wantMax:        0.3,
		},
		{
			name:           "depth 5 medium",
			depth:          5,
			baseDifficulty: 0.5,
			wantMin:        0.7,
			wantMax:        0.8,
		},
		{
			name:           "depth 10 hard clamped",
			depth:          10,
			baseDifficulty: 0.7,
			wantMin:        1.0,
			wantMax:        1.0,
		},
		{
			name:           "depth 0 base only",
			depth:          0,
			baseDifficulty: 0.3,
			wantMin:        0.3,
			wantMax:        0.3,
		},
		{
			name:           "negative base clamped",
			depth:          1,
			baseDifficulty: -0.5,
			wantMin:        0.0,
			wantMax:        0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			difficulty := calculateStationDifficulty(tt.depth, tt.baseDifficulty)

			if difficulty < tt.wantMin || difficulty > tt.wantMax {
				t.Errorf("expected difficulty in range [%.1f, %.1f], got %.1f",
					tt.wantMin, tt.wantMax, difficulty)
			}

			// Always clamped to valid range
			if difficulty < 0.0 || difficulty > 1.0 {
				t.Errorf("difficulty %.1f outside valid range [0.0, 1.0]", difficulty)
			}
		})
	}
}

func TestGetMiniGameStationPrompt(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name        string
		setupFunc   func() *Entity
		wantContain string
	}{
		{
			name: "basic station prompt",
			setupFunc: func() *Entity {
				station, _ := SpawnMiniGameStation(world, 0, 0, 12345, procgen.GenerationParams{
					Difficulty: 0.5,
					Depth:      1,
					GenreID:    "fantasy",
				})
				return station
			},
			wantContain: "Play",
		},
		{
			name: "station with cost",
			setupFunc: func() *Entity {
				station, _ := SpawnMiniGameStation(world, 0, 0, 12345, procgen.GenerationParams{
					Difficulty: 0.5,
					Depth:      1,
					GenreID:    "fantasy",
				})
				stationCompRaw, _ := station.GetComponent("minigameStation")
				stationComp := stationCompRaw.(*MiniGameStationComponent)
				stationComp.EntryCost = 50
				return station
			},
			wantContain: "50g",
		},
		{
			name: "station with level requirement",
			setupFunc: func() *Entity {
				station, _ := SpawnMiniGameStation(world, 0, 0, 12345, procgen.GenerationParams{
					Difficulty: 0.5,
					Depth:      1,
					GenreID:    "fantasy",
				})
				stationCompRaw, _ := station.GetComponent("minigameStation")
				stationComp := stationCompRaw.(*MiniGameStationComponent)
				stationComp.RequiresLevel = 5
				return station
			},
			wantContain: "Lvl 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			station := tt.setupFunc()
			prompt := GetMiniGameStationPrompt(station)

			if prompt == "" {
				t.Error("expected non-empty prompt")
			}

			// Check for expected content
			if len(tt.wantContain) > 0 {
				found := false
				for i := 0; i <= len(prompt)-len(tt.wantContain); i++ {
					if prompt[i:i+len(tt.wantContain)] == tt.wantContain {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("prompt '%s' does not contain '%s'", prompt, tt.wantContain)
				}
			}
		})
	}
}
