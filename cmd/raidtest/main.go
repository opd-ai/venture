// Command raidtest demonstrates the raid dungeon generation and instance management system.
//
// This tool showcases Phase 59.1 features:
//   - Raid generation across all 5 tiers (Normal, Heroic, Mythic, Legendary, Nightmare)
//   - Multi-room boss encounters with procedural mechanics
//   - Instance creation and management
//   - Weekly lockout system
//   - Boss phase transitions and mechanics
//
// Usage:
//
//	# Generate and display raid
//	./raidtest
//
//	# Test specific tier
//	./raidtest -tier mythic
//
//	# Test all tiers
//	./raidtest -mode all
//
//	# Test instance management
//	./raidtest -mode instance
//
//	# Test lockout system
//	./raidtest -mode lockout
//
//	# Verbose output
//	./raidtest -verbose
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/world/raids"
)

var (
	mode    = flag.String("mode", "demo", "Test mode: demo, all, instance, lockout, mechanics, benchmark")
	tier    = flag.String("tier", "normal", "Raid tier: normal, heroic, mythic, legendary, nightmare")
	seed    = flag.Int64("seed", 12345, "Random seed for deterministic generation")
	depth   = flag.Int("depth", 10, "Dungeon depth level")
	verbose = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Parse()

	switch *mode {
	case "demo":
		runDemo()
	case "all":
		runAllTiers()
	case "instance":
		runInstanceTest()
	case "lockout":
		runLockoutTest()
	case "mechanics":
		runMechanicsTest()
	case "benchmark":
		runBenchmark()
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
	}
}

func runDemo() {
	fmt.Println("=== Raid Dungeon Generation Demo ===")
	fmt.Printf("Tier: %s, Seed: %d, Depth: %d\n\n", *tier, *seed, *depth)

	gen := raids.NewGenerator(*seed)
	raidTier := parseTier(*tier)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      *depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier": raidTier,
		},
	}

	result, err := gen.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Failed to generate raid: %v", err)
	}

	raid := result.(*raids.RaidDungeon)
	printRaidInfo(raid)
}

func runAllTiers() {
	fmt.Println("=== All Raid Tiers ===")

	gen := raids.NewGenerator(*seed)
	tiers := []raids.RaidTier{
		raids.TierNormal,
		raids.TierHeroic,
		raids.TierMythic,
		raids.TierLegendary,
		raids.TierNightmare,
	}

	for _, tier := range tiers {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      *depth,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"tier": tier,
			},
		}

		result, err := gen.Generate(*seed, params)
		if err != nil {
			log.Printf("Failed to generate %s tier: %v", tier, err)
			continue
		}

		raid := result.(*raids.RaidDungeon)
		fmt.Printf("=== %s ===\n", tier)
		fmt.Printf("Name: %s\n", raid.Name)
		fmt.Printf("Difficulty: %.1fx\n", tier.DifficultyMultiplier())
		fmt.Printf("Players: %d-%d\n", tier.MinPlayers(), tier.MaxPlayers())
		fmt.Printf("Bosses: %d\n", len(raid.Bosses))
		fmt.Printf("Rooms: %d\n\n", len(raid.Rooms))
	}
}

func runInstanceTest() {
	fmt.Println("=== Instance Management Test ===")

	gen := raids.NewGenerator(*seed)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      *depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier": raids.TierMythic,
		},
	}

	result, err := gen.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Failed to generate raid: %v", err)
	}

	raid := result.(*raids.RaidDungeon)
	manager := raids.NewInstanceManager()

	// Create multiple instances
	groups := []string{"group-alpha", "group-beta", "group-gamma"}
	playerIDs := []string{"player-1", "player-2", "player-3", "player-4", "player-5"}

	fmt.Println("Creating raid instances:")
	var firstInstanceID string
	for _, groupID := range groups {
		instance, err := manager.CreateInstance(raid, groupID, playerIDs)
		if err != nil {
			log.Printf("  Failed to create instance for %s: %v\n", groupID, err)
			continue
		}
		if firstInstanceID == "" {
			firstInstanceID = instance.InstanceID
		}
		fmt.Printf("  Group %s -> Instance %s\n", groupID, instance.InstanceID)
	}

	fmt.Printf("\nTotal active instances: %d\n", manager.GetActiveInstanceCount())

	// Get instance info
	if firstInstanceID != "" {
		instance, exists := manager.GetInstance(firstInstanceID)
		if exists {
			fmt.Printf("\nInstance Details (Group %s):\n", groups[0])
			fmt.Printf("  ID: %s\n", instance.InstanceID)
			fmt.Printf("  Created: %s\n", instance.CreatedAt.Format(time.RFC3339))
			fmt.Printf("  Expires: %s\n", instance.ExpiresAt.Format(time.RFC3339))
			fmt.Printf("  Players: %d\n", len(instance.PlayerIDs))
			fmt.Printf("  Completed: %v\n", instance.Completed)
		}
	}

	// Test cleanup
	fmt.Println("\nTesting cleanup system...")
	cleaned := manager.CleanupExpired()
	fmt.Printf("  Cleaned up %d expired instances\n", cleaned)
}

