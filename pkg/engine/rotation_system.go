// Package engine provides the rotation system for entity orientation.
// This file implements RotationSystem which updates entity facing directions
// based on aim input, supporting smooth rotation interpolation.
package engine

import (
	"github.com/sirupsen/logrus"
)

var rotationLog *logrus.Logger

func init() {
	rotationLog = logrus.New()
	rotationLog.SetReportCaller(true)
	rotationLog.SetLevel(logrus.InfoLevel)
}

// RotationSystem manages entity rotation and orientation.
// Updates RotationComponent based on AimComponent input, enabling
// smooth transitions between facing directions. Works in conjunction
// with InputSystem (sets aim) and RenderSystem (renders rotated sprites).
type RotationSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewRotationSystem creates a new rotation system.
func NewRotationSystem(world *World) *RotationSystem {
	logger := rotationLog.WithFields(logrus.Fields{
		"system_name": "rotation",
	})

	logger.Info("Initializing rotation system")

	return &RotationSystem{
		world:  world,
		logger: logger,
	}
}

// Update processes rotation for all entities with RotationComponent.
// deltaTime: elapsed time in seconds since last update
//
// Update flow:
// 1. Query entities with "rotation" component
// 2. If entity has "aim" component, sync rotation target to aim angle
// 3. Interpolate rotation towards target angle
// 4. Clamp rotation to valid range [0, 2π)
func (s *RotationSystem) Update(deltaTime float64) {
	s.logger.WithFields(logrus.Fields{
		"delta_time": deltaTime,
	}).Debug("Starting rotation system update")

	entities := s.world.GetEntitiesWith("rotation")

	s.logger.WithFields(logrus.Fields{
		"entity_count": len(entities),
	}).Debug("Processing entities with rotation component")

	for _, entity := range entities {
		rotation := s.getRotationComponent(entity)
		if rotation == nil {
			continue
		}

		s.syncRotationWithAim(entity, rotation)
		s.performRotationUpdate(entity, rotation, deltaTime)
	}

	s.logger.Debug("Rotation system update complete")
}

// getRotationComponent retrieves and validates the rotation component from an entity.
func (s *RotationSystem) getRotationComponent(entity *Entity) *RotationComponent {
	rotComp, ok := entity.GetComponent("rotation")
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Warn("Entity has rotation in query but component retrieval failed")
		return nil
	}
	rotation, ok := rotComp.(*RotationComponent)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"component_type": "rotation",
		}).Error("Failed to cast rotation component")
		return nil
	}
	return rotation
}

// syncRotationWithAim synchronizes rotation target with aim component if present.
func (s *RotationSystem) syncRotationWithAim(entity *Entity, rotation *RotationComponent) {
	if !entity.HasComponent("aim") {
		return
	}

	aimComp, ok := entity.GetComponent("aim")
	if !ok {
		return
	}

	aim, ok := aimComp.(*AimComponent)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"component_type": "aim",
		}).Warn("Failed to cast aim component")
		return
	}

	s.logger.WithFields(logrus.Fields{
		"entity_id":    entity.ID,
		"aim_angle":    aim.AimAngle,
		"has_position": entity.HasComponent("position"),
	}).Debug("Syncing rotation with aim component")

	s.updateAimFromPosition(entity, aim)

	oldTarget := rotation.TargetAngle
	rotation.SetTargetAngle(aim.AimAngle)
	s.logger.WithFields(logrus.Fields{
		"entity_id":     entity.ID,
		"old_target":    oldTarget,
		"new_target":    aim.AimAngle,
		"current_angle": rotation.Angle,
	}).Debug("Set rotation target from aim")
}

// updateAimFromPosition updates aim angle based on entity position if available.
func (s *RotationSystem) updateAimFromPosition(entity *Entity, aim *AimComponent) {
	if !entity.HasComponent("position") {
		return
	}

	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	aim.UpdateAimAngle(pos.X, pos.Y)
	s.logger.WithFields(logrus.Fields{
		"entity_id":     entity.ID,
		"position_x":    pos.X,
		"position_y":    pos.Y,
		"updated_angle": aim.AimAngle,
	}).Debug("Updated aim angle from position")
}

// performRotationUpdate executes the rotation interpolation and logs changes.
func (s *RotationSystem) performRotationUpdate(entity *Entity, rotation *RotationComponent, deltaTime float64) {
	oldAngle := rotation.Angle
	rotation.Update(deltaTime)

	if oldAngle != rotation.Angle {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"old_angle":      oldAngle,
			"new_angle":      rotation.Angle,
			"target_angle":   rotation.TargetAngle,
			"smooth_enabled": rotation.SmoothRotation,
			"rotation_speed": rotation.RotationSpeed,
		}).Debug("Entity rotation updated")
	}
}

