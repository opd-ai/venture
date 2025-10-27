package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestRecipeType_String tests RecipeType string representation.
func TestRecipeType_String(t *testing.T) {
	tests := []struct {
		name     string
		recType  RecipeType
		expected string
	}{
		{"potion", RecipePotion, "potion"},
		{"enchanting", RecipeEnchanting, "enchanting"},
		{"magic item", RecipeMagicItem, "magic_item"},
		{"unknown", RecipeType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.recType.String(); got != tt.expected {
				t.Errorf("RecipeType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRecipeRarity_String tests RecipeRarity string representation.
func TestRecipeRarity_String(t *testing.T) {
	tests := []struct {
		name     string
		rarity   RecipeRarity
		expected string
	}{
		{"common", RecipeCommon, "common"},
		{"uncommon", RecipeUncommon, "uncommon"},
		{"rare", RecipeRare, "rare"},
		{"epic", RecipeEpic, "epic"},
		{"legendary", RecipeLegendary, "legendary"},
		{"unknown", RecipeRarity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rarity.String(); got != tt.expected {
				t.Errorf("RecipeRarity.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRecipe_GetEffectiveSuccessChance tests success chance calculation.
func TestRecipe_GetEffectiveSuccessChance(t *testing.T) {
	tests := []struct {
		name           string
		baseChance     float64
		skillRequired  int
		skillLevel     int
		expectedChance float64
		description    string
	}{
		{
			name:           "below minimum skill",
			baseChance:     0.50,
			skillRequired:  5,
			skillLevel:     3,
			expectedChance: 0.0,
			description:    "cannot craft without minimum skill",
		},
		{
			name:           "at minimum skill",
			baseChance:     0.50,
			skillRequired:  5,
			skillLevel:     5,
			expectedChance: 0.50,
			description:    "base chance at minimum skill",
		},
		{
			name:           "5 levels above minimum",
			baseChance:     0.50,
			skillRequired:  5,
			skillLevel:     10,
			expectedChance: 0.75, // 0.50 + 0.05*5
			description:    "5% bonus per level above requirement",
		},
		{
			name:           "caps at 95%",
			baseChance:     0.50,
			skillRequired:  5,
			skillLevel:     20,
			expectedChance: 0.95,
			description:    "capped at 95% to maintain risk",
		},
		{
			name:           "high base chance still caps",
			baseChance:     0.90,
			skillRequired:  1,
			skillLevel:     10,
			expectedChance: 0.95,
			description:    "even with high base, cap applies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe := &Recipe{
				BaseSuccessChance: tt.baseChance,
				SkillRequired:     tt.skillRequired,
			}
			got := recipe.GetEffectiveSuccessChance(tt.skillLevel)
			if got != tt.expectedChance {
				t.Errorf("GetEffectiveSuccessChance() = %v, want %v (%s)", got, tt.expectedChance, tt.description)
			}
		})
	}
}

// TestNewRecipeKnowledgeComponent tests component creation.
func TestNewRecipeKnowledgeComponent(t *testing.T) {
	comp := NewRecipeKnowledgeComponent(10)
	if comp == nil {
		t.Fatal("NewRecipeKnowledgeComponent returned nil")
	}
	if comp.RecipeSlots != 10 {
		t.Errorf("RecipeSlots = %v, want 10", comp.RecipeSlots)
	}
	if comp.KnownRecipes == nil {
		t.Error("KnownRecipes map not initialized")
	}
	if comp.Type() != "recipe_knowledge" {
		t.Errorf("Type() = %v, want 'recipe_knowledge'", comp.Type())
	}
}

// TestRecipeKnowledgeComponent_LearnRecipe tests recipe learning.
func TestRecipeKnowledgeComponent_LearnRecipe(t *testing.T) {
	tests := []struct {
		name          string
		recipeSlots   int
		existingCount int
		recipeID      string
		shouldSucceed bool
		description   string
	}{
		{
			name:          "learn first recipe",
			recipeSlots:   5,
			existingCount: 0,
			recipeID:      "recipe1",
			shouldSucceed: true,
			description:   "should succeed with empty slots",
		},
		{
			name:          "learn duplicate",
			recipeSlots:   5,
			existingCount: 0,
			recipeID:      "recipe1",
			shouldSucceed: true,
			description:   "learning same recipe again should succeed",
		},
		{
			name:          "slots full",
			recipeSlots:   2,
			existingCount: 2,
			recipeID:      "recipe3",
			shouldSucceed: false,
			description:   "should fail when slots full",
		},
		{
			name:          "unlimited slots",
			recipeSlots:   0,
			existingCount: 100,
			recipeID:      "recipe101",
			shouldSucceed: true,
			description:   "0 slots means unlimited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewRecipeKnowledgeComponent(tt.recipeSlots)

			// Add existing recipes
			for i := 0; i < tt.existingCount; i++ {
				recipe := &Recipe{ID: "existing" + string(rune('A'+i))}
				comp.LearnRecipe(recipe)
			}

			// Try to learn new recipe
			recipe := &Recipe{ID: tt.recipeID, Name: "Test Recipe"}
			success := comp.LearnRecipe(recipe)

			if success != tt.shouldSucceed {
				t.Errorf("LearnRecipe() = %v, want %v (%s)", success, tt.shouldSucceed, tt.description)
			}

			// If successful, verify recipe is known
			if success && !comp.KnowsRecipe(tt.recipeID) {
				t.Error("Recipe should be known after successful learning")
			}
		})
	}
}

// TestRecipeKnowledgeComponent_KnowsRecipe tests recipe knowledge check.
func TestRecipeKnowledgeComponent_KnowsRecipe(t *testing.T) {
	comp := NewRecipeKnowledgeComponent(10)
	recipe := &Recipe{ID: "test_recipe"}

	// Should not know before learning
	if comp.KnowsRecipe("test_recipe") {
		t.Error("Should not know recipe before learning")
	}

	// Learn the recipe
	comp.LearnRecipe(recipe)

	// Should know after learning
	if !comp.KnowsRecipe("test_recipe") {
		t.Error("Should know recipe after learning")
	}

	// Should not know different recipe
	if comp.KnowsRecipe("other_recipe") {
		t.Error("Should not know unlearned recipe")
	}
}

// TestNewCraftingStationComponent tests crafting station creation.
func TestNewCraftingStationComponent(t *testing.T) {
	tests := []struct {
		name        string
		stationType RecipeType
	}{
		{"potion station", RecipePotion},
		{"enchanting station", RecipeEnchanting},
		{"magic item station", RecipeMagicItem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewCraftingStationComponent(tt.stationType)
			if comp == nil {
				t.Fatal("NewCraftingStationComponent returned nil")
			}
			if comp.StationType != tt.stationType {
				t.Errorf("StationType = %v, want %v", comp.StationType, tt.stationType)
			}
			if comp.BonusSuccessChance <= 0 {
				t.Error("BonusSuccessChance should be positive")
			}
			if comp.CraftTimeMultiplier <= 0 || comp.CraftTimeMultiplier >= 1 {
				t.Error("CraftTimeMultiplier should be between 0 and 1 for speed bonus")
			}
			if !comp.Available {
				t.Error("Station should be available by default")
			}
			if comp.Type() != "crafting_station" {
				t.Errorf("Type() = %v, want 'crafting_station'", comp.Type())
			}
		})
	}
}

// TestNewCraftingProgressComponent tests crafting progress creation.
func TestNewCraftingProgressComponent(t *testing.T) {
	recipe := &Recipe{
		ID:            "test",
		CraftTimeSec:  10.0,
		SkillRequired: 5,
	}
	stationID := uint64(42)

	comp := NewCraftingProgressComponent(recipe, 7.5, stationID)
	if comp == nil {
		t.Fatal("NewCraftingProgressComponent returned nil")
	}
	if comp.CurrentRecipe != recipe {
		t.Error("CurrentRecipe not set correctly")
	}
	if comp.RequiredTimeSec != 7.5 {
		t.Errorf("RequiredTimeSec = %v, want 7.5", comp.RequiredTimeSec)
	}
	if comp.ElapsedTimeSec != 0 {
		t.Errorf("ElapsedTimeSec should start at 0, got %v", comp.ElapsedTimeSec)
	}
	if comp.UsingStationID != stationID {
		t.Errorf("UsingStationID = %v, want %v", comp.UsingStationID, stationID)
	}
	if comp.MaterialsConsumed {
		t.Error("MaterialsConsumed should be false initially")
	}
	if comp.Type() != "crafting_progress" {
		t.Errorf("Type() = %v, want 'crafting_progress'", comp.Type())
	}
}

// TestCraftingProgressComponent_GetProgress tests progress calculation.
func TestCraftingProgressComponent_GetProgress(t *testing.T) {
	tests := []struct {
		name            string
		elapsed         float64
		required        float64
		expectedPercent float64
	}{
		{"0% progress", 0, 10, 0.0},
		{"25% progress", 2.5, 10, 0.25},
		{"50% progress", 5, 10, 0.50},
		{"75% progress", 7.5, 10, 0.75},
		{"100% progress", 10, 10, 1.0},
		{"over 100% caps at 1.0", 15, 10, 1.0},
		{"zero required time", 5, 0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &CraftingProgressComponent{
				ElapsedTimeSec:  tt.elapsed,
				RequiredTimeSec: tt.required,
			}
			got := comp.GetProgress()
			if got != tt.expectedPercent {
				t.Errorf("GetProgress() = %v, want %v", got, tt.expectedPercent)
			}
		})
	}
}

// TestCraftingProgressComponent_IsComplete tests completion check.
func TestCraftingProgressComponent_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		elapsed  float64
		required float64
		expected bool
	}{
		{"not complete", 5, 10, false},
		{"exactly complete", 10, 10, true},
		{"over complete", 15, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &CraftingProgressComponent{
				ElapsedTimeSec:  tt.elapsed,
				RequiredTimeSec: tt.required,
			}
			if got := comp.IsComplete(); got != tt.expected {
				t.Errorf("IsComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestNewCraftingSkillComponent tests crafting skill creation.
func TestNewCraftingSkillComponent(t *testing.T) {
	comp := NewCraftingSkillComponent()
	if comp == nil {
		t.Fatal("NewCraftingSkillComponent returned nil")
	}
	if comp.SkillLevel != 0 {
		t.Errorf("SkillLevel should start at 0, got %v", comp.SkillLevel)
	}
	if comp.Experience != 0 {
		t.Errorf("Experience should start at 0, got %v", comp.Experience)
	}
	if comp.ExperienceToNextLevel != 100 {
		t.Errorf("ExperienceToNextLevel = %v, want 100", comp.ExperienceToNextLevel)
	}
	if comp.Type() != "crafting_skill" {
		t.Errorf("Type() = %v, want 'crafting_skill'", comp.Type())
	}
}

// TestCraftingSkillComponent_AddExperience tests XP and leveling.
func TestCraftingSkillComponent_AddExperience(t *testing.T) {
	tests := []struct {
		name          string
		initialLevel  int
		initialXP     int
		addXP         int
		expectedLevel int
		expectedXP    int
		shouldLevelUp bool
		description   string
	}{
		{
			name:          "gain XP no level",
			initialLevel:  0,
			initialXP:     50,
			addXP:         30,
			expectedLevel: 0,
			expectedXP:    80,
			shouldLevelUp: false,
			description:   "XP gained but not enough to level",
		},
		{
			name:          "level up exactly",
			initialLevel:  0,
			initialXP:     90,
			addXP:         10,
			expectedLevel: 1,
			expectedXP:    0,
			shouldLevelUp: true,
			description:   "exactly 100 XP levels up",
		},
		{
			name:          "level up with overflow",
			initialLevel:  0,
			initialXP:     90,
			addXP:         30,
			expectedLevel: 1,
			expectedXP:    20,
			shouldLevelUp: true,
			description:   "excess XP carries over",
		},
		{
			name:          "level 5 scaling",
			initialLevel:  4,
			initialXP:     450,
			addXP:         50,
			expectedLevel: 5,
			expectedXP:    0,
			shouldLevelUp: true,
			description:   "XP requirement scales with level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &CraftingSkillComponent{
				SkillLevel:            tt.initialLevel,
				Experience:            tt.initialXP,
				ExperienceToNextLevel: 100 * (tt.initialLevel + 1),
			}

			leveledUp := comp.AddExperience(tt.addXP)

			if leveledUp != tt.shouldLevelUp {
				t.Errorf("AddExperience() returned %v, want %v (%s)", leveledUp, tt.shouldLevelUp, tt.description)
			}
			if comp.SkillLevel != tt.expectedLevel {
				t.Errorf("SkillLevel = %v, want %v", comp.SkillLevel, tt.expectedLevel)
			}
			if comp.Experience != tt.expectedXP {
				t.Errorf("Experience = %v, want %v", comp.Experience, tt.expectedXP)
			}
		})
	}
}

// TestMaterialRequirement tests material requirement structure.
func TestMaterialRequirement(t *testing.T) {
	weaponType := item.TypeWeapon
	req := MaterialRequirement{
		ItemName: "Iron Ore",
		Quantity: 5,
		Optional: false,
		ItemType: &weaponType,
	}

	if req.ItemName != "Iron Ore" {
		t.Errorf("ItemName = %v, want 'Iron Ore'", req.ItemName)
	}
	if req.Quantity != 5 {
		t.Errorf("Quantity = %v, want 5", req.Quantity)
	}
	if req.Optional {
		t.Error("Optional should be false")
	}
	if *req.ItemType != item.TypeWeapon {
		t.Errorf("ItemType = %v, want TypeWeapon", *req.ItemType)
	}
}

// TestRecipe tests full recipe structure.
func TestRecipe(t *testing.T) {
	recipe := &Recipe{
		ID:          "healing_potion_basic",
		Name:        "Basic Healing Potion",
		Description: "Restores 50 HP",
		Type:        RecipePotion,
		Rarity:      RecipeCommon,
		Materials: []MaterialRequirement{
			{ItemName: "Healing Herb", Quantity: 2, Optional: false},
			{ItemName: "Flask", Quantity: 1, Optional: false},
		},
		GoldCost:          10,
		SkillRequired:     0,
		BaseSuccessChance: 0.75,
		CraftTimeSec:      5.0,
		OutputItemSeed:    12345,
		OutputItemType:    item.TypeConsumable,
		GenreID:           "fantasy",
	}

	// Verify all fields set correctly
	if recipe.ID != "healing_potion_basic" {
		t.Errorf("ID = %v, want 'healing_potion_basic'", recipe.ID)
	}
	if recipe.Type != RecipePotion {
		t.Errorf("Type = %v, want RecipePotion", recipe.Type)
	}
	if recipe.Rarity != RecipeCommon {
		t.Errorf("Rarity = %v, want RecipeCommon", recipe.Rarity)
	}
	if len(recipe.Materials) != 2 {
		t.Errorf("Materials count = %v, want 2", len(recipe.Materials))
	}
	if recipe.Materials[0].Quantity != 2 {
		t.Errorf("First material quantity = %v, want 2", recipe.Materials[0].Quantity)
	}
}
