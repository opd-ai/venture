// Package particles provides tests for new particle types (Phase 14.3).
package particles

import (
	"testing"
)

func TestNewParticleTypes_String(t *testing.T) {
	tests := []struct {
		name string
		pt   ParticleType
		want string
	}{
		{"ember", ParticleEmber, "ember"},
		{"sparkle", ParticleSparkle, "sparkle"},
		{"smoke_plume", ParticleSmokePlume, "smoke_plume"},
		{"debris", ParticleDebris, "debris"},
		{"existing_spark", ParticleSpark, "spark"},
		{"existing_magic", ParticleMagic, "magic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pt.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateEmbers(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleEmber,
		Count:    20,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 2.0,
		SpreadX:  100.0,
		SpreadY:  100.0,
		Gravity:  200.0,
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(system.Particles) != 20 {
		t.Errorf("Particle count = %d, want 20", len(system.Particles))
	}

	// Verify embers have rising behavior
	for i, p := range system.Particles {
		if !p.Behavior.Has(BehaviorRising) {
			t.Errorf("Particle %d missing BehaviorRising", i)
		}
		if !p.Behavior.Has(BehaviorAirResistance) {
			t.Errorf("Particle %d missing BehaviorAirResistance", i)
		}
		if p.Size < 2.0 || p.Size > 5.0 {
			t.Errorf("Particle %d size = %f, want 2.0-5.0", i, p.Size)
		}
	}
}

func TestGenerateSparkles(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleSparkle,
		Count:    15,
		GenreID:  "scifi",
		Seed:     54321,
		Duration: 1.5,
		SpreadX:  80.0,
		SpreadY:  80.0,
		Gravity:  0.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(system.Particles) != 15 {
		t.Errorf("Particle count = %d, want 15", len(system.Particles))
	}

	// Verify sparkles have orbital behavior
	for i, p := range system.Particles {
		if !p.Behavior.Has(BehaviorOrbit) {
			t.Errorf("Particle %d missing BehaviorOrbit", i)
		}
		if p.Physics.OrbitSpeed <= 0 {
			t.Errorf("Particle %d has invalid orbit speed: %f", i, p.Physics.OrbitSpeed)
		}
	}
}

func TestGenerateSmokePlume(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleSmokePlume,
		Count:    30,
		GenreID:  "postapoc",
		Seed:     99999,
		Duration: 3.0,
		SpreadX:  120.0,
		SpreadY:  120.0,
		Gravity:  100.0,
		MinSize:  3.0,
		MaxSize:  8.0,
		Custom:   make(map[string]interface{}),
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(system.Particles) != 30 {
		t.Errorf("Particle count = %d, want 30", len(system.Particles))
	}

	// Verify smoke plumes have rising + air resistance
	for i, p := range system.Particles {
		if !p.Behavior.Has(BehaviorRising) {
			t.Errorf("Particle %d missing BehaviorRising", i)
		}
		if !p.Behavior.Has(BehaviorAirResistance) {
			t.Errorf("Particle %d missing BehaviorAirResistance", i)
		}
		// Plumes should be larger than normal particles
		if p.Size < 3.0 {
			t.Errorf("Particle %d size = %f, want >= 3.0", i, p.Size)
		}
	}
}

func TestGenerateDebris(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleDebris,
		Count:    25,
		GenreID:  "fantasy",
		Seed:     11111,
		Duration: 2.5,
		SpreadX:  150.0,
		SpreadY:  150.0,
		Gravity:  300.0,
		MinSize:  2.0,
		MaxSize:  6.0,
		Custom: map[string]interface{}{
			"groundY": 100.0,
		},
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(system.Particles) != 25 {
		t.Errorf("Particle count = %d, want 25", len(system.Particles))
	}

	// Verify debris has gravity + bounce + air resistance
	for i, p := range system.Particles {
		if !p.Behavior.Has(BehaviorGravity) {
			t.Errorf("Particle %d missing BehaviorGravity", i)
		}
		if !p.Behavior.Has(BehaviorBounce) {
			t.Errorf("Particle %d missing BehaviorBounce", i)
		}
		if !p.Behavior.Has(BehaviorAirResistance) {
			t.Errorf("Particle %d missing BehaviorAirResistance", i)
		}
		if p.Physics.GroundY != 100.0 {
			t.Errorf("Particle %d groundY = %f, want 100.0", i, p.Physics.GroundY)
		}
	}
}

func TestNewParticleTypes_Determinism(t *testing.T) {
	gen := NewGenerator()

	types := []ParticleType{
		ParticleEmber,
		ParticleSparkle,
		ParticleSmokePlume,
		ParticleDebris,
	}

	for _, pt := range types {
		t.Run(pt.String(), func(t *testing.T) {
			config := Config{
				Type:     pt,
				Count:    10,
				GenreID:  "fantasy",
				Seed:     42,
				Duration: 1.0,
				SpreadX:  100.0,
				SpreadY:  100.0,
				Gravity:  200.0,
				MinSize:  2.0,
				MaxSize:  4.0,
				Custom:   make(map[string]interface{}),
			}

			// Generate twice with same seed
			system1, err1 := gen.Generate(config)
			if err1 != nil {
				t.Fatalf("First generate error = %v", err1)
			}

			system2, err2 := gen.Generate(config)
			if err2 != nil {
				t.Fatalf("Second generate error = %v", err2)
			}

			// Verify determinism
			if len(system1.Particles) != len(system2.Particles) {
				t.Errorf("Particle counts differ: %d vs %d",
					len(system1.Particles), len(system2.Particles))
			}

			for i := 0; i < len(system1.Particles) && i < len(system2.Particles); i++ {
				p1 := system1.Particles[i]
				p2 := system2.Particles[i]

				if p1.X != p2.X || p1.Y != p2.Y {
					t.Errorf("Particle %d position differs: (%f,%f) vs (%f,%f)",
						i, p1.X, p1.Y, p2.X, p2.Y)
				}
				if p1.VX != p2.VX || p1.VY != p2.VY {
					t.Errorf("Particle %d velocity differs", i)
				}
				if p1.Size != p2.Size {
					t.Errorf("Particle %d size differs: %f vs %f", i, p1.Size, p2.Size)
				}
			}
		})
	}
}

func TestParticleUpdate_WithBehaviors(t *testing.T) {
	// Create a particle system with ember particles
	gen := NewGenerator()
	config := Config{
		Type:     ParticleEmber,
		Count:    5,
		GenreID:  "fantasy",
		Seed:     777,
		Duration: 1.0,
		SpreadX:  100.0,
		SpreadY:  100.0,
		Gravity:  200.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Record initial positions
	initialY := make([]float64, len(system.Particles))
	for i := range system.Particles {
		initialY[i] = system.Particles[i].Y
	}

	// Update for 1 second
	system.Update(1.0)

	// Verify particles moved upward (rising behavior)
	for i := range system.Particles {
		if system.Particles[i].Y >= initialY[i] {
			t.Errorf("Particle %d did not rise: Y %f -> %f",
				i, initialY[i], system.Particles[i].Y)
		}
	}

	// Verify life decreased
	for i := range system.Particles {
		if system.Particles[i].Life >= 1.0 {
			t.Errorf("Particle %d life did not decrease: %f", i, system.Particles[i].Life)
		}
	}
}

func BenchmarkGenerateEmbers(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleEmber,
		Count:    50,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 2.0,
		SpreadX:  100.0,
		SpreadY:  100.0,
		Gravity:  200.0,
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(config)
	}
}

func BenchmarkGenerateSparkles(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleSparkle,
		Count:    50,
		GenreID:  "scifi",
		Seed:     54321,
		Duration: 1.5,
		SpreadX:  80.0,
		SpreadY:  80.0,
		Gravity:  0.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(config)
	}
}

func BenchmarkParticleUpdate_WithBehaviors(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleDebris,
		Count:    100,
		GenreID:  "fantasy",
		Seed:     99999,
		Duration: 2.0,
		SpreadX:  150.0,
		SpreadY:  150.0,
		Gravity:  300.0,
		MinSize:  2.0,
		MaxSize:  6.0,
		Custom:   make(map[string]interface{}),
	}

	system, _ := gen.Generate(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(0.016) // 60 FPS
	}
}

// TestParticle_ColorIsRGBA verifies that Particle.Color is color.RGBA (not interface).
func TestParticle_ColorIsRGBA(t *testing.T) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleSpark,
		Count:    10,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 1.0,
		SpreadX:  50.0,
		SpreadY:  50.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(system.Particles) == 0 {
		t.Fatal("Expected particles to be generated")
	}

	// Verify Color field is color.RGBA
	p := system.Particles[0]
	_ = p.Color.R // Should compile if Color is color.RGBA
	_ = p.Color.G
	_ = p.Color.B
	_ = p.Color.A

	// Verify alpha channel can be manipulated directly
	originalAlpha := p.Color.A
	p.Color.A = uint8(float64(originalAlpha) * 0.5)
	if p.Color.A >= originalAlpha {
		t.Errorf("Alpha modification failed: %d >= %d", p.Color.A, originalAlpha)
	}
}
