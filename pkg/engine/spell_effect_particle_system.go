// Package engine provides the SpellEffectParticleSystem for visual spell effect feedback.
// This system connects SpellEffectSystem with ParticleSystem to spawn genre-aware particle
// effects when spells are cast, providing visual feedback for different effect types.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// SpellEffectParticleSystem spawns particle effects for active spell effects.
// It processes entities with SpellEffectComponents and spawns appropriate
// particles based on the spell effect type (fire, ice, summoning, etc.).
type SpellEffectParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Track which effects have already spawned particles to avoid duplicates
	spawnedEffects map[effectKey]bool

	// Particle configuration per effect type
	particleCount int
	spreadFactor  float64
}

// effectKey uniquely identifies a spell effect for deduplication.
type effectKey struct {
	entityID   uint64
	effectType EffectType
	startTime  float64 // Duration - ElapsedTime at spawn
}

// NewSpellEffectParticleSystem creates a new spell effect particle system.
func NewSpellEffectParticleSystem(world *World, seed int64) *SpellEffectParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "spell_effect_particle")
		logEntry.Debug("spell effect particle system created")
	}

	return &SpellEffectParticleSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		spawnedEffects: make(map[effectKey]bool),
		particleCount:  15,
		spreadFactor:   100.0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *SpellEffectParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *SpellEffectParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities with spell effects and spawns particles.
func (s *SpellEffectParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	// Clean up old tracked effects periodically (every ~60 frames worth)
	if len(s.spawnedEffects) > 100 {
		s.spawnedEffects = make(map[effectKey]bool)
	}

	for _, entity := range entities {
		for _, comp := range entity.Components {
			effect, ok := comp.(*SpellEffectComponent)
			if !ok || !effect.Active {
				continue
			}

			// Create key for deduplication
			key := effectKey{
				entityID:   entity.ID,
				effectType: effect.EffectType,
				startTime:  effect.Duration,
			}

			// Only spawn particles once per effect (at start)
			if s.spawnedEffects[key] {
				continue
			}

			// Only spawn on first frame of effect
			if effect.ElapsedTime > 0.05 {
				s.spawnedEffects[key] = true
				continue
			}

			s.spawnedEffects[key] = true
			s.spawnEffectParticles(entity, effect)
		}
	}
}

// spawnEffectParticles creates particles appropriate for the spell effect type.
func (s *SpellEffectParticleSystem) spawnEffectParticles(entity *Entity, effect *SpellEffectComponent) {
	var x, y float64

	// Use target position if specified, otherwise entity position
	if effect.TargetX != 0 || effect.TargetY != 0 {
		x, y = effect.TargetX, effect.TargetY
	} else if pos := entity.GetPosition(); pos != nil {
		x, y = pos.X, pos.Y
	} else {
		return
	}

	// Select particle type and config based on spell effect type
	config := s.getParticleConfig(effect)

	// Use deterministic seed offset
	config.Seed = s.seed + int64(x*100) + int64(y*100) + int64(effect.EffectType)*1000

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effect.EffectType.String(),
			"x":           x,
			"y":           y,
		}).Debug("spell effect particles spawned")
	}
}

