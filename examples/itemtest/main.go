package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

func main() {
	config := parseFlags()
	logger := initializeLogger(config)
	items := generateItems(config, logger)
	outputFile := prepareOutputFile(config, logger)
	defer outputFile.Close()

	printHeader(outputFile, config, len(items))
	rarityCounts, typeCounts := printItems(outputFile, items, config.verbose)
	printStatistics(outputFile, items, rarityCounts, typeCounts)
	fmt.Fprintf(outputFile, "\nValidation: PASSED\n")
}

type testConfig struct {
	genre    string
	count    int
	depth    int
	itemType string
	seed     int64
	verbose  bool
	output   string
}

// parseFlags parses command-line flags and returns configuration.
func parseFlags() testConfig {
	genre := flag.String("genre", "fantasy", "Genre for items (fantasy, scifi)")
	count := flag.Int("count", 20, "Number of items to generate")
	depth := flag.Int("depth", 5, "Dungeon depth (affects item level and rarity)")
	itemType := flag.String("type", "", "Filter by item type (weapon, armor, consumable)")
	seed := flag.Int64("seed", 0, "Random seed (0 for random)")
	verbose := flag.Bool("verbose", false, "Show detailed item information")
	output := flag.String("output", "", "Output file (default: stdout)")

	flag.Parse()

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	return testConfig{
		genre:    *genre,
		count:    *count,
		depth:    *depth,
		itemType: *itemType,
		seed:     *seed,
		verbose:  *verbose,
		output:   *output,
	}
}

// initializeLogger creates and configures the logger for the test utility.
func initializeLogger(config testConfig) *logrus.Entry {
	logger := logging.TestUtilityLogger("itemtest")
	testLogger := logger.WithFields(logrus.Fields{
		"genre": config.genre,
		"count": config.count,
		"depth": config.depth,
		"seed":  config.seed,
	})
	if config.itemType != "" {
		testLogger = testLogger.WithField("itemType", config.itemType)
	}
	testLogger.Info("generating items")
	return testLogger
}

// generateItems generates and validates items using the configured generator.
func generateItems(config testConfig, logger *logrus.Entry) []*item.Item {
	generator := item.NewItemGeneratorWithLogger(logger.Logger)

	params := procgen.GenerationParams{
		Depth:      config.depth,
		Difficulty: 0.5,
		GenreID:    config.genre,
		Custom: map[string]interface{}{
			"count": config.count,
		},
	}

	if config.itemType != "" {
		params.Custom["type"] = config.itemType
	}

	genLogger := logging.GeneratorLogger(logger.Logger, "item", config.seed, config.genre)
	genLogger.Debug("starting item generation")

	result, err := generator.Generate(config.seed, params)
	if err != nil {
		genLogger.WithError(err).Fatal("failed to generate items")
	}

	items := result.([]*item.Item)

	if err := generator.Validate(items); err != nil {
		genLogger.WithError(err).Fatal("generated items failed validation")
	}

	genLogger.WithField("itemCount", len(items)).Info("items generated successfully")
	return items
}

// prepareOutputFile creates the output file or returns stdout.
func prepareOutputFile(config testConfig, logger *logrus.Entry) *os.File {
	if config.output == "" {
		return os.Stdout
	}

	f, err := os.Create(config.output)
	if err != nil {
		logger.WithError(err).WithField("outputFile", config.output).Fatal("failed to create output file")
	}

	logger.WithField("outputFile", config.output).Info("writing items to file")
	return f
}

// printHeader prints the test header information.
func printHeader(w *os.File, config testConfig, itemCount int) {
	fmt.Fprintf(w, "Item Generator Test - Genre: %s, Depth: %d, Seed: %d\n", config.genre, config.depth, config.seed)
	fmt.Fprintf(w, "Generated %d items\n", itemCount)
	fmt.Fprintf(w, "%s\n\n", separator(80))
}

// printItems prints all items and returns rarity and type counts.
func printItems(w *os.File, items []*item.Item, verbose bool) (map[item.Rarity]int, map[item.ItemType]int) {
	rarityCounts := make(map[item.Rarity]int)
	typeCounts := make(map[item.ItemType]int)

	for i, itm := range items {
		rarityCounts[itm.Rarity]++
		typeCounts[itm.Type]++
		printItem(w, i+1, itm, verbose)
		fmt.Fprintf(w, "\n")
	}

	return rarityCounts, typeCounts
}

// printStatistics prints distribution and average stats.
func printStatistics(w *os.File, items []*item.Item, rarityCounts map[item.Rarity]int, typeCounts map[item.ItemType]int) {
	fmt.Fprintf(w, "%s\n", separator(80))
	fmt.Fprintf(w, "Statistics:\n\n")

	printRarityDistribution(w, rarityCounts, len(items))
	printTypeDistribution(w, typeCounts, len(items))
	printAverageStats(w, items)
}

