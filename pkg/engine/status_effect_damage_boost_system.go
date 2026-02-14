// Package engine provides the StatusEffectDamageBoostSystem which bridges
// status effects with entity outgoing damage in combat.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectDamageBoostSystem applies damage modifiers to attacker stats
// based on active status effects. This connects StatusEffectSystem with
// CombatSystem by modifying Attack/MagicPower stats during combat.
//
// Supported status effects and their damage modifiers:
//   - "enraged": +20% damage (fury increases attack power)
//   - "empowered": +15% damage (magical enhancement)
//   - "berserk": +30% damage, -10% defense (reckless power)
//   - "blessed": +10% damage (divine favor)
//   - "cursed": -15% damage (misfortune weakens attacks)
//   - "weakness": -20% damage (drained strength)
//   - "strength": +15% damage (enhanced muscles)
//   - "poisoned": -5% damage (weakened state)
//   - "burning": +5% damage (adrenaline from pain)
//   - "chilled": -10% damage (cold slows strikes)
//   - "frozen": -25% damage (near-immobile, weak attacks)
type StatusEffectDamageBoostSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
	genre  string

	// Cache for damage modifiers to avoid recalculating each frame.
	// Maps entityID -> (attack modifier, magic modifier)
	attackCache map[uint64]float64
	magicCache  map[uint64]float64
}

// NewStatusEffectDamageBoostSystem creates a new status effect damage boost system.
func NewStatusEffectDamageBoostSystem(world *World, seed int64) *StatusEffectDamageBoostSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_damage_boost")
	}
	return &StatusEffectDamageBoostSystem{
		world:       world,
		rng:         rand.New(rand.NewSource(seed)),
		logger:      logEntry,
		attackCache: make(map[uint64]float64, 64),
		magicCache:  make(map[uint64]float64, 64),
	}
}

// SetGenre sets the genre for genre-aware damage modifiers.
func (s *StatusEffectDamageBoostSystem) SetGenre(genre string) {
	s.genre = genre
}

// Update processes all entities and applies damage modifiers based on
// their active status effects. The modifiers scale Attack and MagicPower.
func (s *StatusEffectDamageBoostSystem) Update(entities []*Entity, deltaTime float64) {
	// Clear cache each frame
	for k := range s.attackCache {
		delete(s.attackCache, k)
	}
	for k := range s.magicCache {
		delete(s.magicCache, k)
	}

	for _, entity := range entities {
		stats := entity.GetStats()
		if stats == nil {
			continue
		}

		attackMod, magicMod := s.calculateDamageModifiers(entity)
		if attackMod == 0.0 && magicMod == 0.0 {
			continue // No modification needed
		}

		// Cache the modifiers
		s.attackCache[entity.ID] = attackMod
		s.magicCache[entity.ID] = magicMod

		// Apply modifiers (multiplicative percentage)
		oldAttack := stats.Attack
		oldMagic := stats.MagicPower

		stats.Attack *= (1.0 + attackMod)
		stats.MagicPower *= (1.0 + magicMod)

		// Clamp to prevent negative damage
		if stats.Attack < 0 {
			stats.Attack = 0
		}
		if stats.MagicPower < 0 {
			stats.MagicPower = 0
		}

		s.logDamageModification(entity, oldAttack, oldMagic, stats.Attack, stats.MagicPower)
	}
}

// calculateDamageModifiers computes the cumulative damage modifiers for an entity
// based on all active status effects. Returns (attackMod, magicMod) as percentages.
func (s *StatusEffectDamageBoostSystem) calculateDamageModifiers(entity *Entity) (attackMod, magicMod float64) {
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}

		aMod, mMod := s.effectToDamageModifier(effect)
		attackMod += aMod
		magicMod += mMod
	}

	// Apply genre-specific scaling
	attackMod = s.applyGenreScaling(attackMod)
	magicMod = s.applyGenreScaling(magicMod)

	return attackMod, magicMod
}

