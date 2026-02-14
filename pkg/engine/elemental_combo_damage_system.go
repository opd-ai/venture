// Package engine provides the ElementalComboDamageSystem for applying bonus damage
// when elemental status effects combine on the same entity.
// This connects StatusEffectComponent elemental effects with combat damage
// for genre-aware combo damage (fire+ice=shatter damage, fire+poison=toxic burst, etc).
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ElementalComboDamageSystem applies damage when elemental status effects combine
// on the same entity. It provides gameplay mechanics for elemental interactions
// with genre-aware damage multipliers.
type ElementalComboDamageSystem struct {
	world   *World
	genreID string
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry

	// Cooldown tracking to prevent spamming damage
	comboCooldowns map[uint64]float64 // entityID -> remaining cooldown
	cooldownTime   float64            // Seconds between damage triggers per entity

	// Damage configuration
	baseDamage       float64 // Base damage for combo effects
	genreMultipliers map[string]float64
}

// comboDefinition defines how two elements interact for damage purposes.
type comboDefinition struct {
	element1   string
	element2   string
	comboName  string
	damageType string  // "physical", "fire", "ice", "electric", "poison"
	baseMult   float64 // Base damage multiplier
}

// comboDamageDefinitions defines all valid elemental combinations with damage.
var comboDamageDefinitions = []comboDefinition{
	{"burning", "frozen", "steam_burst", "physical", 1.5},
	{"burning", "wet", "steam_burst", "fire", 1.3},
	{"burning", "poisoned", "toxic_flames", "fire", 1.8},
	{"frozen", "shocked", "shatter", "physical", 2.0},
	{"poisoned", "wet", "toxic_pool", "poison", 1.4},
	{"burning", "shocked", "plasma_burst", "electric", 1.7},
	{"frozen", "wet", "deep_freeze", "ice", 1.6},
}

// NewElementalComboDamageSystem creates a new elemental combo damage system.
func NewElementalComboDamageSystem(world *World, seed int64) *ElementalComboDamageSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "elemental_combo_damage")
		logEntry.Debug("elemental combo damage system created")
	}

	return &ElementalComboDamageSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		comboCooldowns: make(map[uint64]float64),
		cooldownTime:   2.0, // 2 seconds between damage per entity
		baseDamage:     15.0,
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     1.1,
			"horror":    1.3, // Horror: elemental combos are more devastating
			"cyberpunk": 1.2,
			"postapoc":  0.9, // Post-apoc: lower magic effectiveness
		},
	}
}

// SetGenre sets the genre ID for genre-aware damage multipliers.
func (s *ElementalComboDamageSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities and applies combo damage when elemental effects combine.
func (s *ElementalComboDamageSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Update cooldowns
	s.updateCooldowns(deltaTime)

	for _, entity := range entities {
		// Skip entities on cooldown
		if s.isOnCooldown(entity.ID) {
			continue
		}

		// Skip dead entities
		health := entity.GetHealth()
		if health == nil || health.IsDead() {
			continue
		}

		// Detect elemental combos
		combo := s.detectElementalCombo(entity)
		if combo == nil {
			continue
		}

		// Apply combo damage
		s.applyElementalComboDamage(entity, combo)

		// Start cooldown
		s.comboCooldowns[entity.ID] = s.cooldownTime
	}
}

// updateCooldowns decrements all cooldowns by deltaTime.
func (s *ElementalComboDamageSystem) updateCooldowns(deltaTime float64) {
	for entityID, remaining := range s.comboCooldowns {
		remaining -= deltaTime
		if remaining <= 0 {
			delete(s.comboCooldowns, entityID)
		} else {
			s.comboCooldowns[entityID] = remaining
		}
	}
}

// isOnCooldown returns true if the entity is on combo damage cooldown.
func (s *ElementalComboDamageSystem) isOnCooldown(entityID uint64) bool {
	_, exists := s.comboCooldowns[entityID]
	return exists
}

// detectElementalCombo checks if entity has two compatible elemental status effects.
func (s *ElementalComboDamageSystem) detectElementalCombo(entity *Entity) *comboDefinition {
	var activeEffects []string

	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}

		if s.isElementalEffect(effect.EffectType) {
			activeEffects = append(activeEffects, effect.EffectType)
		}
	}

	if len(activeEffects) < 2 {
		return nil
	}

	// Check for valid combos
	for i := 0; i < len(activeEffects); i++ {
		for j := i + 1; j < len(activeEffects); j++ {
			if combo := s.findCombo(activeEffects[i], activeEffects[j]); combo != nil {
				return combo
			}
		}
	}

	return nil
}

// isElementalEffect returns true if the effect type is elemental.
func (s *ElementalComboDamageSystem) isElementalEffect(effectType string) bool {
	switch effectType {
	case "burning", "frozen", "shocked", "poisoned", "wet", "chilled":
		return true
	default:
		return false
	}
}

// findCombo returns a comboDefinition if the two effects create one.
func (s *ElementalComboDamageSystem) findCombo(effect1, effect2 string) *comboDefinition {
	for i := range comboDamageDefinitions {
		def := &comboDamageDefinitions[i]
		if (def.element1 == effect1 && def.element2 == effect2) ||
			(def.element1 == effect2 && def.element2 == effect1) {
			return def
		}
	}
	return nil
}

// applyElementalComboDamage applies damage to an entity from an elemental combo.
func (s *ElementalComboDamageSystem) applyElementalComboDamage(entity *Entity, combo *comboDefinition) {
	health := entity.GetHealth()
	if health == nil {
		return
	}

	// Calculate damage with genre multiplier
	genreMult := s.getGenreMultiplier()
	damage := s.baseDamage * combo.baseMult * genreMult

	// Apply random variance (±15%)
	variance := 0.85 + s.rng.Float64()*0.30
	damage *= variance

	// Apply damage
	health.TakeDamage(damage)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":        entity.ID,
			"combo_name":       combo.comboName,
			"damage":           damage,
			"damage_type":      combo.damageType,
			"genre_multiplier": genreMult,
			"health_remaining": health.Current,
		}).Debug("elemental combo damage applied")
	}
}

// getGenreMultiplier returns the damage multiplier for the current genre.
func (s *ElementalComboDamageSystem) getGenreMultiplier() float64 {
	if mult, ok := s.genreMultipliers[s.genreID]; ok {
		return mult
	}
	return 1.0
}

// GetBaseDamage returns the base damage for elemental combos.
func (s *ElementalComboDamageSystem) GetBaseDamage() float64 {
	return s.baseDamage
}

// SetBaseDamage sets the base damage for elemental combos.
func (s *ElementalComboDamageSystem) SetBaseDamage(damage float64) {
	if damage > 0 {
		s.baseDamage = damage
	}
}

// GetCooldownTime returns the cooldown time between combo damage triggers.
func (s *ElementalComboDamageSystem) GetCooldownTime() float64 {
	return s.cooldownTime
}

// SetCooldownTime sets the cooldown time between combo damage triggers.
func (s *ElementalComboDamageSystem) SetCooldownTime(cooldown float64) {
	if cooldown > 0 {
		s.cooldownTime = cooldown
	}
}
