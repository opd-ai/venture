package terrain

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewTerrain(t *testing.T) {
	terrain := NewTerrain(10, 10, 12345)

	if terrain.Width != 10 {
		t.Errorf("Expected width 10, got %d", terrain.Width)
	}
	if terrain.Height != 10 {
		t.Errorf("Expected height 10, got %d", terrain.Height)
	}
	if terrain.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", terrain.Seed)
	}

	// All tiles should be walls initially
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) != TileWall {
				t.Errorf("Expected wall at (%d,%d)", x, y)
			}
		}
	}
}

func TestTileOperations(t *testing.T) {
	terrain := NewTerrain(5, 5, 12345)

	// Test SetTile and GetTile
	terrain.SetTile(2, 2, TileFloor)
	if terrain.GetTile(2, 2) != TileFloor {
		t.Error("SetTile/GetTile not working correctly")
	}

	// Test bounds checking
	terrain.SetTile(-1, -1, TileFloor) // Should not panic
	terrain.SetTile(10, 10, TileFloor) // Should not panic

	if terrain.GetTile(-1, -1) != TileWall {
		t.Error("Out of bounds should return wall")
	}
	if terrain.GetTile(10, 10) != TileWall {
		t.Error("Out of bounds should return wall")
	}
}

func TestIsWalkable(t *testing.T) {
	terrain := NewTerrain(5, 5, 12345)

	// Walls are not walkable
	if terrain.IsWalkable(0, 0) {
		t.Error("Walls should not be walkable")
	}

	// Floor is walkable
	terrain.SetTile(2, 2, TileFloor)
	if !terrain.IsWalkable(2, 2) {
		t.Error("Floor should be walkable")
	}

	// Corridors are walkable
	terrain.SetTile(3, 3, TileCorridor)
	if !terrain.IsWalkable(3, 3) {
		t.Error("Corridors should be walkable")
	}

	// Doors are walkable
	terrain.SetTile(4, 4, TileDoor)
	if !terrain.IsWalkable(4, 4) {
		t.Error("Doors should be walkable")
	}
}

func TestRoomCenter(t *testing.T) {
	room := &Room{X: 10, Y: 20, Width: 8, Height: 6}
	cx, cy := room.Center()

	if cx != 14 || cy != 23 {
		t.Errorf("Expected center (14, 23), got (%d, %d)", cx, cy)
	}
}

func TestRoomOverlaps(t *testing.T) {
	room1 := &Room{X: 5, Y: 5, Width: 10, Height: 10}
	room2 := &Room{X: 10, Y: 10, Width: 10, Height: 10}
	room3 := &Room{X: 20, Y: 20, Width: 10, Height: 10}

	if !room1.Overlaps(room2) {
		t.Error("room1 and room2 should overlap")
	}

	if room1.Overlaps(room3) {
		t.Error("room1 and room3 should not overlap")
	}
}

func TestTileType_String(t *testing.T) {
	tests := []struct {
		name     string
		tileType TileType
		expected string
	}{
		{"Wall", TileWall, "wall"},
		{"Floor", TileFloor, "floor"},
		{"Door", TileDoor, "door"},
		{"Corridor", TileCorridor, "corridor"},
		{"Shallow Water", TileWaterShallow, "shallow_water"},
		{"Deep Water", TileWaterDeep, "deep_water"},
		{"Tree", TileTree, "tree"},
		{"Stairs Up", TileStairsUp, "stairs_up"},
		{"Stairs Down", TileStairsDown, "stairs_down"},
		{"Trap Door", TileTrapDoor, "trap_door"},
		{"Secret Door", TileSecretDoor, "secret_door"},
		{"Bridge", TileBridge, "bridge"},
		{"Structure", TileStructure, "structure"},
		{"Unknown", TileType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tileType.String()
			if result != tt.expected {
				t.Errorf("TileType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBSPValidation_InvalidInput(t *testing.T) {
	gen := NewBSPGenerator()

	// Test with non-terrain input
	err := gen.Validate("not a terrain")
	if err == nil {
		t.Error("Validate should fail for non-terrain input")
	}

	// Test with terrain with no rooms
	emptyTerrain := NewTerrain(10, 10, 12345)
	err = gen.Validate(emptyTerrain)
	if err == nil {
		t.Error("Validate should fail for terrain with no rooms")
	}

	// Test with valid terrain
	emptyTerrain.Rooms = []*Room{{X: 1, Y: 1, Width: 5, Height: 5}}
	emptyTerrain.SetTile(2, 2, TileFloor)
	err = gen.Validate(emptyTerrain)
	if err != nil {
		t.Errorf("Validate should succeed for valid terrain: %v", err)
	}

	// Test with out-of-bounds room
	badTerrain := NewTerrain(10, 10, 12345)
	badTerrain.Rooms = []*Room{{X: 50, Y: 50, Width: 5, Height: 5}}
	err = gen.Validate(badTerrain)
	if err == nil {
		t.Error("Validate should fail for out-of-bounds room")
	}
}

func TestBSPGenerator_InvalidDimensions(t *testing.T) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"negative width", -10, 50},
		{"negative height", 80, -10},
		{"both negative", -10, -10},
		{"zero width", 0, 50},
		{"zero height", 80, 0},
		{"both zero", 0, 0},
		{"too large width", 20000, 50},
		{"too large height", 80, 20000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params.Custom = map[string]interface{}{
				"width":  tt.width,
				"height": tt.height,
			}
			_, err := gen.Generate(12345, params)
			if err == nil {
				t.Errorf("Generate should fail for invalid dimensions: width=%d, height=%d", tt.width, tt.height)
			}
		})
	}
}

