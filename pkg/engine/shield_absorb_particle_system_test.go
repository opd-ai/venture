package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewShieldAbsorbParticleSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	sys := NewShieldAbsorbParticleSystem(world, seed)

	if sys == nil {
		t.Fatal("NewShieldAbsorbParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("World reference not set")
	}
	if sys.seed != seed {
		t.Errorf("Seed = %d, want %d", sys.seed, seed)
	}
	if sys.particleCount != 8 {
		t.Errorf("Default particleCount = %d, want 8", sys.particleCount)
	}
	if sys.spreadFactor != 80.0 {
		t.Errorf("Default spreadFactor = %f, want 80.0", sys.spreadFactor)
	}
	if sys.minAbsorbAmount != 1.0 {
		t.Errorf("Default minAbsorbAmount = %f, want 1.0", sys.minAbsorbAmount)
	}
}

func TestShieldAbsorbParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("SetParticleSystem did not set particle system")
	}
}

func TestShieldAbsorbParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)

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
			sys.SetGenre(tt.genreID)
			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestShieldAbsorbParticleSystem_SetMinAbsorbAmount(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"positive", 5.0, 5.0},
		{"zero", 0.0, 0.0},
		{"negative clamped", -10.0, 0.0},
		{"large", 100.0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetMinAbsorbAmount(tt.input)
			if sys.minAbsorbAmount != tt.expected {
				t.Errorf("minAbsorbAmount = %f, want %f", sys.minAbsorbAmount, tt.expected)
			}
		})
	}
}

func TestShieldAbsorbParticleSystem_OnShieldAbsorb_NilChecks(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)

	// Should not panic with nil particle system
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	sys.OnShieldAbsorb(entity, 10.0, 50.0)

	// Should not panic with nil target
	sys.SetParticleSystem(NewParticleSystem())
	sys.OnShieldAbsorb(nil, 10.0, 50.0)

	// Should not panic with nil world
	sys2 := NewShieldAbsorbParticleSystem(nil, 12345)
	sys2.SetParticleSystem(NewParticleSystem())
	sys2.OnShieldAbsorb(entity, 10.0, 50.0)
}

func TestShieldAbsorbParticleSystem_OnShieldAbsorb_MinAbsorbThreshold(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetMinAbsorbAmount(5.0)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not spawn particles if absorbed < minAbsorbAmount
	sys.OnShieldAbsorb(entity, 2.0, 50.0) // absorbed=2 < min=5, should not trigger
}

func TestShieldAbsorbParticleSystem_OnShieldAbsorb_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Entity without position component
	entity := world.CreateEntity()

	// Should not panic without position
	sys.OnShieldAbsorb(entity, 10.0, 50.0)
}

func TestShieldAbsorbParticleSystem_OnShieldAbsorb_WithValidEntity(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 150, Y: 200})

	// Should complete without error
	sys.OnShieldAbsorb(entity, 25.0, 75.0)
}

func TestShieldAbsorbParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)

	// Update is a no-op for this callback-driven system
	// Just verify it doesn't panic
	entity := world.CreateEntity()
	sys.Update([]*Entity{entity}, 0.016)
}

func TestShieldAbsorbParticleSystem_getParticleTypeForGenre(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)

	tests := []struct {
		genre    string
		wantType particles.ParticleType
	}{
		{"fantasy", particles.ParticleSparkle},
		{"scifi", particles.ParticleSpark},
		{"horror", particles.ParticleSmoke},
		{"cyberpunk", particles.ParticleSpark},
		{"postapoc", particles.ParticleDust},
		{"unknown", particles.ParticleSparkle},
		{"", particles.ParticleSparkle},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.genreID = tt.genre
			particleType := sys.getParticleTypeForGenre()
			if particleType != tt.wantType {
				t.Errorf("getParticleTypeForGenre() = %v, want %v", particleType, tt.wantType)
			}
		})
	}
}

func TestShieldAbsorbParticleSystem_SpawnShieldEffect(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)

	// Should not panic without particle system
	sys.SpawnShieldEffect(100, 100, 10)

	// Should not panic without world
	sys2 := NewShieldAbsorbParticleSystem(nil, 12345)
	sys2.SetParticleSystem(NewParticleSystem())
	sys2.SpawnShieldEffect(100, 100, 10)

	// Should complete without error with both set
	sys.SetParticleSystem(NewParticleSystem())
	sys.SetGenre("scifi")
	sys.SpawnShieldEffect(200, 150, 30)
}

func TestShieldAbsorbParticleSystem_ParticleCountScaling(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Test various absorbed amounts to verify scaling doesn't panic
	testAmounts := []float64{1.0, 10.0, 25.0, 55.0, 100.0}
	for _, amount := range testAmounts {
		sys.OnShieldAbsorb(entity, amount, 100.0)
	}
}

func TestShieldAbsorbParticleSystem_ShieldBreakingEffect(t *testing.T) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Test low remaining shield (should trigger breaking effect)
	sys.OnShieldAbsorb(entity, 50.0, 5.0) // remaining < 10 triggers breaking
}

func BenchmarkShieldAbsorbParticleSystem_OnShieldAbsorb(b *testing.B) {
	world := NewWorld()
	sys := NewShieldAbsorbParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.OnShieldAbsorb(entity, 25.0, 75.0)
	}
}
