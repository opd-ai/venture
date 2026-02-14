package engine

import (
	"testing"
)

func TestNewReputationDefenseBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewReputationDefenseBonusSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.defenseMultipliers == nil {
		t.Error("defenseMultipliers map not initialized")
	}
	if len(sys.genreMultipliers) == 0 {
		t.Error("genreMultipliers not initialized")
	}
}

func TestReputationDefenseBonusSystem_SetFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	sys.SetFactionSystem(factionSys)

	if sys.factionSystem != factionSys {
		t.Error("FactionSystem not set correctly")
	}
}

func TestReputationDefenseBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

	sys.SetGenre("cyberpunk")

	if sys.genreID != "cyberpunk" {
		t.Errorf("GenreID = %s, want cyberpunk", sys.genreID)
	}
}

func TestReputationDefenseBonusSystem_CalculateBonusPercent(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

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
		{"neutral_low", 1, 0.02, 0.02},
		{"neutral_mid", 25, 0.02, 0.02},
		{"neutral_high", 50, 0.02, 0.02},
		{"friendly_low", 51, 0.04, 0.05},
		{"friendly_mid", 75, 0.07, 0.09},
		{"friendly_max", 100, 0.11, 0.13},
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

func TestReputationDefenseBonusSystem_GetDefenseMultiplier_NoFaction(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

	attacker := NewEntity()
	world.AddEntity(attacker)

	mult := sys.GetDefenseMultiplier(1, attacker)
	if mult != 1.0 {
		t.Errorf("GetDefenseMultiplier = %f, want 1.0 (no faction)", mult)
	}
}

func TestReputationDefenseBonusSystem_GetDefenseMultiplier_NilAttacker(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

	mult := sys.GetDefenseMultiplier(1, nil)
	if mult != 1.0 {
		t.Errorf("GetDefenseMultiplier = %f, want 1.0 (nil attacker)", mult)
	}
}

func TestReputationDefenseBonusSystem_GetDefenseMultiplier_PlayerFaction(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

	attacker := NewEntity()
	attacker.AddComponent(FactionComponent{
		FactionID:       "player_allies",
		Reputation:      100,
		IsPlayerFaction: true,
	})
	world.AddEntity(attacker)

	mult := sys.GetDefenseMultiplier(1, attacker)
	if mult != 1.0 {
		t.Errorf("GetDefenseMultiplier = %f, want 1.0 (player faction)", mult)
	}
}

func TestReputationDefenseBonusSystem_Update_CachesMultipliers(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	bandits := &Faction{ID: "bandits", Name: "Bandits", Relationships: map[string]int{"guards": -100}}
	guards := &Faction{ID: "guards", Name: "Guards", Relationships: map[string]int{"bandits": -100}}
	factionSys.AddFaction(bandits)
	factionSys.AddFaction(guards)

	factionSys.SetPlayerReputation("guards", 75)
	sys.SetFactionSystem(factionSys)

	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	sys.Update([]*Entity{player}, 0.016)

	if _, ok := sys.defenseMultipliers[player.ID]; !ok {
		t.Error("Player defense multipliers not cached after Update")
	}
}

func TestReputationDefenseBonusSystem_GetDefenseMultiplier_WithBonus(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	bandits := &Faction{ID: "bandits", Name: "Bandits", Relationships: map[string]int{}}
	guards := &Faction{ID: "guards", Name: "Guards", Relationships: map[string]int{"bandits": -100}}
	factionSys.AddFaction(bandits)
	factionSys.AddFaction(guards)

	factionSys.SetPlayerReputation("guards", 75)
	sys.SetFactionSystem(factionSys)

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

	mult := sys.GetDefenseMultiplier(player.ID, bandit)
	if mult <= 1.0 {
		t.Errorf("GetDefenseMultiplier = %f, want > 1.0 (bonus against enemy faction)", mult)
	}
	// At 75 rep: base bonus = 0.04 + (75-50)/50*0.08 = 0.04 + 0.04 = 0.08 => multiplier 1.08
	if mult < 1.07 || mult > 1.09 {
		t.Errorf("GetDefenseMultiplier = %f, expected around 1.08", mult)
	}
}

func TestReputationDefenseBonusSystem_GenreMultipliers(t *testing.T) {
	tests := []struct {
		genre   string
		wantMin float64
		wantMax float64
	}{
		{"fantasy", 0.99, 1.01},
		{"scifi", 1.09, 1.11},
		{"horror", 0.74, 0.76},
		{"cyberpunk", 1.14, 1.16},
		{"postapoc", 0.89, 0.91},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewReputationDefenseBonusSystem(world, 12345)
			sys.SetGenre(tt.genre)

			mult := sys.GetGenreMultiplier()
			if mult < tt.wantMin || mult > tt.wantMax {
				t.Errorf("GetGenreMultiplier() = %f, want between %f and %f",
					mult, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestReputationDefenseBonusSystem_OnDefend_NoBonus(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

	defender := NewEntity()
	defender.AddComponent(NewStubInput())
	world.AddEntity(defender)

	attacker := NewEntity()
	world.AddEntity(attacker)

	result := sys.OnDefend(defender, attacker, 100.0)
	if result != 100.0 {
		t.Errorf("OnDefend = %f, want 100.0 (no bonus)", result)
	}
}

func TestReputationDefenseBonusSystem_OnDefend_WithBonus(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	bandits := &Faction{ID: "bandits", Name: "Bandits", Relationships: map[string]int{}}
	guards := &Faction{ID: "guards", Name: "Guards", Relationships: map[string]int{"bandits": -100}}
	factionSys.AddFaction(bandits)
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100)
	sys.SetFactionSystem(factionSys)

	defender := NewEntity()
	defender.AddComponent(NewStubInput())
	world.AddEntity(defender)

	bandit := NewEntity()
	bandit.AddComponent(FactionComponent{
		FactionID:       "bandits",
		Reputation:      0,
		IsPlayerFaction: false,
	})
	world.AddEntity(bandit)

	sys.Update([]*Entity{defender}, 0.016)

	result := sys.OnDefend(defender, bandit, 100.0)
	if result >= 100.0 {
		t.Errorf("OnDefend = %f, want < 100.0 (with defense bonus)", result)
	}
	// At 100 rep: bonus = 0.04 + (100-50)/50*0.08 = 0.12, multiplier 1.12
	// damage = 100/1.12 ≈ 89.3
	if result < 88.0 || result > 91.0 {
		t.Errorf("OnDefend = %f, expected around 89.3", result)
	}
}

func TestReputationDefenseBonusSystem_OnDefend_NilEntities(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

	result := sys.OnDefend(nil, NewEntity(), 100.0)
	if result != 100.0 {
		t.Errorf("OnDefend with nil defender = %f, want 100.0", result)
	}

	result = sys.OnDefend(NewEntity(), nil, 100.0)
	if result != 100.0 {
		t.Errorf("OnDefend with nil attacker = %f, want 100.0", result)
	}
}

func TestReputationDefenseBonusSystem_GetBonusPercent(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 75)
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	bonus := sys.GetBonusPercent("guards")
	// At 75 rep, base bonus is 0.08, genre mult is 1.0
	if bonus < 0.07 || bonus > 0.09 {
		t.Errorf("GetBonusPercent = %f, expected around 0.08", bonus)
	}
}

func TestReputationDefenseBonusSystem_GetBonusPercent_HorrorGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	cult := &Faction{ID: "cult", Name: "The Cult", Enemies: map[string]bool{}}
	factionSys.AddFaction(cult)
	factionSys.SetPlayerReputation("cult", 75)
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("horror") // 0.75x multiplier

	bonus := sys.GetBonusPercent("cult")
	// At 75 rep, base bonus 0.08, horror 0.75x = 0.06
	if bonus < 0.05 || bonus > 0.07 {
		t.Errorf("GetBonusPercent = %f, expected around 0.06", bonus)
	}
}

func TestReputationDefenseBonusSystem_Update_ClearsCacheEachFrame(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	sys.Update([]*Entity{player}, 0.016)

	sys.defenseMultipliers[999] = map[string]float64{"fake": 1.5}

	sys.Update([]*Entity{player}, 0.016)

	if _, ok := sys.defenseMultipliers[999]; ok {
		t.Error("Cache should be cleared each frame")
	}
}

func TestReputationDefenseBonusSystem_NoFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)

	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	sys.Update([]*Entity{player}, 0.016)

	bonus := sys.GetBonusPercent("any")
	if bonus != 0.0 {
		t.Errorf("GetBonusPercent without FactionSystem = %f, want 0.0", bonus)
	}
}

func TestReputationDefenseBonusSystem_UnknownGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	sys.SetGenre("steampunk") // Not in genreMultipliers

	mult := sys.GetGenreMultiplier()
	if mult != 1.0 {
		t.Errorf("GetGenreMultiplier for unknown genre = %f, want 1.0", mult)
	}
}

func BenchmarkReputationDefenseBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	for i := 0; i < 5; i++ {
		faction := &Faction{
			ID:      "faction" + string(rune('a'+i)),
			Name:    "Faction " + string(rune('A'+i)),
			Enemies: map[string]bool{},
		}
		factionSys.AddFaction(faction)
	}
	sys.SetFactionSystem(factionSys)

	player := NewEntity()
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	entities := []*Entity{player}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkReputationDefenseBonusSystem_GetDefenseMultiplier(b *testing.B) {
	world := NewWorld()
	sys := NewReputationDefenseBonusSystem(world, 12345)
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

	attacker := NewEntity()
	attacker.AddComponent(FactionComponent{FactionID: "bandits"})
	world.AddEntity(attacker)

	sys.Update([]*Entity{player}, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.GetDefenseMultiplier(player.ID, attacker)
	}
}
