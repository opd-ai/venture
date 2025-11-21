package furniture

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
)

// Generator creates procedural furniture items
type Generator struct{}

// NewGenerator creates a new furniture generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a furniture item with deterministic seed-based generation
func (gen *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	rng := rand.New(rand.NewSource(seed))

	// Extract custom parameters
	subType := ""
	if st, ok := params.Custom["SubType"].(string); ok {
		subType = st
	}

	// If no subtype specified, choose random based on depth/difficulty
	if subType == "" {
		subType = gen.chooseRandomSubType(rng, params)
	}

	// Get template
	tmpl := GetTemplate(subType)
	if tmpl == nil {
		return nil, fmt.Errorf("unknown furniture subtype: %s", subType)
	}

	// Determine rarity based on difficulty and depth
	rarity := gen.determineRarity(rng, params)

	// Select material
	material := gen.selectMaterial(rng, tmpl, params.GenreID, rarity)

	// Generate dimensions within template ranges
	width := tmpl.MinWidth + rng.Float64()*(tmpl.MaxWidth-tmpl.MinWidth)
	height := tmpl.MinHeight + rng.Float64()*(tmpl.MaxHeight-tmpl.MinHeight)
	depth := tmpl.MinDepth + rng.Float64()*(tmpl.MaxDepth-tmpl.MinDepth)

	// Scale dimensions slightly with rarity
	rarityScale := 1.0 + (float64(rarity) * 0.1)
	width *= rarityScale
	height *= rarityScale
	depth *= rarityScale

	// Generate name
	name := gen.generateName(rng, material, rarity, tmpl, params.GenreID)

	// Generate colors based on material and genre
	primaryColor := gen.getMaterialColor(rng, material, params.GenreID)
	secondaryColor := gen.getSecondaryColor(rng, primaryColor)

	// Calculate detail level
	detailLevel := tmpl.DetailComplexity * rarity.DetailMultiplier()

	// Calculate collision box (same as dimensions for most furniture)
	collisionWidth := width
	collisionDepth := depth

	// Some furniture types have smaller collision boxes (walkable items)
	if tmpl.Walkable {
		collisionWidth *= 0.5
		collisionDepth *= 0.5
	}

	// Calculate capacity if storage furniture
	capacity := tmpl.BaseCapacity
	if capacity > 0 {
		// Scale capacity with rarity
		capacity = int(float64(capacity) * rarity.DetailMultiplier())
	}

	// Calculate light intensity if lighting furniture
	lightIntensity := tmpl.BaseLightLevel
	if lightIntensity > 0 {
		// Add slight variation
		lightIntensity += (rng.Float64()*0.2 - 0.1)
		if lightIntensity < 0 {
			lightIntensity = 0
		}
		if lightIntensity > 1.0 {
			lightIntensity = 1.0
		}
	}

	// Generate ID
	id := fmt.Sprintf("furniture_%d", seed)

	// Generate description
	description := gen.generateDescription(tmpl, material, rarity, params.GenreID)

	furniture := &Furniture{
		ID:          id,
		Type:        tmpl.Type,
		SubType:     tmpl.SubType,
		Material:    material,
		Rarity:      rarity,
		GenreID:     params.GenreID,
		Name:        name,
		Description: description,

		Width:  width,
		Height: height,
		Depth:  depth,

		PrimaryColor:   primaryColor,
		SecondaryColor: secondaryColor,
		DetailLevel:    detailLevel,

		Direction:      DirNorth, // Default, can be rotated later
		Walkable:       tmpl.Walkable,
		CollisionWidth: collisionWidth,
		CollisionDepth: collisionDepth,

		Functional:     tmpl.Functional,
		Capacity:       capacity,
		LightIntensity: lightIntensity,
	}

	// Set crafting type for crafting furniture
	if tmpl.Type == TypeCrafting {
		furniture.CraftingType = tmpl.SubType
	}

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
