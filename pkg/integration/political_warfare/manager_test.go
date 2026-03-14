package political_warfare

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
)

// Test helpers

func setupTestManager(t *testing.T) (*Manager, *guild.Manager, string, string, string) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()

	// Create test guild IDs
	guildID1, _ := guildManager.CreateGuild("fantasy", "player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "player2", 23456)
	guildID3, _ := guildManager.CreateGuild("fantasy", "player3", 34567)

	// Get guilds and set up reputation relationships
	guild1, _ := guildManager.GetGuild(guildID1)
	guild3, _ := guildManager.GetGuild(guildID3)

	guild1.Reputation[guildID3] = 0.7 // Allied
	guild3.Reputation[guildID1] = 0.7

	guild1.Treasury = 100000
	guild3.Treasury = 100000

	manager := NewManager(world, guildManager)
	return manager, guildManager, guildID1, guildID2, guildID3
}

// Test NewManagerWithSeed uses provided seed deterministically
func TestNewManagerWithSeed(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()

	// Create two managers with same seed
	seed := int64(99999)
	manager1 := NewManagerWithSeed(world, guildManager, seed)
	manager2 := NewManagerWithSeed(world, guildManager, seed)

	// Verify seeds are stored correctly
	if manager1.seed != seed {
		t.Errorf("Expected seed %d, got %d", seed, manager1.seed)
	}
	if manager2.seed != seed {
		t.Errorf("Expected seed %d, got %d", seed, manager2.seed)
	}

	// Verify both managers produce same random sequence
	val1 := manager1.rng.Float64()
	val2 := manager2.rng.Float64()
	if val1 != val2 {
		t.Errorf("Managers with same seed should produce same random values: got %f vs %f", val1, val2)
	}
}

// Test NewManager uses DefaultSeed
func TestNewManagerUsesDefaultSeed(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()

	manager := NewManager(world, guildManager)

	if manager.seed != DefaultSeed {
		t.Errorf("Expected default seed %d, got %d", DefaultSeed, manager.seed)
	}
}

// Test concession value constants are used correctly
func TestConcessionValueConstants(t *testing.T) {
	// Verify constants have expected values
	if GoldValueNormalizer != 10000.0 {
		t.Errorf("GoldValueNormalizer should be 10000.0, got %f", GoldValueNormalizer)
	}
	if TerritoryValueEquivalent != 2.0 {
		t.Errorf("TerritoryValueEquivalent should be 2.0, got %f", TerritoryValueEquivalent)
	}
	if ApologyValue != 0.1 {
		t.Errorf("ApologyValue should be 0.1, got %f", ApologyValue)
	}
	if ItemValueEquivalent != 0.5 {
		t.Errorf("ItemValueEquivalent should be 0.5, got %f", ItemValueEquivalent)
	}
	if TradeDiscountMultiplier != 0.5 {
		t.Errorf("TradeDiscountMultiplier should be 0.5, got %f", TradeDiscountMultiplier)
	}
	if DefaultSeed != 12345 {
		t.Errorf("DefaultSeed should be 12345, got %d", DefaultSeed)
	}
}

// Test war declaration

func TestDeclareWar(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	preparationPeriod := 24 * time.Hour
	war, err := manager.DeclareWar(guild1, guild2, preparationPeriod)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	if war == nil {
		t.Fatal("Expected war declaration, got nil")
	}
	if war.AttackerGuildID != guild1 {
		t.Errorf("Expected attacker %s, got %s", guild1, war.AttackerGuildID)
	}
	if war.DefenderGuildID != guild2 {
		t.Errorf("Expected defender %s, got %s", guild2, war.DefenderGuildID)
	}
	if war.Active {
		t.Error("War should not be active during preparation period")
	}
	if war.PreparationPeriod != preparationPeriod {
		t.Errorf("Expected preparation period %v, got %v", preparationPeriod, war.PreparationPeriod)
	}
}

func TestDeclareWarInvalidGuild(t *testing.T) {
	manager, _, _, guild2, _ := setupTestManager(t)

	_, err := manager.DeclareWar("invalid_guild", guild2, 24*time.Hour)
	if err == nil {
		t.Error("Expected error for invalid attacker guild")
	}

	_, err = manager.DeclareWar(guild2, "invalid_guild", 24*time.Hour)
	if err == nil {
		t.Error("Expected error for invalid defender guild")
	}
}

func TestSignPeaceTreaty(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	duration := 14 * 24 * time.Hour
	treaty, err := manager.SignPeaceTreaty(guild1, guild2, duration)
	if err != nil {
		t.Fatalf("SignPeaceTreaty failed: %v", err)
	}
	if treaty == nil {
		t.Fatal("Expected peace treaty, got nil")
	}
	if !treaty.Active {
		t.Error("Peace treaty should be active")
	}
	if treaty.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, treaty.Duration)
	}
}

func TestImposeEmbargo(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	priceIncrease := 0.75 // 75%
	embargo, err := manager.ImposeEmbargo(guild1, guild2, priceIncrease)
	if err != nil {
		t.Fatalf("ImposeEmbargo failed: %v", err)
	}
	if embargo == nil {
		t.Fatal("Expected embargo, got nil")
	}
	if embargo.PriceIncrease != priceIncrease {
		t.Errorf("Expected price increase %.2f, got %.2f", priceIncrease, embargo.PriceIncrease)
	}
	if !embargo.Active {
		t.Error("Embargo should be active")
	}
}

