//go:build ignore

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestWaterSurfaceRippleComponent_Type(t *testing.T) {
	c := &WaterSurfaceRippleComponent{}
	if got := c.Type(); got != "water_surface_ripple" {
		t.Errorf("Type() = %q, want %q", got, "water_surface_ripple")
	}
}

func TestNewWaterSurfaceRippleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 42)
	if sys == nil {
		t.Fatal("NewWaterSurfaceRippleSystem returned nil")
	}
	if sys.seed != 42 {
		t.Errorf("seed = %d, want 42", sys.seed)
	}
	if sys.tileSize != 32 {
		t.Errorf("tileSize = %d, want 32", sys.tileSize)
	}
}

func TestWaterSurfaceRippleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"scifi", "scifi"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWaterSurfaceRippleSystem(world, 1)
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %q, want %q", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestWaterSurfaceRippleSystem_SetTileSize(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 1)

	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64", sys.tileSize)
	}

	// Zero and negative should be ignored
	sys.SetTileSize(0)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed to %d on zero input", sys.tileSize)
	}
	sys.SetTileSize(-1)
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed to %d on negative input", sys.tileSize)
	}
}

func TestWaterSurfaceRippleSystem_isWaterTile(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 1)

	tests := []struct {
		name   string
		tile   terrain.TileType
		expect bool
	}{
		{"shallow_water", terrain.TileWaterShallow, true},
		{"deep_water", terrain.TileWaterDeep, true},
		{"floor", terrain.TileFloor, false},
		{"wall", terrain.TileWall, false},
		{"corridor", terrain.TileCorridor, false},
		{"bridge", terrain.TileBridge, false},
		{"lava", terrain.TileLavaFlow, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.isWaterTile(tt.tile)
			if got != tt.expect {
				t.Errorf("isWaterTile(%v) = %v, want %v", tt.tile, got, tt.expect)
			}
		})
	}
}

func TestWaterSurfaceRippleSystem_Update_NilParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 1)
	// Should not panic when particleSystem is nil
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&VelocityComponent{VX: 50, VY: 0})
	sys.Update([]*Entity{entity}, 0.016)
}

func TestWaterSurfaceRippleSystem_Update_NilWorld(t *testing.T) {
	sys := NewWaterSurfaceRippleSystem(nil, 1)
	// Should not panic
	sys.Update([]*Entity{}, 0.016)
}

func TestWaterSurfaceRippleSystem_Update_NonWaterTile(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 1)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 0})

	// No terrain set, defaults to TileFloor — no ripples should attach
	sys.Update([]*Entity{entity}, 0.016)

	_, hasComp := entity.GetComponent("water_surface_ripple")
	if hasComp {
		t.Error("should not attach ripple component on non-water tile")
	}
}

func TestWaterSurfaceRippleSystem_Update_WaterTile(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create a terrain with shallow water at (0,0)
	terr := terrain.NewTerrain(10, 10, 42)
	terr.SetTile(0, 0, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 5, Y: 5}) // tile (0,0) with 32px tiles
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 0})

	// First update: should trigger entry splash and attach component
	sys.Update([]*Entity{entity}, 0.016)

	comp, ok := entity.GetComponent("water_surface_ripple")
	if !ok {
		t.Fatal("expected water_surface_ripple component after entering water")
	}
	rc := comp.(*WaterSurfaceRippleComponent)
	if !rc.InWater {
		t.Error("expected InWater=true after entering water tile")
	}
}

