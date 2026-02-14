package engine

import (
	"testing"
)

func TestNewCompanionDamageLifestealSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)

	if system == nil {
		t.Fatal("NewCompanionDamageLifestealSystem returned nil")
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
	if system.baseLifestealPerLoyalty <= 0 {
		t.Error("baseLifestealPerLoyalty not initialized")
	}
}

func TestCompanionDamageLifestealSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestCompanionDamageLifestealSystem_SetGenre(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)

	system.SetGenre("horror")

	if system.genreID != "horror" {
		t.Errorf("genreID = %q, want horror", system.genreID)
	}
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_NilAttacker(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)

	// Should not panic with nil attacker
	system.OnCompanionDamageDealt(nil, nil, 100)
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_ZeroDamage(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	attacker := world.CreateEntity()

	// Should return early with zero damage
	system.OnCompanionDamageDealt(attacker, nil, 0)
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_NotCompanion(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	attacker := world.CreateEntity()

	// Attacker without companion component - should return early
	system.OnCompanionDamageDealt(attacker, nil, 100)
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_NoOwner(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	companion := world.CreateEntity()

	// Add companion component but with no owner
	companion.AddComponent(&CompanionComponent{
		OwnerID:       0,
		CompanionType: CompanionTypePet,
		Loyalty:       80,
	})

	// Should return early with no owner
	system.OnCompanionDamageDealt(companion, nil, 100)
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_OwnerNotFound(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	companion := world.CreateEntity()

	// Add companion component with non-existent owner
	companion.AddComponent(&CompanionComponent{
		OwnerID:       999999,
		CompanionType: CompanionTypePet,
		Loyalty:       80,
	})

	// Should return early when owner not found
	system.OnCompanionDamageDealt(companion, nil, 100)
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_OwnerNoHealth(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)

	owner := world.CreateEntity()
	owner.AddComponent(NewStubInput()) // Mark as player

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       80,
	})

	// Should return early when owner has no health component
	system.OnCompanionDamageDealt(companion, nil, 100)
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_BasicHealing(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	// Create owner with reduced health
	owner := world.CreateEntity()
	owner.AddComponent(NewStubInput())
	owner.AddComponent(&HealthComponent{Current: 50, Max: 100})
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion with high loyalty
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       100, // Max loyalty = 5% base lifesteal * 1.3 fantasy = 6.5%
	})

	// Deal 100 damage through companion
	system.OnCompanionDamageDealt(companion, nil, 100)

	// Check owner was healed
	healthComp, _ := owner.GetComponent("health")
	health := healthComp.(*HealthComponent)

	// Expected: 100 * 0.0005 * 100 * 1.3 = 6.5 heal
	if health.Current <= 50 {
		t.Errorf("owner health = %f, expected > 50", health.Current)
	}
	if health.Current > 60 {
		t.Errorf("owner health = %f, expected < 60 (lifesteal capped)", health.Current)
	}
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_WithPerks(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	owner := world.CreateEntity()
	owner.AddComponent(NewStubInput())
	owner.AddComponent(&HealthComponent{Current: 50, Max: 100})
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Companion with perks
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       80,
		BondingPerks:  []BondingPerk{PerkExtraHealth, PerkSharedExperience},
	})

	system.OnCompanionDamageDealt(companion, nil, 100)

	healthComp, _ := owner.GetComponent("health")
	health := healthComp.(*HealthComponent)

	// Base: 80 * 0.0005 = 4%
	// + PerkExtraHealth: +3%
	// + PerkSharedExp: +2%
	// Total: 9% * 1.3 (fantasy) = 11.7%
	// 100 damage * 11.7% = 11.7 heal
	if health.Current < 55 {
		t.Errorf("owner health = %f, expected >= 55 (with perks)", health.Current)
	}
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_GenreHorror(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	system.SetGenre("horror") // 1.5x multiplier

	owner := world.CreateEntity()
	owner.AddComponent(NewStubInput())
	owner.AddComponent(&HealthComponent{Current: 50, Max: 100})
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypeUndead,
		Loyalty:       100, // 5% * 1.5 = 7.5%
	})

	system.OnCompanionDamageDealt(companion, nil, 100)

	healthComp, _ := owner.GetComponent("health")
	health := healthComp.(*HealthComponent)

	// Horror bonus should give more heal than fantasy
	if health.Current < 56 {
		t.Errorf("owner health = %f, expected >= 56 (horror bonus)", health.Current)
	}
}

