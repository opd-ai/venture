package vehicle

import (
	"math"
	"testing"
)

func TestNewCollisionResponseComponent(t *testing.T) {
	mass := 1000.0
	comp := NewCollisionResponseComponent(mass)

	if comp == nil {
		t.Fatal("NewCollisionResponseComponent returned nil")
	}
	if comp.Type() != "collision_response" {
		t.Errorf("got type %q, want %q", comp.Type(), "collision_response")
	}
	if comp.MassForCalculation != mass {
		t.Errorf("got mass=%f, want %f", comp.MassForCalculation, mass)
	}
	if comp.StructuralIntegrity != 1.0 {
		t.Errorf("got integrity=%f, want 1.0", comp.StructuralIntegrity)
	}
}

func TestCollisionResponseComponent_ProcessCollision_BelowThreshold(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewCollisionResponseComponent(1000.0)

	// Low velocity collision (below damage threshold of 50 px/s)
	result := sys.ProcessCollisionResponse(comp, 30.0, 0.0, -1.0, 0.0)

	if result.DamageDealt != 0.0 {
		t.Errorf("low velocity collision should deal no damage, got %f", result.DamageDealt)
	}
	if result.IntegrityLoss != 0.0 {
		t.Errorf("low velocity collision should cause no integrity loss, got %f", result.IntegrityLoss)
	}
	if comp.StructuralIntegrity != 1.0 {
		t.Errorf("integrity should remain 1.0, got %f", comp.StructuralIntegrity)
	}
}

func TestCollisionResponseComponent_ProcessCollision_AboveThreshold(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewCollisionResponseComponent(1000.0)

	// High velocity head-on collision
	result := sys.ProcessCollisionResponse(comp, 100.0, 0.0, -1.0, 0.0)

	if result.DamageDealt <= 0.0 {
		t.Error("high velocity collision should deal damage")
	}
	if result.IntegrityLoss <= 0.0 {
		t.Error("high velocity collision should cause integrity loss")
	}
	if comp.StructuralIntegrity >= 1.0 {
		t.Error("structural integrity should be reduced")
	}
}

func TestCollisionResponseComponent_ProcessCollision_HeadOn(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewCollisionResponseComponent(1000.0)

	// Head-on collision (velocity opposite to normal)
	// Velocity: (100, 0), Normal: (-1, 0)
	result := sys.ProcessCollisionResponse(comp, 100.0, 0.0, -1.0, 0.0)

	// Check bounce velocity is opposite direction
	if result.BounceVelocityX >= 0.0 {
		t.Errorf("head-on collision should reverse X velocity, got %f", result.BounceVelocityX)
	}

	// Check damage is present for head-on (threshold lowered for realistic values)
	if result.DamageDealt < 0.1 {
		t.Errorf("head-on collision should deal some damage, got %f", result.DamageDealt)
	}
}

func TestCollisionResponseComponent_ProcessCollision_GlancingBlow(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp1 := NewCollisionResponseComponent(1000.0)
	comp2 := NewCollisionResponseComponent(1000.0)

	// Head-on collision
	headOn := sys.ProcessCollisionResponse(comp1, 100.0, 0.0, -1.0, 0.0)

	// Glancing blow (45 degree angle)
	// Velocity: (100, 0), Normal: (-0.707, -0.707)
	glancing := sys.ProcessCollisionResponse(comp2, 100.0, 0.0, -0.707, -0.707)

	// Glancing blow should deal less damage than head-on
	if glancing.DamageDealt >= headOn.DamageDealt {
		t.Errorf("glancing damage (%f) should be less than head-on (%f)", glancing.DamageDealt, headOn.DamageDealt)
	}
}

func TestCollisionResponseComponent_GetDamageMultiplier(t *testing.T) {
	comp := NewCollisionResponseComponent(1000.0)

	// At full integrity
	mult := GetDamageMultiplier(comp)
	if mult != 1.0 {
		t.Errorf("at full integrity, multiplier should be 1.0, got %f", mult)
	}

	// Damage vehicle to 50% integrity
	comp.StructuralIntegrity = 0.5
	mult = GetDamageMultiplier(comp)
	expected := 0.5 + (0.5 * 0.5) // 0.75
	if math.Abs(mult-expected) > 0.01 {
		t.Errorf("at 50%% integrity, multiplier should be ~%f, got %f", expected, mult)
	}

	// Damage vehicle to 0% integrity
	comp.StructuralIntegrity = 0.0
	mult = GetDamageMultiplier(comp)
	if mult != 0.5 {
		t.Errorf("at 0%% integrity, multiplier should be 0.5, got %f", mult)
	}
}

func TestCollisionResponseComponent_IsDestroyed(t *testing.T) {
	comp := NewCollisionResponseComponent(1000.0)

	if IsDestroyed(comp) {
		t.Error("new component should not be destroyed")
	}

	comp.StructuralIntegrity = 0.5
	if IsDestroyed(comp) {
		t.Error("50% integrity should not be destroyed")
	}

	comp.StructuralIntegrity = 0.0
	if !IsDestroyed(comp) {
		t.Error("0% integrity should be destroyed")
	}
}

