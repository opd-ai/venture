// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

import (
	"math"
)

// EnhancedVehicleSystem integrates suspension, weight transfer, terrain deformation, and collision response.
type EnhancedVehicleSystem struct {
	// System configuration
	enabled bool
}

// NewEnhancedVehicleSystem creates a new enhanced vehicle physics system.
func NewEnhancedVehicleSystem() *EnhancedVehicleSystem {
	return &EnhancedVehicleSystem{
		enabled: true,
	}
}

// VehicleState contains current vehicle dynamics for physics calculations.
type VehicleState struct {
	PositionX     float64
	PositionY     float64
	VelocityX     float64
	VelocityY     float64
	Rotation      float64 // Radians
	AngularVel    float64 // Radians/s
	Speed         float64 // Magnitude of velocity
	Acceleration  float64
	IsGrounded    bool
	TerrainHeight []float64 // Height at each wheel position
	TerrainTypes  []TerrainType
}

// UpdateVehiclePhysics performs a complete physics update for a vehicle.
// This integrates suspension, weight transfer, and collision response.
func (evs *EnhancedVehicleSystem) UpdateVehiclePhysics(
	suspension *SuspensionComponent,
	weightTransfer *WeightTransferComponent,
	deformation *TerrainDeformationComponent,
	collision *CollisionResponseComponent,
	state VehicleState,
	deltaTime float64,
) VehicleState {
	
	if !evs.enabled {
		return state
	}

	// Step 1: Update weight transfer based on vehicle dynamics
	if weightTransfer != nil {
		weightTransfer.Update(state.VelocityX, state.VelocityY, state.AngularVel, deltaTime)
		
		// Apply weight distribution to suspension
		if suspension != nil {
			weights := weightTransfer.GetWheelWeights()
			totalMass := suspension.TotalMass
			
			// Distribute mass to wheels based on weight transfer
			// Assuming 4-wheel layout: [FL, FR, RL, RR]
			for i := 0; i < len(suspension.Wheels) && i < 4; i++ {
				wheelMass := totalMass * weights[i]
				wheelLoad := wheelMass * 9.81 // F = mg
				suspension.SetWheelLoad(i, wheelLoad)
			}
		}
	}

	// Step 2: Update suspension physics
	if suspension != nil && len(state.TerrainHeight) == len(suspension.Wheels) {
		_ = suspension.Update(deltaTime, state.TerrainHeight) // Returns vertical offset for rendering
		
		// Check if vehicle is grounded (at least 2 wheels touching)
		groundedCount := suspension.GetGroundedWheelCount()
		state.IsGrounded = groundedCount >= 2
	}

	// Step 3: Add tire tracks if moving on soft terrain
	if deformation != nil && state.Speed > 10.0 { // Only create tracks above 10 pixels/s
		// Create tracks for each wheel touching ground
		if suspension != nil {
			for i := range suspension.Wheels {
				if suspension.IsWheelGrounded(i) {
					wheel := &suspension.Wheels[i]
					
					// Calculate world position of wheel
					wheelWorldX := state.PositionX + wheel.LocalX*math.Cos(state.Rotation) - wheel.LocalY*math.Sin(state.Rotation)
					wheelWorldY := state.PositionY + wheel.LocalX*math.Sin(state.Rotation) + wheel.LocalY*math.Cos(state.Rotation)
					
					// Get terrain type at wheel
					terrainType := TerrainFirm // Default
					if i < len(state.TerrainTypes) {
						terrainType = state.TerrainTypes[i]
					}
					
					// Add track with wheel load
					wheelLoad := suspension.GetWheelLoad(i)
					deformation.AddTrack(wheelWorldX, wheelWorldY, state.Rotation, wheelLoad, terrainType)
				}
			}
		}
		
		// Update existing tracks (aging and removal)
		deformation.Update(deltaTime)
	}

	// Step 4: Apply vertical offset from suspension to position
	// This makes the vehicle body move up/down based on suspension compression
	// (In a real implementation, this would be applied to rendering, not actual position)
	// For now, we just track it in the state
	
	// Step 5: Apply damage multiplier to vehicle performance
	if collision != nil && !collision.IsDestroyed() {
		damageMultiplier := collision.GetDamageMultiplier()
		
		// Reduce speed if damaged
		if damageMultiplier < 1.0 {
			// Scale down speed based on damage
			state.Speed *= damageMultiplier
			state.VelocityX *= damageMultiplier
			state.VelocityY *= damageMultiplier
		}
	}

	return state
}

// ProcessVehicleCollision handles a collision event for a vehicle.
// Returns updated velocity after collision response.
func (evs *EnhancedVehicleSystem) ProcessVehicleCollision(
	collision *CollisionResponseComponent,
	state VehicleState,
	normalX, normalY float64,
) (newVelX, newVelY, damage float64) {
	
	if collision == nil {
		return state.VelocityX, state.VelocityY, 0.0
	}

	// Process the collision
	result := collision.ProcessCollision(state.VelocityX, state.VelocityY, normalX, normalY)

	// Return new velocity and damage dealt
	return result.BounceVelocityX, result.BounceVelocityY, result.DamageDealt
}

// Enable enables the enhanced vehicle physics system.
func (evs *EnhancedVehicleSystem) Enable() {
	evs.enabled = true
}

// Disable disables the enhanced vehicle physics system.
func (evs *EnhancedVehicleSystem) Disable() {
	evs.enabled = false
}

// IsEnabled returns whether the system is enabled.
func (evs *EnhancedVehicleSystem) IsEnabled() bool {
	return evs.enabled
}
