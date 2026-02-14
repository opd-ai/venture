package engine

import (
	"testing"
)

func TestNewFactionDamageBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewFactionDamageBonusSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.damageMultipliers == nil {
		t.Error("damageMultipliers map not initialized")
	}
	if len(sys.genreMultipliers) == 0 {
		t.Error("genreMultipliers not initialized")
	}
}

func TestFactionDamageBonusSystem_SetFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	sys.SetFactionSystem(factionSys)

	if sys.factionSystem != factionSys {
		t.Error("FactionSystem not set correctly")
	}
}

func TestFactionDamageBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)

	sys.SetGenre("cyberpunk")

	if sys.genreID != "cyberpunk" {
		t.Errorf("GenreID = %s, want cyberpunk", sys.genreID)
	}
}

func TestFactionDamageBonusSystem_CalculateBonusPercent(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)

	tests := []struct {
		name       string
		reputation int
		wantMin    float64
		wantMax    float64
	}{
		{"hostile", -100, 0.0, 0.0},
		{"hostile_edge", -50, 0.0, 0.0},
		{"suspicious", -25, 0.0, 0.0},
		{"suspicious_edge", 0, 0.0, 0.0},
		{"neutral_low", 1, 0.03, 0.03},
		{"neutral_mid", 25, 0.03, 0.03},
		{"neutral_high", 50, 0.03, 0.03},
		{"friendly_low", 51, 0.05, 0.06},
		{"friendly_mid", 75, 0.09, 0.11},
		{"friendly_max", 100, 0.14, 0.16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.calculateBonusPercent(tt.reputation)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateBonusPercent(%d) = %f, want between %f and %f",
					tt.reputation, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFactionDamageBonusSystem_GetDamageMultiplier_NoFaction(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)

	target := NewEntity()
	world.AddEntity(target)

	// No faction component - should return 1.0
	mult := sys.GetDamageMultiplier(1, target)
	if mult != 1.0 {
		t.Errorf("GetDamageMultiplier = %f, want 1.0 (no faction)", mult)
	}
}

func TestFactionDamageBonusSystem_GetDamageMultiplier_NilTarget(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)

	mult := sys.GetDamageMultiplier(1, nil)
	if mult != 1.0 {
		t.Errorf("GetDamageMultiplier = %f, want 1.0 (nil target)", mult)
	}
}

func TestFactionDamageBonusSystem_GetDamageMultiplier_PlayerFaction(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)

	target := NewEntity()
	target.AddComponent(FactionComponent{
		FactionID:       "player_allies",
		Reputation:      100,
		IsPlayerFaction: true,
	})
	world.AddEntity(target)

	// Player faction targets should not get bonus damage
	mult := sys.GetDamageMultiplier(1, target)
	if mult != 1.0 {
		t.Errorf("GetDamageMultiplier = %f, want 1.0 (player faction)", mult)
	}
}

func TestFactionDamageBonusSystem_Update_CachesMultipliers(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	// Set up factions with relationships (negative = enemies)
	bandits := &Faction{ID: "bandits", Name: "Bandits", Relationships: map[string]int{"guards": -100}}
	guards := &Faction{ID: "guards", Name: "Guards", Relationships: map[string]int{"bandits": -100}}
	factionSys.AddFaction(bandits)
	factionSys.AddFaction(guards)

	// Player has good reputation with guards
	factionSys.SetPlayerReputation("guards", 75)
	sys.SetFactionSystem(factionSys)

	// Create player entity
	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// Run update
	sys.Update([]*Entity{player}, 0.016)

	// Check that multipliers were cached
	if _, ok := sys.damageMultipliers[player.ID]; !ok {
		t.Error("Player multipliers not cached after Update")
	}
}

func TestFactionDamageBonusSystem_GetDamageMultiplier_WithBonus(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	// Set up factions: Guards hate Bandits (relationship <= -50 means enemy)
	bandits := &Faction{ID: "bandits", Name: "Bandits", Relationships: map[string]int{}}
	guards := &Faction{ID: "guards", Name: "Guards", Relationships: map[string]int{"bandits": -100}}
	factionSys.AddFaction(bandits)
	factionSys.AddFaction(guards)

	// Player is friendly with guards
	factionSys.SetPlayerReputation("guards", 75)
	sys.SetFactionSystem(factionSys)

	// Create player entity
	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// Create bandit target
	bandit := NewEntity()
	bandit.AddComponent(FactionComponent{
		FactionID:       "bandits",
		Reputation:      0,
		IsPlayerFaction: false,
	})
	world.AddEntity(bandit)

	// Run update to cache multipliers
	sys.Update([]*Entity{player}, 0.016)

	// Check damage multiplier
	mult := sys.GetDamageMultiplier(player.ID, bandit)
	if mult <= 1.0 {
		t.Errorf("GetDamageMultiplier = %f, want > 1.0 (bonus against enemy faction)", mult)
	}
	// Expect ~10% bonus at 75 reputation
	if mult < 1.08 || mult > 1.12 {
		t.Errorf("GetDamageMultiplier = %f, expected around 1.10", mult)
	}
}

func TestFactionDamageBonusSystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre   string
		wantMin float64
		wantMax float64
	}{
		{"fantasy", 0.99, 1.01},
		{"scifi", 1.09, 1.11},
		{"horror", 0.79, 0.81},
		{"cyberpunk", 1.14, 1.16},
		{"postapoc", 0.89, 0.91},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewFactionDamageBonusSystem(world, 12345)
			sys.SetGenre(tt.genre)

			mult := sys.GetGenreMultiplier()
			if mult < tt.wantMin || mult > tt.wantMax {
				t.Errorf("GetGenreMultiplier() = %f, want between %f and %f",
					mult, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFactionDamageBonusSystem_OnAttack_NoBonus(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)

	attacker := NewEntity()
	attacker.AddComponent(NewStubInput())
	world.AddEntity(attacker)

	target := NewEntity()
	world.AddEntity(target)

	// No faction - damage should be unchanged
	result := sys.OnAttack(attacker, target, 100.0)
	if result != 100.0 {
		t.Errorf("OnAttack = %f, want 100.0 (no bonus)", result)
	}
}

func TestFactionDamageBonusSystem_OnAttack_WithBonus(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	// Set up factions with relationships
	bandits := &Faction{ID: "bandits", Name: "Bandits", Relationships: map[string]int{}}
	guards := &Faction{ID: "guards", Name: "Guards", Relationships: map[string]int{"bandits": -100}}
	factionSys.AddFaction(bandits)
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100)
	sys.SetFactionSystem(factionSys)

	// Create entities
	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	bandit := NewEntity()
	bandit.AddComponent(FactionComponent{
		FactionID:       "bandits",
		Reputation:      0,
		IsPlayerFaction: false,
	})
	world.AddEntity(bandit)

	sys.Update([]*Entity{player}, 0.016)

	// Test OnAttack applies bonus
	result := sys.OnAttack(player, bandit, 100.0)
	if result <= 100.0 {
		t.Errorf("OnAttack = %f, want > 100.0 (with bonus)", result)
	}
	// Expect ~15% bonus at max reputation
	if result < 114.0 || result > 116.0 {
		t.Errorf("OnAttack = %f, expected around 115.0", result)
	}
}

func TestFactionDamageBonusSystem_OnAttack_NilEntities(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)

	// Nil attacker
	result := sys.OnAttack(nil, NewEntity(), 100.0)
	if result != 100.0 {
		t.Errorf("OnAttack with nil attacker = %f, want 100.0", result)
	}

	// Nil target
	result = sys.OnAttack(NewEntity(), nil, 100.0)
	if result != 100.0 {
		t.Errorf("OnAttack with nil target = %f, want 100.0", result)
	}
}

func TestFactionDamageBonusSystem_GetBonusPercent(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 75)
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	bonus := sys.GetBonusPercent("guards")
	// At 75 rep, base bonus is ~10%, genre mult is 1.0
	if bonus < 0.09 || bonus > 0.11 {
		t.Errorf("GetBonusPercent = %f, expected around 0.10", bonus)
	}
}

func TestFactionDamageBonusSystem_GetBonusPercent_CyberpunkGenre(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	corps := &Faction{ID: "corps", Name: "Corporations", Enemies: map[string]bool{}}
	factionSys.AddFaction(corps)
	factionSys.SetPlayerReputation("corps", 75)
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("cyberpunk") // 1.15x multiplier

	bonus := sys.GetBonusPercent("corps")
	// At 75 rep, base bonus ~10%, cyberpunk 1.15x = ~11.5%
	if bonus < 0.10 || bonus > 0.13 {
		t.Errorf("GetBonusPercent = %f, expected around 0.115", bonus)
	}
}

func TestFactionDamageBonusSystem_Update_ClearsCacheEachFrame(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// First update
	sys.Update([]*Entity{player}, 0.016)

	// Add a fake entry to verify it gets cleared
	sys.damageMultipliers[999] = map[string]float64{"fake": 1.5}

	// Second update should clear the fake entry
	sys.Update([]*Entity{player}, 0.016)

	if _, ok := sys.damageMultipliers[999]; ok {
		t.Error("Cache should be cleared each frame")
	}
}

func TestFactionDamageBonusSystem_NoFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	// No faction system set

	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// Should not panic, should just skip processing
	sys.Update([]*Entity{player}, 0.016)

	bonus := sys.GetBonusPercent("any")
	if bonus != 0.0 {
		t.Errorf("GetBonusPercent without FactionSystem = %f, want 0.0", bonus)
	}
}

func BenchmarkFactionDamageBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	// Create several factions
	for i := 0; i < 5; i++ {
		faction := &Faction{
			ID:      "faction" + string(rune('a'+i)),
			Name:    "Faction " + string(rune('A'+i)),
			Enemies: map[string]bool{},
		}
		factionSys.AddFaction(faction)
	}
	sys.SetFactionSystem(factionSys)

	// Create player entity
	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	entities := []*Entity{player}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkFactionDamageBonusSystem_GetDamageMultiplier(b *testing.B) {
	world := NewWorld()
	sys := NewFactionDamageBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	bandits := &Faction{ID: "bandits", Name: "Bandits", Enemies: map[string]bool{}}
	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{"bandits": true}}
	factionSys.AddFaction(bandits)
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 75)
	sys.SetFactionSystem(factionSys)

	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	target := NewEntity()
	target.AddComponent(FactionComponent{FactionID: "bandits"})
	world.AddEntity(target)

	sys.Update([]*Entity{player}, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.GetDamageMultiplier(player.ID, target)
	}
}
