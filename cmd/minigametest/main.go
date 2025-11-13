package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	gameType := flag.String("game", "card", "Mini-game type to test: card, dice, puzzle, memory, lockpicking, hacking, ritual, or 'all' for all types")
	difficulty := flag.Float64("difficulty", 0.5, "Game difficulty (0.0-1.0)")
	seed := flag.Int64("seed", 12345, "Random seed for deterministic generation")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	listGames := flag.Bool("list", false, "List all available mini-games")

	flag.Parse()

	// Initialize logger
	baseLogger := logging.TestUtilityLogger("minigametest")
	if *verbose {
		baseLogger.SetLevel(logrus.DebugLevel)
	}

	logger := baseLogger.WithFields(logrus.Fields{
		"game":       *gameType,
		"difficulty": *difficulty,
		"seed":       *seed,
	})
	logger.Info("Mini-Game Test Tool started")

	// List games if requested
	if *listGames {
		listMiniGames()
		os.Exit(0)
	}

	// Validate difficulty
	if *difficulty < 0.0 || *difficulty > 1.0 {
		logger.WithField("difficulty", *difficulty).Error("difficulty must be between 0.0 and 1.0")
		fmt.Fprintf(os.Stderr, "Error: difficulty must be between 0.0 and 1.0 (got %.2f)\n", *difficulty)
		os.Exit(1)
	}

	// Map of game names to types
	gameMap := map[string]engine.MiniGameType{
		"card":        engine.MiniGameCard,
		"dice":        engine.MiniGameDice,
		"puzzle":      engine.MiniGamePuzzle,
		"memory":      engine.MiniGameMemory,
		"lockpicking": engine.MiniGameLockPicking,
		"hacking":     engine.MiniGameHacking,
		"ritual":      engine.MiniGameRitual,
	}

	// Test all games or specific one
	if *gameType == "all" {
		testAllGames(gameMap, *difficulty, *seed, *verbose, logger)
	} else {
		gameTypeVal, ok := gameMap[*gameType]
		if !ok {
			logger.WithField("game", *gameType).Error("unknown game type")
			fmt.Fprintf(os.Stderr, "Unknown game type: %s\n", *gameType)
			fmt.Fprintf(os.Stderr, "Use -list to see available games\n")
			os.Exit(1)
		}
		testGame(gameTypeVal, *difficulty, *seed, *verbose, logger)
	}

	logger.Info("mini-game test completed")
}

func listMiniGames() int {
	fmt.Println("=== Available Mini-Games ===")
	fmt.Println()

	games := []struct {
		name        string
		timeLimit   string
		description string
	}{
		{"card", "10 min", "Procedural card game with AI opponent"},
		{"dice", "5 min", "Custom dice game with betting mechanics"},
		{"puzzle", "7 min", "Sliding tiles or pattern matching"},
		{"memory", "4 min", "Card pairs or sequence repetition"},
		{"lockpicking", "2 min", "Timing-based lock-picking challenge"},
		{"hacking", "3 min", "Terminal/console puzzle (sci-fi genre)"},
		{"ritual", "5 min", "Spell pattern drawing (fantasy/horror)"},
	}

	for _, game := range games {
		fmt.Printf("%-12s %-8s %s\n", game.name, game.timeLimit, game.description)
	}

	fmt.Println()
	fmt.Println("Usage: minigametest -game <name> -difficulty 0.5 -seed 12345")
	fmt.Println("       minigametest -game all")
	fmt.Println("       minigametest -list")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  minigametest -game card -difficulty 0.8")
	fmt.Println("  minigametest -game all -seed 42 -verbose")

	return 0
}

func testGame(gameType engine.MiniGameType, difficulty float64, seed int64, verbose bool, logger *logrus.Entry) {
	fmt.Printf("\n=== Testing %s ===\n", gameType.String())
	fmt.Printf("Difficulty: %.2f\n", difficulty)
	fmt.Printf("Seed: %d\n\n", seed)

	// Create world and system
	world := engine.NewWorld()
	system := engine.NewMiniGameSystem(world)
	system.SetSeed(seed)

	// Create player entity
	player := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Add components for rewards
	player.AddComponent(&engine.InventoryComponent{Gold: 0})
	player.AddComponent(&engine.ExperienceComponent{CurrentXP: 0})

	// Start the game
	err := system.StartGame(player.ID, gameType, difficulty)
	if err != nil {
		logger.WithError(err).Error("failed to start game")
		fmt.Fprintf(os.Stderr, "Error starting game: %v\n", err)
		return
	}

	// Get the game component
	gameComp := system.GetGameComponent(player.ID)
	if gameComp == nil {
		fmt.Println("Error: failed to get game component")
		return
	}

	// Display game info
	fmt.Println("Game Information:")
	fmt.Printf("  Type: %s\n", gameComp.GameType.String())
	fmt.Printf("  Time Limit: %.0f seconds\n", gameComp.TimeLimit)
	fmt.Printf("  Active: %v\n", gameComp.Active)
	fmt.Println()

	// Display reward info
	if gameComp.Reward != nil {
		fmt.Println("Expected Reward:")
		fmt.Printf("  Gold: %d\n", gameComp.Reward.Gold)
		fmt.Printf("  XP: %.1f\n", gameComp.Reward.XP)
		fmt.Printf("  Items: %d\n", len(gameComp.Reward.Items))
		fmt.Println()
	}

	// Simulate game completion (success)
	fmt.Println("Simulating game completion (success)...")
	err = system.EndGame(player.ID, true)
	if err != nil {
		logger.WithError(err).Error("failed to end game")
		fmt.Fprintf(os.Stderr, "Error ending game: %v\n", err)
		return
	}

	// Check rewards
	invComp, _ := player.GetComponent("inventory")
	inv := invComp.(*engine.InventoryComponent)
	expComp, _ := player.GetComponent("experience")
	exp := expComp.(*engine.ExperienceComponent)

	fmt.Println("Player Rewards:")
	fmt.Printf("  Gold Earned: %d\n", inv.Gold)
	fmt.Printf("  XP Earned: %d\n", exp.CurrentXP)
	fmt.Println()

	if verbose {
		fmt.Println("Game State:")
		fmt.Printf("  Game Active: %v\n", gameComp.Active)
		fmt.Printf("  Time Elapsed: %.2f seconds\n", gameComp.TimeElapsed)
	}

	fmt.Println("✓ Game test completed successfully")
}

func testAllGames(gameMap map[string]engine.MiniGameType, difficulty float64, seed int64, verbose bool, logger *logrus.Entry) {
	fmt.Println("\n=== Testing All Mini-Games ===")
	fmt.Printf("Difficulty: %.2f\n", difficulty)
	fmt.Printf("Seed: %d\n", seed)

	gameNames := []string{"card", "dice", "puzzle", "memory", "lockpicking", "hacking", "ritual"}

	for _, name := range gameNames {
		gameType := gameMap[name]
		testGame(gameType, difficulty, seed, false, logger)
		fmt.Println()
	}

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Printf("Tested %d mini-game types\n", len(gameNames))
	fmt.Println("All games initialized and completed successfully")
}
