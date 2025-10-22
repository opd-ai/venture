// +build test

package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/genre"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
)

// GenreBlendingDemo demonstrates using blended genres for content generation.
// This shows how cross-genre blending creates unique hybrid content that combines
// themes, naming conventions, and characteristics from both base genres.
func main() {
	fmt.Println("=== Genre Blending Content Generation Demo ===")
	fmt.Println()

	// Create blender and registry
	registry := genre.DefaultRegistry()
	blender := genre.NewGenreBlender(registry)

	// Demonstrate multiple blend scenarios
	demonstrateSciFiHorror(blender)
	fmt.Println("\n" + separator(80) + "\n")
	demonstrateDarkFantasy(blender)
	fmt.Println("\n" + separator(80) + "\n")
	demonstrateCyberHorror(blender)
}

func demonstrateSciFiHorror(blender *genre.GenreBlender) {
	fmt.Println("Scenario 1: Sci-Fi Horror (Space Horror)")
	fmt.Println(separator(80))
	fmt.Println("Creating 'Alien' or 'Dead Space' style content...")
	fmt.Println()

	// Create sci-fi horror blend
	blended, err := blender.CreatePresetBlend("sci-fi-horror", 12345)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Blended Genre: %s\n", blended.Name)
	fmt.Printf("Description: %s\n", blended.Description)
	fmt.Printf("Themes: %v\n", blended.Themes)
	fmt.Println()

	// Generate entities with blended genre
	fmt.Println("Generated Entities:")
	generateEntities(blended.ID, 3)
	fmt.Println()

	// Generate items with blended genre
	fmt.Println("Generated Items:")
	generateItems(blended.ID, 3)
	fmt.Println()

	// Generate spells with blended genre
	fmt.Println("Generated Abilities:")
	generateSpells(blended.ID, 2)
}

func demonstrateDarkFantasy(blender *genre.GenreBlender) {
	fmt.Println("Scenario 2: Dark Fantasy")
	fmt.Println(separator(80))
	fmt.Println("Creating 'Dark Souls' or 'Bloodborne' style content...")
	fmt.Println()

	// Create dark fantasy blend (primarily fantasy with horror elements)
	blended, err := blender.CreatePresetBlend("dark-fantasy", 54321)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Blended Genre: %s\n", blended.Name)
	fmt.Printf("Description: %s\n", blended.Description)
	fmt.Printf("Themes: %v\n", blended.Themes)
	fmt.Println()

	// Generate entities
	fmt.Println("Generated Entities:")
	generateEntities(blended.ID, 3)
	fmt.Println()

	// Generate items
	fmt.Println("Generated Items:")
	generateItems(blended.ID, 3)
	fmt.Println()

	// Generate spells
	fmt.Println("Generated Abilities:")
	generateSpells(blended.ID, 2)
}

func demonstrateCyberHorror(blender *genre.GenreBlender) {
	fmt.Println("Scenario 3: Cyber Horror")
	fmt.Println(separator(80))
	fmt.Println("Creating cyberpunk-horror hybrid content...")
	fmt.Println()

	// Create cyber-horror blend
	blended, err := blender.CreatePresetBlend("cyber-horror", 99999)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Blended Genre: %s\n", blended.Name)
	fmt.Printf("Description: %s\n", blended.Description)
	fmt.Printf("Themes: %v\n", blended.Themes)
	fmt.Println()

	// Generate entities
	fmt.Println("Generated Entities:")
	generateEntities(blended.ID, 3)
	fmt.Println()

	// Generate items
	fmt.Println("Generated Items:")
	generateItems(blended.ID, 3)
	fmt.Println()

	// Generate spells
	fmt.Println("Generated Abilities:")
	generateSpells(blended.ID, 2)
}

func generateEntities(genreID string, count int) {
	gen := entity.NewEntityGenerator()
	params := procgen.GenerationParams{
		GenreID:    genreID,
		Difficulty: 0.5,
		Depth:      5,
		Custom: map[string]interface{}{
			"count": count,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		log.Printf("Entity generation failed: %v", err)
		return
	}

	entities := result.([]*entity.Entity)
	for i, ent := range entities {
		fmt.Printf("  %d. %s (%s) - Level %d\n", 
			i+1, ent.Name, ent.Type, ent.Stats.Level)
		fmt.Printf("     HP: %d | Damage: %d | Defense: %d\n",
			ent.Stats.MaxHealth, ent.Stats.Damage, ent.Stats.Defense)
	}
}

func generateItems(genreID string, count int) {
	gen := item.NewItemGenerator()
	params := procgen.GenerationParams{
		GenreID:    genreID,
		Difficulty: 0.5,
		Depth:      5,
		Custom: map[string]interface{}{
			"count": count,
		},
	}

	result, err := gen.Generate(54321, params)
	if err != nil {
		log.Printf("Item generation failed: %v", err)
		return
	}

	items := result.([]*item.Item)
	for i, itm := range items {
		fmt.Printf("  %d. %s (%s) - %s\n", 
			i+1, itm.Name, itm.Type, itm.Rarity)
		
		// Show relevant stats
		if itm.Type == item.TypeWeapon {
			fmt.Printf("     Damage: %d", itm.Stats.Damage)
			if itm.Stats.AttackSpeed > 0 {
				fmt.Printf(" | Speed: %.1f", itm.Stats.AttackSpeed)
			}
			fmt.Println()
		} else if itm.Type == item.TypeArmor {
			fmt.Printf("     Defense: %d", itm.Stats.Defense)
			if itm.Stats.Weight > 0 {
				fmt.Printf(" | Weight: %.1f", itm.Stats.Weight)
			}
			fmt.Println()
		}
	}
}

func generateSpells(genreID string, count int) {
	gen := magic.NewSpellGenerator()
	params := procgen.GenerationParams{
		GenreID:    genreID,
		Difficulty: 0.5,
		Depth:      5,
		Custom: map[string]interface{}{
			"count": count,
		},
	}

	result, err := gen.Generate(99999, params)
	if err != nil {
		log.Printf("Spell generation failed: %v", err)
		return
	}

	spells := result.([]*magic.Spell)
	for i, spell := range spells {
		fmt.Printf("  %d. %s (%s/%s)\n", 
			i+1, spell.Name, spell.Element, spell.Type)
		
		// Show damage or healing based on type
		if spell.Stats.Damage > 0 {
			fmt.Printf("     Damage: %d", spell.Stats.Damage)
		} else if spell.Stats.Healing > 0 {
			fmt.Printf("     Healing: %d", spell.Stats.Healing)
		}
		
		fmt.Printf(" | Range: %.1f | Cost: %d\n",
			spell.Stats.Range, spell.Stats.ManaCost)
		fmt.Printf("     Target: %s", spell.Target)
		if spell.Stats.Cooldown > 0 {
			fmt.Printf(" | Cooldown: %.1fs", spell.Stats.Cooldown)
		}
		fmt.Println()
	}
}

func separator(width int) string {
	result := ""
	for i := 0; i < width; i++ {
		result += "="
	}
	return result
}
