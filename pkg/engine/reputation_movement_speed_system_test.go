//go:build ignore

package engine

import (
	"testing"
)

func TestNewReputationMovementSpeedSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewReputationMovementSpeedSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.appliedMultipliers == nil {
		t.Error("appliedMultipliers map not initialized")
	}
	if sys.speedBonuses == nil {
		t.Error("speedBonuses map not initialized")
	}
	if len(sys.genreMultipliers) == 0 {
		t.Error("genreMultipliers not initialized")
	}
}

func TestReputationMovementSpeedSystem_SetFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	sys.SetFactionSystem(factionSys)

	if sys.factionSystem != factionSys {
		t.Error("FactionSystem not set correctly")
	}
}

func TestReputationMovementSpeedSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)

	sys.SetGenre("cyberpunk")

	if sys.genreID != "cyberpunk" {
		t.Errorf("GenreID = %s, want cyberpunk", sys.genreID)
	}
}

func TestReputationMovementSpeedSystem_BonusForReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)

	tests := []struct {
		name       string
		reputation int
		wantMin    float64
		wantMax    float64
	}{
		{"hostile", -100, 0.0, 0.0},
		{"hostile_edge", -50, 0.0, 0.0},
		{"suspicious", -25, 0.0, 0.0},
		{"zero", 0, 0.0, 0.0},
		{"neutral_low", 1, 0.03, 0.03},
		{"neutral_mid", 25, 0.03, 0.03},
		{"neutral_high", 50, 0.03, 0.03},
		{"friendly_low", 51, 0.06, 0.06},
		{"friendly_mid", 65, 0.06, 0.06},
		{"friendly_high", 75, 0.06, 0.06},
		{"honored_low", 76, 0.10, 0.10},
		{"honored_max", 100, 0.10, 0.10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.bonusForReputation(tt.reputation)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("bonusForReputation(%d) = %f, want between %f and %f",
					tt.reputation, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestReputationMovementSpeedSystem_GenreMultipliers(t *testing.T) {
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
			sys := NewReputationMovementSpeedSystem(world, 12345)
			sys.SetGenre(tt.genre)

			mult := sys.GetGenreMultiplier()
			if mult < tt.wantMin || mult > tt.wantMax {
				t.Errorf("GetGenreMultiplier() = %f, want between %f and %f",
					mult, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestReputationMovementSpeedSystem_UnknownGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	sys.SetGenre("steampunk")

	mult := sys.GetGenreMultiplier()
	if mult != 1.0 {
		t.Errorf("GetGenreMultiplier for unknown genre = %f, want 1.0", mult)
	}
}

func TestReputationMovementSpeedSystem_Update_NoFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&VelocityComponent{VX: 10.0, VY: 5.0})
	world.AddEntity(player)

	// Force past update interval
	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	vel := player.GetVelocity()
	if vel.VX != 10.0 || vel.VY != 5.0 {
		t.Errorf("Velocity changed without FactionSystem: VX=%f VY=%f", vel.VX, vel.VY)
	}
}

func TestReputationMovementSpeedSystem_Update_NeutralReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 25) // Neutral
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 50.0})
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	vel := player.GetVelocity()
	// Expected: 1.0 + 0.03*1.0 = 1.03 multiplier
	expectedVX := 100.0 * 1.03
	if vel.VX < expectedVX-0.5 || vel.VX > expectedVX+0.5 {
		t.Errorf("VX = %f, expected around %f", vel.VX, expectedVX)
	}
}

func TestReputationMovementSpeedSystem_Update_HonoredReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100) // Honored
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 0.0})
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	vel := player.GetVelocity()
	// Expected: 1.0 + 0.10*1.0 = 1.10 multiplier
	expectedVX := 100.0 * 1.10
	if vel.VX < expectedVX-0.5 || vel.VX > expectedVX+0.5 {
		t.Errorf("VX = %f, expected around %f", vel.VX, expectedVX)
	}

	bonus := sys.GetSpeedBonus(player.ID)
	if bonus < 9.0 || bonus > 11.0 {
		t.Errorf("GetSpeedBonus = %f, expected around 10.0", bonus)
	}
}

