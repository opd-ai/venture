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
func UpdateDensity(component *BuoyancyComponent) {
	if component.Volume > 0 {
		component.Density = component.Mass / component.Volume
	}
}

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
