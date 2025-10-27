// Package recipe provides procedural recipe generation for crafting systems.
// This file implements deterministic recipe generators that create recipes
// for potions, enchanting, and magic items based on genre, difficulty, and seed.
//
// Design Philosophy:
// - Deterministic: same seed + params always generates same recipes
// - Genre-themed: recipes use genre-specific materials and naming
// - Balanced: recipes scale with skill requirements and rarity
// - Extensible: template-based system allows adding new recipe types
package recipe

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// RecipeGenerator generates procedural crafting recipes.
type RecipeGenerator struct {
	potionTemplates    map[string][]RecipeTemplate
	enchantTemplates   map[string][]RecipeTemplate
	magicItemTemplates map[string][]RecipeTemplate
	logger             *logrus.Entry
}

// RecipeTemplate defines a pattern for generating recipes.
type RecipeTemplate struct {
	NamePrefix      string
	NameSuffix      string
	RecipeType      engine.RecipeType
	RecipeRarity    engine.RecipeRarity
	OutputType      item.ItemType
	MaterialNames   []string // Pool of possible material names
	MaterialCount   [2]int   // Min and max materials required
	GoldCostRange   [2]int   // Min and max gold cost
	SkillRange      [2]int   // Min and max skill requirement
	BaseSuccessRange [2]float64 // Min and max base success chance
	CraftTimeRange  [2]float64 // Min and max craft time in seconds
}

// NewRecipeGenerator creates a new recipe generator.
func NewRecipeGenerator() *RecipeGenerator {
	return NewRecipeGeneratorWithLogger(nil)
}

// NewRecipeGeneratorWithLogger creates a new recipe generator with a logger.
func NewRecipeGeneratorWithLogger(logger *logrus.Logger) *RecipeGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "recipe")
	}

	gen := &RecipeGenerator{
		potionTemplates:    make(map[string][]RecipeTemplate),
		enchantTemplates:   make(map[string][]RecipeTemplate),
		magicItemTemplates: make(map[string][]RecipeTemplate),
		logger:             logEntry,
	}

	// Register templates for all genres
	gen.registerFantasyTemplates()
	gen.registerSciFiTemplates()
	gen.registerHorrorTemplates()
	gen.registerCyberpunkTemplates()
	gen.registerPostApocTemplates()

	// Default templates (fantasy)
	gen.potionTemplates[""] = gen.potionTemplates["fantasy"]
	gen.enchantTemplates[""] = gen.enchantTemplates["fantasy"]
	gen.magicItemTemplates[""] = gen.magicItemTemplates["fantasy"]

	if logEntry != nil {
		logEntry.Debug("recipe generator initialized")
	}

	return gen
}

// Generate creates recipes based on seed and parameters.
// Returns a slice of *engine.Recipe.
func (g *RecipeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":    seed,
			"genreID": params.GenreID,
			"depth":   params.Depth,
		}).Debug("starting recipe generation")
	}

	// Get count from custom parameters
	count := 5 // default: generate 5 recipes
	if params.Custom != nil {
		if c, ok := params.Custom["count"].(int); ok {
			count = c
		}
	}

	// Get recipe type filter from custom parameters
	var recipeTypeFilter *engine.RecipeType
	if params.Custom != nil {
		if typeStr, ok := params.Custom["type"].(string); ok {
			switch typeStr {
			case "potion":
				t := engine.RecipePotion
				recipeTypeFilter = &t
			case "enchanting":
				t := engine.RecipeEnchanting
				recipeTypeFilter = &t
			case "magic_item":
				t := engine.RecipeMagicItem
				recipeTypeFilter = &t
			}
		}
	}

	// Create random source from seed
	rng := rand.New(rand.NewSource(seed))

	// Generate recipes
	recipes := make([]*engine.Recipe, 0, count)
	for i := 0; i < count; i++ {
		// Determine recipe type
		var recipeType engine.RecipeType
		if recipeTypeFilter != nil {
			recipeType = *recipeTypeFilter
		} else {
			// Random distribution: 50% potion, 30% enchanting, 20% magic item
			roll := rng.Float64()
			if roll < 0.5 {
				recipeType = engine.RecipePotion
			} else if roll < 0.8 {
				recipeType = engine.RecipeEnchanting
			} else {
				recipeType = engine.RecipeMagicItem
			}
		}

		// Generate recipe
		recipe := g.generateRecipe(rng, params, recipeType, i)
		recipes = append(recipes, recipe)
	}

	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":         seed,
			"recipeCount":  len(recipes),
			"genreID":      params.GenreID,
		}).Debug("recipe generation complete")
	}

	return recipes, nil
}

