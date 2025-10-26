package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/sirupsen/logrus"
)

var (
	genre        = flag.String("genre", "fantasy", "Genre: fantasy, scifi, horror, cyberpunk, or postapoc")
	count        = flag.Int("count", 3, "Number of merchants to generate")
	merchantType = flag.String("type", "fixed", "Merchant type: fixed or nomadic")
	depth        = flag.Int("depth", 1, "Depth level (affects inventory quality)")
	difficulty   = flag.Float64("difficulty", 0.5, "Difficulty multiplier (0.0-1.0)")
	seed         = flag.Int64("seed", 12345, "Generation seed")
	verbose      = flag.Bool("verbose", false, "Show detailed merchant information")
)

func main() {
	flag.Parse()

	// Initialize logger for test utility
	logger := logging.TestUtilityLogger("merchanttest")
	testLogger := logger.WithFields(logrus.Fields{
		"genre":        *genre,
		"count":        *count,
		"merchantType": *merchantType,
		"depth":        *depth,
		"difficulty":   *difficulty,
		"seed":         *seed,
	})

	testLogger.Info("generating merchants")

	// Parse merchant type
	var mType entity.MerchantType
	switch *merchantType {
	case "fixed":
		mType = entity.MerchantFixed
	case "nomadic":
		mType = entity.MerchantNomadic
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid merchant type '%s' (must be 'fixed' or 'nomadic')\n", *merchantType)
		os.Exit(1)
	}

	// Create generator with logger
	gen := entity.NewEntityGeneratorWithLogger(logger)

	// Set up generation parameters
	params := procgen.GenerationParams{
		Difficulty: *difficulty,
		Depth:      *depth,
		GenreID:    *genre,
	}

	// Generate merchants
	fmt.Printf("=== Merchant Generation Test ===\n")
	fmt.Printf("Genre: %s\n", *genre)
	fmt.Printf("Type: %s\n", *merchantType)
	fmt.Printf("Seed: %d\n", *seed)
	fmt.Printf("Count: %d\n\n", *count)

	for i := 0; i < *count; i++ {
		merchantSeed := *seed + int64(i)*1000

		merchant, err := gen.GenerateMerchant(merchantSeed, params, mType)
		if err != nil {
			testLogger.WithError(err).Errorf("failed to generate merchant %d", i)
			fmt.Fprintf(os.Stderr, "Error generating merchant %d: %v\n", i, err)
			continue
		}

		// Print merchant summary
		fmt.Printf("Merchant #%d: %s\n", i+1, merchant.Entity.Name)
		fmt.Printf("  Type: %s\n", merchant.MerchantType.String())
		fmt.Printf("  Level: %d\n", merchant.Entity.Stats.Level)
		fmt.Printf("  Health: %d/%d\n", merchant.Entity.Stats.Health, merchant.Entity.Stats.MaxHealth)
		fmt.Printf("  Defense: %d\n", merchant.Entity.Stats.Defense)
		fmt.Printf("  Speed: %.2f\n", merchant.Entity.Stats.Speed)
		fmt.Printf("  Price Multiplier: %.2fx\n", merchant.PriceMultiplier)
		fmt.Printf("  Buyback Rate: %.0f%%\n", merchant.BuyBackPercentage*100)
		fmt.Printf("  Inventory Size: %d items\n", len(merchant.Inventory))

		if *verbose {
			fmt.Printf("  Inventory:\n")
			for j, itm := range merchant.Inventory {
				if j >= 10 && !*verbose {
					fmt.Printf("    ... and %d more items\n", len(merchant.Inventory)-10)
					break
				}
				fmt.Printf("    - %s (Type: %s, Value: %d gold, Rarity: %s)\n",
					itm.Name, itm.Type.String(), itm.Stats.Value, itm.Rarity.String())
			}
		}

		fmt.Println()
	}

	// Generate spawn points if requested
	if *count > 0 {
		fmt.Printf("=== Spawn Points ===\n")
		worldWidth, worldHeight := 1000, 800
		spawnPoints := entity.GenerateMerchantSpawnPoints(*seed, worldWidth, worldHeight, mType, *count)

		for i, pt := range spawnPoints {
			fmt.Printf("Merchant #%d spawn: (%.1f, %.1f)\n", i+1, pt.X, pt.Y)
		}
		fmt.Println()
	}

	testLogger.WithField("merchantsGenerated", *count).Info("merchant generation complete")
	fmt.Println("Generation complete!")
}