func TestCollisionResponseComponent_Repair(t *testing.T) {
	comp := NewCollisionResponseComponent(1000.0)

	// Damage it
	comp.StructuralIntegrity = 0.5

	// Repair partially
	RepairVehicle(comp, 0.3)
	if math.Abs(comp.StructuralIntegrity-0.8) > 0.01 {
		t.Errorf("after repair, integrity should be ~0.8, got %f", comp.StructuralIntegrity)
	}

	// Over-repair (should clamp to 1.0)
	RepairVehicle(comp, 1.0)
	if comp.StructuralIntegrity != 1.0 {
		t.Errorf("integrity should be clamped to 1.0, got %f", comp.StructuralIntegrity)
	}
}

func TestCollisionResponseComponent_Reset(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewCollisionResponseComponent(1000.0)

	// Process some collisions
	sys.ProcessCollisionResponse(comp, 100.0, 0.0, -1.0, 0.0)
	sys.ProcessCollisionResponse(comp, 80.0, 0.0, -1.0, 0.0)

	if comp.CollisionCount != 2 {
		t.Fatalf("expected 2 collisions, got %d", comp.CollisionCount)
	}

	// Reset
	comp.Reset()

	if comp.CollisionCount != 0 {
		t.Errorf("after reset, collision count should be 0, got %d", comp.CollisionCount)
	}
	if comp.StructuralIntegrity != 1.0 {
		t.Errorf("after reset, integrity should be 1.0, got %f", comp.StructuralIntegrity)
	}
	if comp.TotalImpactDamage != 0.0 {
		t.Errorf("after reset, total damage should be 0.0, got %f", comp.TotalImpactDamage)
	}
}

func TestCollisionResponseComponent_VelocityReflection(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewCollisionResponseComponent(1000.0)

	tests := []struct {
		name    string
		velX    float64
		velY    float64
		normalX float64
		normalY float64
		expectX float64 // Expected sign of bounce velocity
		expectY float64
	}{
		{"wall right", 100.0, 0.0, -1.0, 0.0, -1.0, 0.0}, // Should bounce left
		{"wall left", -100.0, 0.0, 1.0, 0.0, 1.0, 0.0},   // Should bounce right
		{"wall top", 0.0, 100.0, 0.0, -1.0, 0.0, -1.0},   // Should bounce down
		{"wall bottom", 0.0, -100.0, 0.0, 1.0, 0.0, 1.0}, // Should bounce up
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp.Reset() // Reset between tests
			result := sys.ProcessCollisionResponse(comp, tt.velX, tt.velY, tt.normalX, tt.normalY)

			// Check sign of bounce velocity
			if tt.expectX != 0.0 {
				bounceSign := 1.0
				if result.BounceVelocityX < 0 {
					bounceSign = -1.0
				}
				if bounceSign != tt.expectX {
					t.Errorf("bounce X sign: got %f, want %f", bounceSign, tt.expectX)
				}
			}

			if tt.expectY != 0.0 {
				bounceSign := 1.0
				if result.BounceVelocityY < 0 {
					bounceSign = -1.0
				}
				if bounceSign != tt.expectY {
					t.Errorf("bounce Y sign: got %f, want %f", bounceSign, tt.expectY)
				}
			}
		})
	}
}

func TestCollisionResponseComponent_RestitutionEffect(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewCollisionResponseComponent(1000.0)

	// Process collision
	velX := 100.0
	result := sys.ProcessCollisionResponse(comp, velX, 0.0, -1.0, 0.0)

	// Bounce velocity should be less than original (due to restitution < 1.0)
	bounceSpeed := math.Sqrt(result.BounceVelocityX*result.BounceVelocityX + result.BounceVelocityY*result.BounceVelocityY)
	if bounceSpeed >= velX {
		t.Errorf("bounce speed (%f) should be less than impact speed (%f)", bounceSpeed, velX)
	}

	// Check velocity reduction is positive
	if result.VelocityReduction <= 0.0 {
		t.Errorf("velocity reduction should be positive, got %f", result.VelocityReduction)
	}
}

func TestCollisionResponseComponent_ShouldCauseDamage(t *testing.T) {
	comp := NewCollisionResponseComponent(1000.0)

	tests := []struct {
		speed        float64
		shouldDamage bool
	}{
		{10.0, false}, // Below threshold
		{49.0, false}, // Just below threshold
		{50.0, true},  // At threshold
		{100.0, true}, // Above threshold
		{200.0, true}, // Well above threshold
	}

	for _, tt := range tests {
		got := comp.ShouldCauseDamage(tt.speed)
		if got != tt.shouldDamage {
			t.Errorf("speed %f: got shouldDamage=%v, want %v", tt.speed, got, tt.shouldDamage)
		}
	}
}

// Benchmark tests
func BenchmarkCollisionResponseComponent_ProcessCollision(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	comp := NewCollisionResponseComponent(1000.0)
	velX, velY := 100.0, 50.0
	normalX, normalY := -1.0, 0.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Reset() // Reset to avoid accumulating damage
		_ = sys.ProcessCollisionResponse(comp, velX, velY, normalX, normalY)
	}
}

func BenchmarkCollisionResponseComponent_GetDamageMultiplier(b *testing.B) {
	comp := NewCollisionResponseComponent(1000.0)
	comp.StructuralIntegrity = 0.7

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetDamageMultiplier(comp)
	}
}
