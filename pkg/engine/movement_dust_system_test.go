package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestMovementDustComponentType(t *testing.T) {
	c := &MovementDustComponent{}
	if c.Type() != "movement_dust" {
		t.Errorf("expected 'movement_dust', got %q", c.Type())
	}
}

func TestNewMovementDustSystem(t *testing.T) {
	tests := []struct {
		name  string
		world *World
		seed  int64
	}{
		{"with world", NewWorld(), 42},
		{"nil world", nil, 0},
		{"different seed", NewWorld(), 99999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewMovementDustSystem(tt.world, tt.seed)
			if sys == nil {
				t.Fatal("NewMovementDustSystem returned nil")
			}
			if sys.seed != tt.seed {
				t.Errorf("seed = %d, want %d", sys.seed, tt.seed)
			}
		})
	}
}

func TestMovementDustSystemSetGenre(t *testing.T) {
	genres := []struct {
		id       string
		wantType particles.ParticleType
	}{
		{"fantasy", particles.ParticleDust},
		{"horror", particles.ParticleSmoke},
		{"scifi", particles.ParticleSpark},
		{"cyberpunk", particles.ParticleSpark},
		{"postapoc", particles.ParticleDebris},
		{"unknown", particles.ParticleDust},
	}
	for _, tt := range genres {
		t.Run(tt.id, func(t *testing.T) {
			sys := NewMovementDustSystem(NewWorld(), 42)
			sys.SetGenre(tt.id)
			if sys.genreID != tt.id {
				t.Errorf("genreID = %q, want %q", sys.genreID, tt.id)
			}
			if sys.preset.particleType != tt.wantType {
				t.Errorf("particleType = %v, want %v", sys.preset.particleType, tt.wantType)
			}
		})
	}
}

func TestMovementDustSystemSetTileSize(t *testing.T) {
	sys := NewMovementDustSystem(NewWorld(), 1)
	sys.SetTileSize(64)
	if sys.tileSize != 64 {
		t.Errorf("tileSize = %d, want 64", sys.tileSize)
	}
	sys.SetTileSize(0) // should not change
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed to %d on zero input", sys.tileSize)
	}
	sys.SetTileSize(-1) // should not change
	if sys.tileSize != 64 {
		t.Errorf("tileSize changed to %d on negative input", sys.tileSize)
	}
}

func TestMovementDustSystemUpdateNilDeps(t *testing.T) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 0})

	// No particle system - should not panic
	sys.Update([]*Entity{entity}, 0.016)

	// Set particle system but nil world
	sys.world = nil
	sys.SetParticleSystem(&ParticleSystem{})
	sys.Update([]*Entity{entity}, 0.016)
}

func TestMovementDustSystemUpdateBelowThreshold(t *testing.T) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&VelocityComponent{VX: 5, VY: 5}) // Speed ~7, below 50

	sys.Update([]*Entity{entity}, 0.016)

	// Entity should NOT have dust component (too slow)
	if _, ok := entity.GetComponent("movement_dust"); ok {
		t.Error("dust component attached to slow entity")
	}
}

func TestMovementDustSystemUpdateAboveThreshold(t *testing.T) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 150, VY: 0}) // Speed 150, above threshold

	sys.Update([]*Entity{entity}, 0.016)

	comp, ok := entity.GetComponent("movement_dust")
	if !ok {
		t.Fatal("expected dust component on fast entity")
	}
	dc := comp.(*MovementDustComponent)
	if dc.Intensity <= 0 {
		t.Error("expected positive intensity for fast entity")
	}
	if dc.LastSpeedSq <= 0 {
		t.Error("expected positive LastSpeedSq")
	}
}

func TestMovementDustSystemSuppressed(t *testing.T) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})
	entity.AddComponent(&MovementDustComponent{Suppressed: true})

	sys.Update([]*Entity{entity}, 0.5)

	dc, ok := entity.GetComponent("movement_dust")
	if !ok {
		t.Fatal("expected movement_dust component")
	}
	mdc := dc.(*MovementDustComponent)
	if mdc.Cooldown != 0 {
		t.Error("suppressed entity should not have cooldown modified")
	}
}