// getParticleConfig returns a particle configuration appropriate for the effect type.
func (s *SpellEffectParticleSystem) getParticleConfig(effect *SpellEffectComponent) particles.Config {
	config := particles.Config{
		Count:   s.particleCount,
		GenreID: s.genreID,
		SpreadX: s.spreadFactor,
		SpreadY: s.spreadFactor,
		Custom:  make(map[string]interface{}),
	}

	switch effect.EffectType {
	case EffectTerrainManipulation:
		// Earth/rock effects - dust and debris
		config.Type = particles.ParticleDust
		config.Count = 25
		config.Duration = 1.2
		config.Gravity = 150.0
		config.MinSize = 3.0
		config.MaxSize = 8.0
		config.Custom["terrain_effect"] = true

	case EffectTransmutation:
		// Magical transformation - sparkles
		config.Type = particles.ParticleSparkle
		config.Count = 20
		config.Duration = 0.8
		config.Gravity = -40.0
		config.MinSize = 4.0
		config.MaxSize = 7.0
		config.Custom["transmutation"] = true

	case EffectSummoning:
		// Summoning portal - magic particles rising
		config.Type = particles.ParticleMagic
		config.Count = 30
		config.Duration = 1.5
		config.Gravity = -100.0
		config.SpreadX = s.spreadFactor * 1.5
		config.SpreadY = s.spreadFactor * 1.5
		config.MinSize = 5.0
		config.MaxSize = 10.0
		config.Custom["summoning_portal"] = true

	case EffectIllusion:
		// Illusion shimmer - fading sparkles
		config.Type = particles.ParticleSparkle
		config.Count = 15
		config.Duration = 0.6
		config.Gravity = -20.0
		config.SpreadX = s.spreadFactor * 0.8
		config.SpreadY = s.spreadFactor * 0.8
		config.MinSize = 2.0
		config.MaxSize = 5.0
		config.Custom["illusion_shimmer"] = true

	case EffectTimeManipulation:
		// Time distortion - slow spiraling particles
		config.Type = particles.ParticleSparkle
		config.Count = 12
		config.Duration = 2.0
		config.Gravity = 0.0
		config.SpreadX = s.spreadFactor * 0.5
		config.SpreadY = s.spreadFactor * 0.5
		config.MinSize = 3.0
		config.MaxSize = 6.0
		config.Custom["time_distortion"] = true

	case EffectGravityControl:
		// Gravity field - floating debris
		config.Type = particles.ParticleDebris
		config.Count = 18
		config.Duration = 1.8
		config.Gravity = -80.0 // Inverted for levitation
		config.MinSize = 4.0
		config.MaxSize = 9.0
		config.Custom["gravity_field"] = true

	case EffectElementalFusion:
		// Elemental explosion - mixed fire and magic
		config.Type = particles.ParticleFlame
		config.Count = 35
		config.Duration = 1.0
		config.Gravity = -60.0
		config.SpreadX = s.spreadFactor * 1.8
		config.SpreadY = s.spreadFactor * 1.8
		config.MinSize = 5.0
		config.MaxSize = 12.0
		config.Custom["elemental_fusion"] = true

	case EffectLifeDrain:
		// Life drain - blood-like particles flowing
		config.Type = particles.ParticleBlood
		config.Count = 20
		config.Duration = 0.9
		config.Gravity = 50.0
		config.MinSize = 3.0
		config.MaxSize = 6.0
		config.Custom["life_drain"] = true

	case EffectTeleportation:
		// Teleport flash - bright sparks
		config.Type = particles.ParticleSpark
		config.Count = 40
		config.Duration = 0.4
		config.Gravity = 0.0
		config.SpreadX = s.spreadFactor * 2.0
		config.SpreadY = s.spreadFactor * 2.0
		config.MinSize = 4.0
		config.MaxSize = 8.0
		config.Custom["teleport_flash"] = true

	case EffectMetamagic:
		// Meta enhancement - magical aura
		config.Type = particles.ParticleMagic
		config.Count = 25
		config.Duration = 1.2
		config.Gravity = -30.0
		config.MinSize = 4.0
		config.MaxSize = 8.0
		config.Custom["metamagic_aura"] = true

	default:
		// Default magic effect
		config.Type = particles.ParticleMagic
		config.Count = s.particleCount
		config.Duration = 1.0
		config.Gravity = -50.0
		config.MinSize = 4.0
		config.MaxSize = 8.0
	}

	// Scale by effect magnitude
	if effect.Magnitude > 1.0 {
		config.Count = int(float64(config.Count) * (1.0 + (effect.Magnitude-1.0)*0.5))
		if config.Count > 80 {
			config.Count = 80
		}
	}

	// Scale by effect radius for area effects
	if effect.Radius > 0 {
		radiusScale := effect.Radius / 50.0
		if radiusScale > 1.0 {
			config.SpreadX *= radiusScale
			config.SpreadY *= radiusScale
		}
	}

	return config
}

// SpawnEffectParticlesAt allows external systems to spawn spell particles directly.
func (s *SpellEffectParticleSystem) SpawnEffectParticlesAt(effectType EffectType, x, y, magnitude float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	effect := &SpellEffectComponent{
		EffectType: effectType,
		Magnitude:  magnitude,
		TargetX:    x,
		TargetY:    y,
	}

	config := s.getParticleConfig(effect)
	config.Seed = s.seed + int64(x*100) + int64(y*100) + int64(effectType)*1000

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}
