package engine

import (
	"math/rand"
	"testing"
)

func TestNewSpecializationDefenseSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationDefenseSystem(world, 12345)

	if system == nil {
		t.Fatal("NewSpecializationDefenseSystem returned nil")
	}

	if system.world != world {
		t.Error("System world reference not set correctly")
	}

	if system.updateInterval != 1.0 {
		t.Errorf("Expected updateInterval 1.0, got %f", system.updateInterval)
	}

	// Verify the bonus cache maps are initialized. A nil map read also returns zero,
	// so checking GetBonusForEntity alone does not prove initialization.
	if system.defenseMod.applied == nil {
		t.Error("defenseMod.applied cache not initialized")
	}

	if system.defenseMod.original == nil {
		t.Error("defenseMod.original cache not initialized")
	}

	// Unknown entities should still report zero bonus.
	if system.GetBonusForEntity(99999) != 0 {
		t.Error("defenseMod cache not initialized: expected zero bonus for unknown entity")
	}

	if system.genreID != "fantasy" {
		t.Errorf("Expected default genre 'fantasy', got '%s'", system.genreID)
	}
}

func TestSpecializationDefenseSystem_SetGenre(t *testing.T) {
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
			system := NewSpecializationDefenseSystem(nil, 12345)
			system.SetGenre(tt.genreID)

			if system.genreID != tt.genreID {
				t.Errorf("Expected genre '%s', got '%s'", tt.genreID, system.genreID)
			}
		})
	}
}

func TestSpecializationDefenseSystem_Update_IntervalThrottling(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationDefenseSystem(world, 12345)

	// Create entity with required components
	entity := world.CreateEntity()
	entity.AddComponent(NewStatsComponent())
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          10,
		Specialization: SpecializationDefender,
	})

	entities := []*Entity{entity}

	// First update with small delta - should not process
	system.Update(entities, 0.5)

	if system.GetBonusForEntity(entity.ID) != 0 {
		t.Error("Should not apply bonus before interval elapsed")
	}

	// Second update - now interval should be met
	system.Update(entities, 0.6)

	if system.GetBonusForEntity(entity.ID) == 0 {
		t.Error("Should apply bonus after interval elapsed")
	}
}

func TestSpecializationDefenseSystem_TankSpecializations(t *testing.T) {
	tests := []struct {
		name           string
		specialization SpecializationType
		minBonus       float64 // Minimum expected bonus
	}{
		{"defender", SpecializationDefender, 0.70},
		{"guardian", SpecializationGuardian, 0.60},
		{"brewmaster", SpecializationBrewmaster, 0.55},
		{"pack_leader", SpecializationPackLeader, 0.45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewSpecializationDefenseSystem(world, 12345)

			entity := world.CreateEntity()
			stats := NewStatsComponent()
			stats.Defense = 100.0
			entity.AddComponent(stats)
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassWarrior,
				Level:          15,
				Specialization: tt.specialization,
			})

			entities := []*Entity{entity}
			system.Update(entities, 2.0)

			bonus := system.GetBonusForEntity(entity.ID)
			if bonus < tt.minBonus {
				t.Errorf("Expected bonus >= %f for %s, got %f", tt.minBonus, tt.name, bonus)
			}

			// Verify defense was actually increased
			if stats.Defense <= 100.0 {
				t.Errorf("Defense should be increased from 100, got %f", stats.Defense)
			}
		})
	}
}

func TestSpecializationDefenseSystem_CasterSpecializations(t *testing.T) {
	tests := []struct {
		name           string
		specialization SpecializationType
		maxBonus       float64 // Maximum expected bonus (casters should have low)
	}{
		{"arcanist", SpecializationArcanist, 0.15},
		{"elementalist", SpecializationElementalist, 0.15},
		{"affliction", SpecializationAffliction, 0.15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewSpecializationDefenseSystem(world, 12345)

			entity := world.CreateEntity()
			stats := NewStatsComponent()
			stats.Defense = 50.0
			entity.AddComponent(stats)
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassMage,
				Level:          15,
				Specialization: tt.specialization,
			})

			entities := []*Entity{entity}
			system.Update(entities, 2.0)

			bonus := system.GetBonusForEntity(entity.ID)
			if bonus > tt.maxBonus {
				t.Errorf("Expected bonus <= %f for %s, got %f", tt.maxBonus, tt.name, bonus)
			}
		})
	}
}

