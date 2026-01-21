// Package terrain provides tests for cellular automata terrain generation.
package terrain

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// TestCellularGenerator_BasicGeneration verifies basic terrain generation.
func TestCellularGenerator_BasicGeneration(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	terrain, ok := result.(*Terrain)
	if !ok {
		t.Fatalf("Generate() did not return *Terrain")
	}

	if terrain.Width != 80 || terrain.Height != 50 {
		t.Errorf("Generate() dimensions = (%d, %d), want (80, 50)", terrain.Width, terrain.Height)
	}
}

// TestCellularGenerator_Determinism verifies same seed produces same terrain.
func TestCellularGenerator_Determinism(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	result1, _ := gen.Generate(99999, params)
	result2, _ := gen.Generate(99999, params)

	terrain1 := result1.(*Terrain)
	terrain2 := result2.(*Terrain)

	for y := 0; y < terrain1.Height; y++ {
		for x := 0; x < terrain1.Width; x++ {
			if terrain1.GetTile(x, y) != terrain2.GetTile(x, y) {
				t.Fatalf("Determinism failed at (%d, %d): %v != %v",
					x, y, terrain1.GetTile(x, y), terrain2.GetTile(x, y))
			}
		}
	}
}

// TestCellularGenerator_DifferentSeeds verifies different seeds produce different terrain.
func TestCellularGenerator_DifferentSeeds(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	result1, _ := gen.Generate(12345, params)
	result2, _ := gen.Generate(54321, params)

	terrain1 := result1.(*Terrain)
	terrain2 := result2.(*Terrain)

	differences := 0
	for y := 0; y < terrain1.Height; y++ {
		for x := 0; x < terrain1.Width; x++ {
			if terrain1.GetTile(x, y) != terrain2.GetTile(x, y) {
				differences++
			}
		}
	}

	// Should have some differences with different seeds
	if differences == 0 {
		t.Error("Different seeds produced identical terrain")
	}
}

// TestCellularGenerator_CustomParams verifies custom parameters are applied.
func TestCellularGenerator_CustomParams(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":           60,
			"height":          40,
			"fillProbability": 0.50,
			"iterations":      3,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	terrain := result.(*Terrain)
	if terrain.Width != 60 || terrain.Height != 40 {
		t.Errorf("Generate() dimensions = (%d, %d), want (60, 40)", terrain.Width, terrain.Height)
	}
}

// TestCellularGenerator_EdgeWalls verifies edges remain as walls.
func TestCellularGenerator_EdgeWalls(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  30,
			"height": 20,
		},
	}

	result, _ := gen.Generate(12345, params)
	terrain := result.(*Terrain)

	// Check all edges are walls
	for x := 0; x < terrain.Width; x++ {
		if terrain.GetTile(x, 0) != TileWall {
			t.Errorf("Top edge at x=%d is not a wall", x)
		}
		if terrain.GetTile(x, terrain.Height-1) != TileWall {
			t.Errorf("Bottom edge at x=%d is not a wall", x)
		}
	}
	for y := 0; y < terrain.Height; y++ {
		if terrain.GetTile(0, y) != TileWall {
			t.Errorf("Left edge at y=%d is not a wall", y)
		}
		if terrain.GetTile(terrain.Width-1, y) != TileWall {
			t.Errorf("Right edge at y=%d is not a wall", y)
		}
	}
}

// TestCellularGenerator_WithLogger verifies logging doesn't break generation.
func TestCellularGenerator_WithLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	gen := NewCellularGeneratorWithLogger(logger)

	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	terrain := result.(*Terrain)
	if terrain == nil {
		t.Fatal("Generate() returned nil terrain")
	}
}

// TestCellularGenerator_Validate verifies validation works correctly.
func TestCellularGenerator_Validate(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	result, _ := gen.Generate(12345, params)

	err := gen.Validate(result)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil for valid terrain", err)
	}
}

// TestCellularGenerator_Validate_InvalidType verifies validation rejects non-terrain.
func TestCellularGenerator_Validate_InvalidType(t *testing.T) {
	gen := NewCellularGenerator()

	err := gen.Validate("not a terrain")
	if err == nil {
		t.Error("Validate() should error on non-Terrain type")
	}
}

// TestCountWallNeighbors verifies neighbor counting works correctly.
func TestCountWallNeighbors(t *testing.T) {
	gen := NewCellularGenerator()
	terrain := NewTerrain(5, 5, 0)

	// Set center tile as floor, all others as wall (default)
	terrain.SetTile(2, 2, TileFloor)

	count := gen.countWallNeighbors(terrain, 2, 2)
	if count != 8 {
		t.Errorf("countWallNeighbors() = %d, want 8 (all neighbors are walls)", count)
	}

	// Clear some neighbors
	terrain.SetTile(1, 1, TileFloor)
	terrain.SetTile(2, 1, TileFloor)

	count = gen.countWallNeighbors(terrain, 2, 2)
	if count != 6 {
		t.Errorf("countWallNeighbors() = %d, want 6", count)
	}
}

// TestCountWallNeighborsFast verifies fast neighbor counting works correctly.
func TestCountWallNeighborsFast(t *testing.T) {
	gen := NewCellularGenerator()

	// Create row slices for testing
	rowAbove := []TileType{TileWall, TileWall, TileWall, TileWall, TileWall}
	row := []TileType{TileWall, TileWall, TileFloor, TileWall, TileWall}
	rowBelow := []TileType{TileWall, TileWall, TileWall, TileWall, TileWall}

	count := gen.countWallNeighborsFast(row, rowAbove, rowBelow, 2)
	if count != 8 {
		t.Errorf("countWallNeighborsFast() = %d, want 8", count)
	}

	// Clear some neighbors
	rowAbove[1] = TileFloor
	rowAbove[2] = TileFloor

	count = gen.countWallNeighborsFast(row, rowAbove, rowBelow, 2)
	if count != 6 {
		t.Errorf("countWallNeighborsFast() = %d, want 6", count)
	}
}

// TestNeighborOffsets verifies the pre-computed offsets are correct.
func TestNeighborOffsets(t *testing.T) {
	if len(neighborOffsets) != 8 {
		t.Errorf("neighborOffsets has %d entries, want 8", len(neighborOffsets))
	}

	// Verify no offset is (0,0)
	for _, offset := range neighborOffsets {
		if offset[0] == 0 && offset[1] == 0 {
			t.Error("neighborOffsets contains (0,0) which is the center cell")
		}
	}
}

// BenchmarkCountWallNeighbors benchmarks original neighbor counting.
func BenchmarkCountWallNeighbors(b *testing.B) {
	gen := NewCellularGenerator()
	terrain := NewTerrain(80, 50, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.countWallNeighbors(terrain, 40, 25)
	}
}

// BenchmarkCountWallNeighborsFast benchmarks optimized neighbor counting.
func BenchmarkCountWallNeighborsFast(b *testing.B) {
	gen := NewCellularGenerator()
	terrain := NewTerrain(80, 50, 12345)
	row := terrain.Tiles[25]
	rowAbove := terrain.Tiles[24]
	rowBelow := terrain.Tiles[26]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.countWallNeighborsFast(row, rowAbove, rowBelow, 40)
	}
}

// BenchmarkSimulateStep benchmarks a single cellular automata step.
func BenchmarkSimulateStep(b *testing.B) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	// Generate initial terrain
	result, _ := gen.Generate(12345, params)
	terrain := result.(*Terrain)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gen.simulateStep(terrain)
	}
}
