// Package vehicle provides helper functions for vehicle generation.
// This file contains utility functions for rarity determination, name generation,
// stat calculation (cargo slots, weight, upgrade slots), and validation helpers.
//
// All functions moved from generator.go during Phase 3 reorganization to improve
// code navigability by grouping related utility functions together.
//
// Code relocated from: generator.go
// Phase 21.1: Vehicle Foundation
package vehicle

import (
	"fmt"
	"math/rand"
)

// determineRarity calculates vehicle rarity based on depth progression and a random roll.
// At depth 0, distribution is ~70% Common, 20% Uncommon, 8% Rare, 2% Epic/Legendary.
// Each depth level adds +2% to rare+ chances via depthBonus.
// Originally from: generator.go
func (g *VehicleGenerator) determineRarity(depth int, rng *rand.Rand) Rarity {
	// Base probabilities
	roll := rng.Float64()

	// Depth increases rare chances
	depthBonus := float64(depth) * 0.02 // +2% per depth level

	// Common: 60% - depth*2%
	// Uncommon: 25%
	// Rare: 10% + depth*1%
	// Epic: 4% + depth*0.8%
	// Legendary: 1% + depth*0.2%

	commonChance := 0.60 - depthBonus
	uncommonChance := commonChance + 0.25
	rareChance := uncommonChance + 0.10 + (depthBonus * 0.5)
	epicChance := rareChance + 0.04 + (depthBonus * 0.4)
	// Legendary is everything above epicChance

	if roll < commonChance {
		return RarityCommon
	} else if roll < uncommonChance {
		return RarityUncommon
	} else if roll < rareChance {
		return RarityRare
	} else if roll < epicChance {
		return RarityEpic
	}
	return RarityLegendary
}

// generateName creates a vehicle name from template prefix/suffix combined with
// a rarity-based title. Legendary vehicles get a special unique title.
// Originally from: generator.go
func (g *VehicleGenerator) generateName(template VehicleTemplate, rarity Rarity, rng *rand.Rand) string {
	prefix := template.NamePrefix
	suffix := template.NameSuffix

	// Add rarity modifier to uncommon and above
	if rarity >= RarityUncommon {
		modifiers := []string{"Enhanced", "Superior", "Reinforced", "Advanced", "Masterwork"}
		prefix = modifiers[rng.Intn(len(modifiers))] + " " + prefix
	}

	return prefix + " " + suffix
}

// generateCargoSlots determines cargo capacity based on vehicle type and rarity.
// generateCargoSlots returns the number of cargo slots for a vehicle based on type and rarity.
// Mounts start with 2 (saddlebags), cargo vehicles scale up to 20+ for legendary tier.
// Phase 21.2: Cargo/Passenger System
func (g *VehicleGenerator) generateCargoSlots(vehicleType VehicleType, rarity Rarity, rng *rand.Rand) int {
	baseSlots := 0
	switch vehicleType {
	case TypeMount:
		baseSlots = 2 // Saddlebags
	case TypeCart:
		baseSlots = 20 // Large cargo area
	case TypeBoat:
		baseSlots = 10 // Medium storage
	case TypeGlider:
		baseSlots = 1 // Minimal storage
	case TypeMech:
		baseSlots = 5 // Internal compartment
	}

	// Scale with rarity (10-50 range)
	rarityBonus := int(float64(baseSlots) * (rarity.GetMultiplier() - 1.0) * 0.5)
	totalSlots := baseSlots + rarityBonus + rng.Intn(5) // +0 to 4 random

	// Clamp to valid range
	if totalSlots < 1 {
		totalSlots = 1
	}
	if totalSlots > 50 {
		totalSlots = 50
	}

	return totalSlots
}

// generateCargoWeight determines weight capacity based on vehicle type and rarity.
// generateCargoWeight returns the maximum cargo weight capacity in kg for a vehicle.
// Phase 21.2: Cargo/Passenger System
func (g *VehicleGenerator) generateCargoWeight(vehicleType VehicleType, rarity Rarity) float64 {
	baseWeight := 0.0
	switch vehicleType {
	case TypeMount:
		baseWeight = 50.0 // kg
	case TypeCart:
		baseWeight = 500.0 // kg
	case TypeBoat:
		baseWeight = 200.0 // kg
	case TypeGlider:
		baseWeight = 10.0 // kg
	case TypeMech:
		baseWeight = 100.0 // kg
	}

	return baseWeight * rarity.GetMultiplier()
}

// generateUpgradeSlots determines upgrade slot count based on rarity.
// Phase 21.2: Upgrade System
// Originally from: generator.go
func (g *VehicleGenerator) generateUpgradeSlots(rarity Rarity) int {
	switch rarity {
	case RarityCommon:
		return 2
	case RarityUncommon:
		return 3
	case RarityRare:
		return 4
	case RarityEpic:
		return 5
	case RarityLegendary:
		return 6
	default:
		return 2
	}
}

// validateVehicleBasics checks fundamental vehicle properties.
// Originally from: generator.go
func validateVehicleBasics(vehicle *Vehicle, index int) error {
	if vehicle == nil {
		return fmt.Errorf("vehicle %d is nil", index)
	}

	if vehicle.Name == "" {
		return fmt.Errorf("vehicle %d has empty name", index)
	}

	return nil
}

// validateVehicleStats verifies all vehicle stats are positive.
// Originally from: generator.go
func validateVehicleStats(vehicle *Vehicle, index int) error {
	if vehicle.MaxSpeed <= 0 {
		return fmt.Errorf("vehicle %d has invalid MaxSpeed: %f", index, vehicle.MaxSpeed)
	}
	if vehicle.Acceleration <= 0 {
		return fmt.Errorf("vehicle %d has invalid Acceleration: %f", index, vehicle.Acceleration)
	}
	if vehicle.Handling <= 0 {
		return fmt.Errorf("vehicle %d has invalid Handling: %f", index, vehicle.Handling)
	}
	if vehicle.MaxDurability <= 0 {
		return fmt.Errorf("vehicle %d has invalid MaxDurability: %f", index, vehicle.MaxDurability)
	}
	if vehicle.FuelCapacity <= 0 {
		return fmt.Errorf("vehicle %d has invalid FuelCapacity: %f", index, vehicle.FuelCapacity)
	}
	if vehicle.Capacity <= 0 {
		return fmt.Errorf("vehicle %d has invalid Capacity: %d", index, vehicle.Capacity)
	}

	return nil
}

// clamp constrains a value to a range.
// Originally from: generator.go
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
