// Package engine provides the StatusEffectMovementSystem which bridges
// status effects with entity movement speed.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectMovementSystem applies movement speed modifiers based on active
// status effects. This connects the WeatherCombatSystem and StatusEffectSystem
// with the MovementSystem by scaling entity velocities.
//
// Supported status effects:
//   - "chilled": 15% base slow, scales with magnitude
//   - "frozen": 80% slow (near-immobile)
//   - "wet": 10% slow (slippery ground)
//   - "speed_boost": Increases speed based on magnitude
//   - "haste": 25% speed increase
//   - "slow": Generic slow, scales with magnitude
type StatusEffectMovementSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry

	// Cache for speed modifiers to avoid recalculating each frame.
	// Maps entityID -> cumulative speed multiplier (1.0 = normal).
	speedCache map[uint64]float64
}

// NewStatusEffectMovementSystem creates a new status effect movement system.
func NewStatusEffectMovementSystem(world *World, seed int64) *StatusEffectMovementSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_movement")
	}
	return &StatusEffectMovementSystem{
		world:      world,
		rng:        rand.New(rand.NewSource(seed)),
		logger:     logEntry,
		speedCache: make(map[uint64]float64, 64),
	}
}

// Update processes all entities and applies movement speed modifiers based on
// their active status effects.
func (s *StatusEffectMovementSystem) Update(entities []*Entity, deltaTime float64) {
	// Clear cache each frame
	for k := range s.speedCache {
		delete(s.speedCache, k)
	}

	for _, entity := range entities {
		vel := entity.GetVelocity()
		if vel == nil {
			continue
		}

		multiplier := s.calculateSpeedMultiplier(entity)
		if multiplier == 1.0 {
			continue // No modification needed
		}

		// Apply speed multiplier to velocity
		vel.VX *= multiplier
		vel.VY *= multiplier

		s.logSpeedModification(entity, multiplier)
	}
}

// calculateSpeedMultiplier computes the cumulative speed multiplier for an entity
// based on all active status effects.
func (s *StatusEffectMovementSystem) calculateSpeedMultiplier(entity *Entity) float64 {
	// Check cache first
	if cached, ok := s.speedCache[entity.ID]; ok {
		return cached
	}

	multiplier := 1.0

	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}

		effectMult := s.effectToSpeedMultiplier(effect)
		multiplier *= effectMult
	}

	// Clamp multiplier to reasonable bounds [0.1, 3.0]
	if multiplier < 0.1 {
		multiplier = 0.1
	} else if multiplier > 3.0 {
		multiplier = 3.0
	}

	s.speedCache[entity.ID] = multiplier
	return multiplier
}

// effectToSpeedMultiplier converts a status effect to its speed multiplier.
// Returns 1.0 for effects that don't modify movement speed.
func (s *StatusEffectMovementSystem) effectToSpeedMultiplier(effect *StatusEffectComponent) float64 {
	switch effect.EffectType {
	case "chilled":
		// Chilled: 15-40% slow based on magnitude (0.0-1.0)
		// Magnitude from WeatherCombatSystem scales with intensity
		slowAmount := 0.15 + (effect.Magnitude * 0.25)
		return 1.0 - slowAmount

	case "frozen":
		// Frozen: 80% slow (near-immobile but can still inch along)
		return 0.2

	case "wet":
		// Wet: 10% slow (slippery surfaces)
		return 0.9

	case "speed_boost":
		// Speed boost: Uses magnitude as multiplier directly
		// Typical values: 1.5 (50% faster) to 2.0 (100% faster)
		if effect.Magnitude > 1.0 {
			return effect.Magnitude
		}
		return 1.5 // Default 50% boost

	case "haste":
		// Haste: Fixed 25% speed increase
		return 1.25

	case "slow":
		// Generic slow: Magnitude is the percentage to slow (0.0-1.0)
		slowAmount := effect.Magnitude
		if slowAmount > 0.9 {
			slowAmount = 0.9 // Cap at 90% slow
		}
		return 1.0 - slowAmount

	default:
		return 1.0 // No speed modification
	}
}

// logSpeedModification logs when a speed modifier is applied.
func (s *StatusEffectMovementSystem) logSpeedModification(entity *Entity, multiplier float64) {
	if s.logger == nil {
		return
	}
	if s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID":   entity.ID,
		"multiplier": multiplier,
	}).Debug("applied speed modifier from status effects")
}

// GetSpeedMultiplier returns the current speed multiplier for an entity.
// Useful for UI display or debugging. Returns 1.0 if not cached.
func (s *StatusEffectMovementSystem) GetSpeedMultiplier(entityID uint64) float64 {
	if mult, ok := s.speedCache[entityID]; ok {
		return mult
	}
	return 1.0
}

// HasMovementEffect returns true if the entity has any movement-affecting status effect.
func (s *StatusEffectMovementSystem) HasMovementEffect(entity *Entity) bool {
	if entity == nil {
		return false
	}
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}
		switch effect.EffectType {
		case "chilled", "frozen", "wet", "speed_boost", "haste", "slow":
			return true
		}
	}
	return false
}
