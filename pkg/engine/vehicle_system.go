package engine

import (
	"errors"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// VehicleSystem manages vehicle movement, fuel consumption, and terrain validation.
// Operates on entities with VehicleComponent and MountComponent.
//
// Phase 21.2: Vehicle System
type VehicleSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewVehicleSystem creates a new vehicle management system.
func NewVehicleSystem(world *World) *VehicleSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system": "vehicle",
	})
	return &VehicleSystem{
		world:  world,
		logger: logger,
	}
}

// Update processes vehicle physics, fuel consumption, and mounting.
func (s *VehicleSystem) Update(deltaTime float64) {
	// Get all vehicles
	vehicles := s.world.GetEntitiesWith("vehicle")

	for _, vehicle := range vehicles {
		vehicleComp, ok := vehicle.GetComponent("vehicle").(*VehicleComponent)
		if !ok {
			continue
		}

		// Process mounted vehicles
		mountComp, hasMountComp := vehicle.GetComponent("mount").(*MountComponent)
		if hasMountComp && mountComp.RiderID != 0 {
			s.processMountedVehicle(vehicle, vehicleComp, mountComp, deltaTime)
		}

		// Update fuel consumption
		s.updateFuel(vehicle, vehicleComp, deltaTime)

		// Validate terrain traversal
		s.validateTerrain(vehicle, vehicleComp)

		// Apply speed decay (friction/drag)
		s.applySpeedDecay(vehicleComp, deltaTime)
	}
}

// processMountedVehicle handles physics for vehicles with riders.
func (s *VehicleSystem) processMountedVehicle(vehicle *Entity, vehicleComp *VehicleComponent, mountComp *MountComponent, deltaTime float64) {
	// Get rider entity
	rider := s.world.GetEntity(mountComp.RiderID)
	if rider == nil {
		// Rider no longer exists, dismount
		mountComp.RiderID = 0
		mountComp.IsMounted = false
		return
	}

	// Get rider input component for acceleration/turning
	inputComp, hasInput := rider.GetComponent("input").(*InputComponent)
	if !hasInput {
		return
	}

	// Apply acceleration based on input
	if inputComp.MoveX != 0 || inputComp.MoveY != 0 {
		// Calculate desired speed based on input magnitude
		inputMagnitude := 1.0 // Simplified, could use sqrt(moveX^2 + moveY^2)
		targetSpeed := vehicleComp.MaxSpeed * inputMagnitude

		// Accelerate toward target speed
		if vehicleComp.Speed < targetSpeed {
			vehicleComp.Speed += vehicleComp.Acceleration * deltaTime
			if vehicleComp.Speed > targetSpeed {
				vehicleComp.Speed = targetSpeed
			}
		} else if vehicleComp.Speed > targetSpeed {
			vehicleComp.Speed -= vehicleComp.Acceleration * deltaTime * 2.0 // Deceleration faster than acceleration
			if vehicleComp.Speed < targetSpeed {
				vehicleComp.Speed = targetSpeed
			}
		}
	}

	// Apply vehicle speed to rider's velocity
	if velocityComp, hasVelocity := rider.GetComponent("velocity").(*VelocityComponent); hasVelocity {
		// Use rider's input direction for movement
		if inputComp.MoveX != 0 || inputComp.MoveY != 0 {
			// Normalize input direction
			length := 1.0 // Simplified
			velocityComp.Vx = (inputComp.MoveX / length) * vehicleComp.Speed
			velocityComp.Vy = (inputComp.MoveY / length) * vehicleComp.Speed
		}
	}

	// Sync vehicle position with rider
	if riderPos, hasRiderPos := rider.GetComponent("position").(*PositionComponent); hasRiderPos {
		if vehiclePos, hasVehiclePos := vehicle.GetComponent("position").(*PositionComponent); hasVehiclePos {
			vehiclePos.X = riderPos.X
			vehiclePos.Y = riderPos.Y
		}
	}
}

// updateFuel reduces fuel based on speed and deltaTime.
func (s *VehicleSystem) updateFuel(vehicle *Entity, vehicleComp *VehicleComponent, deltaTime float64) {
	// Only consume fuel when moving
	if vehicleComp.Speed > 0 {
		// Fuel consumption proportional to speed
		fuelConsumption := (vehicleComp.Speed / vehicleComp.MaxSpeed) * deltaTime * 10.0
		vehicleComp.FuelAmount -= fuelConsumption

		if vehicleComp.FuelAmount < 0 {
			vehicleComp.FuelAmount = 0
			// Out of fuel - force stop
			vehicleComp.Speed = 0
		}
	}
}

