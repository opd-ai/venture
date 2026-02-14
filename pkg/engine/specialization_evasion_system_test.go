package engine

import (
	"math/rand"
	"testing"
)

func TestNewSpecializationEvasionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewSpecializationEvasionSystem returned nil")
	}
	if sys.world != world {
		t.Error("World reference not set")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.originalEvasion == nil {
		t.Error("originalEvasion map not initialized")
	}
	if sys.appliedBonuses == nil {
		t.Error("appliedBonuses map not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestSpecializationEvasionSystem_SetGenre(t *testing.T) {
	sys := NewSpecializationEvasionSystem(nil, 12345)

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
		sys.SetGenre(tt.genre)
		if sys.genreID != tt.genre {
			t.Errorf("Genre = %q, want %q", sys.genreID, tt.genre)
		}
	}
}

func TestSpecializationEvasionSystem_NoBonus_WithoutRequiredComponents(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)

	// Entity with only position (no stats or class_progression)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	bonus := sys.GetEvasionBonus(entity.ID)
	if bonus != 0.0 {
		t.Errorf("Expected no bonus without components, got %f", bonus)
	}
}

func TestSpecializationEvasionSystem_BaseClassBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	tests := []struct {
		name          string
		class         CharacterClass
		expectedBonus float64
	}{
		{"Rogue base bonus", ClassRogue, 0.08},
		{"Ranger base bonus", ClassRanger, 0.05},
		{"Monk base bonus", ClassMonk, 0.06},
		{"Spellblade base bonus", ClassSpellblade, 0.04},
		{"Warrior no bonus", ClassWarrior, 0.0},
		{"Mage no bonus", ClassMage, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&StatsComponent{
				Evasion: 0.10, // 10% base evasion
			})
			entity.AddComponent(&ClassProgressionComponent{
				Class:          tt.class,
				Specialization: SpecializationNone,
			})

			sys.Update([]*Entity{entity}, 1.0)

			bonus := sys.GetEvasionBonus(entity.ID)
			if bonus != tt.expectedBonus {
				t.Errorf("Bonus = %f, want %f", bonus, tt.expectedBonus)
			}

			// Check that stats were updated
			stats := entity.GetStats()
			expectedEvasion := 0.10 + tt.expectedBonus
			if stats.Evasion != expectedEvasion {
				t.Errorf("Evasion = %f, want %f", stats.Evasion, expectedEvasion)
			}
		})
	}
}

func TestSpecializationEvasionSystem_SpecializationBonuses(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	tests := []struct {
		name           string
		specialization SpecializationType
		expectedBonus  float64
	}{
		{"Shadowdancer max evasion", SpecializationShadowdancer, 0.30},
		{"Windwalker high evasion", SpecializationWindwalker, 0.25},
		{"Assassin good evasion", SpecializationAssassin, 0.20},
		{"Trickster illusion evasion", SpecializationTrickster, 0.20},
		{"Marksman mobile evasion", SpecializationMarksman, 0.15},
		{"Duelist parry evasion", SpecializationDuelist, 0.15},
		{"Beastmaster companion evasion", SpecializationBeastmaster, 0.10},
		{"Exorcist holy evasion", SpecializationExorcist, 0.10},
		{"Defender no evasion", SpecializationDefender, 0.0},
		{"Elementalist no evasion", SpecializationElementalist, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&StatsComponent{
				Evasion: 0.10,
			})
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassRogue,
				Specialization: tt.specialization,
			})

			sys.Update([]*Entity{entity}, 1.0)

			bonus := sys.GetEvasionBonus(entity.ID)
			if bonus != tt.expectedBonus {
				t.Errorf("Bonus = %f, want %f", bonus, tt.expectedBonus)
			}
		})
	}
}

