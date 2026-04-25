// Package engine provides the StatusEffectCriticalChanceSystem which bridges
// status effects with entity critical hit chance in combat.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectCriticalChanceSystem applies critical hit chance modifiers based on
// active status effects. This connects StatusEffectSystem with CombatSystem by
// modifying entity CritChance stats during combat calculations.
//
// Supported status effects and their crit chance modifiers:
//   - "blessed": +10% crit chance (divine favor guides strikes)
//   - "cursed": -10% crit chance (misfortune deflects killing blows)
//   - "haste": +5% crit chance (faster reactions find weak points)
//   - "strength": +8% crit chance (powerful strikes hit harder)
//   - "weakness": -8% crit chance (weakened state reduces precision)
//   - "focused": +15% crit chance (concentration improves aim)
//   - "blinded": -15% crit chance (can't aim for vital points)
//   - "enraged": +12% crit chance, -5% defense (reckless aggression)
//   - "frozen": -20% crit chance (sluggish movements)
//   - "chilled": -3% crit chance (cold slows reactions)
//   - "burning": +5% crit chance (desperation increases aggression)
//   - "poisoned": -5% crit chance (weakened state)
type StatusEffectCriticalChanceSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
	genre  string

	// critCache holds the modifier written to stats.CritChance this frame.
	// Maps entityID -> cumulative crit modifier applied this frame.
	critCache map[uint64]float64

	// prevCache holds the modifier that was applied to stats.CritChance last
	// frame. It is subtracted at the start of each frame before the new
	// modifier is added, making the net change always (new - prev). This fixes
	// G33: without prevCache the accumulated additions were never undone.
	prevCache map[uint64]float64
}

// NewStatusEffectCriticalChanceSystem creates a new status effect crit chance system.
func NewStatusEffectCriticalChanceSystem(world *World, seed int64) *StatusEffectCriticalChanceSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_crit_chance")
	}
	return &StatusEffectCriticalChanceSystem{
		world:     world,
		rng:       rand.New(rand.NewSource(seed)),
		logger:    logEntry,
		critCache: make(map[uint64]float64, 64),
		prevCache: make(map[uint64]float64, 64),
	}
}

// SetGenre sets the genre for genre-aware crit modifiers.
func (s *StatusEffectCriticalChanceSystem) SetGenre(genre string) {
	s.genre = genre
}

// Update processes all entities and calculates crit chance modifiers based on
// their active status effects. The modifiers are cached for combat use.
// G33 fix: each frame the previously applied modifier is subtracted from
// stats.CritChance before the new modifier is added, so the net delta is
// always (newModifier - prevModifier) regardless of how many frames pass.
func (s *StatusEffectCriticalChanceSystem) Update(entities []*Entity, deltaTime float64) {
	// Swap caches: critCache becomes prevCache, then clear critCache for this frame.
	s.critCache, s.prevCache = s.prevCache, s.critCache
	for k := range s.critCache {
		delete(s.critCache, k)
	}

	for _, entity := range entities {
		stats := entity.GetStats()
		if stats == nil {
			continue
		}

		modifier := s.calculateCritModifier(entity)

		// Undo the modifier we applied last frame (0.0 if none).
		prev := s.prevCache[entity.ID]
		stats.CritChance -= prev

		if modifier != 0.0 {
			// Cache the modifier for combat system to use
			s.critCache[entity.ID] = modifier

			// Apply new modifier to stats
			oldCrit := stats.CritChance
			stats.CritChance += modifier

			// Clamp crit chance to valid range [0.0, 1.0]
			if stats.CritChance < 0.0 {
				stats.CritChance = 0.0
			} else if stats.CritChance > 1.0 {
				stats.CritChance = 1.0
			}

			s.logCritModification(entity, oldCrit, stats.CritChance, modifier)
		} else if prev != 0.0 {
			// Effect expired: clamp after removal.
			if stats.CritChance < 0.0 {
				stats.CritChance = 0.0
			} else if stats.CritChance > 1.0 {
				stats.CritChance = 1.0
			}
		}
	}
}

