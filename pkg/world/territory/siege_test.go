package territory

import (
	"fmt"
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
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, fixedTime)

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
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, time.Now())

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
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, time.Now())

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
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, time.Now())

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
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, time.Now())

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
	now := time.Now()
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, now)

	// Cannot advance immediately
	err := siege.AdvancePhaseWithTime(now)
	if err == nil {
		t.Error("Expected error advancing phase before 1 hour")
	}

	// Simulate 1 hour passing
	siege.PhaseStartTime = now.Add(-1 * time.Hour)

	// Should advance to assault
	err = siege.AdvancePhaseWithTime(now)
	if err != nil {
		t.Fatalf("AdvancePhase failed: %v", err)
	}
	if siege.Phase != PhaseAssault {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseAssault)
	}

	// Cannot advance assault immediately
	err = siege.AdvancePhaseWithTime(now)
	if err == nil {
		t.Error("Expected error advancing assault phase before 2 hours")
	}

	// Simulate 2 hours passing or victory
	siege.WinnerGuildID = "guild_attack"

	// Should advance to resolution
	err = siege.AdvancePhaseWithTime(now)
	if err != nil {
		t.Fatalf("AdvancePhase failed: %v", err)
	}
	if siege.Phase != PhaseResolution {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseResolution)
	}

	// Should advance to ended
	err = siege.AdvancePhaseWithTime(now)
	if err != nil {
		t.Fatalf("AdvancePhase failed: %v", err)
	}
	if siege.Phase != PhaseEnded {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseEnded)
	}
}

func TestSiegeCaptureControlPoint(t *testing.T) {
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, time.Now())

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
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, time.Now())

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
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, time.Now())

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
	phasesAdvanced, siegesEnded, err := sm.Update(0.016)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if phasesAdvanced != 1 {
		t.Errorf("phasesAdvanced = %d, want 1", phasesAdvanced)
	}
	if siegesEnded != 0 {
		t.Errorf("siegesEnded = %d, want 0", siegesEnded)
	}

	if siege.Phase != PhaseAssault {
		t.Errorf("Phase = %v, want %v", siege.Phase, PhaseAssault)
	}

	// Simulate 2 hours passing in assault
	siege.PhaseStartTime = time.Now().Add(-121 * time.Minute)

	// Update should set defender victory and advance to resolution
	phasesAdvanced, siegesEnded, err = sm.Update(0.016)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if phasesAdvanced != 1 {
		t.Errorf("phasesAdvanced = %d, want 1", phasesAdvanced)
	}
	if siegesEnded != 0 {
		t.Errorf("siegesEnded = %d, want 0", siegesEnded)
	}

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
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	structures := GenerateDefensiveStructuresWithTime("territory1", 12345, 10, fixedTime)

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
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	structures1 := GenerateDefensiveStructuresWithTime("territory1", 12345, 10, fixedTime)
	structures2 := GenerateDefensiveStructuresWithTime("territory1", 12345, 10, fixedTime)

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
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test minimum count (should clamp to 5)
	structures := GenerateDefensiveStructuresWithTime("territory1", 12345, 2, fixedTime)
	if len(structures) != 5 {
		t.Errorf("Structure count = %v, want 5 (minimum)", len(structures))
	}

	// Test maximum count (should clamp to 15)
	structures = GenerateDefensiveStructuresWithTime("territory1", 12345, 20, fixedTime)
	if len(structures) != 15 {
		t.Errorf("Structure count = %v, want 15 (maximum)", len(structures))
	}
}

func TestGetSiege_DefensiveCopy(t *testing.T) {
	sm := NewSiegeManager()
	created, _ := sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Get a copy and mutate it
	copy1, _ := sm.GetSiege(created.ID)
	copy1.Phase = PhaseEnded
	copy1.Attackers["mutated_player"] = true

	// Get another copy and verify internal state was not affected
	copy2, _ := sm.GetSiege(created.ID)
	if copy2.Phase == PhaseEnded {
		t.Error("internal state mutated: Phase should not be Ended")
	}
	if copy2.Attackers["mutated_player"] {
		t.Error("internal state mutated: mutated_player should not be in Attackers")
	}
}

