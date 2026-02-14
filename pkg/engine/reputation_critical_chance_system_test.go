//go:build ignore

package engine

import (
	"testing"
)

func TestNewReputationCriticalChanceBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewReputationCriticalChanceBonusSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.appliedBonuses == nil {
		t.Error("appliedBonuses map not initialized")
	}
	if len(sys.genreMultipliers) == 0 {
		t.Error("genreMultipliers not initialized")
	}
}

func TestReputationCriticalChanceBonusSystem_SetFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	sys.SetFactionSystem(factionSys)

	if sys.factionSystem != factionSys {
		t.Error("FactionSystem not set correctly")
	}
}

func TestReputationCriticalChanceBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)

	sys.SetGenre("cyberpunk")

	if sys.genreID != "cyberpunk" {
		t.Errorf("GenreID = %s, want cyberpunk", sys.genreID)
	}
}

func TestReputationCriticalChanceBonusSystem_BonusForReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)

	tests := []struct {
		name       string
		reputation int
		want       float64
	}{
		{"hostile", -100, 0.0},
		{"hostile_edge", -50, 0.0},
		{"suspicious", -25, 0.0},
		{"zero", 0, 0.0},
		{"neutral_low", 1, 0.02},
		{"neutral_mid", 25, 0.02},
		{"neutral_high", 50, 0.02},
		{"friendly_low", 51, 0.04},
		{"friendly_mid", 65, 0.04},
		{"friendly_high", 75, 0.04},
		{"honored_low", 76, 0.07},
		{"honored_max", 100, 0.07},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.bonusForReputation(tt.reputation)
			if got != tt.want {
				t.Errorf("bonusForReputation(%d) = %f, want %f",
					tt.reputation, got, tt.want)
			}
		})
	}
}

func TestReputationCriticalChanceBonusSystem_GenreMultipliers(t *testing.T) {
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
			sys := NewReputationCriticalChanceBonusSystem(world, 12345)
			sys.SetGenre(tt.genre)

			mult := sys.GetGenreMultiplier()
			if mult < tt.wantMin || mult > tt.wantMax {
				t.Errorf("GetGenreMultiplier() = %f, want between %f and %f",
					mult, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestReputationCriticalChanceBonusSystem_UnknownGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	sys.SetGenre("steampunk")

	mult := sys.GetGenreMultiplier()
	if mult != 1.0 {
		t.Errorf("GetGenreMultiplier for unknown genre = %f, want 1.0", mult)
	}
}

func TestReputationCriticalChanceBonusSystem_Update_NoFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	player.AddComponent(stats)
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	if stats.CritChance != 0.10 {
		t.Errorf("CritChance changed without FactionSystem: %f", stats.CritChance)
	}
}

func TestReputationCriticalChanceBonusSystem_Update_NeutralReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 25) // Neutral
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	stats := NewStatsComponent()
	stats.CritChance = 0.05
	player.AddComponent(stats)
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	// Expected: 0.05 + 0.02*1.0 = 0.07
	expected := 0.07
	if stats.CritChance < expected-0.001 || stats.CritChance > expected+0.001 {
		t.Errorf("CritChance = %f, expected around %f", stats.CritChance, expected)
	}
}

func TestReputationCriticalChanceBonusSystem_Update_HonoredReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100) // Honored
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	stats := NewStatsComponent()
	stats.CritChance = 0.05
	player.AddComponent(stats)
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	// Expected: 0.05 + 0.07*1.0 = 0.12
	expected := 0.12
	if stats.CritChance < expected-0.001 || stats.CritChance > expected+0.001 {
		t.Errorf("CritChance = %f, expected around %f", stats.CritChance, expected)
	}

	bonus := sys.GetCritBonusPercent(player.ID)
	if bonus < 6.5 || bonus > 7.5 {
		t.Errorf("GetCritBonusPercent = %f, expected around 7.0", bonus)
	}
}

