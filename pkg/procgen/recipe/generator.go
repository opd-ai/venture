// Package recipe provides procedural recipe generation for crafting systems.
// This file implements deterministic recipe generators that create recipes
// for potions, enchanting, and magic items based on genre, difficulty, and seed.
//
// Design Philosophy:
// - Deterministic: same seed + params always generates same recipes
// - Genre-themed: recipes use genre-specific materials and naming
// - Balanced: recipes scale with skill requirements and rarity
// - Extensible: template-based system allows adding new recipe types
//
// # Extending Templates
//
// Recipe templates are registered in initializeTemplates (called from NewGenerator).
// To add new recipe types or materials for a genre, add entries to
// potionTemplates, enchantingTemplates, or magicItemTemplates in that method.
// Mod authors using the JSON mod system (pkg/modding) can contribute new
// recipe templates via code contributions to this file — the mod system
// does not currently support runtime recipe template injection.
package recipe

import (
	"fmt"
	"math"
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
	cookingTemplates   map[string][]RecipeTemplate
	smithingTemplates  map[string][]RecipeTemplate
	logger             *logrus.Entry
}

// RecipeTemplate defines a pattern for generating recipes.
// Templates are used to create procedurally generated recipes with appropriate
// naming, materials, costs, and requirements for different genres.
type RecipeTemplate struct {
	// NamePrefix is prepended to the generated recipe name (e.g., "Mystic" in "Mystic Healing Potion")
	NamePrefix string
	// NameSuffix is appended to the generated recipe name (e.g., "of Power" in "Elixir of Power")
	NameSuffix string
	// RecipeType categorizes the recipe (potion, enchanting, magic_item, cooking, smithing)
	RecipeType engine.RecipeType
	// RecipeRarity determines the rarity tier (common, uncommon, rare, epic, legendary)
	RecipeRarity engine.RecipeRarity
	// OutputType specifies the item type produced by the recipe
	OutputType item.ItemType
	// MaterialNames is the pool of possible material names for ingredient selection
	MaterialNames []string
	// MaterialCount specifies [min, max] number of materials required (inclusive)
	MaterialCount [2]int
	// MinQuantity is the minimum quantity per material ingredient (default 1 if zero)
	MinQuantity int
	// MaxQuantity is the maximum quantity per material ingredient (default 3 if zero)
	MaxQuantity int
	// GoldCostRange specifies [min, max] gold cost for crafting (inclusive)
	GoldCostRange [2]int
	// SkillRange specifies [min, max] skill requirement to attempt crafting (inclusive)
	SkillRange [2]int
	// BaseSuccessRange specifies [min, max] base success chance (0.0-1.0, inclusive)
	BaseSuccessRange [2]float64
	// CraftTimeRange specifies [min, max] craft time in seconds (inclusive)
	CraftTimeRange [2]float64
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
		cookingTemplates:   make(map[string][]RecipeTemplate),
		smithingTemplates:  make(map[string][]RecipeTemplate),
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
	gen.cookingTemplates[""] = gen.cookingTemplates["fantasy"]
	gen.smithingTemplates[""] = gen.smithingTemplates["fantasy"]

	if logEntry != nil {
		logEntry.Debug("recipe generator initialized")
	}

	return gen
}