func TestGetActiveSieges_DefensiveCopy(t *testing.T) {
	sm := NewSiegeManager()
	sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Get copies and mutate
	active := sm.GetActiveSieges()
	active[0].Phase = PhaseEnded

	// Verify internal state was not affected
	active2 := sm.GetActiveSieges()
	if len(active2) != 1 {
		t.Errorf("expected 1 active siege after mutating copy, got %d", len(active2))
	}
}

// Benchmark tests
func BenchmarkSiegeCreate(b *testing.B) {
	now := time.Now()
	for i := 0; i < b.N; i++ {
		NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, now)
	}
}

func BenchmarkSiegeJoin(b *testing.B) {
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, time.Now())
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
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		GenerateDefensiveStructuresWithTime("territory1", int64(i), 10, fixedTime)
	}
}

// MockTimeProvider is a mock implementation of TimeProvider for testing.
type MockTimeProvider struct {
	fixedTime time.Time
}

// Now returns the fixed time.
func (m *MockTimeProvider) Now() time.Time {
	return m.fixedTime
}

// SetTime sets the fixed time for testing.
func (m *MockTimeProvider) SetTime(t time.Time) {
	m.fixedTime = t
}

func TestNewManagerWithTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	m := NewManagerWithTimeProvider(tp)

	territory, err := m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !territory.LastUpdate.Equal(fixedTime) {
		t.Errorf("expected LastUpdate %v, got %v", fixedTime, territory.LastUpdate)
	}
}

func TestManagerTimeProviderDeterminism(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tp1 := &MockTimeProvider{fixedTime: fixedTime}
	tp2 := &MockTimeProvider{fixedTime: fixedTime}

	m1 := NewManagerWithTimeProvider(tp1)
	m2 := NewManagerWithTimeProvider(tp2)

	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}

	territory1, _ := m1.CreateTerritory("terr-1", coords)
	territory2, _ := m2.CreateTerritory("terr-1", coords)

	if !territory1.LastUpdate.Equal(territory2.LastUpdate) {
		t.Error("expected deterministic LastUpdate with same TimeProvider")
	}
}

func TestNewSiegeManagerWithTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	sm := NewSiegeManagerWithTimeProvider(tp)

	siege, err := sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !siege.StartTime.Equal(fixedTime) {
		t.Errorf("expected StartTime %v, got %v", fixedTime, siege.StartTime)
	}
	if !siege.PhaseStartTime.Equal(fixedTime) {
		t.Errorf("expected PhaseStartTime %v, got %v", fixedTime, siege.PhaseStartTime)
	}
}

func TestSiegeManagerUpdateWithTimeProvider(t *testing.T) {
	initialTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: initialTime}
	sm := NewSiegeManagerWithTimeProvider(tp)

	siege, _ := sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)

	// Verify initial phase
	if siege.Phase != PhasePreparation {
		t.Errorf("expected initial phase %v, got %v", PhasePreparation, siege.Phase)
	}

	// Advance time by 1 hour and update
	tp.SetTime(initialTime.Add(61 * time.Minute))
	phasesAdvanced, siegesEnded, err := sm.Update(0.016)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if phasesAdvanced != 1 {
		t.Errorf("phasesAdvanced = %d, want 1", phasesAdvanced)
	}
	if siegesEnded != 0 {
		t.Errorf("siegesEnded = %d, want 0", siegesEnded)
	}

	// Should advance to assault
	if siege.Phase != PhaseAssault {
		t.Errorf("expected phase %v, got %v", PhaseAssault, siege.Phase)
	}
}

