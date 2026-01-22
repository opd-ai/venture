// Package engine provides the lifetime management system.
// This file implements LifetimeSystem which automatically despawns entities
// after their lifetime duration expires. Used for temporary entities like
// spell lights, particle effects, and timed buffs.
//
// Design Philosophy:
// - Simple and efficient: O(n) iteration over entities with lifetime
// - Automatic cleanup: no manual despawn code needed
// - ECS integration: uses standard component and system patterns
package engine

import (
	"github.com/sirupsen/logrus"
)

// LifetimeSystem manages entities with limited lifespans.
// Entities with LifetimeComponent are automatically despawned when their
// duration expires.
type LifetimeSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewLifetimeSystem creates a new lifetime management system.
func NewLifetimeSystem(world *World) *LifetimeSystem {
	system := NewLifetimeSystemWithLogger(world, nil)
	if system.logger != nil {
		system.logger.Debug("LifetimeSystem created without custom logger")
	}
	return system
}

// NewLifetimeSystemWithLogger creates a new lifetime system with a logger.
func NewLifetimeSystemWithLogger(world *World, logger *logrus.Logger) *LifetimeSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "lifetime")
		logEntry.WithFields(logrus.Fields{
			"has_world": world != nil,
		}).Debug("LifetimeSystem created with custom logger")
	}

	return &LifetimeSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes all entities with LifetimeComponent and despawns expired ones.
// logUpdateStart logs the beginning of a lifetime system update cycle
func (s *LifetimeSystem) logUpdateStart(entityCount int, deltaTime float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count": entityCount,
			"delta_time":   deltaTime,
		}).Debug("LifetimeSystem update started")
	}
}

// logLifetimeUpdate logs the lifetime progression for a specific entity
func (s *LifetimeSystem) logLifetimeUpdate(entity *Entity, lifetime *LifetimeComponent, previousElapsed, deltaTime float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":        entity.ID,
			"elapsed":          lifetime.Elapsed,
			"duration":         lifetime.Duration,
			"remaining":        lifetime.Duration - lifetime.Elapsed,
			"previous_elapsed": previousElapsed,
			"delta_time":       deltaTime,
		}).Debug("updated entity lifetime")
	}
}

// logEntityExpired logs when an entity's lifetime has expired and it is despawned
func (s *LifetimeSystem) logEntityExpired(entity *Entity, lifetime *LifetimeComponent) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"duration":  lifetime.Duration,
			"elapsed":   lifetime.Elapsed,
		}).Debug("entity lifetime expired, despawned")
	}
}

// logUpdateComplete logs the completion of a lifetime system update cycle
func (s *LifetimeSystem) logUpdateComplete(processedCount, expiredCount int, deltaTime float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"processed_count": processedCount,
			"expired_count":   expiredCount,
			"delta_time":      deltaTime,
		}).Debug("LifetimeSystem update completed")
	}
}

// processEntityLifetime updates lifetime for a single entity and returns true if expired
func (s *LifetimeSystem) processEntityLifetime(entity *Entity, deltaTime float64) bool {
	lifetimeComp, hasLifetime := entity.GetComponent("lifetime")
	if !hasLifetime {
		return false
	}

	lifetime, ok := lifetimeComp.(*LifetimeComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"component_type": "lifetime",
			}).Warn("failed to cast lifetime component to LifetimeComponent")
		}
		return false
	}

	previousElapsed := lifetime.Elapsed
	lifetime.Elapsed += deltaTime

	s.logLifetimeUpdate(entity, lifetime, previousElapsed, deltaTime)

	if lifetime.Elapsed >= lifetime.Duration {
		s.world.RemoveEntity(entity.ID)
		s.logEntityExpired(entity, lifetime)
		return true
	}

	return false
}

func (s *LifetimeSystem) Update(entities []*Entity, deltaTime float64) {
	s.logUpdateStart(len(entities), deltaTime)

	expiredCount := 0
	for _, entity := range entities {
		if s.processEntityLifetime(entity, deltaTime) {
			expiredCount++
		}
	}

	s.logUpdateComplete(len(entities), expiredCount, deltaTime)
}
