// Package engine provides the rotation system for entity orientation.
// This file implements RotationSystem which updates entity facing directions
// based on aim input, supporting smooth rotation interpolation.
package engine

// RotationSystem manages entity rotation and orientation.
// Updates RotationComponent based on AimComponent input, enabling
// smooth transitions between facing directions. Works in conjunction
// with InputSystem (sets aim) and RenderSystem (renders rotated sprites).
type RotationSystem struct {
	world *World
}

// NewRotationSystem creates a new rotation system.
func NewRotationSystem(world *World) *RotationSystem {
	return &RotationSystem{
		world: world,
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
	entities := s.world.GetEntitiesWith("rotation")

	for _, entity := range entities {
		// Get rotation component
		rotComp, ok := entity.GetComponent("rotation")
		if !ok {
			continue
		}
		rotation := rotComp.(*RotationComponent)

		// Sync rotation target with aim component if present
		if entity.HasComponent("aim") {
			aimComp, ok := entity.GetComponent("aim")
			if ok {
				aim := aimComp.(*AimComponent)

				// Update aim angle from position if target-based
				if entity.HasComponent("position") {
					posComp, ok := entity.GetComponent("position")
					if ok {
						pos := posComp.(*PositionComponent)
						aim.UpdateAimAngle(pos.X, pos.Y)
					}
				}

				// Set rotation target to match aim
				rotation.SetTargetAngle(aim.AimAngle)
			}
		}

		// Perform smooth rotation update
		rotation.Update(deltaTime)
	}
}

// SyncRotationToAim immediately sets an entity's rotation to match aim.
// Useful for initialization or when instant alignment is needed.
// entityID: ID of entity to sync
// Returns true if sync was successful, false if entity or components not found
func (s *RotationSystem) SyncRotationToAim(entityID uint64) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return false
	}

	if !entity.HasComponent("rotation") || !entity.HasComponent("aim") {
		return false
	}

	rotComp, _ := entity.GetComponent("rotation")
	rotation := rotComp.(*RotationComponent)

	aimComp, _ := entity.GetComponent("aim")
	aim := aimComp.(*AimComponent)

	rotation.SetAngleImmediate(aim.AimAngle)
	return true
}

// SetEntityRotation sets an entity's rotation angle immediately.
// entityID: ID of entity to rotate
// angle: rotation angle in radians
// Returns true if successful, false if entity or component not found
func (s *RotationSystem) SetEntityRotation(entityID uint64, angle float64) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return false
	}

	if !entity.HasComponent("rotation") {
		return false
	}

	rotComp, _ := entity.GetComponent("rotation")
	rotation := rotComp.(*RotationComponent)

	rotation.SetAngleImmediate(angle)
	return true
}

// GetEntityRotation returns an entity's current rotation angle.
// entityID: ID of entity to query
// Returns angle in radians and ok status (false if entity or component not found)
func (s *RotationSystem) GetEntityRotation(entityID uint64) (float64, bool) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return 0, false
	}

	if !entity.HasComponent("rotation") {
		return 0, false
	}

	rotComp, _ := entity.GetComponent("rotation")
	rotation := rotComp.(*RotationComponent)

	return rotation.Angle, true
}

// EnableSmoothRotation enables or disables smooth rotation for an entity.
// When disabled, rotation snaps instantly to target angle.
// entityID: ID of entity to configure
// enabled: true for smooth rotation, false for instant
// Returns true if successful, false if entity or component not found
func (s *RotationSystem) EnableSmoothRotation(entityID uint64, enabled bool) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return false
	}

	if !entity.HasComponent("rotation") {
		return false
	}

	rotComp, _ := entity.GetComponent("rotation")
	rotation := rotComp.(*RotationComponent)

	rotation.SmoothRotation = enabled
	return true
}

// SetRotationSpeed sets the maximum rotation rate for an entity.
// entityID: ID of entity to configure
// speed: rotation speed in radians per second
// Returns true if successful, false if entity or component not found
func (s *RotationSystem) SetRotationSpeed(entityID uint64, speed float64) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return false
	}

	if !entity.HasComponent("rotation") {
		return false
	}

	rotComp, _ := entity.GetComponent("rotation")
	rotation := rotComp.(*RotationComponent)

	rotation.RotationSpeed = speed
	return true
}
