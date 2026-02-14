//go:build ignore

package engine

import (
	"testing"
)

func TestNewSpecializationCritDamageSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	if system == nil {
		t.Fatal("NewSpecializationCritDamageSystem returned nil")
	}
	if system.world != world {
		t.Error("World not set correctly")
	}
	if system.updateInterval != 1.0 {
		t.Errorf("updateInterval = %v, want 1.0", system.updateInterval)
	}
	if system.appliedBonuses == nil {
		t.Error("appliedBonuses map not initialized")
	}
	if system.genreID != "fantasy" {
		t.Errorf("genreID = %v, want fantasy", system.genreID)
	}
}

func TestSpecializationCritDamageSystem_SetGenre(t *testing.T) {
	system := NewSpecializationCritDamageSystem(nil, 12345)

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
			system.SetGenre(tt.genre)
			if system.genreID != tt.genre {
				t.Errorf("genreID = %v, want %v", system.genreID, tt.genre)
			}
		})
	}
}

func TestSpecializationCritDamageSystem_Update_NoComponents(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	// Entity without required components
	entity := NewEntity()
	world.AddEntity(entity)

	// Should not panic
	system.Update([]*Entity{entity}, 1.0)

	if len(system.appliedBonuses) != 0 {
		t.Error("Should not apply bonus to entity without required components")
	}
}

func TestSpecializationCritDamageSystem_Update_WithSpecialization(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationAssassin,
	})
	entity.AddComponent(&StatsComponent{
		CritDamage: 2.0, // 200% base crit damage
	})
	world.AddEntity(entity)

	// First update with sufficient time
	system.Update([]*Entity{entity}, 1.0)

	stats := entity.GetComponentByType("stats").(*StatsComponent)
	// Assassin gives 0.50 bonus at fantasy (1.0 multiplier)
	expectedBonus := 0.50
	expectedCritDamage := 2.0 + expectedBonus

	if stats.CritDamage != expectedCritDamage {
		t.Errorf("CritDamage = %v, want %v", stats.CritDamage, expectedCritDamage)
	}

	if system.GetCritDamageBonus(entity.ID) != expectedBonus {
		t.Errorf("GetCritDamageBonus = %v, want %v", system.GetCritDamageBonus(entity.ID), expectedBonus)
	}
}

func TestSpecializationCritDamageSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationAssassin,
	})
	entity.AddComponent(&StatsComponent{
		CritDamage: 2.0,
	})
	world.AddEntity(entity)

	// Small delta time - should not update yet
	system.Update([]*Entity{entity}, 0.5)

	if len(system.appliedBonuses) != 0 {
		t.Error("Should not update before interval elapsed")
	}

	// Another small update to reach interval
	system.Update([]*Entity{entity}, 0.5)

	if len(system.appliedBonuses) == 0 {
		t.Error("Should update after interval elapsed")
	}
}

func TestSpecializationCritDamageSystem_BaseClassBonus(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	tests := []struct {
		class         CharacterClass
		expectedBonus float64
	}{
		{ClassRogue, 0.10},
		{ClassNinja, 0.12},
		{ClassRanger, 0.08},
		{ClassMonk, 0.05},
		{ClassWarrior, 0.03},
		{ClassMage, 0.0},
		{ClassCleric, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.class.String(), func(t *testing.T) {
			entity := NewEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          tt.class,
				Specialization: SpecializationNone,
			})
			entity.AddComponent(&StatsComponent{
				CritDamage: 2.0,
			})
			world.AddEntity(entity)

			system.Update([]*Entity{entity}, 1.0)

			stats := entity.GetComponentByType("stats").(*StatsComponent)
			expectedCritDamage := 2.0 + tt.expectedBonus

			if stats.CritDamage != expectedCritDamage {
				t.Errorf("CritDamage for %v = %v, want %v", tt.class, stats.CritDamage, expectedCritDamage)
			}

			// Clean up
			system.removeBonus(entity.ID)
			world.RemoveEntity(entity.ID)
		})
	}
}

