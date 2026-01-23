package furniture

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// This file contains the furniture Generator and all generation logic including:
// - Core generation algorithm (Generate, Validate)
// - Rarity and material selection
// - Name and color generation (consolidated from naming.go)
// - Helper functions for genre-specific material and naming

// Generator creates procedural furniture items
type Generator struct{}

// NewGenerator creates a new furniture generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a furniture item with deterministic seed-based generation
// resolveSubType determines the furniture subtype from params or random selection
func (gen *Generator) resolveSubType(rng *rand.Rand, params procgen.GenerationParams) string {
	if st, ok := params.Custom["SubType"].(string); ok && st != "" {
		return st
	}
	return gen.chooseRandomSubType(rng, params)
}

// generateDimensions creates width, height, depth dimensions with rarity scaling
func generateDimensions(rng *rand.Rand, tmpl *Template, rarity RarityTier) (width, height, depth float64) {
	width = tmpl.MinWidth + rng.Float64()*(tmpl.MaxWidth-tmpl.MinWidth)
	height = tmpl.MinHeight + rng.Float64()*(tmpl.MaxHeight-tmpl.MinHeight)
	depth = tmpl.MinDepth + rng.Float64()*(tmpl.MaxDepth-tmpl.MinDepth)

	rarityScale := 1.0 + (float64(rarity) * 0.1)
	width *= rarityScale
	height *= rarityScale
	depth *= rarityScale

	return width, height, depth
}

// calculateCollisionBox determines collision dimensions based on walkability
func calculateCollisionBox(width, depth float64, walkable bool) (collisionWidth, collisionDepth float64) {
	collisionWidth = width
	collisionDepth = depth
	if walkable {
		collisionWidth *= 0.5
		collisionDepth *= 0.5
	}
	return collisionWidth, collisionDepth
}

// calculateCapacity determines storage capacity with rarity scaling
func calculateCapacity(baseCapacity int, rarity RarityTier) int {
	if baseCapacity <= 0 {
		return 0
	}
	return int(float64(baseCapacity) * rarity.DetailMultiplier())
}

// calculateLightIntensity determines light intensity with random variation
func calculateLightIntensity(rng *rand.Rand, baseLightLevel float64) float64 {
	if baseLightLevel <= 0 {
		return 0
	}
	intensity := baseLightLevel + (rng.Float64()*0.2 - 0.1)
	if intensity < 0 {
		return 0
	}
	if intensity > 1.0 {
		return 1.0
	}
	return intensity
}

// buildFurniture constructs a Furniture object from all generated components
func buildFurniture(seed int64, tmpl *Template, material MaterialType, rarity RarityTier, genreID, name, description string, width, height, depth float64, primaryColor, secondaryColor color.RGBA, detailLevel, collisionWidth, collisionDepth float64, capacity int, lightIntensity float64) *Furniture {
	furniture := &Furniture{
		ID:          fmt.Sprintf("furniture_%d", seed),
		Type:        tmpl.Type,
		SubType:     tmpl.SubType,
		Material:    material,
		Rarity:      rarity,
		GenreID:     genreID,
		Name:        name,
		Description: description,

		Width:  width,
		Height: height,
		Depth:  depth,

		PrimaryColor:   primaryColor,
		SecondaryColor: secondaryColor,
		DetailLevel:    detailLevel,

		Direction:      DirNorth,
		Walkable:       tmpl.Walkable,
		CollisionWidth: collisionWidth,
		CollisionDepth: collisionDepth,

		Functional:     tmpl.Functional,
		Capacity:       capacity,
		LightIntensity: lightIntensity,
	}

	if tmpl.Type == TypeCrafting {
		furniture.CraftingType = tmpl.SubType
	}

	return furniture
}

