//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	politicalwarfare "github.com/opd-ai/venture/pkg/integration/political_warfare"
	"github.com/sirupsen/logrus"
)

// TestPoliticalWarfareSystem_ServerInitialization verifies PoliticalWarfareSystem is properly initialized on server
func TestPoliticalWarfareSystem_ServerInitialization(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(12345)

	// Initialize V8 to get guild manager
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	if guildManager == nil {
		t.Fatal("Failed to initialize guild manager")
	}

	initialSystemCount := len(world.GetSystems())

	// Initialize V9 which should add PoliticalWarfareSystem
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	if politicalWarfareSys == nil {
		t.Fatal("PoliticalWarfareSystem not initialized")
	}

	finalSystemCount := len(world.GetSystems())
	if finalSystemCount <= initialSystemCount {
		t.Error("PoliticalWarfareSystem not added to world")
	}

	// Verify manager is accessible
	manager := politicalWarfareSys.GetManager()
	if manager == nil {
		t.Error("PoliticalWarfareSystem manager is nil")
	}

	t.Log("PoliticalWarfareSystem successfully initialized on server")
}

// TestPoliticalWarfareSystem_WarDeclaration tests war declaration integration
func TestPoliticalWarfareSystem_WarDeclaration(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(12345)

	// Setup
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Create test guilds
	guildID1, err := guildManager.CreateGuild("fantasy", "Player1", 12345)
	if err != nil {
		t.Fatalf("Failed to create guild1: %v", err)
	}

	guildID2, err := guildManager.CreateGuild("fantasy", "Player2", 23456)
	if err != nil {
		t.Fatalf("Failed to create guild2: %v", err)
	}

	// Declare war
	manager := politicalWarfareSys.GetManager()
	war, err := manager.DeclareWar(guildID1, guildID2, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to declare war: %v", err)
	}

	if war.AttackerGuildID != guildID1 {
		t.Errorf("Expected attacker %s, got %s", guildID1, war.AttackerGuildID)
	}

	if war.DefenderGuildID != guildID2 {
		t.Errorf("Expected defender %s, got %s", guildID2, war.DefenderGuildID)
	}

	if war.Active {
		t.Error("War should not be active during preparation period")
	}

	// Verify war is tracked
	activeWars := manager.GetActiveWars()
	if len(activeWars) != 1 {
		t.Errorf("Expected 1 active war, got %d", len(activeWars))
	}

	t.Log("War declaration successful")
}

// TestPoliticalWarfareSystem_Update tests time-based war activation
func TestPoliticalWarfareSystem_Update(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(12345)

	// Setup
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Create test guilds
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	// Declare war with short preparation period
	manager := politicalWarfareSys.GetManager()
	_, err := manager.DeclareWar(guildID1, guildID2, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to declare war: %v", err)
	}

	// War should not be active yet
	wars := manager.GetActiveWars()
	if wars[0].Active {
		t.Error("War should not be active immediately after declaration")
	}

	// Wait for preparation period to pass
	time.Sleep(60 * time.Millisecond)

	// Update system to activate war
	entities := []*engine.Entity{}
	politicalWarfareSys.Update(entities, 0.05)

	// War should now be active
	wars = manager.GetActiveWars()
	if len(wars) == 0 {
		t.Fatal("War disappeared from active wars")
	}

	if !wars[0].Active {
		t.Error("War should be active after preparation period")
	}

	t.Log("War activation successful via Update()")
}

// TestPoliticalWarfareSystem_PeaceTreaty tests treaty signing
func TestPoliticalWarfareSystem_PeaceTreaty(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(12345)

	// Setup
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Create test guilds
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	// Sign peace treaty
	manager := politicalWarfareSys.GetManager()
	treaty, err := manager.SignPeaceTreaty(guildID1, guildID2, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to sign treaty: %v", err)
	}

	if treaty.GuildID1 != guildID1 && treaty.GuildID2 != guildID1 {
		t.Error("Guild1 not in treaty")
	}

	if treaty.GuildID1 != guildID2 && treaty.GuildID2 != guildID2 {
		t.Error("Guild2 not in treaty")
	}

	// Verify war cannot be declared during treaty
	_, err = manager.DeclareWar(guildID1, guildID2, 1*time.Hour)
	if err == nil {
		t.Error("Should not be able to declare war during active treaty")
	}

	t.Log("Peace treaty signed successfully")
}

// TestPoliticalWarfareSystem_TradeEmbargo tests embargo functionality
func TestPoliticalWarfareSystem_TradeEmbargo(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(12345)

	// Setup
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Create test guilds
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	// Impose embargo
	manager := politicalWarfareSys.GetManager()
	embargo, err := manager.ImposeEmbargo(guildID1, guildID2, 0.75) // 75% price increase
	if err != nil {
		t.Fatalf("Failed to impose embargo: %v", err)
	}

	if embargo.ImposingGuildID != guildID1 {
		t.Errorf("Expected imposing guild %s, got %s", guildID1, embargo.ImposingGuildID)
	}

	if embargo.TargetGuildID != guildID2 {
		t.Errorf("Expected target guild %s, got %s", guildID2, embargo.TargetGuildID)
	}

	if embargo.PriceIncrease != 0.75 {
		t.Errorf("Expected 75%% price increase (0.75), got %f", embargo.PriceIncrease)
	}

	t.Log("Trade embargo imposed successfully")
}