func TestSpecializationCritDamageSystem_SpecializationBonuses(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	tests := []struct {
		spec          SpecializationType
		expectedBonus float64
	}{
		{SpecializationAssassin, 0.50},
		{SpecializationShadowdancer, 0.45},
		{SpecializationShinobi, 0.48},
		{SpecializationMarksman, 0.40},
		{SpecializationDefender, 0.05},
		{SpecializationHealer, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.spec.String(), func(t *testing.T) {
			entity := NewEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassRogue,
				Specialization: tt.spec,
			})
			entity.AddComponent(&StatsComponent{
				CritDamage: 2.0,
			})
			world.AddEntity(entity)

			system.Update([]*Entity{entity}, 1.0)

			stats := entity.GetComponentByType("stats").(*StatsComponent)
			expectedCritDamage := 2.0 + tt.expectedBonus

			if stats.CritDamage != expectedCritDamage {
				t.Errorf("CritDamage for %v = %v, want %v", tt.spec, stats.CritDamage, expectedCritDamage)
			}

			// Clean up
			system.removeBonus(entity.ID)
			world.RemoveEntity(entity.ID)
		})
	}
}

func TestSpecializationCritDamageSystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre      string
		multiplier float64
	}{
		{"fantasy", 1.0},
		{"scifi", 1.15},
		{"horror", 1.2},
		{"cyberpunk", 1.25},
		{"postapoc", 1.1},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			system := NewSpecializationCritDamageSystem(world, 12345)
			system.SetGenre(tt.genre)

			entity := NewEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassRogue,
				Specialization: SpecializationAssassin, // 0.50 base
			})
			entity.AddComponent(&StatsComponent{
				CritDamage: 2.0,
			})
			world.AddEntity(entity)

			system.Update([]*Entity{entity}, 1.0)

			stats := entity.GetComponentByType("stats").(*StatsComponent)
			expectedBonus := 0.50 * tt.multiplier
			expectedCritDamage := 2.0 + expectedBonus

			if stats.CritDamage != expectedCritDamage {
				t.Errorf("CritDamage for %v = %v, want %v", tt.genre, stats.CritDamage, expectedCritDamage)
			}
		})
	}
}

func TestSpecializationCritDamageSystem_RemoveBonus(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationAssassin,
	})
	entity.AddComponent(&StatsComponent{
		CritDamage: 2.0,
	})
	world.AddEntity(entity)

	// Apply bonus
	system.Update([]*Entity{entity}, 1.0)

	stats := entity.GetComponentByType("stats").(*StatsComponent)
	if stats.CritDamage == 2.0 {
		t.Error("Bonus should have been applied")
	}

	// Remove bonus
	system.removeBonus(entity.ID)

	if stats.CritDamage != 2.0 {
		t.Errorf("CritDamage after remove = %v, want 2.0", stats.CritDamage)
	}

	if system.IsSpecializationActive(entity.ID) {
		t.Error("Specialization should not be active after removal")
	}
}

func TestSpecializationCritDamageSystem_SpecializationChange(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	entity := NewEntity()
	progression := &ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationAssassin,
	}
	entity.AddComponent(progression)
	entity.AddComponent(&StatsComponent{
		CritDamage: 2.0,
	})
	world.AddEntity(entity)

	// Apply assassin bonus (0.50)
	system.Update([]*Entity{entity}, 1.0)

	stats := entity.GetComponentByType("stats").(*StatsComponent)
	expectedAssassin := 2.0 + 0.50
	if stats.CritDamage != expectedAssassin {
		t.Errorf("CritDamage with Assassin = %v, want %v", stats.CritDamage, expectedAssassin)
	}

	// Change to defender (0.05)
	progression.Specialization = SpecializationDefender
	system.timeSinceCheck = system.updateInterval // Force update

	system.Update([]*Entity{entity}, 1.0)

	expectedDefender := 2.0 + 0.05
	if stats.CritDamage != expectedDefender {
		t.Errorf("CritDamage with Defender = %v, want %v", stats.CritDamage, expectedDefender)
	}
}

