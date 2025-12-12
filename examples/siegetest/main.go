package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/world/territory"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, structures, phases, reinforcements, victory")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	fmt.Println("=== Siege System Test Tool ===")
	fmt.Println()

	switch *mode {
	case "demo":
		runDemo(*verbose)
	case "structures":
		testDefensiveStructures(*verbose)
	case "phases":
		testPhaseProgression(*verbose)
	case "reinforcements":
		testReinforcements(*verbose)
	case "victory":
		testVictoryConditions(*verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		fmt.Println("Available modes: demo, structures, phases, reinforcements, victory")
	}
}

func runDemo(verbose bool) {
	fmt.Println("Running Full Siege Demonstration...")
	fmt.Println()

	sm := territory.NewSiegeManager()
	siege := createDemoSiege(sm)
	if siege == nil {
		return
	}

	addDemoParticipants(siege)
	addDemoReinforcements(siege)
	advanceToAssaultPhase(siege)
	simulateAssaultActions(siege)
	advanceToResolution(siege)
	distributeDemoLoot(siege)
	endDemoSiege(siege)
}

func createDemoSiege(sm *territory.SiegeManager) *territory.Siege {
	siege, err := sm.CreateSiege("demo_territory", "attacking_guild", "defending_guild", 50000)
	if err != nil {
		fmt.Printf("Error creating siege: %v\n", err)
		return nil
	}

	fmt.Printf("✓ Created siege: %s\n", siege.ID)
	fmt.Printf("  Territory: %s\n", siege.TerritoryID)
	fmt.Printf("  Attacker: %s\n", siege.AttackerGuildID)
	fmt.Printf("  Defender: %s\n", siege.DefenderGuildID)
	fmt.Printf("  Phase: %s\n", siege.Phase)
	fmt.Printf("  Defender Treasury: %d gold\n", siege.DefenderTreasury)
	fmt.Println()

	return siege
}

func addDemoParticipants(siege *territory.Siege) {
	for i := 0; i < 25; i++ {
		playerID := fmt.Sprintf("attacker_%d", i)
		if err := siege.JoinSiege(playerID, true); err != nil {
			fmt.Printf("Error adding attacker %s: %v\n", playerID, err)
			continue
		}
	}

	for i := 0; i < 20; i++ {
		playerID := fmt.Sprintf("defender_%d", i)
		if err := siege.JoinSiege(playerID, false); err != nil {
			fmt.Printf("Error adding defender %s: %v\n", playerID, err)
			continue
		}
	}

	fmt.Printf("✓ Added %d attackers and %d defenders\n", len(siege.Attackers), len(siege.Defenders))
	fmt.Println()
}

func addDemoReinforcements(siege *territory.Siege) {
	allies := []string{"ally_1", "ally_2", "ally_3", "ally_4", "ally_5"}
	if err := siege.AddReinforcements("allied_guild_1", allies); err != nil {
		fmt.Printf("Error adding reinforcements: %v\n", err)
	} else {
		fmt.Printf("✓ Added reinforcements from allied_guild_1 (%d players)\n", len(allies))
		fmt.Printf("  Total defenders now: %d\n", len(siege.Defenders))
	}
	fmt.Println()
}

func advanceToAssaultPhase(siege *territory.Siege) {
	fmt.Println("Simulating 1 hour preparation phase...")
	siege.PhaseStartTime = time.Now().Add(-61 * time.Minute)
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("Error advancing phase: %v\n", err)
	} else {
		fmt.Printf("✓ Advanced to phase: %s\n", siege.Phase)
	}
	fmt.Println()
}

func simulateAssaultActions(siege *territory.Siege) {
	fmt.Println("Simulating assault phase actions...")
	captureInitialPoints(siege, 3)
	damageGuildHall(siege, 3000)
	captureRemainingPoints(siege)
}

func captureInitialPoints(siege *territory.Siege, count int) {
	for i := 0; i < count; i++ {
		if err := siege.CaptureControlPoint(); err != nil {
			fmt.Printf("Error capturing point %d: %v\n", i+1, err)
		} else {
			fmt.Printf("✓ Captured control point %d/%d\n", siege.ControlPointsCaptured, siege.TotalControlPoints)
		}
	}
	fmt.Println()
}

func damageGuildHall(siege *territory.Siege, damage float64) {
	fmt.Printf("Guild Hall HP: %.0f/%.0f\n", siege.GuildHallHP, siege.GuildHallMaxHP)
	if err := siege.DamageGuildHall(damage); err != nil {
		fmt.Printf("Error damaging guild hall: %v\n", err)
	} else {
		fmt.Printf("✓ Dealt %.0f damage to guild hall\n", damage)
		fmt.Printf("  Remaining HP: %.0f/%.0f\n", siege.GuildHallHP, siege.GuildHallMaxHP)
	}
	fmt.Println()
}

