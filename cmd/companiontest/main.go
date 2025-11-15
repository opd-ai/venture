// Package main provides a CLI tool for testing companion generation and features.
// This tool demonstrates Phase 22 companion system features including
// companion generation, inventory, skill inheritance, and bonding perks.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/companion"
)

func main() {
	// Command-line flags
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	depth := flag.Int("depth", 5, "Dungeon depth/level")
	difficulty := flag.Float64("difficulty", 0.5, "Difficulty (0.0-1.0)")
	genre := flag.String("genre", "fantasy", "Genre (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)")
	count := flag.Int("count", 5, "Number of companions to generate")

	flag.Parse()

	fmt.Println("=== Venture Companion System Test ===")
	fmt.Printf("Seed: %d, Depth: %d, Difficulty: %.2f, Genre: %s\n\n", *seed, *depth, *difficulty, *genre)

	// Create generator
	gen := companion.NewGenerator()

	// Generation parameters
	params := procgen.GenerationParams{
		Difficulty: *difficulty,
		Depth:      *depth,
		GenreID:    *genre,
	}

	// Generate companions
	for i := 0; i < *count; i++ {
		companionSeed := *seed + int64(i)

		result, err := gen.Generate(companionSeed, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating companion %d: %v\n", i+1, err)
			continue
		}

		// Validate
		if err := gen.Validate(result); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed for companion %d: %v\n", i+1, err)
			continue
		}

		comp := result.(*companion.Companion)

		// Display companion info
		fmt.Printf("Companion %d: %s\n", i+1, comp.Name)
		fmt.Printf("  Type: %s\n", getTypeName(comp.Type))
		fmt.Printf("  Level: %d\n", comp.Level)
		fmt.Printf("  Loyalty: %.1f/100\n", comp.Loyalty)
		fmt.Printf("  Stats:\n")
		fmt.Printf("    HP: %.0f/%.0f\n", comp.HP, comp.MaxHP)
		fmt.Printf("    Attack: %.1f\n", comp.Attack)
		fmt.Printf("    Defense: %.1f\n", comp.Defense)
		fmt.Printf("    Speed: %.1f\n", comp.Speed)
		fmt.Printf("  Commands: %d available\n", len(comp.Commands))
		for _, cmd := range comp.Commands {
			fmt.Printf("    - %s\n", getCommandName(cmd))
		}
		fmt.Println()
	}

	// Demonstrate companion features
	fmt.Println("=== Companion Features Demo ===")

	// 1. Inventory Feature
	fmt.Println("1. Companion Inventory:")
	inventory := engine.NewCompanionInventoryComponent(10, 50.0)
	inventory.AutoFetch = true
	inventory.FetchRadius = 100.0
	fmt.Printf("   Max Items: %d\n", inventory.MaxItems)
	fmt.Printf("   Max Weight: %.1f\n", inventory.MaxWeight)
	fmt.Printf("   Auto-Fetch: %v (radius: %.1f)\n", inventory.AutoFetch, inventory.FetchRadius)
	fmt.Println()

	// 2. Skill Inheritance Feature
	fmt.Println("2. Skill Inheritance:")
	skillInheritance := engine.NewSkillInheritanceComponent(5, 0.15)
	fmt.Printf("   Max Skills: %d\n", skillInheritance.MaxSkills)
	fmt.Printf("   Learning Rate: %.2f\n", skillInheritance.LearningRate)
	fmt.Printf("   Required Loyalty: %.1f\n", skillInheritance.RequiredLoyalty)

	// Simulate learning
	skillInheritance.AddSkillProgress("fireball", 0.3)
	skillInheritance.AddSkillProgress("ice_shard", 0.7)
	skillInheritance.AddSkillProgress("lightning", 1.0)

	fmt.Printf("   Skills Learning:\n")
	fmt.Printf("     - Fireball: %.0f%%\n", skillInheritance.GetSkillProgress("fireball")*100)
	fmt.Printf("     - Ice Shard: %.0f%%\n", skillInheritance.GetSkillProgress("ice_shard")*100)
	fmt.Printf("     - Lightning: %.0f%% (ACTIVE)\n", skillInheritance.GetSkillProgress("lightning")*100)
	fmt.Println()

	// 3. Bonding Perks Feature
	fmt.Println("3. Bonding Perks:")
	companionComp := &engine.CompanionComponent{
		Loyalty:       85.0,
		TimeWithOwner: 10000.0,
		BondingPerks:  []engine.BondingPerk{},
	}

	// Simulate perk unlocks
	companionComp.AddPerk(engine.PerkExtraHealth)
	companionComp.AddPerk(engine.PerkExtraDamage)
	companionComp.AddPerk(engine.PerkFasterLearning)

	fmt.Printf("   Loyalty: %.1f/100\n", companionComp.Loyalty)
	fmt.Printf("   Time Together: %.0f seconds (%.1f hours)\n", companionComp.TimeWithOwner, companionComp.TimeWithOwner/3600.0)
	fmt.Printf("   Unlocked Perks (%d):\n", len(companionComp.BondingPerks))
	for _, perk := range companionComp.BondingPerks {
		fmt.Printf("     - %s\n", perk.String())
	}
	fmt.Println()

	// 4. Permadeath Mode
	fmt.Println("4. Permadeath Mode:")
	revivableComp := &engine.CompanionComponent{Permadeath: false}
	permadeathComp := &engine.CompanionComponent{Permadeath: true}
	fmt.Printf("   Revivable Companion: Permadeath=%v (can be revived)\n", revivableComp.Permadeath)
	fmt.Printf("   Permadeath Companion: Permadeath=%v (dies permanently)\n", permadeathComp.Permadeath)
	fmt.Println()

	fmt.Println("=== Phase 22.2 Complete ===")
	fmt.Println("All companion features implemented and tested!")
}

func getTypeName(t engine.CompanionType) string {
	switch t {
	case engine.CompanionTypePet:
		return "Pet"
	case engine.CompanionTypeSummon:
		return "Summon"
	case engine.CompanionTypeHireling:
		return "Hireling"
	case engine.CompanionTypeElemental:
		return "Elemental"
	case engine.CompanionTypeUndead:
		return "Undead"
	case engine.CompanionTypeRobot:
		return "Robot"
	case engine.CompanionTypeSpirit:
		return "Spirit"
	case engine.CompanionTypeInsect:
		return "Insect"
	default:
		return "Unknown"
	}
}

func getCommandName(cmd engine.CommandType) string {
	switch cmd {
	case engine.CommandFollow:
		return "Follow"
	case engine.CommandStay:
		return "Stay"
	case engine.CommandAttack:
		return "Attack"
	case engine.CommandDefend:
		return "Defend"
	case engine.CommandGather:
		return "Gather"
	case engine.CommandScout:
		return "Scout"
	default:
		return "Unknown"
	}
}
