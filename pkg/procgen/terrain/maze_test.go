//go:build test
// +build test

package terrain

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestMazeGenerator tests basic maze generation.
func TestMazeGenerator(t *testing.T) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  41, // Odd dimensions for maze
			"height": 41,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain, ok := result.(*Terrain)
	if !ok {
		t.Fatal("Result is not a Terrain")
	}

	// Validate dimensions
	if terrain.Width != 41 || terrain.Height != 41 {
		t.Errorf("Unexpected dimensions: %dx%d, want 41x41", terrain.Width, terrain.Height)
	}

	// Validate generation
	if err := gen.Validate(terrain); err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

// TestMazeGeneratorDeterminism verifies that the same seed produces the same maze.
func TestMazeGeneratorDeterminism(t *testing.T) {
	gen1 := NewMazeGenerator()
	gen2 := NewMazeGenerator()

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  31,
			"height": 31,
		},
	}

	seed := int64(98765)

	// Generate two mazes with the same seed
	result1, err1 := gen1.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	result2, err2 := gen2.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	terrain1 := result1.(*Terrain)
	terrain2 := result2.(*Terrain)

	// Compare dimensions
	if terrain1.Width != terrain2.Width || terrain1.Height != terrain2.Height {
		t.Errorf("Dimensions differ: %dx%d vs %dx%d",
			terrain1.Width, terrain1.Height, terrain2.Width, terrain2.Height)
	}

	// Compare all tiles
	differences := 0
	for y := 0; y < terrain1.Height; y++ {
		for x := 0; x < terrain1.Width; x++ {
			if terrain1.GetTile(x, y) != terrain2.GetTile(x, y) {
				differences++
			}
		}
	}

	if differences > 0 {
		t.Errorf("Mazes are not deterministic: %d tiles differ", differences)
	}

	// Compare room counts
	if len(terrain1.Rooms) != len(terrain2.Rooms) {
		t.Errorf("Room counts differ: %d vs %d", len(terrain1.Rooms), len(terrain2.Rooms))
	}

	// Compare stair counts
	if len(terrain1.StairsUp) != len(terrain2.StairsUp) {
		t.Errorf("Stairs up counts differ: %d vs %d", len(terrain1.StairsUp), len(terrain2.StairsUp))
	}
	if len(terrain1.StairsDown) != len(terrain2.StairsDown) {
		t.Errorf("Stairs down counts differ: %d vs %d", len(terrain1.StairsDown), len(terrain2.StairsDown))
	}
}

// TestMazeGenerator_InvalidDimensions tests error handling for invalid dimensions.
func TestMazeGenerator_InvalidDimensions(t *testing.T) {
	gen := NewMazeGenerator()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"negative width", -10, 50},
		{"negative height", 50, -10},
		{"zero width", 0, 50},
		{"zero height", 50, 0},
		{"too large width", 1001, 50},
		{"too large height", 50, 1001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := procgen.GenerationParams{
				Custom: map[string]interface{}{
					"width":  tt.width,
					"height": tt.height,
				},
			}

			_, err := gen.Generate(12345, params)
			if err == nil {
				t.Error("Expected error for invalid dimensions, got none")
			}
		})
	}
}

// TestMazeGenerator_Connectivity tests that all walkable tiles are reachable.
func TestMazeGenerator_Connectivity(t *testing.T) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  51,
			"height": 51,
		},
	}

	result, err := gen.Generate(55555, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Find a starting walkable tile
	var startX, startY int
	found := false
	for y := 0; y < terrain.Height && !found; y++ {
		for x := 0; x < terrain.Width && !found; x++ {
			if terrain.IsWalkable(x, y) {
				startX, startY = x, y
				found = true
			}
		}
	}

	if !found {
		t.Fatal("No walkable tiles found in maze")
	}

	// Flood fill from start to find all reachable tiles
	visited := make(map[Point]bool)
	var queue []Point
	queue = append(queue, Point{X: startX, Y: startY})
	visited[Point{X: startX, Y: startY}] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check all four neighbors
		neighbors := current.Neighbors()
		for _, n := range neighbors {
			if !terrain.IsInBounds(n.X, n.Y) {
				continue
			}
			if visited[n] {
				continue
			}
			if !terrain.IsWalkable(n.X, n.Y) {
				continue
			}

			visited[n] = true
			queue = append(queue, n)
		}
	}

	// Count total walkable tiles
	totalWalkable := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				totalWalkable++
			}
		}
	}

	reachable := len(visited)

	// All walkable tiles should be reachable
	if reachable != totalWalkable {
		t.Errorf("Not all tiles are connected: %d reachable out of %d walkable (%.1f%%)",
			reachable, totalWalkable, float64(reachable)/float64(totalWalkable)*100)
	}
}

