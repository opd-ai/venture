// Package engine provides the FishingLineTensionParticleSystem for visual line tension feedback.
// This system connects FishingComponent with ParticleSystem to spawn genre-aware particle
// effects during the reeling minigame, visualizing line tension and fish struggle.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// FishingLineTensionParticleSystem spawns particles based on fishing line tension
// during the reeling minigame. High tension creates stress particles, fish struggles
// create splash particles, and successful tension management shows smooth water ripples.
type FishingLineTensionParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Timing control
	emitInterval  float64 // Time between particle emissions
	timeSinceEmit float64

	// Tension thresholds
	lowTensionThreshold    float64 // Below this, emit relaxed particles
	highTensionThreshold   float64 // Above this, emit stress particles
	criticalTensionWarning float64 // Critical threshold for warning particles

	// Track entities in reeling state for transition effects
	previouslyReeling map[uint64]bool
}

// NewFishingLineTensionParticleSystem creates a new fishing line tension particle system.
func NewFishingLineTensionParticleSystem(world *World, seed int64) *FishingLineTensionParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "fishing_line_tension_particle")
		logEntry.Debug("fishing line tension particle system created")
	}

	return &FishingLineTensionParticleSystem{
		world:                  world,
		seed:                   seed,
		rng:                    rand.New(rand.NewSource(seed)),
		logger:                 logEntry,
		genreID:                "fantasy",
		emitInterval:           0.15, // Emit particles every 150ms during reeling
		lowTensionThreshold:    0.3,
		highTensionThreshold:   0.7,
		criticalTensionWarning: 0.9,
		previouslyReeling:      make(map[uint64]bool),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *FishingLineTensionParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *FishingLineTensionParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities and spawns tension particles for active fishing.
func (s *FishingLineTensionParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	shouldEmit := s.timeSinceEmit >= s.emitInterval

	currentlyReeling := make(map[uint64]bool)

	for _, entity := range entities {
		s.processEntity(entity, shouldEmit, currentlyReeling)
	}

	// Handle entities that stopped reeling (fish caught or escaped)
	for entityID := range s.previouslyReeling {
		if !currentlyReeling[entityID] {
			s.onReelingEnded(entityID)
		}
	}

	s.previouslyReeling = currentlyReeling

	if shouldEmit {
		s.timeSinceEmit = 0
	}
}

// processEntity handles tension particles for a single fishing entity.
func (s *FishingLineTensionParticleSystem) processEntity(entity *Entity, shouldEmit bool, currentlyReeling map[uint64]bool) {
	comp, ok := entity.GetComponent("fishing")
	if !ok {
		return
	}

	fishComp, ok := comp.(*FishingComponent)
	if !ok || fishComp == nil {
		return
	}

	// Only process entities actively reeling
	if fishComp.GetState() != FishingStateReeling {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	currentlyReeling[entity.ID] = true

	// Check for transition into reeling
	if !s.previouslyReeling[entity.ID] {
		s.onReelingStarted(entity.ID, pos.X, pos.Y)
	}

	if !shouldEmit {
		return
	}

	tension := fishComp.GetTensionLevel()
	reelProgress := fishComp.GetReelProgress()
	_, fishWeight := fishComp.GetHookedFish()

	// Spawn tension-based particles
	s.spawnTensionParticles(entity.ID, pos.X, pos.Y, tension, reelProgress, fishWeight, fishComp.FishStruggleDirection)
}

// onReelingStarted spawns burst particles when reeling begins.
func (s *FishingLineTensionParticleSystem) onReelingStarted(entityID uint64, x, y float64) {
	effectSeed := s.seed + int64(entityID) + 1000

	config := particles.Config{
		Type:     particles.ParticleDust,
		Count:    8,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.5,
		SpreadX:  25.0,
		SpreadY:  15.0,
		Gravity:  -20.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"reeling_start": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y+10)

	if s.logger != nil {
		s.logger.WithField("entity_id", entityID).Debug("reeling started particles spawned")
	}
}

// onReelingEnded handles cleanup when an entity stops reeling.
func (s *FishingLineTensionParticleSystem) onReelingEnded(entityID uint64) {
	if s.logger != nil {
		s.logger.WithField("entity_id", entityID).Debug("entity stopped reeling")
	}
}

// spawnTensionParticles creates particles based on current line tension.
func (s *FishingLineTensionParticleSystem) spawnTensionParticles(entityID uint64, x, y, tension, reelProgress, fishWeight float64, struggleDir int) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(tension*1000)

	// Determine particle type and count based on tension level
	particleType, count, gravity := s.selectTensionParticles(tension, struggleDir)

	// Scale particles by fish weight (heavier fish = more dramatic effects)
	weightScale := 1.0 + (fishWeight / 20.0) // Max 2x at 20kg fish
	if weightScale > 2.0 {
		weightScale = 2.0
	}
	count = int(float64(count) * weightScale)

	// Cap at reasonable maximum
	if count > 12 {
		count = 12
	}

	// Offset particles toward struggle direction
	offsetX := float64(struggleDir) * 15.0

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.selectDuration(tension),
		SpreadX:  20.0 + tension*15.0,
		SpreadY:  10.0 + tension*10.0,
		Gravity:  gravity,
		MinSize:  1.5 + tension*0.5,
		MaxSize:  3.0 + tension*1.5,
		ZLayer:   particles.ZLayerGround,
		Custom: map[string]interface{}{
			"tension":       tension,
			"reel_progress": reelProgress,
			"struggle_dir":  struggleDir,
		},
	}

	s.particleSystem.SpawnParticles(s.world, config, x+offsetX, y+8)

	// Spawn warning particles at critical tension
	if tension >= s.criticalTensionWarning {
		s.spawnCriticalWarning(entityID, x, y, effectSeed+1)
	}

	// Spawn struggle splash when fish fights
	if struggleDir != 0 && tension > s.lowTensionThreshold {
		s.spawnStruggleSplash(x, y, struggleDir, tension, effectSeed+2)
	}
}

