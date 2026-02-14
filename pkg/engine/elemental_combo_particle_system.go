// Package engine provides the ElementalComboParticleSystem for visual feedback
// when elemental status effects combine on the same entity.
// This connects StatusEffectComponent elemental effects with ParticleSystem to spawn
// genre-aware combo particles (fire+ice=steam, fire+poison=toxic fumes, etc).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ElementalComboParticleSystem spawns particle effects when elemental status effects
// combine on the same entity. It provides visual feedback for elemental interactions
// with genre-aware particle colors and combo effects.
type ElementalComboParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Cooldown tracking to avoid spamming particles
	comboCooldowns map[uint64]float64 // entityID -> remaining cooldown
	cooldownTime   float64            // Seconds between combo triggers per entity

	// Particle configuration
	baseParticleCount int
	spreadFactor      float64
	effectDuration    float64
}

// ElementalCombo represents a detected elemental combination.
type ElementalCombo struct {
	Primary   string // First element (e.g., "burning")
	Secondary string // Second element (e.g., "frozen")
	ComboName string // Resulting combo name (e.g., "steam_burst")
}

// elementalComboDefinition defines how two elements interact.
type elementalComboDefinition struct {
	element1  string
	element2  string
	comboName string
}

// elementalCombos defines all valid elemental combinations.
var elementalCombos = []elementalComboDefinition{
	{"burning", "frozen", "steam_burst"},
	{"burning", "wet", "steam_burst"},
	{"burning", "poisoned", "toxic_flames"},
	{"frozen", "shocked", "shatter"},
	{"poisoned", "wet", "toxic_pool"},
	{"burning", "shocked", "plasma_burst"},
	{"frozen", "wet", "deep_freeze"},
}

