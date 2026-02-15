package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewSprintTrailParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSprintTrailParticleSystem(world, 42)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.world != world {
		t.Error("world not set")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected default genre fantasy, got %s", sys.genreID)
	}
	if sys.sprintSpeedSq != 10000.0 {
		t.Errorf("expected sprint threshold 10000, got %f", sys.sprintSpeedSq)
	}
}

func TestSprintTrailGenrePresets(t *testing.T) {
	world := NewWorld()
	sys := NewSprintTrailParticleSystem(world, 42)

	tests := []struct {
		name         string
		genre        string
		wantType     particles.ParticleType
		wantPositive bool // gravity > 0 means particles settle
	}{
		{"fantasy sparkles", "fantasy", particles.ParticleSparkle, false},
		{"horror smoke", "horror", particles.ParticleSmoke, false},
		{"cyberpunk sparks", "cyberpunk", particles.ParticleSpark, false},
		{"scifi magic", "scifi", particles.ParticleMagic, false},
		{"postapoc dust", "postapoc", particles.ParticleDust, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("genre not set: got %s", sys.genreID)
			}
			preset := sys.getGenrePreset(tt.genre)
			if preset.ParticleType != tt.wantType {
				t.Errorf("particle type: got %v, want %v", preset.ParticleType, tt.wantType)
			}
			if tt.wantPositive && preset.Gravity <= 0 {
				t.Errorf("expected positive gravity for %s, got %f", tt.genre, preset.Gravity)
			}
			if preset.Count <= 0 {
				t.Error("particle count must be positive")
			}
			if preset.Duration <= 0 {
				t.Error("duration must be positive")
			}
			if preset.MinSize <= 0 || preset.MaxSize <= 0 {
				t.Error("sizes must be positive")
			}
			if preset.MinSize > preset.MaxSize {
				t.Error("min size should not exceed max size")
			}
		})
	}
}

func TestSprintTrailUpdateNilDeps(t *testing.T) {
	world := NewWorld()
	sys := NewSprintTrailParticleSystem(world, 42)

	// Should not panic with nil particle system
	entities := []*Entity{world.CreateEntity()}
	sys.Update(entities, 0.016)

	// Should not panic with nil world
	sys2 := NewSprintTrailParticleSystem(nil, 42)
	sys2.particleSystem = &ParticleSystem{}
	sys2.Update(entities, 0.016)
}

func TestSprintTrailIgnoresSlowEntities(t *testing.T) {
	world := NewWorld()
	sys := NewSprintTrailParticleSystem(world, 42)
	// Don't set particleSystem - Update will return early for fast entities
	// but slow entities are filtered before particle spawn

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	// Speed 50 px/s = speedSq 2500, below threshold 10000
	entity.AddComponent(&VelocityComponent{VX: 30, VY: 40})

	// Use a nil particleSystem - slow entities are filtered before spawn is called
	sys.particleSystem = &ParticleSystem{}
	sys.Update([]*Entity{entity}, 0.016)

	// Entity should not be tracked (below sprint threshold)
	if _, exists := sys.lastSpawn[entity.ID]; exists {
		t.Error("slow entity should not have spawn tracking")
	}
}

func TestSprintTrailSpawnThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewSprintTrailParticleSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	// Speed 150 px/s = speedSq 22500, above threshold 10000
	entity.AddComponent(&VelocityComponent{VX: 90, VY: 120})

	// Set initial cooldown so it doesn't try to actually spawn (no generator)
	sys.lastSpawn[entity.ID] = 0
	sys.particleSystem = &ParticleSystem{}

	// With dt=0.01, timer accumulates but stays below interval (0.08)
	// so SpawnParticles won't be called (avoids nil generator panic)
	sys.Update([]*Entity{entity}, 0.01)
	timer := sys.lastSpawn[entity.ID]
	if timer < 0.01 {
		t.Errorf("expected timer >= 0.01, got %f", timer)
	}
}

func TestSprintTrailSetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSprintTrailParticleSystem(world, 42)

	ps := &ParticleSystem{}
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestSprintTrailCleansUpOnStop(t *testing.T) {
	world := NewWorld()
	sys := NewSprintTrailParticleSystem(world, 42)
	sys.particleSystem = &ParticleSystem{}

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&VelocityComponent{VX: 90, VY: 120})

	// Manually populate tracking (don't call Update which would try to spawn)
	sys.lastSpawn[entity.ID] = 0.05
	if _, exists := sys.lastSpawn[entity.ID]; !exists {
		t.Fatal("expected entity to be tracked")
	}

	// Entity slows down - should be cleaned up
	entity.AddComponent(&VelocityComponent{VX: 5, VY: 5})
	sys.Update([]*Entity{entity}, 0.1)
	if _, exists := sys.lastSpawn[entity.ID]; exists {
		t.Error("slow entity should have been removed from tracking")
	}
}
