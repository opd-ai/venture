package companion_housing

import (
	"time"
)

// CompanionHousingComponent tracks a companion's housing integration state.
// Used by the ECS to manage companion-house interactions.
type CompanionHousingComponent struct {
	OwnerHouseID      string    // House where companion is assigned
	BeddingID         string    // Furniture ID of assigned bedding
	LastRestTime      time.Time // Last time companion rested
	LoyaltyBonus      float64   // Daily loyalty gain from housing
	ActiveTraining    string    // Furniture ID of active training area (empty if none)
	TrainingBonus     float64   // XP multiplier from active training
	SharedChestAccess []string  // Furniture IDs of accessible shared chests
}

// Type returns the component type identifier for ECS.
func (c *CompanionHousingComponent) Type() string {
	return "companion_housing"
}

// IsInHouse returns true if companion is assigned to a house.
func (c *CompanionHousingComponent) IsInHouse() bool {
	return c.OwnerHouseID != ""
}

// HasBedding returns true if companion has assigned bedding.
func (c *CompanionHousingComponent) HasBedding() bool {
	return c.BeddingID != ""
}

// IsTraining returns true if companion has active training session.
func (c *CompanionHousingComponent) IsTraining() bool {
	return c.ActiveTraining != ""
}

// HasSharedStorage returns true if companion can access shared chests.
func (c *CompanionHousingComponent) HasSharedStorage() bool {
	return len(c.SharedChestAccess) > 0
}

// DaysSinceRest calculates days since last rest (assumes 24-hour days).
// Returns 0.0 if never rested.
func (c *CompanionHousingComponent) DaysSinceRest() float64 {
	if c.LastRestTime.IsZero() {
		return 0.0
	}
	duration := time.Since(c.LastRestTime)
	return duration.Hours() / 24.0
}

// UpdateFromManager syncs component state from PetHomeManager.
// Should be called by systems that modify housing state.
func (c *CompanionHousingComponent) UpdateFromManager(manager *PetHomeManager, companionID uint64) {
	houseID := manager.GetCompanionHome(companionID)
	c.OwnerHouseID = houseID

	if houseID != "" {
		c.LoyaltyBonus = manager.GetLoyaltyBonus(companionID, houseID)
		c.TrainingBonus = manager.GetTrainingBonus(companionID, houseID)
	} else {
		c.LoyaltyBonus = 0.0
		c.TrainingBonus = 1.0
	}
}
