package engine

import (
	"testing"
)

func TestNewCompanionManaRegenSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewCompanionManaRegenSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
}

func TestCompanionManaRegenSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)
	sys.SetGenre("fantasy")
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", sys.genreID)
	}
}

func TestCompanionManaRegenSystem_NoCompanions(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)

	// Create an entity with mana but no companion
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	sys.Update([]*Entity{owner}, 1.0)

	// Should have no bonus
	if sys.HasActiveBonus(owner.ID) {
		t.Error("expected no bonus without companion")
	}
}

func TestCompanionManaRegenSystem_BasicLoyaltyBonus(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create owner with mana
	owner := world.CreateEntity()
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	// Create companion with high loyalty
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		Loyalty:       100,
		CompanionType: CompanionTypeSpirit, // Highest mana bonus type
	})

	// Update twice to trigger calculation (throttle)
	sys.Update([]*Entity{owner, companion}, 0.5)
	sys.Update([]*Entity{owner, companion}, 0.5)

	// Should have bonus
	if !sys.HasActiveBonus(owner.ID) {
		t.Error("expected bonus with companion")
	}

	// Check mana regen increased
	manaComp, _ := owner.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	// Spirit (1.8) * Fantasy (1.3) * 100 loyalty * 0.002 = 0.468 (46.8% bonus)
	// Original 5.0 * 1.468 = ~7.34
	if mana.Regen <= 5.0 {
		t.Errorf("mana regen should be increased, got %f", mana.Regen)
	}
}

func TestCompanionManaRegenSystem_CompanionTypeBonuses(t *testing.T) {
	tests := []struct {
		compType    CompanionType
		name        string
		expectBonus float64 // relative (higher = more bonus)
	}{
		{CompanionTypeSpirit, "spirit", 1.8},
		{CompanionTypeElemental, "elemental", 1.5},
		{CompanionTypeSummon, "summon", 1.0},
		{CompanionTypeUndead, "undead", 0.8},
		{CompanionTypePet, "pet", 0.6},
		{CompanionTypeInsect, "insect", 0.5},
		{CompanionTypeHireling, "hireling", 0.4},
		{CompanionTypeRobot, "robot", 0.3},
	}

	baseBonus := 0.0
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCompanionManaRegenSystem(world, 12345)
			sys.SetGenre("fantasy")

			owner := world.CreateEntity()
			owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

			companion := world.CreateEntity()
			companion.AddComponent(&CompanionComponent{
				OwnerID:       owner.ID,
				Loyalty:       100,
				CompanionType: tt.compType,
			})

			sys.Update([]*Entity{owner, companion}, 0.5)
			sys.Update([]*Entity{owner, companion}, 0.5)

			bonus := sys.GetBonusMultiplier(owner.ID)

			// First type sets baseline
			if i == 0 {
				baseBonus = bonus
			}

			// Higher multiplier types should give more bonus
			expectedRatio := tt.expectBonus / 1.8 // ratio to spirit
			actualRatio := bonus / baseBonus

			// Allow 10% tolerance
			if actualRatio < expectedRatio*0.8 || actualRatio > expectedRatio*1.2 {
				t.Errorf("%s: bonus ratio = %f, expected around %f", tt.name, actualRatio, expectedRatio)
			}
		})
	}
}

func TestCompanionManaRegenSystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre    string
		wantMult float64
	}{
		{"fantasy", 1.3},
		{"horror", 1.1},
		{"postapoc", 0.7},
		{"scifi", 0.5},
		{"cyberpunk", 0.4},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewCompanionManaRegenSystem(world, 12345)
			sys.SetGenre(tt.genre)

			owner := world.CreateEntity()
			owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

			companion := world.CreateEntity()
			companion.AddComponent(&CompanionComponent{
				OwnerID:       owner.ID,
				Loyalty:       100,
				CompanionType: CompanionTypeSummon, // 1.0 type multiplier
			})

			sys.Update([]*Entity{owner, companion}, 0.5)
			sys.Update([]*Entity{owner, companion}, 0.5)

			// Base: 100 * 0.002 * typeMult(1.0) = 0.2, with genre multiplier
			expectedBonus := 0.2 * tt.wantMult
			bonus := sys.GetBonusMultiplier(owner.ID)

			if bonus < expectedBonus*0.9 || bonus > expectedBonus*1.1 {
				t.Errorf("genre %s: bonus = %f, expected around %f", tt.genre, bonus, expectedBonus)
			}
		})
	}
}