func runLockoutTest() {
	fmt.Println("=== Lockout System Test ===")

	lockoutMgr := raids.NewLockoutManager()
	playerID := "player-123"

	// Apply lockouts
	fmt.Println("Applying lockouts:")
	for _, tier := range []raids.RaidTier{raids.TierNormal, raids.TierHeroic, raids.TierMythic} {
		lockoutMgr.RecordClear(playerID, tier)
		fmt.Printf("  %s tier locked\n", tier)
	}

	// Check lockouts
	fmt.Println("\nChecking lockouts:")
	for _, tier := range []raids.RaidTier{
		raids.TierNormal,
		raids.TierHeroic,
		raids.TierMythic,
		raids.TierLegendary,
		raids.TierNightmare,
	} {
		locked := lockoutMgr.IsLockedOut(playerID, tier)
		if locked {
			remaining := lockoutMgr.TimeUntilReset(playerID, tier)
			fmt.Printf("  %s: LOCKED (expires in %s)\n", tier, remaining.Round(time.Second))
		} else {
			fmt.Printf("  %s: Available\n", tier)
		}
	}

	// Test lockout clearing
	fmt.Println("\nClearing Normal tier lockout...")
	lockoutMgr.ClearLockout(playerID, raids.TierNormal)
	locked := lockoutMgr.IsLockedOut(playerID, raids.TierNormal)
	fmt.Printf("  Normal tier locked: %v\n", locked)
}

func runMechanicsTest() {
	fmt.Println("=== Boss Mechanics Test ===")

	gen := raids.NewGenerator(*seed)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      *depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier": raids.TierMythic,
		},
	}

	result, err := gen.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Failed to generate raid: %v", err)
	}

	raid := result.(*raids.RaidDungeon)
	for i, boss := range raid.Bosses {
		fmt.Printf("=== Boss %d: %s ===\n", i+1, boss.Entity.Name)
		fmt.Printf("Health: %d\n", boss.Entity.Stats.Health)
		fmt.Printf("Damage: %d\n", boss.Entity.Stats.Damage)
		fmt.Printf("Defense: %d\n\n", boss.Entity.Stats.Defense)

		fmt.Printf("Mechanics (%d):\n", len(boss.Mechanics))
		for _, mech := range boss.Mechanics {
			fmt.Printf("  - %s [%s]\n", mech.Name, mech.Type)
			if *verbose {
				fmt.Printf("    %s\n", mech.Description)
				fmt.Printf("    Cooldown: %s, Damage: %d\n", mech.Cooldown, mech.Damage)
				if mech.AoE {
					fmt.Printf("    AoE Radius: %.1f\n", mech.Radius)
				}
			}
		}

		fmt.Printf("\nPhases (%d):\n", len(boss.Phases))
		for _, phase := range boss.Phases {
			fmt.Printf("  Phase %d (at %.0f%% HP):\n", phase.Number, phase.HealthThresh*100)
			fmt.Printf("    Mechanics: %s\n", strings.Join(phase.Mechanics, ", "))
			if phase.AddSpawns > 0 {
				fmt.Printf("    Spawns: %d adds\n", phase.AddSpawns)
			}
		}
		fmt.Println()
	}
}