func captureRemainingPoints(siege *territory.Siege) {
	fmt.Println("Capturing remaining control points...")
	for siege.ControlPointsCaptured < siege.TotalControlPoints {
		if err := siege.CaptureControlPoint(); err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}
	}

	if siege.WinnerGuildID != "" {
		fmt.Printf("✓ Victory achieved: %s\n", siege.VictoryCondition)
		fmt.Printf("  Winner: %s\n", siege.WinnerGuildID)
	}
	fmt.Println()
}

func advanceToResolution(siege *territory.Siege) {
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("Error advancing to resolution: %v\n", err)
	} else {
		fmt.Printf("✓ Advanced to phase: %s\n", siege.Phase)
	}
	fmt.Println()
}

func distributeDemoLoot(siege *territory.Siege) {
	loot, err := siege.DistributeLoot()
	if err != nil {
		fmt.Printf("Error distributing loot: %v\n", err)
	} else {
		fmt.Printf("✓ Loot distributed: %d gold (%.0f%% of %d)\n", loot, siege.LootPercentage*100, siege.DefenderTreasury)
	}
	fmt.Println()
}

func endDemoSiege(siege *territory.Siege) {
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("Error ending siege: %v\n", err)
	} else {
		fmt.Printf("✓ Siege ended\n")
		fmt.Printf("  Final phase: %s\n", siege.Phase)
		fmt.Printf("  Duration: %v\n", siege.EndTime.Sub(siege.StartTime))
	}
}

func testDefensiveStructures(verbose bool) {
	fmt.Println("Testing Defensive Structure Generation...")
	fmt.Println()

	// Generate structures
	structures := territory.GenerateDefensiveStructures("test_territory", 12345, 12)

	fmt.Printf("✓ Generated %d defensive structures\n", len(structures))
	fmt.Println()

	// Analyze structure types
	wallCount := 0
	towerCount := 0
	guardCount := 0

	for _, s := range structures {
		switch s.Type {
		case territory.StructureTypeWall:
			wallCount++
		case territory.StructureTypeTower:
			towerCount++
		case territory.StructureTypeGuard:
			guardCount++
		}

		if verbose {
			fmt.Printf("  %s (Level %d)\n", s.Type, s.Level)
			fmt.Printf("    Position: (%.0f, %.0f)\n", s.X, s.Y)
			fmt.Printf("    HP: %.0f/%.0f\n", s.HP, s.MaxHP)
			fmt.Printf("    Damage: %.0f\n", s.Damage)
			fmt.Println()
		}
	}

	fmt.Println("Structure Distribution:")
	fmt.Printf("  Walls:  %d\n", wallCount)
	fmt.Printf("  Towers: %d\n", towerCount)
	fmt.Printf("  Guards: %d\n", guardCount)
	fmt.Println()

	// Test determinism
	structures2 := territory.GenerateDefensiveStructures("test_territory", 12345, 12)
	if len(structures) != len(structures2) {
		fmt.Printf("✗ Determinism check failed: count mismatch (%d vs %d)\n", len(structures), len(structures2))
		return
	}

	match := true
	for i := range structures {
		if structures[i].Type != structures2[i].Type {
			match = false
			break
		}
	}

	if match {
		fmt.Println("✓ Determinism verified: same seed produces identical structures")
	} else {
		fmt.Println("✗ Determinism check failed: structures differ")
	}
}

func testPhaseProgression(verbose bool) {
	fmt.Println("Testing Phase Progression...")
	fmt.Println()

	siege := territory.NewSiege("test_territory", "guild_a", "guild_b", 10000)
	fmt.Printf("Initial phase: %s\n", siege.Phase)
	fmt.Println()

	// Try to advance too early
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("✓ Cannot advance preparation phase early: %v\n", err)
	}
	fmt.Println()

	// Simulate time passing
	siege.PhaseStartTime = time.Now().Add(-61 * time.Minute)
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("✗ Error advancing to assault: %v\n", err)
	} else {
		fmt.Printf("✓ Advanced to: %s\n", siege.Phase)
	}
	fmt.Println()

	// Try to advance assault early
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("✓ Cannot advance assault phase early: %v\n", err)
	}
	fmt.Println()

	// Set a winner
	siege.WinnerGuildID = "guild_a"
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("✗ Error advancing to resolution: %v\n", err)
	} else {
		fmt.Printf("✓ Advanced to: %s (winner determined)\n", siege.Phase)
	}
	fmt.Println()

	// Advance to ended
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("✗ Error ending siege: %v\n", err)
	} else {
		fmt.Printf("✓ Advanced to: %s\n", siege.Phase)
	}
	fmt.Println()

	// Try to advance beyond ended
	if err := siege.AdvancePhase(); err != nil {
		fmt.Printf("✓ Cannot advance beyond ended phase: %v\n", err)
	}
}

