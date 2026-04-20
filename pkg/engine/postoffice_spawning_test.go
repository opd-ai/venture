package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestPostOfficeSpawner_SpawnInCity(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	// Create a simple city terrain
	cityTerrain := terrain.NewTerrain(50, 50, 12345)
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			cityTerrain.SetTile(x, y, terrain.TileFloor)
		}
	}

	// Add streets
	for x := 0; x < 50; x++ {
		cityTerrain.SetTile(x, 10, terrain.TileCorridor)
		cityTerrain.SetTile(x, 30, terrain.TileCorridor)
	}
	for y := 0; y < 50; y++ {
		cityTerrain.SetTile(10, y, terrain.TileCorridor)
		cityTerrain.SetTile(30, y, terrain.TileCorridor)
	}

	// Create building blocks
	blocks := []*terrain.CityBlock{
		{
			Rect:      terrain.Rect{X: 1, Y: 1, Width: 8, Height: 8},
			BlockType: terrain.BlockBuilding,
		},
		{
			Rect:      terrain.Rect{X: 11, Y: 11, Width: 18, Height: 18},
			BlockType: terrain.BlockBuilding,
		},
		{
			Rect:      terrain.Rect{X: 31, Y: 31, Width: 15, Height: 15},
			BlockType: terrain.BlockBuilding,
		},
	}

	result, err := spawner.SpawnInCity(cityTerrain, blocks, "fantasy", 12345)
	if err != nil {
		t.Fatalf("SpawnInCity failed: %v", err)
	}

	// Process entity creation
	world.Update(0)

	if result.BuildingID == 0 {
		t.Error("BuildingID is zero")
	}

	if result.ClerkID == 0 {
		t.Error("ClerkID is zero")
	}

	if result.ClerkName == "" {
		t.Error("ClerkName is empty")
	}

	// Verify entities exist in world
	building, exists := world.GetEntity(result.BuildingID)
	if !exists {
		t.Error("Building entity not found in world")
	} else {
		// Verify building components
		if _, ok := building.GetComponent("position"); !ok {
			t.Error("Building missing position component")
		}
		if _, ok := building.GetComponent("postoffice"); !ok {
			t.Error("Building missing postoffice component")
		}
	}

	clerk, exists := world.GetEntity(result.ClerkID)
	if !exists {
		t.Error("Clerk entity not found in world")
	} else {
		// Verify clerk components
		if _, ok := clerk.GetComponent("position"); !ok {
			t.Error("Clerk missing position component")
		}
		if _, ok := clerk.GetComponent("postoffice_clerk"); !ok {
			t.Error("Clerk missing postoffice_clerk component")
		}
		if _, ok := clerk.GetComponent("ai"); !ok {
			t.Error("Clerk missing AI component")
		}
	}
}

func TestPostOfficeSpawner_SpawnInCity_NilTerrain(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	blocks := []*terrain.CityBlock{
		{Rect: terrain.Rect{X: 1, Y: 1, Width: 10, Height: 10}, BlockType: terrain.BlockBuilding},
	}

	_, err := spawner.SpawnInCity(nil, blocks, "fantasy", 12345)
	if err == nil {
		t.Error("Expected error for nil terrain")
	}
}

func TestPostOfficeSpawner_SpawnInCity_EmptyBlocks(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	cityTerrain := terrain.NewTerrain(50, 50, 12345)

	_, err := spawner.SpawnInCity(cityTerrain, []*terrain.CityBlock{}, "fantasy", 12345)
	if err == nil {
		t.Error("Expected error for empty blocks")
	}
}

func TestPostOfficeSpawner_SpawnInCity_NoSuitableBlocks(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	cityTerrain := terrain.NewTerrain(50, 50, 12345)

	// Only small blocks (area < 64)
	blocks := []*terrain.CityBlock{
		{Rect: terrain.Rect{X: 1, Y: 1, Width: 5, Height: 5}, BlockType: terrain.BlockBuilding},
	}

	_, err := spawner.SpawnInCity(cityTerrain, blocks, "fantasy", 12345)
	if err == nil {
		t.Error("Expected error for no suitable blocks")
	}
}

func TestPostOfficeSpawner_FindSuitableBlocks(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	cityTerrain := terrain.NewTerrain(50, 50, 12345)
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			cityTerrain.SetTile(x, y, terrain.TileFloor)
		}
	}

	// Add streets
	for x := 0; x < 50; x++ {
		cityTerrain.SetTile(x, 10, terrain.TileCorridor)
	}

	tests := []struct {
		name     string
		blocks   []*terrain.CityBlock
		expected int
	}{
		{
			name: "large building with street access",
			blocks: []*terrain.CityBlock{
				{
					Rect:      terrain.Rect{X: 1, Y: 11, Width: 10, Height: 10},
					BlockType: terrain.BlockBuilding,
				},
			},
			expected: 1,
		},
		{
			name: "small building",
			blocks: []*terrain.CityBlock{
				{
					Rect:      terrain.Rect{X: 1, Y: 11, Width: 5, Height: 5},
					BlockType: terrain.BlockBuilding,
				},
			},
			expected: 0,
		},
		{
			name: "plaza",
			blocks: []*terrain.CityBlock{
				{
					Rect:      terrain.Rect{X: 1, Y: 11, Width: 10, Height: 10},
					BlockType: terrain.BlockPlaza,
				},
			},
			expected: 0,
		},
		{
			name: "no street access",
			blocks: []*terrain.CityBlock{
				{
					Rect:      terrain.Rect{X: 20, Y: 20, Width: 10, Height: 10},
					BlockType: terrain.BlockBuilding,
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suitable := spawner.findSuitableBlocks(tt.blocks, cityTerrain)
			if len(suitable) != tt.expected {
				t.Errorf("Expected %d suitable blocks, got %d", tt.expected, len(suitable))
			}
		})
	}
}

