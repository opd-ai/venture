// Package engine provides vehicle durability system functionality.
// This file implements VehicleDurabilitySystem which handles vehicle damage
// from collisions, terrain, and combat.
//
// Phase 21.1: Vehicle Foundation
package engine

import (
	"github.com/sirupsen/logrus"
)

// VehicleDurabilitySystem manages vehicle damage and destruction.
// It processes damage from collisions, environmental hazards, and combat.
type VehicleDurabilitySystem struct {
	world  *World
	logger *logrus.Entry
}

// NewVehicleDurabilitySystem creates a new vehicle durability system.
func NewVehicleDurabilitySystem(world *World) *VehicleDurabilitySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "vehicle_durability")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("Vehicle durability system created")
		}
	}
	return &VehicleDurabilitySystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes vehicle durability for all entities with vehicle components.
func (vds *VehicleDurabilitySystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check if entity has vehicle component
		vehicleComp, hasVehicle := entity.GetComponent("vehicle")
		if !hasVehicle {
			continue
		}

		vehicle := vehicleComp.(*VehicleComponent)

		// Skip if already destroyed
		if vehicle.IsDestroyed() {
			vds.handleVehicleDestruction(entity, vehicle)
			continue
		}

		// Check for collision damage
		vds.checkCollisionDamage(entity, vehicle, deltaTime)

		// Check for environmental damage
		vds.checkEnvironmentalDamage(entity, vehicle, deltaTime)
	}
}

// checkCollisionDamage applies damage from collisions with walls/obstacles.
func (vds *VehicleDurabilitySystem) checkCollisionDamage(entity *Entity, vehicle *VehicleComponent, deltaTime float64) {
	// Check if entity has collider component
	collComp, hasCollision := entity.GetComponent("collider")
	if !hasCollision {
		return
	}

	collider := collComp.(*ColliderComponent)

	// In a full implementation, this would check if the collider
	// recently hit a solid object and apply damage based on speed.
	// For now, this is simplified - collision detection happens
	// in a separate collision system.
	_ = collider
}

// checkEnvironmentalDamage applies damage from terrain hazards.
func (vds *VehicleDurabilitySystem) checkEnvironmentalDamage(entity *Entity, vehicle *VehicleComponent, deltaTime float64) {
	// Check position to determine if on hazardous terrain
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return
	}

	// In a full implementation, this would query terrain at position
	// and apply damage for hazards like lava, spikes, etc.
	// For now, this is a placeholder for future enhancement
	_ = posComp
}

// handleVehicleDestruction manages what happens when a vehicle is destroyed.
func (vds *VehicleDurabilitySystem) handleVehicleDestruction(entity *Entity, vehicle *VehicleComponent) {
	// Ensure vehicle is stopped
	vehicle.Speed = 0.0

	// Dismount all passengers
	// In the full implementation, this would call the MountingSystem to eject riders
	vehicle.CurrentPassengers = 0

	// Mark entity for removal (in a full implementation)
	// For now, just log the destruction
	if vds.logger != nil {
		vds.logger.WithField("entity_id", entity.ID).Info("Vehicle destroyed")
	}
}

// ApplyDamage directly applies damage to a vehicle entity.
// This is called by external systems (combat, traps, etc.)
func (vds *VehicleDurabilitySystem) ApplyDamage(entity *Entity, damage float64) bool {
	vehicleComp, hasVehicle := entity.GetComponent("vehicle")
	if !hasVehicle {
		return false
	}

	vehicle := vehicleComp.(*VehicleComponent)
	destroyed := vehicle.TakeDamage(damage)

	if destroyed {
		vds.handleVehicleDestruction(entity, vehicle)
	}

	return destroyed
}

// RepairVehicle restores durability to a vehicle entity.
func (vds *VehicleDurabilitySystem) RepairVehicle(entity *Entity, amount float64) bool {
	vehicleComp, hasVehicle := entity.GetComponent("vehicle")
	if !hasVehicle {
		return false
	}

	vehicle := vehicleComp.(*VehicleComponent)
	vehicle.Repair(amount)
	return true
}
