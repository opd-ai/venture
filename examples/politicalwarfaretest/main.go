package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/integration/political_warfare"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
)

func main() {
	mode := flag.String("mode", "all", "Test mode: war, treaty, embargo, alliance, diplomatic, all")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	fmt.Println("=== Political Warfare Integration Test ===")
	fmt.Println()

	switch *mode {
	case "war":
		testWarDeclaration(*verbose)
	case "treaty":
		testPeaceTreaty(*verbose)
	case "embargo":
		testTradeEmbargo(*verbose)
	case "alliance":
		testAllianceCall(*verbose)
	case "diplomatic":
		testDiplomaticVictory(*verbose)
	case "all":
		testWarDeclaration(*verbose)
		testPeaceTreaty(*verbose)
		testTradeEmbargo(*verbose)
		testAllianceCall(*verbose)
		testDiplomaticVictory(*verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
	}
}

func setupManager() (*political_warfare.Manager, *guild.Manager, string, string, string) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()

	// Create test guilds
	guildID1, _ := guildManager.CreateGuild("fantasy", "player1")
	guildID2, _ := guildManager.CreateGuild("fantasy", "player2")
	guildID3, _ := guildManager.CreateGuild("fantasy", "player3")

	// Set up reputation
	guild1, _ := guildManager.GetGuild(guildID1)
	guild3, _ := guildManager.GetGuild(guildID3)
	guild1.Reputation[guildID3] = 0.7
	guild3.Reputation[guildID1] = 0.7
	guild1.Treasury = 100000
	guild3.Treasury = 100000

	manager := political_warfare.NewManager(world, guildManager)
	return manager, guildManager, guildID1, guildID2, guildID3
}

func testWarDeclaration(verbose bool) {
	fmt.Println("--- War Declaration Test ---")
	manager, guildManager, guild1, guild2, _ := setupManager()

	g1, _ := guildManager.GetGuild(guild1)
	g2, _ := guildManager.GetGuild(guild2)

	fmt.Printf("Declaring war: %s vs %s\n", g1.Name, g2.Name)

	preparationPeriod := 24 * time.Hour
	war, err := manager.DeclareWar(guild1, guild2, preparationPeriod)
	if err != nil {
		log.Fatalf("War declaration failed: %v", err)
	}

	fmt.Printf("✓ War declared successfully\n")
	fmt.Printf("  Attacker: %s\n", g1.Name)
	fmt.Printf("  Defender: %s\n", g2.Name)
	fmt.Printf("  Preparation Period: %v\n", war.PreparationPeriod)
	fmt.Printf("  Preparation Ends: %v\n", war.PreparationEnds.Format(time.RFC3339))
	fmt.Printf("  Active: %v\n", war.Active)

	if verbose {
		fmt.Printf("  Declared At: %v\n", war.DeclaredAt.Format(time.RFC3339))
	}

	// Test reputation penalty
	penalties := manager.GetReputationPenalties()
	if len(penalties) > 0 {
		fmt.Printf("  Reputation Penalties Applied: %d\n", len(penalties))
		if verbose {
			for _, p := range penalties {
				fmt.Printf("    - Guild: %s, Action: %s, Penalty: %.2f\n", p.GuildID, p.Action, p.Penalty)
			}
		}
	}

	fmt.Println()
}

