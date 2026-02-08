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
		mountComp, isMounted := entity.GetComponent("mount")
		if !isMounted {
			continue
		}

		mount, ok := mountComp.(*MountComponent)
		if !ok {
			continue
		}

		ms.processMountedEntity(entity, mount, entities)
	}
}

// processMountedEntity handles the logic for a single mounted entity.
func (ms *MountingSystem) processMountedEntity(entity *Entity, mount *MountComponent, entities []*Entity) {
	vehicle := ms.findEntity(entities, mount.MountedEntityID)
	if vehicle == nil {
		ms.forceDismount(entity)
		return
	}

	vehiclePos := ms.getVehiclePosition(vehicle)
	if vehiclePos == nil {
		return
	}

	ms.syncRiderPosition(entity, vehiclePos, mount)
	ms.syncRiderRotation(entity, vehicle)
}

// getVehiclePosition retrieves and validates the vehicle's position component.
func (ms *MountingSystem) getVehiclePosition(vehicle *Entity) *PositionComponent {
	vehiclePosComp, hasVehiclePos := vehicle.GetComponent("position")
	if !hasVehiclePos {
		return nil
	}
	vehiclePos, ok := vehiclePosComp.(*PositionComponent)
	if !ok {
		return nil
	}
	return vehiclePos
}

// syncRiderPosition updates the rider's position to match the vehicle with offset.
func (ms *MountingSystem) syncRiderPosition(entity *Entity, vehiclePos *PositionComponent, mount *MountComponent) {
	riderPosComp, hasRiderPos := entity.GetComponent("position")
	if !hasRiderPos {
		return
	}
	riderPos, ok := riderPosComp.(*PositionComponent)
	if !ok {
		return
	}
	riderPos.X = vehiclePos.X + mount.MountOffset.X
	riderPos.Y = vehiclePos.Y + mount.MountOffset.Y
}

// syncRiderRotation synchronizes the rider's rotation with the vehicle.
func (ms *MountingSystem) syncRiderRotation(entity, vehicle *Entity) {
	vehicleRotComp, hasVehicleRot := vehicle.GetComponent("rotation")
	riderRotComp, hasRiderRot := entity.GetComponent("rotation")
	if !hasVehicleRot || !hasRiderRot {
		return
	}
	vehicleRot, ok := vehicleRotComp.(*RotationComponent)
	if !ok {
		return
	}
	riderRot, ok := riderRotComp.(*RotationComponent)
	if !ok {
		return
	}
	riderRot.Angle = vehicleRot.Angle
}

// Mount attempts to mount a rider entity onto a vehicle entity.
// Implements VehicleController interface.
func (ms *MountingSystem) Mount(rider, vehicle *Entity) error {
	if err := ms.validateMountEntities(rider, vehicle); err != nil {
		return err
	}

	vehicleData, err := ms.validateVehicleState(vehicle)
	if err != nil {
		return err
	}

	riderPos, vehiclePos, err := ms.extractPositionComponents(rider, vehicle)
	if err != nil {
		return err
	}

	ms.createMountRelationship(rider, vehicle.ID, vehicleData, riderPos, vehiclePos)

	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"rider_id":   rider.ID,
			"vehicle_id": vehicle.ID,
		}).Debug("Entity mounted vehicle")
	}

	return nil
}

// validateMountEntities validates rider and vehicle entities for mounting.
func (ms *MountingSystem) validateMountEntities(rider, vehicle *Entity) error {
	if rider == nil {
		return fmt.Errorf("rider entity is nil")
	}
	if vehicle == nil {
		return fmt.Errorf("vehicle entity is nil")
	}

	if _, isMounted := rider.GetComponent("mount"); isMounted {
		return fmt.Errorf("rider is already mounted")
	}

	return nil
}

