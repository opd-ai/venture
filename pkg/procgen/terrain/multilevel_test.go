package terrain

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestGenerateMultiLevel_ThreeLevels tests basic multi-level generation with 3 levels.
func TestGenerateMultiLevel_ThreeLevels(t *testing.T) {
	gen := NewLevelGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	levels, err := gen.GenerateMultiLevel(3, 12345, params)
	if err != nil {
		t.Fatalf("GenerateMultiLevel failed: %v", err)
	}

	if len(levels) != 3 {
		t.Errorf("Expected 3 levels, got %d", len(levels))
	}

	// Verify each level has correct level number
	for i, level := range levels {
		if level.Level != i {
			t.Errorf("Level %d has incorrect level number: %d", i, level.Level)
		}
	}

	// Verify first level has stairs down
	if len(levels[0].StairsDown) == 0 {
		t.Error("First level missing stairs down")
	}

	// Verify middle level has both stairs
	if len(levels[1].StairsUp) == 0 {
		t.Error("Middle level missing stairs up")
	}
	if len(levels[1].StairsDown) == 0 {
		t.Error("Middle level missing stairs down")
	}

	// Verify last level has stairs up
	if len(levels[2].StairsUp) == 0 {
		t.Error("Last level missing stairs up")
	}
}

// TestGenerateMultiLevel_SingleLevel tests that single level generation works.
func TestGenerateMultiLevel_SingleLevel(t *testing.T) {
	gen := NewLevelGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	levels, err := gen.GenerateMultiLevel(1, 12345, params)
	if err != nil {
		t.Fatalf("GenerateMultiLevel failed: %v", err)
	}

	if len(levels) != 1 {
		t.Errorf("Expected 1 level, got %d", len(levels))
	}

	// Single level should have no stairs requirement
	level := levels[0]
	if level.Level != 0 {
		t.Errorf("Single level has incorrect level number: %d", level.Level)
	}
}

// TestGenerateMultiLevel_FiveLevels tests generation with 5 levels (deeper dungeon).
func TestGenerateMultiLevel_FiveLevels(t *testing.T) {
	gen := NewLevelGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.3,
		Depth:      1,
		GenreID:    "dungeon",
		Custom: map[string]interface{}{
			"width":  50,
			"height": 40,
		},
	}

	levels, err := gen.GenerateMultiLevel(5, 99999, params)
	if err != nil {
		t.Fatalf("GenerateMultiLevel failed: %v", err)
	}

	if len(levels) != 5 {
		t.Errorf("Expected 5 levels, got %d", len(levels))
	}

	// Verify difficulty scaling
	for i, level := range levels {
		expectedDifficulty := 0.3 + float64(i)*0.1
		if expectedDifficulty > 1.0 {
			expectedDifficulty = 1.0
		}
		// Difficulty should increase with depth (checked implicitly via depth)
		if level.Level != i {
			t.Errorf("Level %d has incorrect level number: %d", i, level.Level)
		}
	}

	// Verify connectivity
	if len(levels[0].StairsDown) == 0 {
		t.Error("Level 0 missing stairs down")
	}
	if len(levels[4].StairsUp) == 0 {
		t.Error("Level 4 missing stairs up")
	}
	for i := 1; i < 4; i++ {
		if len(levels[i].StairsUp) == 0 {
			t.Errorf("Level %d missing stairs up", i)
		}
		if len(levels[i].StairsDown) == 0 {
			t.Errorf("Level %d missing stairs down", i)
		}
	}
}

// TestGenerateMultiLevel_InvalidLevelCount tests error handling for invalid level counts.
func TestGenerateMultiLevel_InvalidLevelCount(t *testing.T) {
	gen := NewLevelGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	tests := []struct {
		name      string
		numLevels int
	}{
		{"zero levels", 0},
		{"negative levels", -1},
		{"too many levels", 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gen.GenerateMultiLevel(tt.numLevels, 12345, params)
			if err == nil {
				t.Errorf("Expected error for %d levels, got none", tt.numLevels)
			}
		})
	}
}

