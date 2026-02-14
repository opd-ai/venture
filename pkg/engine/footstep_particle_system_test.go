package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestNewFootstepParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewFootstepParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("rng not initialized")
	}
	if sys.tileSize != 32 {
		t.Errorf("tileSize = %d, want 32", sys.tileSize)
	}
	if sys.lastFootstep == nil {
		t.Error("lastFootstep map not initialized")
	}
	if sys.lastTile == nil {
		t.Error("lastTile map not initialized")
	}
}

func TestFootstepParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestFootstepParticleSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	terr := terrain.NewTerrain(10, 10)

	// Add some entries to lastTile to verify it gets cleared
	sys.lastTile[1] = tileCoord{x: 1, y: 1}

	sys.SetTerrain(terr)

	if sys.terrain != terr {
		t.Error("terrain not set correctly")
	}
	if len(sys.lastTile) != 0 {
		t.Error("lastTile should be cleared when terrain changes")
	}
}

func TestFootstepParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)

	sys.SetGenre("horror")

	if sys.genreID != "horror" {
		t.Errorf("genreID = %s, want horror", sys.genreID)
	}
}

func TestFootstepParticleSystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64", sys.tileSize)
	}

	// Invalid size should not change
	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed with invalid value, got %d", sys.tileSize)
	}

	sys.SetTileSize(-1)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed with negative value, got %d", sys.tileSize)
	}
}

func TestFootstepParticleSystem_UpdateWithoutDependencies(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)

	// Create a moving entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 50, VY: 0})

	// Should not panic without particle system or terrain
	sys.Update([]*Entity{entity}, 0.016)
}

func TestFootstepParticleSystem_UpdateEntityTooSlow(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	ps := NewParticleSystem()
	terr := terrain.NewTerrain(10, 10)
	terr.SetTile(3, 3, terrain.TileFloor)

	sys.SetParticleSystem(ps)
	sys.SetTerrain(terr)

	// Create entity moving too slow (speed < 10)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 5, VY: 0}) // Speed = 5, below threshold

	initialEntities := len(world.entities)
	sys.Update([]*Entity{entity}, 0.016)

	// Should not spawn particles for slow-moving entity
	if len(world.entities) != initialEntities {
		t.Error("particles spawned for entity moving too slowly")
	}
}

func TestFootstepParticleSystem_UpdateEntityMoving(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	ps := NewParticleSystem()
	terr := terrain.NewTerrain(20, 20)
	// Set up floor tiles around the entity position
	for x := 0; x < 20; x++ {
		for y := 0; y < 20; y++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	sys.SetParticleSystem(ps)
	sys.SetTerrain(terr)
	sys.SetGenre("fantasy")

	// Create fast-moving entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 160, Y: 160}) // Tile (5, 5)
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 0}) // Fast enough

	// First update should record position
	sys.Update([]*Entity{entity}, 0.016)

	// Simulate time passing and entity moving to new tile
	entity.GetPosition().X = 192                    // Move to tile (6, 5)
	sys.lastFootstep[entity.ID] = sys.spawnInterval // Ensure cooldown passed

	initialEntities := len(world.entities)
	sys.Update([]*Entity{entity}, sys.spawnInterval+0.01)

	// Should spawn particles when crossing tile boundary
	if len(world.entities) <= initialEntities {
		t.Error("expected particles to spawn when entity crosses tile boundary")
	}
}

func TestFootstepParticleSystem_GetParticleConfig(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	sys.SetGenre("fantasy")

	tests := []struct {
		name     string
		tileType terrain.TileType
		wantNil  bool
	}{
		{"floor", terrain.TileFloor, false},
		{"corridor", terrain.TileCorridor, false},
		{"door", terrain.TileDoor, false},
		{"water", terrain.TileWaterShallow, false},
		{"lava", terrain.TileLavaFlow, false},
		{"bridge", terrain.TileBridge, false},
		{"ramp", terrain.TileRamp, false},
		{"ramp_up", terrain.TileRampUp, false},
		{"ramp_down", terrain.TileRampDown, false},
		{"platform", terrain.TilePlatform, false},
		{"wall", terrain.TileWall, true}, // No footsteps on walls
		{"tree", terrain.TileTree, true}, // No footsteps on trees
		{"pit", terrain.TilePit, true},   // No footsteps on pits
		{"deep_water", terrain.TileWaterDeep, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := sys.getParticleConfig(tt.tileType, 100, 100)
			if tt.wantNil && config != nil {
				t.Errorf("expected nil config for %s", tt.name)
			}
			if !tt.wantNil && config == nil {
				t.Errorf("expected non-nil config for %s", tt.name)
			}
		})
	}
}

func TestFootstepParticleSystem_GenreVariations(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys.SetGenre(genre)

			// Test floor config varies by genre
			floorConfig := sys.getFloorConfig(12345)
			if floorConfig == nil {
				t.Error("floor config should not be nil")
			}

			// Test ramp config varies by genre (scifi/cyberpunk use sparks)
			rampConfig := sys.getRampConfig(12345)
			if rampConfig == nil {
				t.Error("ramp config should not be nil")
			}

			// Test platform config varies by genre
			platformConfig := sys.getPlatformConfig(12345)
			if platformConfig == nil {
				t.Error("platform config should not be nil")
			}
		})
	}
}

func TestFootstepParticleSystem_CooldownTracking(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	ps := NewParticleSystem()
	terr := terrain.NewTerrain(20, 20)
	for x := 0; x < 20; x++ {
		for y := 0; y < 20; y++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	sys.SetParticleSystem(ps)
	sys.SetTerrain(terr)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 160, Y: 160})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 0})

	// Update multiple times with small delta
	for i := 0; i < 5; i++ {
		sys.Update([]*Entity{entity}, 0.01) // 10ms each, total 50ms < 150ms cooldown
	}

	// Cooldown should be accumulating
	if sys.lastFootstep[entity.ID] < 0.04 {
		t.Error("cooldown should be accumulating")
	}
}

func TestFootstepParticleSystem_TileTracking(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	terr := terrain.NewTerrain(10, 10)

	sys.SetTerrain(terr)

	// Verify tile tracking is reset when terrain changes
	sys.lastTile[1] = tileCoord{x: 5, y: 5}

	newTerr := terrain.NewTerrain(20, 20)
	sys.SetTerrain(newTerr)

	if _, exists := sys.lastTile[1]; exists {
		t.Error("tile tracking should be cleared when terrain changes")
	}
}

func TestFootstepParticleSystem_EntityWithoutPosition(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	ps := NewParticleSystem()
	terr := terrain.NewTerrain(10, 10)

	sys.SetParticleSystem(ps)
	sys.SetTerrain(terr)

	// Entity without position
	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 0})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestFootstepParticleSystem_EntityWithoutVelocity(t *testing.T) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	ps := NewParticleSystem()
	terr := terrain.NewTerrain(10, 10)

	sys.SetParticleSystem(ps)
	sys.SetTerrain(terr)

	// Entity without velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func BenchmarkFootstepParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewFootstepParticleSystem(world, 12345)
	ps := NewParticleSystem()
	terr := terrain.NewTerrain(100, 100)
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			terr.SetTile(x, y, terrain.TileFloor)
		}
	}

	sys.SetParticleSystem(ps)
	sys.SetTerrain(terr)
	sys.SetGenre("fantasy")

	// Create 100 moving entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entity.AddComponent(&VelocityComponent{VX: 100, VY: 0})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