// SyncRotationToAim immediately sets an entity's rotation to match aim.
// Useful for initialization or when instant alignment is needed.
// entityID: ID of entity to sync
// Returns true if sync was successful, false if entity or components not found
func (s *RotationSystem) SyncRotationToAim(entityID uint64) bool {
	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"operation": "sync_rotation_to_aim",
	}).Debug("Syncing rotation to aim")

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("Entity not found for rotation sync")
		return false
	}

	if !entity.HasComponent("rotation") || !entity.HasComponent("aim") {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entityID,
			"has_rotation": entity.HasComponent("rotation"),
			"has_aim":      entity.HasComponent("aim"),
		}).Warn("Missing required components for rotation sync")
		return false
	}

	rotComp, _ := entity.GetComponent("rotation")
	if rotComp == nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Rotation component is nil")
		return false
	}
	rotation, ok := rotComp.(*RotationComponent)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Failed to cast rotation component")
		return false
	}

	aimComp, _ := entity.GetComponent("aim")
	if aimComp == nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "aim",
		}).Error("Aim component is nil")
		return false
	}
	aim, ok := aimComp.(*AimComponent)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "aim",
		}).Error("Failed to cast aim component")
		return false
	}

	oldAngle := rotation.Angle
	rotation.SetAngleImmediate(aim.AimAngle)

	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"old_angle": oldAngle,
		"new_angle": aim.AimAngle,
	}).Info("Synced rotation to aim angle")

	return true
}

// SetEntityRotation sets an entity's rotation angle immediately.
// entityID: ID of entity to rotate
// angle: rotation angle in radians
// Returns true if successful, false if entity or component not found
func (s *RotationSystem) SetEntityRotation(entityID uint64, angle float64) bool {
	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"angle":     angle,
		"operation": "set_entity_rotation",
	}).Debug("Setting entity rotation")

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("Entity not found for rotation set")
		return false
	}

	if !entity.HasComponent("rotation") {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("Entity missing rotation component")
		return false
	}

	rotComp, _ := entity.GetComponent("rotation")
	if rotComp == nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Rotation component is nil")
		return false
	}
	rotation, ok := rotComp.(*RotationComponent)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Failed to cast rotation component")
		return false
	}

	oldAngle := rotation.Angle
	rotation.SetAngleImmediate(angle)

	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"old_angle": oldAngle,
		"new_angle": angle,
	}).Info("Set entity rotation angle")

	return true
}

// GetEntityRotation returns an entity's current rotation angle.
// entityID: ID of entity to query
// Returns angle in radians and ok status (false if entity or component not found)
func (s *RotationSystem) GetEntityRotation(entityID uint64) (float64, bool) {
	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"operation": "get_entity_rotation",
	}).Debug("Getting entity rotation")

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("Entity not found for rotation query")
		return 0, false
	}

	if !entity.HasComponent("rotation") {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Debug("Entity missing rotation component")
		return 0, false
	}

	rotComp, _ := entity.GetComponent("rotation")
	if rotComp == nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Rotation component is nil")
		return 0, false
	}
	rotation, ok := rotComp.(*RotationComponent)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Failed to cast rotation component")
		return 0, false
	}

	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"angle":     rotation.Angle,
	}).Debug("Retrieved entity rotation")

	return rotation.Angle, true
}

// EnableSmoothRotation enables or disables smooth rotation for an entity.
// When disabled, rotation snaps instantly to target angle.
// entityID: ID of entity to configure
// enabled: true for smooth rotation, false for instant
// Returns true if successful, false if entity or component not found
func (s *RotationSystem) EnableSmoothRotation(entityID uint64, enabled bool) bool {
	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"enabled":   enabled,
		"operation": "enable_smooth_rotation",
	}).Debug("Configuring smooth rotation")

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("Entity not found for smooth rotation config")
		return false
	}

	if !entity.HasComponent("rotation") {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("Entity missing rotation component")
		return false
	}

	rotComp, _ := entity.GetComponent("rotation")
	if rotComp == nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Rotation component is nil")
		return false
	}
	rotation, ok := rotComp.(*RotationComponent)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Failed to cast rotation component")
		return false
	}

	oldValue := rotation.SmoothRotation
	rotation.SmoothRotation = enabled

	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"old_value": oldValue,
		"new_value": enabled,
	}).Info("Updated smooth rotation setting")

	return true
}

// SetRotationSpeed sets the maximum rotation rate for an entity.
// entityID: ID of entity to configure
// speed: rotation speed in radians per second
// Returns true if successful, false if entity or component not found
func (s *RotationSystem) SetRotationSpeed(entityID uint64, speed float64) bool {
	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"speed":     speed,
		"operation": "set_rotation_speed",
	}).Debug("Setting rotation speed")

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("Entity not found for rotation speed config")
		return false
	}

	if !entity.HasComponent("rotation") {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Warn("Entity missing rotation component")
		return false
	}

	rotComp, _ := entity.GetComponent("rotation")
	if rotComp == nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Rotation component is nil")
		return false
	}
	rotation, ok := rotComp.(*RotationComponent)
	if !ok {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entityID,
			"component_type": "rotation",
		}).Error("Failed to cast rotation component")
		return false
	}

	oldSpeed := rotation.RotationSpeed
	rotation.RotationSpeed = speed

	s.logger.WithFields(logrus.Fields{
		"entity_id": entityID,
		"old_speed": oldSpeed,
		"new_speed": speed,
	}).Info("Updated rotation speed")

	return true
}
