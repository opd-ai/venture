package engine

import (
	"testing"
)

// StubPlayerTagComponent is a test stub for marking entities as players.
type StubPlayerTagComponent struct{}

func (s *StubPlayerTagComponent) Type() string { return "input" }

func TestNewEntityIdleAmbientParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 42 {
		t.Errorf("seed = %d, want 42", sys.seed)
	}
	if sys.idleThreshold != 1.5 {
		t.Errorf("idleThreshold = %f, want 1.5", sys.idleThreshold)
	}
	if sys.spawnInterval != 1.2 {
		t.Errorf("spawnInterval = %f, want 1.2", sys.spawnInterval)
	}
	if sys.entityStates == nil {
		t.Error("entityStates map should be initialized")
	}
}

func TestEntityIdleAmbientParticleSystem_NilGuards(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)

	// No particle system set — should not panic
	entities := []*Entity{world.CreateEntity()}
	sys.Update(entities, 0.016)

	// Nil world — should not panic
	sys2 := NewEntityIdleAmbientParticleSystem(nil, 42)
	sys2.Update(entities, 0.016)
}

func TestEntityIdleAmbientParticleSystem_IdleTracking(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	entities := []*Entity{entity}

	// Update for less than idle threshold
	sys.Update(entities, 0.5)
	if dur := sys.GetIdleDuration(entity.ID); dur < 0.4 || dur > 0.6 {
		t.Errorf("idle duration after 0.5s = %f, want ~0.5", dur)
	}

	// Still below threshold — no spawn yet
	sys.Update(entities, 0.5)
	if dur := sys.GetIdleDuration(entity.ID); dur < 0.9 || dur > 1.1 {
		t.Errorf("idle duration after 1.0s = %f, want ~1.0", dur)
	}
}

func TestEntityIdleAmbientParticleSystem_MovementResetsIdle(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	vel := &VelocityComponent{VX: 0, VY: 0}
	entity.AddComponent(vel)

	entities := []*Entity{entity}

	// Accumulate idle time
	sys.Update(entities, 1.0)
	if dur := sys.GetIdleDuration(entity.ID); dur < 0.9 {
		t.Errorf("expected idle duration ~1.0, got %f", dur)
	}

	// Start moving — should reset
	vel.VX = 10.0
	vel.VY = 10.0
	sys.Update(entities, 0.1)
	if dur := sys.GetIdleDuration(entity.ID); dur != 0 {
		t.Errorf("expected idle duration 0 after moving, got %f", dur)
	}
}

func TestEntityIdleAmbientParticleSystem_SkipsEntitiesWithoutHealth(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	// No health component

	entities := []*Entity{entity}
	sys.Update(entities, 2.0)

	// Should not have created state for entity without health
	if sys.GetEntityStateCount() != 0 {
		t.Errorf("expected 0 tracked entities, got %d", sys.GetEntityStateCount())
	}
}

func TestEntityIdleAmbientParticleSystem_SkipsEntitiesWithoutPosition(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	// No position component

	entities := []*Entity{entity}
	sys.Update(entities, 2.0)

	if sys.GetEntityStateCount() != 0 {
		t.Errorf("expected 0 tracked entities, got %d", sys.GetEntityStateCount())
	}
}

func TestEntityIdleAmbientParticleSystem_GenreConfigs(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"horror", "horror"},
		{"scifi", "scifi"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"unknown defaults to fantasy", "unknown_genre"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewEntityIdleAmbientParticleSystem(world, 42)
			sys.SetGenre(tt.genreID)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 50, Y: 50})
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

			config := sys.getIdleParticleConfig(entity, 50, 50)
			if config == nil {
				t.Errorf("expected non-nil particle config for genre %q", tt.genreID)
				return
			}
			if config.Count < 1 {
				t.Errorf("expected count >= 1, got %d", config.Count)
			}
			if config.Duration <= 0 {
				t.Errorf("expected positive duration, got %f", config.Duration)
			}
		})
	}
}

func TestEntityIdleAmbientParticleSystem_PlayerGetsMoreParticles(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)
	sys.SetGenre("fantasy")

	// Non-player entity
	npc := world.CreateEntity()
	npc.AddComponent(&PositionComponent{X: 50, Y: 50})
	npc.AddComponent(&HealthComponent{Current: 100, Max: 100})
	npcConfig := sys.getIdleParticleConfig(npc, 50, 50)

	// Player entity
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&StubPlayerTagComponent{})
	playerConfig := sys.getIdleParticleConfig(player, 50, 50)

	if npcConfig == nil || playerConfig == nil {
		t.Fatal("expected non-nil configs")
	}
	if playerConfig.Count <= npcConfig.Count {
		t.Errorf("player count (%d) should be > npc count (%d)", playerConfig.Count, npcConfig.Count)
	}
}

func TestEntityIdleAmbientParticleSystem_PruneStaleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	e1 := world.CreateEntity()
	e1.AddComponent(&PositionComponent{X: 10, Y: 10})
	e1.AddComponent(&HealthComponent{Current: 100, Max: 100})
	e1.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	e2 := world.CreateEntity()
	e2.AddComponent(&PositionComponent{X: 20, Y: 20})
	e2.AddComponent(&HealthComponent{Current: 100, Max: 100})
	e2.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	entities := []*Entity{e1, e2}
	sys.Update(entities, 0.5)

	if sys.GetEntityStateCount() != 2 {
		t.Errorf("expected 2 tracked entities, got %d", sys.GetEntityStateCount())
	}

	// Prune with only e1 active
	sys.pruneStaleEntities([]*Entity{e1})
	if sys.GetEntityStateCount() != 1 {
		t.Errorf("expected 1 tracked entity after prune, got %d", sys.GetEntityStateCount())
	}
}

func TestEntityIdleAmbientParticleSystem_SpeedThreshold(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)

	threshold := sys.idleParticleSpeedThreshold()
	if threshold != 2.0 {
		t.Errorf("speed threshold = %f, want 2.0", threshold)
	}
}

func TestEntityIdleAmbientParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genreID = %q, want horror", sys.genreID)
	}
}

func TestEntityIdleAmbientParticleSystem_NoVelocityComponent(t *testing.T) {
	world := NewWorld()
	sys := NewEntityIdleAmbientParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	// Entity with no velocity component should still be treated as idle
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entities := []*Entity{entity}
	sys.Update(entities, 2.0)

	if dur := sys.GetIdleDuration(entity.ID); dur < 1.9 {
		t.Errorf("expected idle duration ~2.0 for entity with no velocity, got %f", dur)
	}
}
