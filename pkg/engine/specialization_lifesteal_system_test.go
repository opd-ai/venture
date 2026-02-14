package engine

import (
	"testing"
)

func TestNewSpecializationLifestealSystem(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)

	if system == nil {
		t.Fatal("NewSpecializationLifestealSystem returned nil")
	}
	if system.world != world {
		t.Error("System world reference not set")
	}
	if system.updateInterval != 1.0 {
		t.Errorf("Update interval = %v, want 1.0", system.updateInterval)
	}
	if system.appliedBonuses == nil {
		t.Error("Applied bonuses map not initialized")
	}
}

func TestSpecializationLifestealSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)

	system.SetGenre("horror")
	if system.genreID != "horror" {
		t.Errorf("Genre = %v, want horror", system.genreID)
	}
}

func TestSpecializationLifestealSystem_BloodMageBonus(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create entity with blood mage specialization
	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassNecromancer,
		Level:          15,
		Specialization: SpecializationBloodMage,
	})
	stats := NewStatsComponent()
	stats.Lifesteal = 0.0 // Start with no lifesteal
	entity.AddComponent(stats)

	entities := []*Entity{entity}
	system.Update(entities, 1.0) // Trigger update

	// Blood mage should get +30% lifesteal (0.30)
	expectedBonus := 0.30 * 1.0 // fantasy genre multiplier = 1.0
	if stats.Lifesteal < expectedBonus-0.01 || stats.Lifesteal > expectedBonus+0.01 {
		t.Errorf("Blood mage lifesteal = %v, want ~%v", stats.Lifesteal, expectedBonus)
	}
}

func TestSpecializationLifestealSystem_HorrorGenreMultiplier(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)
	system.SetGenre("horror")

	// Create entity with blood mage specialization
	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassNecromancer,
		Level:          15,
		Specialization: SpecializationBloodMage,
	})
	stats := NewStatsComponent()
	stats.Lifesteal = 0.0
	entity.AddComponent(stats)

	entities := []*Entity{entity}
	system.Update(entities, 1.0)

	// Blood mage in horror: 0.30 * 1.3 = 0.39
	expectedBonus := 0.30 * 1.3
	if stats.Lifesteal < expectedBonus-0.01 || stats.Lifesteal > expectedBonus+0.01 {
		t.Errorf("Horror blood mage lifesteal = %v, want ~%v", stats.Lifesteal, expectedBonus)
	}
}

func TestSpecializationLifestealSystem_NoSpecialization(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create entity with no specialization (just necromancer class)
	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassNecromancer,
		Level:          5,
		Specialization: SpecializationNone,
	})
	stats := NewStatsComponent()
	stats.Lifesteal = 0.0
	entity.AddComponent(stats)

	entities := []*Entity{entity}
	system.Update(entities, 1.0)

	// Necromancer base class should get +5% lifesteal (0.05)
	expectedBonus := 0.05
	if stats.Lifesteal < expectedBonus-0.01 || stats.Lifesteal > expectedBonus+0.01 {
		t.Errorf("Base necromancer lifesteal = %v, want ~%v", stats.Lifesteal, expectedBonus)
	}
}

func TestSpecializationLifestealSystem_HealerNoBonus(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create entity with healer specialization (should get no lifesteal)
	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassCleric,
		Level:          15,
		Specialization: SpecializationHealer,
	})
	stats := NewStatsComponent()
	stats.Lifesteal = 0.0
	entity.AddComponent(stats)

	entities := []*Entity{entity}
	system.Update(entities, 1.0)

	// Healer should get 0% lifesteal
	if stats.Lifesteal != 0.0 {
		t.Errorf("Healer lifesteal = %v, want 0.0", stats.Lifesteal)
	}
}

func TestSpecializationLifestealSystem_BonusRemovalOnChange(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	progression := &ClassProgressionComponent{
		Class:          ClassNecromancer,
		Level:          15,
		Specialization: SpecializationBloodMage,
	}
	entity.AddComponent(progression)
	stats := NewStatsComponent()
	stats.Lifesteal = 0.0
	entity.AddComponent(stats)

	entities := []*Entity{entity}
	system.Update(entities, 1.0)

	initialBonus := stats.Lifesteal

	// Change specialization to healer
	progression.Specialization = SpecializationHealer
	system.Update(entities, 1.0)

	// Should have removed blood mage bonus and not added healer bonus
	if stats.Lifesteal >= initialBonus {
		t.Errorf("Lifesteal not reduced after spec change: %v, was %v", stats.Lifesteal, initialBonus)
	}
	if stats.Lifesteal != 0.0 {
		t.Errorf("Healer lifesteal = %v, want 0.0", stats.Lifesteal)
	}
}

func TestSpecializationLifestealSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassNecromancer,
		Level:          15,
		Specialization: SpecializationBloodMage,
	})
	stats := NewStatsComponent()
	stats.Lifesteal = 0.0
	entity.AddComponent(stats)

	entities := []*Entity{entity}

	// First update with deltaTime < updateInterval should not apply bonus
	system.Update(entities, 0.5)
	if stats.Lifesteal != 0.0 {
		t.Error("Bonus applied before update interval elapsed")
	}

	// Second update pushes us past interval
	system.Update(entities, 0.6)
	if stats.Lifesteal == 0.0 {
		t.Error("Bonus not applied after update interval elapsed")
	}
}

func TestSpecializationLifestealSystem_MissingComponents(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)

	tests := []struct {
		name       string
		components []Component
	}{
		{"no components", nil},
		{"stats only", []Component{NewStatsComponent()}},
		{"class only", []Component{&ClassProgressionComponent{
			Class:          ClassNecromancer,
			Specialization: SpecializationBloodMage,
		}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			for _, comp := range tt.components {
				entity.AddComponent(comp)
			}

			entities := []*Entity{entity}
			// Should not panic with missing components
			system.Update(entities, 1.0)
		})
	}
}

func TestSpecializationLifestealSystem_GetLifestealBonus(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassNecromancer,
		Level:          15,
		Specialization: SpecializationBloodMage,
	})
	entity.AddComponent(NewStatsComponent())

	entities := []*Entity{entity}
	system.Update(entities, 1.0)

	bonus := system.GetLifestealBonus(entity.ID)
	if bonus < 0.29 || bonus > 0.31 {
		t.Errorf("GetLifestealBonus = %v, want ~0.30", bonus)
	}

	// Non-existent entity should return 0
	bonus = system.GetLifestealBonus(99999)
	if bonus != 0.0 {
		t.Errorf("Non-existent entity bonus = %v, want 0.0", bonus)
	}
}

func TestSpecializationLifestealSystem_IsSpecializationActive(t *testing.T) {
	world := NewWorld()
	system := NewSpecializationLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassNecromancer,
		Level:          15,
		Specialization: SpecializationBloodMage,
	})
	entity.AddComponent(NewStatsComponent())

	// Not active before update
	if system.IsSpecializationActive(entity.ID) {
		t.Error("Specialization active before update")
	}

	entities := []*Entity{entity}
	system.Update(entities, 1.0)

	// Active after update
	if !system.IsSpecializationActive(entity.ID) {
		t.Error("Specialization not active after update")
	}
}

func TestSpecializationLifestealSystem_AllGenres(t *testing.T) {
	genres := []struct {
		id         string
		multiplier float64
	}{
		{"fantasy", 1.0},
		{"scifi", 0.8},
		{"horror", 1.3},
		{"cyberpunk", 0.9},
		{"postapoc", 1.1},
	}

	for _, genre := range genres {
		t.Run(genre.id, func(t *testing.T) {
			world := NewWorld()
			system := NewSpecializationLifestealSystem(world, 12345)
			system.SetGenre(genre.id)

			entity := world.CreateEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassNecromancer,
				Level:          15,
				Specialization: SpecializationBloodMage,
			})
			stats := NewStatsComponent()
			entity.AddComponent(stats)

			entities := []*Entity{entity}
			system.Update(entities, 1.0)

			expectedBonus := 0.30 * genre.multiplier
			if stats.Lifesteal < expectedBonus-0.01 || stats.Lifesteal > expectedBonus+0.01 {
				t.Errorf("%s lifesteal = %v, want ~%v", genre.id, stats.Lifesteal, expectedBonus)
			}
		})
	}
}

func TestSpecializationLifestealSystem_AllSpecializations(t *testing.T) {
	// Test a sampling of specializations to ensure they return valid values
	specs := []struct {
		spec     SpecializationType
		minBonus float64
		maxBonus float64
	}{
		{SpecializationBloodMage, 0.25, 0.35},
		{SpecializationCrimsonBlade, 0.24, 0.32},
		{SpecializationHemomancer, 0.20, 0.30},
		{SpecializationBerserker, 0.10, 0.20},
		{SpecializationDefender, 0.03, 0.08},
		{SpecializationHealer, 0.0, 0.0},
		{SpecializationArcanist, 0.0, 0.05},
	}

	for _, tt := range specs {
		t.Run(tt.spec.String(), func(t *testing.T) {
			world := NewWorld()
			system := NewSpecializationLifestealSystem(world, 12345)
			system.SetGenre("fantasy")

			entity := world.CreateEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassWarrior,
				Level:          15,
				Specialization: tt.spec,
			})
			stats := NewStatsComponent()
			entity.AddComponent(stats)

			entities := []*Entity{entity}
			system.Update(entities, 1.0)

			if stats.Lifesteal < tt.minBonus || stats.Lifesteal > tt.maxBonus {
				t.Errorf("%s lifesteal = %v, want %v-%v", tt.spec, stats.Lifesteal, tt.minBonus, tt.maxBonus)
			}
		})
	}
}