func TestSpecializationEvasionSystem_GenreMultipliers(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)

	// Use Shadowdancer (30% base) to test genre multipliers
	tests := []struct {
		genre         string
		expectedBonus float64
	}{
		{"fantasy", 0.30},    // 0.30 * 1.0
		{"scifi", 0.33},      // 0.30 * 1.10
		{"cyberpunk", 0.345}, // 0.30 * 1.15
		{"horror", 0.27},     // 0.30 * 0.90
		{"postapoc", 0.285},  // 0.30 * 0.95
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			// Reset caches
			sys.originalEvasion = make(map[uint64]float64)
			sys.appliedBonuses = make(map[uint64]float64)

			entity := world.CreateEntity()
			entity.AddComponent(&StatsComponent{
				Evasion: 0.10,
			})
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassRogue,
				Specialization: SpecializationShadowdancer,
			})

			sys.Update([]*Entity{entity}, 1.0)

			bonus := sys.GetEvasionBonus(entity.ID)
			// Allow small floating point tolerance
			if bonus < tt.expectedBonus-0.001 || bonus > tt.expectedBonus+0.001 {
				t.Errorf("Bonus = %f, want %f", bonus, tt.expectedBonus)
			}
		})
	}
}

func TestSpecializationEvasionSystem_EvasionCap(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	// High base evasion + high bonus should cap at 85%
	entity := world.CreateEntity()
	entity.AddComponent(&StatsComponent{
		Evasion: 0.70, // 70% base evasion
	})
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationShadowdancer, // +30%
	})

	sys.Update([]*Entity{entity}, 1.0)

	stats := entity.GetStats()
	if stats.Evasion > 0.85 {
		t.Errorf("Evasion = %f, should cap at 0.85", stats.Evasion)
	}
	if stats.Evasion != 0.85 {
		t.Errorf("Evasion = %f, want 0.85", stats.Evasion)
	}
}

func TestSpecializationEvasionSystem_UpdateThrottling(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&StatsComponent{Evasion: 0.10})
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationShadowdancer,
	})

	entities := []*Entity{entity}

	// First update with small deltaTime - shouldn't process
	sys.Update(entities, 0.1)
	bonus := sys.GetEvasionBonus(entity.ID)
	if bonus != 0.0 {
		t.Error("Should not update before interval elapsed")
	}

	// Accumulate time but not enough
	sys.Update(entities, 0.5)
	bonus = sys.GetEvasionBonus(entity.ID)
	if bonus != 0.0 {
		t.Error("Should not update before interval elapsed")
	}

	// Now exceed interval
	sys.Update(entities, 0.5)
	bonus = sys.GetEvasionBonus(entity.ID)
	if bonus == 0.0 {
		t.Error("Should update after interval elapsed")
	}
}

func TestSpecializationEvasionSystem_BonusRemovalOnComponentLoss(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&StatsComponent{Evasion: 0.10})
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationShadowdancer,
	})

	// Apply bonus
	sys.Update([]*Entity{entity}, 1.0)
	if sys.GetEvasionBonus(entity.ID) == 0.0 {
		t.Fatal("Bonus should be applied")
	}

	// Remove class_progression component
	entity.RemoveComponent("class_progression")

	// Update should clean up
	sys.Update([]*Entity{entity}, 1.0)
	if sys.GetEvasionBonus(entity.ID) != 0.0 {
		t.Error("Bonus should be removed when component is missing")
	}
}

func TestSpecializationEvasionSystem_Determinism(t *testing.T) {
	seed := int64(42)

	results := make([]float64, 5)
	for i := 0; i < 5; i++ {
		world := NewWorld()
		sys := NewSpecializationEvasionSystem(world, seed)

		entity := world.CreateEntity()
		entity.AddComponent(&StatsComponent{Evasion: 0.10})
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassRogue,
			Specialization: SpecializationShadowdancer,
		})

		sys.Update([]*Entity{entity}, 1.0)
		results[i] = sys.GetEvasionBonus(entity.ID)
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Result[%d] = %f differs from Result[0] = %f", i, results[i], results[0])
		}
	}
}

func BenchmarkSpecializationEvasionSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewSpecializationEvasionSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create 100 entities with varying specializations
	entities := make([]*Entity, 100)
	rng := rand.New(rand.NewSource(99))
	specs := []SpecializationType{
		SpecializationNone, SpecializationShadowdancer, SpecializationWindwalker,
		SpecializationAssassin, SpecializationMarksman, SpecializationDefender,
	}

	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&StatsComponent{Evasion: 0.1 + float64(rng.Intn(20))/100})
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassRogue,
			Specialization: specs[rng.Intn(len(specs))],
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 0 // Force update
		sys.Update(entities, 1.0)
	}
}
