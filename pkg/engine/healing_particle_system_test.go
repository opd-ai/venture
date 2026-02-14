package engine

import (
	"testing"
)

func TestNewHealingParticleSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)

	system := NewHealingParticleSystem(world, seed)
	if system == nil {
		t.Fatal("NewHealingParticleSystem returned nil")
	}

	if system.world != world {
		t.Error("system.world not set correctly")
	}

	if system.seed != seed {
		t.Errorf("system.seed = %d, want %d", system.seed, seed)
	}

	if system.rng == nil {
		t.Error("system.rng is nil")
	}

	if system.particleCount != 8 {
		t.Errorf("system.particleCount = %d, want 8", system.particleCount)
	}

	if system.spreadFactor != 50.0 {
		t.Errorf("system.spreadFactor = %f, want 50.0", system.spreadFactor)
	}

	if system.minHealAmount != 1.0 {
		t.Errorf("system.minHealAmount = %f, want 1.0", system.minHealAmount)
	}
}

func TestHealingParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particleSystem not set correctly")
	}
}

func TestHealingParticleSystem_SetGenre(t *testing.T) {
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
			system := NewHealingParticleSystem(world, 12345)

			system.SetGenre(tt.genreID)

			if system.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", system.genreID, tt.genreID)
			}
		})
	}
}

func TestHealingParticleSystem_SetMinHealAmount(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)

	// Test normal value
	system.SetMinHealAmount(5.0)
	if system.minHealAmount != 5.0 {
		t.Errorf("minHealAmount = %f, want 5.0", system.minHealAmount)
	}

	// Test negative clamps to 0
	system.SetMinHealAmount(-10.0)
	if system.minHealAmount != 0.0 {
		t.Errorf("minHealAmount = %f, want 0.0 (clamped from negative)", system.minHealAmount)
	}
}

func TestHealingParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)

	// Create test entities
	entities := []*Entity{
		world.CreateEntity(),
		world.CreateEntity(),
	}

	// Update should not panic (it's a no-op for callback-driven systems)
	system.Update(entities, 0.016)
}

func TestHealingParticleSystem_OnHeal_NilChecks(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)

	// Should not panic with nil particle system
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 200})

	system.OnHeal(nil, target, 10.0)

	// Should not panic with nil target
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.OnHeal(nil, nil, 10.0)
}

func TestHealingParticleSystem_OnHeal_MinimumThreshold(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetMinHealAmount(5.0)

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 200})

	// Heal below threshold should not crash
	system.OnHeal(nil, target, 3.0)

	// Heal above threshold should work
	system.OnHeal(nil, target, 10.0)
}

func TestHealingParticleSystem_OnHeal_WithoutPosition(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Target without position - should not panic
	target := world.CreateEntity()
	system.OnHeal(nil, target, 10.0)
}

func TestHealingParticleSystem_OnHeal_SelfHeal(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create player healing themselves
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 200})
	player.AddComponent(NewStubInput())

	// Self-heal: healer == target
	system.OnHeal(player, player, 25.0)
}

func TestHealingParticleSystem_OnHeal_HealerAndTarget(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	// Create healer and target
	healer := world.CreateEntity()
	healer.AddComponent(&PositionComponent{X: 50, Y: 100})

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 200, Y: 300})

	// Healer heals target
	system.OnHeal(healer, target, 50.0)
}

func TestHealingParticleSystem_GenreParticleType(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)

	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "magical_sparkles"},
		{"scifi", "nanobots"},
		{"horror", "blood_mist"},
		{"cyberpunk", "stim_injection"},
		{"postapoc", "medkit_spray"},
		{"unknown", "magical_sparkles"},
		{"", "magical_sparkles"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			result := system.getGenreHealingParticleType()
			if result != tt.expected {
				t.Errorf("getGenreHealingParticleType() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestHealingParticleSystem_ParticleScaling(t *testing.T) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 200})

	// Test various heal amounts for particle scaling
	testAmounts := []float64{1.0, 10.0, 50.0, 100.0, 500.0}
	for _, amount := range testAmounts {
		system.OnHeal(nil, target, amount)
	}
}

func TestCombatSystem_SetHealCallback(t *testing.T) {
	system := NewCombatSystem(12345)

	callbackCalled := false
	var capturedHealer, capturedTarget *Entity
	var capturedAmount float64

	system.SetHealCallback(func(healer, target *Entity, amount float64) {
		callbackCalled = true
		capturedHealer = healer
		capturedTarget = target
		capturedAmount = amount
	})

	// Create a target with health
	world := NewWorld()
	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 50, Max: 100})

	// Heal should trigger callback
	system.Heal(target, 30.0)

	if !callbackCalled {
		t.Error("heal callback was not called")
	}

	if capturedHealer != nil {
		t.Error("healer should be nil for Heal() without healer")
	}

	if capturedTarget != target {
		t.Error("target not captured correctly")
	}

	if capturedAmount != 30.0 {
		t.Errorf("amount = %f, want 30.0", capturedAmount)
	}
}

