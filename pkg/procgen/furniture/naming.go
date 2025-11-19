package furniture

import (
	"fmt"
	"image/color"
	"math/rand"
)

// getMaterialColor returns primary color based on material and genre
func (gen *Generator) getMaterialColor(rng *rand.Rand, material MaterialType, genreID string) color.RGBA {
	switch material {
	case MaterialWood:
		// Browns and tans
		r := uint8(100 + rng.Intn(56)) // 100-155
		g := uint8(60 + rng.Intn(41))  // 60-100
		b := uint8(20 + rng.Intn(31))  // 20-50
		return color.RGBA{R: r, G: g, B: b, A: 255}

	case MaterialMetal:
		// Grays, silvers, with genre tint
		base := uint8(140 + rng.Intn(76)) // 140-215
		var r, g, b uint8
		switch genreID {
		case "scifi", "cyberpunk":
			// Blue-tinted metal
			r = base - 20
			g = base - 10
			b = base
		case "fantasy":
			// Silver/gold tint
			if rng.Float64() < 0.3 {
				// Gold
				r = base + 20
				g = base + 10
				b = base - 30
			} else {
				// Silver
				r = base
				g = base
				b = base
			}
		default:
			r = base
			g = base
			b = base
		}
		return color.RGBA{R: r, G: g, B: b, A: 255}

	case MaterialStone:
		// Grays and browns
		base := uint8(80 + rng.Intn(96))      // 80-175
		variation := uint8(rng.Intn(21) - 10) // -10 to +10
		r := base
		g := base + variation
		b := base + variation/2
		return color.RGBA{R: r, G: g, B: b, A: 255}

	case MaterialCrystal:
		// Bright, saturated colors influenced by genre
		switch genreID {
		case "fantasy":
			// Blue, purple, green crystals
			crystalType := rng.Intn(3)
			switch crystalType {
			case 0: // Blue
				return color.RGBA{R: 100, G: 150, B: 255, A: 255}
			case 1: // Purple
				return color.RGBA{R: 200, G: 100, B: 255, A: 255}
			default: // Green
				return color.RGBA{R: 100, G: 255, B: 150, A: 255}
			}
		case "scifi", "cyberpunk":
			// Neon colors
			return color.RGBA{R: 0, G: 255, B: 255, A: 255} // Cyan
		case "horror":
			// Dark red/purple
			return color.RGBA{R: 150, G: 50, B: 100, A: 255}
		default:
			// White/clear
			return color.RGBA{R: 230, G: 240, B: 255, A: 255}
		}

	case MaterialFabric:
		// Wide variety of colors, genre-influenced
		switch genreID {
		case "fantasy":
			// Rich colors
			colors := []color.RGBA{
				{R: 180, G: 20, B: 20, A: 255},  // Red
				{R: 20, G: 100, B: 180, A: 255}, // Blue
				{R: 100, G: 150, B: 50, A: 255}, // Green
				{R: 120, G: 80, B: 150, A: 255}, // Purple
			}
			return colors[rng.Intn(len(colors))]
		case "horror":
			// Dark, muted colors
			r := uint8(40 + rng.Intn(41))
			g := uint8(30 + rng.Intn(31))
			b := uint8(30 + rng.Intn(31))
			return color.RGBA{R: r, G: g, B: b, A: 255}
		default:
			// Random colors
			r := uint8(100 + rng.Intn(156))
			g := uint8(100 + rng.Intn(156))
			b := uint8(100 + rng.Intn(156))
			return color.RGBA{R: r, G: g, B: b, A: 255}
		}
	}

	// Fallback
	return color.RGBA{R: 128, G: 128, B: 128, A: 255}
}

