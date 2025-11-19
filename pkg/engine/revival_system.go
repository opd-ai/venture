// Package engine provides the revival system for multiplayer death mechanics.
// This file implements RevivalSystem which allows living teammates to revive
// dead players through proximity interaction.
// Priority 1.5: Multiplayer Revival System
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// RevivalSystem handles player revival mechanics in multiplayer.
// Living players can revive dead teammates by standing nearby and pressing
// the revival input key. This system implements multiplayer revival mechanics
// with proximity detection and health restoration.
type RevivalSystem struct {
	world  *World
	logger *logrus.Entry

	// RevivalRange is the maximum distance for revival (in pixels)
	// Default: 32.0 pixels (one tile)
	RevivalRange float64

	// RevivalAmount is the fraction of max health restored (0.0-1.0)
	// Default: 0.2 (20% health)
	RevivalAmount float64

	// RevivalTime is how long the revival action takes (in seconds)
	// Default: 0.0 (instant revival)
	// Future enhancement: could add channeling time
	RevivalTime float64
}

// NewRevivalSystem creates a new revival system with default parameters.
func NewRevivalSystem(world *World) *RevivalSystem {
	return NewRevivalSystemWithLogger(world, nil)
}

// NewRevivalSystemWithLogger creates a new revival system with a logger.
func NewRevivalSystemWithLogger(world *World, logger *logrus.Logger) *RevivalSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "revival")
	}

	system := &RevivalSystem{
		world:         world,
		logger:        logEntry,
		RevivalRange:  32.0, // One tile range
		RevivalAmount: 0.2,  // 20% health restoration
		RevivalTime:   0.0,  // Instant revival (no channeling)
	}

	if system.logger != nil {
		system.logger.WithFields(logrus.Fields{
			"revival_range":  system.RevivalRange,
			"revival_amount": system.RevivalAmount,
			"revival_time":   system.RevivalTime,
		}).Info("RevivalSystem initialized")
	}

	return system
}

// Update processes revival inputs and handles revival logic.
// Checks for living players pressing revival key near dead players.
func (s *RevivalSystem) Update(entities []*Entity, deltaTime float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_count": len(entities),
			"delta_time":   deltaTime,
		}).Debug("RevivalSystem update started")
	}

	// Find all living player entities (have input and not dead)
	var livingPlayers []*Entity
	for _, entity := range entities {
		if entity.HasComponent("input") && !entity.HasComponent("dead") {
			if healthComp, hasHealth := entity.GetComponent("health"); hasHealth {
				// Type assert with safety check
				if health, ok := healthComp.(*HealthComponent); ok {
					if health.IsAlive() {
						livingPlayers = append(livingPlayers, entity)
					}
				}
			}
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"living_player_count": len(livingPlayers),
		}).Debug("Found living players")
	}

	// Find all dead player entities
	var deadPlayers []*Entity
	for _, entity := range entities {
		if entity.HasComponent("input") && entity.HasComponent("dead") {
			deadPlayers = append(deadPlayers, entity)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"dead_player_count": len(deadPlayers),
		}).Debug("Found dead players")
	}

	// No revival possible if no living or dead players
	if len(livingPlayers) == 0 || len(deadPlayers) == 0 {
		if s.logger != nil && len(entities) > 0 {
			s.logger.WithFields(logrus.Fields{
				"living_players": len(livingPlayers),
				"dead_players":   len(deadPlayers),
			}).Debug("No revival possible - missing living or dead players")
		}
		return
	}

	// Check each living player for revival input
	for _, livingPlayer := range livingPlayers {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": livingPlayer.ID,
			}).Debug("Checking living player for revival input")
		}

		// Check for revival input (E key or interact button)
		inputComp, ok := livingPlayer.GetComponent("input")
		if !ok {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": livingPlayer.ID,
				}).Warn("Living player missing input component")
			}
			continue
		}
		// Type assert with safety check
		input, ok := inputComp.(*EbitenInput)
		if !ok {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":      livingPlayer.ID,
					"component_type": "input",
				}).Warn("Failed to cast input component to EbitenInput")
			}
			continue
		}

		// Check if revival action key is pressed (E key = UseItemPressed)
		// In this context, E key serves dual purpose: use item / interact / revive
		if !input.UseItemPressed {
			continue
		}

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": livingPlayer.ID,
			}).Debug("Living player pressed revival key")
		}

		// Get living player position
		livingPosComp, hasLivingPos := livingPlayer.GetComponent("position")
		if !hasLivingPos {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": livingPlayer.ID,
				}).Warn("Living player missing position component")
			}
			continue
		}
		// Type assert with safety check
		livingPos, ok := livingPosComp.(*PositionComponent)
		if !ok {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":      livingPlayer.ID,
					"component_type": "position",
				}).Warn("Failed to cast position component to PositionComponent")
			}
			continue
		}

		// Find closest dead player within range
		var closestDeadPlayer *Entity
		closestDistance := math.MaxFloat64

		for _, deadPlayer := range deadPlayers {
			// Get dead player position
			deadPosComp, hasDeadPos := deadPlayer.GetComponent("position")
			if !hasDeadPos {
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"entity_id": deadPlayer.ID,
					}).Debug("Dead player missing position component")
				}
				continue
			}
			// Type assert with safety check
			deadPos, ok := deadPosComp.(*PositionComponent)
			if !ok {
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"entity_id":      deadPlayer.ID,
						"component_type": "position",
					}).Warn("Failed to cast position component to PositionComponent")
				}
				continue
			}

			// Calculate distance
			dx := deadPos.X - livingPos.X
			dy := deadPos.Y - livingPos.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"living_entity_id": livingPlayer.ID,
					"dead_entity_id":   deadPlayer.ID,
					"distance":         distance,
					"revival_range":    s.RevivalRange,
				}).Debug("Calculated distance between players")
			}

			// Check if within revival range and closest so far
			if distance <= s.RevivalRange && distance < closestDistance {
				closestDistance = distance
				closestDeadPlayer = deadPlayer
			}
		}

		// Revive the closest dead player if found
		if closestDeadPlayer != nil {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"living_entity_id": livingPlayer.ID,
					"dead_entity_id":   closestDeadPlayer.ID,
					"distance":         closestDistance,
				}).Info("Attempting to revive dead player")
			}
			s.revivePlayer(closestDeadPlayer)
		} else {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"living_entity_id": livingPlayer.ID,
					"revival_range":    s.RevivalRange,
				}).Debug("No dead players within revival range")
			}
		}
	}

	if s.logger != nil {
		s.logger.Debug("RevivalSystem update completed")
	}
}