func TestPostOfficeSpawner_HasStreetAccess(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	cityTerrain := terrain.NewTerrain(50, 50, 12345)
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			cityTerrain.SetTile(x, y, terrain.TileFloor)
		}
	}

	tests := []struct {
		name     string
		block    *terrain.CityBlock
		streets  []struct{ X, Y int }
		expected bool
	}{
		{
			name: "street on top",
			block: &terrain.CityBlock{
				Rect:      terrain.Rect{X: 5, Y: 5, Width: 10, Height: 10},
				BlockType: terrain.BlockBuilding,
			},
			streets:  []struct{ X, Y int }{{X: 7, Y: 4}},
			expected: true,
		},
		{
			name: "street on bottom",
			block: &terrain.CityBlock{
				Rect:      terrain.Rect{X: 5, Y: 5, Width: 10, Height: 10},
				BlockType: terrain.BlockBuilding,
			},
			streets:  []struct{ X, Y int }{{X: 7, Y: 15}},
			expected: true,
		},
		{
			name: "street on left",
			block: &terrain.CityBlock{
				Rect:      terrain.Rect{X: 5, Y: 5, Width: 10, Height: 10},
				BlockType: terrain.BlockBuilding,
			},
			streets:  []struct{ X, Y int }{{X: 4, Y: 7}},
			expected: true,
		},
		{
			name: "street on right",
			block: &terrain.CityBlock{
				Rect:      terrain.Rect{X: 5, Y: 5, Width: 10, Height: 10},
				BlockType: terrain.BlockBuilding,
			},
			streets:  []struct{ X, Y int }{{X: 15, Y: 7}},
			expected: true,
		},
		{
			name: "no street access",
			block: &terrain.CityBlock{
				Rect:      terrain.Rect{X: 5, Y: 5, Width: 10, Height: 10},
				BlockType: terrain.BlockBuilding,
			},
			streets:  []struct{ X, Y int }{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear terrain
			for y := 0; y < 50; y++ {
				for x := 0; x < 50; x++ {
					cityTerrain.SetTile(x, y, terrain.TileFloor)
				}
			}

			// Add streets
			for _, pos := range tt.streets {
				cityTerrain.SetTile(pos.X, pos.Y, terrain.TileCorridor)
			}

			result := spawner.hasStreetAccess(tt.block, cityTerrain)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPostOfficeSpawner_SelectCentralBlock(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	cityTerrain := terrain.NewTerrain(50, 50, 12345)

	blocks := []*terrain.CityBlock{
		{
			Rect:      terrain.Rect{X: 1, Y: 1, Width: 10, Height: 10},
			BlockType: terrain.BlockBuilding,
		},
		{
			Rect:      terrain.Rect{X: 20, Y: 20, Width: 10, Height: 10},
			BlockType: terrain.BlockBuilding,
		},
		{
			Rect:      terrain.Rect{X: 40, Y: 40, Width: 5, Height: 5},
			BlockType: terrain.BlockBuilding,
		},
	}

	selected := spawner.selectCentralBlock(blocks, cityTerrain)

	// Center of 50x50 is (25, 25)
	// Block at (20, 20) with size 10x10 has center at (25, 25) - exact match!
	cx, cy := selected.Rect.Center()
	if cx != 25 || cy != 25 {
		t.Errorf("Expected block centered at (25, 25), got (%d, %d)", cx, cy)
	}
}

func TestPostOfficeSpawner_GenerateClerkName(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			// Generate multiple names to ensure variety
			names := make(map[string]bool)

			for i := 0; i < 20; i++ {
				rng := rand.New(rand.NewSource(int64(12345 + i)))
				name := spawner.generateClerkName(rng, genreID)

				if name == "" {
					t.Error("Generated empty name")
				}

				names[name] = true
			}

			// Should have some variety (at least 5 different names in 20 generations)
			if len(names) < 5 {
				t.Errorf("Insufficient name variety: only %d unique names in 20 generations", len(names))
			}
		})
	}
}

func TestPostOfficeSpawner_GenerateClerkName_Deterministic(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	seed := int64(12345)
	rng1 := rand.New(rand.NewSource(seed))
	rng2 := rand.New(rand.NewSource(seed))

	name1 := spawner.generateClerkName(rng1, "fantasy")
	name2 := spawner.generateClerkName(rng2, "fantasy")

	if name1 != name2 {
		t.Errorf("Names not deterministic: %s != %s", name1, name2)
	}
}