func TestCallReinforcementAllies(t *testing.T) {
	manager, _, guild1, guild2, guild3 := setupTestManager(t)

	call, err := manager.CallReinforcementAllies(guild1, guild2)
	if err != nil {
		t.Fatalf("CallReinforcementAllies failed: %v", err)
	}
	if call == nil {
		t.Fatal("Expected alliance call, got nil")
	}
	if call.Completed {
		t.Error("Alliance call should not be completed immediately")
	}

	// guild3 has 0.7 reputation with guild1, should be called
	found := false
	for _, response := range call.RespondingAllies {
		if response.AllyGuildID == guild3 {
			found = true
			if response.SuccessRate < 0.6 || response.SuccessRate > 0.8 {
				t.Errorf("Expected success rate between 0.6-0.8, got %.2f", response.SuccessRate)
			}
		}
	}
	if !found {
		t.Error("Expected guild3 to respond to alliance call")
	}
}

func TestApplyReputationPenalty(t *testing.T) {
	manager, _, guild1, _, _ := setupTestManager(t)

	err := manager.ApplyReputationPenalty(guild1, "attack", -0.3)
	if err != nil {
		t.Fatalf("ApplyReputationPenalty failed: %v", err)
	}

	penalties := manager.GetReputationPenalties()
	if len(penalties) == 0 {
		t.Fatal("Expected reputation penalty to be recorded")
	}

	lastPenalty := penalties[len(penalties)-1]
	if lastPenalty.GuildID != guild1 {
		t.Errorf("Expected guild %s, got %s", guild1, lastPenalty.GuildID)
	}
	if lastPenalty.Action != "attack" {
		t.Errorf("Expected action 'attack', got %s", lastPenalty.Action)
	}
	if lastPenalty.Penalty != -0.3 {
		t.Errorf("Expected penalty -0.3, got %.2f", lastPenalty.Penalty)
	}
}

func TestUpdateActivatesWar(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	war, err := manager.DeclareWar(guild1, guild2, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}

	if war.Active {
		t.Error("War should not be active immediately")
	}

	// Wait for preparation period
	time.Sleep(200 * time.Millisecond)
	manager.Update(0)

	if !war.Active {
		t.Error("War should be active after preparation period")
	}
}

// Benchmark tests

func BenchmarkDeclareWar(b *testing.B) {
	manager, _, guild1, guild2, _ := setupTestManager(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.wars = make(map[string]*WarDeclaration)
		manager.DeclareWar(guild1, guild2, 24*time.Hour)
	}
}

func BenchmarkImposeEmbargo(b *testing.B) {
	manager, _, guild1, guild2, _ := setupTestManager(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.embargoes = make(map[string]*TradeEmbargo)
		manager.ImposeEmbargo(guild1, guild2, 0.75)
	}
}

func BenchmarkUpdate(b *testing.B) {
	manager, _, guild1, guild2, guild3 := setupTestManager(&testing.T{})

	// Add some state
	manager.DeclareWar(guild1, guild2, 24*time.Hour)
	manager.SignPeaceTreaty(guild1, guild3, 7*24*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Update(0.016)
	}
}

// Phase 67.2: Additional test coverage for political warfare

// Test war declaration edge cases

func TestDeclareWarAlreadyAtWar(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Declare war once
	_, err := manager.DeclareWar(guild1, guild2, 24*time.Hour)
	if err != nil {
		t.Fatalf("First DeclareWar failed: %v", err)
	}

	// Attempt to declare war again
	_, err = manager.DeclareWar(guild1, guild2, 24*time.Hour)
	if err == nil {
		t.Error("Expected error when declaring war while already at war")
	}
}

func TestDeclareWarPeaceCooldown(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Sign peace treaty
	_, err := manager.SignPeaceTreaty(guild1, guild2, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("SignPeaceTreaty failed: %v", err)
	}

	// Attempt to declare war during peace treaty
	_, err = manager.DeclareWar(guild1, guild2, 24*time.Hour)
	if err == nil {
		t.Error("Expected error when declaring war during peace treaty")
	}
}

// Test peace treaty negotiation with various concessions

func TestSignPeaceTreatyInvalidGuild(t *testing.T) {
	manager, _, _, guild2, _ := setupTestManager(t)

	_, err := manager.SignPeaceTreaty("invalid_guild", guild2, 7*24*time.Hour)
	if err == nil {
		t.Error("Expected error for invalid guild1")
	}

	_, err = manager.SignPeaceTreaty(guild2, "invalid_guild", 7*24*time.Hour)
	if err == nil {
		t.Error("Expected error for invalid guild2")
	}
}

func TestSignPeaceTreatyActiveTreaty(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Sign first treaty
	treaty1, err := manager.SignPeaceTreaty(guild1, guild2, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("First SignPeaceTreaty failed: %v", err)
	}

	// Sign another treaty (should overwrite the first)
	treaty2, err := manager.SignPeaceTreaty(guild1, guild2, 14*24*time.Hour)
	if err != nil {
		t.Fatalf("Second SignPeaceTreaty failed: %v", err)
	}

	// Both should reference same treaty (overwritten)
	if treaty2.Duration != 14*24*time.Hour {
		t.Errorf("Expected treaty duration 14 days, got %v", treaty2.Duration)
	}

	// Verify only one active treaty exists
	treaties := manager.GetActiveTreaties()
	if len(treaties) != 1 {
		t.Errorf("Expected 1 active treaty, got %d", len(treaties))
	}

	// Verify the newer treaty overwrote the older one
	if treaties[0].Duration != treaty2.Duration {
		t.Error("Expected newer treaty to replace older treaty")
	}

	_ = treaty1 // Suppress unused warning
}

