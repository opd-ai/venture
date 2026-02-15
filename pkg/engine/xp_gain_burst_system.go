// Package engine provides the XPGainBurstSystem for visual XP gain feedback.
// When an entity's experience points increase, this system emits genre-aware
// upward-rising particle bursts around the entity, giving players immediate
// visual confirmation of experience gain. Burst intensity scales with XP amount.
// Genre styles: fantasy=golden sparkles, scifi=cyan data pulses, horror=green wisps,
// cyberpunk=neon sparks, postapoc=amber dust.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// XPGainBurstComponent tracks per-entity previous XP for delta detection.
type XPGainBurstComponent struct {
	PrevTotalXP int     // TotalXP value last frame
	BurstTimer  float64 // Cooldown between burst emissions
	Initialized bool    // Whether PrevTotalXP has been set at least once
}

// Type returns the component type identifier.
func (c *XPGainBurstComponent) Type() string { return "xp_gain_burst" }

// xpGainGenrePreset holds genre-specific XP gain visual parameters.
type xpGainGenrePreset struct {
	particleType particles.ParticleType
	minSize      float64
	maxSize      float64
	duration     float64
	gravity      float64
	baseCount    int
}

// XPGainBurstSystem emits genre-aware particles when entity XP increases.
type XPGainBurstSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry
	preset         xpGainGenrePreset

	burstInterval float64 // Minimum seconds between bursts per entity
	spreadRadius  float64 // Particle spread radius around entity
}

// NewXPGainBurstSystem creates a new XP gain burst visual system.
func NewXPGainBurstSystem(world *World, seed int64) *XPGainBurstSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "xp_gain_burst")
		logEntry.Debug("xp gain burst system created")
	}

	s := &XPGainBurstSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		burstInterval: 0.4,
		spreadRadius:  20.0,
	}
	s.applyGenrePreset("fantasy") // default
	return s
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *XPGainBurstSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetGenre sets the genre ID and updates visual preset.
func (s *XPGainBurstSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenrePreset(genreID)
}

func (s *XPGainBurstSystem) applyGenrePreset(genreID string) {
	switch genreID {
	case "scifi":
		s.preset = xpGainGenrePreset{
			particleType: particles.ParticleMagic,
			minSize:      2.0, maxSize: 5.0,
			duration: 0.6, gravity: -50.0, baseCount: 6,
		}
	case "horror":
		s.preset = xpGainGenrePreset{
			particleType: particles.ParticleEmber,
			minSize:      2.0, maxSize: 4.0,
			duration: 0.5, gravity: -25.0, baseCount: 5,
		}
	case "cyberpunk":
		s.preset = xpGainGenrePreset{
			particleType: particles.ParticleSpark,
			minSize:      2.0, maxSize: 4.0,
			duration: 0.5, gravity: -45.0, baseCount: 7,
		}
	case "postapoc":
		s.preset = xpGainGenrePreset{
			particleType: particles.ParticleDust,
			minSize:      3.0, maxSize: 6.0,
			duration: 0.7, gravity: -15.0, baseCount: 4,
		}
	default: // fantasy
		s.preset = xpGainGenrePreset{
			particleType: particles.ParticleSparkle,
			minSize:      2.0, maxSize: 5.0,
			duration: 0.7, gravity: -35.0, baseCount: 8,
		}
	}
}

// Update checks all entities for XP increases and emits burst particles.
func (s *XPGainBurstSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		expComp, ok := entity.GetComponent("experience")
		if !ok {
			continue
		}
		exp, ok := expComp.(*ExperienceComponent)
		if !ok || exp == nil {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		comp := s.ensureComponent(entity)

		// Tick cooldown
		if comp.BurstTimer > 0 {
			comp.BurstTimer -= deltaTime
		}

		if !comp.Initialized {
			comp.PrevTotalXP = exp.TotalXP
			comp.Initialized = true
			continue
		}

		// Detect XP increase
		xpDelta := exp.TotalXP - comp.PrevTotalXP
		if xpDelta > 0 && comp.BurstTimer <= 0 {
			// Scale intensity by XP gained relative to required XP
			ratio := float64(xpDelta) / float64(max(exp.RequiredXP, 1))
			s.spawnXPBurst(pos.X, pos.Y, entity.ID, ratio)
			comp.BurstTimer = s.burstInterval
		}

		comp.PrevTotalXP = exp.TotalXP
	}
}

// ensureComponent lazily attaches an XPGainBurstComponent to the entity.
func (s *XPGainBurstSystem) ensureComponent(entity *Entity) *XPGainBurstComponent {
	if comp, ok := entity.GetComponent("xp_gain_burst"); ok {
		return comp.(*XPGainBurstComponent)
	}
	comp := &XPGainBurstComponent{}
	entity.AddComponent(comp)
	return comp
}

// spawnXPBurst creates genre-aware upward-rising XP gain particles.
func (s *XPGainBurstSystem) spawnXPBurst(x, y float64, entityID uint64, xpRatio float64) {
	// Scale particle count by XP magnitude (clamped 1x-3x base)
	intensityMul := 1.0
	if xpRatio > 0.1 {
		intensityMul = 1.5
	}
	if xpRatio > 0.3 {
		intensityMul = 2.0
	}
	if xpRatio > 0.6 {
		intensityMul = 3.0
	}
	count := int(float64(s.preset.baseCount) * intensityMul)
	if count < 2 {
		count = 2
	}
	if count > 24 {
		count = 24
	}

	effectSeed := s.seed + int64(entityID)*37 + int64(x*11) + int64(y*17)

	config := particles.Config{
		Type:     s.preset.particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.preset.duration,
		SpreadX:  s.spreadRadius,
		SpreadY:  s.spreadRadius * 0.8,
		Gravity:  s.preset.gravity,
		MinSize:  s.preset.minSize,
		MaxSize:  s.preset.maxSize,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"xp_gain_burst": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}
