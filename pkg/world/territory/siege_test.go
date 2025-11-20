package territory

import (
	"testing"
	"time"
)

func TestSiegePhaseString(t *testing.T) {
	tests := []struct {
		phase    SiegePhase
		expected string
	}{
		{PhasePreparation, "Preparation"},
		{PhaseAssault, "Assault"},
		{PhaseResolution, "Resolution"},
		{PhaseEnded, "Ended"},
		{SiegePhase(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.phase.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVictoryConditionString(t *testing.T) {
	tests := []struct {
		condition VictoryCondition
		expected  string
	}{
		{VictoryCapturePoints, "Captured All Points"},
		{VictoryDestroyHall, "Guild Hall Destroyed"},
		{VictoryDefenseTimeout, "Defense Held"},
		{VictorySurrender, "Surrender"},
		{VictoryCondition(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.condition.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewSiege(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	if siege.ID == "" {
		t.Error("Siege ID should not be empty")
	}
	if siege.TerritoryID != "territory1" {
		t.Errorf("TerritoryID = %v, want territory1", siege.TerritoryID)
	}
	if siege.AttackerGuildID != "guild_attack" {
		t.Errorf("AttackerGuildID = %v, want guild_attack", siege.AttackerGuildID)
	}
	if siege.DefenderGuildID != "guild_defend" {
		t.Errorf("DefenderGuildID = %v, want guild_defend", siege.DefenderGuildID)
	}
	if siege.Phase != PhasePreparation {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhasePreparation)
	}
	if siege.DefenderTreasury != 10000 {
		t.Errorf("DefenderTreasury = %v, want 10000", siege.DefenderTreasury)
	}
	if siege.LootPercentage != 0.15 {
		t.Errorf("LootPercentage = %v, want 0.15", siege.LootPercentage)
	}
	if siege.GuildHallHP != 10000.0 {
		t.Errorf("GuildHallHP = %v, want 10000.0", siege.GuildHallHP)
	}
	if siege.TotalControlPoints != 5 {
		t.Errorf("TotalControlPoints = %v, want 5", siege.TotalControlPoints)
	}
}

func TestSiegeJoinAndParticipants(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Test joining as attacker
	err := siege.JoinSiege("player1", true)
	if err != nil {
		t.Fatalf("JoinSiege failed: %v", err)
	}
	if !siege.Attackers["player1"] {
		t.Error("player1 should be in attackers")
	}

	// Test joining as defender
	err = siege.JoinSiege("player2", false)
	if err != nil {
		t.Fatalf("JoinSiege failed: %v", err)
	}
	if !siege.Defenders["player2"] {
		t.Error("player2 should be in defenders")
	}

	// Test duplicate join
	err = siege.JoinSiege("player1", true)
	if err == nil {
		t.Error("Expected error for duplicate join")
	}
}

func TestSiegeParticipantCap(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Fill to capacity (100 players)
	for i := 0; i < 50; i++ {
		siege.Attackers[string(rune(i))] = true
	}
	for i := 0; i < 50; i++ {
		siege.Defenders[string(rune(i+50))] = true
	}

	// Try to add one more
	err := siege.JoinSiege("player101", true)
	if err == nil {
		t.Error("Expected error for exceeding participant cap")
	}
}

func TestSiegeReinforcements(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Add reinforcements
	players := []string{"ally1", "ally2", "ally3"}
	err := siege.AddReinforcements("guild_ally", players)
	if err != nil {
		t.Fatalf("AddReinforcements failed: %v", err)
	}

	// Verify reinforcements added
	if _, exists := siege.Reinforcements["guild_ally"]; !exists {
		t.Error("Reinforcements not added")
	}

	// Verify players added to defenders
	for _, playerID := range players {
		if !siege.Defenders[playerID] {
			t.Errorf("Player %s not added to defenders", playerID)
		}
	}

	// Test duplicate guild
	err = siege.AddReinforcements("guild_ally", []string{"ally4"})
	if err == nil {
		t.Error("Expected error for duplicate guild reinforcements")
	}
}

func TestSiegeReinforcementsCap(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Add 5 guilds
	for i := 0; i < 5; i++ {
		guildID := string(rune('A' + i))
		err := siege.AddReinforcements(guildID, []string{"player1"})
		if err != nil {
			t.Fatalf("AddReinforcements failed for guild %s: %v", guildID, err)
		}
	}

	// Try to add 6th guild
	err := siege.AddReinforcements("guild_F", []string{"player2"})
	if err == nil {
		t.Error("Expected error for exceeding reinforcement guild cap")
	}
}

func TestSiegePhaseAdvancement(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Cannot advance immediately
	err := siege.AdvancePhase()
	if err == nil {
		t.Error("Expected error advancing phase before 1 hour")
	}

	// Simulate 1 hour passing
	siege.PhaseStartTime = time.Now().Add(-1 * time.Hour)

	// Should advance to assault
	err = siege.AdvancePhase()
	if err != nil {
		t.Fatalf("AdvancePhase failed: %v", err)
	}
	if siege.Phase != PhaseAssault {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseAssault)
	}

	// Cannot advance assault immediately
	err = siege.AdvancePhase()
	if err == nil {
		t.Error("Expected error advancing assault phase before 2 hours")
	}

	// Simulate 2 hours passing or victory
	siege.WinnerGuildID = "guild_attack"

	// Should advance to resolution
	err = siege.AdvancePhase()
	if err != nil {
		t.Fatalf("AdvancePhase failed: %v", err)
	}
	if siege.Phase != PhaseResolution {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseResolution)
	}

	// Should advance to ended
	err = siege.AdvancePhase()
	if err != nil {
		t.Fatalf("AdvancePhase failed: %v", err)
	}
	if siege.Phase != PhaseEnded {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseEnded)
	}
}

func TestSiegeCaptureControlPoint(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Cannot capture during preparation
	err := siege.CaptureControlPoint()
	if err == nil {
		t.Error("Expected error capturing during preparation phase")
	}

	// Advance to assault
	siege.Phase = PhaseAssault

	// Capture points
	for i := 0; i < siege.TotalControlPoints-1; i++ {
		err := siege.CaptureControlPoint()
		if err != nil {
			t.Fatalf("CaptureControlPoint failed: %v", err)
		}
	}

	// No winner yet
	if siege.WinnerGuildID != "" {
		t.Error("Winner should not be set until all points captured")
	}

	// Capture final point
	err = siege.CaptureControlPoint()
	if err != nil {
		t.Fatalf("CaptureControlPoint failed: %v", err)
	}

	// Victory condition should be set
	if siege.VictoryCondition != VictoryCapturePoints {
		t.Errorf("VictoryCondition = %v, want %v", siege.VictoryCondition, VictoryCapturePoints)
	}
	if siege.WinnerGuildID != "guild_attack" {
		t.Errorf("WinnerGuildID = %v, want guild_attack", siege.WinnerGuildID)
	}
}

func TestSiegeDamageGuildHall(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Cannot damage during preparation
	err := siege.DamageGuildHall(100)
	if err == nil {
		t.Error("Expected error damaging hall during preparation phase")
	}

	// Advance to assault
	siege.Phase = PhaseAssault

	// Damage hall
	err = siege.DamageGuildHall(5000)
	if err != nil {
		t.Fatalf("DamageGuildHall failed: %v", err)
	}
	if siege.GuildHallHP != 5000 {
		t.Errorf("GuildHallHP = %v, want 5000", siege.GuildHallHP)
	}

	// No winner yet
	if siege.WinnerGuildID != "" {
		t.Error("Winner should not be set until hall destroyed")
	}

	// Destroy hall
	err = siege.DamageGuildHall(6000)
	if err != nil {
		t.Fatalf("DamageGuildHall failed: %v", err)
	}
	if siege.GuildHallHP != 0 {
		t.Errorf("GuildHallHP = %v, want 0", siege.GuildHallHP)
	}

	// Victory condition should be set
	if siege.VictoryCondition != VictoryDestroyHall {
		t.Errorf("VictoryCondition = %v, want %v", siege.VictoryCondition, VictoryDestroyHall)
	}
	if siege.WinnerGuildID != "guild_attack" {
		t.Errorf("WinnerGuildID = %v, want guild_attack", siege.WinnerGuildID)
	}
}

func TestSiegeDistributeLoot(t *testing.T) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Cannot distribute during preparation
	_, err := siege.DistributeLoot()
	if err == nil {
		t.Error("Expected error distributing loot before resolution phase")
	}

	// Advance to resolution with winner
	siege.Phase = PhaseResolution
	siege.WinnerGuildID = "guild_attack"

	// Distribute loot
	loot, err := siege.DistributeLoot()
	if err != nil {
		t.Fatalf("DistributeLoot failed: %v", err)
	}

	// Verify loot amount (15% of 10000 = 1500)
	expected := int(float64(10000) * 0.15)
	if loot != expected {
		t.Errorf("Loot = %v, want %v", loot, expected)
	}

	// Cannot distribute twice
	_, err = siege.DistributeLoot()
	if err == nil {
		t.Error("Expected error distributing loot twice")
	}
}

func TestSiegeManagerCreate(t *testing.T) {
	sm := NewSiegeManager()

	siege, err := sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)
	if err != nil {
		t.Fatalf("CreateSiege failed: %v", err)
	}
	if siege == nil {
		t.Fatal("Siege should not be nil")
	}

	// Cannot create duplicate siege for same territory
	_, err = sm.CreateSiege("territory1", "guild_attack2", "guild_defend", 10000)
	if err == nil {
		t.Error("Expected error creating duplicate siege for same territory")
	}
}

func TestSiegeManagerGetSiege(t *testing.T) {
	sm := NewSiegeManager()

	created, _ := sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Get existing siege
	siege, err := sm.GetSiege(created.ID)
	if err != nil {
		t.Fatalf("GetSiege failed: %v", err)
	}
	if siege.ID != created.ID {
		t.Errorf("Siege ID = %v, want %v", siege.ID, created.ID)
	}

	// Get non-existent siege
	_, err = sm.GetSiege("nonexistent")
	if err == nil {
		t.Error("Expected error getting non-existent siege")
	}
}

func TestSiegeManagerGetActiveSieges(t *testing.T) {
	sm := NewSiegeManager()

	// Create 3 sieges
	sm.CreateSiege("territory1", "guild1", "guild2", 10000)
	sm.CreateSiege("territory2", "guild3", "guild4", 10000)
	siege3, _ := sm.CreateSiege("territory3", "guild5", "guild6", 10000)

	// End one siege
	siege3.Phase = PhaseEnded

	// Get active sieges (should be 2)
	active := sm.GetActiveSieges()
	if len(active) != 2 {
		t.Errorf("Active sieges = %v, want 2", len(active))
	}
}

func TestSiegeManagerGetSiegeForTerritory(t *testing.T) {
	sm := NewSiegeManager()

	created, _ := sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Get siege for territory
	siege, found := sm.GetSiegeForTerritory("territory1")
	if !found {
		t.Fatal("Siege not found for territory1")
	}
	if siege.ID != created.ID {
		t.Errorf("Siege ID = %v, want %v", siege.ID, created.ID)
	}

	// Get siege for territory without siege
	_, found = sm.GetSiegeForTerritory("territory2")
	if found {
		t.Error("Should not find siege for territory2")
	}
}

func TestSiegeManagerUpdate(t *testing.T) {
	sm := NewSiegeManager()

	siege, _ := sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Simulate 1 hour passing
	siege.PhaseStartTime = time.Now().Add(-61 * time.Minute)

	// Update should advance to assault
	sm.Update(0.016)

	if siege.Phase != PhaseAssault {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseAssault)
	}

	// Simulate 2 hours passing in assault
	siege.PhaseStartTime = time.Now().Add(-121 * time.Minute)

	// Update should set defender victory and advance to resolution
	sm.Update(0.016)

	if siege.VictoryCondition != VictoryDefenseTimeout {
		t.Errorf("VictoryCondition = %v, want %v", siege.VictoryCondition, VictoryDefenseTimeout)
	}
	if siege.WinnerGuildID != "guild_defend" {
		t.Errorf("WinnerGuildID = %v, want guild_defend", siege.WinnerGuildID)
	}
	if siege.Phase != PhaseResolution {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseResolution)
	}
}

