package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/magic"
)

var (
	genre      = flag.String("genre", "fantasy", "Genre: fantasy or scifi")
	count      = flag.Int("count", 20, "Number of spells to generate")
	depth      = flag.Int("depth", 5, "Depth level (affects power and rarity)")
	difficulty = flag.Float64("difficulty", 0.5, "Difficulty multiplier (0.0-1.0)")
	seed       = flag.Int64("seed", 12345, "Generation seed")
	output     = flag.String("output", "", "Output file (leave empty for console)")
	verbose    = flag.Bool("verbose", false, "Show detailed spell information")
	spellType  = flag.String("type", "", "Filter by type: offensive, defensive, healing, buff, debuff, utility, summon")
)

func main() {
	flag.Parse()

	log.Printf("Generating spells for %s genre", *genre)
	log.Printf("Count: %d, Depth: %d, Difficulty: %.1f, Seed: %d", *count, *depth, *difficulty, *seed)

	// Create generator
	gen := magic.NewSpellGenerator()

	// Set up generation parameters
	params := procgen.GenerationParams{
		Difficulty: *difficulty,
		Depth:      *depth,
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"count": *count,
		},
	}

	// Generate spells
	result, err := gen.Generate(*seed, params)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	spells, ok := result.([]*magic.Spell)
	if !ok {
		log.Fatal("Result is not []*Spell")
	}

	// Validate
	if err := gen.Validate(spells); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// Filter by type if specified
	if *spellType != "" {
		spells = filterByType(spells, *spellType)
		log.Printf("Filtered to %d spells of type %s", len(spells), *spellType)
	}

	log.Printf("Generated %d spells", len(spells))

	// Render to string
	rendered := renderSpells(spells, *verbose)

	// Output to file or console
	if *output != "" {
		if err := os.WriteFile(*output, []byte(rendered), 0o644); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		log.Printf("Spells saved to %s", *output)
	} else {
		fmt.Println(rendered)
	}
}

// filterByType filters spells by type name
func filterByType(spells []*magic.Spell, typeName string) []*magic.Spell {
	var filtered []*magic.Spell
	typeName = strings.ToLower(typeName)

	for _, spell := range spells {
		if strings.ToLower(spell.Type.String()) == typeName {
			filtered = append(filtered, spell)
		}
	}

	return filtered
}

// renderSpells converts spells to readable text
func renderSpells(spells []*magic.Spell, verbose bool) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Generated %d Spells\n", len(spells)))
	result.WriteString(strings.Repeat("=", 80) + "\n\n")

	// Count by type and element
	typeCount := make(map[magic.SpellType]int)
	elementCount := make(map[magic.ElementType]int)
	rarityCount := make(map[magic.Rarity]int)

	for _, s := range spells {
		typeCount[s.Type]++
		elementCount[s.Element]++
		rarityCount[s.Rarity]++
	}

	result.WriteString("Summary:\n")
	result.WriteString("  By Type:\n")
	for t := magic.TypeOffensive; t <= magic.TypeSummon; t++ {
		if typeCount[t] > 0 {
			result.WriteString(fmt.Sprintf("    %s: %d\n", strings.Title(t.String()), typeCount[t]))
		}
	}

	result.WriteString("  By Element:\n")
	for e := magic.ElementNone; e <= magic.ElementArcane; e++ {
		if elementCount[e] > 0 {
			result.WriteString(fmt.Sprintf("    %s: %d\n", strings.Title(e.String()), elementCount[e]))
		}
	}

	result.WriteString("  By Rarity:\n")
	result.WriteString(fmt.Sprintf("    Common: %d, Uncommon: %d, Rare: %d, Epic: %d, Legendary: %d\n\n",
		rarityCount[magic.RarityCommon], rarityCount[magic.RarityUncommon],
		rarityCount[magic.RarityRare], rarityCount[magic.RarityEpic],
		rarityCount[magic.RarityLegendary]))

	result.WriteString(strings.Repeat("-", 80) + "\n\n")

	// List spells
	for i, s := range spells {
		if verbose {
			result.WriteString(renderSpellVerbose(i+1, s))
		} else {
			result.WriteString(renderSpellCompact(i+1, s))
		}
	}

	return result.String()
}

