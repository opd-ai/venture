package engine

import (
	"testing"
)

func TestNewFactionXPBonusSystem(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)

	if system == nil {
		t.Fatal("NewFactionXPBonusSystem returned nil")
	}
	if system.world != world {
		t.Error("System world not set correctly")
	}
	if system.rng == nil {
		t.Error("System RNG not initialized")
	}
}

func TestFactionXPBonusSystem_CalculateBonusPercent(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)

	tests := []struct {
		name       string
		reputation int
		wantMin    float64
		wantMax    float64
	}{
		{"hostile_low", -100, 0.0, 0.0},
		{"hostile_high", -50, 0.0, 0.0},
		{"suspicious", -25, 0.0, 0.0},
		{"suspicious_zero", 0, 0.0, 0.0},
		{"neutral_low", 1, 0.05, 0.05},
		{"neutral_mid", 25, 0.05, 0.05},
		{"neutral_high", 50, 0.05, 0.05},
		{"friendly_low", 51, 0.10, 0.11},
		{"friendly_mid", 75, 0.17, 0.18},
		{"friendly_max", 100, 0.24, 0.26},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.calculateBonusPercent(tt.reputation)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateBonusPercent(%d) = %v, want between %v and %v",
					tt.reputation, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFactionXPBonusSystem_OnEnemyKilled_NoFactionSystem(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)

	killer := world.CreateEntity()
	killer.AddComponent(NewStubInput())

	victim := world.CreateEntity()
	victim.AddComponent(FactionComponent{FactionID: "bandits"})

	// Should not panic without faction system
	system.OnEnemyKilled(killer, victim, 100)

	if system.GetPendingBonusCount() != 0 {
		t.Error("Should not queue bonuses without faction system")
	}
}

func TestFactionXPBonusSystem_OnEnemyKilled_NonPlayer(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)
	factionSystem := NewFactionSystem(world, nil)
	system.SetFactionSystem(factionSystem)

	// Killer without input component (NPC)
	killer := world.CreateEntity()
	victim := world.CreateEntity()
	victim.AddComponent(FactionComponent{FactionID: "bandits"})

	system.OnEnemyKilled(killer, victim, 100)

	if system.GetPendingBonusCount() != 0 {
		t.Error("Should not queue bonuses for non-player killers")
	}
}

func TestFactionXPBonusSystem_OnEnemyKilled_WithAlliedFaction(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)
	factionSystem := NewFactionSystem(world, nil)
	system.SetFactionSystem(factionSystem)

	// Create merchant guild faction that is enemies with bandits
	merchantGuild := &Faction{
		ID:            "merchant_guild",
		Name:          "Merchant Guild",
		Type:          FactionTypeMerchants,
		Relationships: map[string]int{"bandits": -75}, // Enemy of bandits
	}
	factionSystem.AddFaction(merchantGuild)

	// Create player with good reputation with merchant guild
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(FactionComponent{
		FactionID:       "merchant_guild",
		Reputation:      75, // Friendly
		IsPlayerFaction: true,
	})

	// Create bandit victim
	victim := world.CreateEntity()
	victim.AddComponent(FactionComponent{
		FactionID:       "bandits",
		IsPlayerFaction: false,
	})

	system.OnEnemyKilled(player, victim, 100)

	if system.GetPendingBonusCount() != 1 {
		t.Errorf("Expected 1 pending bonus, got %d", system.GetPendingBonusCount())
	}
}

func TestFactionXPBonusSystem_OnEnemyKilled_HostileFaction(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)
	factionSystem := NewFactionSystem(world, nil)
	system.SetFactionSystem(factionSystem)

	// Create a faction that is enemies with bandits
	guards := &Faction{
		ID:            "guards",
		Name:          "City Guards",
		Type:          FactionTypeKingdom,
		Relationships: map[string]int{"bandits": -60},
	}
	factionSystem.AddFaction(guards)

	// Player is hostile with guards
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(FactionComponent{
		FactionID:       "guards",
		Reputation:      -60, // Hostile
		IsPlayerFaction: true,
	})

	// Kill a bandit
	victim := world.CreateEntity()
	victim.AddComponent(FactionComponent{
		FactionID:       "bandits",
		IsPlayerFaction: false,
	})

	system.OnEnemyKilled(player, victim, 100)

	// Should not get bonus because player is hostile with guards
	if system.GetPendingBonusCount() != 0 {
		t.Errorf("Expected 0 pending bonuses for hostile faction, got %d", system.GetPendingBonusCount())
	}
}

