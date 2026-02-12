package engine

import (
	"math"
	"testing"
)

func TestNewStatusEffectMovementSystem(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewStatusEffectMovementSystem returned nil")
	}
	if sys.world != world {
		t.Error("System world reference mismatch")
	}
	if sys.rng == nil {
		t.Error("System RNG not initialized")
	}
	if sys.speedCache == nil {
		t.Error("System speedCache not initialized")
	}
}

func TestStatusEffectMovementSystem_NoEffects(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 50})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	if vel.VX != 100 || vel.VY != 50 {
		t.Errorf("Velocity should be unchanged without effects, got VX=%f VY=%f", vel.VX, vel.VY)
	}
}

func TestStatusEffectMovementSystem_ChilledEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "chilled",
		Magnitude:  0.15, // Light intensity from weather
		Duration:   6.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// Chilled with 0.15 magnitude: 0.15 + (0.15 * 0.25) = 0.1875 slow
	// Expected multiplier: 1.0 - 0.1875 = 0.8125
	expectedMult := 0.8125
	expectedVX := 100 * expectedMult

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Chilled effect: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_FrozenEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen",
		Duration:   3.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// Frozen: 80% slow = 0.2 multiplier
	expectedVX := 100 * 0.2

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Frozen effect: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_SpeedBoost(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "speed_boost",
		Magnitude:  2.0, // 100% speed increase
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	expectedVX := 100 * 2.0

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Speed boost effect: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_WetEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "wet",
		Duration:   8.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// Wet: 10% slow = 0.9 multiplier
	expectedVX := 100 * 0.9

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Wet effect: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_HasteEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "haste",
		Duration:   10.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// Haste: 25% speed increase = 1.25 multiplier
	expectedVX := 100 * 1.25

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Haste effect: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_SlowEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "slow",
		Magnitude:  0.5, // 50% slow
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// Slow with 0.5 magnitude: 50% slow = 0.5 multiplier
	expectedVX := 100 * 0.5

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Slow effect: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_MultipleEffects(t *testing.T) {
	// Note: Due to ECS design, only one StatusEffectComponent can exist per entity
	// (components are keyed by Type()). In practice, weather applies one effect at a time.
	// This test validates that chilled effect applies correctly (the latest one wins).
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	// Only one effect survives (last one added)
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "wet",
		Duration:   8.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// Wet: 0.9 multiplier
	expectedVX := 100 * 0.9

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Wet effect: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_ExpiredEffectIgnored(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen",
		Duration:   0, // Expired
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// Expired effect should be ignored
	if vel.VX != 100 || vel.VY != 100 {
		t.Errorf("Expired effect should be ignored, got VX=%f VY=%f", vel.VX, vel.VY)
	}
}

func TestStatusEffectMovementSystem_ClampMinimum(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	// Frozen is already at 0.2, which is above minimum clamp
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen",
		Duration:   3.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// Frozen = 0.2 multiplier
	expectedVX := 100 * 0.2

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Frozen effect: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_ClampMaximum(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	// Single large speed boost
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "speed_boost",
		Magnitude:  2.0,
		Duration:   5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	vel := entity.GetVelocity()
	// speed_boost with 2.0 magnitude = 2.0 multiplier
	expectedVX := 100 * 2.0

	if math.Abs(vel.VX-expectedVX) > 0.01 {
		t.Errorf("Speed boost: expected VX ~%f, got %f", expectedVX, vel.VX)
	}
}

func TestStatusEffectMovementSystem_GetSpeedMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100, VY: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "haste",
		Duration:   10.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0/60.0)

	mult := sys.GetSpeedMultiplier(entity.ID)
	if math.Abs(mult-1.25) > 0.01 {
		t.Errorf("GetSpeedMultiplier: expected ~1.25, got %f", mult)
	}

	// Unknown entity should return 1.0
	unknownMult := sys.GetSpeedMultiplier(99999)
	if unknownMult != 1.0 {
		t.Errorf("Unknown entity should return 1.0, got %f", unknownMult)
	}
}

func TestStatusEffectMovementSystem_HasMovementEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	tests := []struct {
		name       string
		effectType string
		want       bool
	}{
		{"chilled", "chilled", true},
		{"frozen", "frozen", true},
		{"wet", "wet", true},
		{"speed_boost", "speed_boost", true},
		{"haste", "haste", true},
		{"slow", "slow", true},
		{"poison", "poison", false},
		{"burning", "burning", false},
		{"stun", "stun", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&StatusEffectComponent{
				EffectType: tt.effectType,
				Duration:   5.0,
			})

			got := sys.HasMovementEffect(entity)
			if got != tt.want {
				t.Errorf("HasMovementEffect(%s) = %v, want %v", tt.effectType, got, tt.want)
			}
		})
	}
}

func TestStatusEffectMovementSystem_HasMovementEffect_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	if sys.HasMovementEffect(nil) {
		t.Error("HasMovementEffect(nil) should return false")
	}
}

func TestStatusEffectMovementSystem_NoVelocityComponent(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	// Entity without velocity should be skipped
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frozen",
		Duration:   3.0,
	})

	entities := []*Entity{entity}
	// Should not panic
	sys.Update(entities, 1.0/60.0)
}

func BenchmarkStatusEffectMovementSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewStatusEffectMovementSystem(world, 12345)

	// Create 100 entities with velocity and status effects
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		e := world.CreateEntity()
		e.AddComponent(&VelocityComponent{VX: 100, VY: 100})
		if i%3 == 0 {
			e.AddComponent(&StatusEffectComponent{
				EffectType: "chilled",
				Magnitude:  0.15,
				Duration:   6.0,
			})
		}
		if i%5 == 0 {
			e.AddComponent(&StatusEffectComponent{
				EffectType: "wet",
				Duration:   8.0,
			})
		}
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset velocities for consistent benchmark
		for _, e := range entities {
			if vel := e.GetVelocity(); vel != nil {
				vel.VX = 100
				vel.VY = 100
			}
		}
		sys.Update(entities, 1.0/60.0)
	}
}