func TestReputationMovementSpeedSystem_Update_HorrorGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	cult := &Faction{ID: "cult", Name: "The Cult", Enemies: map[string]bool{}}
	factionSys.AddFaction(cult)
	factionSys.SetPlayerReputation("cult", 100) // Honored
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("horror") // 0.75 multiplier

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 0.0})
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	vel := player.GetVelocity()
	// Expected: 1.0 + 0.10*0.75 = 1.075
	expectedVX := 100.0 * 1.075
	if vel.VX < expectedVX-0.5 || vel.VX > expectedVX+0.5 {
		t.Errorf("VX = %f, expected around %f (horror genre)", vel.VX, expectedVX)
	}
}

func TestReputationMovementSpeedSystem_Update_HostileReputation(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	bandits := &Faction{ID: "bandits", Name: "Bandits", Enemies: map[string]bool{}}
	factionSys.AddFaction(bandits)
	factionSys.SetPlayerReputation("bandits", -50) // Hostile
	sys.SetFactionSystem(factionSys)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 50.0})
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	vel := player.GetVelocity()
	if vel.VX != 100.0 || vel.VY != 50.0 {
		t.Errorf("Velocity should not change for hostile rep: VX=%f VY=%f", vel.VX, vel.VY)
	}
}

func TestReputationMovementSpeedSystem_Update_SkipsNonPlayer(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100)
	sys.SetFactionSystem(factionSys)

	npc := NewEntity(0)
	npc.AddComponent(&VelocityComponent{VX: 50.0, VY: 25.0})
	world.AddEntity(npc)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{npc}, 0.016)

	vel := npc.GetVelocity()
	if vel.VX != 50.0 || vel.VY != 25.0 {
		t.Errorf("NPC velocity should not change: VX=%f VY=%f", vel.VX, vel.VY)
	}
}

func TestReputationMovementSpeedSystem_Update_RespectInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.SetPlayerReputation("guards", 100)
	sys.SetFactionSystem(factionSys)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 0.0})
	world.AddEntity(player)

	// Should not update - interval not reached
	sys.Update([]*Entity{player}, 0.016)

	vel := player.GetVelocity()
	if vel.VX != 100.0 {
		t.Errorf("Should not update before interval: VX=%f", vel.VX)
	}
}

func TestReputationMovementSpeedSystem_GetSpeedMultiplier_Default(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)

	mult := sys.GetSpeedMultiplier(999)
	if mult != 1.0 {
		t.Errorf("GetSpeedMultiplier for unknown entity = %f, want 1.0", mult)
	}
}

func TestReputationMovementSpeedSystem_MultipleBestFaction(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	guards := &Faction{ID: "guards", Name: "Guards", Enemies: map[string]bool{}}
	merchants := &Faction{ID: "merchants", Name: "Merchants", Enemies: map[string]bool{}}
	factionSys.AddFaction(guards)
	factionSys.AddFaction(merchants)
	factionSys.SetPlayerReputation("guards", 30)    // Neutral: 3%
	factionSys.SetPlayerReputation("merchants", 80)  // Honored: 10%
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 0.0})
	world.AddEntity(player)

	sys.timeSinceCheck = 1.0
	sys.Update([]*Entity{player}, 0.016)

	vel := player.GetVelocity()
	// Should use best reputation (merchants @ 80 => 10% bonus)
	expectedVX := 100.0 * 1.10
	if vel.VX < expectedVX-0.5 || vel.VX > expectedVX+0.5 {
		t.Errorf("VX = %f, expected around %f (best faction)", vel.VX, expectedVX)
	}
}

func BenchmarkReputationMovementSpeedSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewReputationMovementSpeedSystem(world, 12345)
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
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 50.0})
	world.AddEntity(player)

	entities := []*Entity{player}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 1.0
		sys.Update(entities, 0.016)
	}
}