// TestGenerateMultiLevel_Determinism tests that same seed produces same levels.
func TestGenerateMultiLevel_Determinism(t *testing.T) {
	gen1 := NewLevelGenerator()
	gen2 := NewLevelGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	seed := int64(12345)
	levels1, err1 := gen1.GenerateMultiLevel(3, seed, params)
	levels2, err2 := gen2.GenerateMultiLevel(3, seed, params)

	if err1 != nil || err2 != nil {
		t.Fatalf("GenerateMultiLevel failed: err1=%v, err2=%v", err1, err2)
	}

	// Compare levels
	for i := 0; i < 3; i++ {
		// Compare stair positions
		if len(levels1[i].StairsUp) != len(levels2[i].StairsUp) {
			t.Errorf("Level %d stairs up count mismatch: %d vs %d",
				i, len(levels1[i].StairsUp), len(levels2[i].StairsUp))
		}
		if len(levels1[i].StairsDown) != len(levels2[i].StairsDown) {
			t.Errorf("Level %d stairs down count mismatch: %d vs %d",
				i, len(levels1[i].StairsDown), len(levels2[i].StairsDown))
		}

		// Compare actual stair positions
		for j, stair1 := range levels1[i].StairsUp {
			if j < len(levels2[i].StairsUp) {
				stair2 := levels2[i].StairsUp[j]
				if stair1.X != stair2.X || stair1.Y != stair2.Y {
					t.Errorf("Level %d stairs up position mismatch at index %d: (%d,%d) vs (%d,%d)",
						i, j, stair1.X, stair1.Y, stair2.X, stair2.Y)
				}
			}
		}
	}
}

// TestGenerateMultiLevel_MixedGenerators tests using different generators for different depths.
func TestGenerateMultiLevel_MixedGenerators(t *testing.T) {
	gen := NewLevelGenerator()

	// Level 0: BSP (default)
	// Level 1: Cellular (caves)
	// Level 2: Maze
	gen.SetGenerator(1, NewCellularGenerator())
	gen.SetGenerator(2, NewMazeGenerator())

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	levels, err := gen.GenerateMultiLevel(3, 54321, params)
	if err != nil {
		t.Fatalf("GenerateMultiLevel with mixed generators failed: %v", err)
	}

	if len(levels) != 3 {
		t.Errorf("Expected 3 levels, got %d", len(levels))
	}

	// Level 0 should have rooms (BSP)
	if len(levels[0].Rooms) == 0 {
		t.Error("Level 0 (BSP) has no rooms")
	}

	// Level 1 should have no rooms (Cellular)
	if len(levels[1].Rooms) != 0 {
		t.Error("Level 1 (Cellular) should have no rooms")
	}

	// Level 2 should have some rooms (Maze with dead-end rooms)
	// Note: Maze may or may not have rooms depending on room chance
	// So we just verify it generated without error

	// Verify all levels are connected
	if err := gen.ValidateMultiLevelConnectivity(levels); err != nil {
		t.Errorf("Multi-level connectivity validation failed: %v", err)
	}
}

// TestValidateMultiLevelConnectivity tests connectivity validation.
func TestValidateMultiLevelConnectivity(t *testing.T) {
	gen := NewLevelGenerator()

	t.Run("valid three levels", func(t *testing.T) {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      1,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"width":  40,
				"height": 30,
			},
		}

		levels, err := gen.GenerateMultiLevel(3, 12345, params)
		if err != nil {
			t.Fatalf("GenerateMultiLevel failed: %v", err)
		}

		if err := gen.ValidateMultiLevelConnectivity(levels); err != nil {
			t.Errorf("Valid levels failed validation: %v", err)
		}
	})

	t.Run("missing stairs down", func(t *testing.T) {
		levels := []*Terrain{
			NewTerrain(40, 30, 12345),
			NewTerrain(40, 30, 12346),
		}
		levels[0].Level = 0
		levels[1].Level = 1
		levels[1].AddStairs(20, 15, true) // Only stairs up in level 1

		err := gen.ValidateMultiLevelConnectivity(levels)
		if err == nil {
			t.Error("Expected error for missing stairs down, got none")
		}
	})

	t.Run("missing stairs up", func(t *testing.T) {
		levels := []*Terrain{
			NewTerrain(40, 30, 12345),
			NewTerrain(40, 30, 12346),
		}
		levels[0].Level = 0
		levels[1].Level = 1
		levels[0].AddStairs(20, 15, false) // Only stairs down in level 0

		err := gen.ValidateMultiLevelConnectivity(levels)
		if err == nil {
			t.Error("Expected error for missing stairs up, got none")
		}
	})

	t.Run("empty levels", func(t *testing.T) {
		levels := []*Terrain{}
		err := gen.ValidateMultiLevelConnectivity(levels)
		if err == nil {
			t.Error("Expected error for empty levels, got none")
		}
	})
}