// Generate creates recipes based on seed and parameters.
// Returns a slice of *engine.Recipe.
func (g *RecipeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	g.logGenerationStart(seed, params)

	count := g.extractRecipeCount(params)
	recipeTypeFilter := g.extractRecipeTypeFilter(params)

	rng := rand.New(rand.NewSource(seed))
	recipes := g.generateRecipes(rng, params, count, recipeTypeFilter)

	g.logGenerationComplete(seed, params.GenreID, len(recipes))
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
	// Get templates for genre and recipe type, with fantasy fallback
	templates := g.getTemplatesForType(params.GenreID, recipeType)
	if len(templates) == 0 {
		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"requested_genre": params.GenreID,
				"recipe_type":     recipeType,
				"fallback_to":     "fantasy",
			}).Warn("No templates found for requested genre, falling back to fantasy")
		}
		templates = g.getTemplatesForType("fantasy", recipeType)
	}
	if len(templates) == 0 {
		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"requested_genre": params.GenreID,
				"recipe_type":     recipeType,
				"fallback_to":     "generic",
			}).Warn("No fantasy templates found, falling back to generic templates")
		}
		templates = g.getTemplatesForType("", recipeType)
	}
	if len(templates) == 0 {
		if g.logger != nil {
			g.logger.WithFields(logrus.Fields{
				"requested_genre": params.GenreID,
				"recipe_type":     recipeType,
				"fallback_to":     "any_potion",
			}).Error("No templates found for any genre, using first available potion template")
		}
		// Last resort: use any available potion template to avoid panic
		for _, tmplList := range g.potionTemplates {
			if len(tmplList) > 0 {
				templates = tmplList
				break
			}
		}
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
		minQ, maxQ := template.MinQuantity, template.MaxQuantity
		if minQ <= 0 {
			minQ = 1
		}
		if maxQ <= 0 || maxQ < minQ {
			maxQ = 3
		}
		quantity := minQ + rng.Intn(maxQ-minQ+1)
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
	// Higher rarity = lower base success (more challenging), clamped to valid range
	baseSuccess -= float64(rarity) * 0.05
	baseSuccess = math.Max(0.05, math.Min(1.0, baseSuccess))

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

// logGenerationStart logs the start of recipe generation.
func (g *RecipeGenerator) logGenerationStart(seed int64, params procgen.GenerationParams) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":    seed,
			"genreID": params.GenreID,
			"depth":   params.Depth,
		}).Debug("starting recipe generation")
	}
}

// extractRecipeCount gets the recipe count from custom parameters, defaults to 5.
func (g *RecipeGenerator) extractRecipeCount(params procgen.GenerationParams) int {
	count := 5
	if params.Custom != nil {
		if c, ok := params.Custom["count"].(int); ok && c > 0 {
			count = c
		}
	}
	return count
}

// extractRecipeTypeFilter gets the recipe type filter from custom parameters.
func (g *RecipeGenerator) extractRecipeTypeFilter(params procgen.GenerationParams) *engine.RecipeType {
	if params.Custom == nil {
		return nil
	}

	typeStr, ok := params.Custom["type"].(string)
	if !ok {
		return nil
	}

	switch typeStr {
	case "potion":
		t := engine.RecipePotion
		return &t
	case "enchanting":
		t := engine.RecipeEnchanting
		return &t
	case "magic_item":
		t := engine.RecipeMagicItem
		return &t
	case "cooking":
		t := engine.RecipeCooking
		return &t
	case "smithing":
		t := engine.RecipeSmithing
		return &t
	}
	return nil
}

// generateRecipes creates the specified number of recipes.
func (g *RecipeGenerator) generateRecipes(rng *rand.Rand, params procgen.GenerationParams, count int, typeFilter *engine.RecipeType) []*engine.Recipe {
	recipes := make([]*engine.Recipe, 0, count)
	for i := 0; i < count; i++ {
		recipeType := g.determineRecipeType(rng, typeFilter)
		recipe := g.generateRecipe(rng, params, recipeType, i)
		recipes = append(recipes, recipe)
	}
	return recipes
}

// determineRecipeType selects recipe type based on filter or random distribution.
func (g *RecipeGenerator) determineRecipeType(rng *rand.Rand, typeFilter *engine.RecipeType) engine.RecipeType {
	if typeFilter != nil {
		return *typeFilter
	}

	roll := rng.Float64()
	if roll < 0.30 {
		return engine.RecipePotion
	} else if roll < 0.50 {
		return engine.RecipeEnchanting
	} else if roll < 0.70 {
		return engine.RecipeMagicItem
	} else if roll < 0.85 {
		return engine.RecipeCooking
	}
	return engine.RecipeSmithing
}

// logGenerationComplete logs the completion of recipe generation.
func (g *RecipeGenerator) logGenerationComplete(seed int64, genreID string, recipeCount int) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":        seed,
			"recipeCount": recipeCount,
			"genreID":     genreID,
		}).Debug("recipe generation complete")
	}
}

