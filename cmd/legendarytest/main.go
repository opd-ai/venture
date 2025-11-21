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

	gen := legendary.NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      level,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"player_level":    level,
			"servers_visited": servers,
		},
	}

	result, err := gen.Generate(seed, params)
	if err != nil {
		log.Fatalf("Failed to generate quest: %v", err)
	}

	quest := result.(*legendary.LegendaryQuest)

	fmt.Printf("\nQuest: %s\n", quest.Name)
	fmt.Printf("Description: %s\n", quest.Description)
	fmt.Printf("Min Level: %d\n", quest.MinLevel)
	fmt.Printf("Estimated Hours: %d\n", quest.EstimatedHours)
	fmt.Printf("Servers Required: %d\n", quest.ServersRequired)
	fmt.Printf("Raids Required: %d\n", len(quest.RaidsRequired))

	if verbose {
		fmt.Printf("\nLore:\n%s\n", quest.Lore)
	}

	fmt.Printf("\nPhases (%d):\n", len(quest.Phases))
	for i, phase := range quest.Phases {
		fmt.Printf("  %d. [%s] %s\n", i+1, phase.Type, phase.Name)
		if verbose {
			fmt.Printf("     %s\n", phase.Description)
			fmt.Printf("     Rewards: %d XP, %d Gold\n", phase.XPReward, phase.GoldReward)

			switch phase.Type {
			case legendary.PhaseExploration:
				fmt.Printf("     Server: %s, Location: (%d, %d)\n", phase.ServerID, phase.LocationX, phase.LocationY)
			case legendary.PhaseCombat:
				fmt.Printf("     Boss: %s, Kills: %d\n", phase.BossName, phase.KillCount)
			case legendary.PhaseCrafting:
				fmt.Printf("     Item: %s (Tier %d)\n", phase.ItemName, phase.StationTier)
			case legendary.PhaseCollection:
				fmt.Printf("     Materials: %d types\n", len(phase.MaterialIDs))
			case legendary.PhaseRaid:
				fmt.Printf("     Raid: %s (%s)\n", phase.RaidID, phase.RaidTier)
			}
		}
	}

	fmt.Printf("\nRewards (%d):\n", len(quest.Rewards))
	for i, reward := range quest.Rewards {
		fmt.Printf("  %d. [%s] %s\n", i+1, reward.Type, reward.Name)
		if verbose {
			fmt.Printf("     %s\n", reward.Description)
			if reward.Type == legendary.RewardAccountBonus {
				fmt.Printf("     Bonus: %.1f%%\n", reward.BonusPercent)
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
	gen := legendary.NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      level,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"player_level":    level,
			"servers_visited": servers,
		},
	}

	result, _ := gen.Generate(seed, params)
	quest := result.(*legendary.LegendaryQuest)

	// Create tracker
	tracker := legendary.NewProgressTracker()
	playerID := "test_player"

	fmt.Printf("\nTracking progress for quest: %s\n", quest.Name)

	// Simulate progression
	for i, phase := range quest.Phases {
		fmt.Printf("\nPhase %d: %s\n", i+1, phase.Name)

		// Update phase progress
		for progress := 0.0; progress <= 100.0; progress += 25.0 {
			tracker.UpdatePhase(quest.ID, playerID, i, progress)
			if verbose {
				p := tracker.GetProgress(quest.ID, playerID)
				fmt.Printf("  Progress: %.0f%% (Phase %d)\n", p.PhaseProgress, p.CurrentPhase)
			}
		}

		// Handle phase-specific tracking
		switch phase.Type {
		case legendary.PhaseExploration:
			tracker.AddServerVisited(quest.ID, playerID, phase.ServerID)
			if verbose {
				p := tracker.GetProgress(quest.ID, playerID)
				fmt.Printf("  Servers visited: %d/%d\n", len(p.ServersVisited), quest.ServersRequired)
			}

		case legendary.PhaseRaid:
			tracker.AddRaidCompleted(quest.ID, playerID, phase.RaidID)
			if verbose {
				p := tracker.GetProgress(quest.ID, playerID)
				fmt.Printf("  Raids completed: %d/%d\n", len(p.RaidsCompleted), len(quest.RaidsRequired))
			}

		case legendary.PhaseCollection:
			for j, matID := range phase.MaterialIDs {
				tracker.AddMaterial(quest.ID, playerID, matID, phase.Quantities[j])
			}
			if verbose {
				p := tracker.GetProgress(quest.ID, playerID)
				fmt.Printf("  Materials gathered: %d types\n", len(p.MaterialsGathered))
			}
		}
	}

	// Complete quest
	tracker.CompleteQuest(quest.ID, playerID)
	finalProgress := tracker.GetProgress(quest.ID, playerID)

	fmt.Printf("\n✓ Quest completed!\n")
	fmt.Printf("  Time taken: %s\n", finalProgress.CompletedAt.Sub(finalProgress.StartedAt))
	fmt.Printf("  Servers visited: %d\n", len(finalProgress.ServersVisited))
	fmt.Printf("  Raids completed: %d\n", len(finalProgress.RaidsCompleted))
	fmt.Printf("  Materials gathered: %d types\n", len(finalProgress.MaterialsGathered))
}

func runRewardsTest(seed int64, genreID string, level, servers int, verbose bool) {
	fmt.Println("=== Reward Generation Test ===")

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}

	fmt.Println("\nGenerating legendary quests for all genres...")

	for i, g := range genres {
		gen := legendary.NewGenerator()
		params := procgen.GenerationParams{
			Difficulty: 0.9,
			Depth:      level,
			GenreID:    g,
			Custom: map[string]interface{}{
				"player_level":    level,
				"servers_visited": servers,
			},
		}

		result, err := gen.Generate(seed+int64(i), params)
		if err != nil {
			log.Printf("Failed to generate quest for genre %s: %v", g, err)
			continue
		}

		quest := result.(*legendary.LegendaryQuest)

		fmt.Printf("\n%s - %s:\n", g, quest.Name)
		for _, reward := range quest.Rewards {
			fmt.Printf("  • [%s] %s", reward.Type, reward.Name)
			if reward.Type == legendary.RewardAccountBonus {
				fmt.Printf(" (+%.1f%%)", reward.BonusPercent)
			}
			fmt.Println()

			if verbose {
				fmt.Printf("    %s\n", reward.Description)
			}
		}
	}

	fmt.Println("\n✓ All genres tested successfully")
}
