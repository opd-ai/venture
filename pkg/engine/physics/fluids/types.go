// Package fluids - types.go
// This file contains all type definitions, constants, and helper functions for fluid simulation.

package fluids

import (
	"image/color"
	"sync"
)

// FluidType represents the type of fluid
type FluidType int

const (
	FluidWater FluidType = iota
	FluidLava
	FluidOil
	FluidAcid
	FluidPoison
)

// String returns the human-readable name of the fluid type
func (f FluidType) String() string {
	switch f {
	case FluidWater:
		return "Water"
	case FluidLava:
		return "Lava"
	case FluidOil:
		return "Oil"
	case FluidAcid:
		return "Acid"
	case FluidPoison:
		return "Poison"
	default:
		return "Unknown"
	}
}

// FluidProperties defines the physical properties of a fluid type
type FluidProperties struct {
	Viscosity    float64    // Resistance to flow (0.0-1.0)
	Density      float64    // Mass per unit volume (kg/m³)
	FlowRate     float64    // Speed of flow propagation (0.0-1.0)
	Damage       float64    // Damage per second when submerged
	Color        color.RGBA // Visual color of the fluid
	Transparency float64    // Visual transparency (0.0=opaque, 1.0=transparent)
}

// Cell represents a single grid cell in the fluid simulation
type Cell struct {
	Pressure  float64   // Pressure value (0.0-1.0)
	VelocityX float64   // X-axis velocity
	VelocityY float64   // Y-axis velocity
	Type      FluidType // Type of fluid in this cell
	Amount    float64   // Amount of fluid (0.0=empty, 1.0=full)
}

// Grid represents the fluid simulation grid
type Grid struct {
	Width  int      // Grid width in cells
	Height int      // Grid height in cells
	Cells  [][]Cell // 2D array of cells
	mu     sync.RWMutex
}

// BuoyancyComponent defines buoyancy properties for an entity
type BuoyancyComponent struct {
	Mass         float64 // Entity mass (kg)
	Volume       float64 // Entity volume (m³)
	Density      float64 // Entity density (kg/m³) = Mass/Volume
	Buoyant      bool    // Whether entity floats
	Submerged    float64 // Percentage submerged (0.0-1.0)
	BuoyantForce float64 // Upward force from buoyancy (N)
}

// Type returns the component type identifier
func (b BuoyancyComponent) Type() string {
	return "buoyancy"
}

// SwimmingComponent defines swimming mechanics for an entity
type SwimmingComponent struct {
	IsSwimming     bool    // Currently in water and swimming
	Stamina        float64 // Swimming stamina (0.0-100.0)
	MaxStamina     float64 // Maximum stamina
	StaminaDrain   float64 // Stamina drain per second
	StaminaRegen   float64 // Stamina regen per second (on land)
	SwimSpeed      float64 // Swimming speed multiplier (0.0-1.0)
	TreadingWater  bool    // Staying afloat without moving
	Drowning       bool    // Out of stamina, taking damage
	DrowningDamage float64 // Damage per second when drowning
}

// Type returns the component type identifier
func (s SwimmingComponent) Type() string {
	return "swimming"
}

// FloodingComponent tracks flooding state for an enclosed area
type FloodingComponent struct {
	AreaID        string        // Identifier for the area being flooded
	FloodLevel    float64       // Current flood level (0.0-1.0)
	FloodRate     float64       // Rate of flooding (units per second)
	MaxFloodLevel float64       // Maximum flood level before area is full
	Sources       []FloodSource // Water entry points
}

// Type returns the component type identifier
func (f FloodingComponent) Type() string {
	return "flooding"
}

// FloodSource represents a water entry point
type FloodSource struct {
	X        int     // Grid X position
	Y        int     // Grid Y position
	FlowRate float64 // Amount of water per second
}

// SimulationConfig defines configuration for fluid simulation
type SimulationConfig struct {
	GridWidth       int     // Width of simulation grid
	GridHeight      int     // Height of simulation grid
	CellSize        float64 // Size of each cell in world units
	UpdateRate      float64 // Updates per second (default: 30 FPS)
	Gravity         float64 // Gravity constant (m/s²)
	PressureFactor  float64 // Pressure calculation factor
	ViscosityFactor float64 // Viscosity calculation factor
	MaxIterations   int     // Max iterations per update
	Convergence     float64 // Convergence threshold
}

// DefaultSimulationConfig returns default configuration values
func DefaultSimulationConfig() SimulationConfig {
	return SimulationConfig{
		GridWidth:       100,
		GridHeight:      100,
		CellSize:        1.0,
		UpdateRate:      30.0,
		Gravity:         9.81,
		PressureFactor:  1.0,
		ViscosityFactor: 0.95,
		MaxIterations:   10,
		Convergence:     0.001,
	}
}

// GetFluidProperties returns the properties for a given fluid type
func GetFluidProperties(fluidType FluidType) FluidProperties {
	switch fluidType {
	case FluidWater:
		return FluidProperties{
			Viscosity:    0.1,
			Density:      1000.0,
			FlowRate:     0.8,
			Damage:       0.0,
			Color:        color.RGBA{R: 30, G: 144, B: 255, A: 180},
			Transparency: 0.7,
		}
	case FluidLava:
		return FluidProperties{
			Viscosity:    0.6,
			Density:      3000.0,
			FlowRate:     0.3,
			Damage:       50.0,
			Color:        color.RGBA{R: 255, G: 69, B: 0, A: 220},
			Transparency: 0.2,
		}
	case FluidOil:
		return FluidProperties{
			Viscosity:    0.3,
			Density:      900.0,
			FlowRate:     0.6,
			Damage:       0.0,
			Color:        color.RGBA{R: 50, G: 50, B: 50, A: 150},
			Transparency: 0.5,
		}
	case FluidAcid:
		return FluidProperties{
			Viscosity:    0.15,
			Density:      1200.0,
			FlowRate:     0.7,
			Damage:       25.0,
			Color:        color.RGBA{R: 0, G: 255, B: 0, A: 200},
			Transparency: 0.6,
		}
	case FluidPoison:
		return FluidProperties{
			Viscosity:    0.12,
			Density:      1050.0,
			FlowRate:     0.75,
			Damage:       15.0,
			Color:        color.RGBA{R: 128, G: 0, B: 128, A: 190},
			Transparency: 0.65,
		}
	default:
		return FluidProperties{
			Viscosity:    0.1,
			Density:      1000.0,
			FlowRate:     0.5,
			Damage:       0.0,
			Color:        color.RGBA{R: 255, G: 255, B: 255, A: 255},
			Transparency: 1.0,
		}
	}
}