func TestCompanionManaRegenSystem_PerkBonuses(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	owner := world.CreateEntity()
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	// Companion with mana-boosting perks
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		Loyalty:       50,
		CompanionType: CompanionTypeSummon,
		BondingPerks:  []BondingPerk{PerkFasterLearning, PerkSharedExperience},
	})

	sys.Update([]*Entity{owner, companion}, 0.5)
	sys.Update([]*Entity{owner, companion}, 0.5)

	bonus := sys.GetBonusMultiplier(owner.ID)
	// Should have bonus from loyalty + both perks + perk count stacking
	if bonus <= 0 {
		t.Error("expected positive bonus from perks")
	}

	// Compare to companion without perks
	world2 := NewWorld()
	sys2 := NewCompanionManaRegenSystem(world2, 12345)
	sys2.SetGenre("fantasy")

	owner2 := world2.CreateEntity()
	owner2.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	companion2 := world2.CreateEntity()
	companion2.AddComponent(&CompanionComponent{
		OwnerID:       owner2.ID,
		Loyalty:       50,
		CompanionType: CompanionTypeSummon,
	})

	sys2.Update([]*Entity{owner2, companion2}, 0.5)
	sys2.Update([]*Entity{owner2, companion2}, 0.5)

	bonusNoPerk := sys2.GetBonusMultiplier(owner2.ID)

	if bonus <= bonusNoPerk {
		t.Errorf("perk bonus %f should be > no-perk bonus %f", bonus, bonusNoPerk)
	}
}

func TestCompanionManaRegenSystem_MultipleCompanions(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	owner := world.CreateEntity()
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	// Create two companions for same owner
	companion1 := world.CreateEntity()
	companion1.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		Loyalty:       50,
		CompanionType: CompanionTypeSummon,
	})

	companion2 := world.CreateEntity()
	companion2.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		Loyalty:       50,
		CompanionType: CompanionTypeSpirit,
	})

	sys.Update([]*Entity{owner, companion1, companion2}, 0.5)
	sys.Update([]*Entity{owner, companion1, companion2}, 0.5)

	// Bonuses should stack
	bonus := sys.GetBonusMultiplier(owner.ID)

	// Compare to single companion
	world2 := NewWorld()
	sys2 := NewCompanionManaRegenSystem(world2, 12345)
	sys2.SetGenre("fantasy")

	owner2 := world2.CreateEntity()
	owner2.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	companion3 := world2.CreateEntity()
	companion3.AddComponent(&CompanionComponent{
		OwnerID:       owner2.ID,
		Loyalty:       50,
		CompanionType: CompanionTypeSummon,
	})

	sys2.Update([]*Entity{owner2, companion3}, 0.5)
	sys2.Update([]*Entity{owner2, companion3}, 0.5)

	singleBonus := sys2.GetBonusMultiplier(owner2.ID)

	if bonus <= singleBonus {
		t.Errorf("stacked bonus %f should be > single bonus %f", bonus, singleBonus)
	}
}

func TestCompanionManaRegenSystem_ZeroLoyalty(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)

	owner := world.CreateEntity()
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 0,
	})

	sys.Update([]*Entity{owner, companion}, 0.5)
	sys.Update([]*Entity{owner, companion}, 0.5)

	bonus := sys.GetBonusMultiplier(owner.ID)
	// Should be 0 with no loyalty and no perks
	if bonus != 0 {
		t.Errorf("expected 0 bonus at 0 loyalty, got %f", bonus)
	}
}