// calculateRarity determines recipe rarity based on depth and difficulty.
func (g *RecipeGenerator) calculateRarity(rng *rand.Rand, depth int, difficulty float64) engine.RecipeRarity {
	// Base chances: Common 50%, Uncommon 30%, Rare 15%, Epic 4%, Legendary 1%
	// Modified by depth and difficulty
	roll := rng.Float64()

	// Adjust thresholds based on depth and difficulty, clamped to prevent inversion
	rarityBonus := math.Min(0.45, (float64(depth)*0.02)+(difficulty*0.1))

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
	case engine.RecipeCooking:
		if templates, ok := g.cookingTemplates[genreID]; ok {
			return templates
		}
	case engine.RecipeSmithing:
		if templates, ok := g.smithingTemplates[genreID]; ok {
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
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Healing Herb", "Water Flask", "Honey"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{5, 15},
			SkillRange:       [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.75, 0.85},
			CraftTimeRange:   [2]float64{3.0, 5.0},
		},
		{
			NamePrefix: "Mana", NameSuffix: "Elixir",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeUncommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Mana Crystal", "Purified Water", "Arcane Dust"},
			MaterialCount:    [2]int{2, 4},
			GoldCostRange:    [2]int{15, 30},
			SkillRange:       [2]int{3, 5},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange:   [2]float64{5.0, 8.0},
		},
	}

	g.enchantTemplates["fantasy"] = []RecipeTemplate{
		{
			NamePrefix: "Minor", NameSuffix: "Enchantment",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Enchantment Scroll", "Magic Ink", "Silver Dust"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{20, 40},
			SkillRange:       [2]int{2, 4},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{8.0, 12.0},
		},
	}

	g.magicItemTemplates["fantasy"] = []RecipeTemplate{
		{
			NamePrefix: "Apprentice", NameSuffix: "Wand",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Oak Branch", "Magic Crystal", "Silver Wire"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{30, 60},
			SkillRange:       [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange:   [2]float64{10.0, 15.0},
		},
	}

	g.cookingTemplates["fantasy"] = []RecipeTemplate{
		{
			NamePrefix: "Hearty", NameSuffix: "Stew",
			RecipeType: engine.RecipeCooking, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Meat", "Vegetables", "Spices", "Water"},
			MaterialCount:    [2]int{2, 4},
			GoldCostRange:    [2]int{3, 10},
			SkillRange:       [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.80, 0.90},
			CraftTimeRange:   [2]float64{5.0, 10.0},
		},
		{
			NamePrefix: "Stamina", NameSuffix: "Pie",
			RecipeType: engine.RecipeCooking, RecipeRarity: engine.RecipeUncommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Flour", "Butter", "Honey", "Berries"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{10, 20},
			SkillRange:       [2]int{2, 4},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{8.0, 12.0},
		},
	}

	g.smithingTemplates["fantasy"] = []RecipeTemplate{
		{
			NamePrefix: "Iron", NameSuffix: "Sword",
			RecipeType: engine.RecipeSmithing, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Iron Ore", "Coal", "Leather Grip"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{20, 40},
			SkillRange:       [2]int{3, 5},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{10.0, 15.0},
		},
		{
			NamePrefix: "Steel", NameSuffix: "Plate Armor",
			RecipeType: engine.RecipeSmithing, RecipeRarity: engine.RecipeUncommon,
			OutputType:       item.TypeArmor,
			MaterialNames:    []string{"Steel Ingot", "Leather Straps", "Iron Rivets"},
			MaterialCount:    [2]int{4, 5},
			GoldCostRange:    [2]int{40, 80},
			SkillRange:       [2]int{5, 8},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange:   [2]float64{15.0, 25.0},
		},
	}
}

