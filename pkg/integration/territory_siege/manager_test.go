package territory_siege

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world"
)

// TestNewSiegeManager tests siege manager creation.
func TestNewSiegeManager(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	sm := NewSiegeManager(w, tm, ps, gm)

	if sm == nil {
		t.Fatal("NewSiegeManager() returned nil")
	}

	if sm.world != w {
		t.Error("World not set correctly")
	}

	if sm.territoryManager != tm {
		t.Error("TerritoryManager not set correctly")
	}

	if sm.politicsSystem != ps {
		t.Error("PoliticsSystem not set correctly")
	}

	if sm.guildManager != gm {
		t.Error("GuildManager not set correctly")
	}

	if len(sm.activeSieges) != 0 {
		t.Errorf("activeSieges should be empty, got %d", len(sm.activeSieges))
	}

	if len(sm.completedSieges) != 0 {
		t.Errorf("completedSieges should be empty, got %d", len(sm.completedSieges))
	}
}

// TestDeclareSiege_Basic tests basic siege declaration.
func TestDeclareSiege_Basic(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()
	sm := NewSiegeManager(w, tm, ps, gm)

	// Create guilds
	attackerID, _ := gm.CreateGuild("fantasy", "attacker_leader")
	defenderID, _ := gm.CreateGuild("fantasy", "defender_leader")

	// Create zone owned by defender
	zone := tm.CreateBorderZone("zone1", "server1", "server2", 3)
	zone.OwnerFaction = defenderID

	// Declare siege
	siege, err := sm.DeclareSiege(attackerID, defenderID, "zone1")
	if err != nil {
		t.Fatalf("DeclareSiege() error = %v", err)
	}

	if siege == nil {
		t.Fatal("DeclareSiege() returned nil")
	}

	if siege.AttackerGuildID != attackerID {
		t.Errorf("AttackerGuildID = %s, want %s", siege.AttackerGuildID, attackerID)
	}

	if siege.DefenderGuildID != defenderID {
		t.Errorf("DefenderGuildID = %s, want %s", siege.DefenderGuildID, defenderID)
	}

	if siege.CurrentPhase != PhasePreparation {
		t.Errorf("CurrentPhase = %s, want Preparation", siege.CurrentPhase.String())
	}

	if len(siege.DefensiveStructures) < 5 || len(siege.DefensiveStructures) > 15 {
		t.Errorf("DefensiveStructures count = %d, want 5-15", len(siege.DefensiveStructures))
	}
}

// TestDeclareSiege_InvalidZone tests siege with invalid zone.
func TestDeclareSiege_InvalidZone(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()
	sm := NewSiegeManager(w, tm, ps, gm)

	attackerID, _ := gm.CreateGuild("fantasy", "attacker")
	defenderID, _ := gm.CreateGuild("fantasy", "defender")

	_, err := sm.DeclareSiege(attackerID, defenderID, "nonexistent")
	if err == nil {
		t.Error("DeclareSiege() with invalid zone should error")
	}
}

// TestUpdate_PhaseAdvancement tests phase transitions.
func TestUpdate_PhaseAdvancement(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()
	sm := NewSiegeManager(w, tm, ps, gm)

	attackerID, _ := gm.CreateGuild("fantasy", "attacker")
	defenderID, _ := gm.CreateGuild("fantasy", "defender")

	zone := tm.CreateBorderZone("zone1", "server1", "server2", 3)
	zone.OwnerFaction = defenderID

	siege, _ := sm.DeclareSiege(attackerID, defenderID, "zone1")

	// Force preparation phase to expire
	siege.PhaseStartTime = time.Now().Unix() - 3601

	sm.Update(0.016)

	if siege.CurrentPhase != PhaseAssault {
		t.Errorf("Phase after update = %s, want Assault", siege.CurrentPhase.String())
	}

	// Force assault phase to expire
	siege.PhaseStartTime = time.Now().Unix() - 7201

	sm.Update(0.016)

	if siege.CurrentPhase != PhaseResolution {
		t.Errorf("Phase after update = %s, want Resolution", siege.CurrentPhase.String())
	}

	if siege.Victor != defenderID {
		t.Errorf("Victor = %s, want %s", siege.Victor, defenderID)
	}
}

