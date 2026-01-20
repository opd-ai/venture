// Package fluids - flooding.go
// This file contains the FloodingManager implementation for flooding mechanics in enclosed areas.
// Code relocated from: buoyancy.go

package fluids

import (
	"math"
)

// FloodingManager handles flooding mechanics for enclosed areas
type FloodingManager struct {
	simulator *Simulator
}

// NewFloodingManager creates a new flooding manager
func NewFloodingManager(simulator *Simulator) *FloodingManager {
	return &FloodingManager{
		simulator: simulator,
	}
}

// UpdateFlooding updates flooding state for an area
func (f *FloodingManager) UpdateFlooding(component *FloodingComponent, deltaTime float64) {
	// Add water from all sources
	totalInflow := 0.0
	for _, source := range component.Sources {
		totalInflow += source.FlowRate * deltaTime

		// Add water to simulation grid
		f.simulator.AddFluid(source.X, source.Y, source.FlowRate*deltaTime, FluidWater)
	}

	// Update flood level
	component.FloodLevel = math.Min(component.MaxFloodLevel, component.FloodLevel+component.FloodRate*deltaTime)
}

// AddFloodSource adds a new water entry point
func (f *FloodingManager) AddFloodSource(component *FloodingComponent, x, y int, flowRate float64) {
	component.Sources = append(component.Sources, FloodSource{
		X:        x,
		Y:        y,
		FlowRate: flowRate,
	})
}

// RemoveFloodSource removes a water entry point
func (f *FloodingManager) RemoveFloodSource(component *FloodingComponent, x, y int) {
	for i, source := range component.Sources {
		if source.X == x && source.Y == y {
			component.Sources = append(component.Sources[:i], component.Sources[i+1:]...)
			return
		}
	}
}

// IsFullyFlooded returns true if the area is completely flooded
func (f *FloodingManager) IsFullyFlooded(component *FloodingComponent) bool {
	return component.FloodLevel >= component.MaxFloodLevel
}

// GetFloodPercentage returns the flood level as a percentage
func (f *FloodingManager) GetFloodPercentage(component *FloodingComponent) float64 {
	if component.MaxFloodLevel == 0 {
		return 0.0
	}
	return component.FloodLevel / component.MaxFloodLevel
}
