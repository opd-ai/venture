package engine

import (
	"testing"
)

func TestNewSpecializationManaBoostSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewSpecializationManaBoostSystem returned nil")
	}
	if sys.world != world {
		t.Error("World reference not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.originalRegen == nil {
		t.Error("originalRegen map not initialized")
	}
	if sys.appliedBonuses == nil {
		t.Error("appliedBonuses map not initialized")
	}
}

func TestSpecializationManaBoostSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("Genre not set correctly: got %s, want %s", sys.genreID, genre)
		}
	}
}

func TestSpecializationManaBoostSystem_Update_NoSpecialization(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	// Create entity with mana and class progression but no specialization
	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          5,
		Specialization: SpecializationNone,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Mage without specialization should get 10% bonus
	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Expected: 5.0 * 1.10 = 5.5
	if mana.Regen < 5.4 || mana.Regen > 5.6 {
		t.Errorf("Mage base class bonus incorrect: got %f, want ~5.5", mana.Regen)
	}
}

func TestSpecializationManaBoostSystem_Update_MageSpecialization(t *testing.T) {
	tests := []struct {
		name           string
		specialization SpecializationType
		minBonus       float64 // Expected minimum regen value
		maxBonus       float64 // Expected maximum regen value
	}{
		{"Arcanist", SpecializationArcanist, 7.4, 7.6},         // 5.0 * 1.50 = 7.5
		{"Elementalist", SpecializationElementalist, 6.9, 7.1}, // 5.0 * 1.40 = 7.0
		{"BloodMage", SpecializationBloodMage, 6.6, 6.8},       // 5.0 * 1.35 = 6.75
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationManaBoostSystem(world, 12345)

			entity := world.CreateEntity()
			entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassMage,
				Level:          10,
				Specialization: tt.specialization,
			})

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			manaComp, _ := entity.GetComponent("mana")
			mana := manaComp.(*ManaComponent)

			if mana.Regen < tt.minBonus || mana.Regen > tt.maxBonus {
				t.Errorf("%s bonus incorrect: got %f, want [%f, %f]",
					tt.name, mana.Regen, tt.minBonus, tt.maxBonus)
			}
		})
	}
}

func TestSpecializationManaBoostSystem_Update_WarriorNoBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          10,
		Specialization: SpecializationBerserker,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Warriors should get no bonus
	if mana.Regen != 5.0 {
		t.Errorf("Warrior should get no bonus: got %f, want 5.0", mana.Regen)
	}
}

func TestSpecializationManaBoostSystem_GenreModifier(t *testing.T) {
	tests := []struct {
		genre    string
		expected float64 // Multiplier
	}{
		{"fantasy", 1.0},
		{"scifi", 0.8},
		{"horror", 1.2},
		{"cyberpunk", 0.7},
		{"postapoc", 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationManaBoostSystem(world, 12345)
			sys.SetGenre(tt.genre)

			entity := world.CreateEntity()
			entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassMage,
				Level:          10,
				Specialization: SpecializationArcanist, // 50% base bonus
			})

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			manaComp, _ := entity.GetComponent("mana")
			mana := manaComp.(*ManaComponent)

			// Expected: 10.0 * (1.0 + 0.5 * genreMultiplier)
			expectedBase := 10.0 * (1.0 + 0.5*tt.expected)
			tolerance := 0.5

			if mana.Regen < expectedBase-tolerance || mana.Regen > expectedBase+tolerance {
				t.Errorf("Genre %s bonus incorrect: got %f, want ~%f", tt.genre, mana.Regen, expectedBase)
			}
		})
	}
}

func TestSpecializationManaBoostSystem_Update_NoManaComponent(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          10,
		Specialization: SpecializationArcanist,
	})
	// No ManaComponent added

	entities := []*Entity{entity}
	// Should not panic
	sys.Update(entities, 1.0)
}

func TestSpecializationManaBoostSystem_Update_NoClassProgression(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	// No ClassProgressionComponent added

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)

	// Should remain unchanged
	if mana.Regen != 5.0 {
		t.Errorf("Regen should be unchanged: got %f, want 5.0", mana.Regen)
	}
}

func TestSpecializationManaBoostSystem_GetBonusForEntity(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          10,
		Specialization: SpecializationArcanist,
	})

	// Before update, no bonus
	bonus := sys.GetBonusForEntity(entity.ID)
	if bonus != 0.0 {
		t.Errorf("Bonus before update should be 0: got %f", bonus)
	}

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// After update, should have bonus
	bonus = sys.GetBonusForEntity(entity.ID)
	if bonus < 0.4 || bonus > 0.6 {
		t.Errorf("Arcanist bonus should be ~0.5: got %f", bonus)
	}
}

func TestSpecializationManaBoostSystem_HybridSpecializations(t *testing.T) {
	tests := []struct {
		name     string
		spec     SpecializationType
		hasBonus bool
	}{
		{"Warmage", SpecializationWarmage, true},
		{"Spellsword", SpecializationSpellsword, true},
		{"Trickster", SpecializationTrickster, true},
		{"Duelist", SpecializationDuelist, true},
		{"Templator", SpecializationTemplar, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationManaBoostSystem(world, 12345)

			entity := world.CreateEntity()
			entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassBattlemage, // Hybrid class
				Level:          10,
				Specialization: tt.spec,
			})

			entities := []*Entity{entity}
			sys.Update(entities, 1.0)

			bonus := sys.GetBonusForEntity(entity.ID)
			if tt.hasBonus && bonus <= 0 {
				t.Errorf("%s should have a mana bonus", tt.name)
			}
		})
	}
}

func TestSpecializationManaBoostSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          10,
		Specialization: SpecializationArcanist,
	})

	entities := []*Entity{entity}

	// First update with small deltaTime should not trigger check
	sys.Update(entities, 0.1)
	bonus := sys.GetBonusForEntity(entity.ID)
	if bonus != 0.0 {
		t.Errorf("Bonus should be 0 before interval: got %f", bonus)
	}

	// Update with remaining time to pass interval
	sys.Update(entities, 0.9)
	bonus = sys.GetBonusForEntity(entity.ID)
	if bonus == 0.0 {
		t.Error("Bonus should be applied after interval")
	}
}

func TestSpecializationManaBoostSystem_RemoveBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          10,
		Specialization: SpecializationArcanist,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Verify bonus applied
	bonus := sys.GetBonusForEntity(entity.ID)
	if bonus == 0.0 {
		t.Fatal("Bonus should be applied")
	}

	// Remove class progression component
	entity.RemoveComponent("class_progression")

	// Update should remove bonus tracking
	sys.Update(entities, 1.0)
	bonus = sys.GetBonusForEntity(entity.ID)
	if bonus != 0.0 {
		t.Errorf("Bonus should be removed: got %f", bonus)
	}
}

func BenchmarkSpecializationManaBoostSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	// Create 100 entities with mana and class progression
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassMage,
			Level:          10,
			Specialization: SpecializationArcanist,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016) // 60 FPS frame
	}
}

func BenchmarkSpecializationManaBoostSystem_Update_MixedEntities(b *testing.B) {
	world := NewWorld()
	sys := NewSpecializationManaBoostSystem(world, 12345)

	// Create 200 entities, half with mana/class, half without
	entities := make([]*Entity, 200)
	for i := 0; i < 200; i++ {
		entity := world.CreateEntity()
		if i%2 == 0 {
			entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
			entity.AddComponent(&ClassProgressionComponent{
				Class:          CharacterClass(i % 6),
				Level:          10,
				Specialization: SpecializationType(i % 12),
			})
		}
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