// TestMazeGenerator_RoomGeneration tests that rooms are created at dead ends.
func TestMazeGenerator_RoomGeneration(t *testing.T) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":      41,
			"height":     41,
			"roomChance": 0.5, // 50% chance for testing
		},
	}

	result, err := gen.Generate(11111, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Should have created at least one room with 50% chance
	if len(terrain.Rooms) == 0 {
		t.Log("Warning: No rooms created (this can happen randomly, but is unlikely with 50% chance)")
	}

	// Verify each room is accessible
	for i, room := range terrain.Rooms {
		// Check that room center is walkable
		cx, cy := room.Center()
		if !terrain.IsWalkable(cx, cy) {
			t.Errorf("Room %d center (%d, %d) is not walkable", i, cx, cy)
		}
	}
}

// TestMazeGenerator_StairPlacement tests that stairs are placed in corners.
func TestMazeGenerator_StairPlacement(t *testing.T) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  41,
			"height": 41,
		},
	}

	result, err := gen.Generate(77777, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Should have stairs up and down
	if len(terrain.StairsUp) == 0 {
		t.Error("No stairs up placed")
	}
	if len(terrain.StairsDown) == 0 {
		t.Error("No stairs down placed")
	}

	// Verify stairs are in corner regions (within 10 tiles of edges)
	cornerSize := 10
	for _, stair := range terrain.StairsUp {
		inCorner := (stair.X < cornerSize || stair.X >= terrain.Width-cornerSize) &&
			(stair.Y < cornerSize || stair.Y >= terrain.Height-cornerSize)
		if !inCorner {
			t.Errorf("Stairs up at (%d, %d) not in corner region", stair.X, stair.Y)
		}
	}

	for _, stair := range terrain.StairsDown {
		inCorner := (stair.X < cornerSize || stair.X >= terrain.Width-cornerSize) &&
			(stair.Y < cornerSize || stair.Y >= terrain.Height-cornerSize)
		if !inCorner {
			t.Errorf("Stairs down at (%d, %d) not in corner region", stair.X, stair.Y)
		}
	}

	// Validate stair placement
	if err := terrain.ValidateStairPlacement(); err != nil {
		t.Errorf("Stair validation failed: %v", err)
	}
}

// TestMazeGenerator_CustomParameters tests custom parameter handling.
func TestMazeGenerator_CustomParameters(t *testing.T) {
	tests := []struct {
		name          string
		roomChance    float64
		corridorWidth int
	}{
		{"no rooms", 0.0, 1},
		{"many rooms", 0.9, 1},
		{"wide corridors", 0.1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewMazeGenerator()
			params := procgen.GenerationParams{
				Custom: map[string]interface{}{
					"width":         31,
					"height":        31,
					"roomChance":    tt.roomChance,
					"corridorWidth": tt.corridorWidth,
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			terrain := result.(*Terrain)

			if err := gen.Validate(terrain); err != nil {
				t.Errorf("Validation failed: %v", err)
			}

			// Check room generation matches expectation
			if tt.roomChance == 0.0 && len(terrain.Rooms) > 0 {
				t.Errorf("Expected no rooms with 0 room chance, got %d", len(terrain.Rooms))
			}
		})
	}
}

// TestMazeGenerator_EvenDimensionsAdjustment tests that even dimensions are adjusted to odd.
func TestMazeGenerator_EvenDimensionsAdjustment(t *testing.T) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  40, // Even
			"height": 50, // Even
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain := result.(*Terrain)

	// Dimensions should be adjusted to odd
	if terrain.Width%2 == 0 {
		t.Errorf("Width should be odd, got %d", terrain.Width)
	}
	if terrain.Height%2 == 0 {
		t.Errorf("Height should be odd, got %d", terrain.Height)
	}

	// Should be increased by 1
	if terrain.Width != 41 {
		t.Errorf("Width should be 41 (40+1), got %d", terrain.Width)
	}
	if terrain.Height != 51 {
		t.Errorf("Height should be 51 (50+1), got %d", terrain.Height)
	}
}

// TestMazeValidation_InvalidInput tests validation error handling.
func TestMazeValidation_InvalidInput(t *testing.T) {
	gen := NewMazeGenerator()

	// Test with non-terrain input
	err := gen.Validate("not a terrain")
	if err == nil {
		t.Error("Validate should fail for non-terrain input")
	}

	// Test with terrain with insufficient walkable tiles
	terrain := NewTerrain(10, 10, 12345)
	// Leave it mostly walls (only 1 floor tile = 1%)
	terrain.SetTile(5, 5, TileFloor)

	err = gen.Validate(terrain)
	if err == nil {
		t.Error("Validate should fail for terrain with insufficient walkable tiles")
	}
}

// TestMazeGenerator_SmallMaze tests generation of small mazes.
func TestMazeGenerator_SmallMaze(t *testing.T) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  11,
			"height": 11,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain := result.(*Terrain)

	if err := gen.Validate(terrain); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Even small mazes should have some walkable area
	walkable := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				walkable++
			}
		}
	}

	if walkable == 0 {
		t.Error("Small maze has no walkable tiles")
	}
}

// BenchmarkMazeGenerator benchmarks maze generation performance.
func BenchmarkMazeGenerator(b *testing.B) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  81,
			"height": 81,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

// BenchmarkMazeGenerator_Large benchmarks large maze generation.
func BenchmarkMazeGenerator_Large(b *testing.B) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  201,
			"height": 201,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}