func TestCellularValidation_InvalidInput(t *testing.T) {
	gen := NewCellularGenerator()

	// Test with non-terrain input
	err := gen.Validate("not a terrain")
	if err == nil {
		t.Error("Validate should fail for non-terrain input")
	}

	// Test with terrain with no walkable tiles
	wallTerrain := NewTerrain(10, 10, 12345)
	err = gen.Validate(wallTerrain)
	if err == nil {
		t.Error("Validate should fail for terrain with no walkable tiles")
	}

	// Test with valid terrain (needs at least 30% walkable tiles)
	for x := 2; x < 8; x++ {
		for y := 2; y < 8; y++ {
			wallTerrain.SetTile(x, y, TileFloor)
		}
	}
	err = gen.Validate(wallTerrain)
	if err != nil {
		t.Errorf("Validate should succeed for terrain with walkable tiles: %v", err)
	}
}

func TestCellularGenerator_InvalidDimensions(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "caves",
	}

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"negative width", -10, 50},
		{"negative height", 80, -10},
		{"both negative", -10, -10},
		{"zero width", 0, 50},
		{"zero height", 80, 0},
		{"both zero", 0, 0},
		{"too large width", 20000, 50},
		{"too large height", 80, 20000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params.Custom = map[string]interface{}{
				"width":  tt.width,
				"height": tt.height,
			}
			_, err := gen.Generate(12345, params)
			if err == nil {
				t.Errorf("Generate should fail for invalid dimensions: width=%d, height=%d", tt.width, tt.height)
			}
		})
	}
}

func TestBSPGenerator(t *testing.T) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
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

	// Validate the terrain
	if err := gen.Validate(terrain); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Check basic properties
	if terrain.Width != 40 || terrain.Height != 30 {
		t.Errorf("Expected size 40x30, got %dx%d", terrain.Width, terrain.Height)
	}

	if len(terrain.Rooms) == 0 {
		t.Error("No rooms generated")
	}

	// Check that rooms are carved out
	hasFloor := false
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.GetTile(x, y) == TileFloor {
				hasFloor = true
				break
			}
		}
	}
	if !hasFloor {
		t.Error("No floor tiles found")
	}
}

func TestBSPGeneratorDeterminism(t *testing.T) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	// Generate twice with the same seed
	result1, _ := gen.Generate(12345, params)
	result2, _ := gen.Generate(12345, params)

	terrain1 := result1.(*Terrain)
	terrain2 := result2.(*Terrain)

	// Check that they are identical
	if len(terrain1.Rooms) != len(terrain2.Rooms) {
		t.Error("Room count differs between generations")
	}

	// Check that tile layout is identical
	for y := 0; y < terrain1.Height; y++ {
		for x := 0; x < terrain1.Width; x++ {
			if terrain1.GetTile(x, y) != terrain2.GetTile(x, y) {
				t.Errorf("Tile at (%d,%d) differs", x, y)
			}
		}
	}
}

func TestCellularGenerator(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "caves",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
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

	// Validate the terrain
	if err := gen.Validate(terrain); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Check basic properties
	if terrain.Width != 40 || terrain.Height != 30 {
		t.Errorf("Expected size 40x30, got %dx%d", terrain.Width, terrain.Height)
	}

	// Check that we have walkable areas
	walkableCount := 0
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if terrain.IsWalkable(x, y) {
				walkableCount++
			}
		}
	}

	if walkableCount == 0 {
		t.Error("No walkable tiles generated")
	}

	totalTiles := terrain.Width * terrain.Height
	walkablePercent := float64(walkableCount) / float64(totalTiles)
	if walkablePercent < 0.3 {
		t.Errorf("Too few walkable tiles: %.1f%%", walkablePercent*100)
	}
}

func TestCellularGeneratorDeterminism(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "caves",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	// Generate twice with the same seed
	result1, _ := gen.Generate(12345, params)
	result2, _ := gen.Generate(12345, params)

	terrain1 := result1.(*Terrain)
	terrain2 := result2.(*Terrain)

	// Check that tile layout is identical
	for y := 0; y < terrain1.Height; y++ {
		for x := 0; x < terrain1.Width; x++ {
			if terrain1.GetTile(x, y) != terrain2.GetTile(x, y) {
				t.Errorf("Tile at (%d,%d) differs", x, y)
			}
		}
	}
}

func TestCellularGeneratorCustomParameters(t *testing.T) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "caves",
		Custom: map[string]interface{}{
			"width":           30,
			"height":          20,
			"fillProbability": 0.35,
			"iterations":      3,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	terrain := result.(*Terrain)
	if terrain.Width != 30 || terrain.Height != 20 {
		t.Errorf("Custom size not applied: got %dx%d", terrain.Width, terrain.Height)
	}
}

func BenchmarkBSPGenerator(b *testing.B) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}

func BenchmarkCellularGenerator(b *testing.B) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "caves",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}