// selectTensionParticles returns particle type, count, and gravity based on tension.
func (s *FishingLineTensionParticleSystem) selectTensionParticles(tension float64, struggleDir int) (particles.ParticleType, int, float64) {
	switch s.genreID {
	case "fantasy":
		if tension >= s.criticalTensionWarning {
			return particles.ParticleSpark, 6, -30.0 // Magical strain sparks
		}
		if tension >= s.highTensionThreshold {
			return particles.ParticleSparkle, 5, -15.0 // Tense shimmer
		}
		return particles.ParticleDust, 3, 10.0 // Water ripples

	case "scifi":
		if tension >= s.criticalTensionWarning {
			return particles.ParticleSpark, 7, -40.0 // Tech overload
		}
		if tension >= s.highTensionThreshold {
			return particles.ParticleMagic, 5, -20.0 // Energy buildup
		}
		return particles.ParticleDust, 3, 5.0

	case "horror":
		if tension >= s.criticalTensionWarning {
			return particles.ParticleSmoke, 6, -10.0 // Dark mist warning
		}
		if tension >= s.highTensionThreshold {
			return particles.ParticleSmoke, 4, 0.0 // Eerie tension
		}
		return particles.ParticleDust, 2, 15.0

	case "cyberpunk":
		if tension >= s.criticalTensionWarning {
			return particles.ParticleSpark, 8, -35.0 // Neon overload
		}
		if tension >= s.highTensionThreshold {
			return particles.ParticleSpark, 5, -20.0 // Electric tension
		}
		return particles.ParticleDust, 3, 8.0

	case "postapoc":
		if tension >= s.criticalTensionWarning {
			return particles.ParticleSmoke, 5, 20.0 // Gritty strain
		}
		if tension >= s.highTensionThreshold {
			return particles.ParticleDust, 4, 10.0
		}
		return particles.ParticleDust, 2, 15.0

	default:
		if tension >= s.highTensionThreshold {
			return particles.ParticleSparkle, 4, -15.0
		}
		return particles.ParticleDust, 3, 10.0
	}
}

// selectDuration returns particle duration based on tension.
func (s *FishingLineTensionParticleSystem) selectDuration(tension float64) float64 {
	baseDuration := 0.3
	return baseDuration + tension*0.3 // 0.3-0.6s based on tension
}

// spawnCriticalWarning spawns warning particles at near-breaking tension.
func (s *FishingLineTensionParticleSystem) spawnCriticalWarning(entityID uint64, x, y float64, seed int64) {
	config := particles.Config{
		Type:     particles.ParticleSpark,
		Count:    4,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.25,
		SpreadX:  30.0,
		SpreadY:  20.0,
		Gravity:  -50.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"critical_warning": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
		}).Debug("critical tension warning particles")
	}
}

// spawnStruggleSplash spawns splash particles in the direction of fish struggle.
func (s *FishingLineTensionParticleSystem) spawnStruggleSplash(x, y float64, struggleDir int, tension float64, seed int64) {
	// Position splash in direction of struggle
	splashX := x + float64(struggleDir)*25.0

	count := 3
	if tension > s.highTensionThreshold {
		count = 5
	}

	config := particles.Config{
		Type:     particles.ParticleDust,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.35,
		SpreadX:  15.0,
		SpreadY:  8.0,
		Gravity:  40.0, // Fall back to water
		MinSize:  1.5,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"struggle_splash": true, "direction": struggleDir},
	}

	s.particleSystem.SpawnParticles(s.world, config, splashX, y+5)
}