// printRarityDistribution prints rarity distribution statistics.
func printRarityDistribution(w *os.File, rarityCounts map[item.Rarity]int, totalItems int) {
	fmt.Fprintf(w, "Rarity Distribution:\n")
	for _, rarity := range []item.Rarity{
		item.RarityCommon, item.RarityUncommon, item.RarityRare,
		item.RarityEpic, item.RarityLegendary,
	} {
		count := rarityCounts[rarity]
		percentage := float64(count) / float64(totalItems) * 100
		fmt.Fprintf(w, "  %-12s: %3d (%5.1f%%) %s\n",
			rarity.String(), count, percentage, bar(percentage, 50))
	}
}

// printTypeDistribution prints item type distribution statistics.
func printTypeDistribution(w *os.File, typeCounts map[item.ItemType]int, totalItems int) {
	fmt.Fprintf(w, "\nType Distribution:\n")
	for _, itemType := range []item.ItemType{
		item.TypeWeapon, item.TypeArmor,
		item.TypeConsumable, item.TypeAccessory,
	} {
		count := typeCounts[itemType]
		percentage := float64(count) / float64(totalItems) * 100
		fmt.Fprintf(w, "  %-12s: %3d (%5.1f%%) %s\n",
			itemType.String(), count, percentage, bar(percentage, 50))
	}
}

// printAverageStats calculates and prints average weapon/armor stats.
func printAverageStats(w *os.File, items []*item.Item) {
	totalDamage, totalDefense := 0, 0
	weaponCount, armorCount := 0, 0

	for _, itm := range items {
		if itm.Type == item.TypeWeapon {
			totalDamage += itm.Stats.Damage
			weaponCount++
		}
		if itm.Type == item.TypeArmor {
			totalDefense += itm.Stats.Defense
			armorCount++
		}
	}

	if weaponCount > 0 || armorCount > 0 {
		fmt.Fprintf(w, "\nAverage Stats:\n")
		if weaponCount > 0 {
			fmt.Fprintf(w, "  Weapon Damage: %.1f\n", float64(totalDamage)/float64(weaponCount))
		}
		if armorCount > 0 {
			fmt.Fprintf(w, "  Armor Defense: %.1f\n", float64(totalDefense)/float64(armorCount))
		}
	}
}

func printItem(w *os.File, index int, itm *item.Item, verbose bool) {
	// Header with index and name
	fmt.Fprintf(w, "[%d] %s\n", index, itm.Name)

	// Rarity with color indicator
	rarityIndicator := rarityToIndicator(itm.Rarity)
	fmt.Fprintf(w, "    Rarity:  %s %s\n", rarityIndicator, itm.Rarity.String())

	// Type information
	typeInfo := itm.Type.String()
	switch itm.Type {
	case item.TypeWeapon:
		typeInfo += fmt.Sprintf(" (%s)", itm.WeaponType.String())
	case item.TypeArmor:
		typeInfo += fmt.Sprintf(" (%s)", itm.ArmorType.String())
	case item.TypeConsumable:
		typeInfo += fmt.Sprintf(" (%s)", itm.ConsumableType.String())
	}
	fmt.Fprintf(w, "    Type:    %s\n", typeInfo)

	// Stats
	if itm.Type == item.TypeWeapon {
		fmt.Fprintf(w, "    Damage:  %d\n", itm.Stats.Damage)
		if itm.Stats.AttackSpeed > 0 {
			fmt.Fprintf(w, "    Speed:   %.2f\n", itm.Stats.AttackSpeed)
		}
	}
	if itm.Type == item.TypeArmor {
		fmt.Fprintf(w, "    Defense: %d\n", itm.Stats.Defense)
	}

	fmt.Fprintf(w, "    Value:   %d gold\n", itm.Stats.Value)
	fmt.Fprintf(w, "    Weight:  %.1f\n", itm.Stats.Weight)

	if itm.Stats.RequiredLevel > 1 {
		fmt.Fprintf(w, "    Level:   %d\n", itm.Stats.RequiredLevel)
	}

	if itm.Stats.DurabilityMax > 0 {
		fmt.Fprintf(w, "    Durability: %d/%d\n", itm.Stats.Durability, itm.Stats.DurabilityMax)
	}

	// Verbose information
	if verbose {
		if len(itm.Tags) > 0 {
			fmt.Fprintf(w, "    Tags:    %v\n", itm.Tags)
		}
		if itm.Description != "" {
			fmt.Fprintf(w, "    Desc:    %s\n", itm.Description)
		}
		fmt.Fprintf(w, "    Seed:    %d\n", itm.Seed)
	}
}

func rarityToIndicator(r item.Rarity) string {
	switch r {
	case item.RarityCommon:
		return "⚪"
	case item.RarityUncommon:
		return "🟢"
	case item.RarityRare:
		return "🔵"
	case item.RarityEpic:
		return "🟣"
	case item.RarityLegendary:
		return "🟠"
	default:
		return "  "
	}
}

func separator(width int) string {
	return strings.Repeat("=", width)
}

func bar(percentage float64, maxWidth int) string {
	filled := int(percentage / 100.0 * float64(maxWidth))
	var builder strings.Builder
	for i := 0; i < maxWidth; i++ {
		if i < filled {
			builder.WriteString("█")
		} else {
			builder.WriteString("░")
		}
	}
	return builder.String()
}

func init() {
	// Seed the default random source for any additional randomness
	rand.Seed(time.Now().UnixNano())
}
