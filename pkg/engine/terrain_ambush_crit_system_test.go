package engine

import (
	"math"
	"testing"
)

func TestNewTerrainAmbushCritSystem(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)

	if system == nil {
		t.Fatal("NewTerrainAmbushCritSystem returned nil")
	}

	if system.world != world {
		t.Error("world not set correctly")
	}

	if system.rng == nil {
		t.Error("RNG not initialized")
	}

	if system.critBonuses == nil {
		t.Error("crit bonuses map not initialized")
	}
}

func TestTerrainAmbushCritSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		system.SetGenre(genre)
		if system.genreID != genre {
			t.Errorf("SetGenre(%s) did not set genre correctly", genre)
		}
	}
}

func TestTerrainAmbushCritSystem_CalculateCritBonus(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)

	tests := []struct {
		name        string
		stealthMult float64
		expectedMin float64
		expectedMax float64
	}{
		{"exceptional concealment", 0.4, 0.17, 0.19},
		{"heavy cover", 0.6, 0.11, 0.13},
		{"moderate cover", 0.8, 0.06, 0.08},
		{"light cover", 0.92, 0.02, 0.04},
		{"exposed", 1.0, 0.0, 0.0},
		{"very exposed", 1.3, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := system.calculateCritBonus(tt.stealthMult)
			if bonus < tt.expectedMin || bonus > tt.expectedMax {
				t.Errorf("calculateCritBonus(%f) = %f, want between %f and %f",
					tt.stealthMult, bonus, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestTerrainAmbushCritSystem_GenreMultipliers(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)

	tests := []struct {
		genre    string
		expected float64
	}{
		{"fantasy", 1.10},
		{"scifi", 0.90},
		{"horror", 1.20},
		{"cyberpunk", 1.15},
		{"postapoc", 1.05},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			mult := system.getGenreMultiplier()
			if math.Abs(mult-tt.expected) > 0.001 {
				t.Errorf("getGenreMultiplier() for %s = %f, want %f", tt.genre, mult, tt.expected)
			}
		})
	}
}

func TestTerrainAmbushCritSystem_UpdateWithoutStealthSystem(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&StatsComponent{CritChance: 0.10})

	// Should not panic when stealth system is nil
	system.Update([]*Entity{entity}, 0.5)

	stats := entity.GetStats()
	if stats.CritChance != 0.10 {
		t.Errorf("CritChance changed without stealth system: %f", stats.CritChance)
	}
}

func TestTerrainAmbushCritSystem_UpdateWithStealthSystem(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)
	stealthSystem := NewTerrainStealthSystem(world, 12345)

	system.SetTerrainStealthSystem(stealthSystem)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatsComponent{CritChance: 0.10})

	// Manually set stealth cache for testing
	stealthSystem.stealthCache[entity.ID] = 0.6 // Heavy cover

	// Update enough to trigger check
	system.Update([]*Entity{entity}, 0.3)

	stats := entity.GetStats()
	expectedBonus := 0.12 * 1.10 // heavy cover * fantasy multiplier
	expectedCrit := 0.10 + expectedBonus

	if math.Abs(stats.CritChance-expectedCrit) > 0.01 {
		t.Errorf("CritChance = %f, want ~%f", stats.CritChance, expectedCrit)
	}
}

func TestTerrainAmbushCritSystem_GetCritBonus(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)
	stealthSystem := NewTerrainStealthSystem(world, 12345)

	system.SetTerrainStealthSystem(stealthSystem)
	system.SetGenre("horror") // +20% bonus

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatsComponent{CritChance: 0.05})

	stealthSystem.stealthCache[entity.ID] = 0.4 // Exceptional concealment

	system.Update([]*Entity{entity}, 0.3)

	bonus := system.GetCritBonus(entity.ID)
	expectedBonus := 0.18 * 1.20 // exceptional * horror

	if math.Abs(bonus-expectedBonus) > 0.01 {
		t.Errorf("GetCritBonus() = %f, want ~%f", bonus, expectedBonus)
	}
}