// NewElementalComboParticleSystem creates a new elemental combo particle system.
func NewElementalComboParticleSystem(world *World, seed int64) *ElementalComboParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "elemental_combo_particle")
		logEntry.Debug("elemental combo particle system created")
	}

	return &ElementalComboParticleSystem{
		world:             world,
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		comboCooldowns:    make(map[uint64]float64),
		cooldownTime:      1.0, // 1 second between combo effects per entity
		baseParticleCount: 18,
		spreadFactor:      100.0,
		effectDuration:    0.8,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ElementalComboParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ElementalComboParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities and detects elemental combos to spawn particle effects.
func (s *ElementalComboParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	// Update cooldowns
	s.updateCooldowns(deltaTime)

	for _, entity := range entities {
		// Skip entities on cooldown
		if s.isOnCooldown(entity.ID) {
			continue
		}

		// Detect elemental combos
		combo := s.detectElementalCombo(entity)
		if combo == nil {
			continue
		}

		// Get entity position for particle spawn
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Spawn combo particles
		s.spawnComboParticles(entity.ID, pos.X, pos.Y, combo)

		// Start cooldown
		s.comboCooldowns[entity.ID] = s.cooldownTime
	}
}

// updateCooldowns decrements all cooldowns by deltaTime.
func (s *ElementalComboParticleSystem) updateCooldowns(deltaTime float64) {
	for entityID, remaining := range s.comboCooldowns {
		remaining -= deltaTime
		if remaining <= 0 {
			delete(s.comboCooldowns, entityID)
		} else {
			s.comboCooldowns[entityID] = remaining
		}
	}
}

// isOnCooldown returns true if the entity is on combo cooldown.
func (s *ElementalComboParticleSystem) isOnCooldown(entityID uint64) bool {
	_, exists := s.comboCooldowns[entityID]
	return exists
}

// detectElementalCombo checks if entity has two compatible elemental status effects.
func (s *ElementalComboParticleSystem) detectElementalCombo(entity *Entity) *ElementalCombo {
	// Collect all active elemental effects
	var activeEffects []string

	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok || effect.IsExpired() {
			continue
		}

		// Only track elemental effects
		if s.isElementalEffect(effect.EffectType) {
			activeEffects = append(activeEffects, effect.EffectType)
		}
	}

	// Need at least 2 effects for a combo
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
func (s *ElementalComboParticleSystem) isElementalEffect(effectType string) bool {
	switch effectType {
	case "burning", "frozen", "shocked", "poisoned", "wet", "chilled":
		return true
	default:
		return false
	}
}

// findCombo returns an ElementalCombo if the two effects create one.
func (s *ElementalComboParticleSystem) findCombo(effect1, effect2 string) *ElementalCombo {
	for _, def := range elementalCombos {
		if (def.element1 == effect1 && def.element2 == effect2) ||
			(def.element1 == effect2 && def.element2 == effect1) {
			return &ElementalCombo{
				Primary:   def.element1,
				Secondary: def.element2,
				ComboName: def.comboName,
			}
		}
	}
	return nil
}

// spawnComboParticles creates visual effects for elemental combos.
func (s *ElementalComboParticleSystem) spawnComboParticles(entityID uint64, x, y float64, combo *ElementalCombo) {
	effectSeed := s.seed + int64(entityID*1000) + int64(x) + int64(y)

	// Get particle types based on combo and genre
	primaryType, secondaryType := s.getComboParticleTypes(combo)

	// Primary burst
	primaryConfig := particles.Config{
		Type:     primaryType,
		Count:    s.baseParticleCount,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  s.getComboGravity(combo),
		MinSize:  4.0,
		MaxSize:  10.0,
		Custom: map[string]interface{}{
			"elemental_combo": true,
			"combo_name":      combo.ComboName,
		},
	}
	s.particleSystem.SpawnParticles(s.world, primaryConfig, x, y)

	// Secondary accent particles
	secondaryConfig := particles.Config{
		Type:     secondaryType,
		Count:    s.baseParticleCount / 2,
		GenreID:  s.genreID,
		Seed:     effectSeed + 1,
		Duration: s.effectDuration * 0.7,
		SpreadX:  s.spreadFactor * 0.6,
		SpreadY:  s.spreadFactor * 0.6,
		Gravity:  s.getComboGravity(combo) * 0.5,
		MinSize:  2.0,
		MaxSize:  6.0,
		Custom: map[string]interface{}{
			"elemental_combo": true,
			"combo_accent":    true,
		},
	}
	s.particleSystem.SpawnParticles(s.world, secondaryConfig, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entityID,
			"combo_name": combo.ComboName,
			"primary":    combo.Primary,
			"secondary":  combo.Secondary,
			"x":          x,
			"y":          y,
		}).Debug("elemental combo particles spawned")
	}
}

// getComboParticleTypes returns primary and secondary particle types for combo.
func (s *ElementalComboParticleSystem) getComboParticleTypes(combo *ElementalCombo) (particles.ParticleType, particles.ParticleType) {
	switch combo.ComboName {
	case "steam_burst":
		return s.getSteamBurstTypes()
	case "toxic_flames":
		return s.getToxicFlameTypes()
	case "shatter":
		return s.getShatterTypes()
	case "toxic_pool":
		return s.getToxicPoolTypes()
	case "plasma_burst":
		return s.getPlasmaBurstTypes()
	case "deep_freeze":
		return s.getDeepFreezeTypes()
	default:
		return particles.ParticleMagic, particles.ParticleSparkle
	}
}

// getSteamBurstTypes returns particle types for fire+ice/wet combo.
func (s *ElementalComboParticleSystem) getSteamBurstTypes() (particles.ParticleType, particles.ParticleType) {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSmokePlume, particles.ParticleMagic
	case "scifi":
		return particles.ParticleSmoke, particles.ParticleSpark
	case "horror":
		return particles.ParticleSmoke, particles.ParticleSmokePlume
	case "cyberpunk":
		return particles.ParticleSmoke, particles.ParticleSpark
	case "postapoc":
		return particles.ParticleSmokePlume, particles.ParticleDust
	default:
		return particles.ParticleSmoke, particles.ParticleSparkle
	}
}

