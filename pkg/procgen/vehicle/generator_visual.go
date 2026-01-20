// Package vehicle provides visual variation generation functions.
// This file contains color generation, decoration selection, damage state assignment,
// and decal pattern generation - all contributing to unique visual identity for each vehicle.
//
// Visual variation system (Phase 21.3):
// - Primary/Secondary Colors: Rarity-based color variation with complementary/analogous schemes
// - Decorations: 1-5 genre-specific ornaments scaled by rarity (e.g., "Golden Trim", "Neon Strips")
// - Damage State: Visual wear level (0.0-1.0) independent of actual durability stats
// - Decal Patterns: Genre-specific paint schemes ("Heraldic Stripes", "Racing Stripes", etc.)
//
// All functions moved from generator.go during Phase 3 reorganization to group
// visual/aesthetic functionality separately from core game mechanics.
//
// Code relocated from: generator.go
// Phase 21.3: Visual Variation
package vehicle

import (
	"math/rand"
)

// generateColor varies the base color based on rarity.
// Originally from: generator.go
func (g *VehicleGenerator) generateColor(baseColor uint32, rarity Rarity, rng *rand.Rand) uint32 {
	// Extract RGB components
	r := float64((baseColor >> 16) & 0xFF)
	green := float64((baseColor >> 8) & 0xFF)
	b := float64(baseColor & 0xFF)

	// Apply rarity-based variation
	variation := 30.0
	if rarity >= RarityRare {
		variation = 50.0 // More variation for rare+
	}

	r += (rng.Float64()*2.0 - 1.0) * variation
	green += (rng.Float64()*2.0 - 1.0) * variation
	b += (rng.Float64()*2.0 - 1.0) * variation

	// Clamp to valid range
	r = clamp(r, 0, 255)
	green = clamp(green, 0, 255)
	b = clamp(b, 0, 255)

	return (uint32(r) << 16) | (uint32(green) << 8) | uint32(b)
}

// generateDecorations creates genre-specific visual decorations.
// Phase 21.3: Visual Variation
// Originally from: generator.go
func (g *VehicleGenerator) generateDecorations(genreID string, vehicleType VehicleType, rarity Rarity, rng *rand.Rand) []string {
	// Higher rarity = more decorations
	maxDecorations := int(rarity) + 1 // Common: 1, Uncommon: 2, ..., Legendary: 5
	if maxDecorations > 5 {
		maxDecorations = 5
	}

	genreDecorations := map[string][]string{
		"fantasy": {
			"Golden Trim", "Heraldic Shield", "Feather Plume", "Gemstone Inlay",
			"Engraved Runes", "Banner Flag", "Silver Buckles", "Leather Tassels",
			"Royal Crest", "Mystic Sigil", "Dragon Scales", "Elven Filigree",
		},
		"scifi": {
			"Neon Strips", "Holographic Decal", "Chrome Plating", "LED Arrays",
			"Antenna Array", "Energy Coils", "Carbon Fiber Panel", "Heat Sinks",
			"Sensor Array", "Plasma Vents", "Quantum Stabilizers", "Shield Emitters",
		},
		"horror": {
			"Bone Spikes", "Blood Stains", "Flesh Grafts", "Rusted Chains",
			"Skulls", "Grotesque Faces", "Torn Flesh", "Dark Runes",
			"Cursed Symbols", "Spider Webs", "Dripping Slime", "Jagged Teeth",
		},
		"cyberpunk": {
			"Neon Underglow", "Sponsor Decals", "Chrome Exhaust", "LED Trim",
			"Street Tags", "Corporate Logo", "Holographic Paint", "Spike Wheels",
			"Carbon Spoiler", "Mirror Finish", "Laser Etching", "Tech Panels",
		},
		"postapoc": {
			"Barbed Wire", "Scrap Armor", "Skull Trophy", "Rust Patina",
			"Spikes", "Makeshift Plate", "Bullet Holes", "War Paint",
			"Salvaged Parts", "Chain Mail", "Welded Scrap", "Tire Treads",
		},
	}

	decorations := genreDecorations[genreID]
	if decorations == nil {
		decorations = genreDecorations["fantasy"] // Default
	}

	// Randomly select decorations without repeats
	numDecorations := rng.Intn(maxDecorations) + 1 // At least 1
	if numDecorations > len(decorations) {
		numDecorations = len(decorations)
	}

	// Create a copy and shuffle to get random selection
	selected := make([]string, len(decorations))
	copy(selected, decorations)
	rng.Shuffle(len(selected), func(i, j int) {
		selected[i], selected[j] = selected[j], selected[i]
	})

	return selected[:numDecorations]
}

