// Package fluids - buoyancy_calculator.go
// This file contains the BuoyancyCalculator implementation for buoyancy force calculations.
// Code relocated from: buoyancy.go

package fluids

import (
	"math"
)

// BuoyancyCalculator handles buoyancy calculations for entities
type BuoyancyCalculator struct {
	gravity float64
}

// NewBuoyancyCalculator creates a new buoyancy calculator
func NewBuoyancyCalculator(gravity float64) *BuoyancyCalculator {
	return &BuoyancyCalculator{
		gravity: gravity,
	}
}

// CalculateBuoyancy calculates buoyancy for an entity in fluid
func (b *BuoyancyCalculator) CalculateBuoyancy(component *BuoyancyComponent, fluidAmount float64, fluidType FluidType) {
	if fluidAmount == 0.0 {
		component.Submerged = 0.0
		component.BuoyantForce = 0.0
		component.Buoyant = false
		return
	}

	// Get fluid properties
	props := GetFluidProperties(fluidType)

	// Calculate submersion percentage
	component.Submerged = math.Min(1.0, fluidAmount)

	// Calculate buoyant force: F = ρ * V * g * submersion
	fluidMass := props.Density * component.Volume * component.Submerged
	component.BuoyantForce = fluidMass * b.gravity

	// Check if entity floats
	entityWeight := component.Mass * b.gravity
	component.Buoyant = component.BuoyantForce >= entityWeight
}

// GetNetForce returns the net vertical force (buoyancy - weight)
func (b *BuoyancyCalculator) GetNetForce(component *BuoyancyComponent) float64 {
	weight := component.Mass * b.gravity
	return component.BuoyantForce - weight
}

// UpdateDensity recalculates entity density from mass and volume
func (b *BuoyancyCalculator) UpdateDensity(component *BuoyancyComponent) {
	if component.Volume > 0 {
		component.Density = component.Mass / component.Volume
	}
}

// UpdateDensity recalculates entity density from mass and volume.
// This is a package-level convenience function that doesn't require a BuoyancyCalculator instance.
// For consistency with other methods, prefer using BuoyancyCalculator.UpdateDensity() when
// a calculator instance is already available.
func UpdateDensity(component *BuoyancyComponent) {
	if component.Volume > 0 {
		component.Density = component.Mass / component.Volume
	}
}
