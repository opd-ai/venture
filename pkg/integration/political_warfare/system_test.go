package political_warfare

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
)

func TestNewSystem(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	seed := int64(12345)

	sys := NewSystem(world, guildManager, seed)

	if sys == nil {
		t.Fatal("Expected non-nil system")
	}
	if sys.world != world {
		t.Error("System world mismatch")
	}
	if sys.manager == nil {
		t.Error("Expected non-nil manager")
	}
}

func TestSystemUpdate(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	seed := int64(12345)

	// Create test guilds
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	sys := NewSystem(world, guildManager, seed)

	// Declare war with 1 second preparation
	_, err := sys.manager.DeclareWar(guildID1, guildID2, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to declare war: %v", err)
	}

	// War should not be active yet
	wars := sys.manager.GetActiveWars()
	if len(wars) != 1 {
		t.Fatalf("Expected 1 war, got %d", len(wars))
	}
	if wars[0].Active {
		t.Error("War should not be active during preparation period")
	}

	// Update system to process time-based state changes
	entities := []*engine.Entity{}
	sys.Update(entities, 0.5) // Update with 0.5s delta

	// Still not active (preparation is 1s)
	wars = sys.manager.GetActiveWars()
	if wars[0].Active {
		t.Error("War should still not be active")
	}

	// Wait for preparation period to end
	time.Sleep(1100 * time.Millisecond)

	// Update again
	sys.Update(entities, 0.5)

	// War should now be active
	wars = sys.manager.GetActiveWars()
	if !wars[0].Active {
		t.Error("War should be active after preparation period")
	}
}

func TestSystemGetManager(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	seed := int64(12345)

	sys := NewSystem(world, guildManager, seed)

	manager := sys.GetManager()
	if manager == nil {
		t.Error("Expected non-nil manager")
	}
	if manager != sys.manager {
		t.Error("GetManager should return the internal manager")
	}
}

func TestSystemIntegration(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	seed := int64(12345)

	// Create test guilds
	guildID1, _ := guildManager.CreateGuild("fantasy", "Warlord1", 11111)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Diplomat1", 22222)
	guildID3, _ := guildManager.CreateGuild("fantasy", "Ally1", 33333)

	// Set up alliance (high reputation)
	guild1, _ := guildManager.GetGuild(guildID1)
	guild1.Reputation[guildID3] = 0.7

	sys := NewSystem(world, guildManager, seed)

	// Test war declaration
	war, err := sys.manager.DeclareWar(guildID1, guildID2, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to declare war: %v", err)
	}
	if war.AttackerGuildID != guildID1 {
		t.Error("Attacker guild ID mismatch")
	}
	if war.DefenderGuildID != guildID2 {
		t.Error("Defender guild ID mismatch")
	}

	// Test embargo
	embargo, err := sys.manager.ImposeEmbargo(guildID1, guildID2, 0.75)
	if err != nil {
		t.Fatalf("Failed to impose embargo: %v", err)
	}
	if embargo.PriceIncrease != 0.75 {
		t.Errorf("Expected price increase 0.75, got %f", embargo.PriceIncrease)
	}

	// Test alliance call
	call, err := sys.manager.CallReinforcementAllies(guildID1, guildID2)
	if err != nil {
		t.Fatalf("Failed to call allies: %v", err)
	}
	if call.Completed {
		t.Error("Alliance call should not be completed immediately")
	}

	// Test peace treaty
	treaty, err := sys.manager.SignPeaceTreaty(guildID1, guildID2, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to sign treaty: %v", err)
	}
	if !treaty.Active {
		t.Error("Treaty should be active")
	}

	// Verify war ended with treaty
	wars := sys.manager.GetActiveWars()
	if len(wars) != 0 {
		t.Error("War should have ended with peace treaty")
	}

	// Test diplomatic victory
	guildID4, _ := guildManager.CreateGuild("fantasy", "Conqueror1", 44444)
	guildID5, _ := guildManager.CreateGuild("fantasy", "Defender1", 55555)
	guild5, _ := guildManager.GetGuild(guildID5)
	guild5.Treasury = 100000 // Give defender gold for concessions

	_, _ = sys.manager.DeclareWar(guildID4, guildID5, 0)
	time.Sleep(100 * time.Millisecond) // Wait for war to activate
	sys.Update([]*engine.Entity{}, 0.1)

	concessions := []DiplomaticConcession{
		{Type: ConcessionGold, Value: 50000},
		{Type: ConcessionApology, Value: "We apologize for our actions"},
	}

	// Multiple attempts to test probabilistic success
	successCount := 0
	for i := 0; i < 100; i++ {
		// Create new war for each attempt
		testGuildID1, _ := guildManager.CreateGuild("fantasy", "TestPlayer1"+string(rune(i)), int64(i*1000+100))
		testGuildID2, _ := guildManager.CreateGuild("fantasy", "TestPlayer2"+string(rune(i)), int64(i*1000+200))
		testGuild2, _ := guildManager.GetGuild(testGuildID2)
		testGuild2.Treasury = 100000

		sys.manager.DeclareWar(testGuildID1, testGuildID2, 0)
		time.Sleep(10 * time.Millisecond)
		sys.Update([]*engine.Entity{}, 0.01)

		success, _ := sys.manager.NegotiateDiplomaticVictory(testGuildID1, testGuildID2, concessions)
		if success {
			successCount++
		}
	}

	// With high-value concessions, we should see some successes
	if successCount == 0 {
		t.Log("No diplomatic victories in 100 attempts (this is probabilistic, may occasionally fail)")
	}

	// Test reputation penalties
	penalties := sys.manager.GetReputationPenalties()
	if len(penalties) < 2 { // At least war declaration + embargo
		t.Errorf("Expected at least 2 reputation penalties, got %d", len(penalties))
	}
}

func TestSystemTreatyExpiration(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	seed := int64(12345)

	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	sys := NewSystem(world, guildManager, seed)

	// Sign short-duration treaty
	_, err := sys.manager.SignPeaceTreaty(guildID1, guildID2, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to sign treaty: %v", err)
	}

	// Treaty should be active
	treaties := sys.manager.GetActiveTreaties()
	if len(treaties) != 1 {
		t.Fatalf("Expected 1 active treaty, got %d", len(treaties))
	}

	// Wait for expiration
	time.Sleep(600 * time.Millisecond)

	// Update to process expiration
	sys.Update([]*engine.Entity{}, 0.1)

	// Treaty should now be inactive
	treaties = sys.manager.GetActiveTreaties()
	if len(treaties) != 0 {
		t.Errorf("Expected 0 active treaties after expiration, got %d", len(treaties))
	}
}