// getToxicFlameTypes returns particle types for fire+poison combo.
func (s *ElementalComboParticleSystem) getToxicFlameTypes() (particles.ParticleType, particles.ParticleType) {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleFlame, particles.ParticleSmoke
	case "scifi":
		return particles.ParticleEmber, particles.ParticleSmoke
	case "horror":
		return particles.ParticleFlame, particles.ParticleSmokePlume
	case "cyberpunk":
		return particles.ParticleEmber, particles.ParticleSpark
	case "postapoc":
		return particles.ParticleFlame, particles.ParticleDust
	default:
		return particles.ParticleFlame, particles.ParticleSmoke
	}
}

// getShatterTypes returns particle types for ice+shock combo.
func (s *ElementalComboParticleSystem) getShatterTypes() (particles.ParticleType, particles.ParticleType) {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle, particles.ParticleDebris
	case "scifi":
		return particles.ParticleSpark, particles.ParticleDebris
	case "horror":
		return particles.ParticleDebris, particles.ParticleSpark
	case "cyberpunk":
		return particles.ParticleSpark, particles.ParticleDebris
	case "postapoc":
		return particles.ParticleDebris, particles.ParticleDust
	default:
		return particles.ParticleSparkle, particles.ParticleDebris
	}
}

// getToxicPoolTypes returns particle types for poison+wet combo.
func (s *ElementalComboParticleSystem) getToxicPoolTypes() (particles.ParticleType, particles.ParticleType) {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSmoke, particles.ParticleDust
	case "scifi":
		return particles.ParticleSmoke, particles.ParticleMagic
	case "horror":
		return particles.ParticleSmoke, particles.ParticleBlood
	case "cyberpunk":
		return particles.ParticleSmoke, particles.ParticleSpark
	case "postapoc":
		return particles.ParticleSmoke, particles.ParticleDust
	default:
		return particles.ParticleSmoke, particles.ParticleDust
	}
}

// getPlasmaBurstTypes returns particle types for fire+shock combo.
func (s *ElementalComboParticleSystem) getPlasmaBurstTypes() (particles.ParticleType, particles.ParticleType) {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleMagic, particles.ParticleSpark
	case "scifi":
		return particles.ParticleSpark, particles.ParticleEmber
	case "horror":
		return particles.ParticleEmber, particles.ParticleSpark
	case "cyberpunk":
		return particles.ParticleSpark, particles.ParticleEmber
	case "postapoc":
		return particles.ParticleEmber, particles.ParticleSpark
	default:
		return particles.ParticleSpark, particles.ParticleEmber
	}
}

// getDeepFreezeTypes returns particle types for ice+wet combo.
func (s *ElementalComboParticleSystem) getDeepFreezeTypes() (particles.ParticleType, particles.ParticleType) {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle, particles.ParticleMagic
	case "scifi":
		return particles.ParticleSpark, particles.ParticleSmoke
	case "horror":
		return particles.ParticleSmoke, particles.ParticleSparkle
	case "cyberpunk":
		return particles.ParticleSpark, particles.ParticleSparkle
	case "postapoc":
		return particles.ParticleDust, particles.ParticleSparkle
	default:
		return particles.ParticleSparkle, particles.ParticleMagic
	}
}

// getComboGravity returns appropriate gravity for combo type.
func (s *ElementalComboParticleSystem) getComboGravity(combo *ElementalCombo) float64 {
	switch combo.ComboName {
	case "steam_burst":
		return -120.0 // Steam rises quickly
	case "toxic_flames":
		return -60.0 // Toxic fumes rise
	case "shatter":
		return 150.0 // Debris falls
	case "toxic_pool":
		return 80.0 // Pool effect sinks
	case "plasma_burst":
		return -30.0 // Plasma hovers
	case "deep_freeze":
		return 20.0 // Ice crystals slowly fall
	default:
		return 0.0
	}
}