func TestNewSiegeWithTime(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, fixedTime)

	if !siege.StartTime.Equal(fixedTime) {
		t.Errorf("expected StartTime %v, got %v", fixedTime, siege.StartTime)
	}

	expectedID := "siege_territory1_" + fmt.Sprintf("%d", fixedTime.Unix())
	if siege.ID != expectedID {
		t.Errorf("expected ID %v, got %v", expectedID, siege.ID)
	}
}

func TestAdvancePhaseWithTime(t *testing.T) {
	initialTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	siege := NewSiegeWithTime("territory1", "guild_attack", "guild_defend", 10000, initialTime)

	// Try to advance before 1 hour - should fail
	err := siege.AdvancePhaseWithTime(initialTime.Add(30 * time.Minute))
	if err == nil {
		t.Error("expected error when advancing before 1 hour")
	}

	// Advance after 1 hour - should succeed
	err = siege.AdvancePhaseWithTime(initialTime.Add(61 * time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if siege.Phase != PhaseAssault {
		t.Errorf("expected phase %v, got %v", PhaseAssault, siege.Phase)
	}
}

func TestGenerateDefensiveStructuresWithTime(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	structures := GenerateDefensiveStructuresWithTime("territory1", 12345, 10, fixedTime)

	// Verify count
	if len(structures) != 10 {
		t.Errorf("Structure count = %v, want 10", len(structures))
	}

	// Verify all structures have the fixed construction time
	for i, s := range structures {
		if !s.ConstructedAt.Equal(fixedTime) {
			t.Errorf("Structure %d ConstructedAt = %v, want %v", i, s.ConstructedAt, fixedTime)
		}
	}
}

func TestGenerateDefensiveStructuresWithTimeDeterminism(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	structures1 := GenerateDefensiveStructuresWithTime("territory1", 12345, 10, fixedTime)
	structures2 := GenerateDefensiveStructuresWithTime("territory1", 12345, 10, fixedTime)

	// Verify complete determinism including ConstructedAt
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
		if !structures1[i].ConstructedAt.Equal(structures2[i].ConstructedAt) {
			t.Errorf("Structure %d ConstructedAt mismatch", i)
		}
	}
}

// TestSiegeManagerUpdate_CompleteLifecycle tests the full siege lifecycle with return values.
func TestSiegeManagerUpdate_CompleteLifecycle(t *testing.T) {
	initialTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: initialTime}
	sm := NewSiegeManagerWithTimeProvider(tp)

	siege, err := sm.CreateSiege("territory1", "guild_attack", "guild_defend", 10000)
	if err != nil {
		t.Fatalf("CreateSiege() error = %v", err)
	}

	tests := []struct {
		name            string
		advanceTime     time.Duration
		wantPhasesAdv   int
		wantSiegesEnded int
		wantPhase       SiegePhase
		wantVictoryCond VictoryCondition
		wantWinner      string
	}{
		{
			name:            "no advancement before 1 hour",
			advanceTime:     30 * time.Minute,
			wantPhasesAdv:   0,
			wantSiegesEnded: 0,
			wantPhase:       PhasePreparation,
		},
		{
			name:            "advance to assault after 1 hour",
			advanceTime:     31 * time.Minute,
			wantPhasesAdv:   1,
			wantSiegesEnded: 0,
			wantPhase:       PhaseAssault,
		},
		{
			name:            "no advancement before 2 hours in assault",
			advanceTime:     60 * time.Minute,
			wantPhasesAdv:   0,
			wantSiegesEnded: 0,
			wantPhase:       PhaseAssault,
		},
		{
			name:            "advance to resolution with defender victory",
			advanceTime:     61 * time.Minute,
			wantPhasesAdv:   1,
			wantSiegesEnded: 0,
			wantPhase:       PhaseResolution,
			wantVictoryCond: VictoryDefenseTimeout,
			wantWinner:      "guild_defend",
		},
		{
			name:            "no advancement before 5 minutes in resolution",
			advanceTime:     2 * time.Minute,
			wantPhasesAdv:   0,
			wantSiegesEnded: 0,
			wantPhase:       PhaseResolution,
		},
		{
			name:            "end siege after 5 minutes in resolution",
			advanceTime:     4 * time.Minute,
			wantPhasesAdv:   1,
			wantSiegesEnded: 1,
			wantPhase:       PhaseEnded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp.SetTime(tp.Now().Add(tt.advanceTime))
			phasesAdvanced, siegesEnded, err := sm.Update(0.016)
			if err != nil {
				t.Errorf("Update() error = %v", err)
			}
			if phasesAdvanced != tt.wantPhasesAdv {
				t.Errorf("phasesAdvanced = %d, want %d", phasesAdvanced, tt.wantPhasesAdv)
			}
			if siegesEnded != tt.wantSiegesEnded {
				t.Errorf("siegesEnded = %d, want %d", siegesEnded, tt.wantSiegesEnded)
			}
			if siege.Phase != tt.wantPhase {
				t.Errorf("phase = %v, want %v", siege.Phase, tt.wantPhase)
			}
			if tt.wantVictoryCond != 0 && siege.VictoryCondition != tt.wantVictoryCond {
				t.Errorf("victoryCondition = %v, want %v", siege.VictoryCondition, tt.wantVictoryCond)
			}
			if tt.wantWinner != "" && siege.WinnerGuildID != tt.wantWinner {
				t.Errorf("winnerGuildID = %v, want %v", siege.WinnerGuildID, tt.wantWinner)
			}
		})
	}
}

