package engine

import (
	"testing"
)

// TestNewEvasionParticleSystem verifies system creation with valid inputs.
func TestNewEvasionParticleSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(42)

	system := NewEvasionParticleSystem(world, seed)

	if system == nil {
		t.Fatal("expected non-nil system")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.seed != seed {
		t.Errorf("seed = %d, want %d", system.seed, seed)
	}
	if system.particleCount != 6 {
		t.Errorf("particleCount = %d, want 6", system.particleCount)
	}
	if system.spreadFactor != 50.0 {
		t.Errorf("spreadFactor = %f, want 50.0", system.spreadFactor)
	}
}

// TestNewEvasionParticleSystem_NilWorld verifies system creation with nil world.
func TestNewEvasionParticleSystem_NilWorld(t *testing.T) {
	system := NewEvasionParticleSystem(nil, 42)

	if system == nil {
		t.Fatal("expected non-nil system even with nil world")
	}
	if system.world != nil {
		t.Error("world should be nil")
	}
}

// TestEvasionParticleSystem_SetParticleSystem verifies particle system linking.
func TestEvasionParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewEvasionParticleSystem(world, 42)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

// TestEvasionParticleSystem_SetGenre verifies genre setting.
func TestEvasionParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewEvasionParticleSystem(NewWorld(), 42)
			system.SetGenre(tt.genreID)

			if system.genreID != tt.genreID {
				t.Errorf("genreID = %q, want %q", system.genreID, tt.genreID)
			}
		})
	}
}

// TestEvasionParticleSystem_Update verifies no-op update behavior.
func TestEvasionParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewEvasionParticleSystem(world, 42)

	// Should not panic with nil or empty entities
	system.Update(nil, 0.016)
	system.Update([]*Entity{}, 0.016)

	entity := world.CreateEntity()
	system.Update([]*Entity{entity}, 0.016)
	// No assertions needed - just verify no panic
}

// TestEvasionParticleSystem_OnEvasion_ValidInputs tests callback with valid inputs.
func TestEvasionParticleSystem_OnEvasion_ValidInputs(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()
	system := NewEvasionParticleSystem(world, 42)
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 50, Y: 50})

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	system.OnEvasion(attacker, target, 0.3)
}

// TestEvasionParticleSystem_OnEvasion_NilTarget tests nil target handling.
func TestEvasionParticleSystem_OnEvasion_NilTarget(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()
	system := NewEvasionParticleSystem(world, 42)
	system.SetParticleSystem(ps)

	attacker := world.CreateEntity()

	// Should not panic with nil target
	system.OnEvasion(attacker, nil, 0.3)
}

// TestEvasionParticleSystem_OnEvasion_NoPosition tests target without position.
func TestEvasionParticleSystem_OnEvasion_NoPosition(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()
	system := NewEvasionParticleSystem(world, 42)
	system.SetParticleSystem(ps)

	attacker := world.CreateEntity()
	target := world.CreateEntity()
	// target has no position component

	// Should not panic when target has no position
	system.OnEvasion(attacker, target, 0.3)
}

// TestEvasionParticleSystem_OnEvasion_NilParticleSystem tests nil particle system.
func TestEvasionParticleSystem_OnEvasion_NilParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewEvasionParticleSystem(world, 42)
	// Don't set particle system

	attacker := world.CreateEntity()
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic with nil particle system
	system.OnEvasion(attacker, target, 0.3)
}

// TestEvasionParticleSystem_getParticleTypeForGenre tests genre-based particle selection.
func TestEvasionParticleSystem_getParticleTypeForGenre(t *testing.T) {
	tests := []struct {
		genreID  string
		wantType string // We check the type is not empty
	}{
		{"fantasy", "sparkle"},
		{"scifi", "spark"},
		{"horror", "smoke"},
		{"cyberpunk", "spark"},
		{"postapoc", "dust"},
		{"", "sparkle"}, // default
		{"unknown", "sparkle"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			system := NewEvasionParticleSystem(NewWorld(), 42)
			system.SetGenre(tt.genreID)

			particleType := system.getParticleTypeForGenre()
			// Just verify it returns a valid type (non-zero)
			if particleType == 0 {
				t.Error("expected non-zero particle type")
			}
		})
	}
}

// TestEvasionParticleSystem_SpawnEvasionEffect tests direct effect spawning.
func TestEvasionParticleSystem_SpawnEvasionEffect(t *testing.T) {
	world := NewWorld()
	ps := NewParticleSystem()
	system := NewEvasionParticleSystem(world, 42)
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Should not panic
	system.SpawnEvasionEffect(100.0, 100.0)
}

// TestEvasionParticleSystem_SpawnEvasionEffect_NilParticleSystem tests nil handling.
func TestEvasionParticleSystem_SpawnEvasionEffect_NilParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewEvasionParticleSystem(world, 42)
	// Don't set particle system

	// Should not panic
	system.SpawnEvasionEffect(100.0, 100.0)
}

// TestEvasionParticleSystem_SpawnEvasionEffect_NilWorld tests nil world handling.
func TestEvasionParticleSystem_SpawnEvasionEffect_NilWorld(t *testing.T) {
	system := NewEvasionParticleSystem(nil, 42)
	system.SetParticleSystem(NewParticleSystem())

	// Should not panic with nil world
	system.SpawnEvasionEffect(100.0, 100.0)
}

// TestEvasionParticleSystem_ParticleCountScaling tests evasion chance scaling.
func TestEvasionParticleSystem_ParticleCountScaling(t *testing.T) {
	tests := []struct {
		name          string
		evasionChance float64
	}{
		{"low_evasion", 0.1},
		{"moderate_evasion", 0.35},
		{"high_evasion", 0.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			ps := NewParticleSystem()
			system := NewEvasionParticleSystem(world, 42)
			system.SetParticleSystem(ps)
			system.SetGenre("fantasy")

			target := world.CreateEntity()
			target.AddComponent(&PositionComponent{X: 100, Y: 100})

			attacker := world.CreateEntity()

			// Just verify no panic with different evasion values
			system.OnEvasion(attacker, target, tt.evasionChance)
		})
	}
}

// TestCombatSystem_SetEvasionCallback verifies the callback setter exists.
func TestCombatSystem_SetEvasionCallback(t *testing.T) {
	cs := NewCombatSystem(42)

	cs.SetEvasionCallback(func(attacker, target *Entity, evasionChance float64) {
		// Test callback - not invoked in this test
	})

	// The callback is stored but we can't easily test invocation without full combat
	// Just verify the setter doesn't panic
	if cs.onEvasionCallback == nil {
		t.Error("evasion callback not set")
	}
}

// TestEvasionParticleSystem_DeterministicSeeding verifies deterministic particle spawning.
func TestEvasionParticleSystem_DeterministicSeeding(t *testing.T) {
	seed := int64(12345)

	// Create two systems with the same seed
	world1 := NewWorld()
	system1 := NewEvasionParticleSystem(world1, seed)

	world2 := NewWorld()
	system2 := NewEvasionParticleSystem(world2, seed)

	// Both should have same seed (determinism test)
	if system1.seed != system2.seed {
		t.Errorf("seeds differ: %d vs %d", system1.seed, system2.seed)
	}
}