func TestSpecializationCritDamageSystem_EntityWithoutStats(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationAssassin,
	})
	// No stats component
	world.AddEntity(entity)

	// Should not panic
	system.Update([]*Entity{entity}, 1.0)

	if len(system.appliedBonuses) != 0 {
		t.Error("Should not apply bonus to entity without stats")
	}
}

func TestSpecializationCritDamageSystem_GetBonusForEntity(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationAssassin,
	})
	entity.AddComponent(&StatsComponent{
		CritDamage: 2.0,
	})
	world.AddEntity(entity)

	// Before update
	if system.GetBonusForEntity(entity.ID) != 0.0 {
		t.Error("Should return 0 before update")
	}

	// After update
	system.Update([]*Entity{entity}, 1.0)

	bonus := system.GetBonusForEntity(entity.ID)
	expected := 0.50 // Assassin bonus
	if bonus != expected {
		t.Errorf("GetBonusForEntity = %v, want %v", bonus, expected)
	}
}

func TestSpecializationCritDamageSystem_NilWorld(t *testing.T) {
	system := NewSpecializationCritDamageSystem(nil, 12345)

	if system == nil {
		t.Fatal("Should create system even with nil world")
	}

	// Should not panic
	entity := NewEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationAssassin,
	})
	entity.AddComponent(&StatsComponent{
		CritDamage: 2.0,
	})

	system.Update([]*Entity{entity}, 1.0)
}

func TestSpecializationCritDamageSystem_AllSpecializations(t *testing.T) {
	// Test that all specializations have defined bonuses (no panics)
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	specializations := []SpecializationType{
		SpecializationNone,
		SpecializationBerserker, SpecializationDefender,
		SpecializationArcanist, SpecializationElementalist,
		SpecializationHealer, SpecializationOracle,
		SpecializationAssassin, SpecializationShadowdancer,
		SpecializationMarksman, SpecializationSpellshot,
		SpecializationDemonologist, SpecializationAffliction,
		SpecializationTemplar, SpecializationCrusader,
		SpecializationBeastmaster, SpecializationPackLeader,
		SpecializationWindwalker, SpecializationBrewmaster,
		SpecializationShinobi, SpecializationStriker,
		SpecializationSpellsword, SpecializationDuelist,
		SpecializationWarmage, SpecializationDeathKnight,
		SpecializationFrost, SpecializationUnholy,
		SpecializationTrickster, SpecializationSeeker,
		SpecializationGuardian, SpecializationExorcist,
		SpecializationTheurgist, SpecializationBloodMage,
		SpecializationVoidcaller, SpecializationNaturemage,
		SpecializationShapeshifter, SpecializationFeralRage,
		SpecializationPurifier, SpecializationJudge,
		SpecializationInterrogator, SpecializationHemomancer,
		SpecializationSoulweaver, SpecializationCrimsonBlade,
	}

	for _, spec := range specializations {
		t.Run(spec.String(), func(t *testing.T) {
			entity := NewEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassWarrior,
				Specialization: spec,
			})
			entity.AddComponent(&StatsComponent{
				CritDamage: 2.0,
			})
			world.AddEntity(entity)

			// Should not panic
			system.Update([]*Entity{entity}, 1.0)

			// Clean up
			system.removeBonus(entity.ID)
			world.RemoveEntity(entity.ID)
		})
	}
}

func BenchmarkSpecializationCritDamageSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	// Create test entities
	entities := make([]*Entity, 100)
	for i := range entities {
		entity := NewEntity()
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassRogue,
			Specialization: SpecializationAssassin,
		})
		entity.AddComponent(&StatsComponent{
			CritDamage: 2.0,
		})
		world.AddEntity(entity)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.timeSinceCheck = system.updateInterval // Force update
		system.Update(entities, 1.0)
	}
}

func BenchmarkSpecializationCritDamageSystem_GetBonus(b *testing.B) {
	world := NewWorld()
	system := NewSpecializationCritDamageSystem(world, 12345)

	entity := NewEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Specialization: SpecializationAssassin,
	})
	entity.AddComponent(&StatsComponent{
		CritDamage: 2.0,
	})
	world.AddEntity(entity)
	system.Update([]*Entity{entity}, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.GetCritDamageBonus(entity.ID)
	}
}
