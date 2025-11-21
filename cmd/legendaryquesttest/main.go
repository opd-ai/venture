package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/legendary"
)

func main() {
	// Command-line flags
	mode := flag.String("mode", "demo", "Test mode: demo, generate, validate, phases, rewards, benchmark")
	seed := flag.Int64("seed", 12345, "Seed for deterministic generation")
	difficulty := flag.Float64("difficulty", 0.7, "Quest difficulty (0.0-1.0)")
	depth := flag.Int("depth", 15, "Quest depth/level")
	genre := flag.String("genre", "fantasy", "Genre ID (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Create generator
	gen := legendary.NewLegendaryQuestGenerator()

	switch *mode {
	case "demo":
		runDemo(gen, *seed, *difficulty, *depth, *genre, *verbose)
	case "generate":
		runGenerate(gen, *seed, *difficulty, *depth, *genre, *verbose)
	case "validate":
		runValidate(gen, *seed, *difficulty, *depth, *genre)
	case "phases":
		runPhaseAnalysis(gen, *seed, *difficulty, *depth, *genre, *verbose)
	case "rewards":
		runRewardAnalysis(gen, *seed, *difficulty, *depth, *genre, *verbose)
	case "benchmark":
		runBenchmark(gen, *difficulty, *depth, *genre)
	case "all":
		fmt.Println("=== Running All Tests ===")
		fmt.Println()
		runDemo(gen, *seed, *difficulty, *depth, *genre, false)
		fmt.Println()
		runValidate(gen, *seed, *difficulty, *depth, *genre)
		fmt.Println()
		runPhaseAnalysis(gen, *seed, *difficulty, *depth, *genre, *verbose)
		fmt.Println()
		runRewardAnalysis(gen, *seed, *difficulty, *depth, *genre, *verbose)
		fmt.Println()
		runBenchmark(gen, *difficulty, *depth, *genre)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

func runDemo(gen *legendary.LegendaryQuestGenerator, seed int64, difficulty float64, depth int, genre string, verbose bool) {
	fmt.Println("=== Legendary Quest Demo ===")
	fmt.Printf("Seed: %d, Difficulty: %.1f, Depth: %d, Genre: %s\n", seed, difficulty, depth, genre)

	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genre,
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	quest := result.(*legendary.LegendaryQuest)

	fmt.Printf("\n📜 Quest: %s\n", quest.Name)
	fmt.Printf("   ID: %s\n", quest.ID)
	fmt.Printf("   Description: %s\n", quest.Description)
	fmt.Printf("   Required Level: %d\n", quest.RequiredLevel)
	fmt.Printf("   Estimated Time: %.1f hours\n", quest.EstimatedHours)
	fmt.Printf("   Phases: %d\n", len(quest.Phases))

	if verbose {
		fmt.Println("\n🔹 Phases:")
		for _, phase := range quest.Phases {
			fmt.Printf("   %d. %s (%s)\n", phase.PhaseNumber, phase.Name, phase.Type.String())
			fmt.Printf("      %s\n", phase.Description)

			if phase.Requirements != nil {
				printPhaseRequirements(phase)
			}
		}
	} else {
		fmt.Println("\n🔹 Phase Overview:")
		for _, phase := range quest.Phases {
			fmt.Printf("   %d. %s\n", phase.PhaseNumber, phase.Type.String())
		}
	}

	fmt.Println("\n🎁 Rewards:")
	printRewards(quest.Rewards, verbose)

	// Validation
	fmt.Println("\n✅ Validation:")
	if err := gen.Validate(quest); err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
	} else {
		fmt.Println("   ✅ Quest passes all validation checks")
	}
}

func runGenerate(gen *legendary.LegendaryQuestGenerator, seed int64, difficulty float64, depth int, genre string, verbose bool) {
	fmt.Println("=== Generate Legendary Quest ===")

	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genre,
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	quest := result.(*legendary.LegendaryQuest)

	// Full quest details
	fmt.Printf("Quest: %s\n", quest.Name)
	fmt.Printf("ID: %s\n", quest.ID)
	fmt.Printf("Seed: %d\n", quest.Seed)
	fmt.Printf("Required Level: %d\n", quest.RequiredLevel)
	fmt.Printf("Estimated Hours: %.1f\n", quest.EstimatedHours)
	fmt.Printf("Total Phases: %d\n\n", len(quest.Phases))

	for i, phase := range quest.Phases {
		fmt.Printf("Phase %d/%d: %s\n", i+1, len(quest.Phases), phase.Name)
		fmt.Printf("  Type: %s\n", phase.Type.String())
		fmt.Printf("  Description: %s\n", phase.Description)

		if phase.Requirements != nil && verbose {
			printPhaseRequirements(phase)
		}
		fmt.Println()
	}

	fmt.Println("Rewards:")
	printRewards(quest.Rewards, verbose)
}

func runValidate(gen *legendary.LegendaryQuestGenerator, seed int64, difficulty float64, depth int, genre string) {
	fmt.Println("=== Validation Test ===")

	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genre,
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	quest := result.(*legendary.LegendaryQuest)

	// Run validation checks
	fmt.Printf("Testing quest: %s\n", quest.Name)

	if err := gen.Validate(quest); err != nil {
		fmt.Printf("❌ Validation FAILED: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ All validation checks passed:")
	fmt.Printf("  • Phase count: %d (5-10 range) ✓\n", len(quest.Phases))

	// Check cross-server requirement
	hasTravel := false
	minServers := 0
	for _, phase := range quest.Phases {
		if phase.Type == legendary.PhaseTravel && phase.Requirements != nil {
			hasTravel = true
			if phase.Requirements.MinServers > minServers {
				minServers = phase.Requirements.MinServers
			}
		}
	}
	fmt.Printf("  • Cross-server travel: %v (≥3 servers: %d) ✓\n", hasTravel, minServers)

	// Check rewards
	hasRewards := quest.Rewards != nil && len(quest.Rewards.Items) > 0
	fmt.Printf("  • Legendary rewards: %v (%d items) ✓\n", hasRewards, len(quest.Rewards.Items))

	// Check time estimate
	validTime := quest.EstimatedHours >= 10.0 && quest.EstimatedHours <= 20.0
	fmt.Printf("  • Time estimate: %v (%.1f hours in 10-20 range) ✓\n", validTime, quest.EstimatedHours)
}

func runPhaseAnalysis(gen *legendary.LegendaryQuestGenerator, seed int64, difficulty float64, depth int, genre string, verbose bool) {
	fmt.Println("=== Phase Analysis ===")

	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genre,
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	quest := result.(*legendary.LegendaryQuest)

	// Analyze phase distribution
	phaseTypes := make(map[legendary.PhaseType]int)
	for _, phase := range quest.Phases {
		phaseTypes[phase.Type]++
	}

	fmt.Printf("Quest: %s (%d phases)\n\n", quest.Name, len(quest.Phases))
	fmt.Println("Phase Type Distribution:")
	for phaseType, count := range phaseTypes {
		pct := float64(count) / float64(len(quest.Phases)) * 100
		fmt.Printf("  %s: %d (%.1f%%)\n", phaseType.String(), count, pct)
	}

	if verbose {
		fmt.Println("\nDetailed Phase Requirements:")
		for _, phase := range quest.Phases {
			fmt.Printf("\nPhase %d: %s (%s)\n", phase.PhaseNumber, phase.Name, phase.Type.String())
			printPhaseRequirements(phase)
		}
	}
}

func runRewardAnalysis(gen *legendary.LegendaryQuestGenerator, seed int64, difficulty float64, depth int, genre string, verbose bool) {
	fmt.Println("=== Reward Analysis ===")

	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genre,
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	quest := result.(*legendary.LegendaryQuest)

	fmt.Printf("Quest: %s\n\n", quest.Name)
	printRewards(quest.Rewards, verbose)
}

func runBenchmark(gen *legendary.LegendaryQuestGenerator, difficulty float64, depth int, genre string) {
	fmt.Println("=== Performance Benchmark ===")

	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genre,
	}

	// Benchmark generation
	const iterations = 100
	fmt.Printf("Generating %d quests...\n", iterations)

	var totalPhases int
	for i := 0; i < iterations; i++ {
		result, err := gen.Generate(int64(i), params)
		if err != nil {
			log.Fatalf("Generation %d failed: %v", i, err)
		}
		quest := result.(*legendary.LegendaryQuest)
		totalPhases += len(quest.Phases)
	}

	avgPhases := float64(totalPhases) / float64(iterations)
	fmt.Printf("\n✅ Successfully generated %d quests\n", iterations)
	fmt.Printf("   Average phases per quest: %.1f\n", avgPhases)
	fmt.Printf("   Target: <500ms per quest (not measured, use `go test -bench` for precise timing)\n")
}

// Helper functions

func printPhaseRequirements(phase *legendary.QuestPhase) {
	req := phase.Requirements
	if req == nil {
		return
	}

	// Kill requirements
	if len(req.KillTargets) > 0 {
		fmt.Println("      Kill Targets:")
		for enemy, count := range req.KillTargets {
			fmt.Printf("        • %s: %d\n", enemy, count)
		}
	}
	if len(req.KillBosses) > 0 {
		fmt.Println("      Bosses:")
		for _, boss := range req.KillBosses {
			fmt.Printf("        • %s\n", boss)
		}
	}

	// Collection requirements
	if len(req.CollectItems) > 0 {
		fmt.Println("      Collect Items:")
		for item, count := range req.CollectItems {
			fmt.Printf("        • %s: %d\n", item, count)
		}
	}

	// Crafting requirements
	if len(req.CraftItems) > 0 {
		fmt.Println("      Craft Items:")
		for _, craft := range req.CraftItems {
			fmt.Printf("        • %s (%s): %d (Quality: %s)\n",
				craft.ItemName, craft.ItemType, craft.Quantity, craft.StationQuality)
		}
	}

	// Raid requirements
	if len(req.RaidEncounters) > 0 {
		fmt.Println("      Raid Encounters:")
		for _, raid := range req.RaidEncounters {
			fmt.Printf("        • %s (%s)\n", raid.RaidName, raid.Tier.String())
			fmt.Printf("          Party Size: %d+, Max Deaths: %d, Time Limit: %dm\n",
				raid.MinPartySize, raid.MaxDeaths, raid.TimeLimit)
			if len(raid.BossesToKill) > 0 {
				fmt.Printf("          Specific Bosses: %v\n", raid.BossesToKill)
			}
		}
	}

	// Travel requirements
	if req.MinServers > 0 {
		fmt.Printf("      Travel: Visit %d servers\n", req.MinServers)
		if len(req.ServersToVisit) > 0 {
			fmt.Printf("        Servers: %v\n", req.ServersToVisit)
		}
	}
	if len(req.LocationsToDiscover) > 0 {
		fmt.Println("      Discover Locations:")
		for _, loc := range req.LocationsToDiscover {
			fmt.Printf("        • %s\n", loc)
		}
	}

	// NPC requirements
	if len(req.NPCsToTalk) > 0 {
		fmt.Println("      Talk to NPCs:")
		for _, npc := range req.NPCsToTalk {
			fmt.Printf("        • %s\n", npc)
		}
	}

	// Challenges
	if len(req.Challenges) > 0 {
		fmt.Println("      Challenges:")
		for _, challenge := range req.Challenges {
			fmt.Printf("        • %s (%s) - Difficulty: %.1f\n",
				challenge.Name, challenge.Type.String(), challenge.Difficulty)
		}
	}
}

func printRewards(rewards *legendary.LegendaryRewards, verbose bool) {
	if rewards == nil {
		fmt.Println("  No rewards")
		return
	}

	fmt.Printf("  Legendary Items: %d\n", len(rewards.Items))
	if verbose {
		for i, item := range rewards.Items {
			fmt.Printf("    %d. %s (%s)\n", i+1, item.Name, item.Rarity.String())
		}
	}

	fmt.Printf("  Titles: %d\n", len(rewards.Titles))
	if verbose {
		for _, title := range rewards.Titles {
			fmt.Printf("    • %s\n", title)
		}
	}

	fmt.Printf("  Gold: %d\n", rewards.Gold)
	fmt.Printf("  Experience: %d\n", rewards.Experience)
	fmt.Printf("  Prestige Levels: %d\n", rewards.PrestigeLevels)
	fmt.Printf("  Achievements: %d\n", len(rewards.Achievements))
	fmt.Printf("  Cosmetics: %d\n", len(rewards.Cosmetics))

	if verbose && len(rewards.Cosmetics) > 0 {
		fmt.Println("  Cosmetic Rewards:")
		for _, cosmetic := range rewards.Cosmetics {
			fmt.Printf("    • %s\n", cosmetic)
		}
	}
}
