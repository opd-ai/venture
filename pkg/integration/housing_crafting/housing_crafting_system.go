package housing_crafting

// HousingCraftingSystem provides operations on HousingCraftingComponent.
// Following ECS pattern, all logic that was previously in component methods
// is now in this system. Components remain pure data structures.
//
// Deprecated: This system is a thin wrapper around StationManager and is not
// used in runtime code. Use StationManager directly instead, which is injected
// into CraftingSystem and other systems that need crafting station functionality.
// This struct is kept for backward compatibility with existing tests but may be
// removed in a future version.
type HousingCraftingSystem struct {
	manager *StationManager
}

// NewHousingCraftingSystem creates a new housing crafting system.
func NewHousingCraftingSystem(manager *StationManager) *HousingCraftingSystem {
	return &HousingCraftingSystem{
		manager: manager,
	}
}

// GetCraftingBonus returns the bonus multiplier for this station.
// Returns 1.0 (no bonus) if the multiplier is not set or is invalid.
func (s *HousingCraftingSystem) GetCraftingBonus(c *HousingCraftingComponent) float64 {
	if c == nil || c.BonusMultiplier <= 0 {
		return 1.0
	}
	return c.BonusMultiplier
}

// GetSkillBonus returns the skill XP bonus percentage for a skill.
// Returns 0 if the skill is not found or the bonus map is nil.
func (s *HousingCraftingSystem) GetSkillBonus(c *HousingCraftingComponent, skillName string) int {
	if c == nil || c.SkillBonus == nil {
		return 0
	}
	return c.SkillBonus[skillName]
}

// HasRecipe checks if a recipe is unlocked at this station.
// Returns false if the component is nil or the recipe is not found.
func (s *HousingCraftingSystem) HasRecipe(c *HousingCraftingComponent, recipeID string) bool {
	if c == nil {
		return false
	}
	for _, id := range c.ActiveRecipes {
		if id == recipeID {
			return true
		}
	}
	return false
}

// SyncFromStation updates the component state from a registered CraftingStation.
// This should be called during system update cycles to keep the component in sync.
func (s *HousingCraftingSystem) SyncFromStation(c *HousingCraftingComponent) error {
	if c == nil {
		return nil
	}
	if c.StationID == "" {
		return nil
	}

	station, err := s.manager.GetStation(c.StationID)
	if err != nil {
		return err
	}

	c.StationType = station.Type
	c.BonusMultiplier = station.Quality.Multiplier()

	// Deep copy map to avoid sharing mutable references with the station
	if station.SkillBonus != nil {
		c.SkillBonus = make(map[string]int, len(station.SkillBonus))
		for k, v := range station.SkillBonus {
			c.SkillBonus[k] = v
		}
	} else {
		c.SkillBonus = nil
	}

	// Deep copy slice to avoid sharing mutable references with the station
	if station.ActiveRecipes != nil {
		c.ActiveRecipes = make([]string, len(station.ActiveRecipes))
		copy(c.ActiveRecipes, station.ActiveRecipes)
	} else {
		c.ActiveRecipes = nil
	}

	return nil
}