func TestReputationCriticalChanceBonusSystem_Update_HorrorGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	cult := &Faction{ID: "cult", Name: "The Cult", Enemies: map[string]bool{}}
	factionSys.AddFaction(cult)
	factionSys.SetPlayerReputation("cult", 100) // Honored
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("horror") // 0.75 multiplier

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	stats := NewStatsComponent()
	stats.CritChance = 0.05
	player.AddComponent(stats)
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	// Expected: 0.05 + 0.07*0.75 = 0.05 + 0.0525 = 0.1025
	expected := 0.1025
	if stats.CritChance < expected-0.001 || stats.CritChance > expected+0.001 {
		t.Errorf("CritChance = %f, expected around %f (horror genre)", stats.CritChance, expected)
	}
}

func TestReputationCriticalChanceBonusSystem_Update_HostileReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	bandits := &Faction{ID: "bandits", Name: "Bandits", Enemies: map[string]bool{}}
	factionSys.AddFaction(bandits)
	factionSys.SetPlayerReputation("bandits", -50) // Hostile
	sys.SetFactionSystem(factionSys)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	stats := NewStatsComponent()
	stats.CritChance = 0.10
	player.AddComponent(stats)
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	if stats.CritChance != 0.10 {
		t.Errorf("CritChance should not change for hostile rep: %f", stats.CritChance)
	}
}

func TestReputationCriticalChanceBonusSystem_Update_SkipsNonPlayer(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100)
	sys.SetFactionSystem(factionSys)

	npc := NewEntity(0)
	stats := NewStatsComponent()
	stats.CritChance = 0.05
	npc.AddComponent(stats)
	world.AddEntity(npc)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{npc}, 0.016)

	if stats.CritChance != 0.05 {
		t.Errorf("NPC CritChance should not change: %f", stats.CritChance)
	}
}

func TestReputationCriticalChanceBonusSystem_Update_RespectInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100)
	sys.SetFactionSystem(factionSys)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	stats := NewStatsComponent()
	stats.CritChance = 0.05
	player.AddComponent(stats)
	world.AddEntity(player)

	// Should not update - interval not reached
	sys.Update([]*Entity{player}, 0.016)

	if stats.CritChance != 0.05 {
		t.Errorf("Should not update before interval: CritChance=%f", stats.CritChance)
	}
}

func TestReputationCriticalChanceBonusSystem_ClampsToMax(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100)
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	stats := NewStatsComponent()
	stats.CritChance = 0.97 // Near max
	player.AddComponent(stats)
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	if stats.CritChance > 1.0 {
		t.Errorf("CritChance exceeded 1.0: %f", stats.CritChance)
	}
}

func TestReputationCriticalChanceBonusSystem_GetCritBonus_Default(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)

	bonus := sys.GetCritBonus(999)
	if bonus != 0.0 {
		t.Errorf("GetCritBonus for unknown entity = %f, want 0.0", bonus)
	}
}

func TestReputationCriticalChanceBonusSystem_MultipleBestFaction(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	merchants := &Faction{ID: "merchants", Name: "Merchants", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.AddFaction(merchants)
	factionSys.SetPlayerReputation("guards", 30)    // Neutral: 2%
	factionSys.SetPlayerReputation("merchants", 80) // Honored: 7%
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	stats := NewStatsComponent()
	stats.CritChance = 0.05
	player.AddComponent(stats)
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	// Should use best reputation (merchants @ 80 => 7% bonus)
	expected := 0.12
	if stats.CritChance < expected-0.001 || stats.CritChance > expected+0.001 {
		t.Errorf("CritChance = %f, expected around %f (best faction)", stats.CritChance, expected)
	}
}

func BenchmarkReputationCriticalChanceBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewReputationCriticalChanceBonusSystem(world, 12345)
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

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(NewStatsComponent())
	world.AddEntity(player)

	entities := []*Entity{player}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 1.0
		sys.Update(entities, 0.016)
	}
}