// calculateCritModifier computes the cumulative crit modifier for an entity
// based on all active status effects.
func (s *StatusEffectCriticalChanceSystem) calculateCritModifier(entity *Entity) float64 {
	// Check cache first
	if cached, ok := s.critCache[entity.ID]; ok {
		return cached
	}

	modifier := 0.0

	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}

		effectMod := s.effectToCritModifier(effect)
		modifier += effectMod
	}

	// Apply genre-specific scaling
	modifier = s.applyGenreScaling(modifier)

	return modifier
}

// effectToCritModifier converts a status effect to its crit chance modifier.
// Returns 0.0 for effects that don't modify crit chance.
func (s *StatusEffectCriticalChanceSystem) effectToCritModifier(effect *StatusEffectComponent) float64 {
	switch effect.EffectType {
	case "blessed":
		// Blessed: +10% crit chance (divine favor guides strikes)
		return 0.10

	case "cursed":
		// Cursed: -10% crit chance (misfortune deflects killing blows)
		return -0.10

	case "haste":
		// Haste: +5% crit chance (faster reactions find weak points)
		return 0.05

	case "strength":
		// Strength: +8% crit chance (powerful strikes are more precise)
		return 0.08

	case "weakness":
		// Weakness: -8% crit chance (weakened state reduces precision)
		return -0.08

	case "focused":
		// Focused: +15% crit chance (concentration improves targeting)
		return 0.15

	case "blinded", "blindness":
		// Blinded: -15% crit chance (can't aim for vital points)
		return -0.15

	case "enraged":
		// Enraged: +12% crit chance (reckless aggression hits harder)
		return 0.12

	case "frozen":
		// Frozen: -20% crit chance (sluggish movements)
		return -0.20

	case "chilled":
		// Chilled: -3% crit chance (cold slows reactions slightly)
		return -0.03

	case "burning":
		// Burning: +5% crit chance (desperation increases aggression)
		return 0.05

	case "poisoned":
		// Poisoned: -5% crit chance (weakened state)
		return -0.05

	case "fortify":
		// Fortify: No crit effect (defense focused)
		return 0.0

	case "regeneration":
		// Regeneration: +3% crit chance (vitality improves combat focus)
		return 0.03

	case "vulnerability":
		// Vulnerability: No crit effect (affects incoming damage, not outgoing)
		return 0.0

	default:
		return 0.0 // No crit modification
	}
}

// applyGenreScaling adjusts the modifier based on genre preferences.
func (s *StatusEffectCriticalChanceSystem) applyGenreScaling(modifier float64) float64 {
	switch s.genre {
	case "fantasy":
		// Fantasy: Blessed/cursed effects are 20% stronger
		return modifier * 1.0 // Standard fantasy balance

	case "scifi":
		// Sci-fi: Technical precision bonuses are stronger
		return modifier * 1.1 // 10% bonus to all crit modifiers

	case "horror":
		// Horror: Negative effects are stronger, positive weaker
		if modifier < 0 {
			return modifier * 1.3 // 30% stronger debuffs
		}
		return modifier * 0.8 // 20% weaker buffs

	case "cyberpunk":
		// Cyberpunk: All modifiers slightly enhanced (augmentation)
		return modifier * 1.15

	case "postapoc":
		// Post-apocalyptic: Survival instincts boost positive effects
		if modifier > 0 {
			return modifier * 1.2 // 20% stronger buffs
		}
		return modifier * 0.9 // 10% weaker debuffs

	default:
		return modifier
	}
}

// GetCritModifier returns the cached crit modifier for an entity.
// Returns 0.0 if no modifier is cached (no status effects affecting crit).
func (s *StatusEffectCriticalChanceSystem) GetCritModifier(entityID uint64) float64 {
	return s.critCache[entityID]
}

// logCritModification logs crit chance changes for debugging.
func (s *StatusEffectCriticalChanceSystem) logCritModification(entity *Entity, oldCrit, newCrit, modifier float64) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}

	s.logger.WithFields(logrus.Fields{
		"entity_id":      entity.ID,
		"old_crit":       oldCrit,
		"new_crit":       newCrit,
		"modifier":       modifier,
		"genre":          s.genre,
		"clamped":        newCrit != oldCrit+modifier,
		"component_type": "status_effect_crit_chance",
	}).Debug("Crit chance modified by status effects")
}
