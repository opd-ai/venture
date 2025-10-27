// Package engine provides crafting components for the ECS.
// This file defines components for recipe-based crafting mechanics including
// potion brewing, enchanting, and magic item crafting. The crafting system
// integrates with the existing item generation, inventory, and skill systems.
//
// Design Philosophy:
// - Components contain only data, no behavior
// - Recipes are deterministic based on seed and ingredients
// - Success chance scales with skill level (crafting skill)
// - Failed crafts consume 50% of materials (risk/reward balance)
// - Server-authoritative crafting prevents multiplayer exploits
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// RecipeType represents different categories of crafting.
type RecipeType int

const (
	// RecipePotion represents potion brewing (herbs + flask -> consumables)
	RecipePotion RecipeType = iota
	// RecipeEnchanting represents equipment enhancement (item + scroll + gold -> enhanced item)
	RecipeEnchanting
	// RecipeMagicItem represents magic item creation (base + essence + materials -> magic item)
	RecipeMagicItem
)

// String returns the string representation of a recipe type.
func (r RecipeType) String() string {
	switch r {
	case RecipePotion:
		return "potion"
	case RecipeEnchanting:
		return "enchanting"
	case RecipeMagicItem:
		return "magic_item"
	default:
		return "unknown"
	}
}

// RecipeRarity represents the rarity tier of a recipe, affecting success chance and output quality.
type RecipeRarity int

const (
	// RecipeCommon represents basic recipes available to all players
	RecipeCommon RecipeRarity = iota
	// RecipeUncommon represents recipes requiring some skill or discovery
	RecipeUncommon
	// RecipeRare represents advanced recipes requiring higher skill
	RecipeRare
	// RecipeEpic represents master-level recipes with powerful outputs
	RecipeEpic
	// RecipeLegendary represents legendary recipes with exceptional outputs
	RecipeLegendary
)

// String returns the string representation of a recipe rarity.
func (r RecipeRarity) String() string {
	switch r {
	case RecipeCommon:
		return "common"
	case RecipeUncommon:
		return "uncommon"
	case RecipeRare:
		return "rare"
	case RecipeEpic:
		return "epic"
	case RecipeLegendary:
		return "legendary"
	default:
		return "unknown"
	}
}

// MaterialRequirement represents a single ingredient needed for a recipe.
type MaterialRequirement struct {
	// ItemName is the name of the required item (e.g., "Healing Herb", "Iron Ore")
	ItemName string

	// Quantity is how many of this item are needed
	Quantity int

	// Optional indicates this material is not strictly required (can substitute gold)
	Optional bool

	// ItemType restricts to specific item types (nil = any item with matching name)
	ItemType *item.ItemType
}

// Recipe represents a crafting recipe with requirements and output.
type Recipe struct {
	// ID is a unique identifier for this recipe (used for deterministic generation)
	ID string

	// Name is the display name of this recipe
	Name string

	// Description explains what this recipe creates
	Description string

	// Type is the crafting category (potion, enchanting, magic item)
	Type RecipeType

	// Rarity affects difficulty and output quality
	Rarity RecipeRarity

	// Materials lists all required ingredients
	Materials []MaterialRequirement

	// GoldCost is the additional gold required to craft (in addition to materials)
	GoldCost int

	// SkillRequired is the minimum crafting skill level needed
	SkillRequired int

	// BaseSuccessChance is the success probability at minimum skill (0.0-1.0)
	BaseSuccessChance float64

	// CraftTimeSec is how long the crafting takes (for UI feedback)
	CraftTimeSec float64

	// OutputItemSeed is used to generate the output item deterministically
	OutputItemSeed int64

	// OutputItemType specifies what type of item is created
	OutputItemType item.ItemType

	// GenreID links recipe to genre for thematic consistency
	GenreID string
}

// GetEffectiveSuccessChance calculates actual success chance based on skill level.
// Formula: BaseSuccessChance + (0.05 * (skillLevel - SkillRequired))
// Capped at 95% to maintain some risk.
func (r *Recipe) GetEffectiveSuccessChance(skillLevel int) float64 {
	if skillLevel < r.SkillRequired {
		return 0.0 // Cannot craft without minimum skill
	}

	// Increase success chance by 5% per skill level above requirement
	bonus := 0.05 * float64(skillLevel-r.SkillRequired)
	effectiveChance := r.BaseSuccessChance + bonus

	// Cap at 95% to maintain risk
	if effectiveChance > 0.95 {
		return 0.95
	}

	return effectiveChance
}

// RecipeKnowledgeComponent tracks which recipes an entity has discovered.
// Players learn recipes through gameplay (world drops, quest rewards, NPC teaching).
type RecipeKnowledgeComponent struct {
	// KnownRecipes is a map of recipe ID -> Recipe
	KnownRecipes map[string]*Recipe

	// RecipeSlots limits how many recipes can be learned (0 = unlimited)
	RecipeSlots int
}

// Type returns the component type identifier.
func (r *RecipeKnowledgeComponent) Type() string {
	return "recipe_knowledge"
}

