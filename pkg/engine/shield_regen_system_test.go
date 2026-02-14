package engine

import (
	"testing"
)

func TestNewShieldRegenSystem(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewShieldRegenSystem returned nil")
	}

	if sys.world != world {
		t.Error("world not set correctly")
	}

	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}

	if sys.rng == nil {
		t.Error("rng not initialized")
	}

	if sys.baseRegenRate != 5.0 {
		t.Errorf("baseRegenRate = %f, want 5.0", sys.baseRegenRate)
	}

	if sys.regenDelay != 3.0 {
		t.Errorf("regenDelay = %f, want 3.0", sys.regenDelay)
	}
}

func TestShieldRegenSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particleSystem not set correctly")
	}
}

func TestShieldRegenSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("genreID = %s, want %s", sys.genreID, tt.genre)
			}
		})
	}
}

func TestShieldRegenSystem_SetBaseRegenRate(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)

	tests := []struct {
		name     string
		rate     float64
		expected float64
	}{
		{"positive rate", 10.0, 10.0},
		{"zero rate", 0.0, 0.0},
		{"negative rate clamped", -5.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetBaseRegenRate(tt.rate)
			if sys.baseRegenRate != tt.expected {
				t.Errorf("baseRegenRate = %f, want %f", sys.baseRegenRate, tt.expected)
			}
		})
	}
}

func TestShieldRegenSystem_SetRegenDelay(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)

	tests := []struct {
		name     string
		delay    float64
		expected float64
	}{
		{"positive delay", 5.0, 5.0},
		{"zero delay", 0.0, 0.0},
		{"negative delay clamped", -2.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetRegenDelay(tt.delay)
			if sys.regenDelay != tt.expected {
				t.Errorf("regenDelay = %f, want %f", sys.regenDelay, tt.expected)
			}
		})
	}
}

func TestShieldRegenSystem_Update_RegeneratesShield(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	sys.SetBaseRegenRate(10.0) // 10 shield per second
	sys.SetRegenDelay(0.0)     // No delay for testing

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ShieldComponent{
		Amount:      50.0,
		MaxAmount:   100.0,
		Duration:    60.0,
		MaxDuration: 60.0,
	})

	entities := []*Entity{entity}

	// Update for 1 second
	sys.Update(entities, 1.0)

	shield := entity.GetComponent("shield").(*ShieldComponent)
	expectedAmount := 60.0 // 50 + 10*1.0

	if shield.Amount != expectedAmount {
		t.Errorf("shield.Amount = %f, want %f", shield.Amount, expectedAmount)
	}
}

func TestShieldRegenSystem_Update_CapsAtMaxAmount(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	sys.SetBaseRegenRate(100.0) // High regen rate
	sys.SetRegenDelay(0.0)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ShieldComponent{
		Amount:      95.0,
		MaxAmount:   100.0,
		Duration:    60.0,
		MaxDuration: 60.0,
	})

	entities := []*Entity{entity}

	// Update for 1 second - should cap at max
	sys.Update(entities, 1.0)

	shield := entity.GetComponent("shield").(*ShieldComponent)

	if shield.Amount != 100.0 {
		t.Errorf("shield.Amount = %f, want 100.0 (capped at max)", shield.Amount)
	}
}

func TestShieldRegenSystem_Update_NoRegenWhenFull(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	sys.SetBaseRegenRate(10.0)
	sys.SetRegenDelay(0.0)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ShieldComponent{
		Amount:      100.0, // Already full
		MaxAmount:   100.0,
		Duration:    60.0,
		MaxDuration: 60.0,
	})

	entities := []*Entity{entity}

	sys.Update(entities, 1.0)

	shield := entity.GetComponent("shield").(*ShieldComponent)

	if shield.Amount != 100.0 {
		t.Errorf("shield.Amount = %f, want 100.0 (no change when full)", shield.Amount)
	}
}

func TestShieldRegenSystem_Update_NoRegenWhenExpired(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	sys.SetBaseRegenRate(10.0)
	sys.SetRegenDelay(0.0)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ShieldComponent{
		Amount:      50.0,
		MaxAmount:   100.0,
		Duration:    0.0, // Expired
		MaxDuration: 60.0,
	})

	entities := []*Entity{entity}

	sys.Update(entities, 1.0)

	shield := entity.GetComponent("shield").(*ShieldComponent)

	if shield.Amount != 50.0 {
		t.Errorf("shield.Amount = %f, want 50.0 (no regen when expired)", shield.Amount)
	}
}