// renderSpellCompact renders spell in compact format
func renderSpellCompact(num int, s *magic.Spell) string {
	// Build stat summary based on spell type
	statSummary := ""
	if s.Stats.Damage > 0 {
		statSummary += fmt.Sprintf("DMG:%d ", s.Stats.Damage)
	}
	if s.Stats.Healing > 0 {
		statSummary += fmt.Sprintf("HEAL:%d ", s.Stats.Healing)
	}
	statSummary += fmt.Sprintf("MP:%d CD:%.1fs", s.Stats.ManaCost, s.Stats.Cooldown)

	return fmt.Sprintf("%2d. %-30s [%s] Lv.%-2d | %s | %s %s\n",
		num, s.Name, getRaritySymbol(s.Rarity), s.Stats.RequiredLevel,
		statSummary, s.Element.String(), s.Target.String())
}

// renderSpellVerbose renders spell with full details
func renderSpellVerbose(num int, s *magic.Spell) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Spell #%d: %s\n", num, s.Name))
	result.WriteString(fmt.Sprintf("  Type:        %s\n", s.Type.String()))
	result.WriteString(fmt.Sprintf("  Element:     %s\n", s.Element.String()))
	result.WriteString(fmt.Sprintf("  Rarity:      %s %s\n", getRaritySymbol(s.Rarity), s.Rarity.String()))
	result.WriteString(fmt.Sprintf("  Target:      %s\n", s.Target.String()))
	result.WriteString(fmt.Sprintf("  Level:       %d\n", s.Stats.RequiredLevel))
	result.WriteString(fmt.Sprintf("  Power:       %d/100\n", s.GetPowerLevel()))
	result.WriteString(fmt.Sprintf("  Stats:\n"))

	if s.Stats.Damage > 0 {
		result.WriteString(fmt.Sprintf("    Damage:    %d\n", s.Stats.Damage))
	}
	if s.Stats.Healing > 0 {
		result.WriteString(fmt.Sprintf("    Healing:   %d\n", s.Stats.Healing))
	}

	result.WriteString(fmt.Sprintf("    Mana Cost: %d\n", s.Stats.ManaCost))
	result.WriteString(fmt.Sprintf("    Cooldown:  %.2fs\n", s.Stats.Cooldown))
	result.WriteString(fmt.Sprintf("    Cast Time: %.2fs\n", s.Stats.CastTime))

	if s.Stats.Range > 0 {
		result.WriteString(fmt.Sprintf("    Range:     %.1f\n", s.Stats.Range))
	}
	if s.Stats.AreaSize > 0 {
		result.WriteString(fmt.Sprintf("    Area Size: %.1f\n", s.Stats.AreaSize))
	}
	if s.Stats.Duration > 0 {
		result.WriteString(fmt.Sprintf("    Duration:  %.1fs\n", s.Stats.Duration))
	}

	if len(s.Tags) > 0 {
		result.WriteString(fmt.Sprintf("  Tags:        %s\n", strings.Join(s.Tags, ", ")))
	}

	result.WriteString(fmt.Sprintf("  Description: %s\n", s.Description))
	result.WriteString("\n")

	return result.String()
}

// getRaritySymbol returns a symbol representing the rarity
func getRaritySymbol(r magic.Rarity) string {
	switch r {
	case magic.RarityCommon:
		return "●"
	case magic.RarityUncommon:
		return "◆"
	case magic.RarityRare:
		return "★"
	case magic.RarityEpic:
		return "◈"
	case magic.RarityLegendary:
		return "♛"
	default:
		return "?"
	}
}
