// Package engine provides mounting system functionality.
// This file implements MountingSystem which handles mounting and dismounting
// entities on vehicles, position synchronization, and mount state management.
//
// Phase 21.1: Vehicle Foundation
package engine

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// MountingSystem manages rider-vehicle relationships.
// It handles mounting, dismounting, and position synchronization.
type MountingSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewMountingSystem creates a new mounting system.
func NewMountingSystem(world *World) *MountingSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "mounting")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("Mounting system created")
		}
	}
	return &MountingSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update synchronizes rider positions with their mounted vehicles.
func (ms *MountingSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check if entity is mounted on a vehicle
		mountComp, isMounted := entity.GetComponent("mount")
		if !isMounted {
			continue
		}

		mount, ok := mountComp.(*MountComponent)
		if !ok {
			continue
		}

		// Find the vehicle entity
		vehicle := ms.findEntity(entities, mount.MountedEntityID)
		if vehicle == nil {
			// Vehicle no longer exists - dismount
			ms.forceDismount(entity)
			continue
		}

		// Get vehicle position
		vehiclePosComp, hasVehiclePos := vehicle.GetComponent("position")
		if !hasVehiclePos {
			continue
		}
		vehiclePos, ok := vehiclePosComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Update rider position to match vehicle + offset
		riderPosComp, hasRiderPos := entity.GetComponent("position")
		if hasRiderPos {
			if riderPos, ok := riderPosComp.(*PositionComponent); ok {
				riderPos.X = vehiclePos.X + mount.MountOffset.X
				riderPos.Y = vehiclePos.Y + mount.MountOffset.Y
			}
		}

		// Synchronize rider rotation with vehicle
		vehicleRotComp, hasVehicleRot := vehicle.GetComponent("rotation")
		riderRotComp, hasRiderRot := entity.GetComponent("rotation")
		if hasVehicleRot && hasRiderRot {
			if vehicleRot, ok := vehicleRotComp.(*RotationComponent); ok {
				if riderRot, ok := riderRotComp.(*RotationComponent); ok {
					riderRot.Angle = vehicleRot.Angle
				}
			}
		}
	}
}

// Mount attempts to mount a rider entity onto a vehicle entity.
// Implements VehicleController interface.
func (ms *MountingSystem) Mount(rider, vehicle *Entity) error {
	if rider == nil {
		return fmt.Errorf("rider entity is nil")
	}
	if vehicle == nil {
		return fmt.Errorf("vehicle entity is nil")
	}

	// Check if rider is already mounted
	if _, isMounted := rider.GetComponent("mount"); isMounted {
		return fmt.Errorf("rider is already mounted")
	}

	// Check if vehicle has vehicle component
	vehicleComp, hasVehicle := vehicle.GetComponent("vehicle")
	if !hasVehicle {
		return fmt.Errorf("entity is not a vehicle")
	}

	vehicleData, ok := vehicleComp.(*VehicleComponent)
	if !ok {
		return fmt.Errorf("invalid vehicle component type")
	}

	// Check if vehicle has capacity
	if !vehicleData.CanAddPassenger() {
		return fmt.Errorf("vehicle is at full capacity")
	}

	// Check if vehicle is destroyed
	if vehicleData.IsDestroyed() {
		return fmt.Errorf("vehicle is destroyed")
	}

	// Get rider and vehicle positions for offset calculation
	riderPosComp, hasRiderPos := rider.GetComponent("position")
	vehiclePosComp, hasVehiclePos := vehicle.GetComponent("position")
	if !hasRiderPos || !hasVehiclePos {
		return fmt.Errorf("missing position component")
	}

	riderPos, ok := riderPosComp.(*PositionComponent)
	if !ok {
		return fmt.Errorf("invalid rider position component type")
	}
	vehiclePos, ok := vehiclePosComp.(*PositionComponent)
	if !ok {
		return fmt.Errorf("invalid vehicle position component type")
	}

	// Calculate offset (preserve relative position)
	offsetX := riderPos.X - vehiclePos.X
	offsetY := riderPos.Y - vehiclePos.Y

	// Create mount component
	mountComp := NewMountComponent(vehicle.ID, offsetX, offsetY)
	rider.AddComponent(mountComp)

	// Update vehicle passenger count
	vehicleData.AddPassenger()

	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"rider_id":   rider.ID,
			"vehicle_id": vehicle.ID,
		}).Debug("Entity mounted vehicle")
	}

	return nil
}

// Dismount removes a rider entity from their current vehicle.
// Implements VehicleController interface.
func (ms *MountingSystem) Dismount(rider *Entity) error {
	if rider == nil {
		return fmt.Errorf("rider entity is nil")
	}

	// Check if rider is mounted
	mountComp, isMounted := rider.GetComponent("mount")
	if !isMounted {
		return fmt.Errorf("rider is not mounted")
	}

	mount, ok := mountComp.(*MountComponent)
	if !ok {
		return fmt.Errorf("invalid mount component type")
	}

	// Find vehicle and update passenger count
	if ms.world != nil {
		vehicle, exists := ms.world.GetEntity(mount.MountedEntityID)
		if exists && vehicle != nil {
			if vehicleComp, hasVehicle := vehicle.GetComponent("vehicle"); hasVehicle {
				if vehicleData, ok := vehicleComp.(*VehicleComponent); ok {
					vehicleData.RemovePassenger()
				}
			}
		}
	}

	// Remove mount component
	rider.RemoveComponent("mount")

	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"rider_id":   rider.ID,
			"vehicle_id": mount.MountedEntityID,
		}).Debug("Entity dismounted vehicle")
	}

	return nil
}

// IsMounted checks if an entity is currently riding a vehicle.
// Implements VehicleController interface.
func (ms *MountingSystem) IsMounted(rider *Entity) bool {
	if rider == nil {
		return false
	}
	_, isMounted := rider.GetComponent("mount")
	return isMounted
}

// GetMountedVehicle returns the vehicle entity the rider is on.
// Implements VehicleController interface.
// Returns nil if rider is not mounted or vehicle doesn't exist.
func (ms *MountingSystem) GetMountedVehicle(rider *Entity) *Entity {
	if rider == nil {
		return nil
	}

	mountComp, isMounted := rider.GetComponent("mount")
	if !isMounted {
		return nil
	}

	mount, ok := mountComp.(*MountComponent)
	if !ok {
		return nil
	}

	if ms.world != nil {
		vehicle, exists := ms.world.GetEntity(mount.MountedEntityID)
		if exists {
			return vehicle
		}
	}

	return nil
}

// forceDismount removes mount component without vehicle cleanup.
// Used when vehicle no longer exists.
func (ms *MountingSystem) forceDismount(rider *Entity) {
	if ms.logger != nil {
		ms.logger.WithField("rider_id", rider.ID).Warn("Force dismounting - vehicle not found")
	}
	rider.RemoveComponent("mount")
}

// findEntity finds an entity by ID in the entity list.
// Returns nil if not found.
func (ms *MountingSystem) findEntity(entities []*Entity, id uint64) *Entity {
	for _, entity := range entities {
		if entity.ID == id {
			return entity
		}
	}
	return nil
}
