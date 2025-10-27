package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestNewFirePropagationSystem verifies system creation.
func TestNewFirePropagationSystem(t *testing.T) {
	system := NewFirePropagationSystem(32, 12345)

	if system == nil {
		t.Fatal("expected system, got nil")
	}
	if system.tileSize != 32 {
		t.Errorf("expected tileSize=32, got %d", system.tileSize)
	}
	if system.rng == nil {
		t.Error("expected rng to be initialized")
	}
	if system.fireEntities == nil {
		t.Error("expected fireEntities map to be initialized")
	}
}

// TestFirePropagationSystem_SetReferences verifies setting world and terrain references.
func TestFirePropagationSystem_SetReferences(t *testing.T) {
	system := NewFirePropagationSystem(32, 12345)
	world := NewWorld()
	terr := createTestTerrain(10, 10)

	system.SetWorld(world)
	system.SetTerrain(terr)

	if system.world != world {
		t.Error("expected world reference to be set")
	}
	if system.terrain != terr {
		t.Error("expected terrain reference to be set")
	}
}

// TestFirePropagationSystem_IsTileFlammable verifies flammability checks.
func TestFirePropagationSystem_IsTileFlammable(t *testing.T) {
	tests := []struct {
		name        string
		tileType    terrain.TileType
		material    MaterialType
		hasDestComp bool
		want        bool
	}{
		{"floor no component", terrain.TileFloor, MaterialStone, false, true},
		{"wall no component", terrain.TileWall, MaterialStone, false, false},
		{"wood wall", terrain.TileWall, MaterialWood, true, true},
		{"stone wall", terrain.TileWall, MaterialStone, true, false},
		{"metal wall", terrain.TileWall, MaterialMetal, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewFirePropagationSystem(32, 12345)
			world := NewWorld()
			terr := createTestTerrain(10, 10)
			system.SetWorld(world)
			system.SetTerrain(terr)

			terr.SetTile(5, 5, tt.tileType)

			if tt.hasDestComp {
				entity := world.CreateEntity()
				destComp := NewDestructibleComponent(tt.material, 5, 5)
				entity.AddComponent(destComp)
				entity.AddComponent(&PositionComponent{X: 160, Y: 160})
				world.Update(0)
			}

			got := system.isTileFlammable(5, 5)
			if got != tt.want {
				t.Errorf("isTileFlammable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFirePropagationSystem_IgniteTile verifies manual tile ignition.
func TestFirePropagationSystem_IgniteTile(t *testing.T) {
	system := NewFirePropagationSystem(32, 12345)
	world := NewWorld()
	terr := createTestTerrain(10, 10)
	system.SetWorld(world)
	system.SetTerrain(terr)

	terr.SetTile(5, 5, terrain.TileFloor)
	system.IgniteTile(5, 5, 0.8)
	world.Update(0)

	entities := world.GetEntities()
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}

	comp, ok := entities[0].GetComponent("fire")
	if !ok {
		t.Fatal("expected fire component")
	}
	fireComp, ok := comp.(*FireComponent)
	if !ok {
		t.Fatal("expected FireComponent type")
	}
	if fireComp.Intensity != 0.8 {
		t.Errorf("expected intensity=0.8, got %f", fireComp.Intensity)
	}
}

// TestFirePropagationSystem_IgniteTile_NonFlammable verifies ignition fails on non-flammable tiles.
func TestFirePropagationSystem_IgniteTile_NonFlammable(t *testing.T) {
	system := NewFirePropagationSystem(32, 12345)
	world := NewWorld()
	terr := createTestTerrain(10, 10)
	system.SetWorld(world)
	system.SetTerrain(terr)

	terr.SetTile(5, 5, terrain.TileWall)
	system.IgniteTile(5, 5, 0.8)
	world.Update(0)

	entities := world.GetEntities()
	if len(entities) != 0 {
		t.Errorf("expected 0 entities, got %d", len(entities))
	}
}

// TestFirePropagationSystem_IgniteTilesInArea verifies area ignition.
func TestFirePropagationSystem_IgniteTilesInArea(t *testing.T) {
	system := NewFirePropagationSystem(32, 12345)
	world := NewWorld()
	terr := createTestTerrain(10, 10)
	system.SetWorld(world)
	system.SetTerrain(terr)

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	centerX := 5.5 * 32
	centerY := 5.5 * 32
	system.IgniteTilesInArea(centerX, centerY, 64, 1.0)
	world.Update(0)

	entities := world.GetEntities()
	if len(entities) < 5 || len(entities) > 21 {
		t.Errorf("expected 5-21 fire entities, got %d", len(entities))
	}
}

// TestFirePropagationSystem_Update verifies fire component updates.
func TestFirePropagationSystem_Update(t *testing.T) {
	system := NewFirePropagationSystem(32, 12345)
	world := NewWorld()
	terr := createTestTerrain(10, 10)
	system.SetWorld(world)
	system.SetTerrain(terr)

	entity := world.CreateEntity()
	fireComp := NewFireComponent(1.0, 5, 5, 1.0)
	entity.AddComponent(fireComp)
	entity.AddComponent(&PositionComponent{X: 160, Y: 160})
	world.Update(0)

	entities := world.GetEntities()
	system.Update(entities, 1.5)

	// Flush entity removal
	world.Update(0)

	if len(world.GetEntities()) != 0 {
		t.Errorf("expected fire entity to be removed after burning out, got %d entities", len(world.GetEntities()))
	}
}

// TestFirePropagationSystem_GetActiveFireCount verifies fire count tracking.
func TestFirePropagationSystem_GetActiveFireCount(t *testing.T) {
	system := NewFirePropagationSystem(32, 12345)
	world := NewWorld()
	terr := createTestTerrain(10, 10)
	system.SetWorld(world)
	system.SetTerrain(terr)

	if count := system.GetActiveFireCount(); count != 0 {
		t.Errorf("expected 0 fires initially, got %d", count)
	}

	terr.SetTile(3, 3, terrain.TileFloor)
	terr.SetTile(5, 5, terrain.TileFloor)
	system.IgniteTile(3, 3, 1.0)
	system.IgniteTile(5, 5, 1.0)

	if count := system.GetActiveFireCount(); count != 2 {
		t.Errorf("expected 2 fires, got %d", count)
	}
}

// TestFirePropagationSystem_FireSpread verifies cellular automata spread.
func TestFirePropagationSystem_FireSpread(t *testing.T) {
	system := NewFirePropagationSystem(32, 54321)
	world := NewWorld()
	terr := createTestTerrain(10, 10)
	system.SetWorld(world)
	system.SetTerrain(terr)

	for y := 4; y <= 6; y++ {
		for x := 4; x <= 6; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	system.IgniteTile(5, 5, 1.0)
	initialCount := system.GetActiveFireCount()

	for i := 0; i < 50; i++ {
		world.Update(0)
		entities := world.GetEntities()
		system.Update(entities, 0.1)
	}

	finalCount := system.GetActiveFireCount()
	if finalCount <= initialCount {
		t.Errorf("expected fire to spread (>%d fires), got %d", initialCount, finalCount)
	}
}

// Benchmark fire propagation update performance.
func BenchmarkFirePropagationSystem_Update(b *testing.B) {
	system := NewFirePropagationSystem(32, 12345)
	world := NewWorld()
	terr := createTestTerrain(50, 50)
	system.SetWorld(world)
	system.SetTerrain(terr)

	for i := 0; i < 50; i++ {
		terr.SetTile(i, i, terrain.TileFloor)
		system.IgniteTile(i, i, 1.0)
	}
	world.Update(0)

	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

// Benchmark area ignition performance.
func BenchmarkFirePropagationSystem_IgniteTilesInArea(b *testing.B) {
	system := NewFirePropagationSystem(32, 12345)
	world := NewWorld()
	terr := createTestTerrain(100, 100)
	system.SetWorld(world)
	system.SetTerrain(terr)

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.IgniteTilesInArea(1600, 1600, 128, 1.0)
	}
}

// createTestTerrain creates a terrain for testing.
func createTestTerrain(width, height int) *terrain.Terrain {
	tiles := make([][]terrain.TileType, height)
	for y := range tiles {
		tiles[y] = make([]terrain.TileType, width)
		for x := range tiles[y] {
			tiles[y][x] = terrain.TileWall
		}
	}

	return &terrain.Terrain{
		Width:  width,
		Height: height,
		Tiles:  tiles,
	}
}
