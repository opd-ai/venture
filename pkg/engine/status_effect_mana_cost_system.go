// Package engine provides the StatusEffectManaCostSystem which bridges status
// effects with spell mana cost modifications. This system modifies how much mana
// spells cost based on active status effects, connecting StatusEffectSystem with
// SpellCastingSystem for tactical spell management.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// StatusEffectManaCostSystem modifies spell mana costs based on active status
// effects. This connects the StatusEffectSystem with the SpellCastingSystem by
// applying mana cost multipliers during spell casting calculations.
//
// Supported status effects and their mana cost modifiers:
//   - "haste": -15% mana cost (efficient casting)
//   - "focused": -20% mana cost (enhanced concentration)
//   - "blessed": -10% mana cost (divine efficiency)
//   - "regeneration": -5% mana cost (vitality overflow)
//   - "chilled": +15% mana cost (sluggish spellcasting)
//   - "poisoned": +10% mana cost (disrupted concentration)
//   - "cursed": +20% mana cost (magical interference)
//   - "weakness": +10% mana cost (reduced magical stamina)
//   - "stunned": +25% mana cost (disoriented casting)
//   - "burning": +5% mana cost (painful distraction)
type StatusEffectManaCostSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
	genre  string

	// Cache for mana cost multipliers to avoid recalculating each frame.
	// Maps entityID -> cumulative mana cost multiplier (1.0 = no change).
	costMultiplierCache map[uint64]float64
}

// NewStatusEffectManaCostSystem creates a new status effect mana cost system.
func NewStatusEffectManaCostSystem(world *World, seed int64) *StatusEffectManaCostSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_mana_cost")
		logEntry.Debug("StatusEffectManaCostSystem created")
	}
	return &StatusEffectManaCostSystem{
		world:               world,
		rng:                 rand.New(rand.NewSource(seed)),
		logger:              logEntry,
		costMultiplierCache: make(map[uint64]float64, 64),
		genre:               "fantasy",
	}
}

// SetGenre sets the genre for genre-aware mana cost modifiers.
func (s *StatusEffectManaCostSystem) SetGenre(genre string) {
	s.genre = genre
	if s.logger != nil {
		s.logger.WithField("genre", genre).Debug("Genre set for mana cost modifiers")
	}
}

// Update processes all entities and calculates mana cost multipliers based on
// their active status effects. The multipliers are cached for spell casting use.
func (s *StatusEffectManaCostSystem) Update(entities []*Entity, deltaTime float64) {
	// Clear cache each frame
	for k := range s.costMultiplierCache {
		delete(s.costMultiplierCache, k)
	}

	for _, entity := range entities {
		// Only process entities that can cast spells
		if !entity.HasComponent("mana") && !entity.HasComponent("spell_slots") {
			continue
		}

		multiplier := s.calculateCostMultiplier(entity)
		if multiplier == 1.0 {
			continue // No modification needed
		}

		// Cache the multiplier for spell casting system to use
		s.costMultiplierCache[entity.ID] = multiplier

		s.logCostModification(entity, multiplier)
	}
}

// calculateCostMultiplier computes the cumulative mana cost multiplier for an entity
// based on all active status effects. Returns 1.0 for no modification.
func (s *StatusEffectManaCostSystem) calculateCostMultiplier(entity *Entity) float64 {
	multiplier := 1.0

	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}

		effectMod := s.effectToCostModifier(effect)
		// Apply multiplicatively for compounding effects
		multiplier *= effectMod
	}

	// Apply genre-specific scaling
	multiplier = s.applyGenreScaling(multiplier)

	// Clamp to reasonable bounds [0.4, 2.0] (60% discount to 100% cost increase)
	if multiplier < 0.4 {
		multiplier = 0.4
	} else if multiplier > 2.0 {
		multiplier = 2.0
	}

	return multiplier
}

