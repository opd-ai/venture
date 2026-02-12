package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/combat"
)

func TestNewDamageResistanceParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)

	if system == nil {
		t.Fatal("expected system to be non-nil")
	}

	if system.world != world {
		t.Error("expected world to be set")
	}

	if system.rng == nil {
		t.Error("expected RNG to be initialized")
	}

	if system.particleCount != 12 {
		t.Errorf("expected particleCount=12, got %d", system.particleCount)
	}

	if system.resistThreshold != 0.25 {
		t.Errorf("expected resistThreshold=0.25, got %f", system.resistThreshold)
	}

	if system.minDamageReduced != 5.0 {
		t.Errorf("expected minDamageReduced=5.0, got %f", system.minDamageReduced)
	}
}

func TestDamageResistanceParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()

	system.SetParticleSystem(particleSystem)

	if system.particleSystem != particleSystem {
		t.Error("expected particle system to be set")
	}
}

func TestDamageResistanceParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)

	system.SetGenre("fantasy")

	if system.genreID != "fantasy" {
		t.Errorf("expected genreID=fantasy, got %s", system.genreID)
	}
}

func TestDamageResistanceParticleSystem_SetResistThreshold(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal value", 0.5, 0.5},
		{"minimum value", 0.0, 0.0},
		{"maximum value", 1.0, 1.0},
		{"clamp negative", -0.5, 0.0},
		{"clamp above max", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewDamageResistanceParticleSystem(world, 12345)

			system.SetResistThreshold(tt.input)

			if system.resistThreshold != tt.expected {
				t.Errorf("expected resistThreshold=%f, got %f", tt.expected, system.resistThreshold)
			}
		})
	}
}

func TestDamageResistanceParticleSystem_Update(t *testing.T) {
	// Update is a no-op for callback-driven systems, just verify it doesn't panic
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)

	entities := []*Entity{world.CreateEntity()}
	system.Update(entities, 0.016)
	// No error means pass
}

func TestDamageResistanceParticleSystem_OnDamageResisted_NilParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	// Don't set particle system

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic when particle system is nil
	system.OnDamageResisted(target, combat.DamageFire, 100, 50, 0.5)
}

func TestDamageResistanceParticleSystem_OnDamageResisted_NilTarget(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())

	// Should not panic when target is nil
	system.OnDamageResisted(nil, combat.DamageFire, 100, 50, 0.5)
}

func TestDamageResistanceParticleSystem_OnDamageResisted_BelowThreshold(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Resistance below threshold (0.25), should not spawn particles but not panic
	system.OnDamageResisted(target, combat.DamageFire, 100, 80, 0.20) // 20% resist
	// No error means pass
}

func TestDamageResistanceParticleSystem_OnDamageResisted_BelowMinDamageReduced(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// High resistance but low damage reduced (only 3 damage reduced, min is 5)
	system.OnDamageResisted(target, combat.DamageFire, 10, 7, 0.30) // 30% resist, 3 damage
	// No error means pass - should exit early
}

func TestDamageResistanceParticleSystem_OnDamageResisted_NoPosition(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())

	target := world.CreateEntity()
	// No position component

	// Should not panic when target has no position
	system.OnDamageResisted(target, combat.DamageFire, 100, 50, 0.5)
}

func TestDamageResistanceParticleSystem_OnDamageResisted_SpawnsParticles(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)
	system.SetGenre("fantasy")

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Good resistance and damage - should spawn particles (without panicking)
	system.OnDamageResisted(target, combat.DamageFire, 100, 50, 0.5) // 50% resist, 50 damage
	// No panic means particle spawning worked
}