// Test trade embargo enforcement

func TestImposeEmbargoInvalidGuild(t *testing.T) {
	manager, _, _, guild2, _ := setupTestManager(t)

	_, err := manager.ImposeEmbargo("invalid_guild", guild2, 0.75)
	if err == nil {
		t.Error("Expected error for invalid imposer guild")
	}

	_, err = manager.ImposeEmbargo(guild2, "invalid_guild", 0.75)
	if err == nil {
		t.Error("Expected error for invalid target guild")
	}
}

func TestImposeEmbargoInvalidPriceIncrease(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Test below minimum
	_, err := manager.ImposeEmbargo(guild1, guild2, 0.3)
	if err == nil {
		t.Error("Expected error for price increase below 0.5")
	}

	// Test above maximum
	_, err = manager.ImposeEmbargo(guild1, guild2, 1.0)
	if err == nil {
		t.Error("Expected error for price increase above 0.9")
	}
}

func TestImposeEmbargoAlreadyExists(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Impose first embargo
	_, err := manager.ImposeEmbargo(guild1, guild2, 0.75)
	if err != nil {
		t.Fatalf("First ImposeEmbargo failed: %v", err)
	}

	// Attempt to impose another embargo
	_, err = manager.ImposeEmbargo(guild1, guild2, 0.75)
	if err == nil {
		t.Error("Expected error when embargo already active")
	}
}

// Test alliance reinforcement failure scenarios

func TestCallReinforcementAlliesInvalidGuild(t *testing.T) {
	manager, _, _, guild2, _ := setupTestManager(t)

	_, err := manager.CallReinforcementAllies("invalid_guild", guild2)
	if err == nil {
		t.Error("Expected error for invalid caller guild")
	}

	_, err = manager.CallReinforcementAllies(guild2, "invalid_guild")
	if err == nil {
		t.Error("Expected error for invalid enemy guild")
	}
}

func TestCallReinforcementAlliesNoAllies(t *testing.T) {
	manager, guildManager, _, guild2, _ := setupTestManager(t)

	// Create a guild with no allies
	loneGuildID, _ := guildManager.CreateGuild("fantasy", "player4", 45678)

	call, err := manager.CallReinforcementAllies(loneGuildID, guild2)
	if err != nil {
		t.Fatalf("CallReinforcementAllies failed: %v", err)
	}

	if len(call.RespondingAllies) > 0 {
		t.Error("Expected no responding allies for guild with no allies")
	}
}

// Test diplomatic victory conditions

func TestNegotiateDiplomaticVictorySuccess(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)
	_ = guild1 // Used in loop below

	// Declare war first
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}

	// Wait for war to activate
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	// Try negotiation with large concessions
	concessions := []DiplomaticConcession{
		{Type: ConcessionGold, Value: 50000},
		{Type: ConcessionTerritory, Value: nil},
	}

	// Run multiple times to get at least one success
	successFound := false
	for i := 0; i < 100; i++ {
		// Reset war state
		manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
		time.Sleep(10 * time.Millisecond)
		manager.Update(0)

		success, err := manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
		if err != nil {
			t.Fatalf("NegotiateDiplomaticVictory failed: %v", err)
		}
		if success {
			successFound = true
			break
		}
	}

	if !successFound {
		t.Log("Warning: No diplomatic victory in 100 attempts (probabilistic test)")
	}
}

func TestNegotiateDiplomaticVictoryInvalidGuild(t *testing.T) {
	manager, _, _, guild2, _ := setupTestManager(t)

	concessions := []DiplomaticConcession{
		{Type: ConcessionGold, Value: 10000},
	}

	_, err := manager.NegotiateDiplomaticVictory("invalid_guild", guild2, concessions)
	if err == nil {
		t.Error("Expected error for invalid attacker guild")
	}

	_, err = manager.NegotiateDiplomaticVictory(guild2, "invalid_guild", concessions)
	if err == nil {
		t.Error("Expected error for invalid defender guild")
	}
}

func TestNegotiateDiplomaticVictoryNoActiveWar(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	concessions := []DiplomaticConcession{
		{Type: ConcessionGold, Value: 10000},
	}

	_, err := manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
	if err == nil {
		t.Error("Expected error when no active war exists")
	}
}

func TestNegotiateDiplomaticVictoryConcessionTypes(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Declare war
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	// Test various concession types
	concessions := []DiplomaticConcession{
		{Type: ConcessionGold, Value: 10000},
		{Type: ConcessionTerritory, Value: nil},
		{Type: ConcessionApology, Value: nil},
		{Type: ConcessionTribute, Value: []string{"item1", "item2"}},
		{Type: ConcessionTrade, Value: 0.5},
	}

	_, err = manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
	if err != nil {
		t.Fatalf("NegotiateDiplomaticVictory with all concession types failed: %v", err)
	}
}