func TestFactionXPBonusSystem_Update_AwardsBonuses(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)
	progressionSystem := NewProgressionSystem(world)
	system.SetProgressionSystem(progressionSystem)

	// Create player with experience component
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	exp := NewExperienceComponent()
	exp.Level = 1
	exp.CurrentXP = 0
	exp.RequiredXP = 100
	player.AddComponent(exp)

	// Flush pending entity additions by calling World.Update
	// This moves entities from entitiesToAdd to the entities map
	world.Update(0)

	// Verify entity is in world
	retrievedEntity, ok := world.GetEntity(player.ID)
	if !ok || retrievedEntity == nil {
		t.Fatalf("Entity not found in world after flush")
	}

	// Manually add a pending bonus
	system.pendingBonuses = append(system.pendingBonuses, xpBonus{
		entityID:  player.ID,
		bonusXP:   10,
		factionID: "merchant_guild",
		reason:    "Killed enemy of Merchant Guild",
	})

	// Run update
	system.Update([]*Entity{player}, 0.016)

	// Check XP was awarded - retrieve component fresh from the same entity
	expComp, ok := player.GetComponent("experience")
	if !ok {
		t.Fatal("Experience component not found on player")
	}
	gotXP := expComp.(*ExperienceComponent).CurrentXP
	if gotXP != 10 {
		t.Errorf("Expected 10 XP after bonus, got %d", gotXP)
	}

	// Check bonuses were cleared
	if system.GetPendingBonusCount() != 0 {
		t.Errorf("Expected 0 pending bonuses after update, got %d", system.GetPendingBonusCount())
	}
}

func TestFactionXPBonusSystem_GetBonusPercent(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)

	// Without faction system
	bonus := system.GetBonusPercent("any_faction")
	if bonus != 0.0 {
		t.Errorf("Expected 0.0 without faction system, got %v", bonus)
	}

	// With faction system
	factionSystem := NewFactionSystem(world, nil)
	system.SetFactionSystem(factionSystem)

	bonus = system.GetBonusPercent("unknown_faction")
	if bonus != 0.0 {
		t.Errorf("Expected 0.0 for unknown faction (0 rep), got %v", bonus)
	}
}

func TestFactionXPBonusSystem_MultipleFactions(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)
	factionSystem := NewFactionSystem(world, nil)
	system.SetFactionSystem(factionSystem)

	// Create two factions both enemies with bandits
	merchantGuild := &Faction{
		ID:            "merchant_guild",
		Name:          "Merchant Guild",
		Type:          FactionTypeMerchants,
		Relationships: map[string]int{"bandits": -75},
	}
	cityGuards := &Faction{
		ID:            "city_guards",
		Name:          "City Guards",
		Type:          FactionTypeKingdom,
		Relationships: map[string]int{"bandits": -80},
	}
	factionSystem.AddFaction(merchantGuild)
	factionSystem.AddFaction(cityGuards)

	// Player friendly with both
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	// Note: In real game, player would have reputation tracked differently
	// For this test, GetPlayerReputation returns 0 (neutral) by default

	// Kill a bandit
	victim := world.CreateEntity()
	victim.AddComponent(FactionComponent{
		FactionID:       "bandits",
		IsPlayerFaction: false,
	})

	system.OnEnemyKilled(player, victim, 100)

	// Player has 0 reputation with both factions, so no bonuses
	// (reputation must be > 0 for bonus)
	if system.GetPendingBonusCount() != 0 {
		t.Errorf("Expected 0 bonuses for neutral reputation, got %d", system.GetPendingBonusCount())
	}
}

func TestFactionXPBonusSystem_NilInputs(t *testing.T) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)
	factionSystem := NewFactionSystem(world, nil)
	system.SetFactionSystem(factionSystem)

	// Should not panic with nil inputs
	system.OnEnemyKilled(nil, nil, 100)
	system.OnEnemyKilled(world.CreateEntity(), nil, 100)

	victim := world.CreateEntity()
	system.OnEnemyKilled(nil, victim, 100)

	if system.GetPendingBonusCount() != 0 {
		t.Error("Should not queue bonuses for nil inputs")
	}
}

func BenchmarkFactionXPBonusSystem_OnEnemyKilled(b *testing.B) {
	world := NewWorld()
	system := NewFactionXPBonusSystem(world, 12345)
	factionSystem := NewFactionSystem(world, nil)
	system.SetFactionSystem(factionSystem)

	merchantGuild := &Faction{
		ID:            "merchant_guild",
		Name:          "Merchant Guild",
		Type:          FactionTypeMerchants,
		Relationships: map[string]int{"bandits": -75},
	}
	factionSystem.AddFaction(merchantGuild)

	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(FactionComponent{
		FactionID:       "merchant_guild",
		Reputation:      75,
		IsPlayerFaction: true,
	})

	victim := world.CreateEntity()
	victim.AddComponent(FactionComponent{
		FactionID:       "bandits",
		IsPlayerFaction: false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnEnemyKilled(player, victim, 100)
		system.pendingBonuses = system.pendingBonuses[:0] // Reset for next iteration
	}
}