// TestDamageStructure tests applying damage to structures.
func TestDamageStructure(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()
	sm := NewSiegeManager(w, tm, ps, gm)

	attackerID, _ := gm.CreateGuild("fantasy", "attacker")
	defenderID, _ := gm.CreateGuild("fantasy", "defender")

	zone := tm.CreateBorderZone("zone1", "server1", "server2", 3)
	zone.OwnerFaction = defenderID

	siege, _ := sm.DeclareSiege(attackerID, defenderID, "zone1")
	siege.CurrentPhase = PhaseAssault

	structure := siege.DefensiveStructures[0]
	initialHP := structure.CurrentHP

	err := sm.DamageStructure(siege.SiegeID, structure.StructureID, 100)
	if err != nil {
		t.Fatalf("DamageStructure() error = %v", err)
	}

	if structure.CurrentHP != initialHP-100 {
		t.Errorf("Structure HP = %d, want %d", structure.CurrentHP, initialHP-100)
	}
}

// TestUpdatePlayerCounts tests player count tracking.
func TestUpdatePlayerCounts(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()
	sm := NewSiegeManager(w, tm, ps, gm)

	attackerID, _ := gm.CreateGuild("fantasy", "attacker")
	defenderID, _ := gm.CreateGuild("fantasy", "defender")

	zone := tm.CreateBorderZone("zone1", "server1", "server2", 3)
	zone.OwnerFaction = defenderID

	siege, _ := sm.DeclareSiege(attackerID, defenderID, "zone1")

	err := sm.UpdatePlayerCounts(siege.SiegeID, 15, 20)
	if err != nil {
		t.Fatalf("UpdatePlayerCounts() error = %v", err)
	}

	if siege.AttackerPlayerCount != 15 {
		t.Errorf("AttackerPlayerCount = %d, want 15", siege.AttackerPlayerCount)
	}

	if siege.DefenderPlayerCount != 20 {
		t.Errorf("DefenderPlayerCount = %d, want 20", siege.DefenderPlayerCount)
	}
}

// TestGetActiveSiege tests retrieving active sieges.
func TestGetActiveSiege(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()
	sm := NewSiegeManager(w, tm, ps, gm)

	attackerID, _ := gm.CreateGuild("fantasy", "attacker")
	defenderID, _ := gm.CreateGuild("fantasy", "defender")

	zone := tm.CreateBorderZone("zone1", "server1", "server2", 3)
	zone.OwnerFaction = defenderID

	originalSiege, _ := sm.DeclareSiege(attackerID, defenderID, "zone1")

	siege, err := sm.GetActiveSiege(originalSiege.SiegeID)
	if err != nil {
		t.Fatalf("GetActiveSiege() error = %v", err)
	}

	if siege.SiegeID != originalSiege.SiegeID {
		t.Errorf("Retrieved siege ID = %s, want %s", siege.SiegeID, originalSiege.SiegeID)
	}

	// Try nonexistent siege
	_, err = sm.GetActiveSiege("nonexistent")
	if err == nil {
		t.Error("GetActiveSiege() with invalid ID should error")
	}
}

// BenchmarkDeclareSiege benchmarks siege declaration.
func BenchmarkDeclareSiege(b *testing.B) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	// Pre-create guilds
	attackerID, _ := gm.CreateGuild("fantasy", "attacker")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm := NewSiegeManager(w, tm, ps, gm)
		defenderID, _ := gm.CreateGuild("fantasy", "defender")
		zone := tm.CreateBorderZone("zone"+string(rune(i)), "s1", "s2", 3)
		zone.OwnerFaction = defenderID
		_, _ = sm.DeclareSiege(attackerID, defenderID, zone.ZoneID)
	}
}

// BenchmarkUpdate benchmarks siege update processing.
func BenchmarkUpdate(b *testing.B) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()
	sm := NewSiegeManager(w, tm, ps, gm)

	// Create 10 active sieges
	for i := 0; i < 10; i++ {
		attackerID, _ := gm.CreateGuild("fantasy", "attacker")
		defenderID, _ := gm.CreateGuild("fantasy", "defender")
		zone := tm.CreateBorderZone("zone"+string(rune(i)), "s1", "s2", 3)
		zone.OwnerFaction = defenderID
		_, _ = sm.DeclareSiege(attackerID, defenderID, zone.ZoneID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Update(0.016)
	}
}