// revivePlayer performs the actual revival, restoring health and removing dead state.
func (s *RevivalSystem) revivePlayer(deadPlayer *Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": deadPlayer.ID,
		}).Debug("revivePlayer called")
	}

	// Get health component
	healthComp, hasHealth := deadPlayer.GetComponent("health")
	if !hasHealth {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": deadPlayer.ID,
			}).Error("Cannot revive player - missing health component")
		}
		return
	}
	// Type assert with safety check
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      deadPlayer.ID,
				"component_type": "health",
			}).Error("Failed to cast health component to HealthComponent")
		}
		return
	}

	// Restore health (20% of max by default)
	restoredHealth := health.Max * s.RevivalAmount
	previousHealth := health.Current
	health.Current = restoredHealth

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":       deadPlayer.ID,
			"previous_health": previousHealth,
			"restored_health": restoredHealth,
			"max_health":      health.Max,
			"revival_amount":  s.RevivalAmount,
		}).Info("Player health restored")
	}

	// Remove dead component to restore full functionality
	deadPlayer.RemoveComponent("dead")

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":       deadPlayer.ID,
			"restored_health": restoredHealth,
			"max_health":      health.Max,
		}).Info("Player successfully revived")
	}

	// Future enhancement: play revival sound effect, show particles, etc.
	// This would integrate with audio and particle systems
}

// IsPlayerRevivable checks if a specific entity can be revived.
// Returns true if entity is a dead player with health component.
func IsPlayerRevivable(entity *Entity) bool {
	// Must be a player (has input component)
	if !entity.HasComponent("input") {
		return false
	}

	// Must be dead
	if !entity.HasComponent("dead") {
		return false
	}

	// Must have health component to restore health
	if !entity.HasComponent("health") {
		return false
	}

	return true
}

// FindRevivablePlayersInRange finds all dead players within revival range of an entity.
// Useful for UI indicators showing which dead players can be revived.
func FindRevivablePlayersInRange(world *World, livingPlayer *Entity, maxRange float64) []*Entity {
	return FindRevivablePlayersInRangeWithLogger(world, livingPlayer, maxRange, nil)
}

// FindRevivablePlayersInRangeWithLogger finds all dead players within revival range with logging.
func FindRevivablePlayersInRangeWithLogger(world *World, livingPlayer *Entity, maxRange float64, logger *logrus.Logger) []*Entity {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system":    "revival",
			"operation": "find_revivable_players",
		})
	}

	if logEntry != nil {
		logEntry.WithFields(logrus.Fields{
			"living_entity_id": livingPlayer.ID,
			"max_range":        maxRange,
		}).Debug("Finding revivable players in range")
	}

	// Get living player position
	livingPosComp, hasPos := livingPlayer.GetComponent("position")
	if !hasPos {
		if logEntry != nil {
			logEntry.WithFields(logrus.Fields{
				"entity_id": livingPlayer.ID,
			}).Warn("Living player missing position component")
		}
		return nil
	}
	// Type assert with safety check
	livingPos, ok := livingPosComp.(*PositionComponent)
	if !ok {
		if logEntry != nil {
			logEntry.WithFields(logrus.Fields{
				"entity_id":      livingPlayer.ID,
				"component_type": "position",
			}).Warn("Failed to cast position component to PositionComponent")
		}
		return nil
	}

	var revivablePlayers []*Entity

	for _, entity := range world.GetEntities() {
		// Check if entity is revivable
		if !IsPlayerRevivable(entity) {
			continue
		}

		// Get dead player position
		deadPosComp, hasDeadPos := entity.GetComponent("position")
		if !hasDeadPos {
			if logEntry != nil {
				logEntry.WithFields(logrus.Fields{
					"entity_id": entity.ID,
				}).Debug("Dead player missing position component")
			}
			continue
		}
		// Type assert with safety check
		deadPos, ok := deadPosComp.(*PositionComponent)
		if !ok {
			if logEntry != nil {
				logEntry.WithFields(logrus.Fields{
					"entity_id":      entity.ID,
					"component_type": "position",
				}).Warn("Failed to cast position component to PositionComponent")
			}
			continue
		}

		// Calculate distance
		dx := deadPos.X - livingPos.X
		dy := deadPos.Y - livingPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// Add if within range
		if distance <= maxRange {
			revivablePlayers = append(revivablePlayers, entity)
			if logEntry != nil {
				logEntry.WithFields(logrus.Fields{
					"dead_entity_id": entity.ID,
					"distance":       distance,
				}).Debug("Found revivable player in range")
			}
		}
	}

	if logEntry != nil {
		logEntry.WithFields(logrus.Fields{
			"revivable_count": len(revivablePlayers),
		}).Debug("Completed finding revivable players")
	}

	return revivablePlayers
}
