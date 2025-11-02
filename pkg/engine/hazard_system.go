// Package engine provides the hazard system for Phase 11.3.
// Environmental Destruction & Manipulation
//
// This file implements HazardSystem which manages environmental hazards
// (poison clouds, oil puddles, water pools, smoke) and their effects on
// entities (damage, movement speed changes, vision obscuration).
//
// Design Philosophy:
// - Hazards affect entities in their radius
// - Poison deals damage over time
// - Oil and water slow movement
// - Smoke obscures vision (future feature)
// - Hazards have duration and despawn when expired
// - Server-authoritative for multiplayer synchronization
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// HazardSystem manages environmental hazards and their effects.
type HazardSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewHazardSystem creates a new hazard system.
func NewHazardSystem() *HazardSystem {
	return NewHazardSystemWithLogger(nil)
}

// NewHazardSystemWithLogger creates a system with a logger.
func NewHazardSystemWithLogger(logger *logrus.Logger) *HazardSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "hazard",
		})
		logEntry.Debug("hazard system created")
	}

	return &HazardSystem{
		logger: logEntry,
	}
}

// SetWorld sets the ECS world reference.
func (s *HazardSystem) SetWorld(world *World) {
	s.world = world
}

// Update implements the System interface.
// Updates hazard durations, applies effects to entities in range, and removes expired hazards.
func (s *HazardSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Find all hazard entities
	hazards := s.world.GetEntitiesWith("hazard")

	for _, hazardEntity := range hazards {
		// Get hazard component
		hazComp, ok := hazardEntity.GetComponent("hazard")
		if !ok {
			continue
		}
		hazard := hazComp.(*HazardComponent)

		// Update hazard duration
		hazard.Update(deltaTime)

		// Check if hazard should be removed
		if hazard.ShouldRemove() {
			s.world.RemoveEntity(hazardEntity.ID)
			continue
		}

		// Get hazard position
		hazPosComp, ok := hazardEntity.GetComponent("position")
		if !ok {
			continue
		}
		hazPos := hazPosComp.(*PositionComponent)

		// Apply hazard effects to entities in range
		s.applyHazardEffects(hazard, hazPos, deltaTime)
	}
}

// applyHazardEffects applies hazard effects to entities in range.
func (s *HazardSystem) applyHazardEffects(hazard *HazardComponent, hazPos *PositionComponent, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Find all entities
	entities := s.world.GetEntities()

	for _, entity := range entities {
		// Get entity position
		entPosComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		entPos := entPosComp.(*PositionComponent)

		// Calculate distance to hazard
		dx := entPos.X - hazPos.X
		dy := entPos.Y - hazPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Check if entity is in hazard radius
		if dist <= hazard.Radius {
			// Apply damage if hazard is damaging
			if hazard.IsDamaging() {
				s.applyDamage(entity, hazard, deltaTime)
			}

			// Apply movement speed modifier if hazard affects movement
			if hazard.AffectsMovement() {
				s.applyMovementModifier(entity, hazard)
			}
		}
	}
}

// applyDamage applies hazard damage to an entity.
func (s *HazardSystem) applyDamage(entity *Entity, hazard *HazardComponent, deltaTime float64) {
	// Get health component
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return // Entity has no health
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	// Calculate damage for this frame
	damage := hazard.DamagePerSecond * deltaTime

	// Apply damage
	health.TakeDamage(damage)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":   entity.ID,
			"hazardType": hazard.HazardType.String(),
			"damage":     damage,
		}).Debug("hazard damage applied")
	}
}

// applyMovementModifier applies movement speed modifier to an entity.
// Note: This is a simplified implementation that modifies current velocity.
// A more sophisticated approach would store original velocity and restore it
// when the entity leaves the hazard. This requires tracking which entities
// are in which hazards, which is deferred to future enhancement.
func (s *HazardSystem) applyMovementModifier(entity *Entity, hazard *HazardComponent) {
	// For now, we skip movement modification to avoid cumulative effects.
	// Future enhancement: Track entity-hazard relationships and apply modifiers
	// only when entering/exiting hazards, not every frame.
	// TODO: Implement proper hazard zone tracking system
}

// CreateHazard creates a hazard entity at the specified position.
// This is a helper method for external systems (e.g., destructible objects, spells).
func (s *HazardSystem) CreateHazard(hazardType HazardType, x, y, duration, radius float64) uint64 {
	if s.world == nil {
		return 0
	}

	// Create hazard entity
	entity := s.world.CreateEntity()

	// Add position
	posComp := &PositionComponent{X: x, Y: y}
	entity.AddComponent(posComp)

	// Add hazard component
	hazardComp := NewHazardComponent(hazardType, duration, radius)
	entity.AddComponent(hazardComp)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"hazardType": hazardType.String(),
			"x":          x,
			"y":          y,
			"duration":   duration,
			"radius":     radius,
		}).Debug("hazard created")
	}

	return entity.ID
}

// GetHazardsAt returns all hazards affecting the given position.
func (s *HazardSystem) GetHazardsAt(x, y float64) []*Entity {
	if s.world == nil {
		return nil
	}

	var result []*Entity

	// Find all hazard entities
	hazards := s.world.GetEntitiesWith("hazard")

	for _, hazardEntity := range hazards {
		// Get hazard component
		hazComp, ok := hazardEntity.GetComponent("hazard")
		if !ok {
			continue
		}
		hazard := hazComp.(*HazardComponent)

		// Get hazard position
		hazPosComp, ok := hazardEntity.GetComponent("position")
		if !ok {
			continue
		}
		hazPos := hazPosComp.(*PositionComponent)

		// Calculate distance
		dx := x - hazPos.X
		dy := y - hazPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Check if position is in hazard radius
		if dist <= hazard.Radius {
			result = append(result, hazardEntity)
		}
	}

	return result
}

// IsPositionHazardous returns true if the position is affected by any hazards.
func (s *HazardSystem) IsPositionHazardous(x, y float64) bool {
	hazards := s.GetHazardsAt(x, y)
	return len(hazards) > 0
}

// GetActiveHazardCount returns the number of active hazard entities.
func (s *HazardSystem) GetActiveHazardCount() int {
	if s.world == nil {
		return 0
	}
	return len(s.world.GetEntitiesWith("hazard"))
}
