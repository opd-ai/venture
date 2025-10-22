package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

func main() {
	// Command line flags
	genre := flag.String("genre", "fantasy", "Genre for items (fantasy, scifi)")
	count := flag.Int("count", 20, "Number of items to generate")
	depth := flag.Int("depth", 5, "Dungeon depth (affects item level and rarity)")
	itemType := flag.String("type", "", "Filter by item type (weapon, armor, consumable)")
	seed := flag.Int64("seed", 0, "Random seed (0 for random)")
	verbose := flag.Bool("verbose", false, "Show detailed item information")
	output := flag.String("output", "", "Output file (default: stdout)")

	flag.Parse()

	// Generate seed if not provided
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	// Create generator
	generator := item.NewItemGenerator()

	// Set up parameters
	params := procgen.GenerationParams{
		Depth:      *depth,
		Difficulty: 0.5,
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"count": *count,
		},
	}

	// Add type filter if specified
	if *itemType != "" {
		params.Custom["type"] = *itemType
	}

	// Generate items
	result, err := generator.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Failed to generate items: %v", err)
	}

	items := result.([]*item.Item)

	// Validate items
	if err := generator.Validate(items); err != nil {
		log.Fatalf("Generated items failed validation: %v", err)
	}

	// Prepare output
	var outputFile *os.File
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer f.Close()
		outputFile = f
	} else {
		outputFile = os.Stdout
	}

	// Print header
	fmt.Fprintf(outputFile, "Item Generator Test - Genre: %s, Depth: %d, Seed: %d\n", *genre, *depth, *seed)
	fmt.Fprintf(outputFile, "Generated %d items\n", len(items))
	fmt.Fprintf(outputFile, "%s\n\n", separator(80))

	// Print items
	rarityCounts := make(map[item.Rarity]int)
	typeCounts := make(map[item.ItemType]int)

	for i, itm := range items {
		rarityCounts[itm.Rarity]++
		typeCounts[itm.Type]++

		printItem(outputFile, i+1, itm, *verbose)
		fmt.Fprintf(outputFile, "\n")
	}

	// Print statistics
	fmt.Fprintf(outputFile, "%s\n", separator(80))
	fmt.Fprintf(outputFile, "Statistics:\n\n")

	fmt.Fprintf(outputFile, "Rarity Distribution:\n")
	for _, rarity := range []item.Rarity{
		item.RarityCommon,
		item.RarityUncommon,
		item.RarityRare,
		item.RarityEpic,
		item.RarityLegendary,
	} {
		count := rarityCounts[rarity]
		percentage := float64(count) / float64(len(items)) * 100
		fmt.Fprintf(outputFile, "  %-12s: %3d (%5.1f%%) %s\n",
			rarity.String(), count, percentage, bar(percentage, 50))
	}

	fmt.Fprintf(outputFile, "\nType Distribution:\n")
	for _, itemType := range []item.ItemType{
		item.TypeWeapon,
		item.TypeArmor,
		item.TypeConsumable,
		item.TypeAccessory,
	} {
		count := typeCounts[itemType]
		percentage := float64(count) / float64(len(items)) * 100
		fmt.Fprintf(outputFile, "  %-12s: %3d (%5.1f%%) %s\n",
			itemType.String(), count, percentage, bar(percentage, 50))
	}

	// Calculate average stats
	totalDamage := 0
	totalDefense := 0
	weaponCount := 0
	armorCount := 0

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
		fmt.Fprintf(outputFile, "\nAverage Stats:\n")
		if weaponCount > 0 {
			fmt.Fprintf(outputFile, "  Weapon Damage: %.1f\n", float64(totalDamage)/float64(weaponCount))
		}
		if armorCount > 0 {
			fmt.Fprintf(outputFile, "  Armor Defense: %.1f\n", float64(totalDefense)/float64(armorCount))
		}
	}

	fmt.Fprintf(outputFile, "\nValidation: PASSED\n")
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
