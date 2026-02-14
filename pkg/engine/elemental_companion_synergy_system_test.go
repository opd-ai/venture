package engine

import (
	"testing"
)

func TestNewElementalCompanionSynergySystem(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)

	if system == nil {
		t.Fatal("NewElementalCompanionSynergySystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.seed != 12345 {
		t.Errorf("seed = %d, want 12345", system.seed)
	}
	if system.rng == nil {
		t.Error("rng not initialized")
	}
	if system.activeSynergies == nil {
		t.Error("activeSynergies map not initialized")
	}
}

func TestElementalCompanionSynergySystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)

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
				t.Errorf("genreID = %s, want %s", system.genreID, tt.genre)
			}
		})
	}
}

func TestElementalCompanionSynergySystem_Update_NoCompanions(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)

	// Create entity without companion component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	if system.GetActiveSynergyCount() != 0 {
		t.Errorf("GetActiveSynergyCount = %d, want 0", system.GetActiveSynergyCount())
	}
}

func TestElementalCompanionSynergySystem_Update_ElementalCompanionWithOwnerEffect(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)
	system.SetGenre("fantasy")

	// Create owner with burning status effect
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(&StatusEffectComponent{
		EffectType: "burning",
		Duration:   5.0,
		Magnitude:  10.0,
	})

	// Create elemental companion
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 110, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypeElemental,
		Loyalty:       100.0,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  100.0,
		Defense: 50.0,
		Speed:   20.0,
	})

	entities := []*Entity{owner, companion}
	system.Update(entities, 0.016)

	// Check that synergy was applied
	if !system.HasActiveSynergy(companion.ID) {
		t.Error("companion should have active synergy")
	}
	if system.GetActiveSynergyCount() != 1 {
		t.Errorf("GetActiveSynergyCount = %d, want 1", system.GetActiveSynergyCount())
	}

	// Check stats were boosted
	statsComp, _ := companion.GetComponent("companion_stats")
	stats := statsComp.(*CompanionStatsComponent)
	expectedAttack := 100.0 * (1.0 + 0.25) // 25% bonus
	if stats.Attack != expectedAttack {
		t.Errorf("Attack = %f, want %f", stats.Attack, expectedAttack)
	}
}

func TestElementalCompanionSynergySystem_Update_NonElementalCompanion(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)
	system.SetGenre("fantasy")

	// Create owner with burning status effect
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(&StatusEffectComponent{
		EffectType: "burning",
		Duration:   5.0,
		Magnitude:  10.0,
	})

	// Create non-elemental companion (pet type)
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 110, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet, // Not elemental
		Loyalty:       100.0,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  100.0,
		Defense: 50.0,
		Speed:   20.0,
	})

	entities := []*Entity{owner, companion}
	system.Update(entities, 0.016)

	// Should NOT have synergy - not an elemental companion
	if system.HasActiveSynergy(companion.ID) {
		t.Error("non-elemental companion should not have synergy")
	}

	// Stats should be unchanged
	statsComp, _ := companion.GetComponent("companion_stats")
	stats := statsComp.(*CompanionStatsComponent)
	if stats.Attack != 100.0 {
		t.Errorf("Attack = %f, want 100.0 (unchanged)", stats.Attack)
	}
}

func TestElementalCompanionSynergySystem_Update_OwnerWithoutEffect(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)
	system.SetGenre("fantasy")

	// Create owner WITHOUT status effect
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create elemental companion
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 110, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypeElemental,
		Loyalty:       100.0,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  100.0,
		Defense: 50.0,
		Speed:   20.0,
	})

	entities := []*Entity{owner, companion}
	system.Update(entities, 0.016)

	// Should NOT have synergy - owner has no elemental effect
	if system.HasActiveSynergy(companion.ID) {
		t.Error("companion should not have synergy when owner lacks elemental effects")
	}
}

func TestElementalCompanionSynergySystem_Update_SynergyRemoval(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)
	system.SetGenre("fantasy")

	// Create owner with burning status effect
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	burningEffect := &StatusEffectComponent{
		EffectType: "burning",
		Duration:   5.0,
		Magnitude:  10.0,
	}
	owner.AddComponent(burningEffect)

	// Create elemental companion
	companion := world.CreateEntity()
	companion.AddComponent(&PositionComponent{X: 110, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypeElemental,
		Loyalty:       100.0,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  100.0,
		Defense: 50.0,
		Speed:   20.0,
	})

	entities := []*Entity{owner, companion}

	// First update - apply synergy
	system.Update(entities, 0.016)
	if !system.HasActiveSynergy(companion.ID) {
		t.Fatal("expected synergy after first update")
	}

	// Expire the status effect
	burningEffect.Duration = 0

	// Second update - remove synergy
	system.Update(entities, 0.016)
	if system.HasActiveSynergy(companion.ID) {
		t.Error("synergy should be removed when owner effect expires")
	}

	// Stats should be back to original
	statsComp, _ := companion.GetComponent("companion_stats")
	stats := statsComp.(*CompanionStatsComponent)
	if stats.Attack < 99.9 || stats.Attack > 100.1 {
		t.Errorf("Attack = %f, want ~100.0 (restored)", stats.Attack)
	}
}

func TestElementalCompanionSynergySystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre        string
		expectedMult float64
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
			system := NewElementalCompanionSynergySystem(world, 12345)
			system.SetGenre(tt.genre)

			// Create owner with burning status effect
			owner := world.CreateEntity()
			owner.AddComponent(&StatusEffectComponent{
				EffectType: "burning",
				Duration:   5.0,
				Magnitude:  10.0,
			})

			// Create elemental companion
			companion := world.CreateEntity()
			companion.AddComponent(&CompanionComponent{
				OwnerID:       owner.ID,
				CompanionType: CompanionTypeElemental,
			})
			companion.AddComponent(&CompanionStatsComponent{
				Attack:  100.0,
				Defense: 50.0,
				Speed:   20.0,
			})

			entities := []*Entity{owner, companion}
			system.Update(entities, 0.016)

			statsComp, _ := companion.GetComponent("companion_stats")
			stats := statsComp.(*CompanionStatsComponent)

			expectedAttack := 100.0 * (1.0 + 0.25*tt.expectedMult)
			if stats.Attack < expectedAttack-0.1 || stats.Attack > expectedAttack+0.1 {
				t.Errorf("Attack = %f, want ~%f for genre %s", stats.Attack, expectedAttack, tt.genre)
			}
		})
	}
}

func TestElementalCompanionSynergySystem_AllElementalEffects(t *testing.T) {
	effects := []string{"burning", "frozen", "shocked", "poisoned", "wet"}

	for _, effect := range effects {
		t.Run(effect, func(t *testing.T) {
			world := NewWorld()
			system := NewElementalCompanionSynergySystem(world, 12345)
			system.SetGenre("fantasy")

			owner := world.CreateEntity()
			owner.AddComponent(&StatusEffectComponent{
				EffectType: effect,
				Duration:   5.0,
				Magnitude:  10.0,
			})

			companion := world.CreateEntity()
			companion.AddComponent(&CompanionComponent{
				OwnerID:       owner.ID,
				CompanionType: CompanionTypeElemental,
			})
			companion.AddComponent(&CompanionStatsComponent{
				Attack:  100.0,
				Defense: 50.0,
				Speed:   20.0,
			})

			entities := []*Entity{owner, companion}
			system.Update(entities, 0.016)

			if !system.HasActiveSynergy(companion.ID) {
				t.Errorf("elemental companion should have synergy with owner %s effect", effect)
			}
		})
	}
}

func TestElementalCompanionSynergySystem_NilWorld(t *testing.T) {
	system := NewElementalCompanionSynergySystem(nil, 12345)

	// Should not panic
	entities := []*Entity{}
	system.Update(entities, 0.016)
}

func TestElementalCompanionSynergySystem_MissingOwner(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)

	// Create companion with non-existent owner
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       99999, // Non-existent
		CompanionType: CompanionTypeElemental,
	})
	companion.AddComponent(&CompanionStatsComponent{
		Attack:  100.0,
		Defense: 50.0,
		Speed:   20.0,
	})

	entities := []*Entity{companion}
	system.Update(entities, 0.016)

	// Should not have synergy - owner doesn't exist
	if system.HasActiveSynergy(companion.ID) {
		t.Error("companion with missing owner should not have synergy")
	}
}

func TestElementalCompanionSynergySystem_WithRegularStatsComponent(t *testing.T) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)
	system.SetGenre("fantasy")

	// Create owner with burning effect
	owner := world.CreateEntity()
	owner.AddComponent(&StatusEffectComponent{
		EffectType: "burning",
		Duration:   5.0,
		Magnitude:  10.0,
	})

	// Create companion with regular StatsComponent instead of CompanionStatsComponent
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypeElemental,
	})
	companion.AddComponent(NewStatsComponent()) // Regular stats

	entities := []*Entity{owner, companion}
	system.Update(entities, 0.016)

	// Should still have synergy
	if !system.HasActiveSynergy(companion.ID) {
		t.Error("companion should have synergy even with regular StatsComponent")
	}

	// Check that stats were boosted
	statsComp, _ := companion.GetComponent("stats")
	stats := statsComp.(*StatsComponent)
	expectedAttack := 10.0 * (1.0 + 0.25) // 25% bonus on base 10
	if stats.Attack < expectedAttack-0.1 || stats.Attack > expectedAttack+0.1 {
		t.Errorf("Attack = %f, want ~%f", stats.Attack, expectedAttack)
	}
}

func BenchmarkElementalCompanionSynergySystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewElementalCompanionSynergySystem(world, 12345)
	system.SetGenre("fantasy")

	// Create 100 owner-companion pairs
	entities := make([]*Entity, 0, 200)
	for i := 0; i < 100; i++ {
		owner := world.CreateEntity()
		owner.AddComponent(&StatusEffectComponent{
			EffectType: "burning",
			Duration:   5.0,
			Magnitude:  10.0,
		})
		entities = append(entities, owner)

		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			OwnerID:       owner.ID,
			CompanionType: CompanionTypeElemental,
		})
		companion.AddComponent(&CompanionStatsComponent{
			Attack:  100.0,
			Defense: 50.0,
			Speed:   20.0,
		})
		entities = append(entities, companion)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