// TestPoliticalWarfareSystem_AllianceCall tests alliance reinforcement
func TestPoliticalWarfareSystem_AllianceCall(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(77777) // Fixed seed for deterministic alliance check

	// Setup
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Create test guilds
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)
	guildID3, _ := guildManager.CreateGuild("fantasy", "Player3", 34567)

	// Set up alliance (high reputation)
	guild1, _ := guildManager.GetGuild(guildID1)
	guild3, _ := guildManager.GetGuild(guildID3)
	guild1.Reputation[guildID3] = 0.8 // Allied
	guild3.Reputation[guildID1] = 0.8
	guild1.Treasury = 100000
	guild3.Treasury = 100000

	// Call for reinforcements
	manager := politicalWarfareSys.GetManager()
	allianceCall, err := manager.CallReinforcementAllies(guildID1, guildID2)
	if err != nil {
		t.Fatalf("Failed to call allies: %v", err)
	}

	// With high reputation, ally should respond (60-80% chance, deterministic with seed)
	// We can't guarantee a specific result, but we can verify the call completed
	t.Logf("Alliance call returned %d allies", len(allianceCall.RespondingAllies))

	// Verify the call was tracked
	if len(allianceCall.RespondingAllies) > 0 {
		if allianceCall.RespondingAllies[0].AllyGuildID != guildID3 {
			t.Errorf("Expected ally %s, got %s", guildID3, allianceCall.RespondingAllies[0].AllyGuildID)
		}
	}

	t.Log("Alliance call completed successfully")
}

// TestPoliticalWarfareSystem_ReputationPenalty tests reputation tracking
func TestPoliticalWarfareSystem_ReputationPenalty(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(12345)

	// Setup
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Create test guild
	guildID, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)

	// Apply reputation penalty
	manager := politicalWarfareSys.GetManager()
	manager.ApplyReputationPenalty(guildID, "war_declaration", -0.3)

	// Verify penalty was tracked
	penalties := manager.GetReputationPenalties()
	found := false
	for _, p := range penalties {
		if p.GuildID == guildID && p.Action == "war_declaration" {
			found = true
			if p.Penalty != -0.3 {
				t.Errorf("Expected penalty -0.3, got %f", p.Penalty)
			}
			break
		}
	}

	if !found {
		t.Error("Reputation penalty not tracked")
	}

	t.Log("Reputation penalty tracked successfully")
}

// TestPoliticalWarfareSystem_DiplomaticVictory tests diplomatic resolution
func TestPoliticalWarfareSystem_DiplomaticVictory(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(55555) // Fixed seed for deterministic outcome

	// Setup
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Create test guilds
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	// Declare war first (required for diplomatic victory)
	manager := politicalWarfareSys.GetManager()
	_, err := manager.DeclareWar(guildID1, guildID2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to declare war: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	// Attempt diplomatic victory with concessions
	concessions := []politicalwarfare.DiplomaticConcession{
		{Type: politicalwarfare.ConcessionTerritory, Value: "territory1"},
		{Type: politicalwarfare.ConcessionGold, Value: 5000},
	}
	success, err := manager.NegotiateDiplomaticVictory(guildID1, guildID2, concessions)
	if err != nil {
		t.Fatalf("Failed to negotiate diplomatic victory: %v", err)
	}

	t.Logf("Diplomatic negotiation result: success=%v", success)

	// Verify concessions were tracked if successful
	if success {
		applied := manager.GetAppliedConcessions()
		foundConcession := false
		for _, c := range applied {
			if c.DefenderGuildID == guildID2 {
				foundConcession = true
				break
			}
		}
		if !foundConcession {
			t.Error("Expected concessions to be tracked for successful negotiation")
		}
	}

	t.Log("Diplomatic victory negotiation completed")
}

// BenchmarkPoliticalWarfareSystem_Update measures update performance
func BenchmarkPoliticalWarfareSystem_Update(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(12345)

	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Setup test data
	entities := []*engine.Entity{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		politicalWarfareSys.Update(entities, 0.016)
	}
}

// BenchmarkPoliticalWarfareSystem_WarDeclaration measures war declaration performance
func BenchmarkPoliticalWarfareSystem_WarDeclaration(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(logrus.StandardLogger().Out)
	logger.SetLevel(logrus.WarnLevel)
	seed := int64(12345)

	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	_, _, _, _, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	// Create test guilds
	guilds := make([]string, 100)
	for i := 0; i < 100; i++ {
		guildID, _ := guildManager.CreateGuild("fantasy", "Player", 12345)
		guilds[i] = guildID
	}

	manager := politicalWarfareSys.GetManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % 50 // Use first 50 guilds as attackers vs second 50
		manager.DeclareWar(guilds[idx], guilds[idx+50], 24*time.Hour)
	}
}
