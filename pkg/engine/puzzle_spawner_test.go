// Package engine provides tests for puzzle spawning functionality.
package engine

import (
	"sort"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestSpawnPuzzlesInTerrain(t *testing.T) {
	tests := []struct {
		name        string
		targetCount int
		wantErr     bool
		minSpawned  int
	}{
		{
			name:        "spawn 5 puzzles",
			targetCount: 5,
			wantErr:     false,
			minSpawned:  1,
		},
		{
			name:        "spawn 10 puzzles",
			targetCount: 10,
			wantErr:     false,
			minSpawned:  1,
		},
		{
			name:        "zero target count",
			targetCount: 0,
			wantErr:     false,
			minSpawned:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create world
			world := NewWorld()

			// Generate terrain
			terrainGen := terrain.NewBSPGenerator()
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			}
			result, err := terrainGen.Generate(12345, params)
			if err != nil {
				t.Fatalf("failed to generate terrain: %v", err)
			}
			terrainData := result.(*terrain.Terrain)

			// Spawn puzzles
			count, err := SpawnPuzzlesInTerrain(world, terrainData, 54321, params, tt.targetCount)

			// Process entity additions
			world.Update(0)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("SpawnPuzzlesInTerrain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check puzzle count
			if count < tt.minSpawned {
				t.Errorf("SpawnPuzzlesInTerrain() spawned %d puzzles, want at least %d", count, tt.minSpawned)
			}

			// Verify puzzle entities exist
			puzzleEntities := world.GetEntitiesWith("puzzle")
			if len(puzzleEntities) != count {
				t.Errorf("Expected %d puzzle entities, got %d", count, len(puzzleEntities))
			}

			// Verify each puzzle has elements
			for _, entity := range puzzleEntities {
				comp, ok := entity.GetComponent("puzzle")
				if !ok {
					t.Errorf("Puzzle entity missing puzzle component")
					continue
				}

				puzzleComp := comp.(*PuzzleComponent)
				if len(puzzleComp.ElementIDs) == 0 {
					t.Errorf("Puzzle %s has no element IDs", puzzleComp.PuzzleID)
				}

				// Verify element entities exist
				for _, elemID := range puzzleComp.ElementIDs {
					elemEntity, exists := world.GetEntity(elemID)
					if !exists || elemEntity == nil {
						t.Errorf("Element entity %d not found", elemID)
						continue
					}

					// Verify element has required components
					if _, ok := elemEntity.GetComponent("position"); !ok {
						t.Errorf("Element entity %d missing position component", elemID)
					}
					if _, ok := elemEntity.GetComponent("puzzleElement"); !ok {
						t.Errorf("Element entity %d missing puzzleElement component", elemID)
					}
				}
			}
		})
	}
}

func TestSpawnPuzzlesInTerrain_NilInputs(t *testing.T) {
	world := NewWorld()

	// Generate terrain
	terrainGen := terrain.NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}
	result, _ := terrainGen.Generate(12345, params)
	terrainData := result.(*terrain.Terrain)

	tests := []struct {
		name    string
		world   *World
		terrain *terrain.Terrain
		wantErr bool
	}{
		{
			name:    "nil world",
			world:   nil,
			terrain: terrainData,
			wantErr: true,
		},
		{
			name:    "nil terrain",
			world:   world,
			terrain: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			}
			_, err := SpawnPuzzlesInTerrain(tt.world, tt.terrain, 12345, params, 5)
			if (err != nil) != tt.wantErr {
				t.Errorf("SpawnPuzzlesInTerrain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSpawnPuzzlesInTerrain_Determinism(t *testing.T) {
	// Create two worlds
	world1 := NewWorld()
	world2 := NewWorld()

	// Generate terrain
	terrainGen := terrain.NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}
	result, err := terrainGen.Generate(12345, params)
	if err != nil {
		t.Fatalf("failed to generate terrain: %v", err)
	}
	terrainData := result.(*terrain.Terrain)

	// Spawn puzzles with same seed in both worlds
	seed := int64(98765)
	count1, err1 := SpawnPuzzlesInTerrain(world1, terrainData, seed, params, 5)
	count2, err2 := SpawnPuzzlesInTerrain(world2, terrainData, seed, params, 5)

	// Process entity additions
	world1.Update(0)
	world2.Update(0)

	// Both should succeed
	if err1 != nil || err2 != nil {
		t.Fatalf("Spawning failed: err1=%v, err2=%v", err1, err2)
	}

	// Should spawn same number of puzzles
	if count1 != count2 {
		t.Errorf("Determinism failed: count1=%d, count2=%d", count1, count2)
	}

	// Verify puzzles have same IDs and types
	puzzles1 := world1.GetEntitiesWith("puzzle")
	puzzles2 := world2.GetEntitiesWith("puzzle")

	if len(puzzles1) != len(puzzles2) {
		t.Errorf("Different puzzle counts: %d vs %d", len(puzzles1), len(puzzles2))
		return
	}

	// Sort puzzles by PuzzleID for deterministic comparison
	sort.Slice(puzzles1, func(i, j int) bool {
		comp1, _ := puzzles1[i].GetComponent("puzzle")
		comp2, _ := puzzles1[j].GetComponent("puzzle")
		puz1 := comp1.(*PuzzleComponent)
		puz2 := comp2.(*PuzzleComponent)
		return puz1.PuzzleID < puz2.PuzzleID
	})
	sort.Slice(puzzles2, func(i, j int) bool {
		comp1, _ := puzzles2[i].GetComponent("puzzle")
		comp2, _ := puzzles2[j].GetComponent("puzzle")
		puz1 := comp1.(*PuzzleComponent)
		puz2 := comp2.(*PuzzleComponent)
		return puz1.PuzzleID < puz2.PuzzleID
	})

	for i := 0; i < len(puzzles1); i++ {
		comp1, _ := puzzles1[i].GetComponent("puzzle")
		comp2, _ := puzzles2[i].GetComponent("puzzle")

		puz1 := comp1.(*PuzzleComponent)
		puz2 := comp2.(*PuzzleComponent)

		if puz1.PuzzleType != puz2.PuzzleType {
			t.Errorf("Puzzle %d type mismatch: %s vs %s", i, puz1.PuzzleType, puz2.PuzzleType)
		}

		if puz1.Difficulty != puz2.Difficulty {
			t.Errorf("Puzzle %d difficulty mismatch: %d vs %d", i, puz1.Difficulty, puz2.Difficulty)
		}

		if len(puz1.ElementIDs) != len(puz2.ElementIDs) {
			t.Errorf("Puzzle %d element count mismatch: %d vs %d", i, len(puz1.ElementIDs), len(puz2.ElementIDs))
		}
	}
}