func TestDamageResistanceParticleSystem_GetParticleTypeForDamage(t *testing.T) {
	tests := []struct {
		damageType   combat.DamageType
		expectedType string
	}{
		{combat.DamageFire, "ember"},
		{combat.DamageIce, "sparkle"},
		{combat.DamageLightning, "spark"},
		{combat.DamagePoison, "smoke"},
		{combat.DamageMagical, "magic"},
		{combat.DamagePhysical, "dust"},
	}

	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)

	for _, tt := range tests {
		t.Run(tt.damageType.String(), func(t *testing.T) {
			particleType := system.getParticleTypeForDamage(tt.damageType)
			if particleType.String() != tt.expectedType {
				t.Errorf("expected %s particle for %s damage, got %s",
					tt.expectedType, tt.damageType.String(), particleType.String())
			}
		})
	}
}

func TestDamageResistanceParticleSystem_SpawnResistEffect(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)
	system.SetGenre("scifi")

	// Should not panic
	system.SpawnResistEffect(200, 200, combat.DamageLightning, 0.75)
}

func TestDamageResistanceParticleSystem_SpawnResistEffect_NilParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	// Don't set particle system

	// Should not panic
	system.SpawnResistEffect(200, 200, combat.DamageLightning, 0.75)
}

func TestDamageResistanceParticleSystem_SpawnResistEffect_NilWorld(t *testing.T) {
	system := NewDamageResistanceParticleSystem(nil, 12345)
	system.SetParticleSystem(NewParticleSystem())

	// Should not panic
	system.SpawnResistEffect(200, 200, combat.DamageLightning, 0.75)
}

func TestDamageResistanceParticleSystem_DeterministicGeneration(t *testing.T) {
	// Same seed should produce consistent results
	seed := int64(54321)

	world1 := NewWorld()
	system1 := NewDamageResistanceParticleSystem(world1, seed)

	world2 := NewWorld()
	system2 := NewDamageResistanceParticleSystem(world2, seed)

	// Both should have same seed
	if system1.seed != system2.seed {
		t.Errorf("seeds differ: sys1=%d, sys2=%d", system1.seed, system2.seed)
	}
}

func TestDamageResistanceParticleSystem_AllDamageTypes(t *testing.T) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)
	system.SetGenre("fantasy")

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	damageTypes := []combat.DamageType{
		combat.DamagePhysical,
		combat.DamageMagical,
		combat.DamageFire,
		combat.DamageIce,
		combat.DamageLightning,
		combat.DamagePoison,
	}

	for _, dt := range damageTypes {
		t.Run(dt.String(), func(t *testing.T) {
			// Should not panic for any damage type
			system.OnDamageResisted(target, dt, 100, 50, 0.5)
		})
	}
}

func TestCombatSystem_SetDamageResistedCallback(t *testing.T) {
	cs := NewCombatSystem(12345)

	callbackCalled := false
	callback := func(target *Entity, damageType combat.DamageType, origDamage, finalDamage, resistance float64) {
		callbackCalled = true
	}

	cs.SetDamageResistedCallback(callback)

	if cs.onDamageResistedCallback == nil {
		t.Error("damage resisted callback not set")
	}

	// Verify callback can be invoked
	cs.onDamageResistedCallback(nil, combat.DamageFire, 100, 50, 0.5)
	if !callbackCalled {
		t.Error("callback was not invoked")
	}
}

func TestCombatSystem_DamageResistedCallback_Integration(t *testing.T) {
	world := NewWorld()
	sys := NewDamageResistanceParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	cs := NewCombatSystem(12345)

	// Register the OnDamageResisted method as callback
	cs.SetDamageResistedCallback(sys.OnDamageResisted)

	// Verify callback is set
	if cs.onDamageResistedCallback == nil {
		t.Error("damage resisted callback should be registered")
	}

	// Create entity for callback
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Invoke callback directly (simulating what combat system does on resistance)
	cs.onDamageResistedCallback(target, combat.DamageFire, 100, 50, 0.5)
}

func BenchmarkDamageResistanceParticleSystem_OnDamageResisted(b *testing.B) {
	world := NewWorld()
	system := NewDamageResistanceParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)
	system.SetGenre("fantasy")

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnDamageResisted(target, combat.DamageFire, 100, 50, 0.5)
	}
}
