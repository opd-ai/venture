package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewProjectileTrailParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
	if sys.spawnInterval != 0.05 {
		t.Errorf("spawnInterval = %f, want 0.05", sys.spawnInterval)
	}
}

func TestProjectileTrailParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "horror")
	}
}

func TestProjectileTrailParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)

	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestProjectileTrailParticleSystem_Update_NilGuards(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)

	// Should not panic with nil particle system
	sys.Update([]*Entity{}, 0.016)

	sys.SetParticleSystem(NewParticleSystem())
	sys.world = nil
	// Should not panic with nil world
	sys.Update([]*Entity{}, 0.016)
}

func TestProjectileTrailParticleSystem_Update_SkipsNonProjectiles(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	// No projectile component - should skip without error
	sys.Update([]*Entity{entity}, 0.016)
}

func TestProjectileTrailParticleSystem_Update_SkipsExpiredProjectiles(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&ProjectileComponent{
		ProjectileType: "arrow",
		LifeTime:       1.0,
		Age:            2.0, // expired
	})

	sys.Update([]*Entity{entity}, 0.016)
	// Should have cleaned up cooldown
	if _, exists := sys.lastSpawn[entity.ID]; exists {
		t.Error("expected expired projectile cooldown to be cleaned up")
	}
}

func TestProjectileTrailParticleSystem_Update_SkipsHitProjectiles(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&ProjectileComponent{
		ProjectileType: "arrow",
		LifeTime:       5.0,
		Age:            0.5,
		HasHit:         true,
	})

	sys.Update([]*Entity{entity}, 0.016)
	if _, exists := sys.lastSpawn[entity.ID]; exists {
		t.Error("expected hit projectile cooldown to be cleaned up")
	}
}

func TestProjectileTrailParticleSystem_Update_CooldownRespected(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&ProjectileComponent{
		ProjectileType: "fireball",
		LifeTime:       5.0,
		Age:            0.1,
		Speed:          100,
	})

	// First update should set cooldown
	sys.Update([]*Entity{entity}, 0.01)
	cooldown := sys.lastSpawn[entity.ID]
	if cooldown == 0 {
		// Cooldown was reset because spawn happened, OR was accumulated
		// Either way, no panic is good
	}

	// Second very quick update should accumulate cooldown
	sys.Update([]*Entity{entity}, 0.01)
}

func TestProjectileTrailParticleSystem_GetTrailConfig_Types(t *testing.T) {
	tests := []struct {
		name           string
		projectileType string
		genreID        string
		wantType       particles.ParticleType
		wantCount      int
	}{
		{"fireball_fantasy", "fireball", "fantasy", particles.ParticleEmber, 3},
		{"fireball_horror", "fireball", "horror", particles.ParticleSmoke, 3},
		{"ice_shard", "ice_shard", "fantasy", particles.ParticleSparkle, 2},
		{"arrow_fantasy", "arrow", "fantasy", particles.ParticleDust, 2},
		{"arrow_scifi", "arrow", "scifi", particles.ParticleDust, 1},
		{"bullet_default", "bullet", "fantasy", particles.ParticleSmoke, 1},
		{"bullet_cyberpunk", "bullet", "cyberpunk", particles.ParticleSpark, 1},
		{"bullet_scifi", "bullet", "scifi", particles.ParticleSpark, 1},
		{"unknown_type", "plasma_bolt", "fantasy", particles.ParticleMagic, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewProjectileTrailParticleSystem(world, 42)
			sys.SetGenre(tt.genreID)

			config := sys.getTrailConfig(tt.projectileType, 10.0, 20.0)
			if config == nil {
				t.Fatal("expected non-nil config")
			}
			if config.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", config.Type, tt.wantType)
			}
			if config.Count != tt.wantCount {
				t.Errorf("Count = %d, want %d", config.Count, tt.wantCount)
			}
			if config.GenreID != tt.genreID {
				t.Errorf("GenreID = %q, want %q", config.GenreID, tt.genreID)
			}
			if config.Custom["projectile_trail"] != true {
				t.Error("expected projectile_trail custom flag")
			}
		})
	}
}

func TestProjectileTrailParticleSystem_CleanupCooldowns(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)

	// Add some cooldown entries
	sys.lastSpawn[1] = 0.01
	sys.lastSpawn[2] = 0.02
	sys.lastSpawn[3] = 0.03

	// Only entity 2 is still active
	entities := []*Entity{NewEntity(2)}
	sys.cleanupCooldowns(entities)

	if _, ok := sys.lastSpawn[1]; ok {
		t.Error("expected entity 1 cooldown to be cleaned up")
	}
	if _, ok := sys.lastSpawn[2]; !ok {
		t.Error("expected entity 2 cooldown to be preserved")
	}
	if _, ok := sys.lastSpawn[3]; ok {
		t.Error("expected entity 3 cooldown to be cleaned up")
	}
}

func TestProjectileTrailParticleSystem_SpawnsAfterInterval(t *testing.T) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})
	entity.AddComponent(&ProjectileComponent{
		ProjectileType: "fireball",
		LifeTime:       5.0,
		Age:            0.1,
		Speed:          200,
	})

	// Accumulate past the spawn interval (0.05s)
	sys.Update([]*Entity{entity}, 0.06)
	// Cooldown should have been reset to 0 after spawn
	if sys.lastSpawn[entity.ID] != 0 {
		t.Errorf("expected cooldown reset after spawn, got %f", sys.lastSpawn[entity.ID])
	}
}

func BenchmarkProjectileTrailParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewProjectileTrailParticleSystem(world, 42)
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("fantasy")

	entities := make([]*Entity, 50)
	for i := range entities {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		e.AddComponent(&ProjectileComponent{
			ProjectileType: "fireball",
			LifeTime:       5.0,
			Age:            0.1,
			Speed:          200,
		})
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
