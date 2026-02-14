//go:build ignore

package engine

import (
	"testing"
)

func TestNewReputationDefenseBonusParticleSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	system := NewReputationDefenseBonusParticleSystem(world, seed)

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
	if system.spreadFactor != 35.0 {
		t.Errorf("spreadFactor = %f, want 35.0", system.spreadFactor)
	}
	if system.minAbsorbed != 1.0 {
		t.Errorf("minAbsorbed = %f, want 1.0", system.minAbsorbed)
	}
}

func TestReputationDefenseBonusParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestReputationDefenseBonusParticleSystem_SetGenre(t *testing.T) {
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
			system := NewReputationDefenseBonusParticleSystem(world, 12345)

			system.SetGenre(tt.genreID)

			if system.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", system.genreID, tt.genreID)
			}
		})
	}
}

func TestReputationDefenseBonusParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)

	// Update is a no-op, just verify it doesn't panic
	system.Update(nil, 0.016)
	system.Update([]*Entity{}, 0.016)
}

func TestReputationDefenseBonusParticleSystem_OnReputationDefense_NilChecks(t *testing.T) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)

	// No particle system — should not panic
	attacker := world.CreateEntity()
	defender := world.CreateEntity()
	system.OnReputationDefense(defender, attacker, 20.0, 18.0, 0.05)

	// Nil defender — should not panic
	system.SetParticleSystem(NewParticleSystem())
	system.OnReputationDefense(nil, attacker, 20.0, 18.0, 0.05)
}

func TestReputationDefenseBonusParticleSystem_OnReputationDefense_BelowMinAbsorbed(t *testing.T) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")

	defender := world.CreateEntity()
	defender.AddComponent(&PositionComponent{X: 50, Y: 50})

	// Absorbed = 20.0 - 19.5 = 0.5, below minAbsorbed=1.0
	system.OnReputationDefense(defender, nil, 20.0, 19.5, 0.01)
	// Should not panic, particles not triggered
}

func TestReputationDefenseBonusParticleSystem_OnReputationDefense_NoPosition(t *testing.T) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")

	defender := world.CreateEntity()
	// No position component

	system.OnReputationDefense(defender, nil, 20.0, 15.0, 0.08)
	// Should return early without panic
}

func TestReputationDefenseBonusParticleSystem_OnReputationDefense_WithPosition(t *testing.T) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")

	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})

	defender := world.CreateEntity()
	defender.AddComponent(&PositionComponent{X: 100, Y: 100})

	// 5 damage absorbed — should spawn particles
	system.OnReputationDefense(defender, attacker, 25.0, 20.0, 0.08)
}

func TestReputationDefenseBonusParticleSystem_OnReputationDefense_HighAbsorb(t *testing.T) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("scifi")

	defender := world.CreateEntity()
	defender.AddComponent(&PositionComponent{X: 50, Y: 50})

	// Large absorbed damage should scale up particle count
	system.OnReputationDefense(defender, nil, 100.0, 60.0, 0.12)
}

func TestReputationDefenseBonusParticleSystem_getParticleTypeForGenre(t *testing.T) {
	tests := []struct {
		genreID      string
		expectedType string
	}{
		{"fantasy", "sparkle"},
		{"scifi", "spark"},
		{"horror", "smoke"},
		{"cyberpunk", "spark"},
		{"postapoc", "dust"},
		{"unknown", "sparkle"},
		{"", "sparkle"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			world := NewWorld()
			system := NewReputationDefenseBonusParticleSystem(world, 12345)
			system.SetGenre(tt.genreID)

			particleType := system.getParticleTypeForGenre()

			if string(particleType) != tt.expectedType {
				t.Errorf("getParticleTypeForGenre() = %s, want %s", particleType, tt.expectedType)
			}
		})
	}
}

func TestReputationDefenseBonusParticleSystem_SpawnDefenseEffect(t *testing.T) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)

	// Should not panic without particle system
	system.SpawnDefenseEffect(100, 100)

	// Should work with particle system
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")
	system.SpawnDefenseEffect(100, 100)
}

func TestReputationDefenseBonusParticleSystem_DeterministicSeeding(t *testing.T) {
	world := NewWorld()
	seed := int64(99999)

	s1 := NewReputationDefenseBonusParticleSystem(world, seed)
	s2 := NewReputationDefenseBonusParticleSystem(world, seed)

	if s1.seed != s2.seed {
		t.Error("same seed should produce same system seed")
	}
}

func BenchmarkReputationDefenseBonusParticleSystem_OnReputationDefense(b *testing.B) {
	world := NewWorld()
	system := NewReputationDefenseBonusParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetGenre("fantasy")

	defender := world.CreateEntity()
	defender.AddComponent(&PositionComponent{X: 100, Y: 100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnReputationDefense(defender, nil, 30.0, 25.0, 0.08)
	}
}
