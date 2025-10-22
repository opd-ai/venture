package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
)

var (
	genre      = flag.String("genre", "fantasy", "Genre: fantasy or scifi")
	count      = flag.Int("count", 20, "Number of entities to generate")
	depth      = flag.Int("depth", 5, "Depth level (affects difficulty)")
	difficulty = flag.Float64("difficulty", 0.5, "Difficulty multiplier (0.0-1.0)")
	seed       = flag.Int64("seed", 12345, "Generation seed")
	output     = flag.String("output", "", "Output file (leave empty for console)")
	verbose    = flag.Bool("verbose", false, "Show detailed entity information")
)

func main() {
	flag.Parse()

	log.Printf("Generating entities for %s genre", *genre)
	log.Printf("Count: %d, Depth: %d, Difficulty: %.1f, Seed: %d", *count, *depth, *difficulty, *seed)

	// Create generator
	gen := entity.NewEntityGenerator()

	// Set up generation parameters
	params := procgen.GenerationParams{
		Difficulty: *difficulty,
		Depth:      *depth,
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"count": *count,
		},
	}

	// Generate entities
	result, err := gen.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	entities, ok := result.([]*entity.Entity)
	if !ok {
		log.Fatal("Result is not []*Entity")
	}

	// Validate
	if err := gen.Validate(entities); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	log.Printf("Generated %d entities", len(entities))

	// Render to string
	rendered := renderEntities(entities, *verbose)

	// Output to file or console
	if *output != "" {
		if err := os.WriteFile(*output, []byte(rendered), 0o644); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		log.Printf("Entities saved to %s", *output)
	} else {
		fmt.Println(rendered)
	}
}

// renderEntities converts entities to readable text
func renderEntities(entities []*entity.Entity, verbose bool) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Generated %d Entities\n", len(entities)))
	result.WriteString(strings.Repeat("=", 80) + "\n\n")

	// Count by type
	typeCount := make(map[entity.EntityType]int)
	rarityCount := make(map[entity.Rarity]int)
	for _, e := range entities {
		typeCount[e.Type]++
		rarityCount[e.Rarity]++
	}

	result.WriteString("Summary:\n")
	result.WriteString(fmt.Sprintf("  Monsters: %d, Bosses: %d, Minions: %d, NPCs: %d\n",
		typeCount[entity.TypeMonster], typeCount[entity.TypeBoss],
		typeCount[entity.TypeMinion], typeCount[entity.TypeNPC]))
	result.WriteString(fmt.Sprintf("  Common: %d, Uncommon: %d, Rare: %d, Epic: %d, Legendary: %d\n\n",
		rarityCount[entity.RarityCommon], rarityCount[entity.RarityUncommon],
		rarityCount[entity.RarityRare], rarityCount[entity.RarityEpic],
		rarityCount[entity.RarityLegendary]))

	result.WriteString(strings.Repeat("-", 80) + "\n\n")

	// List entities
	for i, e := range entities {
		if verbose {
			result.WriteString(renderEntityVerbose(i+1, e))
		} else {
			result.WriteString(renderEntityCompact(i+1, e))
		}
	}

	return result.String()
}

// renderEntityCompact renders entity in compact format
func renderEntityCompact(num int, e *entity.Entity) string {
	hostile := "HOSTILE"
	if !e.IsHostile() {
		hostile = "FRIENDLY"
	}

	return fmt.Sprintf("%2d. %-25s [%s] Lv.%-2d | HP:%-4d DMG:%-3d DEF:%-3d SPD:%.1f | %s %s\n",
		num, e.Name, getRaritySymbol(e.Rarity), e.Stats.Level,
		e.Stats.MaxHealth, e.Stats.Damage, e.Stats.Defense, e.Stats.Speed,
		e.Type.String(), hostile)
}

// renderEntityVerbose renders entity with full details
func renderEntityVerbose(num int, e *entity.Entity) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Entity #%d: %s\n", num, e.Name))
	result.WriteString(fmt.Sprintf("  Type:        %s (%s)\n", e.Type.String(), e.Size.String()))
	result.WriteString(fmt.Sprintf("  Rarity:      %s %s\n", getRaritySymbol(e.Rarity), e.Rarity.String()))
	result.WriteString(fmt.Sprintf("  Level:       %d\n", e.Stats.Level))
	result.WriteString(fmt.Sprintf("  Stats:\n"))
	result.WriteString(fmt.Sprintf("    Health:    %d / %d\n", e.Stats.Health, e.Stats.MaxHealth))
	result.WriteString(fmt.Sprintf("    Damage:    %d\n", e.Stats.Damage))
	result.WriteString(fmt.Sprintf("    Defense:   %d\n", e.Stats.Defense))
	result.WriteString(fmt.Sprintf("    Speed:     %.2f\n", e.Stats.Speed))
	result.WriteString(fmt.Sprintf("  Hostile:     %v\n", e.IsHostile()))
	result.WriteString(fmt.Sprintf("  Threat:      %d/100\n", e.GetThreatLevel()))

	if len(e.Tags) > 0 {
		result.WriteString(fmt.Sprintf("  Tags:        %s\n", strings.Join(e.Tags, ", ")))
	}

	result.WriteString("\n")

	return result.String()
}

// getRaritySymbol returns a symbol representing the rarity
func getRaritySymbol(r entity.Rarity) string {
	switch r {
	case entity.RarityCommon:
		return "●"
	case entity.RarityUncommon:
		return "◆"
	case entity.RarityRare:
		return "★"
	case entity.RarityEpic:
		return "◈"
	case entity.RarityLegendary:
		return "♛"
	default:
		return "?"
	}
}