func TestShieldRegenSystem_OnShieldDamaged_DelaysRegen(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	sys.SetBaseRegenRate(10.0)
	sys.SetRegenDelay(3.0) // 3 second delay

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ShieldComponent{
		Amount:      50.0,
		MaxAmount:   100.0,
		Duration:    60.0,
		MaxDuration: 60.0,
	})

	entities := []*Entity{entity}

	// Simulate damage
	sys.OnShieldDamaged(entity, 20.0)

	// Update immediately after damage - should not regen
	sys.Update(entities, 1.0)

	shield := entity.GetComponent("shield").(*ShieldComponent)

	if shield.Amount != 50.0 {
		t.Errorf("shield.Amount = %f, want 50.0 (no regen during delay)", shield.Amount)
	}

	// Update to pass the delay period
	sys.Update(entities, 2.5) // Now at 3.5 seconds total game time

	// Shield should still be 50 because we just passed delay
	// Another update should allow regen
	sys.Update(entities, 1.0)

	shield = entity.GetComponent("shield").(*ShieldComponent)

	if shield.Amount <= 50.0 {
		t.Errorf("shield.Amount = %f, want > 50.0 (regen after delay)", shield.Amount)
	}
}

func TestShieldRegenSystem_OnShieldDamaged_NilEntity(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)

	// Should not panic
	sys.OnShieldDamaged(nil, 10.0)
}

func TestShieldRegenSystem_OnShieldDamaged_ZeroDamage(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	sys.SetRegenDelay(3.0)

	entity := world.CreateEntity()

	// Zero damage should not trigger delay
	sys.OnShieldDamaged(entity, 0.0)

	_, exists := sys.lastDamageTime[entity.ID]
	if exists {
		t.Error("zero damage should not record damage time")
	}
}

func TestShieldRegenSystem_getShieldParticleType(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)

	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "sparkle"},
		{"scifi", "spark"},
		{"horror", "smoke"},
		{"cyberpunk", "spark"},
		{"postapoc", "dust"},
		{"unknown", "sparkle"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			particleType := sys.getShieldParticleType()
			if particleType.String() != tt.expected {
				t.Errorf("particle type = %s, want %s", particleType.String(), tt.expected)
			}
		})
	}
}

func TestShieldRegenSystem_Update_CleansUpTracking(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ShieldComponent{
		Amount:      50.0,
		MaxAmount:   100.0,
		Duration:    60.0,
		MaxDuration: 60.0,
	})

	// Record damage time
	sys.lastDamageTime[entity.ID] = 0.0

	entities := []*Entity{entity}

	// Update with entity
	sys.Update(entities, 0.1)

	// Remove shield component
	entity.RemoveComponent("shield")

	// Update again - should clean up tracking
	sys.Update(entities, 0.1)

	_, exists := sys.lastDamageTime[entity.ID]
	if exists {
		t.Error("tracking should be cleaned up when shield component removed")
	}
}

func TestShieldRegenSystem_SpawnParticles_WithParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)
	sys.SetGenre("fantasy")
	sys.SetBaseRegenRate(10.0)
	sys.SetRegenDelay(0.0)

	// Set pulse interval to 0 to force particle spawn
	sys.pulseInterval = 0.0
	sys.timeSinceEmit = 0.0

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&ShieldComponent{
		Amount:      50.0,
		MaxAmount:   100.0,
		Duration:    60.0,
		MaxDuration: 60.0,
	})

	entities := []*Entity{entity}

	// Update should spawn particles (no panic expected)
	sys.Update(entities, 1.0)

	// Verify shield regenerated
	shield := entity.GetComponent("shield").(*ShieldComponent)
	if shield.Amount <= 50.0 {
		t.Errorf("shield.Amount = %f, want > 50.0", shield.Amount)
	}
}

func BenchmarkShieldRegenSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewShieldRegenSystem(world, 12345)
	sys.SetBaseRegenRate(10.0)
	sys.SetRegenDelay(0.0)

	// Create 100 entities with shields
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entity.AddComponent(&ShieldComponent{
			Amount:      50.0,
			MaxAmount:   100.0,
			Duration:    60.0,
			MaxDuration: 60.0,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016) // ~60fps
	}
}
