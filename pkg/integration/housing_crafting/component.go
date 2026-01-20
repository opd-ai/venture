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

// Type returns the component type identifier.
// This is the only method allowed on components per ECS pattern.
// All logic has been moved to HousingCraftingSystem.
func (hcc *HousingCraftingComponent) Type() string {
	return "housing_crafting"
}