// validateTerrain checks if vehicle can traverse current terrain.
func (s *VehicleSystem) validateTerrain(vehicle *Entity, vehicleComp *VehicleComponent) {
	posComp, hasPos := vehicle.GetComponent("position").(*PositionComponent)
	if !hasPos {
		return
	}

	// Get terrain at vehicle position (simplified - would need terrain query in real implementation)
	// For now, we'll skip actual terrain validation as it requires terrain data access
	// In full implementation, this would query the terrain map and check TerrainTypes

	// Placeholder validation
	_ = posComp
	_ = vehicleComp.TerrainTypes
}

// applySpeedDecay reduces speed over time (friction/drag).
func (s *VehicleSystem) applySpeedDecay(vehicleComp *VehicleComponent, deltaTime float64) {
	// Apply drag - reduces speed exponentially
	dragFactor := 0.95 // 5% speed loss per second
	decayRate := 1.0 - ((1.0 - dragFactor) * deltaTime)
	vehicleComp.Speed *= decayRate

	// Stop completely if speed falls below threshold
	if vehicleComp.Speed < 0.1 {
		vehicleComp.Speed = 0
	}
}

// CanTraverse checks if a vehicle can traverse a specific terrain type.
func (s *VehicleSystem) CanTraverse(vehicleComp *VehicleComponent, terrainType terrain.TileType) bool {
	for _, allowedType := range vehicleComp.TerrainTypes {
		if allowedType == terrainType {
			return true
		}
	}
	return false
}

// Mount attaches a rider to a vehicle.
func (s *VehicleSystem) Mount(riderEntity, vehicleEntity *Entity) error {
	vehicleComp, ok := vehicleEntity.GetComponent("vehicle").(*VehicleComponent)
	if !ok {
		return ErrComponentNotFound
	}

	mountComp, ok := vehicleEntity.GetComponent("mount").(*MountComponent)
	if !ok {
		// Create mount component if it doesn't exist
		mountComp = &MountComponent{}
		vehicleEntity.AddComponent(mountComp)
	}

	// Check capacity
	if vehicleComp.CurrentPassengers >= vehicleComp.Capacity {
		s.logger.Warn("Vehicle at capacity, cannot mount")
		return ErrVehicleFullCapacity
	}

	// Mount the rider
	mountComp.RiderID = riderEntity.ID
	mountComp.IsMounted = true
	vehicleComp.CurrentPassengers++

	s.logger.WithFields(logrus.Fields{
		"rider":   riderEntity.ID,
		"vehicle": vehicleEntity.ID,
	}).Info("Rider mounted vehicle")

	return nil
}

// Dismount detaches a rider from a vehicle.
func (s *VehicleSystem) Dismount(riderEntity, vehicleEntity *Entity) error {
	vehicleComp, ok := vehicleEntity.GetComponent("vehicle").(*VehicleComponent)
	if !ok {
		return ErrComponentNotFound
	}

	mountComp, ok := vehicleEntity.GetComponent("mount").(*MountComponent)
	if !ok {
		return ErrComponentNotFound
	}

	if mountComp.RiderID != riderEntity.ID {
		s.logger.Warn("Rider is not mounted on this vehicle")
		return ErrNotMounted
	}

	// Dismount the rider
	mountComp.RiderID = 0
	mountComp.IsMounted = false
	vehicleComp.CurrentPassengers--

	// Stop vehicle when rider dismounts
	vehicleComp.Speed = 0

	s.logger.WithFields(logrus.Fields{
		"rider":   riderEntity.ID,
		"vehicle": vehicleEntity.ID,
	}).Info("Rider dismounted vehicle")

	return nil
}

// IsMounted checks if a rider is currently mounted on any vehicle.
func (s *VehicleSystem) IsMounted(riderEntity *Entity) bool {
	// Check all vehicles for this rider
	vehicles := s.world.GetEntitiesWith("vehicle", "mount")
	for _, vehicle := range vehicles {
		mountComp, ok := vehicle.GetComponent("mount").(*MountComponent)
		if ok && mountComp.RiderID == riderEntity.ID && mountComp.IsMounted {
			return true
		}
	}
	return false
}

// GetMountedVehicle returns the vehicle entity that the rider is mounted on, or nil.
func (s *VehicleSystem) GetMountedVehicle(riderEntity *Entity) *Entity {
	vehicles := s.world.GetEntitiesWith("vehicle", "mount")
	for _, vehicle := range vehicles {
		mountComp, ok := vehicle.GetComponent("mount").(*MountComponent)
		if ok && mountComp.RiderID == riderEntity.ID && mountComp.IsMounted {
			return vehicle
		}
	}
	return nil
}

// Custom errors for vehicle system
var (
	ErrComponentNotFound   = errors.New("component not found")
	ErrVehicleFullCapacity = errors.New("vehicle at full capacity")
	ErrNotMounted          = errors.New("entity not mounted on vehicle")
)