// Validate ensures generated recipes meet quality criteria.
func (g *RecipeGenerator) Validate(result interface{}) error {
	recipes, ok := result.([]*engine.Recipe)
	if !ok {
		return fmt.Errorf("result is not []*engine.Recipe")
	}

	if len(recipes) == 0 {
		return fmt.Errorf("no recipes generated")
	}

	// Validate each recipe
	for i, recipe := range recipes {
		if recipe.ID == "" {
			return fmt.Errorf("recipe %d has empty ID", i)
		}
		if recipe.Name == "" {
			return fmt.Errorf("recipe %d has empty name", i)
		}
		if len(recipe.Materials) == 0 {
			return fmt.Errorf("recipe %d (%s) has no materials", i, recipe.Name)
		}
		if recipe.BaseSuccessChance < 0 || recipe.BaseSuccessChance > 1.0 {
			return fmt.Errorf("recipe %d (%s) has invalid success chance: %f", i, recipe.Name, recipe.BaseSuccessChance)
		}
		if recipe.CraftTimeSec <= 0 {
			return fmt.Errorf("recipe %d (%s) has invalid craft time: %f", i, recipe.Name, recipe.CraftTimeSec)
		}
	}

	return nil
}

// generateRecipe creates a single recipe from a template.
func (g *RecipeGenerator) generateRecipe(rng *rand.Rand, params procgen.GenerationParams, recipeType engine.RecipeType, index int) *engine.Recipe {
	// Get templates for genre and recipe type
	templates := g.getTemplatesForType(params.GenreID, recipeType)
	if len(templates) == 0 {
		templates = g.getTemplatesForType("fantasy", recipeType) // Fallback
	}

	// Select random template
	template := templates[rng.Intn(len(templates))]

	// Determine rarity based on depth and difficulty
	rarity := g.calculateRarity(rng, params.Depth, params.Difficulty)

	// Generate recipe ID
	recipeID := fmt.Sprintf("%s_%s_%d", params.GenreID, recipeType.String(), index)

	// Generate name
	name := fmt.Sprintf("%s %s", template.NamePrefix, template.NameSuffix)

	// Generate materials
	materialCount := template.MaterialCount[0] + rng.Intn(template.MaterialCount[1]-template.MaterialCount[0]+1)
	materials := make([]engine.MaterialRequirement, materialCount)
	for i := 0; i < materialCount; i++ {
		materialName := template.MaterialNames[rng.Intn(len(template.MaterialNames))]
		quantity := 1 + rng.Intn(3) // 1-3 of each material
		materials[i] = engine.MaterialRequirement{
			ItemName: materialName,
			Quantity: quantity,
			Optional: false,
		}
	}

	// Calculate stats based on rarity and depth
	skillRequired := template.SkillRange[0] + int(float64(template.SkillRange[1]-template.SkillRange[0])*params.Difficulty)
	skillRequired += params.Depth / 2 // Deeper dungeons have harder recipes

	goldCost := template.GoldCostRange[0] + rng.Intn(template.GoldCostRange[1]-template.GoldCostRange[0]+1)
	goldCost = int(float64(goldCost) * (1.0 + float64(rarity)*0.5)) // Scale with rarity

	baseSuccess := template.BaseSuccessRange[0] + rng.Float64()*(template.BaseSuccessRange[1]-template.BaseSuccessRange[0])
	// Higher rarity = lower base success (more challenging)
	baseSuccess -= float64(rarity) * 0.05

	craftTime := template.CraftTimeRange[0] + rng.Float64()*(template.CraftTimeRange[1]-template.CraftTimeRange[0])

	// Generate description
	description := fmt.Sprintf("A %s recipe for crafting %s", rarity.String(), name)

	return &engine.Recipe{
		ID:                recipeID,
		Name:              name,
		Description:       description,
		Type:              recipeType,
		Rarity:            rarity,
		Materials:         materials,
		GoldCost:          goldCost,
		SkillRequired:     skillRequired,
		BaseSuccessChance: baseSuccess,
		CraftTimeSec:      craftTime,
		OutputItemSeed:    int64(rng.Int63()),
		OutputItemType:    template.OutputType,
		GenreID:           params.GenreID,
	}
}

