// Package engine provides the HealthRegenPulseSystem for visual healing feedback.
// When an entity's health increases (from regen, potions, or heals), this system
// emits genre-aware upward-drifting particles around the entity, giving players
// immediate visual confirmation that healing is occurring.
// Genre styles: fantasy=green sparkles, scifi=cyan magic, horror=red embers,
// cyberpunk=magenta sparks, postapoc=amber dust.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// HealthRegenPulseComponent tracks per-entity previous health for delta detection.
type HealthRegenPulseComponent struct {
	PrevHealth  float64 // Health value last frame
	PulseTimer  float64 // Cooldown between pulse emissions
	Accumulator float64 // Accumulated heal amount for intensity scaling
	Initialized bool    // Whether PrevHealth has been set at least once
}

// Type returns the component type identifier.
func (c *HealthRegenPulseComponent) Type() string { return "health_regen_pulse" }

// healthRegenGenrePreset holds genre-specific healing visual parameters.
type healthRegenGenrePreset struct {
	particleType particles.ParticleType
	minSize      float64
	maxSize      float64
	duration     float64
	gravity      float64
	baseCount    int
}

// HealthRegenPulseSystem emits genre-aware particles when entity health increases.
type HealthRegenPulseSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry
	preset         healthRegenGenrePreset

	// Tuning
	pulseInterval float64 // Minimum seconds between pulses per entity
	spreadRadius  float64 // Particle spread radius around entity
}

// NewHealthRegenPulseSystem creates a new health regeneration pulse visual system.
func NewHealthRegenPulseSystem(world *World, seed int64) *HealthRegenPulseSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "health_regen_pulse")
		logEntry.Debug("health regen pulse system created")
	}

	s := &HealthRegenPulseSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		pulseInterval: 0.6,
		spreadRadius:  24.0,
	}
	s.applyGenrePreset("fantasy") // default
	return s
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *HealthRegenPulseSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetGenre sets the genre ID and updates visual preset.
func (s *HealthRegenPulseSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenrePreset(genreID)
}

func (s *HealthRegenPulseSystem) applyGenrePreset(genreID string) {
	switch genreID {
	case "scifi":
		s.preset = healthRegenGenrePreset{
			particleType: particles.ParticleMagic,
			minSize:      2.0, maxSize: 5.0,
			duration: 0.7, gravity: -40.0, baseCount: 6,
		}
	case "horror":
		s.preset = healthRegenGenrePreset{
			particleType: particles.ParticleEmber,
			minSize:      2.0, maxSize: 4.0,
			duration: 0.5, gravity: -20.0, baseCount: 5,
		}
	case "cyberpunk":
		s.preset = healthRegenGenrePreset{
			particleType: particles.ParticleSpark,
			minSize:      2.0, maxSize: 4.0,
			duration: 0.5, gravity: -35.0, baseCount: 6,
		}
	case "postapoc":
		s.preset = healthRegenGenrePreset{
			particleType: particles.ParticleDust,
			minSize:      3.0, maxSize: 6.0,
			duration: 0.8, gravity: -15.0, baseCount: 4,
		}
	default: // fantasy
		s.preset = healthRegenGenrePreset{
			particleType: particles.ParticleSparkle,
			minSize:      2.0, maxSize: 5.0,
			duration: 0.8, gravity: -30.0, baseCount: 8,
		}
	}
}

// Update checks all entities for health increases and emits healing particles.
func (s *HealthRegenPulseSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		health := entity.GetHealth()
		if health == nil || health.Max <= 0 {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		comp := s.ensureComponent(entity)

		// Tick cooldown
		if comp.PulseTimer > 0 {
			comp.PulseTimer -= deltaTime
		}

		if !comp.Initialized {
			comp.PrevHealth = health.Current
			comp.Initialized = true
			continue
		}

		// Detect health increase
		delta := health.Current - comp.PrevHealth
		if delta > 0 {
			comp.Accumulator += delta
		}

		comp.PrevHealth = health.Current

		// Emit pulse when accumulated heal is meaningful and cooldown elapsed
		healRatio := comp.Accumulator / health.Max
		if healRatio >= 0.02 && comp.PulseTimer <= 0 {
			s.spawnHealPulse(pos.X, pos.Y, entity.ID, healRatio)
			comp.Accumulator = 0
			comp.PulseTimer = s.pulseInterval
		}
	}
}

// ensureComponent lazily attaches a HealthRegenPulseComponent to the entity.
func (s *HealthRegenPulseSystem) ensureComponent(entity *Entity) *HealthRegenPulseComponent {
	if comp, ok := entity.GetComponent("health_regen_pulse"); ok {
		return comp.(*HealthRegenPulseComponent)
	}
	comp := &HealthRegenPulseComponent{}
	entity.AddComponent(comp)
	return comp
}

// spawnHealPulse creates genre-aware upward-drifting healing particles.
func (s *HealthRegenPulseSystem) spawnHealPulse(x, y float64, entityID uint64, healRatio float64) {
	// Scale particle count by heal magnitude (clamped 1x-2x base)
	intensityMul := 1.0
	if healRatio > 0.1 {
		intensityMul = 1.5
	}
	if healRatio > 0.25 {
		intensityMul = 2.0
	}
	count := int(float64(s.preset.baseCount) * intensityMul)
	if count < 2 {
		count = 2
	}
	if count > 20 {
		count = 20
	}

	effectSeed := s.seed + int64(entityID)*31 + int64(x*7) + int64(y*13)

	config := particles.Config{
		Type:     s.preset.particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.preset.duration,
		SpreadX:  s.spreadRadius,
		SpreadY:  s.spreadRadius * 0.6, // Slightly taller than wide
		Gravity:  s.preset.gravity,
		MinSize:  s.preset.minSize,
		MaxSize:  s.preset.maxSize,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"health_regen_pulse": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}