func (gen *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	rng := rand.New(rand.NewSource(seed))

	subType := gen.resolveSubType(rng, params)

	tmpl := GetTemplate(subType)
	if tmpl == nil {
		return nil, fmt.Errorf("unknown furniture subtype: %s", subType)
	}

	rarity := gen.determineRarity(rng, params)
	material := gen.selectMaterial(rng, tmpl, params.GenreID, rarity)

	width, height, depth := generateDimensions(rng, tmpl, rarity)

	name := gen.generateName(rng, material, rarity, tmpl, params.GenreID)
	primaryColor := gen.getMaterialColor(rng, material, params.GenreID)
	secondaryColor := gen.getSecondaryColor(rng, primaryColor)

	detailLevel := tmpl.DetailComplexity * rarity.DetailMultiplier()
	collisionWidth, collisionDepth := calculateCollisionBox(width, depth, tmpl.Walkable)
	capacity := calculateCapacity(tmpl.BaseCapacity, rarity)
	lightIntensity := calculateLightIntensity(rng, tmpl.BaseLightLevel)

	description := gen.generateDescription(tmpl, material, rarity, params.GenreID)

	furniture := buildFurniture(seed, tmpl, material, rarity, params.GenreID, name, description, width, height, depth, primaryColor, secondaryColor, detailLevel, collisionWidth, collisionDepth, capacity, lightIntensity)

	return furniture, nil
}

// Validate checks if the generated furniture is valid
func (gen *Generator) Validate(result interface{}) error {
	furniture, ok := result.(*Furniture)
	if !ok {
		return fmt.Errorf("result is not a Furniture")
	}

	// Check dimensions are positive
	if furniture.Width <= 0 || furniture.Height <= 0 || furniture.Depth <= 0 {
		return fmt.Errorf("invalid dimensions: %.2f×%.2f×%.2f", furniture.Width, furniture.Height, furniture.Depth)
	}

	// Check collision dimensions are positive
	if furniture.CollisionWidth <= 0 || furniture.CollisionDepth <= 0 {
		return fmt.Errorf("invalid collision dimensions: %.2f×%.2f", furniture.CollisionWidth, furniture.CollisionDepth)
	}

	// Check detail level is reasonable
	if furniture.DetailLevel < 0 || furniture.DetailLevel > 10.0 {
		return fmt.Errorf("invalid detail level: %.2f", furniture.DetailLevel)
	}

	// Check light intensity is in valid range
	if furniture.LightIntensity < 0 || furniture.LightIntensity > 1.0 {
		return fmt.Errorf("invalid light intensity: %.2f", furniture.LightIntensity)
	}

	// Check capacity is non-negative
	if furniture.Capacity < 0 {
		return fmt.Errorf("invalid capacity: %d", furniture.Capacity)
	}

	return nil
}

// chooseRandomSubType selects a random furniture subtype weighted by depth/difficulty
func (gen *Generator) chooseRandomSubType(rng *rand.Rand, params procgen.GenerationParams) string {
	// At low depth, prefer common furniture types
	// At high depth, prefer rare/decorative types

	allSubTypes := GetAllSubTypes()
	if len(allSubTypes) == 0 {
		return "Chair" // Fallback
	}

	// Simple random selection for now
	// Could be weighted by depth/difficulty in future
	return allSubTypes[rng.Intn(len(allSubTypes))]
}

// determineRarity calculates rarity tier based on difficulty and depth
func (gen *Generator) determineRarity(rng *rand.Rand, params procgen.GenerationParams) RarityTier {
	// Base probability affected by difficulty and depth
	roll := rng.Float64()

	// Difficulty and depth increase chance of higher rarity
	modifier := params.Difficulty*0.3 + float64(params.Depth)*0.02
	roll += modifier

	if roll < 0.5 {
		return RarityCommon
	} else if roll < 0.75 {
		return RarityUncommon
	} else if roll < 0.9 {
		return RarityRare
	} else if roll < 0.97 {
		return RarityEpic
	}
	return RarityLegendary
}

// selectMaterial chooses material from allowed materials, influenced by genre and rarity
func (gen *Generator) selectMaterial(rng *rand.Rand, tmpl *Template, genreID string, rarity RarityTier) MaterialType {
	if len(tmpl.AllowedMaterials) == 0 {
		return MaterialWood
	}

	// Pre-generate all RNG values to ensure determinism
	exoticRoll := rng.Float64()
	genreRoll := rng.Float64()
	finalRoll := rng.Intn(len(tmpl.AllowedMaterials))

	// Try exotic material selection for high rarity
	if mat, ok := gen.trySelectExoticMaterialDeterministic(exoticRoll, tmpl, rarity); ok {
		return mat
	}

	// Try genre-specific material selection
	if mat, ok := gen.trySelectGenreMaterialDeterministic(genreRoll, tmpl, genreID); ok {
		return mat
	}

	// Fallback to random selection
	return tmpl.AllowedMaterials[finalRoll]
}