func TestGenerateDefensiveStructures(t *testing.T) {
	structures := GenerateDefensiveStructures("territory1", 12345, 10)

	// Verify count
	if len(structures) != 10 {
		t.Errorf("Structure count = %v, want 10", len(structures))
	}

	// Verify each structure has valid values
	for i, s := range structures {
		if s.ID == "" {
			t.Errorf("Structure %d has empty ID", i)
		}
		if s.HP <= 0 || s.HP > 5000 {
			t.Errorf("Structure %d HP = %v, want 0-5000", i, s.HP)
		}
		if s.Damage <= 0 || s.Damage > 200 {
			t.Errorf("Structure %d Damage = %v, want 0-200", i, s.Damage)
		}
		if s.Level < 1 || s.Level > 5 {
			t.Errorf("Structure %d Level = %v, want 1-5", i, s.Level)
		}
	}
}

func TestGenerateDefensiveStructuresDeterminism(t *testing.T) {
	structures1 := GenerateDefensiveStructures("territory1", 12345, 10)
	structures2 := GenerateDefensiveStructures("territory1", 12345, 10)

	// Verify determinism
	for i := range structures1 {
		if structures1[i].Type != structures2[i].Type {
			t.Errorf("Structure %d Type mismatch", i)
		}
		if structures1[i].HP != structures2[i].HP {
			t.Errorf("Structure %d HP mismatch", i)
		}
		if structures1[i].Damage != structures2[i].Damage {
			t.Errorf("Structure %d Damage mismatch", i)
		}
	}
}

