// Package engine provides the StealthIndicatorParticleSystem for terrain cover feedback.
// This system connects TerrainStealthSystem with ParticleSystem to spawn genre-aware
// particle effects when entities enter or exit terrain that provides stealth bonuses,
// giving players immediate visual feedback about their concealment state.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// StealthIndicatorParticleSystem spawns particle effects when entities enter/exit
// stealth-enhancing terrain. It bridges TerrainStealthSystem with ParticleSystem
// to provide visual feedback for cover and concealment states.
//
// Particle effects are spawned when:
//   - Entity enters terrain with stealth multiplier <0.9 (entering cover)
//   - Entity exits cover terrain back to normal visibility
//
// Genre-specific particle types:
//   - Fantasy: Sparkle particles (magical concealment)
//   - Scifi: Spark particles (cloaking field activation)
//   - Horror: Smoke particles (shadows enveloping)
//   - Cyberpunk: Spark particles (camouflage system)
//   - Postapoc: Dust particles (blending with debris)
type StealthIndicatorParticleSystem struct {
	world                *World
	particleSystem       *ParticleSystem
	terrainStealthSystem *TerrainStealthSystem
	genreID              string
	seed                 int64
	rng                  *rand.Rand
	logger               *logrus.Entry

	// Track last stealth state per entity for change detection
	lastStealthState map[uint64]stealthState

	// Particle configuration
	enterCoverCount int     // Particles when entering cover
	exitCoverCount  int     // Particles when leaving cover
	spreadFactor    float64 // Spread radius for particles

	// Thresholds for triggering effects
	coverThreshold float64 // Stealth multiplier below this triggers "in cover" effect
}

// stealthState tracks an entity's previous stealth condition
type stealthState struct {
	multiplier float64
	inCover    bool
}

