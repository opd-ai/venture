// Package fluids - swimming.go
// This file contains the SwimmingManager implementation for swimming mechanics and stamina.
// Code relocated from: buoyancy.go

package fluids

import (
	"math"
)

// SwimmingManager handles swimming mechanics
type SwimmingManager struct {
	calculator *BuoyancyCalculator
}

// NewSwimmingManager creates a new swimming manager
func NewSwimmingManager(gravity float64) *SwimmingManager {
	return &SwimmingManager{
		calculator: NewBuoyancyCalculator(gravity),
	}
}

// UpdateSwimming updates swimming state based on fluid conditions
func (s *SwimmingManager) UpdateSwimming(
	swimming *SwimmingComponent,
	buoyancy *BuoyancyComponent,
	fluidAmount float64,
	deltaTime float64,
) {
	// Check if entity is in water
	inWater := fluidAmount > 0.3

	if !inWater {
		// On land - regenerate stamina
		swimming.IsSwimming = false
		swimming.TreadingWater = false
		swimming.Drowning = false

		swimming.Stamina = math.Min(swimming.MaxStamina, swimming.Stamina+swimming.StaminaRegen*deltaTime)
		return
	}

	// In water
	swimming.IsSwimming = true

	// Check if treading water (not moving)
	if swimming.TreadingWater {
		// Treading water drains less stamina
		swimming.Stamina = math.Max(0, swimming.Stamina-swimming.StaminaDrain*0.5*deltaTime)
	} else {
		// Swimming drains full stamina
		swimming.Stamina = math.Max(0, swimming.Stamina-swimming.StaminaDrain*deltaTime)
	}

	// Check for drowning
	if swimming.Stamina == 0 {
		swimming.Drowning = true
	} else {
		swimming.Drowning = false
	}
}

// GetSwimSpeedMultiplier returns the speed multiplier based on stamina
func (s *SwimmingManager) GetSwimSpeedMultiplier(swimming *SwimmingComponent) float64 {
	if !swimming.IsSwimming {
		return 1.0
	}

	// Speed scales with stamina
	staminaRatio := swimming.Stamina / swimming.MaxStamina
	return swimming.SwimSpeed * (0.5 + 0.5*staminaRatio)
}

// GetDrowningDamage returns drowning damage per second
func (s *SwimmingManager) GetDrowningDamage(swimming *SwimmingComponent) float64 {
	if swimming.Drowning {
		return swimming.DrowningDamage
	}
	return 0.0
}
