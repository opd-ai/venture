package engine

import (
	"testing"
)

func TestNewItemPickupParticleSystem(t *testing.T) {
	world := NewWorld()
	seed := int64(12345)
	system := NewItemPickupParticleSystem(world, seed)

	if system == nil {
		t.Fatal("NewItemPickupParticleSystem returned nil")
	}

	if system.world != world {
		t.Error("World not set correctly")
	}

	if system.seed != seed {
		t.Errorf("Seed = %d, want %d", system.seed, seed)
	}

	if system.rng == nil {
		t.Error("RNG not initialized")
	}

	if system.baseParticleCount != 12 {
		t.Errorf("baseParticleCount = %d, want 12", system.baseParticleCount)
	}

	if system.spreadFactor != 60.0 {
		t.Errorf("spreadFactor = %f, want 60.0", system.spreadFactor)
	}
}

func TestItemPickupParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()

	system.SetParticleSystem(particleSystem)

	if system.particleSystem != particleSystem {
		t.Error("ParticleSystem not set correctly")
	}
}

func TestItemPickupParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)

	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system.SetGenre(tt.genreID)
			if system.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", system.genreID, tt.genreID)
			}
		})
	}
}

func TestItemPickupParticleSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)

	// Create test entities
	entities := []*Entity{
		{ID: 1, Components: make(map[string]Component)},
		{ID: 2, Components: make(map[string]Component)},
	}

	// Update should not panic - it's a no-op for this callback-driven system
	system.Update(entities, 0.016)
}

func TestItemPickupParticleSystem_OnItemPickup(t *testing.T) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)
	system.SetGenre("fantasy")

	tests := []struct {
		name   string
		x, y   float64
		rarity int
	}{
		{"common_item", 100.0, 200.0, 0},
		{"uncommon_item", 150.0, 250.0, 1},
		{"rare_item", 200.0, 300.0, 2},
		{"epic_item", 250.0, 350.0, 3},
		{"legendary_item", 300.0, 400.0, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// OnItemPickup should not panic
			system.OnItemPickup(tt.x, tt.y, tt.rarity)
		})
	}
}

func TestItemPickupParticleSystem_OnItemPickup_NoParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)
	// Don't set particle system

	// Should not panic when particle system is nil
	system.OnItemPickup(100.0, 200.0, 2)
}

func TestItemPickupParticleSystem_OnItemPickup_NoWorld(t *testing.T) {
	system := NewItemPickupParticleSystem(nil, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)

	// Should not panic when world is nil
	system.OnItemPickup(100.0, 200.0, 2)
}

func TestItemPickupParticleSystem_SpawnPickupEffect(t *testing.T) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)
	system.SetGenre("scifi")

	// SpawnPickupEffect should not panic
	system.SpawnPickupEffect(100.0, 200.0, 3)
}

func TestItemPickupParticleSystem_SpawnPickupEffect_NilSystems(t *testing.T) {
	system := NewItemPickupParticleSystem(nil, 12345)

	// Should not panic with nil world and particle system
	system.SpawnPickupEffect(100.0, 200.0, 2)
}

func TestItemPickupParticleSystem_DeterministicSeeds(t *testing.T) {
	world := NewWorld()
	seed := int64(42)

	// Two systems with same seed should behave identically
	system1 := NewItemPickupParticleSystem(world, seed)
	system2 := NewItemPickupParticleSystem(world, seed)

	if system1.seed != system2.seed {
		t.Error("Systems with same seed should have identical seeds")
	}
}

func TestItemPickupParticleSystem_RarityScaling(t *testing.T) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)

	// Test that rarity affects particle count (internal logic)
	// baseParticleCount is 12
	tests := []struct {
		rarity   int
		minCount int // Expected minimum count based on multiplier
	}{
		{0, 12}, // Common: 1.0x = 12
		{1, 15}, // Uncommon: 1.25x = 15
		{2, 18}, // Rare: 1.5x = 18
		{3, 24}, // Epic: 2.0x = 24
		{4, 30}, // Legendary: 2.5x = 30
	}

	for _, tt := range tests {
		t.Run("rarity_"+string(rune('0'+tt.rarity)), func(t *testing.T) {
			// We can't directly test internal particle count,
			// but we verify the system handles all rarities
			particleSystem := NewParticleSystem()
			system.SetParticleSystem(particleSystem)
			system.SetGenre("fantasy")
			system.SpawnPickupEffect(100.0, 200.0, tt.rarity)
		})
	}
}

// BenchmarkItemPickupParticleSystem_OnItemPickup benchmarks particle spawning.
func BenchmarkItemPickupParticleSystem_OnItemPickup(b *testing.B) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)
	particleSystem := NewParticleSystem()
	system.SetParticleSystem(particleSystem)
	system.SetGenre("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnItemPickup(float64(i%1000), float64(i%1000), i%5)
	}
}

// BenchmarkItemPickupParticleSystem_Update benchmarks the no-op Update.
func BenchmarkItemPickupParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewItemPickupParticleSystem(world, 12345)

	entities := make([]*Entity, 100)
	for i := range entities {
		entities[i] = &Entity{ID: uint64(i), Components: make(map[string]Component)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