func TestGenerateDefensiveStructuresCount(t *testing.T) {
	// Test minimum count (should clamp to 5)
	structures := GenerateDefensiveStructures("territory1", 12345, 2)
	if len(structures) != 5 {
		t.Errorf("Structure count = %v, want 5 (minimum)", len(structures))
	}

	// Test maximum count (should clamp to 15)
	structures = GenerateDefensiveStructures("territory1", 12345, 20)
	if len(structures) != 15 {
		t.Errorf("Structure count = %v, want 15 (maximum)", len(structures))
	}
}

// Benchmark tests
func BenchmarkSiegeCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewSiege("territory1", "guild_attack", "guild_defend", 10000)
	}
}

func BenchmarkSiegeJoin(b *testing.B) {
	siege := NewSiege("territory1", "guild_attack", "guild_defend", 10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		playerID := string(rune(i % 100))
		siege.JoinSiege(playerID, i%2 == 0)
	}
}

func BenchmarkSiegeManagerUpdate(b *testing.B) {
	sm := NewSiegeManager()
	for i := 0; i < 10; i++ {
		sm.CreateSiege(string(rune(i)), "guild_attack", "guild_defend", 10000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.Update(0.016)
	}
}

func BenchmarkGenerateDefensiveStructures(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateDefensiveStructures("territory1", int64(i), 10)
	}
}
