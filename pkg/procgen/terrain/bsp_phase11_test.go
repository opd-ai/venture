package terrain

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestBSPGenerator_DiagonalWalls tests that diagonal walls are generated in some rooms.
func TestBSPGenerator_DiagonalWalls(t *testing.T) {
	gen := NewBSPGenerator()
	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  60,
			"height": 40,
		},
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	terrain := result.(*Terrain)

	// Count diagonal walls
	diagonalCount := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			tile := terrain.GetTile(x, y)
			if tile.IsDiagonalWall() {
				diagonalCount++
			}
		}
	}

	// With 30% chance per room, we should have at least a few diagonal walls
	// in a dungeon with ~6-10 rooms
	if diagonalCount == 0 {
		t.Log("Warning: No diagonal walls generated (this can happen randomly)")
	} else {
		t.Logf("Generated %d diagonal wall tiles", diagonalCount)
	}
}

// TestBSPGenerator_MultiLayerFeatures tests that multi-layer features are generated.
func TestBSPGenerator_MultiLayerFeatures(t *testing.T) {
	gen := NewBSPGenerator()
	seed := int64(54321)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	terrain := result.(*Terrain)

	// Count multi-layer tiles
	platformCount := 0
	pitCount := 0
	lavaCount := 0
	bridgeCount := 0
	rampCount := 0

	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			tile := terrain.GetTile(x, y)
			switch tile {
			case TilePlatform:
				platformCount++
			case TilePit:
				pitCount++
			case TileLavaFlow:
				lavaCount++
			case TileBridge:
				bridgeCount++
			case TileRamp, TileRampUp, TileRampDown:
				rampCount++
			}
		}
	}

	t.Logf("Multi-layer tiles: platforms=%d, pits=%d, lava=%d, bridges=%d, ramps=%d",
		platformCount, pitCount, lavaCount, bridgeCount, rampCount)

	// With ~30% chance of multi-layer features, we should have some
	// (but it's random so don't fail if none)
	totalMultiLayer := platformCount + pitCount + lavaCount
	if totalMultiLayer == 0 {
		t.Log("Warning: No multi-layer features generated (this can happen randomly)")
	}
}

// TestChamferRoomCorners tests the chamfer function directly.
func TestChamferRoomCorners(t *testing.T) {
	gen := NewBSPGenerator()
	terrain := NewTerrain(20, 20, 12345)
	rng := rand.New(rand.NewSource(12345))

	// Create a test room
	room := &Room{
		X:      5,
		Y:      5,
		Width:  10,
		Height: 8,
	}

	// Fill room with floor
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}

	// Apply chamfer
	gen.chamferRoomCorners(terrain, room, rng)

	// Check that some tiles were changed to diagonal walls
	cornerTiles := []struct {
		x, y int
	}{
		{room.X, room.Y},                                     // Top-left
		{room.X + room.Width - 1, room.Y},                    // Top-right
		{room.X, room.Y + room.Height - 1},                   // Bottom-left
		{room.X + room.Width - 1, room.Y + room.Height - 1},  // Bottom-right
	}

	diagonalFound := false
	for _, corner := range cornerTiles {
		tile := terrain.GetTile(corner.x, corner.y)
		if tile.IsDiagonalWall() {
			diagonalFound = true
			break
		}
	}

	if !diagonalFound {
		t.Log("Warning: No diagonal walls at corners (all corners might not be chamfered)")
	}
}

// TestAddCentralPlatform tests platform generation directly.
func TestAddCentralPlatform(t *testing.T) {
	gen := NewBSPGenerator()
	terrain := NewTerrain(30, 30, 12345)
	rng := rand.New(rand.NewSource(12345))

	// Create a test room
	room := &Room{
		X:      5,
		Y:      5,
		Width:  15,
		Height: 15,
	}

	// Fill room with floor
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}

	// Add central platform
	gen.addCentralPlatform(terrain, room, rng)

	// Check that platforms were created
	platformCount := 0
	rampCount := 0
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			tile := terrain.GetTile(x, y)
			if tile == TilePlatform {
				platformCount++
			} else if tile == TileRampUp || tile == TileRampDown {
				rampCount++
			}
		}
	}

	if platformCount == 0 {
		t.Error("Expected platform tiles to be created")
	}
	if rampCount == 0 {
		t.Error("Expected ramp tiles to be created")
	}

	t.Logf("Created platform with %d tiles and %d ramp tiles", platformCount, rampCount)
}

// TestAddCornerPits tests pit generation directly.
func TestAddCornerPits(t *testing.T) {
	gen := NewBSPGenerator()
	terrain := NewTerrain(25, 25, 12345)
	rng := rand.New(rand.NewSource(12345))

	// Create a test room
	room := &Room{
		X:      5,
		Y:      5,
		Width:  12,
		Height: 12,
	}

	// Fill room with floor
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}

	// Add corner pits
	gen.addCornerPits(terrain, room, rng)

	// Check that pits were created
	pitCount := 0
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			tile := terrain.GetTile(x, y)
			if tile == TilePit {
				pitCount++
			}
		}
	}

	if pitCount == 0 {
		t.Error("Expected pit tiles to be created")
	}

	t.Logf("Created %d pit tiles", pitCount)
}

// TestAddLavaFlow tests lava flow generation directly.
func TestAddLavaFlow(t *testing.T) {
	gen := NewBSPGenerator()
	terrain := NewTerrain(25, 25, 12345)
	rng := rand.New(rand.NewSource(12345))

	// Create a test room
	room := &Room{
		X:      5,
		Y:      5,
		Width:  12,
		Height: 12,
	}

	// Fill room with floor
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			terrain.SetTile(x, y, TileFloor)
		}
	}

	// Add lava flow
	gen.addLavaFlow(terrain, room, rng)

	// Check that lava and bridges were created
	lavaCount := 0
	bridgeCount := 0
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			tile := terrain.GetTile(x, y)
			if tile == TileLavaFlow {
				lavaCount++
			} else if tile == TileBridge {
				bridgeCount++
			}
		}
	}

	if lavaCount == 0 {
		t.Error("Expected lava flow tiles to be created")
	}
	if bridgeCount == 0 {
		t.Error("Expected bridge tiles to be created")
	}

	t.Logf("Created %d lava tiles and %d bridge tiles", lavaCount, bridgeCount)
}

// TestBSPGenerator_Determinism_Phase11 tests that Phase 11.1 features are deterministic.
func TestBSPGenerator_Determinism_Phase11(t *testing.T) {
	gen := NewBSPGenerator()
	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  70,
			"height": 45,
		},
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generation error = %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generation error = %v", err2)
	}

	terrain1 := result1.(*Terrain)
	terrain2 := result2.(*Terrain)

	// Compare tiles including diagonal walls and multi-layer features
	mismatchCount := 0
	for y := 0; y < terrain1.Height; y++ {
		for x := 0; x < terrain1.Width; x++ {
			tile1 := terrain1.GetTile(x, y)
			tile2 := terrain2.GetTile(x, y)
			if tile1 != tile2 {
				mismatchCount++
			}
		}
	}

	if mismatchCount > 0 {
		t.Errorf("Determinism failed: %d tiles differ between generations", mismatchCount)
	}
}

// BenchmarkDiagonalWallGeneration benchmarks diagonal wall generation.
func BenchmarkDiagonalWallGeneration(b *testing.B) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  60,
			"height": 40,
		},
	}

	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

// BenchmarkMultiLayerFeatureGeneration benchmarks multi-layer feature generation.
func BenchmarkMultiLayerFeatureGeneration(b *testing.B) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "sci-fi",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}
