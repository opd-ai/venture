package engine

import (
	"testing"
)

func TestNewSpecializationSpellDamageSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewSpecializationSpellDamageSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.updateInterval != 1.0 {
		t.Errorf("updateInterval = %f, want 1.0", sys.updateInterval)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", sys.genreID)
	}
}

func TestSpecializationSpellDamageSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genreID = %s, want horror", sys.genreID)
	}
}

func TestSpecializationSpellDamageSystem_Update_NoSpecialization(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationNone,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0) // Trigger update

	// Warrior with no specialization gets 0 bonus
	multiplier := sys.GetDamageMultiplier(entity.ID)
	if multiplier != 1.0 {
		t.Errorf("GetDamageMultiplier = %f, want 1.0", multiplier)
	}
}

func TestSpecializationSpellDamageSystem_Update_MageSpecialization(t *testing.T) {
	tests := []struct {
		name           string
		specialization SpecializationType
		minMultiplier  float64
		maxMultiplier  float64
	}{
		{"Arcanist", SpecializationArcanist, 1.38, 1.42},         // 40% base
		{"Elementalist", SpecializationElementalist, 1.33, 1.37}, // 35% base
		{"BloodMage", SpecializationBloodMage, 1.28, 1.32},       // 30% base
		{"Voidcaller", SpecializationVoidcaller, 1.31, 1.35},     // 33% base
		{"Demonologist", SpecializationDemonologist, 1.33, 1.37}, // 35% base
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationSpellDamageSystem(world, 12345)
			sys.SetGenre("fantasy")

			entity := world.CreateEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassMage,
				Level:          15,
				Specialization: tt.specialization,
			})
			world.AddEntity(entity)

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			multiplier := sys.GetDamageMultiplier(entity.ID)
			if multiplier < tt.minMultiplier || multiplier > tt.maxMultiplier {
				t.Errorf("GetDamageMultiplier = %f, want between %f and %f",
					multiplier, tt.minMultiplier, tt.maxMultiplier)
			}
		})
	}
}

func TestSpecializationSpellDamageSystem_Update_WarriorNoBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          15,
		Specialization: SpecializationBerserker,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Berserker gets no spell damage bonus
	multiplier := sys.GetDamageMultiplier(entity.ID)
	if multiplier != 1.0 {
		t.Errorf("GetDamageMultiplier = %f, want 1.0 (no bonus for Berserker)", multiplier)
	}
}

func TestSpecializationSpellDamageSystem_GenreModifier(t *testing.T) {
	tests := []struct {
		genre         string
		minMultiplier float64
		maxMultiplier float64
	}{
		{"fantasy", 1.38, 1.42},   // 1.0 genre modifier
		{"horror", 1.44, 1.48},    // 1.15 genre modifier
		{"scifi", 1.32, 1.36},     // 0.85 genre modifier
		{"cyberpunk", 1.28, 1.32}, // 0.75 genre modifier
		{"postapoc", 1.36, 1.40},  // 0.95 genre modifier
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationSpellDamageSystem(world, 12345)
			sys.SetGenre(tt.genre)

			entity := world.CreateEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassMage,
				Level:          15,
				Specialization: SpecializationArcanist, // 40% base bonus
			})
			world.AddEntity(entity)

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			multiplier := sys.GetDamageMultiplier(entity.ID)
			if multiplier < tt.minMultiplier || multiplier > tt.maxMultiplier {
				t.Errorf("GetDamageMultiplier = %f, want between %f and %f for genre %s",
					multiplier, tt.minMultiplier, tt.maxMultiplier, tt.genre)
			}
		})
	}
}

func TestSpecializationSpellDamageSystem_Update_NoClassProgression(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	// Entity without ClassProgressionComponent
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Should return 1.0 (no bonus)
	multiplier := sys.GetDamageMultiplier(entity.ID)
	if multiplier != 1.0 {
		t.Errorf("GetDamageMultiplier = %f, want 1.0", multiplier)
	}
}

func TestSpecializationSpellDamageSystem_GetBonusForEntity(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          15,
		Specialization: SpecializationArcanist,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	bonus := sys.GetBonusForEntity(entity.ID)
	// Arcanist should have ~40% bonus in fantasy genre
	if bonus < 0.38 || bonus > 0.42 {
		t.Errorf("GetBonusForEntity = %f, want between 0.38 and 0.42", bonus)
	}
}