// trySelectExoticMaterialDeterministic attempts to select exotic materials for high rarity items
// Uses pre-rolled random value to ensure determinism
func (gen *Generator) trySelectExoticMaterialDeterministic(roll float64, tmpl *Template, rarity RarityTier) (MaterialType, bool) {
	if rarity < RarityEpic || len(tmpl.AllowedMaterials) <= 1 {
		return 0, false
	}

	// Check for crystal first (60% chance if available)
	for _, mat := range tmpl.AllowedMaterials {
		if mat == MaterialCrystal && roll < 0.6 {
			return MaterialCrystal, true
		}
	}

	// Check for metal (50% chance if available and crystal didn't match)
	for _, mat := range tmpl.AllowedMaterials {
		if mat == MaterialMetal && roll >= 0.6 && roll < 0.8 {
			return MaterialMetal, true
		}
	}

	return 0, false
}

// trySelectGenreMaterialDeterministic attempts to select genre-appropriate materials
// Uses pre-rolled random value to ensure determinism
func (gen *Generator) trySelectGenreMaterialDeterministic(roll float64, tmpl *Template, genreID string) (MaterialType, bool) {
	switch genreID {
	case "fantasy":
		return gen.trySelectFantasyMaterialDeterministic(roll, tmpl)
	case "scifi", "cyberpunk":
		return gen.trySelectScifiMaterialDeterministic(roll, tmpl)
	case "horror":
		return gen.trySelectHorrorMaterialDeterministic(roll, tmpl)
	case "postapoc":
		return gen.trySelectPostapocMaterialDeterministic(roll, tmpl)
	}
	return 0, false
}

// trySelectFantasyMaterialDeterministic attempts to select fantasy-themed materials
// Uses pre-rolled random value to ensure determinism
func (gen *Generator) trySelectFantasyMaterialDeterministic(roll float64, tmpl *Template) (MaterialType, bool) {
	for _, mat := range tmpl.AllowedMaterials {
		if mat == MaterialWood && roll < 0.4 {
			return MaterialWood, true
		}
		if mat == MaterialStone && roll >= 0.4 && roll < 0.7 {
			return MaterialStone, true
		}
	}
	return 0, false
}

// trySelectScifiMaterialDeterministic attempts to select sci-fi themed materials
// Uses pre-rolled random value to ensure determinism
func (gen *Generator) trySelectScifiMaterialDeterministic(roll float64, tmpl *Template) (MaterialType, bool) {
	for _, mat := range tmpl.AllowedMaterials {
		if mat == MaterialMetal && roll < 0.5 {
			return MaterialMetal, true
		}
		if mat == MaterialCrystal && roll >= 0.5 && roll < 0.8 {
			return MaterialCrystal, true
		}
	}
	return 0, false
}

// trySelectHorrorMaterialDeterministic attempts to select horror-themed materials
// Uses pre-rolled random value to ensure determinism
func (gen *Generator) trySelectHorrorMaterialDeterministic(roll float64, tmpl *Template) (MaterialType, bool) {
	for _, mat := range tmpl.AllowedMaterials {
		if mat == MaterialStone && roll < 0.4 {
			return MaterialStone, true
		}
		if mat == MaterialWood && roll >= 0.4 && roll < 0.7 {
			return MaterialWood, true
		}
	}
	return 0, false
}

// trySelectPostapocMaterialDeterministic attempts to select post-apocalyptic themed materials
// Uses pre-rolled random value to ensure determinism
func (gen *Generator) trySelectPostapocMaterialDeterministic(roll float64, tmpl *Template) (MaterialType, bool) {
	for _, mat := range tmpl.AllowedMaterials {
		if mat == MaterialMetal && roll < 0.4 {
			return MaterialMetal, true
		}
		if mat == MaterialWood && roll >= 0.4 && roll < 0.7 {
			return MaterialWood, true
		}
	}
	return 0, false
}

// Code relocated from: naming.go

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
		// Pre-roll for fantasy gold/silver choice to ensure determinism
		goldRoll := rng.Float64()
		var r, g, b uint8
		switch genreID {
		case "scifi", "cyberpunk":
			// Blue-tinted metal
			r = base - 20
			g = base - 10
			b = base
		case "fantasy":
			// Silver/gold tint
			if goldRoll < 0.3 {
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