func (g *RecipeGenerator) registerSciFiTemplates() {
	g.potionTemplates["scifi"] = []RecipeTemplate{
		{
			NamePrefix: "Nano", NameSuffix: "Stim",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Nano-Gel", "Synth Fluid", "Med-Pack"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{10, 20},
			SkillRange:       [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.75, 0.85},
			CraftTimeRange:   [2]float64{3.0, 5.0},
		},
	}

	g.enchantTemplates["scifi"] = []RecipeTemplate{
		{
			NamePrefix: "Basic", NameSuffix: "Mod-Chip",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Circuit Board", "Nano-Wire", "Power Cell"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{25, 45},
			SkillRange:       [2]int{2, 4},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{8.0, 12.0},
		},
	}

	g.magicItemTemplates["scifi"] = []RecipeTemplate{
		{
			NamePrefix: "Plasma", NameSuffix: "Pistol",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeUncommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Plasma Core", "Weapon Frame", "Energy Coil"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{40, 80},
			SkillRange:       [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange:   [2]float64{10.0, 15.0},
		},
	}

	g.cookingTemplates["scifi"] = []RecipeTemplate{
		{
			NamePrefix: "Nutrient", NameSuffix: "Paste",
			RecipeType: engine.RecipeCooking, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Protein Powder", "Vitamin Mix", "Flavor Pack"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{5, 12},
			SkillRange:       [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.80, 0.90},
			CraftTimeRange:   [2]float64{3.0, 5.0},
		},
	}

	g.smithingTemplates["scifi"] = []RecipeTemplate{
		{
			NamePrefix: "Titanium", NameSuffix: "Armor Plating",
			RecipeType: engine.RecipeSmithing, RecipeRarity: engine.RecipeUncommon,
			OutputType:       item.TypeArmor,
			MaterialNames:    []string{"Titanium Alloy", "Ceramic Plate", "Flex-Gel"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{50, 100},
			SkillRange:       [2]int{4, 7},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange:   [2]float64{12.0, 18.0},
		},
	}
}

func (g *RecipeGenerator) registerHorrorTemplates() {
	g.potionTemplates["horror"] = []RecipeTemplate{
		{
			NamePrefix: "Blood", NameSuffix: "Tincture",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Dried Blood", "Bone Dust", "Dark Herb"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{8, 18},
			SkillRange:       [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{4.0, 6.0},
		},
	}

	g.enchantTemplates["horror"] = []RecipeTemplate{
		{
			NamePrefix: "Cursed", NameSuffix: "Binding",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeUncommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Ritual Scroll", "Soul Fragment", "Black Ink"},
			MaterialCount:    [2]int{2, 4},
			GoldCostRange:    [2]int{20, 50},
			SkillRange:       [2]int{3, 5},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange:   [2]float64{10.0, 15.0},
		},
	}

	g.magicItemTemplates["horror"] = []RecipeTemplate{
		{
			NamePrefix: "Bone", NameSuffix: "Dagger",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Human Bone", "Dark Crystal", "Sinew"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{35, 65},
			SkillRange:       [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange:   [2]float64{12.0, 18.0},
		},
	}

	g.cookingTemplates["horror"] = []RecipeTemplate{
		{
			NamePrefix: "Grave", NameSuffix: "Rations",
			RecipeType: engine.RecipeCooking, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Preserved Meat", "Fungus", "Bitter Herb"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{4, 10},
			SkillRange:       [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.75, 0.85},
			CraftTimeRange:   [2]float64{4.0, 8.0},
		},
	}

	g.smithingTemplates["horror"] = []RecipeTemplate{
		{
			NamePrefix: "Rusted", NameSuffix: "Chainmail",
			RecipeType: engine.RecipeSmithing, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeArmor,
			MaterialNames:    []string{"Scrap Iron", "Old Chain", "Blood Oil"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{25, 50},
			SkillRange:       [2]int{3, 6},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{10.0, 15.0},
		},
	}
}

func (g *RecipeGenerator) registerCyberpunkTemplates() {
	g.potionTemplates["cyberpunk"] = []RecipeTemplate{
		{
			NamePrefix: "Street", NameSuffix: "Juice",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Synth-Chem", "Neuro-Booster", "Filter Capsule"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{12, 22},
			SkillRange:       [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.75, 0.85},
			CraftTimeRange:   [2]float64{3.0, 5.0},
		},
	}

	g.enchantTemplates["cyberpunk"] = []RecipeTemplate{
		{
			NamePrefix: "Neural", NameSuffix: "Upgrade",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeUncommon,
			OutputType:       item.TypeAccessory,
			MaterialNames:    []string{"Neural Link", "Bio-Circuit", "Interface Chip"},
			MaterialCount:    [2]int{2, 4},
			GoldCostRange:    [2]int{30, 60},
			SkillRange:       [2]int{3, 5},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange:   [2]float64{8.0, 12.0},
		},
	}

	g.magicItemTemplates["cyberpunk"] = []RecipeTemplate{
		{
			NamePrefix: "Cyber", NameSuffix: "Blade",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeUncommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Titanium Alloy", "Mono-Wire", "Power Core"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{45, 85},
			SkillRange:       [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.60, 0.70},
			CraftTimeRange:   [2]float64{10.0, 15.0},
		},
	}

	g.cookingTemplates["cyberpunk"] = []RecipeTemplate{
		{
			NamePrefix: "Synth", NameSuffix: "Ramen",
			RecipeType: engine.RecipeCooking, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Noodle Block", "Flavor Powder", "Synth-Protein"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{2, 8},
			SkillRange:       [2]int{0, 1},
			BaseSuccessRange: [2]float64{0.85, 0.95},
			CraftTimeRange:   [2]float64{2.0, 4.0},
		},
	}

	g.smithingTemplates["cyberpunk"] = []RecipeTemplate{
		{
			NamePrefix: "Cyber", NameSuffix: "Exo-Suit",
			RecipeType: engine.RecipeSmithing, RecipeRarity: engine.RecipeRare,
			OutputType:       item.TypeArmor,
			MaterialNames:    []string{"Carbon Fiber", "Servo Motors", "Power Cell", "Neural Interface"},
			MaterialCount:    [2]int{4, 5},
			GoldCostRange:    [2]int{70, 120},
			SkillRange:       [2]int{6, 9},
			BaseSuccessRange: [2]float64{0.55, 0.65},
			CraftTimeRange:   [2]float64{18.0, 30.0},
		},
	}
}

func (g *RecipeGenerator) registerPostApocTemplates() {
	g.potionTemplates["postapoc"] = []RecipeTemplate{
		{
			NamePrefix: "Wasteland", NameSuffix: "Remedy",
			RecipeType: engine.RecipePotion, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Purified Water", "Scrap Medicine", "Mutant Plant"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{6, 16},
			SkillRange:       [2]int{0, 2},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{4.0, 6.0},
		},
	}

	g.enchantTemplates["postapoc"] = []RecipeTemplate{
		{
			NamePrefix: "Scrap", NameSuffix: "Modification",
			RecipeType: engine.RecipeEnchanting, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Scrap Metal", "Duct Tape", "Rusty Nails"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{15, 30},
			SkillRange:       [2]int{2, 4},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{6.0, 10.0},
		},
	}

	g.magicItemTemplates["postapoc"] = []RecipeTemplate{
		{
			NamePrefix: "Makeshift", NameSuffix: "Weapon",
			RecipeType: engine.RecipeMagicItem, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeWeapon,
			MaterialNames:    []string{"Scrap Metal", "Pipe", "Wire"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{25, 50},
			SkillRange:       [2]int{5, 7},
			BaseSuccessRange: [2]float64{0.65, 0.75},
			CraftTimeRange:   [2]float64{8.0, 12.0},
		},
	}

	g.cookingTemplates["postapoc"] = []RecipeTemplate{
		{
			NamePrefix: "Canned", NameSuffix: "Rations",
			RecipeType: engine.RecipeCooking, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeConsumable,
			MaterialNames:    []string{"Canned Goods", "Clean Water", "Salt"},
			MaterialCount:    [2]int{2, 3},
			GoldCostRange:    [2]int{3, 9},
			SkillRange:       [2]int{0, 1},
			BaseSuccessRange: [2]float64{0.85, 0.95},
			CraftTimeRange:   [2]float64{2.0, 5.0},
		},
	}

	g.smithingTemplates["postapoc"] = []RecipeTemplate{
		{
			NamePrefix: "Salvaged", NameSuffix: "Body Armor",
			RecipeType: engine.RecipeSmithing, RecipeRarity: engine.RecipeCommon,
			OutputType:       item.TypeArmor,
			MaterialNames:    []string{"Scrap Metal", "Leather", "Bolts"},
			MaterialCount:    [2]int{3, 4},
			GoldCostRange:    [2]int{20, 40},
			SkillRange:       [2]int{3, 5},
			BaseSuccessRange: [2]float64{0.70, 0.80},
			CraftTimeRange:   [2]float64{10.0, 15.0},
		},
	}
}