func TestSpecializationDefenseSystem_NoSpecialization(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationDefenseSystem(world, 12345)

	// Warrior without specialization should get base class bonus
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Defense = 100.0
	entity.AddComponent(stats)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          5, // Below specialization level
		Specialization: SpecializationNone,
	})

	entities := []*Entity{entity}
	system.Update(entities, 2.0)

	bonus := system.GetBonusForEntity(entity.ID)
	// Warrior base class bonus is 0.15 (15%)
	if bonus < 0.10 || bonus > 0.20 {
		t.Errorf("Expected base class bonus around 0.15, got %f", bonus)
	}

	// Mage without specialization should get no bonus
	mage := world.CreateEntity()
	mageStats := NewStatsComponent()
	mageStats.Defense = 50.0
	mage.AddComponent(mageStats)
	mage.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          5,
		Specialization: SpecializationNone,
	})

	entities = []*Entity{mage}
	system.Update(entities, 2.0)

	mageBonus := system.GetBonusForEntity(mage.ID)
	if mageBonus > 0.05 {
		t.Errorf("Mage should have minimal base bonus, got %f", mageBonus)
	}
}

func TestSpecializationDefenseSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name       string
		genreID    string
		multiplier float64 // Expected genre multiplier
	}{
		{"fantasy_standard", "fantasy", 1.0},
		{"scifi_tech_armor", "scifi", 1.2},
		{"horror_reduced", "horror", 0.8},
		{"cyberpunk_enhanced", "cyberpunk", 1.1},
		{"postapoc_degraded", "postapoc", 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewSpecializationDefenseSystem(world, 12345)
			system.SetGenre(tt.genreID)

			entity := world.CreateEntity()
			stats := NewStatsComponent()
			stats.Defense = 100.0
			entity.AddComponent(stats)
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassWarrior,
				Level:          15,
				Specialization: SpecializationDefender,
			})

			entities := []*Entity{entity}
			system.Update(entities, 2.0)

			// Defender base is 0.80, multiplied by genre
			expectedBonus := 0.80 * tt.multiplier
			actualBonus := system.GetBonusForEntity(entity.ID)

			// Allow small floating point tolerance
			if actualBonus < expectedBonus-0.01 || actualBonus > expectedBonus+0.01 {
				t.Errorf("Expected bonus ~%f for %s, got %f", expectedBonus, tt.genreID, actualBonus)
			}
		})
	}
}

func TestSpecializationDefenseSystem_MissingComponents(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationDefenseSystem(world, 12345)

	// Entity with only stats (no class_progression)
	entity1 := world.CreateEntity()
	entity1.AddComponent(NewStatsComponent())

	// Entity with only class_progression (no stats)
	entity2 := world.CreateEntity()
	entity2.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Specialization: SpecializationDefender,
	})

	// Entity with neither
	entity3 := world.CreateEntity()

	entities := []*Entity{entity1, entity2, entity3}
	system.Update(entities, 2.0)

	// None should have bonuses applied
	if system.GetBonusForEntity(entity1.ID) != 0 {
		t.Error("Entity without class_progression should not get bonus")
	}
	if system.GetBonusForEntity(entity2.ID) != 0 {
		t.Error("Entity without stats should not get bonus")
	}
	if system.GetBonusForEntity(entity3.ID) != 0 {
		t.Error("Entity without components should not get bonus")
	}
}

func TestSpecializationDefenseSystem_BonusRemoval(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationDefenseSystem(world, 12345)

	// Create entity with both components
	entity := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Defense = 100.0
	entity.AddComponent(stats)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationDefender,
	})

	entities := []*Entity{entity}
	system.Update(entities, 2.0)

	// Verify bonus was applied
	if system.GetBonusForEntity(entity.ID) == 0 {
		t.Fatal("Bonus should be applied initially")
	}
	if stats.Defense <= 100.0 {
		t.Fatal("Defense should have been boosted above 100")
	}

	// Remove class_progression component
	entity.RemoveComponent("class_progression")

	// Update again — processEntity should call restoreAndRemove
	system.Update(entities, 2.0)

	// Bonus tracking should be removed
	if system.GetBonusForEntity(entity.ID) != 0 {
		t.Error("Bonus tracking should be removed when component is removed")
	}

	// The original defense value should be restored
	if stats.Defense != 100.0 {
		t.Errorf("Defense should be restored to 100 after component removal, got %f", stats.Defense)
	}
}