// effectToDamageModifier converts a status effect to damage modifiers.
// Returns (attackMod, magicMod) as percentage values (e.g., 0.20 = +20%).
func (s *StatusEffectDamageBoostSystem) effectToDamageModifier(effect *StatusEffectComponent) (attackMod, magicMod float64) {
	switch effect.EffectType {
	case "enraged":
		// Enraged: +20% physical damage (fury)
		return 0.20, 0.05

	case "empowered":
		// Empowered: +15% magical damage (arcane enhancement)
		return 0.05, 0.15

	case "berserk":
		// Berserk: +30% physical damage (reckless power)
		return 0.30, 0.10

	case "blessed":
		// Blessed: +10% all damage (divine favor)
		return 0.10, 0.10

	case "cursed":
		// Cursed: -15% all damage (misfortune)
		return -0.15, -0.15

	case "weakness":
		// Weakness: -20% physical damage
		return -0.20, -0.10

	case "strength":
		// Strength: +15% physical damage
		return 0.15, 0.0

	case "poisoned":
		// Poisoned: -5% damage (weakened state)
		return -0.05, -0.05

	case "burning":
		// Burning: +5% physical damage (adrenaline from pain)
		return 0.05, 0.0

	case "chilled":
		// Chilled: -10% damage (cold slows strikes)
		return -0.10, -0.10

	case "frozen":
		// Frozen: -25% damage (near-immobile)
		return -0.25, -0.25

	case "haste":
		// Haste: +10% attack speed translates to effective damage
		return 0.10, 0.10

	case "fortify":
		// Fortify: Defensive, no damage bonus
		return 0.0, 0.0

	case "vulnerability":
		// Vulnerability: Target takes more damage, no attacker bonus
		return 0.0, 0.0

	case "regeneration":
		// Regeneration: No damage effect
		return 0.0, 0.0

	default:
		return 0.0, 0.0
	}
}

// applyGenreScaling applies genre-specific scaling to damage modifiers.
func (s *StatusEffectDamageBoostSystem) applyGenreScaling(modifier float64) float64 {
	switch s.genre {
	case "fantasy":
		// Fantasy: Standard magical effects
		return modifier
	case "scifi":
		// Sci-fi: Tech buffs are slightly stronger
		if modifier > 0 {
			return modifier * 1.1
		}
		return modifier * 0.9
	case "horror":
		// Horror: Debuffs are more punishing
		if modifier < 0 {
			return modifier * 1.25
		}
		return modifier
	case "cyberpunk":
		// Cyberpunk: Augmented damage buffs are stronger
		if modifier > 0 {
			return modifier * 1.15
		}
		return modifier * 0.85
	case "postapoc":
		// Post-apocalyptic: Survival adrenaline boosts buffs
		if modifier > 0 {
			return modifier * 1.1
		}
		return modifier * 1.1
	default:
		return modifier
	}
}

// logDamageModification logs when damage modifiers are applied.
func (s *StatusEffectDamageBoostSystem) logDamageModification(entity *Entity, oldAttack, oldMagic, newAttack, newMagic float64) {
	if s.logger == nil {
		return
	}
	if s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID":       entity.ID,
		"oldAttack":      oldAttack,
		"newAttack":      newAttack,
		"oldMagicPower":  oldMagic,
		"newMagicPower":  newMagic,
		"attackModifier": s.attackCache[entity.ID],
		"magicModifier":  s.magicCache[entity.ID],
		"genre":          s.genre,
	}).Debug("applied damage modifier from status effects")
}

// GetDamageModifiers returns the current damage modifiers for an entity.
// Useful for UI display or debugging.
func (s *StatusEffectDamageBoostSystem) GetDamageModifiers(entityID uint64) (attackMod, magicMod float64) {
	attackMod = s.attackCache[entityID]
	magicMod = s.magicCache[entityID]
	return attackMod, magicMod
}

// HasDamageEffect returns true if the entity has any damage-affecting status effect.
func (s *StatusEffectDamageBoostSystem) HasDamageEffect(entity *Entity) bool {
	if entity == nil {
		return false
	}
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}
		aMod, mMod := s.effectToDamageModifier(effect)
		if aMod != 0.0 || mMod != 0.0 {
			return true
		}
	}
	return false
}

// GetEffectiveDamage returns the effective attack power for an entity after
// applying all status effect modifiers. This is the value combat should use.
func (s *StatusEffectDamageBoostSystem) GetEffectiveDamage(entity *Entity) (attack, magic float64) {
	stats := entity.GetStats()
	if stats == nil {
		return 0.0, 0.0
	}
	return stats.Attack, stats.MagicPower // Already modified by Update()
}
