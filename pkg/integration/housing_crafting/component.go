package housing_crafting

// HousingCraftingComponent integrates crafting stations with the ECS.
// This component is attached to furniture entities that represent crafting stations.
type HousingCraftingComponent struct {
	StationID       string         // Links to StationManager
	StationType     StationType    // Type of crafting station
	BonusMultiplier float64        // Crafting bonus (1.0-2.0)
	SkillBonus      map[string]int // Skill XP bonuses
	ActiveRecipes   []string       // Unlocked recipes
}

// Type returns the component type identifier
func (hcc *HousingCraftingComponent) Type() string {
	return "housing_crafting"
}

// GetCraftingBonus returns the bonus multiplier for this station
func (hcc *HousingCraftingComponent) GetCraftingBonus() float64 {
	if hcc.BonusMultiplier <= 0 {
		return 1.0
	}
	return hcc.BonusMultiplier
}

// GetSkillBonus returns the skill XP bonus percentage for a skill
func (hcc *HousingCraftingComponent) GetSkillBonus(skillName string) int {
	if hcc.SkillBonus == nil {
		return 0
	}
	return hcc.SkillBonus[skillName]
}

// HasRecipe checks if a recipe is unlocked at this station
func (hcc *HousingCraftingComponent) HasRecipe(recipeID string) bool {
	for _, id := range hcc.ActiveRecipes {
		if id == recipeID {
			return true
		}
	}
	return false
}
