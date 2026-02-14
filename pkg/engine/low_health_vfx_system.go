// Package engine provides the LowHealthVFXSystem for visual low-health warnings.
// This system connects HealthComponent with ParticleSystem to spawn genre-aware
// particle effects when player health drops below a threshold.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// LowHealthVFXSystem spawns particle effects when player health is critically low.
// It monitors player health and provides visual warning feedback with genre-aware
// pulsing particle effects to alert players of dangerous health levels.
type LowHealthVFXSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Health thresholds (0.0-1.0 as percentage of max health)
	criticalThreshold float64 // Below this triggers critical effects
	warningThreshold  float64 // Below this triggers warning effects

	// Effect timing
	pulseInterval   float64 // Seconds between pulse effects
	timeSinceEmit   float64 // Accumulator for pulse timing
	baseCount       int     // Base particle count per pulse
	effectDuration  float64 // How long particles live
	spreadFactor    float64 // Particle spread radius
	activeEntityIDs map[uint64]bool
}

// NewLowHealthVFXSystem creates a new low health visual effects system.
func NewLowHealthVFXSystem(world *World, seed int64) *LowHealthVFXSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "low_health_vfx")
		logEntry.Debug("low health VFX system created")
	}

	return &LowHealthVFXSystem{
		world:             world,
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		criticalThreshold: 0.20, // Below 20% health = critical
		warningThreshold:  0.35, // Below 35% health = warning
		pulseInterval:     0.8,  // Pulse every 0.8 seconds
		timeSinceEmit:     0.0,
		baseCount:         12,
		effectDuration:    0.6,
		spreadFactor:      80.0,
		activeEntityIDs:   make(map[uint64]bool),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *LowHealthVFXSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *LowHealthVFXSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetThresholds allows customizing health thresholds.
// warning should be > critical, both in range [0, 1].
func (s *LowHealthVFXSystem) SetThresholds(warning, critical float64) {
	if warning > 0 && warning <= 1.0 {
		s.warningThreshold = warning
	}
	if critical > 0 && critical <= 1.0 && critical < s.warningThreshold {
		s.criticalThreshold = critical
	}
}

// Update monitors player health and spawns warning particles when low.
func (s *LowHealthVFXSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	if s.timeSinceEmit < s.pulseInterval {
		return
	}

	for _, entity := range entities {
		if !s.isPlayerEntity(entity) {
			continue
		}

		healthRatio := s.getHealthRatio(entity)
		if healthRatio < 0 || healthRatio >= s.warningThreshold {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		// Spawn appropriate particles based on health level
		if healthRatio < s.criticalThreshold {
			s.spawnCriticalEffect(pos.X, pos.Y, entity.ID, healthRatio)
		} else {
			s.spawnWarningEffect(pos.X, pos.Y, entity.ID, healthRatio)
		}
	}

	// Reset timer after processing
	if s.timeSinceEmit >= s.pulseInterval {
		s.timeSinceEmit = 0
	}
}

// isPlayerEntity checks if entity is a player (has input component).
func (s *LowHealthVFXSystem) isPlayerEntity(entity *Entity) bool {
	_, hasInput := entity.GetComponent("input")
	return hasInput
}

// getHealthRatio returns current/max health as 0.0-1.0 ratio.
func (s *LowHealthVFXSystem) getHealthRatio(entity *Entity) float64 {
	health := entity.GetHealth()
	if health == nil || health.Max <= 0 {
		return -1
	}
	return health.Current / health.Max
}

// spawnWarningEffect creates amber/yellow warning particles around player.
func (s *LowHealthVFXSystem) spawnWarningEffect(x, y float64, entityID uint64, healthRatio float64) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	// Fewer particles for warning level
	count := s.baseCount

	config := particles.Config{
		Type:     s.getWarningParticleType(),
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -30.0, // Slight upward float
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   map[string]interface{}{"low_health_warning": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entityID,
			"health_ratio": healthRatio,
			"level":        "warning",
		}).Debug("low health warning particles spawned")
	}
}

// spawnCriticalEffect creates red critical danger particles around player.
func (s *LowHealthVFXSystem) spawnCriticalEffect(x, y float64, entityID uint64, healthRatio float64) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	// More particles for critical level, scaling with how low health is
	intensityScale := 1.0 + (s.criticalThreshold-healthRatio)/s.criticalThreshold
	count := int(float64(s.baseCount) * intensityScale)
	if count > 30 {
		count = 30
	}

	// Primary critical effect
	primaryConfig := particles.Config{
		Type:     s.getCriticalParticleType(),
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration * 0.8,
		SpreadX:  s.spreadFactor * 1.2,
		SpreadY:  s.spreadFactor * 1.2,
		Gravity:  -50.0,
		MinSize:  3.0,
		MaxSize:  7.0,
		Custom:   map[string]interface{}{"low_health_critical": true},
	}

	s.particleSystem.SpawnParticles(s.world, primaryConfig, x, y)

	// Secondary pulse for extra urgency
	secondaryConfig := particles.Config{
		Type:     particles.ParticleSpark,
		Count:    count / 2,
		GenreID:  s.genreID,
		Seed:     effectSeed + 1,
		Duration: s.effectDuration * 0.5,
		SpreadX:  s.spreadFactor * 0.6,
		SpreadY:  s.spreadFactor * 0.6,
		Gravity:  0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   map[string]interface{}{"low_health_critical": true},
	}

	s.particleSystem.SpawnParticles(s.world, secondaryConfig, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entityID,
			"health_ratio": healthRatio,
			"level":        "critical",
			"intensity":    intensityScale,
		}).Debug("low health critical particles spawned")
	}
}

// getWarningParticleType returns genre-appropriate warning particle type.
func (s *LowHealthVFXSystem) getWarningParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleMagic
	case "scifi":
		return particles.ParticleSpark
	case "horror":
		return particles.ParticleSmoke
	case "cyberpunk":
		return particles.ParticleEmber
	case "postapoc":
		return particles.ParticleDust
	default:
		return particles.ParticleSparkle
	}
}

// getCriticalParticleType returns genre-appropriate critical particle type.
func (s *LowHealthVFXSystem) getCriticalParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleFlame
	case "scifi":
		return particles.ParticleEmber
	case "horror":
		return particles.ParticleBlood
	case "cyberpunk":
		return particles.ParticleSpark
	case "postapoc":
		return particles.ParticleEmber
	default:
		return particles.ParticleFlame
	}
}
