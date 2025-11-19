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
func (s *LifetimeSystem) Update(entities []*Entity, deltaTime float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("LifetimeSystem update started")
	}

	expiredCount := 0
	for _, entity := range entities {
		lifetimeComp, hasLifetime := entity.GetComponent("lifetime")
		if !hasLifetime {
			continue
		}

		lifetime, ok := lifetimeComp.(*LifetimeComponent)
		if !ok {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":      entity.ID,
					"component_type": "lifetime",
				}).Warn("failed to cast lifetime component to LifetimeComponent")
			}
			continue
		}

		previousElapsed := lifetime.Elapsed
		lifetime.Elapsed += deltaTime

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

		// Check if lifetime expired
		if lifetime.Elapsed >= lifetime.Duration {
			// Despawn the entity
			s.world.RemoveEntity(entity.ID)
			expiredCount++

			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"duration":  lifetime.Duration,
					"elapsed":   lifetime.Elapsed,
				}).Debug("entity lifetime expired, despawned")
			}
		}
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"processed_count": len(entities),
			"expired_count":   expiredCount,
			"delta_time":      deltaTime,
		}).Debug("LifetimeSystem update completed")
	}
}
