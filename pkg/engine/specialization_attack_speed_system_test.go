package engine

import (
	"testing"
)

func TestNewSpecializationAttackSpeedSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewSpecializationAttackSpeedSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.updateInterval != 1.0 {
		t.Errorf("Expected updateInterval 1.0, got %v", sys.updateInterval)
	}
	if sys.appliedBonuses == nil {
		t.Error("appliedBonuses map not initialized")
	}
}

func TestSpecializationAttackSpeedSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("Expected genreID 'cyberpunk', got '%s'", sys.genreID)
	}
}

func TestSpecializationAttackSpeedSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	// Create entity with class progression and attack
	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          10,
		Specialization: SpecializationBerserker,
	})
	entity.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity)

	// First update should not process (interval not met)
	sys.Update([]*Entity{entity}, 0.5)

	if _, exists := sys.appliedBonuses[entity.ID]; exists {
		t.Error("Should not have applied bonus before interval")
	}

	// Second update should process (interval met)
	sys.Update([]*Entity{entity}, 0.6)

	if _, exists := sys.appliedBonuses[entity.ID]; !exists {
		t.Error("Should have applied bonus after interval")
	}
}

func TestSpecializationAttackSpeedSystem_BerserkerBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)
	sys.SetGenre("fantasy")

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          10,
		Specialization: SpecializationBerserker,
	})
	entity.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity)

	// Force update
	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{entity}, 0.1)

	// Berserker should have 30% cooldown reduction
	bonus := sys.GetBonusForEntity(entity.ID)
	if bonus < 0.29 || bonus > 0.31 {
		t.Errorf("Expected ~30%% bonus for Berserker, got %v", bonus*100)
	}

	// Verify cooldown was reduced
	attackComp, _ := entity.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	expectedCooldown := 1.0 * (1.0 - 0.30)
	if attack.Cooldown < expectedCooldown-0.01 || attack.Cooldown > expectedCooldown+0.01 {
		t.Errorf("Expected cooldown ~%v, got %v", expectedCooldown, attack.Cooldown)
	}
}

func TestSpecializationAttackSpeedSystem_AssassinBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Level:          10,
		Specialization: SpecializationAssassin,
	})
	entity.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{entity}, 0.1)

	// Assassin should have 28% cooldown reduction
	bonus := sys.GetBonusForEntity(entity.ID)
	if bonus < 0.27 || bonus > 0.29 {
		t.Errorf("Expected ~28%% bonus for Assassin, got %v", bonus*100)
	}
}

func TestSpecializationAttackSpeedSystem_CasterNoBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassMage,
		Level:          10,
		Specialization: SpecializationArcanist,
	})
	entity.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{entity}, 0.1)

	// Arcanist (pure caster) should have 0% attack speed bonus
	bonus := sys.GetBonusForEntity(entity.ID)
	if bonus != 0.0 {
		t.Errorf("Expected 0%% bonus for Arcanist, got %v", bonus*100)
	}

	// Cooldown should be unchanged
	attackComp, _ := entity.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if attack.Cooldown != 1.0 {
		t.Errorf("Expected cooldown 1.0 unchanged, got %v", attack.Cooldown)
	}
}

func TestSpecializationAttackSpeedSystem_NoSpecialization(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          5, // Below spec level
		Specialization: SpecializationNone,
	})
	entity.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{entity}, 0.1)

	// Base warrior class should have small 5% bonus
	bonus := sys.GetBonusForEntity(entity.ID)
	if bonus < 0.04 || bonus > 0.06 {
		t.Errorf("Expected ~5%% base bonus for Warrior, got %v", bonus*100)
	}
}

func TestSpecializationAttackSpeedSystem_GenreModifier(t *testing.T) {
	tests := []struct {
		name           string
		genre          string
		expectedMin    float64
		expectedMax    float64
		baseBonus      float64
		specialization SpecializationType
	}{
		{"cyberpunk_berserker", "cyberpunk", 0.33, 0.36, 0.30, SpecializationBerserker}, // 30% * 1.15
		{"horror_berserker", "horror", 0.26, 0.28, 0.30, SpecializationBerserker},       // 30% * 0.9
		{"scifi_assassin", "scifi", 0.30, 0.32, 0.28, SpecializationAssassin},           // 28% * 1.1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationAttackSpeedSystem(world, 12345)
			sys.SetGenre(tt.genre)

			entity := NewEntity(1)
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassWarrior,
				Level:          10,
				Specialization: tt.specialization,
			})
			entity.AddComponent(&AttackComponent{
				Cooldown:      1.0,
				CooldownTimer: 0.0,
			})
			world.AddEntity(entity)

			sys.timeSinceCheck = 1.0
			sys.Update([]*Entity{entity}, 0.1)

			bonus := sys.GetBonusForEntity(entity.ID)
			if bonus < tt.expectedMin || bonus > tt.expectedMax {
				t.Errorf("Expected bonus between %.2f and %.2f for %s, got %.4f",
					tt.expectedMin, tt.expectedMax, tt.name, bonus)
			}
		})
	}
}