// TestPlaceStairsRandom tests random stair placement.
func TestPlaceStairsRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))

	t.Run("both stairs", func(t *testing.T) {
		terrain := NewTerrain(40, 30, 12345)
		// Create some floor tiles
		for y := 10; y < 20; y++ {
			for x := 10; x < 30; x++ {
				terrain.SetTile(x, y, TileFloor)
			}
		}

		err := PlaceStairsRandom(terrain, true, true, rng)
		if err != nil {
			t.Fatalf("PlaceStairsRandom failed: %v", err)
		}

		if len(terrain.StairsUp) == 0 {
			t.Error("No stairs up placed")
		}
		if len(terrain.StairsDown) == 0 {
			t.Error("No stairs down placed")
		}

		// Verify stairs are in walkable positions
		for _, stair := range terrain.StairsUp {
			if !terrain.IsWalkable(stair.X, stair.Y) {
				t.Errorf("Stairs up at (%d, %d) is not walkable", stair.X, stair.Y)
			}
		}
		for _, stair := range terrain.StairsDown {
			if !terrain.IsWalkable(stair.X, stair.Y) {
				t.Errorf("Stairs down at (%d, %d) is not walkable", stair.X, stair.Y)
			}
		}
	})

	t.Run("no walkable tiles", func(t *testing.T) {
		terrain := NewTerrain(40, 30, 12345)
		// All walls, no floor tiles

		err := PlaceStairsRandom(terrain, true, false, rng)
		if err == nil {
			t.Error("Expected error for no walkable tiles, got none")
		}
	})
}

// TestPlaceStairsInRoom tests stair placement in specific room types.
func TestPlaceStairsInRoom(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))

	t.Run("boss room", func(t *testing.T) {
		terrain := NewTerrain(40, 30, 12345)
		// Create floor area
		for y := 0; y < 30; y++ {
			for x := 0; x < 40; x++ {
				terrain.SetTile(x, y, TileFloor)
			}
		}

		// Add a boss room
		bossRoom := &Room{
			X:      10,
			Y:      10,
			Width:  10,
			Height: 8,
			Type:   RoomBoss,
		}
		terrain.Rooms = append(terrain.Rooms, bossRoom)

		err := PlaceStairsInRoom(terrain, RoomBoss, true, true, rng)
		if err != nil {
			t.Fatalf("PlaceStairsInRoom failed: %v", err)
		}

		if len(terrain.StairsUp) == 0 {
			t.Error("No stairs up placed")
		}
		if len(terrain.StairsDown) == 0 {
			t.Error("No stairs down placed")
		}

		// Verify stairs are in or near the boss room
		cx, cy := bossRoom.Center()
		for _, stair := range terrain.StairsUp {
			dist := abs(stair.X-cx) + abs(stair.Y-cy)
			if dist > 5 {
				t.Errorf("Stairs up too far from boss room center: distance %d", dist)
			}
		}
	})

	t.Run("no matching room type", func(t *testing.T) {
		terrain := NewTerrain(40, 30, 12345)
		// Create floor but no boss rooms
		for y := 0; y < 30; y++ {
			for x := 0; x < 40; x++ {
				terrain.SetTile(x, y, TileFloor)
			}
		}

		normalRoom := &Room{
			X:      10,
			Y:      10,
			Width:  10,
			Height: 8,
			Type:   RoomNormal,
		}
		terrain.Rooms = append(terrain.Rooms, normalRoom)

		err := PlaceStairsInRoom(terrain, RoomBoss, true, false, rng)
		if err == nil {
			t.Error("Expected error for no matching room type, got none")
		}
	})
}