// effectToCostModifier converts a status effect to its mana cost multiplier.
// Returns 1.0 for effects that don't modify mana cost.
func (s *StatusEffectManaCostSystem) effectToCostModifier(effect *StatusEffectComponent) float64 {
	switch effect.EffectType {
	case "haste":
		// Haste: -15% mana cost (faster, more efficient casting)
		return 0.85

	case "focused", "focus":
		// Focused: -20% mana cost (enhanced magical concentration)
		return 0.80

	case "blessed":
		// Blessed: -10% mana cost (divine efficiency)
		return 0.90

	case "regeneration":
		// Regeneration: -5% mana cost (vitality overflow aids casting)
		return 0.95

	case "empowered":
		// Empowered: -10% mana cost (magical surge)
		return 0.90

	case "chilled":
		// Chilled: +15% mana cost (cold disrupts magical flow)
		return 1.15

	case "frozen":
		// Frozen: +30% mana cost (severe magical disruption)
		return 1.30

	case "poisoned":
		// Poisoned: +10% mana cost (toxins interfere with concentration)
		return 1.10

	case "cursed":
		// Cursed: +20% mana cost (magical interference from curse)
		return 1.20

	case "weakness":
		// Weakness: +10% mana cost (reduced magical stamina)
		return 1.10

	case "stunned":
		// Stunned: +25% mana cost (disoriented, inefficient casting)
		return 1.25

	case "burning":
		// Burning: +5% mana cost (painful distraction)
		return 1.05

	case "blinded", "blindness":
		// Blinded: +10% mana cost (harder to focus)
		return 1.10

	case "wet":
		// Wet: +5% mana cost for fire spells, -5% for water (average to neutral)
		return 1.0

	case "vulnerability":
		// Vulnerability: +5% mana cost (weakened magical defense)
		return 1.05

	case "fortify":
		// Fortify: No mana cost effect (defense focused)
		return 1.0

	case "strength":
		// Strength: No mana cost effect (physical focused)
		return 1.0

	default:
		return 1.0 // No mana cost modification
	}
}

// applyGenreScaling applies genre-specific scaling to mana cost multipliers.
func (s *StatusEffectManaCostSystem) applyGenreScaling(multiplier float64) float64 {
	// Calculate how far from 1.0 the multiplier is
	deviation := multiplier - 1.0

	switch s.genre {
	case "fantasy":
		// Fantasy: Standard magical effects, slightly stronger buffs
		if deviation < 0 {
			deviation *= 1.1 // 10% stronger cost reductions
		}
	case "scifi":
		// Sci-fi: Tech assists magic, stronger cost reductions
		if deviation < 0 {
			deviation *= 1.2 // 20% stronger cost reductions
		}
	case "horror":
		// Horror: Dark magic penalties are harsher
		if deviation > 0 {
			deviation *= 1.25 // 25% stronger cost increases
		}
	case "cyberpunk":
		// Cyberpunk: Neural implants help with focus
		if deviation < 0 {
			deviation *= 1.15 // 15% stronger cost reductions
		}
	case "postapoc":
		// Post-apocalyptic: Scarce resources, all costs slightly higher
		deviation *= 1.1 // 10% stronger effect in either direction
	}

	return 1.0 + deviation
}

// logCostModification logs when a mana cost modifier is calculated.
func (s *StatusEffectManaCostSystem) logCostModification(entity *Entity, multiplier float64) {
	if s.logger == nil {
		return
	}
	if s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}
	s.logger.WithFields(logrus.Fields{
		"entityID":   entity.ID,
		"multiplier": multiplier,
		"genre":      s.genre,
	}).Debug("calculated mana cost modifier from status effects")
}

// GetCostMultiplier returns the current mana cost multiplier for an entity.
// Returns 1.0 if no modifier is cached (no status effects affecting cost).
func (s *StatusEffectManaCostSystem) GetCostMultiplier(entityID uint64) float64 {
	if mult, ok := s.costMultiplierCache[entityID]; ok {
		return mult
	}
	return 1.0
}

// HasCostEffect returns true if the entity has any mana-cost-affecting status effect.
func (s *StatusEffectManaCostSystem) HasCostEffect(entity *Entity) bool {
	if entity == nil {
		return false
	}
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}
		if s.effectToCostModifier(effect) != 1.0 {
			return true
		}
	}
	return false
}

// GetEffectiveManaCost calculates the effective mana cost for a spell,
// applying the entity's status effect modifiers. Returns the modified cost.
func (s *StatusEffectManaCostSystem) GetEffectiveManaCost(entityID uint64, baseCost int) int {
	multiplier := s.GetCostMultiplier(entityID)
	effectiveCost := float64(baseCost) * multiplier

	// Round to nearest integer, minimum 1 mana
	result := int(effectiveCost + 0.5)
	if result < 1 {
		result = 1
	}

	return result
}