func TestCombatSystem_HealWithHealer(t *testing.T) {
	system := NewCombatSystem(12345)

	var capturedHealer, capturedTarget *Entity
	var capturedAmount float64

	system.SetHealCallback(func(healer, target *Entity, amount float64) {
		capturedHealer = healer
		capturedTarget = target
		capturedAmount = amount
	})

	world := NewWorld()
	healer := world.CreateEntity()
	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 50, Max: 100})

	system.HealWithHealer(healer, target, 25.0)

	if capturedHealer != healer {
		t.Error("healer not captured correctly")
	}

	if capturedTarget != target {
		t.Error("target not captured correctly")
	}

	if capturedAmount != 25.0 {
		t.Errorf("amount = %f, want 25.0", capturedAmount)
	}
}

func TestCombatSystem_HealWithHealer_NoOverheal(t *testing.T) {
	system := NewCombatSystem(12345)

	var capturedAmount float64

	system.SetHealCallback(func(healer, target *Entity, amount float64) {
		capturedAmount = amount
	})

	world := NewWorld()
	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 90, Max: 100})

	// Try to heal 50, but only 10 should be effective
	system.HealWithHealer(nil, target, 50.0)

	// Callback should receive actual amount healed
	if capturedAmount != 10.0 {
		t.Errorf("actual heal amount = %f, want 10.0 (capped by max health)", capturedAmount)
	}
}

func TestCombatSystem_HealWithHealer_ZeroHeal(t *testing.T) {
	system := NewCombatSystem(12345)

	callbackCalled := false
	system.SetHealCallback(func(healer, target *Entity, amount float64) {
		callbackCalled = true
	})

	world := NewWorld()
	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 100, Max: 100}) // Already full

	// No actual healing occurs
	system.HealWithHealer(nil, target, 50.0)

	if callbackCalled {
		t.Error("callback should not be called when no actual healing occurs")
	}
}

func TestCombatSystem_HealWithHealer_NilTarget(t *testing.T) {
	system := NewCombatSystem(12345)

	callbackCalled := false
	system.SetHealCallback(func(healer, target *Entity, amount float64) {
		callbackCalled = true
	})

	// Should not panic with nil target
	system.HealWithHealer(nil, nil, 50.0)

	if callbackCalled {
		t.Error("callback should not be called with nil target")
	}
}

func TestCombatSystem_HealWithHealer_NegativeAmount(t *testing.T) {
	system := NewCombatSystem(12345)

	callbackCalled := false
	system.SetHealCallback(func(healer, target *Entity, amount float64) {
		callbackCalled = true
	})

	world := NewWorld()
	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 50, Max: 100})

	// Negative amount should be rejected
	system.HealWithHealer(nil, target, -10.0)

	if callbackCalled {
		t.Error("callback should not be called with negative amount")
	}
}

func TestCombatSystem_HealWithHealer_NoHealthComponent(t *testing.T) {
	system := NewCombatSystem(12345)

	callbackCalled := false
	system.SetHealCallback(func(healer, target *Entity, amount float64) {
		callbackCalled = true
	})

	world := NewWorld()
	target := world.CreateEntity() // No health component

	// Should not panic or call callback
	system.HealWithHealer(nil, target, 50.0)

	if callbackCalled {
		t.Error("callback should not be called when target has no health component")
	}
}

func BenchmarkHealingParticleSystem_OnHeal(b *testing.B) {
	world := NewWorld()
	system := NewHealingParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 200})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnHeal(nil, target, 25.0)
	}
}

func BenchmarkCombatSystem_HealWithCallback(b *testing.B) {
	system := NewCombatSystem(12345)
	system.SetHealCallback(func(healer, target *Entity, amount float64) {})

	world := NewWorld()
	target := world.CreateEntity()
	// Add health with room to heal each iteration
	target.AddComponent(&HealthComponent{Current: 50, Max: 1000000})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.HealWithHealer(nil, target, 10.0)
	}
}
