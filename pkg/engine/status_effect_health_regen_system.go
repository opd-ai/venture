// Package engine provides the StatusEffectHealthRegenSystem which bridges
// status effects with entity health regeneration rates.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectHealthRegenSystem applies health regeneration modifiers to entities
// based on active status effects. This connects StatusEffectSystem with
// HealthComponent by modifying regeneration rates during gameplay.
//
// Supported status effects and their health regen modifiers:
//   - "regeneration": +50% health regen (active healing)
//   - "blessed": +25% health regen (divine favor)
//   - "empowered": +15% health regen (enhanced vitality)
//   - "strength": +10% health regen (physical resilience)
//   - "haste": +5% health regen (accelerated metabolism)
//   - "poisoned": -30% health regen (weakened recovery)
//   - "burning": -20% health regen (pain disrupts healing)
//   - "cursed": -25% health regen (dark affliction)
//   - "weakness": -35% health regen (severely impaired)
//   - "chilled": -15% health regen (slowed metabolism)
//   - "frozen": -50% health regen (near-stasis state)
//   - "bleeding": -40% health regen (blood loss prevents recovery)
type StatusEffectHealthRegenSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
	genre  string

	// updateInterval controls how often we apply regen (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache for regen modifiers to avoid recalculating each frame
	// Maps entityID -> total regen modifier
	regenModifierCache map[uint64]float64
}

// NewStatusEffectHealthRegenSystem creates a new status effect health regen system.
func NewStatusEffectHealthRegenSystem(world *World, seed int64) *StatusEffectHealthRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_health_regen")
	}
	return &StatusEffectHealthRegenSystem{
		world:              world,
		rng:                rand.New(rand.NewSource(seed)),
		logger:             logEntry,
		updateInterval:     0.5, // Apply regen every 0.5 seconds
		timeSinceCheck:     0,
		regenModifierCache: make(map[uint64]float64, 64),
	}
}

// SetGenre sets the genre for genre-aware regen modifiers.
func (s *StatusEffectHealthRegenSystem) SetGenre(genre string) {
	s.genre = genre
}

// Update processes all entities and applies health regen based on
// their active status effects.
func (s *StatusEffectHealthRegenSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	elapsed := s.timeSinceCheck
	s.timeSinceCheck = 0

	// Clear cache each update cycle
	for k := range s.regenModifierCache {
		delete(s.regenModifierCache, k)
	}

	for _, entity := range entities {
		s.processEntity(entity, elapsed)
	}
}

// processEntity applies health regen based on status effects.
func (s *StatusEffectHealthRegenSystem) processEntity(entity *Entity, elapsed float64) {
	health := entity.GetHealth()
	if health == nil || health.IsDead() {
		return
	}

	// Don't regen if at full health
	if health.Current >= health.Max {
		return
	}

	// Calculate total regen modifier from status effects
	regenModifier := s.calculateRegenModifier(entity)
	s.regenModifierCache[entity.ID] = regenModifier

	// Skip if no modifier or negative total
	if regenModifier == 0.0 {
		return
	}

	// Base regen: 0.5% of max health per second
	baseRate := health.Max * 0.005

	// Apply genre modifier
	genreModifier := s.getGenreModifier()

	// Calculate final regen amount
	finalRate := baseRate * (1.0 + regenModifier) * genreModifier
	regenAmount := finalRate * elapsed

	// Apply regen if positive, or damage if negative (e.g., bleeding)
	if regenAmount > 0 {
		oldHealth := health.Current
		health.Heal(regenAmount)

		if s.logger != nil && health.Current != oldHealth {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"regen_modifier": regenModifier,
				"genre_modifier": genreModifier,
				"regen_amount":   regenAmount,
				"old_health":     oldHealth,
				"new_health":     health.Current,
				"max_health":     health.Max,
			}).Debug("Status effect health regen applied")
		}
	} else if regenAmount < 0 {
		// Negative regen (e.g., from bleeding preventing recovery)
		// This reduces the entity's natural regen capability
		oldHealth := health.Current
		// Apply a small health loss from severe debuffs
		loss := -regenAmount * 0.25 // 25% of negative as actual damage
		health.TakeDamage(loss)

		if s.logger != nil && health.Current != oldHealth {
			s.logger.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"regen_modifier": regenModifier,
				"health_loss":    loss,
				"old_health":     oldHealth,
				"new_health":     health.Current,
			}).Debug("Status effect health drain applied")
		}
	}
}

// calculateRegenModifier computes the total health regen modifier from all status effects.
func (s *StatusEffectHealthRegenSystem) calculateRegenModifier(entity *Entity) float64 {
	var totalModifier float64

	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok {
			continue
		}

		modifier := s.getEffectRegenModifier(effect.EffectType)
		if modifier != 0.0 {
			// Scale by effect magnitude (0.5-2.0 typical)
			magnitudeScale := 1.0
			if effect.Magnitude > 0 {
				magnitudeScale = 0.5 + (effect.Magnitude / 100.0)
				if magnitudeScale > 2.0 {
					magnitudeScale = 2.0
				}
			}
			totalModifier += modifier * magnitudeScale
		}
	}

	return totalModifier
}

// getEffectRegenModifier returns the health regen modifier for a specific effect type.
func (s *StatusEffectHealthRegenSystem) getEffectRegenModifier(effectType string) float64 {
	switch effectType {
	// Positive effects that boost health regen
	case "regeneration":
		return 0.50 // +50% regen
	case "blessed":
		return 0.25 // +25% regen
	case "empowered":
		return 0.15 // +15% regen
	case "strength":
		return 0.10 // +10% regen
	case "haste":
		return 0.05 // +5% regen
	case "shield":
		return 0.10 // +10% regen (protected state)
	case "invulnerable":
		return 0.75 // +75% regen (divine protection)

	// Negative effects that reduce health regen
	case "poisoned":
		return -0.30 // -30% regen
	case "burning":
		return -0.20 // -20% regen
	case "cursed":
		return -0.25 // -25% regen
	case "weakness":
		return -0.35 // -35% regen
	case "chilled":
		return -0.15 // -15% regen
	case "frozen":
		return -0.50 // -50% regen
	case "bleeding":
		return -0.40 // -40% regen
	case "stunned":
		return -0.10 // -10% regen
	case "shocked":
		return -0.15 // -15% regen
	case "feared":
		return -0.20 // -20% regen (panic disrupts recovery)
	case "confused":
		return -0.10 // -10% regen

	default:
		return 0.0
	}
}

// getGenreModifier returns a multiplier based on genre.
func (s *StatusEffectHealthRegenSystem) getGenreModifier() float64 {
	switch s.genre {
	case "fantasy":
		return 1.0 // Standard magical healing
	case "scifi":
		return 1.2 // Advanced medical nanobots
	case "horror":
		return 0.6 // Slow, painful recovery
	case "cyberpunk":
		return 1.15 // Cybernetic enhancement
	case "postapoc":
		return 0.75 // Limited medical supplies
	default:
		return 1.0
	}
}

// GetRegenModifierForEntity returns the cached regen modifier for an entity (for UI).
func (s *StatusEffectHealthRegenSystem) GetRegenModifierForEntity(entityID uint64) float64 {
	if modifier, exists := s.regenModifierCache[entityID]; exists {
		return modifier
	}
	return 0.0
}
