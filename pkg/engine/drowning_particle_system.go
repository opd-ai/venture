// Package engine provides the DrowningParticleSystem for visual drowning feedback.
// This system connects SwimmingComponent drowning state with ParticleSystem to spawn
// genre-aware particle effects when entities are drowning, giving players immediate
// visual feedback about suffocation in water.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// DrowningParticleSystem spawns particle effects when entities are drowning.
// It monitors SwimmingComponent.Drowning state and spawns bubble/suffocation
// particles with genre-aware visuals.
type DrowningParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Effect timing
	emitInterval  float64 // Seconds between particle bursts
	timeSinceEmit float64 // Accumulator for emit timing
	baseCount     int     // Base particle count per emit
	spreadFactor  float64 // Particle spread radius
}

// NewDrowningParticleSystem creates a new drowning particle system.
func NewDrowningParticleSystem(world *World, seed int64) *DrowningParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "drowning_particle")
		logEntry.Debug("drowning particle system created")
	}

	return &DrowningParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		emitInterval:  0.5, // Emit bubbles every 0.5 seconds
		timeSinceEmit: 0.0,
		baseCount:     6,
		spreadFactor:  30.0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *DrowningParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *DrowningParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update monitors entities with swimming components and spawns drowning particles.
func (s *DrowningParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	if s.timeSinceEmit < s.emitInterval {
		return
	}

	for _, entity := range entities {
		if !s.isDrowning(entity) {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		s.spawnDrowningParticles(pos.X, pos.Y, entity.ID)

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"x":         pos.X,
				"y":         pos.Y,
			}).Debug("drowning particles spawned")
		}
	}

	// Reset timer after processing
	s.timeSinceEmit = 0
}

// isDrowning checks if entity has swimming component with Drowning = true.
func (s *DrowningParticleSystem) isDrowning(entity *Entity) bool {
	swimmingComp, ok := entity.GetComponent("swimming")
	if !ok {
		return false
	}

	swimming, ok := swimmingComp.(*fluids.SwimmingComponent)
	if !ok {
		return false
	}

	return swimming.Drowning
}

// spawnDrowningParticles creates the drowning bubble/suffocation effect.
func (s *DrowningParticleSystem) spawnDrowningParticles(x, y float64, entityID uint64) {
	// Use deterministic seed offset for consistent visuals
	effectSeed := s.seed + int64(entityID*1000) + int64(x*10) + int64(y*10)

	config := particles.Config{
		Type:     s.getParticleType(),
		Count:    s.baseCount,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 1.2, // Bubbles rise for 1.2 seconds
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.5,
		Gravity:  -100.0, // Bubbles rise upward
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as drowning effect for special underwater rendering
	config.Custom["drowning_effect"] = true
	config.Custom["bubble_wobble"] = true

	s.particleSystem.SpawnParticles(s.world, config, x, y-10) // Spawn slightly above head
}

// getParticleType returns the genre-appropriate particle type for drowning.
func (s *DrowningParticleSystem) getParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleMagic // Blue magical bubbles
	case "scifi":
		return particles.ParticleSpark // Oxygen leak sparks
	case "horror":
		return particles.ParticleSmoke // Dark suffocation mist
	case "cyberpunk":
		return particles.ParticleSpark // Coolant bubbles
	case "postapoc":
		return particles.ParticleDebris // Contaminated water bubbles
	default:
		return particles.ParticleMagic
	}
}