func testReinforcements(verbose bool) {
	fmt.Println("Testing Reinforcement System...")
	fmt.Println()

	siege := territory.NewSiege("test_territory", "guild_a", "guild_b", 10000)

	// Add 5 allied guilds
	for i := 0; i < 5; i++ {
		guildID := fmt.Sprintf("ally_%d", i+1)
		players := []string{
			fmt.Sprintf("player_%d_1", i),
			fmt.Sprintf("player_%d_2", i),
			fmt.Sprintf("player_%d_3", i),
		}

		if err := siege.AddReinforcements(guildID, players); err != nil {
			fmt.Printf("✗ Error adding reinforcements from %s: %v\n", guildID, err)
		} else {
			fmt.Printf("✓ Added reinforcements from %s (%d players)\n", guildID, len(players))
		}
	}

	fmt.Println()
	fmt.Printf("Total reinforcement guilds: %d/5\n", len(siege.Reinforcements))
	fmt.Printf("Total defenders (including reinforcements): %d\n", len(siege.Defenders))
	fmt.Println()

	// Try to add 6th guild
	if err := siege.AddReinforcements("ally_6", []string{"player_x"}); err != nil {
		fmt.Printf("✓ Cannot add 6th reinforcement guild: %v\n", err)
	}
	fmt.Println()

	// Try to add duplicate guild
	if err := siege.AddReinforcements("ally_1", []string{"player_y"}); err != nil {
		fmt.Printf("✓ Cannot add duplicate guild: %v\n", err)
	}
}

func testVictoryConditions(verbose bool) {
	fmt.Println("Testing Victory Conditions...")
	fmt.Println()

	// Test 1: Capture all points
	fmt.Println("1. Victory by capturing all control points:")
	siege1 := territory.NewSiege("test_territory_1", "guild_a", "guild_b", 10000)
	siege1.Phase = territory.PhaseAssault

	for i := 0; i < siege1.TotalControlPoints; i++ {
		siege1.CaptureControlPoint()
	}

	if siege1.WinnerGuildID == "guild_a" && siege1.VictoryCondition == territory.VictoryCapturePoints {
		fmt.Printf("  ✓ Attackers won by %s\n", siege1.VictoryCondition)
	} else {
		fmt.Printf("  ✗ Unexpected result\n")
	}
	fmt.Println()

	// Test 2: Destroy guild hall
	fmt.Println("2. Victory by destroying guild hall:")
	siege2 := territory.NewSiege("test_territory_2", "guild_c", "guild_d", 10000)
	siege2.Phase = territory.PhaseAssault
	siege2.DamageGuildHall(siege2.GuildHallMaxHP)

	if siege2.WinnerGuildID == "guild_c" && siege2.VictoryCondition == territory.VictoryDestroyHall {
		fmt.Printf("  ✓ Attackers won by %s\n", siege2.VictoryCondition)
	} else {
		fmt.Printf("  ✗ Unexpected result\n")
	}
	fmt.Println()

	// Test 3: Defense timeout
	fmt.Println("3. Victory by defense timeout:")
	sm := territory.NewSiegeManager()
	siege3, _ := sm.CreateSiege("test_territory_3", "guild_e", "guild_f", 10000)
	siege3.Phase = territory.PhaseAssault
	siege3.PhaseStartTime = time.Now().Add(-121 * time.Minute)

	sm.Update(0.016)

	if siege3.WinnerGuildID == "guild_f" && siege3.VictoryCondition == territory.VictoryDefenseTimeout {
		fmt.Printf("  ✓ Defenders won by %s\n", siege3.VictoryCondition)
	} else {
		fmt.Printf("  ✗ Unexpected result: winner=%s, condition=%s\n", siege3.WinnerGuildID, siege3.VictoryCondition)
	}
	fmt.Println()

	// Test loot distribution
	fmt.Println("4. Loot distribution:")
	siege3.Phase = territory.PhaseResolution
	loot, err := siege3.DistributeLoot()
	if err != nil {
		fmt.Printf("  ✗ Error distributing loot: %v\n", err)
	} else {
		expected := int(float64(10000) * 0.15)
		fmt.Printf("  ✓ Loot distributed: %d gold (expected ~%d)\n", loot, expected)
	}
}