// validateVehicleState validates the vehicle component and its state.
func (ms *MountingSystem) validateVehicleState(vehicle *Entity) (*VehicleComponent, error) {
	vehicleComp, hasVehicle := vehicle.GetComponent("vehicle")
	if !hasVehicle {
		return nil, fmt.Errorf("entity is not a vehicle")
	}

	vehicleData, ok := vehicleComp.(*VehicleComponent)
	if !ok {
		return nil, fmt.Errorf("invalid vehicle component type")
	}

	if !vehicleData.CanAddPassenger() {
		return nil, fmt.Errorf("vehicle is at full capacity")
	}

	if vehicleData.IsDestroyed() {
		return nil, fmt.Errorf("vehicle is destroyed")
	}

	return vehicleData, nil
}

// extractPositionComponents extracts and validates position components.
func (ms *MountingSystem) extractPositionComponents(rider, vehicle *Entity) (*PositionComponent, *PositionComponent, error) {
	riderPosComp, hasRiderPos := rider.GetComponent("position")
	vehiclePosComp, hasVehiclePos := vehicle.GetComponent("position")
	if !hasRiderPos || !hasVehiclePos {
		return nil, nil, fmt.Errorf("missing position component")
	}

	riderPos, ok := riderPosComp.(*PositionComponent)
	if !ok {
		return nil, nil, fmt.Errorf("invalid rider position component type")
	}

	vehiclePos, ok := vehiclePosComp.(*PositionComponent)
	if !ok {
		return nil, nil, fmt.Errorf("invalid vehicle position component type")
	}

	return riderPos, vehiclePos, nil
}

// createMountRelationship creates mount component and updates vehicle.
func (ms *MountingSystem) createMountRelationship(rider *Entity, vehicleID uint64, vehicleData *VehicleComponent, riderPos, vehiclePos *PositionComponent) {
	offsetX := riderPos.X - vehiclePos.X
	offsetY := riderPos.Y - vehiclePos.Y

	mountComp := NewMountComponent(vehicleID, offsetX, offsetY)
	rider.AddComponent(mountComp)

	vehicleData.AddPassenger()
}

// Dismount removes a rider entity from their current vehicle.
// Implements VehicleController interface.
func (ms *MountingSystem) Dismount(rider *Entity) error {
	if rider == nil {
		return fmt.Errorf("rider entity is nil")
	}

	mount, err := ms.validateMountComponent(rider)
	if err != nil {
		return err
	}

	ms.updateVehiclePassengerCount(mount.MountedEntityID)
	rider.RemoveComponent("mount")
	ms.logDismount(rider.ID, mount.MountedEntityID)

	return nil
}

// validateMountComponent retrieves and validates the mount component from a rider.
func (ms *MountingSystem) validateMountComponent(rider *Entity) (*MountComponent, error) {
	mountComp, isMounted := rider.GetComponent("mount")
	if !isMounted {
		return nil, fmt.Errorf("rider is not mounted")
	}

	mount, ok := mountComp.(*MountComponent)
	if !ok {
		return nil, fmt.Errorf("invalid mount component type")
	}

	return mount, nil
}

// updateVehiclePassengerCount decrements the passenger count on the vehicle.
func (ms *MountingSystem) updateVehiclePassengerCount(vehicleID uint64) {
	if ms.world == nil {
		return
	}

	vehicle, exists := ms.world.GetEntity(vehicleID)
	if !exists || vehicle == nil {
		return
	}

	vehicleComp, hasVehicle := vehicle.GetComponent("vehicle")
	if !hasVehicle {
		return
	}

	vehicleData, ok := vehicleComp.(*VehicleComponent)
	if ok {
		vehicleData.RemovePassenger()
	}
}

// logDismount logs the dismount event.
func (ms *MountingSystem) logDismount(riderID, vehicleID uint64) {
	if ms.logger != nil {
		ms.logger.WithFields(logrus.Fields{
			"rider_id":   riderID,
			"vehicle_id": vehicleID,
		}).Debug("Entity dismounted vehicle")
	}
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
