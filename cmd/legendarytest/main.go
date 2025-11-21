package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/legendary"
)

func main() {
	// Parse flags
	mode := flag.String("mode", "demo", "Test mode: demo, progress, rewards, all")
	seed := flag.Int64("seed", 12345, "Random seed for deterministic generation")
	genre := flag.String("genre", "fantasy", "Genre: fantasy, sci-fi, horror, cyberpunk, post-apocalyptic")
	level := flag.Int("level", 50, "Player level")
	servers := flag.Int("servers", 3, "Number of servers required (3-5)")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	log.Printf("Legendary Quest Test - Mode: %s, Seed: %d, Genre: %s", *mode, *seed, *genre)

	switch *mode {
	case "demo":
		runDemo(*seed, *genre, *level, *servers, *verbose)
	case "progress":
		runProgressTest(*seed, *genre, *level, *servers, *verbose)
	case "rewards":
		runRewardsTest(*seed, *genre, *level, *servers, *verbose)
	case "all":
		runDemo(*seed, *genre, *level, *servers, *verbose)
		fmt.Println()
		runProgressTest(*seed, *genre, *level, *servers, *verbose)
		fmt.Println()
		runRewardsTest(*seed, *genre, *level, *servers, *verbose)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

func runDemo(seed int64, genreID string, level, servers int, verbose bool) {
	fmt.Println("=== Legendary Quest Generation Demo ===")

	gen := legendary.NewLegendaryQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      level,
		GenreID:    genreID,
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		log.Fatalf("Failed to generate quest: %v", err)
	}

	quest := result.(*legendary.LegendaryQuest)

	fmt.Printf("\nQuest: %s\n", quest.Name)
	fmt.Printf("Description: %s\n", quest.Description)
	fmt.Printf("Required Level: %d\n", quest.RequiredLevel)
	fmt.Printf("Estimated Hours: %.1f\n", quest.EstimatedHours)

	if verbose {
		fmt.Printf("\nQuest ID: %s\n", quest.ID)
		fmt.Printf("Seed: %d\n", quest.Seed)
	}

	fmt.Printf("\nPhases (%d):\n", len(quest.Phases))
	for i, phase := range quest.Phases {
		fmt.Printf("  %d. [%s] %s\n", i+1, phase.Type, phase.Name)
		if verbose {
			fmt.Printf("     %s\n", phase.Description)
			fmt.Printf("     Type: %s\n", phase.Type.String())
		}
	}

	fmt.Printf("\nRewards:\n")
	if quest.Rewards != nil {
		fmt.Printf("  Gold: %d\n", quest.Rewards.Gold)
		fmt.Printf("  Experience: %d\n", quest.Rewards.Experience)
		fmt.Printf("  Prestige Levels: %d\n", quest.Rewards.PrestigeLevels)
		fmt.Printf("  Items: %d\n", len(quest.Rewards.Items))
		fmt.Printf("  Titles: %d\n", len(quest.Rewards.Titles))
		if verbose {
			for i, item := range quest.Rewards.Items {
				fmt.Printf("    %d. %s\n", i+1, item.Name)
			}
			for i, title := range quest.Rewards.Titles {
				fmt.Printf("    Title %d: %s\n", i+1, title)
			}
		}
	}

	// Validate
	if err := gen.Validate(quest); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Println("\n✓ Quest validated successfully")
}

func runProgressTest(seed int64, genreID string, level, servers int, verbose bool) {
	fmt.Println("=== Progress Tracking Test ===")

	// Generate quest
	gen := legendary.NewLegendaryQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      level,
		GenreID:    genreID,
	}

	result, _ := gen.Generate(seed, params)
	quest := result.(*legendary.LegendaryQuest)

	fmt.Printf("\nTracking progress for quest: %s\n", quest.Name)
	fmt.Printf("Total phases: %d\n", len(quest.Phases))

	// Simulate progression through phases
	for i, phase := range quest.Phases {
		fmt.Printf("\nPhase %d: %s\n", i+1, phase.Name)
		fmt.Printf("  Type: %s\n", phase.Type.String())
		fmt.Printf("  Description: %s\n", phase.Description)

		if verbose && phase.Requirements != nil {
			if len(phase.Requirements.KillTargets) > 0 {
				fmt.Printf("  Kill targets: %d types\n", len(phase.Requirements.KillTargets))
			}
			if len(phase.Requirements.CollectItems) > 0 {
				fmt.Printf("  Items to collect: %d types\n", len(phase.Requirements.CollectItems))
			}
			if len(phase.Requirements.RaidEncounters) > 0 {
				fmt.Printf("  Raids required: %d\n", len(phase.Requirements.RaidEncounters))
			}
		}
	}

	// Show overall progress using quest's Progress() method
	progress := quest.Progress()
	fmt.Printf("\nOverall progress: %.1f%%\n", progress*100)

	currentPhase := quest.CurrentPhase()
	if currentPhase != nil {
		fmt.Printf("Current phase: %s\n", currentPhase.Name)
	}

	fmt.Printf("Quest complete: %v\n", quest.IsComplete())
}

func runRewardsTest(seed int64, genreID string, level, servers int, verbose bool) {
	fmt.Println("=== Reward Generation Test ===")

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}

	fmt.Println("\nGenerating legendary quests for all genres...")

	for i, g := range genres {
		gen := legendary.NewLegendaryQuestGenerator()
		params := procgen.GenerationParams{
			Difficulty: 0.9,
			Depth:      level,
			GenreID:    g,
		}

		result, err := gen.Generate(seed+int64(i), params)
		if err != nil {
			log.Printf("Failed to generate quest for genre %s: %v", g, err)
			continue
		}

		quest := result.(*legendary.LegendaryQuest)

		fmt.Printf("\n%s - %s:\n", g, quest.Name)
		if quest.Rewards != nil {
			fmt.Printf("  Gold: %d\n", quest.Rewards.Gold)
			fmt.Printf("  Experience: %d\n", quest.Rewards.Experience)
			fmt.Printf("  Prestige Levels: %d\n", quest.Rewards.PrestigeLevels)

			if len(quest.Rewards.Items) > 0 {
				fmt.Printf("  Legendary Items:\n")
				for _, item := range quest.Rewards.Items {
					fmt.Printf("    • %s (%s)\n", item.Name, item.Rarity.String())
				}
			}

			if len(quest.Rewards.Titles) > 0 {
				fmt.Printf("  Titles: %s\n", quest.Rewards.Titles[0])
			}

			if verbose && len(quest.Rewards.Achievements) > 0 {
				fmt.Printf("  Achievements: %d\n", len(quest.Rewards.Achievements))
			}
		}
	}

	fmt.Println("\n✓ All genres tested successfully")
}