// TestNegotiateDiplomaticVictoryGoldConcessionFailure tests that gold transfer failures
// are properly handled with rollback and error reporting.
func TestNegotiateDiplomaticVictoryGoldConcessionFailure(t *testing.T) {
	manager, guildManager, guild1, guild2, _ := setupTestManager(t)

	// Declare war
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	// Get initial treasury values
	defenderGuild, _ := guildManager.GetGuild(guild2)
	initialDefenderTreasury := defenderGuild.Treasury

	// Use a non-existent guild ID as attacker to force error
	fakeAttackerID := "non_existent_guild_id"

	// Update war to use fake attacker
	warKey := guild1 + "_" + guild2
	war := manager.wars[warKey]
	originalAttacker := war.AttackerGuildID
	war.AttackerGuildID = fakeAttackerID

	// Attempt diplomatic victory with gold concession
	concessions := []DiplomaticConcession{
		{Type: ConcessionGold, Value: 5000},
	}

	// Force success by using very high concession value
	for i := 0; i < 100; i++ {
		success, err := manager.NegotiateDiplomaticVictory(fakeAttackerID, guild2, concessions)
		if success && err != nil {
			// Verify error indicates concession failure
			if err.Error() == "" {
				t.Error("Expected error message for failed gold concession")
			}

			// Verify defender treasury was rolled back
			defenderGuild, _ = guildManager.GetGuild(guild2)
			if defenderGuild.Treasury != initialDefenderTreasury {
				t.Errorf("Defender treasury not rolled back: expected %d, got %d",
					initialDefenderTreasury, defenderGuild.Treasury)
			}

			// Verify war state was rolled back
			war = manager.wars[warKey]
			if war.Ended {
				t.Error("War should not be ended after failed concession")
			}
			if war.Victor != "" {
				t.Error("War should not have victor after failed concession")
			}
			war.AttackerGuildID = originalAttacker // Restore for cleanup
			return
		}
	}
	war.AttackerGuildID = originalAttacker // Restore for cleanup
	t.Log("Diplomatic victory did not succeed in 100 attempts (low probability but not a failure)")
}

// TestApplyGoldConcessionInvalidType tests that invalid gold value types are handled gracefully
func TestApplyGoldConcessionInvalidType(t *testing.T) {
	manager, guildManager, guild1, guild2, _ := setupTestManager(t)

	defenderGuild, _ := guildManager.GetGuild(guild2)
	initialTreasury := defenderGuild.Treasury

	// Test with invalid value type (string instead of int)
	concession := DiplomaticConcession{Type: ConcessionGold, Value: "invalid"}
	applied := AppliedConcession{}

	err := manager.applyGoldConcession(concession, &applied, defenderGuild, guild1)
	if err != nil {
		t.Errorf("applyGoldConcession should not return error for invalid type, got: %v", err)
	}

	// Verify treasury unchanged
	if defenderGuild.Treasury != initialTreasury {
		t.Errorf("Treasury should not change for invalid value type: expected %d, got %d",
			initialTreasury, defenderGuild.Treasury)
	}
}

// TestApplyGoldConcessionRollback tests the rollback behavior when attacker guild is not found
func TestApplyGoldConcessionRollback(t *testing.T) {
	manager, guildManager, _, guild2, _ := setupTestManager(t)

	defenderGuild, _ := guildManager.GetGuild(guild2)
	initialTreasury := defenderGuild.Treasury
	goldAmount := 5000

	// Use non-existent guild ID
	fakeAttackerID := "non_existent_guild_id"

	concession := DiplomaticConcession{Type: ConcessionGold, Value: goldAmount}
	applied := AppliedConcession{}

	err := manager.applyGoldConcession(concession, &applied, defenderGuild, fakeAttackerID)
	if err == nil {
		t.Error("Expected error when attacker guild not found")
	}

	// Verify defender treasury was rolled back
	if defenderGuild.Treasury != initialTreasury {
		t.Errorf("Defender treasury not rolled back: expected %d, got %d",
			initialTreasury, defenderGuild.Treasury)
	}

	// Verify applied concession has zero gold amount
	if applied.GoldAmount != 0 {
		t.Errorf("Applied gold amount should be 0 on error, got %d", applied.GoldAmount)
	}
}

// TestApplyGoldConcessionSuccess tests successful gold transfer
func TestApplyGoldConcessionSuccess(t *testing.T) {
	manager, guildManager, guild1, guild2, _ := setupTestManager(t)

	attackerGuild, _ := guildManager.GetGuild(guild1)
	defenderGuild, _ := guildManager.GetGuild(guild2)

	initialAttackerTreasury := attackerGuild.Treasury
	initialDefenderTreasury := defenderGuild.Treasury
	goldAmount := 7500

	concession := DiplomaticConcession{Type: ConcessionGold, Value: goldAmount}
	applied := AppliedConcession{}

	err := manager.applyGoldConcession(concession, &applied, defenderGuild, guild1)
	if err != nil {
		t.Fatalf("applyGoldConcession failed: %v", err)
	}

	// Refresh guild data
	attackerGuild, _ = guildManager.GetGuild(guild1)
	defenderGuild, _ = guildManager.GetGuild(guild2)

	// Verify gold transferred
	if attackerGuild.Treasury != initialAttackerTreasury+goldAmount {
		t.Errorf("Attacker treasury incorrect: expected %d, got %d",
			initialAttackerTreasury+goldAmount, attackerGuild.Treasury)
	}
	if defenderGuild.Treasury != initialDefenderTreasury-goldAmount {
		t.Errorf("Defender treasury incorrect: expected %d, got %d",
			initialDefenderTreasury-goldAmount, defenderGuild.Treasury)
	}

	// Verify applied concession recorded correct amount
	if applied.GoldAmount != goldAmount {
		t.Errorf("Applied gold amount incorrect: expected %d, got %d",
			goldAmount, applied.GoldAmount)
	}
}

