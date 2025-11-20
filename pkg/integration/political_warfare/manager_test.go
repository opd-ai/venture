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
	guildID1, _ := guildManager.CreateGuild("fantasy", "player1")
	guildID2, _ := guildManager.CreateGuild("fantasy", "player2")
	guildID3, _ := guildManager.CreateGuild("fantasy", "player3")

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
	if !call.Completed {
		t.Error("Alliance call should be completed")
	}

	// guild3 has 0.7 reputation with guild1, should be called
	found := false
	for _, response := range call.ResponingAllies {
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
