package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewTerrainMovementSpeedSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTerrainMovementSpeedSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.tileSize != 32 {
		t.Errorf("tileSize = %d, want 32", sys.tileSize)
	}
	if sys.speedCache == nil {
		t.Error("speedCache is nil")
	}
}

func TestTerrainMovementSpeedSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10, 12345)
	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("terrain not set correctly")
	}
}

func TestTerrainMovementSpeedSystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64", sys.tileSize)
	}

	// Invalid tile size should be ignored
	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64 (invalid size should be ignored)", sys.tileSize)
	}

	sys.SetTileSize(-1)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64 (negative size should be ignored)", sys.tileSize)
	}
}

func TestTerrainMovementSpeedSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)

	sys.SetGenre("fantasy")
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", sys.genreID)
	}
}

func TestTerrainMovementSpeedSystem_NoTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	// Should not panic without terrain
	sys.Update([]*Entity{entity}, 0.016)

	vel := entity.GetVelocity()
	if vel.VX != 10 || vel.VY != 10 {
		t.Errorf("velocity changed without terrain, got (%f, %f)", vel.VX, vel.VY)
	}
}

func TestTerrainMovementSpeedSystem_FloorTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileFloor)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100}) // Tile 3,3 at tileSize=32
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	sys.Update([]*Entity{entity}, 0.016)

	vel := entity.GetVelocity()
	// Floor has movement cost 1.0, so velocity should be unchanged
	if vel.VX != 10 || vel.VY != 10 {
		t.Errorf("velocity on floor should be unchanged, got (%f, %f)", vel.VX, vel.VY)
	}
}

func TestTerrainMovementSpeedSystem_ShallowWater(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100}) // Tile 3,3
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	sys.Update([]*Entity{entity}, 0.016)

	vel := entity.GetVelocity()
	// Shallow water has movement cost 2.0, so velocity should be halved
	if vel.VX > 5.1 || vel.VX < 4.9 {
		t.Errorf("velocity in shallow water should be ~5, got VX=%f", vel.VX)
	}
	if vel.VY > 5.1 || vel.VY < 4.9 {
		t.Errorf("velocity in shallow water should be ~5, got VY=%f", vel.VY)
	}
}

func TestTerrainMovementSpeedSystem_Ramp(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileRamp)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100}) // Tile 3,3
	entity.AddComponent(&VelocityComponent{VX: 12, VY: 12})

	sys.Update([]*Entity{entity}, 0.016)

	vel := entity.GetVelocity()
	// Ramp has movement cost 1.2, so velocity should be divided by 1.2
	expectedVel := 12.0 / 1.2
	if vel.VX > expectedVel+0.1 || vel.VX < expectedVel-0.1 {
		t.Errorf("velocity on ramp should be ~%f, got VX=%f", expectedVel, vel.VX)
	}
}

func TestTerrainMovementSpeedSystem_GenreFantasyWater(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)
	sys.SetGenre("fantasy")

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	sys.Update([]*Entity{entity}, 0.016)

	vel := entity.GetVelocity()
	// Fantasy water: base cost 2.0 * genre modifier 1.15 = 2.3
	expectedVel := 10.0 / 2.3
	if vel.VX > expectedVel+0.2 || vel.VX < expectedVel-0.2 {
		t.Errorf("fantasy water velocity should be ~%f, got VX=%f", expectedVel, vel.VX)
	}
}

func TestTerrainMovementSpeedSystem_GenreHorrorSlower(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)
	sys.SetGenre("horror")

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileFloor)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	sys.Update([]*Entity{entity}, 0.016)

	vel := entity.GetVelocity()
	// Horror: floor cost 1.0 * genre modifier 1.08 = 1.08
	expectedVel := 10.0 / 1.08
	if vel.VX > expectedVel+0.1 || vel.VX < expectedVel-0.1 {
		t.Errorf("horror floor velocity should be ~%f, got VX=%f", expectedVel, vel.VX)
	}
}

