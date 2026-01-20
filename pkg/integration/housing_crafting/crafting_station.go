package housing_crafting

// CraftingStation represents a crafting station placed in a player's house.
// This file contains the CraftingStation type and its methods for querying
// skill bonuses and recipe availability.
//
// Code relocated from: types.go

// CraftingStation represents a crafting station placed in a player's house
type CraftingStation struct {
	ID          string      // Unique station ID
	Type        StationType // Station type
	Quality     QualityTier // Quality tier
	OwnerID     string      // Player ID who owns the house
	HouseID     string      // House ID where station is located
	FurnitureID uint64      // Entity ID of the furniture representing this station

	// Skill bonuses provided by this station
	SkillBonus map[string]int // Skill name → XP bonus percentage (e.g., "smithing" → 50 for +50%)

	// Active recipes unlocked by this station
	ActiveRecipes []string // Recipe IDs available at this quality tier
}

// GetSkillTrainingBonus returns the XP bonus percentage for a skill
func (cs *CraftingStation) GetSkillTrainingBonus(skillName string) int {
	if cs.SkillBonus == nil {
		return 0
	}
	return cs.SkillBonus[skillName]
}

// HasRecipe checks if the station has unlocked a specific recipe
func (cs *CraftingStation) HasRecipe(recipeID string) bool {
	for _, id := range cs.ActiveRecipes {
		if id == recipeID {
			return true
		}
	}
	return false
}