func TestSpecializationDefenseSystem_OriginalDefensePreserved(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationDefenseSystem(world, 12345)

	entity := world.CreateEntity()
	stats := NewStatsComponent()
	originalDefense := 75.0
	stats.Defense = originalDefense
	entity.AddComponent(stats)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationDefender,
	})

	entities := []*Entity{entity}

	// Apply bonus multiple times
	system.Update(entities, 2.0)
	firstDefense := stats.Defense

	system.Update(entities, 2.0)
	secondDefense := stats.Defense

	// Defense should be same (not compounding)
	if firstDefense != secondDefense {
		t.Errorf("Defense should not compound: first=%f, second=%f", firstDefense, secondDefense)
	}

	// Verify the original was preserved: applying the same spec twice should produce
	// the same result as applying it once (bonus = originalDefense * (1+spec_bonus)).
	bonus := system.GetBonusForEntity(entity.ID)
	expectedDefense := originalDefense * (1.0 + bonus)
	if firstDefense != expectedDefense {
		t.Errorf("Original defense not preserved correctly: expected %f, got %f", expectedDefense, firstDefense)
	}
}

func TestSpecializationDefenseSystem_DeterministicRNG(t *testing.T) {
	// Create two systems with same seed
	world1 := NewWorld()
	system1 := NewSpecializationDefenseSystem(world1, 54321)

	world2 := NewWorld()
	system2 := NewSpecializationDefenseSystem(world2, 54321)

	// Verify both have same RNG state
	val1 := system1.rng.Float64()
	val2 := system2.rng.Float64()

	if val1 != val2 {
		t.Errorf("Systems with same seed should produce same RNG values: %f != %f", val1, val2)
	}
}

func TestSpecializationDefenseSystem_AllSpecializationsHaveBonus(t *testing.T) {
	// Ensure all specializations have defined bonuses (no panic)
	system := NewSpecializationDefenseSystem(nil, 12345)

	specializations := []SpecializationType{
		SpecializationNone,
		SpecializationDefender, SpecializationBerserker,
		SpecializationGuardian, SpecializationCrusader, SpecializationTemplar,
		SpecializationAssassin, SpecializationShadowdancer, SpecializationTrickster, SpecializationDuelist,
		SpecializationBeastmaster, SpecializationMarksman,
		SpecializationArcanist, SpecializationElementalist, SpecializationBloodMage,
		SpecializationHealer, SpecializationOracle, SpecializationTheurgist,
		SpecializationWindwalker, SpecializationBrewmaster,
		SpecializationDeathKnight, SpecializationUnholy, SpecializationFrost,
		SpecializationWarmage, SpecializationSpellsword,
		SpecializationAffliction, SpecializationDemonologist, SpecializationVoidcaller,
		SpecializationFeralRage, SpecializationPackLeader, SpecializationShapeshifter,
		SpecializationSoulweaver, SpecializationHemomancer, SpecializationNaturemage,
		SpecializationCrimsonBlade,
		SpecializationExorcist, SpecializationPurifier, SpecializationJudge, SpecializationInterrogator,
		SpecializationSpellshot, SpecializationSeeker,
		SpecializationShinobi, SpecializationStriker,
	}

	for _, spec := range specializations {
		// Should not panic
		bonus := system.getSpecializationBonus(spec)
		if bonus < 0 {
			t.Errorf("Specialization %v has negative bonus: %f", spec, bonus)
		}
		if bonus > 1.0 {
			t.Errorf("Specialization %v has bonus > 100%%: %f", spec, bonus)
		}
	}
}

func BenchmarkSpecializationDefenseSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewSpecializationDefenseSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create 100 entities with varying specializations
	entities := make([]*Entity, 100)
	specs := []SpecializationType{
		SpecializationDefender, SpecializationBerserker, SpecializationArcanist,
		SpecializationHealer, SpecializationAssassin, SpecializationMarksman,
	}
	rng := rand.New(rand.NewSource(12345))

	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		stats := NewStatsComponent()
		stats.Defense = 50.0 + float64(i)
		entity.AddComponent(stats)
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassWarrior,
			Level:          10 + i%10,
			Specialization: specs[rng.Intn(len(specs))],
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.timeSinceCheck = 0 // Force processing
		system.Update(entities, 1.5)
	}
}