// calculateRarity determines recipe rarity based on depth and difficulty.
func (g *RecipeGenerator) calculateRarity(rng *rand.Rand, depth int, difficulty float64) engine.RecipeRarity {
	// Base chances: Common 50%, Uncommon 30%, Rare 15%, Epic 4%, Legendary 1%
	// Modified by depth and difficulty
	roll := rng.Float64()
	
	// Adjust thresholds based on depth and difficulty
	rarityBonus := (float64(depth) * 0.02) + (difficulty * 0.1)
	
	if roll < 0.50-rarityBonus {
		return engine.RecipeCommon
	} else if roll < 0.80-rarityBonus/2 {
		return engine.RecipeUncommon
	} else if roll < 0.95 {
		return engine.RecipeRare
	} else if roll < 0.99 {
		return engine.RecipeEpic
	}
	return engine.RecipeLegendary
}

// getTemplatesForType returns templates for a specific genre and recipe type.
func (g *RecipeGenerator) getTemplatesForType(genreID string, recipeType engine.RecipeType) []RecipeTemplate {
	switch recipeType {
	case engine.RecipePotion:
		if templates, ok := g.potionTemplates[genreID]; ok {
			return templates
		}
	case engine.RecipeEnchanting:
		if templates, ok := g.enchantTemplates[genreID]; ok {
			return templates
		}
	case engine.RecipeMagicItem:
		if templates, ok := g.magicItemTemplates[genreID]; ok {
			return templates
		}
	}
	return nil
}

// Template registration methods