// NewStealthIndicatorParticleSystem creates a new stealth indicator particle system.
func NewStealthIndicatorParticleSystem(world *World, seed int64) *StealthIndicatorParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "stealth_indicator_particle")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("stealth indicator particle system created")
	}

	return &StealthIndicatorParticleSystem{
		world:            world,
		seed:             seed,
		rng:              rand.New(rand.NewSource(seed)),
		logger:           logEntry,
		lastStealthState: make(map[uint64]stealthState, 64),
		enterCoverCount:  5,    // Subtle particle count
		exitCoverCount:   3,    // Minimal particles on exit
		spreadFactor:     30.0, // Tight spread around entity
		coverThreshold:   0.85, // Below 85% detection = in cover
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *StealthIndicatorParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetTerrainStealthSystem sets the terrain stealth system for multiplier lookups.
func (s *StealthIndicatorParticleSystem) SetTerrainStealthSystem(tss *TerrainStealthSystem) {
	s.terrainStealthSystem = tss
	if s.logger != nil {
		s.logger.Debug("terrain stealth system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors and types.
func (s *StealthIndicatorParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all entities and spawns particles on stealth state changes.
func (s *StealthIndicatorParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.terrainStealthSystem == nil {
		return
	}

	for _, entity := range entities {
		s.updateEntityStealthIndicator(entity)
	}
}

// updateEntityStealthIndicator checks for stealth state changes and spawns particles.
func (s *StealthIndicatorParticleSystem) updateEntityStealthIndicator(entity *Entity) {
	// Only process entities that can be targets (have health)
	if !entity.HasComponent("health") {
		s.cleanupEntity(entity.ID)
		return
	}

	// Get current stealth multiplier from terrain system
	currentMult := s.terrainStealthSystem.GetStealthMultiplier(entity.ID)
	currentInCover := currentMult < s.coverThreshold

	// Get previous state
	prevState, hadState := s.lastStealthState[entity.ID]

	// Update tracked state
	s.lastStealthState[entity.ID] = stealthState{
		multiplier: currentMult,
		inCover:    currentInCover,
	}

	// Skip if first frame for this entity (no transition)
	if !hadState {
		return
	}

	// Check for state change
	if currentInCover && !prevState.inCover {
		// Entered cover
		s.spawnEnterCoverParticles(entity, currentMult)
	} else if !currentInCover && prevState.inCover {
		// Left cover
		s.spawnExitCoverParticles(entity, prevState.multiplier)
	}
}

// cleanupEntity removes tracking for an entity.
func (s *StealthIndicatorParticleSystem) cleanupEntity(entityID uint64) {
	delete(s.lastStealthState, entityID)
}

// spawnEnterCoverParticles creates particles when entering stealth-enhancing terrain.
func (s *StealthIndicatorParticleSystem) spawnEnterCoverParticles(entity *Entity, stealthMult float64) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Scale particle count with stealth effectiveness
	// Better cover (lower multiplier) = more particles
	count := s.enterCoverCount
	if stealthMult < 0.7 {
		count = int(float64(count) * 1.4)
	}
	if stealthMult < 0.5 {
		count = int(float64(count) * 1.3)
	}
	if count > 12 {
		count = 12
	}

	effectSeed := s.seed + int64(pos.X*997) + int64(pos.Y*991)
	particleType := s.getEnterCoverParticleType()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.6,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -20.0, // Gentle upward drift
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["stealth_enter"] = true
	config.Custom["stealth_mult"] = stealthMult

	s.particleSystem.SpawnParticles(s.world, config, pos.X, pos.Y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"stealth_mult": stealthMult,
			"x":            pos.X,
			"y":            pos.Y,
			"particles":    count,
		}).Debug("enter cover particles spawned")
	}
}

// spawnExitCoverParticles creates particles when leaving stealth-enhancing terrain.
func (s *StealthIndicatorParticleSystem) spawnExitCoverParticles(entity *Entity, prevStealthMult float64) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	effectSeed := s.seed + int64(pos.X*1009) + int64(pos.Y*1013)
	particleType := s.getExitCoverParticleType()

	config := particles.Config{
		Type:     particleType,
		Count:    s.exitCoverCount,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.4,
		SpreadX:  s.spreadFactor * 1.5, // Wider spread on exit
		SpreadY:  s.spreadFactor * 1.5,
		Gravity:  10.0, // Slight downward to suggest "revealing"
		MinSize:  1.5,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["stealth_exit"] = true

	s.particleSystem.SpawnParticles(s.world, config, pos.X, pos.Y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"x":         pos.X,
			"y":         pos.Y,
		}).Debug("exit cover particles spawned")
	}
}

// getEnterCoverParticleType returns the particle type for entering cover.
func (s *StealthIndicatorParticleSystem) getEnterCoverParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Magical concealment shimmer
	case "scifi":
		return particles.ParticleSpark // Cloaking field activation
	case "horror":
		return particles.ParticleSmoke // Shadows gathering
	case "cyberpunk":
		return particles.ParticleSpark // Camouflage system online
	case "postapoc":
		return particles.ParticleDust // Blending with environment
	default:
		return particles.ParticleSparkle
	}
}

// getExitCoverParticleType returns the particle type for leaving cover.
func (s *StealthIndicatorParticleSystem) getExitCoverParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleDust // Magical concealment fading
	case "scifi":
		return particles.ParticleSpark // Cloak deactivating
	case "horror":
		return particles.ParticleSmoke // Shadows dispersing
	case "cyberpunk":
		return particles.ParticleSpark // Camouflage offline
	case "postapoc":
		return particles.ParticleDust // Disturbed debris
	default:
		return particles.ParticleDust
	}
}

// SetCoverThreshold configures when entities are considered "in cover".
// Default is 0.85 (stealth multiplier below 85% triggers cover state).
func (s *StealthIndicatorParticleSystem) SetCoverThreshold(threshold float64) {
	if threshold > 0 && threshold < 1.0 {
		s.coverThreshold = threshold
	}
}

// GetCoverState returns whether an entity is currently in cover and their stealth multiplier.
func (s *StealthIndicatorParticleSystem) GetCoverState(entityID uint64) (inCover bool, multiplier float64) {
	if state, ok := s.lastStealthState[entityID]; ok {
		return state.inCover, state.multiplier
	}
	return false, 1.0
}