func testPeaceTreaty(verbose bool) {
	fmt.Println("--- Peace Treaty Test ---")
	manager, guildManager, guild1, guild2, _ := setupManager()

	g1, _ := guildManager.GetGuild(guild1)
	g2, _ := guildManager.GetGuild(guild2)

	fmt.Printf("Signing peace treaty: %s and %s\n", g1.Name, g2.Name)

	duration := 14 * 24 * time.Hour
	treaty, err := manager.SignPeaceTreaty(guild1, guild2, duration)
	if err != nil {
		log.Fatalf("Peace treaty failed: %v", err)
	}

	fmt.Printf("✓ Peace treaty signed successfully\n")
	fmt.Printf("  Guilds: %s <-> %s\n", g1.Name, g2.Name)
	fmt.Printf("  Duration: %v\n", treaty.Duration)
	fmt.Printf("  Expires: %v\n", treaty.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("  Cooldown Ends: %v\n", treaty.CooldownEnds.Format(time.RFC3339))
	fmt.Printf("  Active: %v\n", treaty.Active)

	if verbose {
		fmt.Printf("  Signed At: %v\n", treaty.SignedAt.Format(time.RFC3339))
	}

	fmt.Println()
}

func testTradeEmbargo(verbose bool) {
	fmt.Println("--- Trade Embargo Test ---")
	manager, guildManager, guild1, guild2, _ := setupManager()

	g1, _ := guildManager.GetGuild(guild1)
	g2, _ := guildManager.GetGuild(guild2)

	fmt.Printf("Imposing embargo: %s against %s\n", g1.Name, g2.Name)

	priceIncrease := 0.75 // 75%
	embargo, err := manager.ImposeEmbargo(guild1, guild2, priceIncrease)
	if err != nil {
		log.Fatalf("Embargo failed: %v", err)
	}

	fmt.Printf("✓ Embargo imposed successfully\n")
	fmt.Printf("  Imposing Guild: %s\n", g1.Name)
	fmt.Printf("  Target Guild: %s\n", g2.Name)
	fmt.Printf("  Price Increase: %.0f%%\n", embargo.PriceIncrease*100)
	fmt.Printf("  Active: %v\n", embargo.Active)

	if verbose {
		fmt.Printf("  Imposed At: %v\n", embargo.ImposedAt.Format(time.RFC3339))
	}

	fmt.Println()
}

func testAllianceCall(verbose bool) {
	fmt.Println("--- Alliance Reinforcement Call Test ---")
	manager, guildManager, guild1, guild2, guild3 := setupManager()

	g1, _ := guildManager.GetGuild(guild1)
	g2, _ := guildManager.GetGuild(guild2)
	g3, _ := guildManager.GetGuild(guild3)

	fmt.Printf("Calling allies for siege: %s against %s\n", g1.Name, g2.Name)

	call, err := manager.CallReinforcementAllies(guild1, guild2)
	if err != nil {
		log.Fatalf("Alliance call failed: %v", err)
	}

	fmt.Printf("✓ Alliance call completed\n")
	fmt.Printf("  Calling Guild: %s\n", g1.Name)
	fmt.Printf("  Target Guild: %s\n", g2.Name)
	fmt.Printf("  Allies Responded: %d\n", len(call.ResponingAllies))

	for i, response := range call.ResponingAllies {
		allyGuild, _ := guildManager.GetGuild(response.AllyGuildID)
		status := "DECLINED"
		if response.Accepted {
			status = "ACCEPTED"
		}
		fmt.Printf("  %d. %s - %s (Success Rate: %.1f%%)\n", i+1, allyGuild.Name, status, response.SuccessRate*100)
		if verbose {
			fmt.Printf("      Responded At: %v\n", response.RespondedAt.Format(time.RFC3339))
		}
	}

	if len(call.ResponingAllies) == 0 {
		fmt.Printf("  (No allies with sufficient reputation >= 0.6)\n")
		fmt.Printf("  Note: %s has reputation 0.7 with %s\n", g1.Name, g3.Name)
	}

	fmt.Println()
}

func testDiplomaticVictory(verbose bool) {
	fmt.Println("--- Diplomatic Victory Test ---")
	manager, guildManager, guild1, guild2, _ := setupManager()

	g1, _ := guildManager.GetGuild(guild1)
	g2, _ := guildManager.GetGuild(guild2)

	// First declare war
	fmt.Printf("Declaring war: %s vs %s\n", g1.Name, g2.Name)
	war, err := manager.DeclareWar(guild1, guild2, 100*time.Millisecond)
	if err != nil {
		log.Fatalf("War declaration failed: %v", err)
	}

	// Wait for war to activate
	time.Sleep(200 * time.Millisecond)
	manager.Update(0)

	fmt.Printf("War now active\n")
	fmt.Printf("Attempting diplomatic negotiation...\n")

	// Attempt diplomatic victory with concessions
	concessions := []political_warfare.DiplomaticConcession{
		{Type: political_warfare.ConcessionGold, Value: 50000},
		{Type: political_warfare.ConcessionApology, Value: "We seek peace"},
	}

	// Try multiple times to demonstrate probabilistic nature
	attempts := 10
	successes := 0
	for i := 0; i < attempts; i++ {
		success, _ := manager.NegotiateDiplomaticVictory(guild1, guild2, concessions)
		if success {
			successes++
		}
		// Reset war for next attempt
		if i < attempts-1 {
			war.Ended = false
			war.Active = true
		}
	}

	successRate := float64(successes) / float64(attempts)
	fmt.Printf("✓ Diplomatic negotiations completed\n")
	fmt.Printf("  Attempts: %d\n", attempts)
	fmt.Printf("  Successes: %d\n", successes)
	fmt.Printf("  Success Rate: %.1f%% (Expected: 10-20%%)\n", successRate*100)
	fmt.Printf("  Concessions Offered:\n")
	fmt.Printf("    - Gold: 50,000\n")
	fmt.Printf("    - Apology: \"We seek peace\"\n")

	if verbose {
		fmt.Printf("  War Status:\n")
		fmt.Printf("    Ended: %v\n", war.Ended)
		fmt.Printf("    Victor: %s\n", war.Victor)
		fmt.Printf("    Victory Type: %s\n", war.VictoryType)
	}

	fmt.Println()
}
