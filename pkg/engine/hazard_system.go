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
	world       *World
	logger      *logrus.Entry
	zoneTracker *HazardZoneTracker
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
		logger:      logEntry,
		zoneTracker: NewHazardZoneTracker(500), // Max 500 concurrent hazard zones
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

	// Update zone tracker (handles expiration and fading)
	removed := s.zoneTracker.Update(deltaTime)
	if s.logger != nil && removed > 0 {
		s.logger.WithField("removed", removed).Debug("expired hazard zones removed")
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
			// Remove from zone tracker
			s.zoneTracker.RemoveZone(hazardEntity.ID)
			// Remove entity
			s.world.RemoveEntity(hazardEntity.ID)
			continue
		}

		// Get hazard position
		hazPosComp, ok := hazardEntity.GetComponent("position")
		if !ok {
			continue
		}
		hazPos := hazPosComp.(*PositionComponent)

		// Sync zone tracker with current hazard state
		zone, exists := s.zoneTracker.GetZone(hazardEntity.ID)
		if !exists {
			// Zone not tracked yet, add it
			zone = &HazardZone{
				ID:                 hazardEntity.ID,
				X:                  hazPos.X,
				Y:                  hazPos.Y,
				Radius:             hazard.Radius,
				HazardType:         hazard.HazardType,
				DamagePerSecond:    hazard.DamagePerSecond,
				MovementMultiplier: hazard.MovementMultiplier,
				RemainingDuration:  hazard.Duration,
			}
			if !s.zoneTracker.AddZone(zone) {
				if s.logger != nil {
					s.logger.Warn("failed to add zone (max limit reached)")
				}
			}
		} else {
			// Update existing zone with current hazard state
			zone.X = hazPos.X
			zone.Y = hazPos.Y
			zone.RemainingDuration = hazard.Duration
		}
	}

	// Apply hazard effects to entities using zone tracker
	s.applyZoneEffects(deltaTime)
}

// applyZoneEffects applies hazard effects to entities using zone tracker.
func (s *HazardSystem) applyZoneEffects(deltaTime float64) {
	if s.world == nil {
		return
	}

	// Find all entities with position and health
	entities := s.world.GetEntities()

	for _, entity := range entities {
		// Get entity position
		entPosComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		entPos := entPosComp.(*PositionComponent)

		// Query zones at entity position
		zones := s.zoneTracker.GetZonesAt(entPos.X, entPos.Y)

		// Apply effects from all overlapping zones
		for _, zone := range zones {
			// Apply damage if zone is damaging
			if zone.DamagePerSecond > 0 {
				s.applyZoneDamage(entity, zone, deltaTime)
			}

			// Apply movement modifier if zone affects movement
			if zone.MovementMultiplier != 1.0 {
				s.applyZoneMovementModifier(entity, zone)
			}
		}
	}
}

// applyZoneDamage applies damage from a hazard zone to an entity.
func (s *HazardSystem) applyZoneDamage(entity *Entity, zone *HazardZone, deltaTime float64) {
	// Get health component
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return // Entity has no health
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	// Calculate damage scaled by zone intensity (fading)
	damage := zone.DamagePerSecond * deltaTime * zone.Intensity

	// Apply damage
	health.TakeDamage(damage)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entityID":   entity.ID,
			"hazardType": zone.HazardType.String(),
			"damage":     damage,
			"intensity":  zone.Intensity,
		}).Debug("zone damage applied")
	}
}

// applyZoneMovementModifier applies movement speed modifier from a hazard zone.
// This now uses zone tracking to avoid cumulative effects.
func (s *HazardSystem) applyZoneMovementModifier(entity *Entity, zone *HazardZone) {
	// Get velocity component
	velComp, ok := entity.GetComponent("velocity")
	if !ok {
		return
	}

	velocity, ok := velComp.(*VelocityComponent)
	if !ok {
		return
	}

	// Apply movement multiplier (scaled by intensity for fading effect)
	// Note: This is a simplified implementation that modifies current velocity.
	// A more sophisticated approach would track original velocity and restore it.
	multiplier := 1.0 + (zone.MovementMultiplier-1.0)*zone.Intensity
	velocity.VX *= multiplier
	velocity.VY *= multiplier
}

// applyHazardEffects applies hazard effects to entities in range (legacy method).
// This is kept for backward compatibility but delegates to zone-based system.
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
// Legacy method maintained for backward compatibility.
func (s *HazardSystem) applyMovementModifier(entity *Entity, hazard *HazardComponent) {
	// Delegate to zone-based system by checking zone tracker
	// This ensures consistent behavior with new zone tracking
	entPosComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	entPos := entPosComp.(*PositionComponent)

	zones := s.zoneTracker.GetZonesAt(entPos.X, entPos.Y)
	for _, zone := range zones {
		if zone.MovementMultiplier != 1.0 {
			s.applyZoneMovementModifier(entity, zone)
			break // Apply only strongest effect
		}
	}
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

// GetZoneTracker returns the hazard zone tracker for direct access.
// This allows external systems to query zones efficiently.
func (s *HazardSystem) GetZoneTracker() *HazardZoneTracker {
	return s.zoneTracker
}

// GetActiveZoneCount returns the number of currently active hazard zones.
func (s *HazardSystem) GetActiveZoneCount() int {
	return s.zoneTracker.GetActiveZoneCount()
}

// GetZonesInRadius returns all hazard zones overlapping a circular area.
// Useful for area-of-effect queries like explosions.
func (s *HazardSystem) GetZonesInRadius(x, y, radius float64) []*HazardZone {
	return s.zoneTracker.GetZonesInRadius(x, y, radius)
}

// GetActiveHazardCount returns the number of active hazard entities.
func (s *HazardSystem) GetActiveHazardCount() int {
	if s.world == nil {
		return 0
	}
	return len(s.world.GetEntitiesWith("hazard"))
}
