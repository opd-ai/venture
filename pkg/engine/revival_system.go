// Package engine provides the revival system for multiplayer death mechanics.
// This file implements RevivalSystem which allows living teammates to revive
// dead players through proximity interaction.
// Priority 1.5: Multiplayer Revival System
package engine

import "math"

// RevivalSystem handles player revival mechanics in multiplayer.
// Living players can revive dead teammates by standing nearby and pressing
// the revival input key. This system implements the PLAN.md Priority 1.5
// revival mechanics with proximity detection and health restoration.
type RevivalSystem struct {
	world *World

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
	return &RevivalSystem{
		world:         world,
		RevivalRange:  32.0, // One tile range
		RevivalAmount: 0.2,  // 20% health restoration
		RevivalTime:   0.0,  // Instant revival (no channeling)
	}
}

// Update processes revival inputs and handles revival logic.
// Checks for living players pressing revival key near dead players.
func (s *RevivalSystem) Update(entities []*Entity, deltaTime float64) {
	// Find all living player entities (have input and not dead)
	var livingPlayers []*Entity
	for _, entity := range entities {
		if entity.HasComponent("input") && !entity.HasComponent("dead") {
			if healthComp, hasHealth := entity.GetComponent("health"); hasHealth {
				health := healthComp.(*HealthComponent)
				if health.IsAlive() {
					livingPlayers = append(livingPlayers, entity)
				}
			}
		}
	}

	// Find all dead player entities
	var deadPlayers []*Entity
	for _, entity := range entities {
		if entity.HasComponent("input") && entity.HasComponent("dead") {
			deadPlayers = append(deadPlayers, entity)
		}
	}

	// No revival possible if no living or dead players
	if len(livingPlayers) == 0 || len(deadPlayers) == 0 {
		return
	}

	// Check each living player for revival input
	for _, livingPlayer := range livingPlayers {
		// Check for revival input (E key or interact button)
		inputComp, _ := livingPlayer.GetComponent("input")
		input := inputComp.(*EbitenInput)

		// Check if revival action key is pressed (E key = UseItemPressed)
		// In this context, E key serves dual purpose: use item / interact / revive
		if !input.UseItemPressed {
			continue
		}

		// Get living player position
		livingPosComp, hasLivingPos := livingPlayer.GetComponent("position")
		if !hasLivingPos {
			continue
		}
		livingPos := livingPosComp.(*PositionComponent)

		// Find closest dead player within range
		var closestDeadPlayer *Entity
		closestDistance := math.MaxFloat64

		for _, deadPlayer := range deadPlayers {
			// Get dead player position
			deadPosComp, hasDeadPos := deadPlayer.GetComponent("position")
			if !hasDeadPos {
				continue
			}
			deadPos := deadPosComp.(*PositionComponent)

			// Calculate distance
			dx := deadPos.X - livingPos.X
			dy := deadPos.Y - livingPos.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			// Check if within revival range and closest so far
			if distance <= s.RevivalRange && distance < closestDistance {
				closestDistance = distance
				closestDeadPlayer = deadPlayer
			}
		}

		// Revive the closest dead player if found
		if closestDeadPlayer != nil {
			s.revivePlayer(closestDeadPlayer)
		}
	}
}

// revivePlayer performs the actual revival, restoring health and removing dead state.
func (s *RevivalSystem) revivePlayer(deadPlayer *Entity) {
	// Get health component
	healthComp, hasHealth := deadPlayer.GetComponent("health")
	if !hasHealth {
		return
	}
	health := healthComp.(*HealthComponent)

	// Restore health (20% of max by default)
	restoredHealth := health.Max * s.RevivalAmount
	health.Current = restoredHealth

	// Remove dead component to restore full functionality
	deadPlayer.RemoveComponent("dead")

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
	// Get living player position
	livingPosComp, hasPos := livingPlayer.GetComponent("position")
	if !hasPos {
		return nil
	}
	livingPos := livingPosComp.(*PositionComponent)

	var revivablePlayers []*Entity

	for _, entity := range world.GetEntities() {
		// Check if entity is revivable
		if !IsPlayerRevivable(entity) {
			continue
		}

		// Get dead player position
		deadPosComp, hasDeadPos := entity.GetComponent("position")
		if !hasDeadPos {
			continue
		}
		deadPos := deadPosComp.(*PositionComponent)

		// Calculate distance
		dx := deadPos.X - livingPos.X
		dy := deadPos.Y - livingPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// Add if within range
		if distance <= maxRange {
			revivablePlayers = append(revivablePlayers, entity)
		}
	}

	return revivablePlayers
}
