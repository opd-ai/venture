// Package particles provides particle behavior tests.
package particles

import (
	"math"
	"testing"
)

func TestParticleBehavior_Has(t *testing.T) {
	tests := []struct {
		name     string
		behavior ParticleBehavior
		flag     ParticleBehavior
		want     bool
	}{
		{
			name:     "gravity has gravity",
			behavior: BehaviorGravity,
			flag:     BehaviorGravity,
			want:     true,
		},
		{
			name:     "gravity doesn't have air resistance",
			behavior: BehaviorGravity,
			flag:     BehaviorAirResistance,
			want:     false,
		},
		{
			name:     "combined flags has both",
			behavior: BehaviorGravity | BehaviorAirResistance,
			flag:     BehaviorGravity,
			want:     true,
		},
		{
			name:     "combined flags has both (second)",
			behavior: BehaviorGravity | BehaviorAirResistance,
			flag:     BehaviorAirResistance,
			want:     true,
		},
		{
			name:     "all behaviors has bounce",
			behavior: BehaviorGravity | BehaviorAirResistance | BehaviorBounce | BehaviorTrail,
			flag:     BehaviorBounce,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.behavior.Has(tt.flag); got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultPhysicsConfig(t *testing.T) {
	config := DefaultPhysicsConfig()

	// Verify sensible defaults
	if config.Gravity != 200.0 {
		t.Errorf("Gravity = %f, want 200.0", config.Gravity)
	}
	if config.AirResistance != 0.05 {
		t.Errorf("AirResistance = %f, want 0.05", config.AirResistance)
	}
	if config.BounceDamping != 0.5 {
		t.Errorf("BounceDamping = %f, want 0.5", config.BounceDamping)
	}
	if config.AttractorStrength != 100.0 {
		t.Errorf("AttractorStrength = %f, want 100.0", config.AttractorStrength)
	}
}

func TestApplyPhysics_Gravity(t *testing.T) {
	p := &Particle{
		X:  0,
		Y:  0,
		VX: 0,
		VY: 0,
	}

	config := PhysicsConfig{
		Gravity: 100.0,
	}

	// Apply gravity for 1 second
	ApplyPhysics(p, BehaviorGravity, config, 1.0)

	// Velocity should increase by 100 pixels/second
	if p.VY != 100.0 {
		t.Errorf("VY after gravity = %f, want 100.0", p.VY)
	}

	// X velocity unchanged
	if p.VX != 0.0 {
		t.Errorf("VX after gravity = %f, want 0.0", p.VX)
	}
}

func TestApplyPhysics_AirResistance(t *testing.T) {
	p := &Particle{
		X:  0,
		Y:  0,
		VX: 100.0,
		VY: 100.0,
	}

	config := PhysicsConfig{
		AirResistance: 0.5, // 50% damping per second
	}

	// Apply air resistance for 1 second
	ApplyPhysics(p, BehaviorAirResistance, config, 1.0)

	// Velocity should be reduced by 50%
	if p.VX != 50.0 {
		t.Errorf("VX after air resistance = %f, want 50.0", p.VX)
	}
	if p.VY != 50.0 {
		t.Errorf("VY after air resistance = %f, want 50.0", p.VY)
	}
}

func TestApplyPhysics_Bounce(t *testing.T) {
	tests := []struct {
		name      string
		initialY  float64
		initialVY float64
		groundY   float64
		damping   float64
		wantY     float64
		wantVY    float64
	}{
		{
			name:      "bounce at ground",
			initialY:  10.0,
			initialVY: 50.0,
			groundY:   0.0,
			damping:   0.5,
			wantY:     0.0,
			wantVY:    -25.0, // Reversed and dampened
		},
		{
			name:      "no bounce above ground",
			initialY:  -10.0,
			initialVY: 50.0,
			groundY:   0.0,
			damping:   0.5,
			wantY:     -10.0,
			wantVY:    50.0, // Unchanged
		},
		{
			name:      "bounce with high damping",
			initialY:  5.0,
			initialVY: 100.0,
			groundY:   0.0,
			damping:   0.1,
			wantY:     0.0,
			wantVY:    -10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Particle{
				X:  0,
				Y:  tt.initialY,
				VX: 10.0, // Should be dampened on bounce
				VY: tt.initialVY,
			}

			config := PhysicsConfig{
				GroundY:       tt.groundY,
				BounceDamping: tt.damping,
			}

			ApplyPhysics(p, BehaviorBounce, config, 0.01)

			if math.Abs(p.Y-tt.wantY) > 0.01 {
				t.Errorf("Y = %f, want %f", p.Y, tt.wantY)
			}
			if math.Abs(p.VY-tt.wantVY) > 0.01 {
				t.Errorf("VY = %f, want %f", p.VY, tt.wantVY)
			}
		})
	}
}

func TestApplyPhysics_Rising(t *testing.T) {
	p := &Particle{
		X:  0,
		Y:  0,
		VX: 0,
		VY: 0,
	}

	config := PhysicsConfig{
		Gravity: 100.0,
	}

	// Apply rising for 1 second
	ApplyPhysics(p, BehaviorRising, config, 1.0)

	// Velocity should decrease by gravity * 0.5
	if p.VY != -50.0 {
		t.Errorf("VY after rising = %f, want -50.0", p.VY)
	}
}

func TestApplyPhysics_Attract(t *testing.T) {
	p := &Particle{
		X:  100.0,
		Y:  0,
		VX: 0,
		VY: 0,
	}

	config := PhysicsConfig{
		AttractorX:        0,
		AttractorY:        0,
		AttractorStrength: 1000.0,
	}

	// Apply attraction for 1 second
	ApplyPhysics(p, BehaviorAttract, config, 1.0)

	// Velocity should be toward attractor (negative X direction)
	if p.VX >= 0 {
		t.Errorf("VX after attraction = %f, want < 0", p.VX)
	}

	// Y velocity should remain 0 (attractor on same Y)
	if math.Abs(p.VY) > 0.01 {
		t.Errorf("VY after attraction = %f, want ~0", p.VY)
	}
}

func TestApplyPhysics_CombinedBehaviors(t *testing.T) {
	p := &Particle{
		X:  0,
		Y:  0,
		VX: 100.0,
		VY: 0,
	}

	config := PhysicsConfig{
		Gravity:       100.0,
		AirResistance: 0.1,
	}

	// Apply gravity + air resistance
	ApplyPhysics(p, BehaviorGravity|BehaviorAirResistance, config, 1.0)

	// Gravity should increase VY by 100, then air resistance dampens it
	// Damping = 1.0 - (0.1 * 1.0) = 0.9
	// VY = 100.0 * 0.9 = 90.0
	if math.Abs(p.VY-90.0) > 0.01 {
		t.Errorf("VY after combined = %f, want 90.0", p.VY)
	}

	// Air resistance should reduce VX
	if p.VX >= 100.0 {
		t.Errorf("VX after combined = %f, want < 100.0", p.VX)
	}
}

func TestDefaultTrailConfig(t *testing.T) {
	config := DefaultTrailConfig()

	if config.MaxTrailLength != 10 {
		t.Errorf("MaxTrailLength = %d, want 10", config.MaxTrailLength)
	}
	if config.TrailFadeRate != 0.5 {
		t.Errorf("TrailFadeRate = %f, want 0.5", config.TrailFadeRate)
	}
	if config.TrailSpacing != 5.0 {
		t.Errorf("TrailSpacing = %f, want 5.0", config.TrailSpacing)
	}
}

func TestUpdateTrail_AddPoints(t *testing.T) {
	pwt := &ParticleWithTrail{
		Particle: Particle{
			X: 10.0,
			Y: 0,
		},
		Trail:         []TrailPoint{},
		LastTrailX:    0,
		LastTrailY:    0,
		TrailDistance: 0,
	}

	config := TrailConfig{
		MaxTrailLength: 5,
		TrailFadeRate:  0.0, // No fading for this test
		TrailSpacing:   5.0,
	}

	// First update - should add trail point (distance = 10)
	UpdateTrail(pwt, config, 0.1)

	if len(pwt.Trail) != 1 {
		t.Errorf("Trail length = %d, want 1", len(pwt.Trail))
	}

	// Move further
	pwt.X = 20.0
	UpdateTrail(pwt, config, 0.1)

	if len(pwt.Trail) != 2 {
		t.Errorf("Trail length after 2nd move = %d, want 2", len(pwt.Trail))
	}

	// Verify trail point positions
	if pwt.Trail[0].X != 10.0 {
		t.Errorf("Latest trail X = %f, want 10.0", pwt.Trail[0].X)
	}
	if pwt.Trail[1].X != 0.0 {
		t.Errorf("Oldest trail X = %f, want 0.0", pwt.Trail[1].X)
	}
}

func TestUpdateTrail_MaxLength(t *testing.T) {
	pwt := &ParticleWithTrail{
		Particle: Particle{
			X: 0,
			Y: 0,
		},
		LastTrailX:    0,
		LastTrailY:    0,
		TrailDistance: 0,
	}

	config := TrailConfig{
		MaxTrailLength: 3,
		TrailFadeRate:  0.0,
		TrailSpacing:   1.0, // Very frequent
	}

	// Add many trail points
	for i := 0; i < 10; i++ {
		pwt.X = float64(i * 2)
		UpdateTrail(pwt, config, 0.1)
	}

	// Should be capped at max length
	if len(pwt.Trail) > 3 {
		t.Errorf("Trail length = %d, want <= 3", len(pwt.Trail))
	}
}

func TestUpdateTrail_Fading(t *testing.T) {
	pwt := &ParticleWithTrail{
		Particle: Particle{
			X: 10.0,
			Y: 0,
		},
		Trail: []TrailPoint{
			{X: 0, Y: 0, Alpha: 1.0},
		},
		LastTrailX: 0,
		LastTrailY: 0,
	}

	config := TrailConfig{
		MaxTrailLength: 10,
		TrailFadeRate:  0.5, // 50% fade per second
		TrailSpacing:   1.0,
	}

	// Update for 1 second
	UpdateTrail(pwt, config, 1.0)

	// Trail point should fade
	if len(pwt.Trail) > 0 && pwt.Trail[0].Alpha >= 1.0 {
		t.Errorf("Trail alpha = %f, want < 1.0", pwt.Trail[0].Alpha)
	}
}

func TestUpdateTrail_RemoveDead(t *testing.T) {
	pwt := &ParticleWithTrail{
		Particle: Particle{
			X: 0,
			Y: 0,
		},
		Trail: []TrailPoint{
			{X: 0, Y: 0, Alpha: 0.1},
		},
		LastTrailX: 0,
		LastTrailY: 0,
	}

	config := TrailConfig{
		MaxTrailLength: 10,
		TrailFadeRate:  1.0, // 100% fade per second
		TrailSpacing:   1.0,
	}

	// Update for 1 second - trail should die
	UpdateTrail(pwt, config, 1.0)

	// Trail should be empty (alpha went below 0)
	if len(pwt.Trail) != 0 {
		t.Errorf("Trail length = %d, want 0 (dead trails removed)", len(pwt.Trail))
	}
}

func BenchmarkApplyPhysics_Gravity(b *testing.B) {
	p := &Particle{X: 0, Y: 0, VX: 10, VY: 10}
	config := DefaultPhysicsConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyPhysics(p, BehaviorGravity, config, 0.016)
	}
}

func BenchmarkApplyPhysics_AllBehaviors(b *testing.B) {
	p := &Particle{X: 100, Y: 100, VX: 50, VY: 50}
	config := DefaultPhysicsConfig()
	behavior := BehaviorGravity | BehaviorAirResistance | BehaviorBounce | BehaviorAttract

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyPhysics(p, behavior, config, 0.016)
	}
}

func BenchmarkUpdateTrail(b *testing.B) {
	pwt := &ParticleWithTrail{
		Particle:   Particle{X: 0, Y: 0},
		LastTrailX: 0,
		LastTrailY: 0,
		Trail:      make([]TrailPoint, 0, 10),
	}
	config := DefaultTrailConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pwt.X = float64(i)
		UpdateTrail(pwt, config, 0.016)
	}
}
