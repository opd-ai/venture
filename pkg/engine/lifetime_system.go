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
	return NewLifetimeSystemWithLogger(world, nil)
}

// NewLifetimeSystemWithLogger creates a new lifetime system with a logger.
func NewLifetimeSystemWithLogger(world *World, logger *logrus.Logger) *LifetimeSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "lifetime")
	}

	return &LifetimeSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes all entities with LifetimeComponent and despawns expired ones.
func (s *LifetimeSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		lifetimeComp, hasLifetime := entity.GetComponent("lifetime")
		if !hasLifetime {
			continue
		}

		lifetime := lifetimeComp.(*LifetimeComponent)
		lifetime.Elapsed += deltaTime

		// Check if lifetime expired
		if lifetime.Elapsed >= lifetime.Duration {
			// Despawn the entity
			s.world.RemoveEntity(entity.ID)

			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"entityID": entity.ID,
					"duration": lifetime.Duration,
				}).Debug("entity lifetime expired, despawned")
			}
		}
	}
}