func TestTerrainAmbushCritSystem_BonusRemoval(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)
	stealthSystem := NewTerrainStealthSystem(world, 12345)

	system.SetTerrainStealthSystem(stealthSystem)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatsComponent{CritChance: 0.10})

	// Apply bonus
	stealthSystem.stealthCache[entity.ID] = 0.6
	system.Update([]*Entity{entity}, 0.3)

	stats := entity.GetStats()
	critWithBonus := stats.CritChance

	// Remove concealment
	stealthSystem.stealthCache[entity.ID] = 1.2 // Exposed
	system.Update([]*Entity{entity}, 0.3)

	// Crit should return to near original
	if math.Abs(stats.CritChance-0.10) > 0.01 {
		t.Errorf("CritChance after losing cover = %f, started at %f with bonus",
			stats.CritChance, critWithBonus)
	}
}

func TestTerrainAmbushCritSystem_CritClamping(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)
	stealthSystem := NewTerrainStealthSystem(world, 12345)

	system.SetTerrainStealthSystem(stealthSystem)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatsComponent{CritChance: 0.95}) // Very high base

	stealthSystem.stealthCache[entity.ID] = 0.4 // Exceptional concealment (+18%)

	system.Update([]*Entity{entity}, 0.3)

	stats := entity.GetStats()
	if stats.CritChance > 1.0 {
		t.Errorf("CritChance exceeded 1.0: %f", stats.CritChance)
	}
}

func TestTerrainAmbushCritSystem_NoStatsEntity(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)
	stealthSystem := NewTerrainStealthSystem(world, 12345)

	system.SetTerrainStealthSystem(stealthSystem)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	// No StatsComponent

	stealthSystem.stealthCache[entity.ID] = 0.4

	// Should not panic
	system.Update([]*Entity{entity}, 0.3)

	bonus := system.GetCritBonus(entity.ID)
	if bonus != 0 {
		t.Errorf("GetCritBonus() = %f for entity without stats, want 0", bonus)
	}
}

func TestTerrainAmbushCritSystem_SetUpdateInterval(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)

	system.SetUpdateInterval(0.5)
	if system.updateInterval != 0.5 {
		t.Errorf("updateInterval = %f, want 0.5", system.updateInterval)
	}

	// Should ignore invalid values
	system.SetUpdateInterval(0)
	if system.updateInterval != 0.5 {
		t.Errorf("updateInterval changed to %f with invalid input", system.updateInterval)
	}

	system.SetUpdateInterval(-1)
	if system.updateInterval != 0.5 {
		t.Errorf("updateInterval changed to %f with negative input", system.updateInterval)
	}
}

func TestTerrainAmbushCritSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)
	stealthSystem := NewTerrainStealthSystem(world, 12345)

	system.SetTerrainStealthSystem(stealthSystem)

	// Create entities with different concealment levels
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&StatsComponent{CritChance: 0.10})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 200, Y: 200})
	entity2.AddComponent(&StatsComponent{CritChance: 0.10})

	entity3 := world.CreateEntity()
	entity3.AddComponent(&PositionComponent{X: 300, Y: 300})
	entity3.AddComponent(&StatsComponent{CritChance: 0.10})

	stealthSystem.stealthCache[entity1.ID] = 0.4 // Exceptional
	stealthSystem.stealthCache[entity2.ID] = 0.8 // Moderate
	stealthSystem.stealthCache[entity3.ID] = 1.0 // Exposed

	system.Update([]*Entity{entity1, entity2, entity3}, 0.3)

	stats1 := entity1.GetStats()
	stats2 := entity2.GetStats()
	stats3 := entity3.GetStats()

	// Entity1 should have highest bonus
	if stats1.CritChance <= stats2.CritChance {
		t.Errorf("Entity1 (exceptional) crit %f should be > entity2 (moderate) %f",
			stats1.CritChance, stats2.CritChance)
	}

	// Entity3 should have no bonus
	if math.Abs(stats3.CritChance-0.10) > 0.001 {
		t.Errorf("Entity3 (exposed) crit = %f, should be base 0.10", stats3.CritChance)
	}
}

// BenchmarkTerrainAmbushCritSystem_Update benchmarks the update performance.
func BenchmarkTerrainAmbushCritSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewTerrainAmbushCritSystem(world, 12345)
	stealthSystem := NewTerrainStealthSystem(world, 12345)

	system.SetTerrainStealthSystem(stealthSystem)
	system.SetGenre("fantasy")

	// Create 100 entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entity.AddComponent(&StatsComponent{CritChance: 0.10})
		stealthSystem.stealthCache[entity.ID] = 0.5 + float64(i%5)*0.15
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.3)
	}
}