func TestTerrainMovementSpeedSystem_GenreScifiFaster(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)
	sys.SetGenre("scifi")

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileFloor)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	sys.Update([]*Entity{entity}, 0.016)

	vel := entity.GetVelocity()
	// Scifi: floor cost 1.0 * genre modifier 0.95 = 0.95 (faster)
	expectedVel := 10.0 / 0.95
	if vel.VX > expectedVel+0.2 || vel.VX < expectedVel-0.2 {
		t.Errorf("scifi floor velocity should be ~%f, got VX=%f", expectedVel, vel.VX)
	}
}

func TestTerrainMovementSpeedSystem_GetSpeedMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	// Before update, multiplier should be 1.0 (default)
	mult := sys.GetSpeedMultiplier(entity.ID)
	if mult != 1.0 {
		t.Errorf("default multiplier should be 1.0, got %f", mult)
	}

	sys.Update([]*Entity{entity}, 0.016)

	// After update, multiplier should be 2.0 (shallow water)
	mult = sys.GetSpeedMultiplier(entity.ID)
	if mult < 1.9 || mult > 2.1 {
		t.Errorf("water multiplier should be ~2.0, got %f", mult)
	}
}

func TestTerrainMovementSpeedSystem_GetTerrainTypeAt(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(5, 5, terrain.TileWaterShallow)
	terr.SetTile(3, 3, terrain.TileLavaFlow)
	sys.SetTerrain(terr)

	// Test water tile
	tileType := sys.GetTerrainTypeAt(160, 160) // Tile 5,5
	if tileType != terrain.TileWaterShallow {
		t.Errorf("expected TileWaterShallow at (160,160), got %s", tileType.String())
	}

	// Test lava tile
	tileType = sys.GetTerrainTypeAt(100, 100) // Tile 3,3
	if tileType != terrain.TileLavaFlow {
		t.Errorf("expected TileLavaFlow at (100,100), got %s", tileType.String())
	}

	// Test without terrain
	sys.SetTerrain(nil)
	tileType = sys.GetTerrainTypeAt(100, 100)
	if tileType != terrain.TileFloor {
		t.Errorf("expected TileFloor without terrain, got %s", tileType.String())
	}
}

func TestTerrainMovementSpeedSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(10, 10, 12345)
	terr.SetTile(3, 3, terrain.TileFloor)
	terr.SetTile(5, 5, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100}) // Floor
	entity1.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 160, Y: 160}) // Water
	entity2.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	sys.Update([]*Entity{entity1, entity2}, 0.016)

	vel1 := entity1.GetVelocity()
	vel2 := entity2.GetVelocity()

	// Entity1 on floor should have unchanged velocity
	if vel1.VX != 10 || vel1.VY != 10 {
		t.Errorf("entity1 on floor velocity should be unchanged, got (%f, %f)", vel1.VX, vel1.VY)
	}

	// Entity2 in water should have halved velocity
	if vel2.VX > 5.1 || vel2.VX < 4.9 {
		t.Errorf("entity2 in water velocity should be ~5, got VX=%f", vel2.VX)
	}
}

func TestTerrainMovementSpeedSystem_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10, 12345)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})

	// Should not panic without position
	sys.Update([]*Entity{entity}, 0.016)

	vel := entity.GetVelocity()
	if vel.VX != 10 {
		t.Errorf("velocity should be unchanged without position, got VX=%f", vel.VX)
	}
}

func TestTerrainMovementSpeedSystem_NoVelocity(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)

	terr := terrain.NewTerrain(10, 10, 12345)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic without velocity
	sys.Update([]*Entity{entity}, 0.016)
}

func BenchmarkTerrainMovementSpeedSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainMovementSpeedSystem(world, 12345)
	sys.SetTileSize(32)

	terr := terrain.NewTerrain(100, 100, 12345)
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if (x+y)%3 == 0 {
				terr.SetTile(x, y, terrain.TileWaterShallow)
			} else {
				terr.SetTile(x, y, terrain.TileFloor)
			}
		}
	}
	sys.SetTerrain(terr)

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entity.AddComponent(&VelocityComponent{VX: 10, VY: 10})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
