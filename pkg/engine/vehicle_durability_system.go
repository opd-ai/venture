// Package engine provides vehicle durability system functionality.
// This file implements VehicleDurabilitySystem which handles vehicle damage
// from collisions, terrain, and combat.
//
// Phase 21.1: Vehicle Foundation
// Phase 21.2: Terrain Hazard Integration
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// Terrain hazard damage rates (damage per second)
const (
	// LavaFlowDamagePerSecond is the damage rate for vehicles on lava tiles.
	LavaFlowDamagePerSecond = 10.0

	// DeepWaterDamagePerSecond is the damage rate for vehicles in deep water
	// (only applies to vehicles that cannot traverse water).
	DeepWaterDamagePerSecond = 5.0

	// PitDamagePerSecond is the damage rate for vehicles that fall into pits.
	PitDamagePerSecond = 15.0
)

// VehicleDurabilitySystem manages vehicle damage and destruction.
// It processes damage from collisions, environmental hazards, and combat.
type VehicleDurabilitySystem struct {
	world                   *World
	logger                  *logrus.Entry
	terrainCollisionChecker *TerrainCollisionChecker
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

// SetTerrainCollisionChecker sets the terrain collision checker for hazard detection.
// This enables the system to query terrain tiles and apply damage from hazardous terrain.
func (vds *VehicleDurabilitySystem) SetTerrainCollisionChecker(checker *TerrainCollisionChecker) {
	vds.terrainCollisionChecker = checker
}

// Update processes vehicle durability for all entities with vehicle components.
func (vds *VehicleDurabilitySystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check if entity has vehicle component
		vehicleComp, hasVehicle := entity.GetComponent("vehicle")
		if !hasVehicle {
			continue
		}

		// Type assert with safety check
		vehicle, ok := vehicleComp.(*VehicleComponent)
		if !ok {
			continue
		}

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

	// Type assert with safety check
	collider, ok := collComp.(*ColliderComponent)
	if !ok {
		return
	}

	// In a full implementation, this would check if the collider
	// recently hit a solid object and apply damage based on speed.
	// For now, this is simplified - collision detection happens
	// in a separate collision system.
	_ = collider
}

// checkEnvironmentalDamage applies damage from terrain hazards.
// Gap A6 Resolution: Now properly checks terrain hazards using TerrainCollisionChecker.
func (vds *VehicleDurabilitySystem) checkEnvironmentalDamage(entity *Entity, vehicle *VehicleComponent, deltaTime float64) {
	// Check position to determine if on hazardous terrain
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return
	}

	position, ok := posComp.(*PositionComponent)
	if !ok || position == nil {
		return
	}

	// Skip if no terrain collision checker is set
	if vds.terrainCollisionChecker == nil || vds.terrainCollisionChecker.terrain == nil {
		return
	}

	// Convert world position to tile coordinates
	tileX, tileY := vds.terrainCollisionChecker.worldToTileCoords(position.X, position.Y)

	// Get the tile type at the vehicle's position
	tile := vds.terrainCollisionChecker.terrain.GetTile(tileX, tileY)

	// Apply damage based on terrain hazard type
	damage := vds.calculateTerrainHazardDamage(tile, vehicle, deltaTime)
	if damage > 0 {
		destroyed := vehicle.TakeDamage(damage)
		if vds.logger != nil {
			vds.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"tile_type":  tile.String(),
				"damage":     damage,
				"durability": vehicle.Durability,
			}).Debug("Vehicle took terrain hazard damage")
		}
		if destroyed {
			vds.handleVehicleDestruction(entity, vehicle)
		}
	}
}

// calculateTerrainHazardDamage determines damage based on terrain type and vehicle capabilities.
// Returns the damage to apply this frame based on deltaTime.
func (vds *VehicleDurabilitySystem) calculateTerrainHazardDamage(tile terrain.TileType, vehicle *VehicleComponent, deltaTime float64) float64 {
	switch tile {
	case terrain.TileLavaFlow:
		// Lava damages all vehicles regardless of type
		return LavaFlowDamagePerSecond * deltaTime

	case terrain.TilePit:
		// Pits cause heavy damage - vehicles fall in
		// Gliders are immune to pit damage (they fly over)
		if vehicle.VehicleType == VehicleGlider {
			return 0
		}
		return PitDamagePerSecond * deltaTime

	case terrain.TileWaterDeep:
		// Deep water damages non-water vehicles
		// Boats and gliders are immune
		if vehicle.VehicleType == VehicleBoat || vehicle.VehicleType == VehicleGlider {
			return 0
		}
		return DeepWaterDamagePerSecond * deltaTime

	default:
		return 0
	}
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

	// Type assert with safety check
	vehicle, ok := vehicleComp.(*VehicleComponent)
	if !ok {
		return false
	}
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

	// Type assert with safety check
	vehicle, ok := vehicleComp.(*VehicleComponent)
	if !ok {
		return false
	}
	vehicle.Repair(amount)
	return true
}
