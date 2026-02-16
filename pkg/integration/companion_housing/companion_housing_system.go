package companion_housing

import (
	"time"
)

// CompanionHousingSystem provides operations on CompanionHousingComponent.
// Following ECS pattern, all logic that was previously in component methods
// is now in this system. Components remain pure data structures.
//
// Deprecated: Use PetHomeManager directly instead. Migration example:
//
//	// Old (deprecated):
//	system := NewCompanionHousingSystem(manager)
//	system.IsInHouse(component)
//
//	// New (recommended):
//	manager := NewPetHomeManager()
//	houseID := manager.GetCompanionHome(companionID)
//	isInHouse := houseID != ""
//
// PetHomeManager is injected into CompanionLoyaltySystem and other systems
// that need companion housing functionality. This struct is kept for backward
// compatibility with existing tests but may be removed in a future version.
type CompanionHousingSystem struct {
	manager *PetHomeManager
}

// NewCompanionHousingSystem creates a new companion housing system.
func NewCompanionHousingSystem(manager *PetHomeManager) *CompanionHousingSystem {
	return &CompanionHousingSystem{
		manager: manager,
	}
}

// IsInHouse returns true if companion is assigned to a house.
func (s *CompanionHousingSystem) IsInHouse(c *CompanionHousingComponent) bool {
	return c.OwnerHouseID != ""
}

// HasBedding returns true if companion has assigned bedding.
func (s *CompanionHousingSystem) HasBedding(c *CompanionHousingComponent) bool {
	return c.BeddingID != ""
}

// IsTraining returns true if companion has active training session.
func (s *CompanionHousingSystem) IsTraining(c *CompanionHousingComponent) bool {
	return c.ActiveTraining != ""
}

// HasSharedStorage returns true if companion can access shared chests.
func (s *CompanionHousingSystem) HasSharedStorage(c *CompanionHousingComponent) bool {
	return len(c.SharedChestAccess) > 0
}

// DaysSinceRest calculates days since last rest using the provided current time.
// This allows deterministic testing by injecting a fixed "now" time.
// Returns 0.0 if never rested (LastRestTime is zero).
func (s *CompanionHousingSystem) DaysSinceRest(c *CompanionHousingComponent, now time.Time) float64 {
	if c.LastRestTime.IsZero() {
		return 0.0
	}
	duration := now.Sub(c.LastRestTime)
	return duration.Hours() / 24.0
}

// UpdateFromManager syncs component state from PetHomeManager.
// Should be called during system update cycles.
func (s *CompanionHousingSystem) UpdateFromManager(c *CompanionHousingComponent, companionID uint64) {
	houseID := s.manager.GetCompanionHome(companionID)
	c.OwnerHouseID = houseID

	if houseID != "" {
		c.LoyaltyBonus = s.manager.GetLoyaltyBonus(companionID, houseID)
		c.TrainingBonus = s.manager.GetTrainingBonus(companionID, houseID)
	} else {
		c.LoyaltyBonus = 0.0
		c.TrainingBonus = 1.0
	}
}