func TestCompanionManaRegenSystem_NegativeLoyalty(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)

	owner := world.CreateEntity()
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: -50, // Negative loyalty
	})

	sys.Update([]*Entity{owner, companion}, 0.5)
	sys.Update([]*Entity{owner, companion}, 0.5)

	bonus := sys.GetBonusMultiplier(owner.ID)
	// Should be clamped to 0
	if bonus < 0 {
		t.Errorf("negative loyalty should not give negative bonus, got %f", bonus)
	}
}

func TestCompanionManaRegenSystem_RegenRestored(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	owner := world.CreateEntity()
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		Loyalty:       100,
		CompanionType: CompanionTypeSpirit,
	})

	// Update with companion
	sys.Update([]*Entity{owner, companion}, 0.5)
	sys.Update([]*Entity{owner, companion}, 0.5)

	manaComp, _ := owner.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	boostedRegen := mana.Regen

	if boostedRegen <= 5.0 {
		t.Fatal("mana regen should be boosted")
	}

	// Update without companion (companion despawned)
	sys.Update([]*Entity{owner}, 0.5)
	sys.Update([]*Entity{owner}, 0.5)

	// Regen should be restored
	if mana.Regen != 5.0 {
		t.Errorf("mana regen should be restored to 5.0, got %f", mana.Regen)
	}
}

func TestCompanionManaRegenSystem_ThrottledUpdate(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)

	owner := world.CreateEntity()
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 50,
	})

	// First update with small delta - should not calculate yet
	sys.Update([]*Entity{owner, companion}, 0.1)

	// Should not have bonus yet (throttled)
	if sys.HasActiveBonus(owner.ID) {
		t.Error("should not calculate on first small delta update")
	}

	// Continue updating until threshold
	sys.Update([]*Entity{owner, companion}, 0.5)

	// Now should have bonus
	if !sys.HasActiveBonus(owner.ID) {
		t.Error("should have bonus after throttle threshold")
	}
}

func TestCompanionManaRegenSystem_GetOriginalRegen(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	owner := world.CreateEntity()
	owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 7.5})

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 100,
	})

	sys.Update([]*Entity{owner, companion}, 0.5)
	sys.Update([]*Entity{owner, companion}, 0.5)

	original := sys.GetOriginalRegen(owner.ID)
	if original != 7.5 {
		t.Errorf("GetOriginalRegen = %f, want 7.5", original)
	}
}

func TestCompanionManaRegenSystem_GetActiveBonusCount(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)

	owner1 := world.CreateEntity()
	owner1.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	owner2 := world.CreateEntity()
	owner2.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	companion1 := world.CreateEntity()
	companion1.AddComponent(&CompanionComponent{
		OwnerID: owner1.ID,
		Loyalty: 50,
	})

	companion2 := world.CreateEntity()
	companion2.AddComponent(&CompanionComponent{
		OwnerID: owner2.ID,
		Loyalty: 50,
	})

	sys.Update([]*Entity{owner1, owner2, companion1, companion2}, 0.5)
	sys.Update([]*Entity{owner1, owner2, companion1, companion2}, 0.5)

	if count := sys.GetActiveBonusCount(); count != 2 {
		t.Errorf("GetActiveBonusCount = %d, want 2", count)
	}
}

func BenchmarkCompanionManaRegenSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewCompanionManaRegenSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create 100 owner-companion pairs
	entities := make([]*Entity, 200)
	for i := 0; i < 100; i++ {
		owner := world.CreateEntity()
		owner.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
		entities[i*2] = owner

		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			OwnerID:       owner.ID,
			Loyalty:       float64(50 + i%50),
			CompanionType: CompanionType(i % 8),
			BondingPerks:  []BondingPerk{PerkFasterLearning},
		})
		entities[i*2+1] = companion
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.5)
	}
}