func runBenchmark() {
	fmt.Println("=== Performance Benchmark ===")

	iterations := 10
	tiers := []raids.RaidTier{
		raids.TierNormal,
		raids.TierHeroic,
		raids.TierMythic,
		raids.TierLegendary,
		raids.TierNightmare,
	}

	for _, tier := range tiers {
		total := time.Duration(0)
		gen := raids.NewGenerator(*seed)

		for i := 0; i < iterations; i++ {
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      *depth,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"tier": tier,
				},
			}

			start := time.Now()
			_, err := gen.Generate(*seed, params)
			if err != nil {
				log.Printf("Error generating %s: %v", tier, err)
				continue
			}
			total += time.Since(start)
		}

		avg := total / time.Duration(iterations)
		fmt.Printf("%s: %s average (target: <5s)\n", tier, avg)
	}

	fmt.Println("\nInstance operations:")
	manager := raids.NewInstanceManager()
	gen := raids.NewGenerator(*seed)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      *depth,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier": raids.TierMythic,
		},
	}
	result, _ := gen.Generate(*seed, params)
	raid := result.(*raids.RaidDungeon)

	start := time.Now()
	playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}
	for i := 0; i < 100; i++ {
		manager.CreateInstance(raid, fmt.Sprintf("group-%d", i), playerIDs)
	}
	elapsed := time.Since(start)
	fmt.Printf("Create 100 instances: %s (%.2fms per instance, target: <100ms)\n",
		elapsed, float64(elapsed.Microseconds())/100/1000)

	start = time.Now()
	// Get the instance ID from first created instance
	testInstance, _ := manager.CreateInstance(raid, "test-group", playerIDs)
	testInstanceID := testInstance.InstanceID
	for i := 0; i < 10000; i++ {
		manager.GetInstance(testInstanceID)
	}
	elapsed = time.Since(start)
	fmt.Printf("10,000 instance lookups: %s (%.2fµs per lookup)\n",
		elapsed, float64(elapsed.Microseconds())/10000)

	fmt.Println("\nLockout operations:")
	lockoutMgr := raids.NewLockoutManager()

	start = time.Now()
	for i := 0; i < 1000; i++ {
		lockoutMgr.RecordClear(fmt.Sprintf("player-%d", i), raids.TierMythic)
	}
	elapsed = time.Since(start)
	fmt.Printf("1,000 lockout applications: %s (%.2fµs per lockout)\n",
		elapsed, float64(elapsed.Microseconds())/1000)

	start = time.Now()
	for i := 0; i < 100000; i++ {
		lockoutMgr.IsLockedOut("player-500", raids.TierMythic)
	}
	elapsed = time.Since(start)
	fmt.Printf("100,000 lockout checks: %s (%.2fns per check, target: <1ms)\n",
		elapsed, float64(elapsed.Nanoseconds())/100000)
}

func printRaidInfo(raid *raids.RaidDungeon) {
	fmt.Printf("=== %s ===\n", raid.Name)
	fmt.Printf("Tier: %s (%.1fx difficulty)\n", raid.Tier, raid.Tier.DifficultyMultiplier())
	fmt.Printf("Players: %d-%d\n", raid.Tier.MinPlayers(), raid.Tier.MaxPlayers())
	fmt.Printf("Seed: %d\n", raid.Seed)
	fmt.Printf("Created: %s\n\n", raid.CreatedAt.Format(time.RFC3339))

	fmt.Printf("Description: %s\n\n", raid.Description)

	fmt.Printf("Layout:\n")
	fmt.Printf("  Terrain: %dx%d tiles\n", raid.Terrain.Width, raid.Terrain.Height)
	fmt.Printf("  Rooms: %d total\n", len(raid.Rooms))

	roomCounts := make(map[raids.RoomType]int)
	for _, room := range raid.Rooms {
		roomCounts[room.Type]++
	}
	for roomType, count := range roomCounts {
		fmt.Printf("    %s: %d\n", roomType, count)
	}

	fmt.Printf("\nBosses (%d):\n", len(raid.Bosses))
	for i, boss := range raid.Bosses {
		fmt.Printf("  %d. %s\n", i+1, boss.Entity.Name)
		fmt.Printf("     HP: %d, Damage: %d, Defense: %d\n",
			boss.Entity.Stats.Health, boss.Entity.Stats.Damage, boss.Entity.Stats.Defense)
		fmt.Printf("     Mechanics: %d, Phases: %d\n",
			len(boss.Mechanics), len(boss.Phases))

		if *verbose {
			for _, mech := range boss.Mechanics {
				fmt.Printf("       - %s [%s]: %s\n", mech.Name, mech.Type, mech.Description)
			}
		}
	}

	if *verbose {
		fmt.Printf("\nRoom Details:\n")
		for _, room := range raid.Rooms {
			fmt.Printf("  %s (%s): %dx%d at (%d, %d)\n",
				room.ID, room.Type, room.W, room.H, room.X, room.Y)
			if len(room.Connections) > 0 {
				fmt.Printf("    Connections: %s\n", strings.Join(room.Connections, ", "))
			}
			if room.BossID != "" {
				fmt.Printf("    Boss: %s\n", room.BossID)
			}
		}
	}
}

func parseTier(tierStr string) raids.RaidTier {
	switch strings.ToLower(tierStr) {
	case "normal":
		return raids.TierNormal
	case "heroic":
		return raids.TierHeroic
	case "mythic":
		return raids.TierMythic
	case "legendary":
		return raids.TierLegendary
	case "nightmare":
		return raids.TierNightmare
	default:
		return raids.TierNormal
	}
}
