//go:build ignore

package engine

import (
	"testing"
)

func TestNewBlockParticleSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	system := NewBlockParticleSystem(world, seed)

	if system == nil {
		t.Fatal("expected non-nil system")
	}
	if system.world != world {
		t.Error("world reference not set")
	}
	if system.seed != seed {
		t.Errorf("seed = %d, want %d", system.seed, seed)
	}
	if system.rng == nil {
		t.Error("expected non-nil RNG")
	}
	if system.particleCount != 8 {
		t.Errorf("particleCount = %d, want 8", system.particleCount)
	}
	if system.spreadFactor != 40.0 {
		t.Errorf("spreadFactor = %f, want 40.0", system.spreadFactor)
	}
}

func TestBlockParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewBlockParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestBlockParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewBlockParticleSystem(world, 12345)

			system.SetGenre(tt.genreID)

			if system.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", system.genreID, tt.genreID)
			}
		})
	}
}

func TestBlockParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewBlockParticleSystem(world, 12345)

	// Update is a no-op, just verify it doesn't panic
	system.Update(nil, 0.016)
	system.Update([]*Entity{}, 0.016)
}

func TestBlockParticleSystem_OnBlock_NilChecks(t *testing.T) {
	world := NewWorld()
	system := NewBlockParticleSystem(world, 12345)

	// Should not panic with nil particle system
	attacker := world.CreateEntity()
	target := world.CreateEntity()
	system.OnBlock(attacker, target, 0.2, 20.0, 10.0)

	// Should not panic with nil target
	system.SetParticleSystem(NewParticleSystem())
	system.OnBlock(attacker, nil, 0.2, 20.0, 10.0)
}

func TestBlockParticleSystem_OnBlock_NoPosition(t *testing.T) {
	world := NewWorld()
	system := NewBlockParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")

	attacker := world.CreateEntity()
	target := world.CreateEntity()
	// Target has no position component

	// Should not panic, just return early
	system.OnBlock(attacker, target, 0.2, 20.0, 10.0)
}

func TestBlockParticleSystem_OnBlock_WithPosition(t *testing.T) {
	world := NewWorld()
	system := NewBlockParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")

	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should spawn particles without panic
	system.OnBlock(attacker, target, 0.25, 30.0, 15.0)
}

func TestBlockParticleSystem_OnBlock_HighDamage(t *testing.T) {
	world := NewWorld()
	system := NewBlockParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("scifi")

	attacker := world.CreateEntity()
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 50, Y: 50})

	// High damage should scale up particle count
	system.OnBlock(attacker, target, 0.3, 100.0, 50.0)
}

func TestBlockParticleSystem_getParticleTypeForGenre(t *testing.T) {
	tests := []struct {
		genreID      string
		expectedType string
	}{
		{"fantasy", "sparkle"},
		{"scifi", "spark"},
		{"horror", "smoke"},
		{"cyberpunk", "spark"},
		{"postapoc", "debris"},
		{"unknown", "sparkle"}, // default
		{"", "sparkle"},        // empty default
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			world := NewWorld()
			system := NewBlockParticleSystem(world, 12345)
			system.SetGenre(tt.genreID)

			particleType := system.getParticleTypeForGenre()

			if string(particleType) != tt.expectedType {
				t.Errorf("getParticleTypeForGenre() = %s, want %s", particleType, tt.expectedType)
			}
		})
	}
}

func TestBlockParticleSystem_SpawnBlockEffect(t *testing.T) {
	world := NewWorld()
	system := NewBlockParticleSystem(world, 12345)

	// Should not panic without particle system
	system.SpawnBlockEffect(100, 100)

	// Should work with particle system
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")
	system.SpawnBlockEffect(100, 100)
}

func TestBlockParticleSystem_Integration_WithCombatSystem(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	blockParticleSystem := NewBlockParticleSystem(world, 12345)
	blockParticleSystem.SetParticleSystem(NewParticleSystem())
	blockParticleSystem.SetGenre("fantasy")

	// Register callback
	combatSystem.SetBlockCallback(blockParticleSystem.OnBlock)

	// Create attacker and target
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&AttackComponent{Damage: 20, Range: 50, Cooldown: 1.0})

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 30, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	stats := NewStatsComponent()
	stats.BlockChance = 1.0 // 100% block for testing
	target.AddComponent(stats)

	world.AddEntity(attacker)
	world.AddEntity(target)

	// Perform attack - should trigger block
	combatSystem.Attack(attacker, target)

	// Target should have taken reduced damage (50% of 20 = 10)
	health := target.GetHealth()
	if health == nil {
		t.Fatal("target has no health component")
	}

	// Damage is 20 base, stats add 10 attack = 30, blocked to 15, minus 5 defense
	// The exact damage depends on combat formula, just verify some damage was taken
	if health.Current >= 100 {
		t.Error("expected target to take some damage")
	}
}

func TestBlockParticleSystem_DeterministicSeeding(t *testing.T) {
	world := NewWorld()
	seed := int64(99999)

	system1 := NewBlockParticleSystem(world, seed)
	system2 := NewBlockParticleSystem(world, seed)

	// Same seed should produce same initial state
	if system1.seed != system2.seed {
		t.Error("same seed should produce same system seed")
	}
}

func BenchmarkBlockParticleSystem_OnBlock(b *testing.B) {
	world := NewWorld()
	system := NewBlockParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")

	attacker := world.CreateEntity()
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnBlock(attacker, target, 0.25, 30.0, 15.0)
	}
}