// NewRecipeKnowledgeComponent creates a recipe knowledge component.
func NewRecipeKnowledgeComponent(recipeSlots int) *RecipeKnowledgeComponent {
	return &RecipeKnowledgeComponent{
		KnownRecipes: make(map[string]*Recipe),
		RecipeSlots:  recipeSlots,
	}
}

// KnowsRecipe checks if a recipe is known.
func (r *RecipeKnowledgeComponent) KnowsRecipe(recipeID string) bool {
	_, ok := r.KnownRecipes[recipeID]
	return ok
}

// LearnRecipe adds a recipe to known recipes.
// Returns false if recipe slots are full.
func (r *RecipeKnowledgeComponent) LearnRecipe(recipe *Recipe) bool {
	if r.KnownRecipes == nil {
		r.KnownRecipes = make(map[string]*Recipe)
	}

	// Check if already known
	if r.KnowsRecipe(recipe.ID) {
		return true
	}

	// Check slot limit (0 = unlimited)
	if r.RecipeSlots > 0 && len(r.KnownRecipes) >= r.RecipeSlots {
		return false
	}

	r.KnownRecipes[recipe.ID] = recipe
	return true
}

// CraftingStationComponent marks an entity as a crafting station.
// Different station types enable different recipe categories.
// Optional: can craft without stations, but stations provide bonuses.
type CraftingStationComponent struct {
	// StationType indicates what can be crafted here
	StationType RecipeType

	// BonusSuccessChance is added to recipe success chance when using this station
	BonusSuccessChance float64

	// CraftTimeMultiplier speeds up crafting (0.5 = twice as fast)
	CraftTimeMultiplier float64

	// Available indicates if station is usable (not in use by another player)
	Available bool
}

// Type returns the component type identifier.
func (c *CraftingStationComponent) Type() string {
	return "crafting_station"
}

// NewCraftingStationComponent creates a crafting station component.
func NewCraftingStationComponent(stationType RecipeType) *CraftingStationComponent {
	return &CraftingStationComponent{
		StationType:         stationType,
		BonusSuccessChance:  0.05, // 5% bonus when using station
		CraftTimeMultiplier: 0.75, // 25% faster at station
		Available:           true,
	}
}

// CraftingProgressComponent tracks ongoing crafting for an entity.
// Only one craft can be in progress at a time per entity.
type CraftingProgressComponent struct {
	// CurrentRecipe is the recipe being crafted
	CurrentRecipe *Recipe

	// ElapsedTimeSec tracks crafting progress
	ElapsedTimeSec float64

	// RequiredTimeSec is total time needed (from recipe, modified by station)
	RequiredTimeSec float64

	// MaterialsConsumed indicates if materials were already deducted from inventory
	MaterialsConsumed bool

	// UsingStationID is the entity ID of the crafting station being used (0 = no station)
	UsingStationID uint64
}

// Type returns the component type identifier.
func (c *CraftingProgressComponent) Type() string {
	return "crafting_progress"
}

// NewCraftingProgressComponent creates a crafting progress component.
func NewCraftingProgressComponent(recipe *Recipe, requiredTime float64, stationID uint64) *CraftingProgressComponent {
	return &CraftingProgressComponent{
		CurrentRecipe:     recipe,
		ElapsedTimeSec:    0,
		RequiredTimeSec:   requiredTime,
		MaterialsConsumed: false,
		UsingStationID:    stationID,
	}
}

// GetProgress returns crafting progress as a percentage (0.0-1.0).
func (c *CraftingProgressComponent) GetProgress() float64 {
	if c.RequiredTimeSec <= 0 {
		return 1.0
	}
	progress := c.ElapsedTimeSec / c.RequiredTimeSec
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// IsComplete returns true if crafting is finished.
func (c *CraftingProgressComponent) IsComplete() bool {
	return c.ElapsedTimeSec >= c.RequiredTimeSec
}

// CraftingSkillComponent tracks crafting skill progression.
// Skill level affects success chance and unlocks higher-tier recipes.
type CraftingSkillComponent struct {
	// SkillLevel is the current crafting skill level (0-100)
	SkillLevel int

	// Experience points toward next level
	Experience int

	// ExperienceToNextLevel defines leveling threshold
	ExperienceToNextLevel int
}

// Type returns the component type identifier.
func (c *CraftingSkillComponent) Type() string {
	return "crafting_skill"
}

// NewCraftingSkillComponent creates a crafting skill component.
func NewCraftingSkillComponent() *CraftingSkillComponent {
	return &CraftingSkillComponent{
		SkillLevel:            0,
		Experience:            0,
		ExperienceToNextLevel: 100, // Linear progression: 100 XP per level
	}
}

// AddExperience adds XP and handles level-ups.
// Returns true if leveled up.
func (c *CraftingSkillComponent) AddExperience(xp int) bool {
	c.Experience += xp
	if c.Experience >= c.ExperienceToNextLevel {
		c.SkillLevel++
		c.Experience -= c.ExperienceToNextLevel
		c.ExperienceToNextLevel = 100 * (c.SkillLevel + 1) // Scaling XP requirements
		return true
	}
	return false
}