func TestSpecializationAttackSpeedSystem_MissingComponents(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	// Entity with only class progression (no attack)
	entity1 := NewEntity(1)
	entity1.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          10,
		Specialization: SpecializationBerserker,
	})
	world.AddEntity(entity1)

	// Entity with only attack (no class progression)
	entity2 := NewEntity(2)
	entity2.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity2)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{entity1, entity2}, 0.1)

	// Neither should have bonuses
	if sys.IsSpecializationActive(entity1.ID) {
		t.Error("Entity without attack component should not have bonus")
	}
	if sys.IsSpecializationActive(entity2.ID) {
		t.Error("Entity without class progression should not have bonus")
	}
}

func TestSpecializationAttackSpeedSystem_GetCooldownMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassRogue,
		Level:          10,
		Specialization: SpecializationAssassin, // 28% bonus
	})
	entity.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity)

	// Before update
	mult := sys.GetCooldownMultiplier(entity.ID)
	if mult != 1.0 {
		t.Errorf("Expected multiplier 1.0 before update, got %v", mult)
	}

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{entity}, 0.1)

	// After update (28% bonus = 0.72 multiplier)
	mult = sys.GetCooldownMultiplier(entity.ID)
	if mult < 0.71 || mult > 0.73 {
		t.Errorf("Expected multiplier ~0.72 after update, got %v", mult)
	}
}

func TestSpecializationAttackSpeedSystem_RemoveBonus(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          10,
		Specialization: SpecializationBerserker,
	})
	entity.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity)

	// Apply bonus
	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{entity}, 0.1)

	if !sys.IsSpecializationActive(entity.ID) {
		t.Fatal("Bonus should be active")
	}

	// Remove class progression component
	entity.RemoveComponent("class_progression")

	// Update again
	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{entity}, 0.1)

	if sys.IsSpecializationActive(entity.ID) {
		t.Error("Bonus should be removed after component removed")
	}
}

func TestSpecializationAttackSpeedSystem_HybridSpecializations(t *testing.T) {
	tests := []struct {
		name        string
		class       CharacterClass
		spec        SpecializationType
		expectedMin float64
		expectedMax float64
	}{
		{"spellsword", ClassBattlemage, SpecializationSpellsword, 0.17, 0.19},       // 18%
		{"duelist", ClassSpellblade, SpecializationDuelist, 0.25, 0.27},             // 26%
		{"crusader", ClassPaladin, SpecializationCrusader, 0.15, 0.17},              // 16%
		{"windwalker", ClassMonk, SpecializationWindwalker, 0.29, 0.31},             // 30%
		{"shapeshifter", ClassDruid, SpecializationShapeshifter, 0.17, 0.19},        // 18%
		{"crimson_blade", ClassBloodKnight, SpecializationCrimsonBlade, 0.23, 0.25}, // 24%
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewSpecializationAttackSpeedSystem(world, 12345)

			entity := NewEntity(1)
			entity.AddComponent(&ClassProgressionComponent{
				Class:          tt.class,
				Level:          10,
				Specialization: tt.spec,
			})
			entity.AddComponent(&AttackComponent{
				Cooldown:      1.0,
				CooldownTimer: 0.0,
			})
			world.AddEntity(entity)

			sys.timeSinceCheck = 1.0
			sys.Update([]*Entity{entity}, 0.1)

			bonus := sys.GetBonusForEntity(entity.ID)
			if bonus < tt.expectedMin || bonus > tt.expectedMax {
				t.Errorf("Expected bonus between %.2f and %.2f for %s, got %.4f",
					tt.expectedMin, tt.expectedMax, tt.name, bonus)
			}
		})
	}
}

func TestSpecializationAttackSpeedSystem_BonusDoesNotDouble(t *testing.T) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          10,
		Specialization: SpecializationBerserker,
	})
	entity.AddComponent(&AttackComponent{
		Cooldown:      1.0,
		CooldownTimer: 0.0,
	})
	world.AddEntity(entity)

	// Apply bonus multiple times
	for i := 0; i < 5; i++ {
		sys.timeSinceCheck = 1.0
		sys.Update([]*Entity{entity}, 0.1)
	}

	// Cooldown should only be reduced once
	attackComp, _ := entity.GetComponent("attack")
	attack := attackComp.(*AttackComponent)

	// 30% reduction of 1.0 = 0.7
	if attack.Cooldown < 0.69 || attack.Cooldown > 0.71 {
		t.Errorf("Expected cooldown ~0.7 (single reduction), got %v", attack.Cooldown)
	}
}

func BenchmarkSpecializationAttackSpeedSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewSpecializationAttackSpeedSystem(world, 12345)

	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := NewEntity(1)
		entity.AddComponent(&ClassProgressionComponent{
			Class:          ClassWarrior,
			Level:          10,
			Specialization: SpecializationBerserker,
		})
		entity.AddComponent(&AttackComponent{
			Cooldown:      1.0,
			CooldownTimer: 0.0,
		})
		world.AddEntity(entity)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 1.0
		sys.Update(entities, 0.016)
	}
}
