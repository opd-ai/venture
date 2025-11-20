// Package engine provides vehicle movement system functionality.
// This file implements VehicleMovementSystem which handles vehicle physics
// including momentum, turning radius, terrain interaction, and fuel consumption.
//
// Phase 21.1: Vehicle Foundation
package engine

import (
	"math"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// VehicleMovementSystem manages vehicle physics and movement.
// It processes entities with both VehicleComponent and PositionComponent,
// applying momentum, turning, fuel consumption, and terrain interaction.
type VehicleMovementSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewVehicleMovementSystem creates a new vehicle movement system.
func NewVehicleMovementSystem(world *World) *VehicleMovementSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "vehicle_movement")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("Vehicle movement system created")
		}
	}
	return &VehicleMovementSystem{
		world:  world,
		logger: logEntry,
	}
}

// getVehicleComponents retrieves and validates vehicle, position and rotation components.
func (vms *VehicleMovementSystem) getVehicleComponents(entity *Entity) (*VehicleComponent, *PositionComponent, *RotationComponent, bool) {
	vehicleComp, hasVehicle := entity.GetComponent("vehicle")
	if !hasVehicle {
		return nil, nil, nil, false
	}

	vehicle, ok := vehicleComp.(*VehicleComponent)
	if !ok {
		return nil, nil, nil, false
	}

	if vehicle.IsDestroyed() || vehicle.IsFuelDepleted() {
		vehicle.Speed = 0.0
		return nil, nil, nil, false
	}

	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return nil, nil, nil, false
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, nil, nil, false
	}

	rotComp, hasRot := entity.GetComponent("rotation")
	if !hasRot {
		return nil, nil, nil, false
	}
	rot, ok := rotComp.(*RotationComponent)
	if !ok {
		return nil, nil, nil, false
	}

	return vehicle, pos, rot, true
}

// handleVehicleControl processes vehicle control input or deceleration.
func (vms *VehicleMovementSystem) handleVehicleControl(entity *Entity, vehicle *VehicleComponent, rot *RotationComponent, deltaTime float64) {
	hasControl := vms.hasControlInput(entity)

	if hasControl {
		inputComp, hasInput := entity.GetComponent("input")
		if hasInput {
			input, ok := inputComp.(InputProvider)
			if ok {
				vms.processInput(vehicle, rot, input, deltaTime)
			}
		}
	} else {
		vms.applyDeceleration(vehicle, deltaTime)
	}
}

// consumeVehicleFuel consumes fuel and stops vehicle if depleted.
func (vms *VehicleMovementSystem) consumeVehicleFuel(entity *Entity, vehicle *VehicleComponent, deltaTime float64) {
	fuelCost := vehicle.GetFuelCost() * deltaTime
	if !vehicle.ConsumeFuel(fuelCost) {
		vehicle.Speed = 0.0
		if vms.logger != nil {
			vms.logger.WithField("entity_id", entity.ID).Debug("Vehicle out of fuel")
		}
	}
}

// Update processes vehicle movement for all entities with vehicle components.
func (vms *VehicleMovementSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		vehicle, pos, rot, ok := vms.getVehicleComponents(entity)
		if !ok {
			continue
		}

		vms.handleVehicleControl(entity, vehicle, rot, deltaTime)

		if vehicle.Speed > 0 {
			vms.applyMovement(vehicle, pos, rot, deltaTime)
			vms.consumeVehicleFuel(entity, vehicle, deltaTime)
		}
	}
}

// hasControlInput checks if the vehicle has active control.
// Returns true if entity has input component or is mounted by a player.
func (vms *VehicleMovementSystem) hasControlInput(entity *Entity) bool {
	// Check for direct input component (player-controlled vehicle)
	if _, hasInput := entity.GetComponent("input"); hasInput {
		return true
	}

	// Check if any entity is mounted on this vehicle
	vehicleComp, ok := entity.GetComponent("vehicle")
	if !ok {
		return false
	}
	// Type assert with safety check
	vehicle, ok := vehicleComp.(*VehicleComponent)
	if !ok {
		return false
	}
	return vehicle.CurrentPassengers > 0
}

// processInput handles acceleration and turning based on input.
func (vms *VehicleMovementSystem) processInput(vehicle *VehicleComponent, rot *RotationComponent, input InputProvider, deltaTime float64) {
	// Get movement input
	moveX, moveY := input.GetMovement()

	// Calculate desired direction from input
	if moveX != 0 || moveY != 0 {
		// Normalize input
		inputMag := math.Sqrt(moveX*moveX + moveY*moveY)
		if inputMag > 1.0 {
			moveX /= inputMag
			moveY /= inputMag
		}

		// Calculate target angle from input
		targetAngle := math.Atan2(moveY, moveX)

		// Apply turning with handling limits
		currentAngle := rot.Angle
		angleDiff := shortestAngularDistance(currentAngle, targetAngle)

		// Limit turn rate based on handling
		maxTurn := vehicle.Handling * deltaTime
		if math.Abs(angleDiff) > maxTurn {
			if angleDiff > 0 {
				rot.Angle = normalizeAngle(currentAngle + maxTurn)
			} else {
				rot.Angle = normalizeAngle(currentAngle - maxTurn)
			}
		} else {
			rot.Angle = normalizeAngle(targetAngle)
		}

		// Apply acceleration
		vehicle.Speed += vehicle.Acceleration * deltaTime
		if vehicle.Speed > vehicle.MaxSpeed {
			vehicle.Speed = vehicle.MaxSpeed
		}
	} else {
		// No input - decelerate
		vms.applyDeceleration(vehicle, deltaTime)
	}
}

// applyDeceleration slows down the vehicle over time.
func (vms *VehicleMovementSystem) applyDeceleration(vehicle *VehicleComponent, deltaTime float64) {
	// Deceleration is half of acceleration for smooth feel
	decel := vehicle.Acceleration * 0.5 * deltaTime
	vehicle.Speed -= decel
	if vehicle.Speed < 0 {
		vehicle.Speed = 0
	}
}

// applyMovement updates position based on current speed and rotation.
func (vms *VehicleMovementSystem) applyMovement(vehicle *VehicleComponent, pos *PositionComponent, rot *RotationComponent, deltaTime float64) {
	if vehicle.Speed <= 0 {
		return
	}

	// Calculate movement vector
	moveDistance := vehicle.Speed * deltaTime
	dx := math.Cos(rot.Angle) * moveDistance
	dy := math.Sin(rot.Angle) * moveDistance

	// Update position
	newX := pos.X + dx
	newY := pos.Y + dy

	// Check terrain at new position (simplified - assumes we can query terrain)
	// In a real implementation, this would query the world's terrain system
	canMove := vms.canMoveToPosition(vehicle, newX, newY)

	if canMove {
		pos.X = newX
		pos.Y = newY
	} else {
		// Hit obstacle - stop
		vehicle.Speed = 0.0
		if vms.logger != nil {
			vms.logger.Debug("Vehicle blocked by terrain")
		}
	}
}

// canMoveToPosition checks if vehicle can move to the given position.
// This is a simplified version - a full implementation would query the
// terrain system to get the actual tile type at the position.
func (vms *VehicleMovementSystem) canMoveToPosition(vehicle *VehicleComponent, x, y float64) bool {
	// For now, assume floor terrain (this would be enhanced with actual terrain queries)
	// In the full implementation, this would call world.GetTerrainAt(x, y)
	terrainType := terrain.TileFloor
	return vehicle.CanTraverse(int(terrainType))
}