func TestWaterSurfaceRippleSystem_Update_LeavingWater(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	terr := terrain.NewTerrain(10, 10, 42)
	terr.SetTile(0, 0, terrain.TileWaterShallow)
	terr.SetTile(1, 0, terrain.TileFloor)
	sys.SetTerrain(terr)

	entity := NewEntity(1)
	posComp := &PositionComponent{X: 5, Y: 5}
	entity.AddComponent(posComp)
	entity.AddComponent(&VelocityComponent{VX: 50, VY: 0})

	// Enter water
	sys.Update([]*Entity{entity}, 0.016)
	comp, _ := entity.GetComponent("water_surface_ripple")
	rc := comp.(*WaterSurfaceRippleComponent)
	if !rc.InWater {
		t.Fatal("expected InWater=true")
	}

	// Move to floor tile
	posComp.X = 40 // tile (1,0)
	sys.Update([]*Entity{entity}, 0.016)
	if rc.InWater {
		t.Error("expected InWater=false after leaving water")
	}
}

func TestWaterSurfaceRippleSystem_Update_IdleInWater(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	terr := terrain.NewTerrain(10, 10, 42)
	terr.SetTile(0, 0, terrain.TileWaterShallow)
	sys.SetTerrain(terr)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 5, Y: 5})
	// No velocity component — idle entity

	// Enter water
	sys.Update([]*Entity{entity}, 0.016)
	// Second update: idle ripple
	sys.Update([]*Entity{entity}, 1.0) // large dt to exhaust cooldown

	comp, ok := entity.GetComponent("water_surface_ripple")
	if !ok {
		t.Fatal("expected component")
	}
	rc := comp.(*WaterSurfaceRippleComponent)
	// Idle entities should still produce low-intensity ripples
	if rc.Intensity > 0.2 {
		t.Errorf("idle intensity = %f, expected <=0.2", rc.Intensity)
	}
}

func TestWaterSurfaceRippleSystem_Update_DeepWater(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	terr := terrain.NewTerrain(10, 10, 42)
	terr.SetTile(0, 0, terrain.TileWaterDeep)
	sys.SetTerrain(terr)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 5, Y: 5})
	entity.AddComponent(&VelocityComponent{VX: 150, VY: 0})

	// Enter deep water
	sys.Update([]*Entity{entity}, 0.016)
	comp, ok := entity.GetComponent("water_surface_ripple")
	if !ok {
		t.Fatal("expected component on deep water")
	}
	if !comp.(*WaterSurfaceRippleComponent).InWater {
		t.Error("expected InWater=true on deep water")
	}
}

func TestWaterSurfaceRippleSystem_GenrePresets(t *testing.T) {
	world := NewWorld()
	genres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			sys := NewWaterSurfaceRippleSystem(world, 1)
			sys.SetGenre(g)
			if sys.preset.minSize <= 0 {
				t.Errorf("genre %s has invalid minSize %f", g, sys.preset.minSize)
			}
			if sys.preset.maxSize < sys.preset.minSize {
				t.Errorf("genre %s maxSize %f < minSize %f", g, sys.preset.maxSize, sys.preset.minSize)
			}
			if sys.preset.duration <= 0 {
				t.Errorf("genre %s has non-positive duration %f", g, sys.preset.duration)
			}
		})
	}
}

func TestWaterSurfaceRippleSystem_Determinism(t *testing.T) {
	// Same seed should produce same internal state
	world1 := NewWorld()
	world2 := NewWorld()
	sys1 := NewWaterSurfaceRippleSystem(world1, 12345)
	sys2 := NewWaterSurfaceRippleSystem(world2, 12345)

	if sys1.rng.Int63() != sys2.rng.Int63() {
		t.Error("same seed produced different RNG sequences")
	}
}

func TestWaterSurfaceRippleSystem_getRippleComponent(t *testing.T) {
	world := NewWorld()
	sys := NewWaterSurfaceRippleSystem(world, 1)

	entity := NewEntity(1)

	// First call creates component
	c1 := sys.getRippleComponent(entity)
	if c1 == nil {
		t.Fatal("getRippleComponent returned nil")
	}

	// Second call returns same component
	c2 := sys.getRippleComponent(entity)
	if c1 != c2 {
		t.Error("getRippleComponent should return existing component")
	}
}