func TestCompanionDamageLifestealSystem_OnCompanionDamageDealt_CapAt15Percent(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	system.SetGenre("horror") // 1.5x multiplier

	owner := world.CreateEntity()
	owner.AddComponent(NewStubInput())
	owner.AddComponent(&HealthComponent{Current: 50, Max: 100})
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Max perks + max loyalty + horror = exceeds 15% cap
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypeUndead,
		Loyalty:       100,
		BondingPerks:  []BondingPerk{PerkExtraHealth, PerkSharedExperience},
	})

	system.OnCompanionDamageDealt(companion, nil, 200)

	healthComp, _ := owner.GetComponent("health")
	health := healthComp.(*HealthComponent)

	// Max heal = 200 * 0.15 = 30, but capped at 15% of max health = 15
	// So health should be 50 + 15 = 65
	if health.Current > 70 {
		t.Errorf("owner health = %f, expected <= 70 (15%% cap)", health.Current)
	}
}

func TestCompanionDamageLifestealSystem_GetLifestealForCompanion(t *testing.T) {
	tests := []struct {
		name      string
		companion *CompanionComponent
		genre     string
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "nil companion",
			companion: nil,
			genre:     "fantasy",
			wantMin:   0,
			wantMax:   0,
		},
		{
			name: "zero loyalty",
			companion: &CompanionComponent{
				Loyalty: 0,
			},
			genre:   "fantasy",
			wantMin: 0,
			wantMax: 0.001,
		},
		{
			name: "max loyalty fantasy",
			companion: &CompanionComponent{
				Loyalty: 100,
			},
			genre:   "fantasy",
			wantMin: 0.05, // 5% * 1.3 = 6.5%
			wantMax: 0.08,
		},
		{
			name: "with perks",
			companion: &CompanionComponent{
				Loyalty:      80,
				BondingPerks: []BondingPerk{PerkExtraHealth},
			},
			genre:   "fantasy",
			wantMin: 0.08,
			wantMax: 0.12,
		},
		{
			name: "capped at 15%",
			companion: &CompanionComponent{
				Loyalty:      100,
				BondingPerks: []BondingPerk{PerkExtraHealth, PerkSharedExperience},
			},
			genre:   "horror", // 1.5x would exceed cap
			wantMin: 0.14,
			wantMax: 0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld(nil)
			system := NewCompanionDamageLifestealSystem(world, 12345)
			system.SetGenre(tt.genre)

			got := system.GetLifestealForCompanion(tt.companion)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetLifestealForCompanion() = %v, want [%v, %v]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCompanionDamageLifestealSystem_Update(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)

	// Update should be a no-op (callback-driven)
	system.Update(nil, 0.016)
	system.Update([]*Entity{}, 0.016)
}

func TestCompanionDamageLifestealSystem_ParticleSpawning(t *testing.T) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	system.SetGenre("fantasy")

	owner := world.CreateEntity()
	owner.AddComponent(NewStubInput())
	owner.AddComponent(&HealthComponent{Current: 10, Max: 100})
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       100,
	})

	// Deal enough damage to trigger particles (heal > 0.5)
	system.OnCompanionDamageDealt(companion, nil, 100)

	// Particles should have been spawned
	// We can't easily check particle count, but at least verify no panic
}

func BenchmarkCompanionDamageLifestealSystem_OnCompanionDamageDealt(b *testing.B) {
	world := NewWorld(nil)
	system := NewCompanionDamageLifestealSystem(world, 12345)
	system.SetGenre("fantasy")

	owner := world.CreateEntity()
	owner.AddComponent(NewStubInput())
	owner.AddComponent(&HealthComponent{Current: 50, Max: 100})
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       80,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset owner health
		healthComp, _ := owner.GetComponent("health")
		health := healthComp.(*HealthComponent)
		health.Current = 50

		system.OnCompanionDamageDealt(companion, nil, 50)
	}
}