// TestSiegeManagerUpdate_MultipleSieges tests update with multiple concurrent sieges.
func TestSiegeManagerUpdate_MultipleSieges(t *testing.T) {
	initialTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: initialTime}
	sm := NewSiegeManagerWithTimeProvider(tp)

	// Create 3 sieges at different times
	siege1, _ := sm.CreateSiege("territory1", "guild_a", "guild_b", 10000)
	tp.SetTime(initialTime.Add(10 * time.Minute))
	siege2, _ := sm.CreateSiege("territory2", "guild_c", "guild_d", 10000)
	tp.SetTime(initialTime.Add(20 * time.Minute))
	siege3, _ := sm.CreateSiege("territory3", "guild_e", "guild_f", 10000)

	// Advance time by 1 hour from first siege start
	tp.SetTime(initialTime.Add(61 * time.Minute))
	phasesAdvanced, siegesEnded, err := sm.Update(0.016)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Only siege1 should advance (51 minutes for siege2, 41 for siege3)
	if phasesAdvanced != 1 {
		t.Errorf("phasesAdvanced = %d, want 1", phasesAdvanced)
	}
	if siegesEnded != 0 {
		t.Errorf("siegesEnded = %d, want 0", siegesEnded)
	}
	if siege1.Phase != PhaseAssault {
		t.Errorf("siege1 phase = %v, want %v", siege1.Phase, PhaseAssault)
	}
	if siege2.Phase != PhasePreparation {
		t.Errorf("siege2 phase = %v, want %v", siege2.Phase, PhasePreparation)
	}
	if siege3.Phase != PhasePreparation {
		t.Errorf("siege3 phase = %v, want %v", siege3.Phase, PhasePreparation)
	}

	// Advance another 20 minutes
	tp.SetTime(initialTime.Add(81 * time.Minute))
	phasesAdvanced, siegesEnded, err = sm.Update(0.016)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Now siege2 and siege3 should also advance
	if phasesAdvanced != 2 {
		t.Errorf("phasesAdvanced = %d, want 2", phasesAdvanced)
	}
	if siege2.Phase != PhaseAssault {
		t.Errorf("siege2 phase = %v, want %v", siege2.Phase, PhaseAssault)
	}
	if siege3.Phase != PhaseAssault {
		t.Errorf("siege3 phase = %v, want %v", siege3.Phase, PhaseAssault)
	}
}