func TestMovementDustSystemCooldownRespected(t *testing.T) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})

	// First update spawns dust and sets cooldown
	sys.Update([]*Entity{entity}, 0.016)
	dcRaw, ok := entity.GetComponent("movement_dust")
	if !ok {
		t.Fatal("expected movement_dust component")
	}
	dc := dcRaw.(*MovementDustComponent)
	cooldownAfterFirst := dc.Cooldown

	if cooldownAfterFirst <= 0 {
		t.Error("expected positive cooldown after first spawn")
	}

	// Very small delta should NOT trigger another spawn
	sys.Update([]*Entity{entity}, 0.001)
	cooldownAfterSecond := dc.Cooldown
	if cooldownAfterSecond >= cooldownAfterFirst {
		t.Error("cooldown should have decreased")
	}
}

func TestIsWalkableDustTerrain(t *testing.T) {
	sys := NewMovementDustSystem(NewWorld(), 1)
	tests := []struct {
		tile terrain.TileType
		want bool
	}{
		{terrain.TileFloor, true},
		{terrain.TileCorridor, true},
		{terrain.TileDoor, true},
		{terrain.TileBridge, true},
		{terrain.TileRamp, true},
		{terrain.TileRampUp, true},
		{terrain.TileRampDown, true},
		{terrain.TilePlatform, true},
		{terrain.TileStairsUp, true},
		{terrain.TileStairsDown, true},
		{terrain.TileLavaFlow, true},
		{terrain.TileWall, false},
		{terrain.TileTree, false},
		{terrain.TileWaterDeep, false},
		{terrain.TileWaterShallow, false},
		{terrain.TilePit, false},
	}
	for _, tt := range tests {
		t.Run(tt.tile.String(), func(t *testing.T) {
			got := sys.isWalkableDustTerrain(tt.tile)
			if got != tt.want {
				t.Errorf("isWalkableDustTerrain(%v) = %v, want %v", tt.tile, got, tt.want)
			}
		})
	}
}

func TestGetTerrainParticleStyle(t *testing.T) {
	sys := NewMovementDustSystem(NewWorld(), 1)
	sys.SetGenre("fantasy")

	tests := []struct {
		tile     terrain.TileType
		wantType particles.ParticleType
	}{
		{terrain.TileLavaFlow, particles.ParticleEmber},
		{terrain.TileBridge, particles.ParticleDebris},
		{terrain.TileStairsUp, particles.ParticleSpark},
		{terrain.TileStairsDown, particles.ParticleSpark},
		{terrain.TileFloor, particles.ParticleDust},
		{terrain.TileCorridor, particles.ParticleDust},
	}
	for _, tt := range tests {
		t.Run(tt.tile.String(), func(t *testing.T) {
			pType, dur, _ := sys.getTerrainParticleStyle(tt.tile)
			if pType != tt.wantType {
				t.Errorf("tile %v: particleType = %v, want %v", tt.tile, pType, tt.wantType)
			}
			if dur <= 0 {
				t.Errorf("tile %v: duration should be positive", tt.tile)
			}
		})
	}
}

func TestMovementDustIntensityScaling(t *testing.T) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	tests := []struct {
		name      string
		speed     float64
		wantRange [2]float64 // min, max intensity
	}{
		{"at threshold", 51, [2]float64{0.0, 0.1}},
		{"medium speed", 125, [2]float64{0.3, 0.7}},
		{"max speed", 200, [2]float64{0.9, 1.0}},
		{"above max", 300, [2]float64{1.0, 1.0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})
			entity.AddComponent(&VelocityComponent{VX: tt.speed, VY: 0})

			sys.Update([]*Entity{entity}, 0.5) // Large dt to ensure past cooldown

			comp, ok := entity.GetComponent("movement_dust")
			if !ok {
				t.Fatal("expected dust component")
			}
			dc := comp.(*MovementDustComponent)
			if dc.Intensity < tt.wantRange[0] || dc.Intensity > tt.wantRange[1] {
				t.Errorf("intensity = %f, want [%f, %f]", dc.Intensity, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

func TestMovementDustNoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 200, VY: 0})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestMovementDustNoVelocity(t *testing.T) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func BenchmarkMovementDustSystemUpdate(b *testing.B) {
	world := NewWorld()
	sys := NewMovementDustSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entities := make([]*Entity, 100)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		e.AddComponent(&VelocityComponent{VX: 100 + float64(i), VY: 50})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