// generateDamageState creates a visual wear level.
// Phase 21.3: Visual Variation
// Originally from: generator.go
func (g *VehicleGenerator) generateDamageState(rng *rand.Rand) float64 {
	// Most vehicles start in good condition
	// 60% pristine (0.0-0.1), 30% worn (0.1-0.3), 10% damaged (0.3-0.7)
	roll := rng.Float64()
	if roll < 0.6 {
		return rng.Float64() * 0.1 // 0.0-0.1 (pristine)
	} else if roll < 0.9 {
		return 0.1 + rng.Float64()*0.2 // 0.1-0.3 (worn)
	}
	return 0.3 + rng.Float64()*0.4 // 0.3-0.7 (damaged)
}

// generateSecondaryColor creates a complementary or contrasting color.
// Phase 21.3: Visual Variation
// Originally from: generator.go
func (g *VehicleGenerator) generateSecondaryColor(baseColor uint32, rarity Rarity, rng *rand.Rand) uint32 {
	// Extract RGB components from base color
	r := float64((baseColor >> 16) & 0xFF)
	green := float64((baseColor >> 8) & 0xFF)
	b := float64(baseColor & 0xFF)

	// Choose color scheme based on random roll
	roll := rng.Float64()

	var r2, g2, b2 float64
	if roll < 0.33 {
		// Complementary color (invert hue)
		r2 = 255 - r
		g2 = 255 - green
		b2 = 255 - b
	} else if roll < 0.66 {
		// Analogous color (shift hue slightly)
		shift := 30.0 + rng.Float64()*30.0 // Shift by 30-60
		r2 = r + shift
		g2 = green + shift
		b2 = b - shift
	} else {
		// Monochromatic (lighter or darker version)
		factor := 0.5 + rng.Float64()*0.5 // 0.5x to 1.0x
		if rng.Float64() < 0.5 {
			factor = 1.0 + rng.Float64()*0.5 // Or 1.0x to 1.5x
		}
		r2 = r * factor
		g2 = green * factor
		b2 = b * factor
	}

	// Clamp to valid range
	r2 = clamp(r2, 0, 255)
	g2 = clamp(g2, 0, 255)
	b2 = clamp(b2, 0, 255)

	return (uint32(r2) << 16) | (uint32(g2) << 8) | uint32(b2)
}

// generateDecalPattern selects a paint/decal pattern.
// Phase 21.3: Visual Variation
// Originally from: generator.go
func (g *VehicleGenerator) generateDecalPattern(genreID string, rarity Rarity, rng *rand.Rand) string {
	// Common vehicles less likely to have patterns
	if rarity == RarityCommon && rng.Float64() < 0.7 {
		return "Solid" // No pattern, 70% chance for common
	}

	genrePatterns := map[string][]string{
		"fantasy": {
			"Solid", "Heraldic Stripes", "Diamond Pattern", "Checkered",
			"Flames", "Royal Banners", "Engraved", "Gemstone Mosaic",
		},
		"scifi": {
			"Solid", "Racing Stripes", "Hexagonal Camo", "Digital Camo",
			"Flames", "Geometric Panels", "Holographic", "Circuit Pattern",
		},
		"horror": {
			"Solid", "Blood Splatter", "Cracked", "Decaying",
			"Veins", "Bone Pattern", "Shadow Wisps", "Cursed Symbols",
		},
		"cyberpunk": {
			"Solid", "Neon Stripes", "Urban Camo", "Grid Pattern",
			"Flames", "Graffiti", "Corporate Livery", "Holo-Flake",
		},
		"postapoc": {
			"Solid", "Desert Camo", "Rust Stripes", "Scrap Patchwork",
			"War Paint", "Tribal Marks", "Bullet Impacts", "Welded Panels",
		},
	}

	patterns := genrePatterns[genreID]
	if patterns == nil {
		patterns = genrePatterns["fantasy"] // Default
	}

	return patterns[rng.Intn(len(patterns))]
}
