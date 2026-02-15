// Package engine provides the EntityIdleAmbientParticleSystem for idle entity visuals.
// This system connects VelocityComponent with ParticleSystem to spawn subtle
// genre-aware ambient particles around entities that have been stationary for a
// configurable duration, adding visual life to idle players, NPCs, and creatures.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// idleEntityState tracks per-entity idle duration and last spawn time.
type idleEntityState struct {
	idleDuration float64 // Seconds entity has been stationary
	lastSpawn    float64 // Accumulated time at last particle spawn
}

// EntityIdleAmbientParticleSystem spawns genre-aware ambient particles around
// entities that have been stationary (velocity ≈ 0) for a configurable duration.
type EntityIdleAmbientParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Configurable thresholds
	idleThreshold   float64 // Seconds before considered idle (default: 1.5)
	spawnInterval   float64 // Seconds between particle spawns per entity (default: 1.2)
	velocityEpsilon float64 // Max speed² to count as stationary (default: 4.0 i.e. 2px/s)
	spawnRadius     float64 // Pixel radius around entity for particle offset (default: 12.0)

	// Per-entity state tracking
	entityStates map[uint64]*idleEntityState
	elapsed      float64
}

// NewEntityIdleAmbientParticleSystem creates a new idle ambient particle system.
func NewEntityIdleAmbientParticleSystem(world *World, seed int64) *EntityIdleAmbientParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "entity_idle_ambient_particle")
		logEntry.Debug("entity idle ambient particle system created")
	}

	return &EntityIdleAmbientParticleSystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		idleThreshold:   1.5,
		spawnInterval:   1.2,
		velocityEpsilon: 4.0,
		spawnRadius:     12.0,
		entityStates:    make(map[uint64]*idleEntityState, 32),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *EntityIdleAmbientParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetGenre sets the genre ID for genre-aware particle selection.
func (s *EntityIdleAmbientParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes entities and spawns idle ambient particles for stationary ones.
func (s *EntityIdleAmbientParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.elapsed += deltaTime

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}

	// Prune stale entries every ~10 seconds
	if int(s.elapsed)%10 == 0 && len(s.entityStates) > 100 {
		s.pruneStaleEntities(entities)
	}
}

// processEntity checks if an entity is idle and spawns particles if appropriate.
func (s *EntityIdleAmbientParticleSystem) processEntity(entity *Entity, deltaTime float64) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}
	if !entity.HasComponent("health") {
		return
	}

	vel := entity.GetVelocity()
	speedSq := 0.0
	if vel != nil {
		speedSq = vel.VX*vel.VX + vel.VY*vel.VY
	}

	state, exists := s.entityStates[entity.ID]
	if !exists {
		state = &idleEntityState{}
		s.entityStates[entity.ID] = state
	}

	if speedSq > s.velocityEpsilon {
		// Entity is moving — reset idle timer
		state.idleDuration = 0
		return
	}

	state.idleDuration += deltaTime

	if state.idleDuration < s.idleThreshold {
		return
	}

	// Check spawn cooldown
	if (s.elapsed - state.lastSpawn) < s.spawnInterval {
		return
	}
	state.lastSpawn = s.elapsed

	config := s.getIdleParticleConfig(entity, pos.X, pos.Y)
	if config == nil {
		return
	}

	offsetX := (s.rng.Float64() - 0.5) * s.spawnRadius * 2
	offsetY := (s.rng.Float64() - 0.5) * s.spawnRadius * 2
	s.particleSystem.SpawnParticles(s.world, *config, pos.X+offsetX, pos.Y+offsetY)
}

// getIdleParticleConfig returns genre-aware particle config for an idle entity.
func (s *EntityIdleAmbientParticleSystem) getIdleParticleConfig(entity *Entity, x, y float64) *particles.Config {
	seed := s.seed + int64(x*31) + int64(y*17) + int64(entity.ID)

	isPlayer := entity.HasComponent("player") || entity.HasComponent("input")

	switch s.genreID {
	case "fantasy":
		return s.fantasyIdleConfig(seed, isPlayer)
	case "horror":
		return s.horrorIdleConfig(seed, isPlayer)
	case "scifi":
		return s.scifiIdleConfig(seed, isPlayer)
	case "cyberpunk":
		return s.cyberpunkIdleConfig(seed, isPlayer)
	case "postapoc":
		return s.postapocIdleConfig(seed, isPlayer)
	default:
		return s.fantasyIdleConfig(seed, isPlayer)
	}
}