func TestSpecializationSpellDamageSystem_HybridSpecializations(t *testing.T) {
	tests := []struct {
		name     string
		spec     SpecializationType
		hasBonus bool
	}{
		{"Warmage", SpecializationWarmage, true},
		{"Spellsword", SpecializationSpellsword, true},
		{"Templar", SpecializationTemplar, true},
		{"Spellshot", SpecializationSpellshot, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationSpellDamageSystem(world, 12345)

			entity := world.CreateEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassBattlemage,
				Level:          15,
				Specialization: tt.spec,
			})
			world.AddEntity(entity)

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			bonus := sys.GetBonusForEntity(entity.ID)
			if tt.hasBonus && bonus <= 0 {
				t.Errorf("Expected bonus > 0 for %s, got %f", tt.name, bonus)
			}
		})
	}
}

func TestSpecializationSpellDamageSystem_IsSpecializationActive(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          15,
		Specialization: SpecializationArcanist,
	})
	world.AddEntity(entity)

	// Before update, not active
	if sys.IsSpecializationActive(entity.ID) {
		t.Error("Expected IsSpecializationActive to be false before update")
	}

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// After update, should be active
	if !sys.IsSpecializationActive(entity.ID) {
		t.Error("Expected IsSpecializationActive to be true after update")
	}
}

func TestSpecializationSpellDamageSystem_RemoveBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          15,
		Specialization: SpecializationArcanist,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Should have bonus
	if sys.GetDamageMultiplier(entity.ID) <= 1.0 {
		t.Error("Expected multiplier > 1.0 after update")
	}

	// Remove the class progression component
	entity.RemoveComponent("class_progression")
	sys.Update(entities, 1.0)

	// Bonus should be removed
	if sys.GetDamageMultiplier(entity.ID) != 1.0 {
		t.Errorf("Expected multiplier = 1.0 after component removal, got %f",
			sys.GetDamageMultiplier(entity.ID))
	}
}

func TestSpecializationSpellDamageSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          15,
		Specialization: SpecializationArcanist,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}

	// Update with less than interval should not process
	sys.Update(entities, 0.5)
	if sys.IsSpecializationActive(entity.ID) {
		t.Error("Should not process before interval reached")
	}

	// Update past interval should process
	sys.Update(entities, 0.6)
	if !sys.IsSpecializationActive(entity.ID) {
		t.Error("Should process after interval reached")
	}
}

func TestSpecializationSpellDamageSystem_BaseClassBonus(t *testing.T) {
	tests := []struct {
		class        CharacterClass
		hasBaseBonus bool
	}{
		{ClassMage, true},
		{ClassNecromancer, true},
		{ClassCleric, true},
		{ClassWarrior, false},
		{ClassRogue, false},
		{ClassRanger, false},
	}

	for _, tt := range tests {
		t.Run(tt.class.String(), func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationSpellDamageSystem(world, 12345)

			entity := world.CreateEntity()
			entity.AddComponent(&ClassProgressionComponent{
				Class:          tt.class,
				Level:          5, // Not high enough for specialization
				Specialization: SpecializationNone,
			})
			world.AddEntity(entity)

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			multiplier := sys.GetDamageMultiplier(entity.ID)
			if tt.hasBaseBonus && multiplier <= 1.0 {
				t.Errorf("Expected base bonus for %s, got multiplier %f", tt.class.String(), multiplier)
			}
			if !tt.hasBaseBonus && multiplier != 1.0 {
				t.Errorf("Expected no base bonus for %s, got multiplier %f", tt.class.String(), multiplier)
			}
		})
	}
}

func BenchmarkSpecializationSpellDamageSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	// Create 100 entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassMage,
			Level:          15,
			Specialization: SpecializationArcanist,
		})
		world.AddEntity(entity)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 0 // Force update
		sys.Update(entities, 1.0)
	}
}

func BenchmarkSpecializationSpellDamageSystem_GetDamageMultiplier(b *testing.B) {
	world := NewWorld()
	sys := NewSpecializationSpellDamageSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          15,
		Specialization: SpecializationArcanist,
	})
	world.AddEntity(entity)

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.GetDamageMultiplier(entity.ID)
	}
}