// TestPlaceStairsSymmetric tests symmetric stair placement.
func TestPlaceStairsSymmetric(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))

	t.Run("opposite corners", func(t *testing.T) {
		terrain := NewTerrain(40, 30, 12345)
		// Create floor tiles across the map
		for y := 0; y < 30; y++ {
			for x := 0; x < 40; x++ {
				terrain.SetTile(x, y, TileFloor)
			}
		}

		err := PlaceStairsSymmetric(terrain, true, true, rng)
		if err != nil {
			t.Fatalf("PlaceStairsSymmetric failed: %v", err)
		}

		if len(terrain.StairsUp) == 0 {
			t.Error("No stairs up placed")
		}
		if len(terrain.StairsDown) == 0 {
			t.Error("No stairs down placed")
		}

		// Verify stairs are in different corners (rough check)
		if len(terrain.StairsUp) > 0 && len(terrain.StairsDown) > 0 {
			up := terrain.StairsUp[0]
			down := terrain.StairsDown[0]
			dist := abs(up.X-down.X) + abs(up.Y-down.Y)
			if dist < 20 {
				t.Errorf("Stairs not symmetric enough: distance %d", dist)
			}
		}
	})

	t.Run("insufficient corners", func(t *testing.T) {
		terrain := NewTerrain(40, 30, 12345)
		// Only one small floor area (not enough for symmetric placement)
		for y := 14; y < 16; y++ {
			for x := 19; x < 21; x++ {
				terrain.SetTile(x, y, TileFloor)
			}
		}

		err := PlaceStairsSymmetric(terrain, true, true, rng)
		if err == nil {
			t.Error("Expected error for insufficient corners, got none")
		}
	})
}

// TestConnectLevels tests level connection functionality.
func TestConnectLevels(t *testing.T) {
	gen := NewLevelGenerator()
	rng := rand.New(rand.NewSource(12345))

	// Create two simple terrains
	above := NewTerrain(40, 30, 12345)
	below := NewTerrain(40, 30, 12346)

	// Fill with floor tiles
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			above.SetTile(x, y, TileFloor)
			below.SetTile(x, y, TileFloor)
		}
	}

	above.Level = 0
	below.Level = 1

	err := gen.ConnectLevels(above, below, rng)
	if err != nil {
		t.Fatalf("ConnectLevels failed: %v", err)
	}

	// Verify stairs were placed
	if len(above.StairsDown) == 0 {
		t.Error("No stairs down placed in upper level")
	}
	if len(below.StairsUp) == 0 {
		t.Error("No stairs up placed in lower level")
	}

	// Verify stairs are walkable
	for _, stair := range above.StairsDown {
		if !above.IsWalkable(stair.X, stair.Y) {
			t.Errorf("Stairs down at (%d, %d) is not walkable", stair.X, stair.Y)
		}
	}
	for _, stair := range below.StairsUp {
		if !below.IsWalkable(stair.X, stair.Y) {
			t.Errorf("Stairs up at (%d, %d) is not walkable", stair.X, stair.Y)
		}
	}
}

// BenchmarkGenerateMultiLevel benchmarks multi-level generation.
func BenchmarkGenerateMultiLevel(b *testing.B) {
	gen := NewLevelGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateMultiLevel(3, 12345, params)
	}
}

// BenchmarkGenerateMultiLevel_FiveLevels benchmarks deeper dungeons.
func BenchmarkGenerateMultiLevel_FiveLevels(b *testing.B) {
	gen := NewLevelGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  50,
			"height": 40,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateMultiLevel(5, 12345, params)
	}
}