func (s *EntityIdleAmbientParticleSystem) fantasyIdleConfig(seed int64, isPlayer bool) *particles.Config {
	pType := particles.ParticleSparkle
	count := 1
	if isPlayer {
		count = 2
	}
	return &particles.Config{
		Type:     pType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 2.5,
		SpreadX:  4.0,
		SpreadY:  4.0,
		Gravity:  -2.0,
		MinSize:  1.0,
		MaxSize:  2.0,
		ZLayer:   particles.ZLayerAbove,
	}
}

func (s *EntityIdleAmbientParticleSystem) horrorIdleConfig(seed int64, isPlayer bool) *particles.Config {
	count := 1
	if isPlayer {
		count = 2
	}
	return &particles.Config{
		Type:     particles.ParticleSmoke,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 3.0,
		SpreadX:  3.0,
		SpreadY:  3.0,
		Gravity:  -1.5,
		MinSize:  1.0,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerGround,
	}
}

func (s *EntityIdleAmbientParticleSystem) scifiIdleConfig(seed int64, isPlayer bool) *particles.Config {
	count := 1
	if isPlayer {
		count = 2
	}
	return &particles.Config{
		Type:     particles.ParticleSpark,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 1.8,
		SpreadX:  5.0,
		SpreadY:  5.0,
		Gravity:  0.0,
		MinSize:  1.0,
		MaxSize:  2.0,
		ZLayer:   particles.ZLayerAbove,
	}
}

func (s *EntityIdleAmbientParticleSystem) cyberpunkIdleConfig(seed int64, isPlayer bool) *particles.Config {
	count := 1
	if isPlayer {
		count = 2
	}
	// Glitch-like sparks
	return &particles.Config{
		Type:     particles.ParticleSpark,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 1.5,
		SpreadX:  6.0,
		SpreadY:  3.0,
		Gravity:  1.0,
		MinSize:  1.0,
		MaxSize:  2.0,
		ZLayer:   particles.ZLayerAbove,
	}
}

func (s *EntityIdleAmbientParticleSystem) postapocIdleConfig(seed int64, isPlayer bool) *particles.Config {
	count := 1
	if isPlayer {
		count = 2
	}
	return &particles.Config{
		Type:     particles.ParticleDust,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 2.0,
		SpreadX:  5.0,
		SpreadY:  5.0,
		Gravity:  -1.0,
		MinSize:  1.0,
		MaxSize:  2.0,
		ZLayer:   particles.ZLayerGround,
	}
}

// pruneStaleEntities removes tracking for entities no longer in the update set.
func (s *EntityIdleAmbientParticleSystem) pruneStaleEntities(entities []*Entity) {
	active := make(map[uint64]struct{}, len(entities))
	for _, e := range entities {
		active[e.ID] = struct{}{}
	}
	for id := range s.entityStates {
		if _, ok := active[id]; !ok {
			delete(s.entityStates, id)
		}
	}
}

// GetIdleDuration returns how long an entity has been idle (for testing).
func (s *EntityIdleAmbientParticleSystem) GetIdleDuration(entityID uint64) float64 {
	if state, ok := s.entityStates[entityID]; ok {
		return state.idleDuration
	}
	return 0
}

// GetEntityStateCount returns the number of tracked entity states (for testing).
func (s *EntityIdleAmbientParticleSystem) GetEntityStateCount() int {
	return len(s.entityStates)
}

// idleParticleSpeedThreshold returns the velocity epsilon (for testing).
func (s *EntityIdleAmbientParticleSystem) idleParticleSpeedThreshold() float64 {
	return math.Sqrt(s.velocityEpsilon)
}