// Benchmark tests
func BenchmarkPostOfficeSpawner_SpawnInCity(b *testing.B) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	cityTerrain := terrain.NewTerrain(80, 50, 12345)
	for y := 0; y < 50; y++ {
		for x := 0; x < 80; x++ {
			cityTerrain.SetTile(x, y, terrain.TileFloor)
		}
	}

	for x := 0; x < 80; x++ {
		cityTerrain.SetTile(x, 10, terrain.TileCorridor)
		cityTerrain.SetTile(x, 30, terrain.TileCorridor)
	}

	blocks := []*terrain.CityBlock{
		{Rect: terrain.Rect{X: 11, Y: 11, Width: 18, Height: 18}, BlockType: terrain.BlockBuilding},
		{Rect: terrain.Rect{X: 31, Y: 31, Width: 15, Height: 15}, BlockType: terrain.BlockBuilding},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spawner.SpawnInCity(cityTerrain, blocks, "fantasy", 12345)
	}
}

func BenchmarkPostOfficeSpawner_GenerateClerkName(b *testing.B) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	rng := rand.New(rand.NewSource(12345))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spawner.generateClerkName(rng, "fantasy")
	}
}

func BenchmarkPostOfficeSpawner_FindSuitableBlocks(b *testing.B) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	cityTerrain := terrain.NewTerrain(80, 50, 12345)
	for y := 0; y < 50; y++ {
		for x := 0; x < 80; x++ {
			cityTerrain.SetTile(x, y, terrain.TileFloor)
		}
	}
	for x := 0; x < 80; x++ {
		cityTerrain.SetTile(x, 10, terrain.TileCorridor)
	}

	blocks := make([]*terrain.CityBlock, 20)
	for i := 0; i < 20; i++ {
		blocks[i] = &terrain.CityBlock{
			Rect:      terrain.Rect{X: i * 4, Y: 11, Width: 10, Height: 10},
			BlockType: terrain.BlockBuilding,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spawner.findSuitableBlocks(blocks, cityTerrain)
	}
}

// Tests for SpawnInTerrain (generic fallback for non-city terrains)

func TestPostOfficeSpawner_SpawnInTerrain(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	// Create terrain with rooms
	ter := terrain.NewTerrain(100, 100, 42)
	ter.Rooms = []*terrain.Room{
		{X: 10, Y: 10, Width: 8, Height: 8},
	}

	result, err := spawner.SpawnInTerrain(ter, "fantasy", 42)
	if err != nil {
		t.Fatalf("SpawnInTerrain() error = %v", err)
	}

	if result.ClerkName == "" {
		t.Error("clerk name should not be empty")
	}

	if result.BuildingID == 0 {
		t.Error("building ID should not be zero")
	}
}

func TestPostOfficeSpawner_SpawnInTerrainNilTerrain(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	_, err := spawner.SpawnInTerrain(nil, "fantasy", 42)
	if err == nil {
		t.Error("SpawnInTerrain(nil) should return error")
	}
}

func TestPostOfficeSpawner_SpawnInTerrainNoRooms(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	ter := terrain.NewTerrain(50, 50, 42)
	// No rooms added
	_, err := spawner.SpawnInTerrain(ter, "fantasy", 42)
	if err == nil {
		t.Error("SpawnInTerrain() with no rooms should return error")
	}
}

func TestPostOfficeSpawner_SpawnInTerrainTooSmallRooms(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	courierSystem := NewCourierSystem(world, mailSystem)
	spawner := NewPostOfficeSpawner(world, courierSystem)

	ter := terrain.NewTerrain(50, 50, 42)
	ter.Rooms = []*terrain.Room{
		{X: 5, Y: 5, Width: 3, Height: 3}, // area 9, too small (need >= 36)
	}

	_, err := spawner.SpawnInTerrain(ter, "fantasy", 42)
	if err == nil {
		t.Error("SpawnInTerrain() with only small rooms should return error")
	}
}

func TestPostOfficeSpawner_SpawnInTerrainDeterministic(t *testing.T) {
	ter := terrain.NewTerrain(100, 100, 42)
	ter.Rooms = []*terrain.Room{
		{X: 10, Y: 10, Width: 10, Height: 10},
	}

	// Same seed should produce same clerk name
	world1 := NewWorld()
	spawner1 := NewPostOfficeSpawner(world1, NewCourierSystem(world1, NewMailSystem(world1)))
	result1, _ := spawner1.SpawnInTerrain(ter, "fantasy", 42)

	world2 := NewWorld()
	spawner2 := NewPostOfficeSpawner(world2, NewCourierSystem(world2, NewMailSystem(world2)))
	result2, _ := spawner2.SpawnInTerrain(ter, "fantasy", 42)

	if result1.ClerkName != result2.ClerkName {
		t.Errorf("determinism failed: %q != %q", result1.ClerkName, result2.ClerkName)
	}
}
