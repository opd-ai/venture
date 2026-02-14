package engine

import (
	"testing"
)

func TestNewCompanionSpellAmplificationSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)
	if sys == nil {
		t.Fatal("NewCompanionSpellAmplificationSystem returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("seed = %d, want 12345", sys.seed)
	}
}

func TestCompanionSpellAmplificationSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)
	sys.SetGenre("fantasy")
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", sys.genreID)
	}
}

func TestCompanionSpellAmplificationSystem_NoCompanions(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)

	// Create an entity without companion component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.Update([]*Entity{entity}, 1.0)

	// Should return 1.0 (no bonus) for any owner
	if mult := sys.GetDamageMultiplier(entity.ID); mult != 1.0 {
		t.Errorf("GetDamageMultiplier = %f, want 1.0", mult)
	}
}

func TestCompanionSpellAmplificationSystem_BasicLoyaltyBonus(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion with high loyalty
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 100, // Max loyalty
	})

	// Update twice to trigger calculation (throttle)
	sys.Update([]*Entity{companion}, 0.5)
	sys.Update([]*Entity{companion}, 0.5)

	// Fantasy genre multiplier is 1.2
	// Base damage bonus: 100 * 0.001 * 1.2 = 0.12 (12%)
	expectedDamage := 1.12
	if mult := sys.GetDamageMultiplier(owner.ID); mult < 1.1 || mult > 1.15 {
		t.Errorf("GetDamageMultiplier = %f, expected around %f", mult, expectedDamage)
	}
}

func TestCompanionSpellAmplificationSystem_PerkBonus(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)
	sys.SetGenre("fantasy")

	owner := world.CreateEntity()

	// Create companion with perks
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:      owner.ID,
		Loyalty:      50,
		BondingPerks: []BondingPerk{PerkExtraDamage, PerkExtraHealth},
	})

	sys.Update([]*Entity{companion}, 0.5)
	sys.Update([]*Entity{companion}, 0.5)

	// Should have both damage and healing bonuses
	dmgBonus, healBonus := sys.GetBonusForOwner(owner.ID)
	if dmgBonus <= 0 {
		t.Error("expected positive damage bonus from perk")
	}
	if healBonus <= 0 {
		t.Error("expected positive healing bonus from perk")
	}
}

func TestCompanionSpellAmplificationSystem_MultipleCompanions(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)
	sys.SetGenre("fantasy")

	owner := world.CreateEntity()

	// Create two companions for same owner
	companion1 := world.CreateEntity()
	companion1.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 50,
	})

	companion2 := world.CreateEntity()
	companion2.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 50,
	})

	sys.Update([]*Entity{companion1, companion2}, 0.5)
	sys.Update([]*Entity{companion1, companion2}, 0.5)

	// Bonuses should stack
	dmgBonus, _ := sys.GetBonusForOwner(owner.ID)
	// Two companions at 50 loyalty: 2 * (50 * 0.001 * 1.2) = 0.12
	if dmgBonus < 0.1 {
		t.Errorf("stacked damage bonus = %f, expected >= 0.1", dmgBonus)
	}
}

func TestCompanionSpellAmplificationSystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre    string
		wantMult float64
	}{
		{"fantasy", 1.2},
		{"scifi", 0.7},
		{"horror", 1.0},
		{"cyberpunk", 0.6},
		{"postapoc", 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewCompanionSpellAmplificationSystem(world, 12345)
			sys.SetGenre(tt.genre)

			owner := world.CreateEntity()

			companion := world.CreateEntity()
			companion.AddComponent(&CompanionComponent{
				OwnerID: owner.ID,
				Loyalty: 100,
			})

			sys.Update([]*Entity{companion}, 0.5)
			sys.Update([]*Entity{companion}, 0.5)

			// Base: 100 * 0.001 = 0.1, with genre multiplier
			expectedBonus := 0.1 * tt.wantMult
			dmgBonus, _ := sys.GetBonusForOwner(owner.ID)

			if dmgBonus < expectedBonus*0.9 || dmgBonus > expectedBonus*1.1 {
				t.Errorf("genre %s: damage bonus = %f, expected around %f", tt.genre, dmgBonus, expectedBonus)
			}
		})
	}
}

func TestCompanionSpellAmplificationSystem_HasActiveBonus(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)

	owner := world.CreateEntity()

	// No companion yet
	if sys.HasActiveBonus(owner.ID) {
		t.Error("should not have bonus without companion")
	}

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 50,
	})

	sys.Update([]*Entity{companion}, 0.5)
	sys.Update([]*Entity{companion}, 0.5)

	if !sys.HasActiveBonus(owner.ID) {
		t.Error("should have bonus with companion")
	}
}

func TestCompanionSpellAmplificationSystem_ZeroLoyalty(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)

	owner := world.CreateEntity()

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 0, // No loyalty
	})

	sys.Update([]*Entity{companion}, 0.5)
	sys.Update([]*Entity{companion}, 0.5)

	dmgBonus, healBonus := sys.GetBonusForOwner(owner.ID)
	// Should still be 0 with no loyalty and no perks
	if dmgBonus != 0 || healBonus != 0 {
		t.Errorf("expected 0 bonus at 0 loyalty, got dmg=%f heal=%f", dmgBonus, healBonus)
	}
}

func TestCompanionSpellAmplificationSystem_NegativeLoyalty(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)

	owner := world.CreateEntity()

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: -50, // Negative loyalty (clamped to 0)
	})

	sys.Update([]*Entity{companion}, 0.5)
	sys.Update([]*Entity{companion}, 0.5)

	dmgBonus, _ := sys.GetBonusForOwner(owner.ID)
	// Should be clamped to 0
	if dmgBonus < 0 {
		t.Errorf("negative loyalty should not give negative bonus, got %f", dmgBonus)
	}
}

func TestCompanionSpellAmplificationSystem_GetActiveBonusCount(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)

	owner1 := world.CreateEntity()
	owner2 := world.CreateEntity()

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

	sys.Update([]*Entity{companion1, companion2}, 0.5)
	sys.Update([]*Entity{companion1, companion2}, 0.5)

	if count := sys.GetActiveBonusCount(); count != 2 {
		t.Errorf("GetActiveBonusCount = %d, want 2", count)
	}
}

func TestCompanionSpellAmplificationSystem_ThrottledUpdate(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)

	owner := world.CreateEntity()

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID: owner.ID,
		Loyalty: 50,
	})

	// First update with small delta - should not calculate yet
	sys.Update([]*Entity{companion}, 0.1)

	// Should not have bonus yet (throttled)
	if sys.HasActiveBonus(owner.ID) {
		t.Error("should not calculate on first small delta update")
	}

	// Continue updating until threshold
	sys.Update([]*Entity{companion}, 0.5)

	// Now should have bonus
	if !sys.HasActiveBonus(owner.ID) {
		t.Error("should have bonus after throttle threshold")
	}
}

func BenchmarkCompanionSpellAmplificationSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewCompanionSpellAmplificationSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create 100 companions with different owners
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		owner := world.CreateEntity()
		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			OwnerID:      owner.ID,
			Loyalty:      float64(50 + i%50),
			BondingPerks: []BondingPerk{PerkExtraDamage},
		})
		entities[i] = companion
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.5)
	}
}