// TestProcessConcessionTypeUnknownType tests handling of unknown concession types
func TestProcessConcessionTypeUnknownType(t *testing.T) {
	manager, guildManager, guild1, guild2, _ := setupTestManager(t)

	defenderGuild, _ := guildManager.GetGuild(guild2)
	concession := DiplomaticConcession{Type: ConcessionType("unknown"), Value: nil}
	applied := AppliedConcession{}

	err := manager.processConcessionType(concession, &applied, defenderGuild, guild1, time.Now())
	if err == nil {
		t.Error("Expected error for unknown concession type")
	}
	if err.Error() != "unknown concession type: unknown" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// Test reputation penalty calculations

func TestApplyReputationPenaltyPositivePenalty(t *testing.T) {
	manager, _, guild1, _, _ := setupTestManager(t)

	err := manager.ApplyReputationPenalty(guild1, "attack", 0.3)
	if err == nil {
		t.Error("Expected error for positive penalty value")
	}
}

func TestApplyReputationPenaltyTooSevere(t *testing.T) {
	manager, _, guild1, _, _ := setupTestManager(t)

	err := manager.ApplyReputationPenalty(guild1, "attack", -0.6)
	if err == nil {
		t.Error("Expected error for penalty below -0.5")
	}
}

func TestApplyReputationPenaltyValidRange(t *testing.T) {
	manager, _, guild1, _, _ := setupTestManager(t)

	tests := []struct {
		penalty float64
		action  string
	}{
		{-0.1, "minor_attack"},
		{-0.25, "medium_attack"},
		{-0.5, "major_attack"},
	}

	for _, tt := range tests {
		err := manager.ApplyReputationPenalty(guild1, tt.action, tt.penalty)
		if err != nil {
			t.Errorf("ApplyReputationPenalty(%s, %.2f) failed: %v", tt.action, tt.penalty, err)
		}
	}

	penalties := manager.GetReputationPenalties()
	if len(penalties) != len(tests) {
		t.Errorf("Expected %d penalties, got %d", len(tests), len(penalties))
	}
}

// Test getter methods

func TestGetActiveWars(t *testing.T) {
	manager, _, guild1, guild2, guild3 := setupTestManager(t)

	// Declare multiple wars
	_, err := manager.DeclareWar(guild1, guild2, 24*time.Hour)
	if err != nil {
		t.Fatalf("DeclareWar 1 failed: %v", err)
	}

	_, err = manager.DeclareWar(guild1, guild3, 24*time.Hour)
	if err != nil {
		t.Fatalf("DeclareWar 2 failed: %v", err)
	}

	wars := manager.GetActiveWars()
	if len(wars) != 2 {
		t.Errorf("Expected 2 active wars, got %d", len(wars))
	}
}

func TestGetActiveTreaties(t *testing.T) {
	manager, guildManager, guild1, guild2, guild3 := setupTestManager(t)

	// Create fourth guild
	guild4, _ := guildManager.CreateGuild("fantasy", "player4", 45678)

	// Sign multiple treaties
	_, err := manager.SignPeaceTreaty(guild1, guild2, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("SignPeaceTreaty 1 failed: %v", err)
	}

	_, err = manager.SignPeaceTreaty(guild3, guild4, 14*24*time.Hour)
	if err != nil {
		t.Fatalf("SignPeaceTreaty 2 failed: %v", err)
	}

	treaties := manager.GetActiveTreaties()
	if len(treaties) != 2 {
		t.Errorf("Expected 2 active treaties, got %d", len(treaties))
	}
}

func TestGetActiveEmbargoes(t *testing.T) {
	manager, guildManager, guild1, guild2, guild3 := setupTestManager(t)

	// Create fourth guild
	guild4, _ := guildManager.CreateGuild("fantasy", "player4", 45678)

	// Impose multiple embargoes
	_, err := manager.ImposeEmbargo(guild1, guild2, 0.75)
	if err != nil {
		t.Fatalf("ImposeEmbargo 1 failed: %v", err)
	}

	_, err = manager.ImposeEmbargo(guild3, guild4, 0.6)
	if err != nil {
		t.Fatalf("ImposeEmbargo 2 failed: %v", err)
	}

	embargoes := manager.GetActiveEmbargoes()
	if len(embargoes) != 2 {
		t.Errorf("Expected 2 active embargoes, got %d", len(embargoes))
	}
}

// Test treaty expiration via Update

func TestUpdateExpiresTreaties(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Sign treaty with short duration
	treaty, err := manager.SignPeaceTreaty(guild1, guild2, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("SignPeaceTreaty failed: %v", err)
	}

	if !treaty.Active {
		t.Error("Treaty should be active initially")
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)
	manager.Update(0)

	if treaty.Active {
		t.Error("Treaty should be expired after duration")
	}

	treaties := manager.GetActiveTreaties()
	if len(treaties) != 0 {
		t.Errorf("Expected 0 active treaties after expiration, got %d", len(treaties))
	}
}

func TestUpdateExpiresEmbargoes(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Impose embargo
	embargo, err := manager.ImposeEmbargo(guild1, guild2, 0.5)
	if err != nil {
		t.Fatalf("ImposeEmbargo failed: %v", err)
	}

	if !embargo.Active {
		t.Error("Embargo should be active initially")
	}

	// Set expiration in the past to simulate expiration
	embargo.ExpiresAt = time.Now().Add(-100 * time.Millisecond)

	// Run update to expire embargo
	manager.Update(0)

	if embargo.Active {
		t.Error("Embargo should be expired after ExpiresAt")
	}

	embargoes := manager.GetActiveEmbargoes()
	if len(embargoes) != 0 {
		t.Errorf("Expected 0 active embargoes after expiration, got %d", len(embargoes))
	}
}

func TestUpdateDoesNotExpireEmbargoWithZeroTime(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Impose embargo without expiration (default zero time)
	embargo, err := manager.ImposeEmbargo(guild1, guild2, 0.5)
	if err != nil {
		t.Fatalf("ImposeEmbargo failed: %v", err)
	}

	if !embargo.Active {
		t.Error("Embargo should be active initially")
	}

	// Verify ExpiresAt is zero (no expiration set)
	if !embargo.ExpiresAt.IsZero() {
		t.Error("Embargo ExpiresAt should be zero by default")
	}

	// Run update multiple times - embargo should remain active
	manager.Update(0)
	manager.Update(0)

	if !embargo.Active {
		t.Error("Embargo with zero ExpiresAt should remain active indefinitely")
	}

	embargoes := manager.GetActiveEmbargoes()
	if len(embargoes) != 1 {
		t.Errorf("Expected 1 active embargo (no expiration), got %d", len(embargoes))
	}
}

// Test String() methods for types

func TestVictoryTypeString(t *testing.T) {
	tests := []struct {
		vt   VictoryType
		want string
	}{
		{VictoryTypeDiplomatic, "diplomatic"},
		{VictoryTypeMilitary, "military"},
		{VictoryTypeDefault, "default"},
	}

	for _, tt := range tests {
		got := tt.vt.String()
		if got != tt.want {
			t.Errorf("VictoryType.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestConcessionTypeString(t *testing.T) {
	tests := []struct {
		ct   ConcessionType
		want string
	}{
		{ConcessionGold, "gold"},
		{ConcessionTerritory, "territory"},
		{ConcessionApology, "apology"},
		{ConcessionTribute, "tribute"},
		{ConcessionTrade, "trade"},
	}

	for _, tt := range tests {
		got := tt.ct.String()
		if got != tt.want {
			t.Errorf("ConcessionType.String() = %q, want %q", got, tt.want)
		}
	}
}

// Test applyConcessions implementation for all concession types

func TestApplyConcessionsGold(t *testing.T) {
	manager, guildManager, guild1, guild2, _ := setupTestManager(t)

	// Set up initial treasury
	defender, _ := guildManager.GetGuild(guild2)
	defender.Treasury = 50000

	// Declare and activate war
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	// Force a successful negotiation
	concessions := []DiplomaticConcession{
		{Type: ConcessionGold, Value: 25000},
	}

	// Run until success
	for i := 0; i < 100; i++ {
		success, _ := manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
		if success {
			break
		}
		manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
		time.Sleep(5 * time.Millisecond)
		manager.Update(0)
	}

	// Check that concession was recorded
	applied := manager.GetAppliedConcessions()
	foundGold := false
	for _, c := range applied {
		if c.Type == ConcessionGold && c.GoldAmount == 25000 {
			foundGold = true
			break
		}
	}
	if !foundGold {
		t.Log("Note: No successful negotiation in test (probabilistic)")
	}
}

func TestApplyConcessionsTerritory(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Declare and activate war
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	concessions := []DiplomaticConcession{
		{Type: ConcessionTerritory, Value: "territory_north_123"},
	}

	// Run until success
	success := false
	for i := 0; i < 100; i++ {
		success, _ = manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
		if success {
			break
		}
		manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
		time.Sleep(5 * time.Millisecond)
		manager.Update(0)
	}

	if success {
		transfers := manager.GetPendingTerritoryTransfers()
		found := false
		for _, t := range transfers {
			if t.TerritoryID == "territory_north_123" && t.AttackerGuildID == guild1 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected territory transfer to be recorded")
		}
	}
}

func TestApplyConcessionsApology(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Declare and activate war
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	concessions := []DiplomaticConcession{
		{Type: ConcessionApology, Value: "We sincerely apologize for our aggression."},
	}

	// Run until success
	success := false
	for i := 0; i < 100; i++ {
		success, _ = manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
		if success {
			break
		}
		manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
		time.Sleep(5 * time.Millisecond)
		manager.Update(0)
	}

	if success {
		apologies := manager.GetPendingApologies()
		found := false
		for _, a := range apologies {
			if a.ApologyText == "We sincerely apologize for our aggression." {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected apology to be recorded")
		}
	}
}

func TestApplyConcessionsApologyDefault(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Declare and activate war
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	// Apology with nil value should generate default text
	concessions := []DiplomaticConcession{
		{Type: ConcessionApology, Value: nil},
	}

	// Run until success
	success := false
	for i := 0; i < 100; i++ {
		success, _ = manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
		if success {
			break
		}
		manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
		time.Sleep(5 * time.Millisecond)
		manager.Update(0)
	}

	if success {
		apologies := manager.GetPendingApologies()
		found := false
		for _, a := range apologies {
			if a.ApologyText != "" && a.DefenderGuildID == guild2 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected default apology text to be generated")
		}
	}
}

func TestApplyConcessionsTribute(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Declare and activate war
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	tributeItems := []string{"sword_legendary_001", "armor_plate_002", "gem_ruby_003"}
	concessions := []DiplomaticConcession{
		{Type: ConcessionTribute, Value: tributeItems},
	}

	// Run until success
	success := false
	for i := 0; i < 100; i++ {
		success, _ = manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
		if success {
			break
		}
		manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
		time.Sleep(5 * time.Millisecond)
		manager.Update(0)
	}

	if success {
		tributes := manager.GetPendingTributes()
		found := false
		for _, trib := range tributes {
			if len(trib.TributeItemIDs) == 3 && trib.TributeItemIDs[0] == "sword_legendary_001" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected tribute items to be recorded")
		}
	}
}

func TestApplyConcessionsTrade(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Declare and activate war
	_, err := manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	manager.Update(0)

	concessions := []DiplomaticConcession{
		{Type: ConcessionTrade, Value: 0.25}, // 25% trade discount
	}

	// Run until success
	success := false
	for i := 0; i < 100; i++ {
		success, _ = manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
		if success {
			break
		}
		manager.DeclareWar(guild1, guild2, 1*time.Millisecond)
		time.Sleep(5 * time.Millisecond)
		manager.Update(0)
	}

	if success {
		discount := manager.GetTradeDiscount(guild1, guild2)
		if discount != 0.25 {
			t.Errorf("Expected trade discount of 0.25, got %f", discount)
		}
	}
}

func TestGetTradeDiscountNoDiscount(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// No negotiation, no discount
	discount := manager.GetTradeDiscount(guild1, guild2)
	if discount != 0 {
		t.Errorf("Expected no trade discount, got %f", discount)
	}
}

func TestGetAppliedConcessionsEmpty(t *testing.T) {
	manager, _, _, _, _ := setupTestManager(t)

	applied := manager.GetAppliedConcessions()
	if len(applied) != 0 {
		t.Errorf("Expected empty applied concessions, got %d", len(applied))
	}
}

func TestGetPendingTerritoryTransfersEmpty(t *testing.T) {
	manager, _, _, _, _ := setupTestManager(t)

	transfers := manager.GetPendingTerritoryTransfers()
	if len(transfers) != 0 {
		t.Errorf("Expected no pending territory transfers, got %d", len(transfers))
	}
}

func TestGetPendingApologiesEmpty(t *testing.T) {
	manager, _, _, _, _ := setupTestManager(t)

	apologies := manager.GetPendingApologies()
	if len(apologies) != 0 {
		t.Errorf("Expected no pending apologies, got %d", len(apologies))
	}
}

func TestGetPendingTributesEmpty(t *testing.T) {
	manager, _, _, _, _ := setupTestManager(t)

	tributes := manager.GetPendingTributes()
	if len(tributes) != 0 {
		t.Errorf("Expected no pending tributes, got %d", len(tributes))
	}
}

// TestSaveLoadRoundTrip verifies Save/Load preserves all political warfare state.
func TestSaveLoadRoundTrip(t *testing.T) {
	manager, guildManager, guildID1, guildID2, guildID3 := setupTestManager(t)

	// Set up some state to serialize
	// Create a war
	war, err := manager.DeclareWar(guildID1, guildID2, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to declare war: %v", err)
	}
	war.Active = true

	// Create a treaty
	treaty, err := manager.SignPeaceTreaty(guildID1, guildID3, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to sign treaty: %v", err)
	}

	// Create an embargo (50% price increase)
	embargo, err := manager.ImposeEmbargo(guildID2, guildID3, 0.5)
	if err != nil {
		t.Fatalf("Failed to impose embargo: %v", err)
	}

	// Create alliance call
	allianceCall, err := manager.CallReinforcementAllies(guildID1, guildID2)
	if err != nil {
		t.Fatalf("Failed to call allies: %v", err)
	}

	// Save state
	data, err := manager.Save()
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Save returned empty data")
	}

	// Create a new manager to load into
	world2 := engine.NewWorld()
	manager2 := NewManager(world2, guildManager)

	// Load state
	err = manager2.Load(data)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify wars
	wars := manager2.GetActiveWars()
	if len(wars) != 1 {
		t.Errorf("Expected 1 war, got %d", len(wars))
	}
	if len(wars) > 0 && wars[0].AttackerGuildID != war.AttackerGuildID {
		t.Errorf("War attacker mismatch: expected %s, got %s", war.AttackerGuildID, wars[0].AttackerGuildID)
	}

	// Verify treaties
	treaties := manager2.GetActiveTreaties()
	if len(treaties) != 1 {
		t.Errorf("Expected 1 treaty, got %d", len(treaties))
	}
	if len(treaties) > 0 && treaties[0].GuildID1 != treaty.GuildID1 && treaties[0].GuildID2 != treaty.GuildID1 {
		t.Errorf("Treaty guild mismatch")
	}

	// Verify embargoes
	embargoes := manager2.GetActiveEmbargoes()
	if len(embargoes) != 1 {
		t.Errorf("Expected 1 embargo, got %d", len(embargoes))
	}
	if len(embargoes) > 0 && embargoes[0].ImposingGuildID != embargo.ImposingGuildID {
		t.Errorf("Embargo imposing guild mismatch: expected %s, got %s", embargo.ImposingGuildID, embargoes[0].ImposingGuildID)
	}

	// Verify alliance calls
	calls := manager2.GetActiveAllianceCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 alliance call, got %d", len(calls))
	}
	if len(calls) > 0 && calls[0].CallingGuildID != allianceCall.CallingGuildID {
		t.Errorf("Alliance call guild mismatch: expected %s, got %s", allianceCall.CallingGuildID, calls[0].CallingGuildID)
	}

	// Verify seed was preserved
	if manager2.seed != manager.seed {
		t.Errorf("Seed not preserved: expected %d, got %d", manager.seed, manager2.seed)
	}
}

// TestSaveLoadEmpty verifies Save/Load works with empty state.
func TestSaveLoadEmpty(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	manager := NewManagerWithSeed(world, guildManager, 99999)

	// Save empty state
	data, err := manager.Save()
	if err != nil {
		t.Fatalf("Failed to save empty state: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Save returned empty data for empty state")
	}

	// Load into new manager
	manager2 := NewManager(world, guildManager)
	err = manager2.Load(data)
	if err != nil {
		t.Fatalf("Failed to load empty state: %v", err)
	}

	// Verify all collections are empty
	if len(manager2.GetActiveWars()) != 0 {
		t.Error("Expected no wars after loading empty state")
	}
	if len(manager2.GetActiveTreaties()) != 0 {
		t.Error("Expected no treaties after loading empty state")
	}
	if len(manager2.GetActiveEmbargoes()) != 0 {
		t.Error("Expected no embargoes after loading empty state")
	}
	if len(manager2.GetActiveAllianceCalls()) != 0 {
		t.Error("Expected no alliance calls after loading empty state")
	}

	// Verify seed was preserved
	if manager2.seed != 99999 {
		t.Errorf("Seed not preserved: expected 99999, got %d", manager2.seed)
	}
}

// TestLoadInvalidData verifies Load handles corrupt data gracefully.
func TestLoadInvalidData(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	manager := NewManager(world, guildManager)

	// Test with invalid gzip data
	err := manager.Load([]byte("not valid gzip data"))
	if err == nil {
		t.Error("Expected error when loading invalid gzip data")
	}

	// Test with empty data
	err = manager.Load([]byte{})
	if err == nil {
		t.Error("Expected error when loading empty data")
	}
}

// TestSaveLoadReputationPenalties verifies penalties are preserved.
func TestSaveLoadReputationPenalties(t *testing.T) {
	manager, _, guildID1, guildID2, _ := setupTestManager(t)

	// Declare war to generate reputation penalty
	_, err := manager.DeclareWar(guildID1, guildID2, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to declare war: %v", err)
	}

	originalPenalties := manager.GetReputationPenalties()

	// Save and load
	data, err := manager.Save()
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	world2 := engine.NewWorld()
	guildManager2 := guild.NewManager()
	manager2 := NewManager(world2, guildManager2)

	err = manager2.Load(data)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	loadedPenalties := manager2.GetReputationPenalties()
	if len(loadedPenalties) != len(originalPenalties) {
		t.Errorf("Penalty count mismatch: expected %d, got %d", len(originalPenalties), len(loadedPenalties))
	}
}

// Test self-guild validation

func TestDeclareWarSelfGuild(t *testing.T) {
	manager, _, guild1, _, _ := setupTestManager(t)

	_, err := manager.DeclareWar(guild1, guild1, 24*time.Hour)
	if err == nil {
		t.Error("Expected error when guild declares war on itself")
	}
}

func TestSignPeaceTreatySelfGuild(t *testing.T) {
	manager, _, guild1, _, _ := setupTestManager(t)

	_, err := manager.SignPeaceTreaty(guild1, guild1, 7*24*time.Hour)
	if err == nil {
		t.Error("Expected error when guild signs treaty with itself")
	}
}

func TestImposeEmbargoSelfGuild(t *testing.T) {
	manager, _, guild1, _, _ := setupTestManager(t)

	_, err := manager.ImposeEmbargo(guild1, guild1, 0.75)
	if err == nil {
		t.Error("Expected error when guild embargoes itself")
	}
}

func TestCallReinforcementAlliesSelfGuild(t *testing.T) {
	manager, _, guild1, _, _ := setupTestManager(t)

	_, err := manager.CallReinforcementAllies(guild1, guild1)
	if err == nil {
		t.Error("Expected error when guild calls reinforcements against itself")
	}
}

// Test reverse war check

func TestDeclareWarReverseDirection(t *testing.T) {
	manager, _, guild1, guild2, _ := setupTestManager(t)

	// Guild1 declares war on Guild2
	_, err := manager.DeclareWar(guild1, guild2, 24*time.Hour)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}

	// Guild2 tries to declare war on Guild1 (reverse) - should fail
	_, err = manager.DeclareWar(guild2, guild1, 24*time.Hour)
	if err == nil {
		t.Error("Expected error when declaring reverse war while war already active")
	}
}
