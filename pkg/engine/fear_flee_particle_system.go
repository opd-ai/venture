// Package engine provides the FearFleeParticleSystem for visual fear/flee feedback.
// This system connects StatusEffectAISystem with ParticleSystem to spawn genre-aware
// particle effects when entities become feared and start fleeing.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// FearFleeParticleSystem spawns particle effects when entities are feared and flee.
// It monitors entities with fear status effects and in flee state, providing visual
// feedback with genre-aware particle effects (ghostly wisps, panic sparks, etc).
type FearFleeParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Effect timing
	emitInterval   float64         // Seconds between particle emissions
	timeSinceEmit  float64         // Accumulator for timing
	baseCount      int             // Base particle count per emit
	effectRadius   float64         // Particle spread radius
	trailOffset    float64         // Distance behind entity for trail
	activeEntities map[uint64]bool // Track entities currently fleeing in fear
}

// NewFearFleeParticleSystem creates a new fear flee particle system.
func NewFearFleeParticleSystem(world *World, seed int64) *FearFleeParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "fear_flee_particle")
		logEntry.Debug("fear flee particle system created")
	}

	return &FearFleeParticleSystem{
		world:          world,
		genreID:        "fantasy",
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		emitInterval:   0.25, // Emit every 0.25s for responsive trail
		timeSinceEmit:  0.0,
		baseCount:      8,
		effectRadius:   40.0,
		trailOffset:    20.0,
		activeEntities: make(map[uint64]bool, 32),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *FearFleeParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *FearFleeParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update monitors entities for fear status effects and flee states.
func (s *FearFleeParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	if s.timeSinceEmit < s.emitInterval {
		return
	}

	for _, entity := range entities {
		aiComp := s.getAIComponent(entity)
		if aiComp == nil {
			continue
		}

		// Check for flee state from fear
		isFeared := s.hasFearEffect(entity)
		isFleeing := aiComp.State == AIStateFlee

		wasActive := s.activeEntities[entity.ID]

		if isFeared && isFleeing {
			pos := entity.GetPosition()
			if pos == nil {
				continue
			}

			// Spawn fear particles trailing behind entity
			s.spawnFearParticles(pos.X, pos.Y, entity, aiComp)

			// Mark as active for tracking
			s.activeEntities[entity.ID] = true

			// Log on first enter
			if !wasActive && s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"x":         pos.X,
					"y":         pos.Y,
				}).Debug("fear flee particles started")
			}
		} else if wasActive {
			// Entity stopped fleeing in fear
			delete(s.activeEntities, entity.ID)
			if s.logger != nil {
				s.logger.WithField("entity_id", entity.ID).Debug("fear flee particles stopped")
			}
		}
	}

	// Reset timer
	s.timeSinceEmit = 0
}

// getAIComponent retrieves the AI component from an entity.
func (s *FearFleeParticleSystem) getAIComponent(entity *Entity) *AIComponent {
	comp, ok := entity.GetComponent("ai")
	if !ok {
		return nil
	}
	aiComp, ok := comp.(*AIComponent)
	if !ok {
		return nil
	}
	return aiComp
}

// hasFearEffect checks if entity has an active fear status effect.
func (s *FearFleeParticleSystem) hasFearEffect(entity *Entity) bool {
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok {
			continue
		}
		if (effect.EffectType == "fear" || effect.EffectType == "feared") && !effect.IsExpired() {
			return true
		}
	}
	return false
}

// spawnFearParticles creates the fear particle trail effect.
func (s *FearFleeParticleSystem) spawnFearParticles(x, y float64, entity *Entity, aiComp *AIComponent) {
	// Calculate trail position behind entity based on velocity
	vel := entity.GetVelocity()
	trailX, trailY := x, y
	if vel != nil && (vel.VX != 0 || vel.VY != 0) {
		// Normalize velocity and offset backwards
		speed := vel.VX*vel.VX + vel.VY*vel.VY
		if speed > 0.01 {
			invSpeed := 1.0 / (speed * 0.5) // Approximate inverse sqrt
			if invSpeed > 1.0 {
				invSpeed = 1.0
			}
			trailX = x - vel.VX*invSpeed*s.trailOffset
			trailY = y - vel.VY*invSpeed*s.trailOffset
		}
	}

	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entity.ID)

	// Primary fear effect - ghostly/panic particles rising from entity
	primaryType := s.getPrimaryParticleType()
	primaryConfig := particles.Config{
		Type:     primaryType,
		Count:    s.baseCount,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.6,
		SpreadX:  s.effectRadius,
		SpreadY:  s.effectRadius * 0.7, // Slightly compressed vertically
		Gravity:  -80.0,                // Rise upward for ghostly effect
		MinSize:  3.0,
		MaxSize:  7.0,
		Custom:   map[string]interface{}{"fear_effect": true, "entity_id": entity.ID},
	}

	s.particleSystem.SpawnParticles(s.world, primaryConfig, trailX, trailY)

	// Secondary wispy trail for motion emphasis
	secondaryType := s.getSecondaryParticleType()
	secondaryConfig := particles.Config{
		Type:     secondaryType,
		Count:    s.baseCount / 2,
		GenreID:  s.genreID,
		Seed:     effectSeed + 1,
		Duration: 0.4,
		SpreadX:  s.effectRadius * 0.5,
		SpreadY:  s.effectRadius * 0.5,
		Gravity:  -40.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   map[string]interface{}{"fear_trail": true},
	}

	s.particleSystem.SpawnParticles(s.world, secondaryConfig, trailX, trailY)
}

// getPrimaryParticleType returns the primary particle type based on genre.
func (s *FearFleeParticleSystem) getPrimaryParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleMagic // Ghostly magical wisps
	case "scifi":
		return particles.ParticleSpark // Panic electrical sparks
	case "horror":
		return particles.ParticleSmoke // Dark shadowy smoke
	case "cyberpunk":
		return particles.ParticleEmber // Glitchy embers
	case "postapoc":
		return particles.ParticleDust // Panic dust clouds
	default:
		return particles.ParticleSparkle
	}
}

// getSecondaryParticleType returns the secondary particle type based on genre.
func (s *FearFleeParticleSystem) getSecondaryParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle
	case "scifi":
		return particles.ParticleSmoke
	case "horror":
		return particles.ParticleSmokePlume
	case "cyberpunk":
		return particles.ParticleSpark
	case "postapoc":
		return particles.ParticleSmoke
	default:
		return particles.ParticleDust
	}
}

// GetActiveCount returns the number of entities currently showing fear effects.
func (s *FearFleeParticleSystem) GetActiveCount() int {
	return len(s.activeEntities)
}

// IsEntityFleeing returns true if the entity is currently showing fear particles.
func (s *FearFleeParticleSystem) IsEntityFleeing(entityID uint64) bool {
	return s.activeEntities[entityID]
}