// getSecondaryColor generates a complementary or accent color
func (gen *Generator) getSecondaryColor(rng *rand.Rand, primary color.RGBA) color.RGBA {
	// Create slight variation of primary color for most materials
	variation := 30

	r := int(primary.R) + rng.Intn(variation*2) - variation
	g := int(primary.G) + rng.Intn(variation*2) - variation
	b := int(primary.B) + rng.Intn(variation*2) - variation

	// Clamp to valid range
	if r < 0 {
		r = 0
	}
	if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	}
	if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	}
	if b > 255 {
		b = 255
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

// generateName creates a name with material and rarity prefixes
func (gen *Generator) generateName(rng *rand.Rand, material MaterialType, rarity RarityTier, tmpl *Template, genreID string) string {
	name := ""

	// Add rarity prefix for uncommon and above
	if rarity >= RarityUncommon {
		name += gen.getRarityPrefix(rng, rarity, genreID) + " "
	}

	// Add material adjective for rare and above
	if rarity >= RarityRare {
		name += material.String() + " "
	}

	// Base name
	name += tmpl.BaseName

	// Add quality suffix for epic and above
	if rarity >= RarityEpic {
		name += gen.getQualitySuffix(rng, rarity, genreID)
	}

	return name
}

// getRarityPrefix returns a genre-appropriate prefix for the rarity
func (gen *Generator) getRarityPrefix(rng *rand.Rand, rarity RarityTier, genreID string) string {
	switch rarity {
	case RarityUncommon:
		prefixes := map[string][]string{
			"fantasy":   {"Fine", "Quality", "Sturdy", "Well-Made"},
			"scifi":     {"Standard", "Certified", "Regulation"},
			"horror":    {"Worn", "Aged", "Weathered"},
			"cyberpunk": {"Branded", "Corporate", "Upscale"},
			"postapoc":  {"Salvaged", "Repaired", "Functional"},
		}
		list := prefixes[genreID]
		if len(list) == 0 {
			list = prefixes["fantasy"]
		}
		return list[rng.Intn(len(list))]

	case RarityRare:
		prefixes := map[string][]string{
			"fantasy":   {"Exquisite", "Masterwork", "Superior", "Elegant"},
			"scifi":     {"Advanced", "Prototype", "Enhanced"},
			"horror":    {"Ancient", "Cursed", "Haunted"},
			"cyberpunk": {"Premium", "Luxury", "Elite"},
			"postapoc":  {"Pristine", "Military-Grade", "Pre-War"},
		}
		list := prefixes[genreID]
		if len(list) == 0 {
			list = prefixes["fantasy"]
		}
		return list[rng.Intn(len(list))]

	case RarityEpic:
		prefixes := map[string][]string{
			"fantasy":   {"Legendary", "Royal", "Grand", "Majestic"},
			"scifi":     {"Quantum", "Experimental", "Military"},
			"horror":    {"Eldritch", "Corrupted", "Profane"},
			"cyberpunk": {"Executive", "Black Market", "Custom"},
			"postapoc":  {"Legendary", "Vault", "Preserved"},
		}
		list := prefixes[genreID]
		if len(list) == 0 {
			list = prefixes["fantasy"]
		}
		return list[rng.Intn(len(list))]

	case RarityLegendary:
		prefixes := map[string][]string{
			"fantasy":   {"Mythical", "Divine", "Transcendent"},
			"scifi":     {"Alien", "Singularity", "Godtech"},
			"horror":    {"Nightmare", "Abyssal", "Forsaken"},
			"cyberpunk": {"Prototype", "Nexus", "Ultimate"},
			"postapoc":  {"Artifact", "Relic", "Ancient"},
		}
		list := prefixes[genreID]
		if len(list) == 0 {
			list = prefixes["fantasy"]
		}
		return list[rng.Intn(len(list))]
	}

	return ""
}

// getQualitySuffix returns a quality descriptor suffix
func (gen *Generator) getQualitySuffix(rng *rand.Rand, rarity RarityTier, genreID string) string {
	switch rarity {
	case RarityEpic:
		suffixes := map[string][]string{
			"fantasy":   {" of Power", " of the Realm", " of Excellence"},
			"scifi":     {" Mk.VII", " Series X", " Elite Edition"},
			"horror":    {" of Dread", " of Torment", " of Shadows"},
			"cyberpunk": {" Prime", " Ultra", " Black Edition"},
			"postapoc":  {" Supreme", " Ultimate", " Prime"},
		}
		list := suffixes[genreID]
		if len(list) == 0 {
			list = suffixes["fantasy"]
		}
		return list[rng.Intn(len(list))]

	case RarityLegendary:
		suffixes := map[string][]string{
			"fantasy":   {" of Legend", " of the Gods", " of Eternity"},
			"scifi":     {" Omega", " Genesis", " Infinity"},
			"horror":    {" of the Void", " of Madness", " Beyond"},
			"cyberpunk": {" Zero", " Nexus", " Prime"},
			"postapoc":  {" Eternal", " Everlasting", " Immortal"},
		}
		list := suffixes[genreID]
		if len(list) == 0 {
			list = suffixes["fantasy"]
		}
		return list[rng.Intn(len(list))]
	}

	return ""
}

// generateDescription creates a descriptive text for the furniture
func (gen *Generator) generateDescription(tmpl *Template, material MaterialType, rarity RarityTier, genreID string) string {
	desc := fmt.Sprintf("A %s %s made from %s. ", rarity.String(), tmpl.BaseName, material.String())

	// Add functional description
	if tmpl.Functional {
		switch tmpl.Type {
		case TypeStorage:
			desc += "Can be used to store items. "
		case TypeCrafting:
			desc += fmt.Sprintf("Used for %s. ", tmpl.SubType)
		case TypeLighting:
			desc += "Provides illumination. "
		case TypeBedding:
			desc += "Can be used for rest. "
		case TypeSeating:
			desc += "Provides comfortable seating. "
		}
	}

	// Add genre-specific flavor
	switch genreID {
	case "fantasy":
		if rarity >= RarityRare {
			desc += "Crafted by master artisans. "
		}
	case "scifi":
		if rarity >= RarityRare {
			desc += "Features advanced technology. "
		}
	case "horror":
		desc += "Has an unsettling presence. "
	case "cyberpunk":
		if rarity >= RarityRare {
			desc += "Shows corporate quality. "
		}
	case "postapoc":
		desc += "Remarkably well-preserved. "
	}

	return desc
}