func (g *RecipeGenerator) registerFantasyTemplates() {
	g.potionTemplates["fantasy"] = []RecipeTemplate{
		{
			NamePrefix: "Healing", NameSuffix: "Potion",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeConsumable,
			MaterialNames: []string{"Healing Herb", "Water Flask", "Honey"},
			MaterialCount: [2]int{2, 3},
			GoldCostRange: [2]int{5, 15},
			SkillRange: [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.75, 0.85},
			CraftTimeRange: [2]float64{3.0, 5.0},
		},
		{
			NamePrefix: "Mana", NameSuffix: "Elixir",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeUncommon,
			OutputType: item.TypeConsumable,
			MaterialNames: []string{"Mana Crystal", "Purified Water", "Arcane Dust"},
			MaterialCount: [2]int{2, 4},
			GoldCostRange: [2]int{15, 30},
			SkillRange: [2]int{3, 5},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange: [2]float64{5.0, 8.0},
		},
	}

	g.enchantTemplates["fantasy"] = []RecipeTemplate{
		{
			NamePrefix: "Minor", NameSuffix: "Enchantment",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Enchantment Scroll", "Magic Ink", "Silver Dust"},
			MaterialCount: [2]int{2, 3},
			GoldCostRange: [2]int{20, 40},
			SkillRange: [2]int{2, 4},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange: [2]float64{8.0, 12.0},
		},
	}

	g.magicItemTemplates["fantasy"] = []RecipeTemplate{
		{
			NamePrefix: "Apprentice", NameSuffix: "Wand",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Oak Branch", "Magic Crystal", "Silver Wire"},
			MaterialCount: [2]int{3, 4},
			GoldCostRange: [2]int{30, 60},
			SkillRange: [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange: [2]float64{10.0, 15.0},
		},
	}
}

func (g *RecipeGenerator) registerSciFiTemplates() {
	g.potionTemplates["scifi"] = []RecipeTemplate{
		{
			NamePrefix: "Nano", NameSuffix: "Stim",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeConsumable,
			MaterialNames: []string{"Nano-Gel", "Synth Fluid", "Med-Pack"},
			MaterialCount: [2]int{2, 3},
			GoldCostRange: [2]int{10, 20},
			SkillRange: [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.75, 0.85},
			CraftTimeRange: [2]float64{3.0, 5.0},
		},
	}

	g.enchantTemplates["scifi"] = []RecipeTemplate{
		{
			NamePrefix: "Basic", NameSuffix: "Mod-Chip",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Circuit Board", "Nano-Wire", "Power Cell"},
			MaterialCount: [2]int{2, 3},
			GoldCostRange: [2]int{25, 45},
			SkillRange: [2]int{2, 4},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange: [2]float64{8.0, 12.0},
		},
	}

	g.magicItemTemplates["scifi"] = []RecipeTemplate{
		{
			NamePrefix: "Plasma", NameSuffix: "Pistol",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeUncommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Plasma Core", "Weapon Frame", "Energy Coil"},
			MaterialCount: [2]int{3, 4},
			GoldCostRange: [2]int{40, 80},
			SkillRange: [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange: [2]float64{10.0, 15.0},
		},
	}
}

func (g *RecipeGenerator) registerHorrorTemplates() {
	g.potionTemplates["horror"] = []RecipeTemplate{
		{
			NamePrefix: "Blood", NameSuffix: "Tincture",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeConsumable,
			MaterialNames: []string{"Dried Blood", "Bone Dust", "Dark Herb"},
			MaterialCount: [2]int{2, 3},
			GoldCostRange: [2]int{8, 18},
			SkillRange: [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange: [2]float64{4.0, 6.0},
		},
	}

	g.enchantTemplates["horror"] = []RecipeTemplate{
		{
			NamePrefix: "Cursed", NameSuffix: "Binding",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeUncommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Ritual Scroll", "Soul Fragment", "Black Ink"},
			MaterialCount: [2]int{2, 4},
			GoldCostRange: [2]int{20, 50},
			SkillRange: [2]int{3, 5},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange: [2]float64{10.0, 15.0},
		},
	}

	g.magicItemTemplates["horror"] = []RecipeTemplate{
		{
			NamePrefix: "Bone", NameSuffix: "Dagger",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Human Bone", "Dark Crystal", "Sinew"},
			MaterialCount: [2]int{3, 4},
			GoldCostRange: [2]int{35, 65},
			SkillRange: [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange: [2]float64{12.0, 18.0},
		},
	}
}

func (g *RecipeGenerator) registerCyberpunkTemplates() {
	g.potionTemplates["cyberpunk"] = []RecipeTemplate{
		{
			NamePrefix: "Street", NameSuffix: "Juice",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeConsumable,
			MaterialNames: []string{"Synth-Chem", "Neuro-Booster", "Filter Capsule"},
			MaterialCount: [2]int{2, 3},
			GoldCostRange: [2]int{12, 22},
			SkillRange: [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.75, 0.85},
			CraftTimeRange: [2]float64{3.0, 5.0},
		},
	}

	g.enchantTemplates["cyberpunk"] = []RecipeTemplate{
		{
			NamePrefix: "Neural", NameSuffix: "Upgrade",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeUncommon,
			OutputType: item.TypeAccessory,
			MaterialNames: []string{"Neural Link", "Bio-Circuit", "Interface Chip"},
			MaterialCount: [2]int{2, 4},
			GoldCostRange: [2]int{30, 60},
			SkillRange: [2]int{3, 5},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange: [2]float64{8.0, 12.0},
		},
	}

	g.magicItemTemplates["cyberpunk"] = []RecipeTemplate{
		{
			NamePrefix: "Cyber", NameSuffix: "Blade",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeUncommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Titanium Alloy", "Mono-Wire", "Power Core"},
			MaterialCount: [2]int{3, 4},
			GoldCostRange: [2]int{45, 85},
			SkillRange: [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange: [2]float64{10.0, 15.0},
		},
	}
}

func (g *RecipeGenerator) registerPostApocTemplates() {
	g.potionTemplates["postapoc"] = []RecipeTemplate{
		{
			NamePrefix: "Wasteland", NameSuffix: "Remedy",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeConsumable,
			MaterialNames: []string{"Purified Water", "Scrap Medicine", "Mutant Plant"},
			MaterialCount: [2]int{2, 3},
			GoldCostRange: [2]int{6, 16},
			SkillRange: [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange: [2]float64{4.0, 6.0},
		},
	}

	g.enchantTemplates["postapoc"] = []RecipeTemplate{
		{
			NamePrefix: "Scrap", NameSuffix: "Modification",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Scrap Metal", "Duct Tape", "Rusty Nails"},
			MaterialCount: [2]int{2, 3},
			GoldCostRange: [2]int{15, 30},
			SkillRange: [2]int{2, 4},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange: [2]float64{6.0, 10.0},
		},
	}

	g.magicItemTemplates["postapoc"] = []RecipeTemplate{
		{
			NamePrefix: "Makeshift", NameSuffix: "Weapon",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeCommon,
			OutputType: item.TypeWeapon,
			MaterialNames: []string{"Scrap Metal", "Pipe", "Wire"},
			MaterialCount: [2]int{3, 4},
			GoldCostRange: [2]int{25, 50},
			SkillRange: [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange: [2]float64{8.0, 12.0},
		},
	}
}
