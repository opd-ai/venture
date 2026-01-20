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

	evs.updateWeightTransfer(suspension, weightTransfer, state, deltaTime)
	state = evs.updateSuspension(suspension, weightTransfer, state, deltaTime)
	evs.updateTireTracks(suspension, deformation, state, deltaTime)
	state = evs.applyCollisionDamage(collision, state)

	return state
}

// updateWeightTransfer updates weight transfer and applies it to suspension.
func (evs *EnhancedVehicleSystem) updateWeightTransfer(
	suspension *SuspensionComponent,
	weightTransfer *WeightTransferComponent,
	state VehicleState,
	deltaTime float64,
) {
	if weightTransfer == nil {
		return
	}

	weightTransfer.Update(state.VelocityX, state.VelocityY, state.AngularVel, deltaTime)

	if suspension != nil {
		evs.applyWeightToSuspension(suspension, weightTransfer)
	}
}

// applyWeightToSuspension distributes mass to wheels based on weight transfer.
func (evs *EnhancedVehicleSystem) applyWeightToSuspension(suspension *SuspensionComponent, weightTransfer *WeightTransferComponent) {
	weights := weightTransfer.GetWheelWeights()
	totalMass := suspension.TotalMass

	for i := 0; i < len(suspension.Wheels) && i < 4; i++ {
		wheelMass := totalMass * weights[i]
		wheelLoad := wheelMass * 9.81
		suspension.SetWheelLoad(i, wheelLoad)
	}
}

// updateSuspension updates suspension physics and grounding state.
func (evs *EnhancedVehicleSystem) updateSuspension(
	suspension *SuspensionComponent,
	weightTransfer *WeightTransferComponent,
	state VehicleState,
	deltaTime float64,
) VehicleState {
	if suspension != nil && len(state.TerrainHeight) == len(suspension.Wheels) {
		_ = suspension.Update(deltaTime, state.TerrainHeight)
		groundedCount := suspension.GetGroundedWheelCount()
		state.IsGrounded = groundedCount >= 2
	}
	return state
}

// updateTireTracks adds tire tracks for moving vehicles on soft terrain.
func (evs *EnhancedVehicleSystem) updateTireTracks(
	suspension *SuspensionComponent,
	deformation *TerrainDeformationComponent,
	state VehicleState,
	deltaTime float64,
) {
	if deformation == nil || state.Speed <= 10.0 {
		return
	}

	if suspension != nil {
		for i := range suspension.Wheels {
			if suspension.IsWheelGrounded(i) {
				evs.addWheelTrack(suspension, deformation, state, i)
			}
		}
	}

	deformation.Update(deltaTime)
}

// addWheelTrack adds a single wheel track to terrain deformation.
func (evs *EnhancedVehicleSystem) addWheelTrack(
	suspension *SuspensionComponent,
	deformation *TerrainDeformationComponent,
	state VehicleState,
	wheelIndex int,
) {
	wheel := &suspension.Wheels[wheelIndex]

	wheelWorldX := state.PositionX + wheel.LocalX*math.Cos(state.Rotation) - wheel.LocalY*math.Sin(state.Rotation)
	wheelWorldY := state.PositionY + wheel.LocalX*math.Sin(state.Rotation) + wheel.LocalY*math.Cos(state.Rotation)

	terrainType := TerrainFirm
	if wheelIndex < len(state.TerrainTypes) {
		terrainType = state.TerrainTypes[wheelIndex]
	}

	wheelLoad := suspension.GetWheelLoad(wheelIndex)
	deformation.AddTrack(wheelWorldX, wheelWorldY, state.Rotation, wheelLoad, terrainType)
}

// applyCollisionDamage applies damage multiplier to vehicle performance.
func (evs *EnhancedVehicleSystem) applyCollisionDamage(collision *CollisionResponseComponent, state VehicleState) VehicleState {
	if collision != nil && !collision.IsDestroyed() {
		damageMultiplier := collision.GetDamageMultiplier()

		if damageMultiplier < 1.0 {
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
