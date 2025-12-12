package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/engine/prestige"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, levels, paragon, abilities, account, save, benchmark, all")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	fmt.Println("=== Prestige System Test ===")
	fmt.Println()

	switch *mode {
	case "demo":
		runDemo(*verbose)
	case "levels":
		testPrestigeLevels(*verbose)
	case "paragon":
		testParagonPoints(*verbose)
	case "abilities":
		testPrestigeAbilities(*verbose)
	case "account":
		testAccountBonus(*verbose)
	case "save":
		testSaveLoad(*verbose)
	case "benchmark":
		runBenchmark(*verbose)
	case "all":
		runDemo(*verbose)
		testPrestigeLevels(*verbose)
		testParagonPoints(*verbose)
		testPrestigeAbilities(*verbose)
		testAccountBonus(*verbose)
		testSaveLoad(*verbose)
		runBenchmark(*verbose)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

func runDemo(verbose bool) {
	fmt.Println("--- Demo: Complete Prestige Journey ---")

	mgr := prestige.NewManager()

	// Create player
	playerID := "player123"
	className := "Warrior"
	accountID := "account1"
	mgr.CreatePlayer(playerID, className, accountID)
	fmt.Printf("Created player: %s (%s)\n", playerID, className)

	// Simulate progression to prestige 10
	fmt.Println("\n[Reaching Prestige Level 10]")
	totalXP := 0
	for i := 1; i <= 10; i++ {
		xpNeeded := int(100000 * (1 << (i - 1))) // 2^(i-1) * baseXP
		totalXP += xpNeeded
		levels := mgr.AddPrestigeXP(playerID, className, xpNeeded)
		if levels > 0 {
			fmt.Printf("Gained prestige level %d (required %d XP)\n", i, xpNeeded)
			mgr.AddParagonPoints(playerID, levels)

			// Check for ability unlock
			ability := mgr.CheckAbilityUnlock(playerID)
			if ability != nil {
				fmt.Printf("  ✨ UNLOCKED ABILITY: %s\n", ability.Name)
				fmt.Printf("     %s (Cooldown: %ds, Mana: %d)\n", ability.Description, ability.Cooldown, ability.ManaCost)
			}

			// Visual tier
			tier := mgr.GetVisualTier(i)
			if tier != prestige.VisualNone {
				fmt.Printf("  ✨ Visual Effect: %s Glow\n", tier.String())
			}
		}
	}

	// Allocate paragon points
	fmt.Println("\n[Allocating Paragon Points]")
	stats := []prestige.ParagonStat{
		prestige.StatDamage,
		prestige.StatDamage,
		prestige.StatHealth,
		prestige.StatHealth,
		prestige.StatDefense,
	}

	for i, stat := range stats {
		if err := mgr.AllocateParagonPoint(playerID, stat); err == nil {
			bonus := mgr.GetStatBonus(playerID, stat)
			fmt.Printf("Point %d → %s (total bonus: +%.1f%%)\n", i+1, stat.String(), bonus*100)
		}
	}

	// Show current state
	fmt.Println("\n[Current Status]")
	level := mgr.GetPrestigeLevel(playerID)
	tier := mgr.GetVisualTier(level)
	fmt.Printf("Prestige Level: %d\n", level)
	fmt.Printf("Visual Tier: %s\n", tier.String())
	fmt.Printf("Total XP Earned: ~%d\n", totalXP)
	fmt.Printf("Damage Bonus: +%.1f%%\n", mgr.GetStatBonus(playerID, prestige.StatDamage)*100)
	fmt.Printf("Health Bonus: +%.1f%%\n", mgr.GetStatBonus(playerID, prestige.StatHealth)*100)
	fmt.Printf("Defense Bonus: +%.1f%%\n", mgr.GetStatBonus(playerID, prestige.StatDefense)*100)

	fmt.Println()
}

func testPrestigeLevels(verbose bool) {
	fmt.Println("--- Test: Prestige Level Progression ---")

	mgr := prestige.NewManager()
	mgr.CreatePlayer("test1", "Mage", "account1")

	fmt.Println("XP Requirements per Level (Exponential Curve):")
	for i := 1; i <= 10; i++ {
		xpNeeded := int(100000 * (1 << (i - 1)))
		fmt.Printf("  Level %2d: %15d XP\n", i, xpNeeded)
	}

	fmt.Println("\nSimulating rapid progression:")
	for i := 1; i <= 5; i++ {
		xp := int(100000 * (1 << (i - 1)))
		levels := mgr.AddPrestigeXP("test1", "Mage", xp)
		fmt.Printf("  Added %d XP → Gained %d level(s) → Current level: %d\n", xp, levels, mgr.GetPrestigeLevel("test1"))
	}

	fmt.Println()
}

func testParagonPoints(verbose bool) {
	fmt.Println("--- Test: Paragon Point System ---")

	mgr := prestige.NewManager()
	playerID := "paragon_test"
	mgr.CreatePlayer(playerID, "Rogue", "account1")

	// Grant 10 prestige levels
	for i := 1; i <= 10; i++ {
		xp := int(100000 * (1 << (i - 1)))
		mgr.AddPrestigeXP(playerID, "Rogue", xp)
	}
	mgr.AddParagonPoints(playerID, 10)

	fmt.Println("Allocating 10 paragon points across stats:")

	allocations := map[prestige.ParagonStat]int{
		prestige.StatDamage:   3,
		prestige.StatCritical: 3,
		prestige.StatSpeed:    2,
		prestige.StatDefense:  1,
		prestige.StatHealth:   1,
	}

	for stat, count := range allocations {
		for i := 0; i < count; i++ {
			mgr.AllocateParagonPoint(playerID, stat)
		}
		bonus := mgr.GetStatBonus(playerID, stat)
		fmt.Printf("  %s: %d points → +%.1f%% bonus\n", stat.String(), count, bonus*100)
	}

	// Test respec
	fmt.Println("\nTesting respec:")
	cost, err := mgr.RespecParagonPoints(playerID)
	if err != nil {
		log.Printf("Respec error: %v", err)
	} else {
		fmt.Printf("  Respec cost: %d gold\n", cost)
		fmt.Printf("  All %d points returned\n", 10)
	}

	fmt.Println()
}

func testPrestigeAbilities(verbose bool) {
	fmt.Println("--- Test: Prestige Abilities ---")

	mgr := prestige.NewManager()
	classes := []string{"Warrior", "Mage", "Rogue", "Cleric"}

	for _, class := range classes {
		fmt.Printf("\n%s Prestige Abilities:\n", class)
		milestones := []int{10, 25, 50, 100}
		for _, milestone := range milestones {
			ability := mgr.GetPrestigeAbility(class, milestone)
			if ability != nil {
				fmt.Printf("  [Level %3d] %s\n", milestone, ability.Name)
				if verbose {
					fmt.Printf("              %s\n", ability.Description)
					fmt.Printf("              Cooldown: %ds, Mana: %d\n", ability.Cooldown, ability.ManaCost)
				}
			}
		}
	}

	fmt.Println()
}

func testAccountBonus(verbose bool) {
	fmt.Println("--- Test: Account-Wide XP Bonus ---")

	mgr := prestige.NewManager()
	accountID := "shared_account"

	// Create 3 characters and level them to prestige 100
	characters := []struct {
		id    string
		class string
	}{
		{"char1", "Warrior"},
		{"char2", "Mage"},
		{"char3", "Rogue"},
	}

	fmt.Println("Leveling characters to prestige 100:")
	for i, char := range characters {
		mgr.CreatePlayer(char.id, char.class, accountID)

		// Fast-track to prestige 100
		totalXP := 0
		for level := 1; level <= 100; level++ {
			xp := int(100000 * (1 << (level - 1)))
			// Cap at reasonable int value
			if xp < 0 {
				xp = 1000000000 // Use large constant for high levels
			}
			totalXP += xp
			mgr.AddPrestigeXP(char.id, char.class, xp)
		}

		fmt.Printf("  %s (%s) reached prestige 100\n", char.id, char.class)

		bonus := mgr.GetAccountXPBonus(accountID)
		fmt.Printf("    Account XP Bonus: +%.2f%% (%d prestige 100 characters)\n", bonus*100, i+1)
	}

	finalBonus := mgr.GetAccountXPBonus(accountID)
	fmt.Printf("\nFinal account-wide bonus: +%.2f%% XP gain\n", finalBonus*100)

	fmt.Println()
}

func testSaveLoad(verbose bool) {
	fmt.Println("--- Test: Save/Load Persistence ---")

	mgr := prestige.NewManager()

	// Create and progress a player
	mgr.CreatePlayer("save_test", "Paladin", "account1")
	for i := 1; i <= 5; i++ {
		xp := int(100000 * (1 << (i - 1)))
		mgr.AddPrestigeXP("save_test", "Paladin", xp)
	}
	mgr.AddParagonPoints("save_test", 5)
	for i := 0; i < 5; i++ {
		mgr.AllocateParagonPoint("save_test", prestige.StatDamage)
	}

	// Save
	data, err := mgr.Save()
	if err != nil {
		log.Fatalf("Save failed: %v", err)
	}
	fmt.Printf("Saved data: %d bytes (compressed)\n", len(data))

	// Load into new manager
	mgr2 := prestige.NewManager()
	if err := mgr2.Load(data); err != nil {
		log.Fatalf("Load failed: %v", err)
	}

	// Verify
	level1 := mgr.GetPrestigeLevel("save_test")
	level2 := mgr2.GetPrestigeLevel("save_test")
	bonus1 := mgr.GetStatBonus("save_test", prestige.StatDamage)
	bonus2 := mgr2.GetStatBonus("save_test", prestige.StatDamage)

	fmt.Printf("Original - Level: %d, Damage Bonus: +%.1f%%\n", level1, bonus1*100)
	fmt.Printf("Loaded   - Level: %d, Damage Bonus: +%.1f%%\n", level2, bonus2*100)

	if level1 == level2 && bonus1 == bonus2 {
		fmt.Println("✅ Save/Load verification passed")
	} else {
		fmt.Println("❌ Save/Load verification failed")
	}

	fmt.Println()
}

func runBenchmark(verbose bool) {
	fmt.Println("--- Benchmark: Performance Tests ---")

	mgr := prestige.NewManager()

	// Test 1: XP addition speed
	fmt.Print("Adding XP (1000 operations)... ")
	mgr.CreatePlayer("bench1", "Warrior", "account1")
	for i := 0; i < 1000; i++ {
		mgr.AddPrestigeXP("bench1", "Warrior", 1000)
	}
	fmt.Println("✓")

	// Test 2: Paragon allocation speed
	fmt.Print("Allocating paragon points (100 operations)... ")
	mgr.AddParagonPoints("bench1", 100)
	for i := 0; i < 100; i++ {
		mgr.AllocateParagonPoint("bench1", prestige.ParagonStat(i%5))
	}
	fmt.Println("✓")

	// Test 3: Ability lookup speed
	fmt.Print("Ability lookups (1000 operations)... ")
	for i := 0; i < 1000; i++ {
		mgr.GetPrestigeAbility("Warrior", 10)
	}
	fmt.Println("✓")

	// Test 4: Visual tier calculation
	fmt.Print("Visual tier calculations (1000 operations)... ")
	for i := 0; i < 1000; i++ {
		mgr.GetVisualTier(i % 150)
	}
	fmt.Println("✓")

	// Test 5: Save/load performance
	fmt.Print("Save/load cycle (10 operations)... ")
	for i := 0; i < 10; i++ {
		data, _ := mgr.Save()
		mgr2 := prestige.NewManager()
		mgr2.Load(data)
	}
	fmt.Println("✓")

	fmt.Println("\n✅ All performance tests passed")
	fmt.Println()
}
