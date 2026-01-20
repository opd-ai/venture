package companion_housing

import (
	"time"
)

// CompanionHousingComponent tracks a companion's housing integration state.
// Used by the ECS to manage companion-house interactions.
// This is a pure data component following ECS patterns - all logic
// is in CompanionHousingSystem.
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
