// Package engine provides the SprintTrailParticleSystem for speed-based movement visuals.
// This system monitors entity velocity and spawns genre-aware trail particles behind
// fast-moving entities, creating visual speed feedback for sprinting/dashing characters.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// sprintTrailPreset holds genre-specific trail particle configuration.
type sprintTrailPreset struct {
	ParticleType particles.ParticleType
	Count        int     // Particles per spawn
	Duration     float64 // Particle lifetime
	MinSize      float64
	MaxSize      float64
	SpreadX      float64
	SpreadY      float64
	Gravity      float64
}

// SprintTrailParticleSystem spawns trail particles behind entities moving above
// a sprint speed threshold. Trail visuals vary by genre: fantasy emits sparkle dust,
// sci-fi emits energy streaks, horror emits shadow wisps, cyberpunk emits neon traces.
type SprintTrailParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Speed threshold (squared) to trigger trail
	sprintSpeedSq float64
	// Minimum time between trail spawns per entity
	spawnInterval float64
	// Per-entity cooldown tracking
	lastSpawn map[uint64]float64
	// Genre preset
	preset sprintTrailPreset
}

// NewSprintTrailParticleSystem creates a sprint trail particle system.
func NewSprintTrailParticleSystem(world *World, seed int64) *SprintTrailParticleSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "sprint_trail_particle",
	})

	sys := &SprintTrailParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logger,
		sprintSpeedSq: 10000.0, // 100 px/s threshold
		spawnInterval: 0.08,    // 80ms between spawns
		lastSpawn:     make(map[uint64]float64, 32),
		genreID:       "fantasy",
	}

	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("sprint trail particle system created")
	return sys
}

// SetParticleSystem sets the particle system for spawning effects.
func (s *SprintTrailParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetGenre configures genre-specific trail visuals.
func (s *SprintTrailParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
}

// getGenrePreset returns trail configuration for the given genre.
func (s *SprintTrailParticleSystem) getGenrePreset(genreID string) sprintTrailPreset {
	switch genreID {
	case "horror":
		return sprintTrailPreset{
			ParticleType: particles.ParticleSmoke,
			Count:        4,
			Duration:     0.6,
			MinSize:      3.0,
			MaxSize:      6.0,
			SpreadX:      30.0,
			SpreadY:      30.0,
			Gravity:      -15.0, // Wisps float upward
		}
	case "cyberpunk":
		return sprintTrailPreset{
			ParticleType: particles.ParticleSpark,
			Count:        6,
			Duration:     0.3,
			MinSize:      1.0,
			MaxSize:      3.0,
			SpreadX:      50.0,
			SpreadY:      20.0,
			Gravity:      0.0,
		}
	case "scifi":
		return sprintTrailPreset{
			ParticleType: particles.ParticleMagic,
			Count:        5,
			Duration:     0.35,
			MinSize:      2.0,
			MaxSize:      4.0,
			SpreadX:      40.0,
			SpreadY:      15.0,
			Gravity:      0.0,
		}
	case "postapoc":
		return sprintTrailPreset{
			ParticleType: particles.ParticleDust,
			Count:        6,
			Duration:     0.5,
			MinSize:      2.0,
			MaxSize:      5.0,
			SpreadX:      40.0,
			SpreadY:      40.0,
			Gravity:      30.0, // Dust settles
		}
	default: // fantasy
		return sprintTrailPreset{
			ParticleType: particles.ParticleSparkle,
			Count:        5,
			Duration:     0.4,
			MinSize:      1.0,
			MaxSize:      3.0,
			SpreadX:      35.0,
			SpreadY:      25.0,
			Gravity:      -10.0, // Sparkles float slightly
		}
	}
}

// Update processes moving entities and spawns trail particles for sprinting ones.
func (s *SprintTrailParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		vel := entity.GetVelocity()
		if vel == nil {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		speedSq := vel.VX*vel.VX + vel.VY*vel.VY
		if speedSq < s.sprintSpeedSq {
			delete(s.lastSpawn, entity.ID)
			continue
		}

		// Throttle spawning per entity
		s.lastSpawn[entity.ID] += deltaTime
		if s.lastSpawn[entity.ID] < s.spawnInterval {
			continue
		}
		s.lastSpawn[entity.ID] = 0

		// Spawn trail offset behind the movement direction
		speed := math.Sqrt(speedSq)
		offsetX := -(vel.VX / speed) * 8.0
		offsetY := -(vel.VY / speed) * 8.0

		// Scale particle count with speed (more particles when faster)
		speedRatio := math.Min(speed/200.0, 2.0)
		count := int(float64(s.preset.Count) * speedRatio)
		if count < 1 {
			count = 1
		}

		config := particles.Config{
			Type:    s.preset.ParticleType,
			Count:   count,
			GenreID: s.genreID,
			Seed:    s.seed + int64(entity.ID) + int64(s.rng.Intn(1000)),
			Duration: s.preset.Duration,
			SpreadX:  s.preset.SpreadX,
			SpreadY:  s.preset.SpreadY,
			Gravity:  s.preset.Gravity,
			MinSize:  s.preset.MinSize,
			MaxSize:  s.preset.MaxSize,
			ZLayer:   particles.ZLayerGround,
			Custom:   make(map[string]interface{}),
		}

		s.particleSystem.SpawnParticles(s.world, config, pos.X+offsetX, pos.Y+offsetY)
	}
}
