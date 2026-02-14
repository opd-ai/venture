// Package engine provides the StatusEffectEvasionSystem which bridges
// status effects with entity evasion chance in combat.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectEvasionSystem applies evasion modifiers based on active status
// effects. This connects the StatusEffectSystem with the CombatSystem by
// modifying entity evasion stats during combat calculations.
//
// Supported status effects and their evasion modifiers:
//   - "chilled": -5% evasion (cold slows reactions)
//   - "frozen": -30% evasion (near-immobile, easy target)
//   - "wet": +5% evasion (slippery, harder to target)
//   - "haste": +15% evasion (faster reactions)
//   - "blinded": -20% evasion (can't see attacks)
//   - "poisoned": -10% evasion (weakened state)
//   - "stunned": -25% evasion (disoriented)
//   - "blessed": +10% evasion (divine protection)
//   - "cursed": -10% evasion (misfortune)
type StatusEffectEvasionSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
	genre  string

	// Cache for evasion modifiers to avoid recalculating each frame.
	// Maps entityID -> cumulative evasion modifier (added to base evasion).
	evasionCache map[uint64]float64
}

// NewStatusEffectEvasionSystem creates a new status effect evasion system.
func NewStatusEffectEvasionSystem(world *World, seed int64) *StatusEffectEvasionSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_evasion")
	}
	return &StatusEffectEvasionSystem{
		world:        world,
		rng:          rand.New(rand.NewSource(seed)),
		logger:       logEntry,
		evasionCache: make(map[uint64]float64, 64),
	}
}

// SetGenre sets the genre for genre-aware evasion modifiers.
func (s *StatusEffectEvasionSystem) SetGenre(genre string) {
	s.genre = genre
}

// Update processes all entities and calculates evasion modifiers based on
// their active status effects. The modifiers are cached for combat use.
func (s *StatusEffectEvasionSystem) Update(entities []*Entity, deltaTime float64) {
	// Clear cache each frame
	for k := range s.evasionCache {
		delete(s.evasionCache, k)
	}

	for _, entity := range entities {
		stats := entity.GetStats()
		if stats == nil {
			continue
		}

		modifier := s.calculateEvasionModifier(entity)
		if modifier == 0.0 {
			continue // No modification needed
		}

		// Cache the modifier for combat system to use
		s.evasionCache[entity.ID] = modifier

		// Apply modifier to stats (additive, not multiplicative)
		oldEvasion := stats.Evasion
		stats.Evasion += modifier

		// Clamp evasion to valid range [0.0, 0.95]
		if stats.Evasion < 0.0 {
			stats.Evasion = 0.0
		} else if stats.Evasion > 0.95 {
			stats.Evasion = 0.95
		}

		s.logEvasionModification(entity, oldEvasion, stats.Evasion, modifier)
	}
}

// calculateEvasionModifier computes the cumulative evasion modifier for an entity
// based on all active status effects.
func (s *StatusEffectEvasionSystem) calculateEvasionModifier(entity *Entity) float64 {
	// Check cache first
	if cached, ok := s.evasionCache[entity.ID]; ok {
		return cached
	}

	modifier := 0.0

	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}

		effectMod := s.effectToEvasionModifier(effect)
		modifier += effectMod
	}

	// Apply genre-specific scaling
	modifier = s.applyGenreScaling(modifier)

	return modifier
}

// effectToEvasionModifier converts a status effect to its evasion modifier.
// Returns 0.0 for effects that don't modify evasion.
func (s *StatusEffectEvasionSystem) effectToEvasionModifier(effect *StatusEffectComponent) float64 {
	switch effect.EffectType {
	case "chilled":
		// Chilled: -5% evasion (cold makes reactions slower)
		return -0.05

	case "frozen":
		// Frozen: -30% evasion (near-immobile, easy target)
		return -0.30

	case "wet":
		// Wet: +5% evasion (slippery surface makes targeting harder)
		return 0.05

	case "haste":
		// Haste: +15% evasion (faster reflexes)
		return 0.15

	case "blinded", "blindness":
		// Blinded: -20% evasion (can't see attacks coming)
		return -0.20

	case "poisoned":
		// Poisoned: -10% evasion (weakened, slower reactions)
		return -0.10

	case "stunned":
		// Stunned: -25% evasion (disoriented, can't dodge)
		return -0.25

	case "blessed":
		// Blessed: +10% evasion (divine favor aids dodging)
		return 0.10

	case "cursed":
		// Cursed: -10% evasion (misfortune makes dodging harder)
		return -0.10

	case "regeneration":
		// Regeneration: +5% evasion (vitality improves reflexes)
		return 0.05

	case "strength":
		// Strength: No evasion effect
		return 0.0

	case "weakness":
		// Weakness: -5% evasion (weakened state)
		return -0.05

	case "fortify":
		// Fortify: No evasion effect (defense focused)
		return 0.0

	case "vulnerability":
		// Vulnerability: -5% evasion (exposed state)
		return -0.05

	default:
		return 0.0 // No evasion modification
	}
}

// applyGenreScaling applies genre-specific scaling to evasion modifiers.
func (s *StatusEffectEvasionSystem) applyGenreScaling(modifier float64) float64 {
	switch s.genre {
	case "scifi":
		// Sci-fi: Enhanced technology reduces status effect impact on evasion
		return modifier * 0.8
	case "horror":
		// Horror: Debuffs are more punishing (fear is stronger)
		if modifier < 0 {
			return modifier * 1.2
		}
		return modifier
	case "cyberpunk":
		// Cyberpunk: Augments provide partial status effect resistance
		return modifier * 0.9
	case "fantasy":
		// Fantasy: Standard magical effects
		return modifier
	case "postapoc":
		// Post-apocalyptic: Harsh conditions amplify negative effects
		if modifier < 0 {
			return modifier * 1.1
		}
		return modifier
	default:
		return modifier
	}
}

// logEvasionModification logs when an evasion modifier is applied.
func (s *StatusEffectEvasionSystem) logEvasionModification(entity *Entity, oldEvasion, newEvasion, modifier float64) {
	if s.logger == nil {
		return
	}
	if s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID":   entity.ID,
		"oldEvasion": oldEvasion,
		"newEvasion": newEvasion,
		"modifier":   modifier,
		"genre":      s.genre,
	}).Debug("applied evasion modifier from status effects")
}

// GetEvasionModifier returns the current evasion modifier for an entity.
// Useful for UI display or debugging. Returns 0.0 if not cached.
func (s *StatusEffectEvasionSystem) GetEvasionModifier(entityID uint64) float64 {
	if mod, ok := s.evasionCache[entityID]; ok {
		return mod
	}
	return 0.0
}

// HasEvasionEffect returns true if the entity has any evasion-affecting status effect.
func (s *StatusEffectEvasionSystem) HasEvasionEffect(entity *Entity) bool {
	if entity == nil {
		return false
	}
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}
		if s.effectToEvasionModifier(effect) != 0.0 {
			return true
		}
	}
	return false
}

// GetEffectiveEvasion returns the effective evasion for an entity after
// applying all status effect modifiers. This is the value combat should use.
func (s *StatusEffectEvasionSystem) GetEffectiveEvasion(entity *Entity) float64 {
	stats := entity.GetStats()
	if stats == nil {
		return 0.0
	}
	return stats.Evasion // Already modified by Update()
}
