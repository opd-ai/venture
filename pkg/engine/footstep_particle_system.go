// Package engine provides the FootstepParticleSystem for terrain-aware movement visuals.
// This system connects MovementSystem with ParticleSystem to spawn genre-aware particle
// effects when entities move on different terrain types (dust on floors, splashes in water,
// sparks on metal), enhancing movement visual feedback.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// FootstepParticleSystem spawns particle effects based on terrain type during movement.
// It connects terrain data and ParticleSystem to provide visual feedback when entities
// move across different surfaces, with genre-aware particle colors and behaviors.
type FootstepParticleSystem struct {
	world          *World
	terrain        *terrain.Terrain
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Configuration
	tileSize      int
	spawnInterval float64 // Minimum time between footsteps per entity
	minSpeedSq    float64 // Minimum squared speed to trigger footsteps

	// Per-entity cooldown tracking to limit particle spawning
	lastFootstep map[uint64]float64
	// Last tile position per entity to detect tile changes
	lastTile map[uint64]tileCoord
}

// tileCoord stores tile coordinates for tracking position changes.
type tileCoord struct {
	x, y int
}

// NewFootstepParticleSystem creates a new footstep particle system.
func NewFootstepParticleSystem(world *World, seed int64) *FootstepParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "footstep_particle")
		logEntry.Debug("footstep particle system created")
	}

	return &FootstepParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		spawnInterval: 0.15,  // 150ms between footsteps
		minSpeedSq:    100.0, // sqrt(100) = 10 px/s minimum
		lastFootstep:  make(map[uint64]float64, 32),
		lastTile:      make(map[uint64]tileCoord, 32),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *FootstepParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetTerrain sets the terrain data for tile lookups.
func (s *FootstepParticleSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
	s.lastTile = make(map[uint64]tileCoord, 32)
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *FootstepParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetTileSize sets the tile size in pixels.
func (s *FootstepParticleSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// Update processes moving entities and spawns terrain-appropriate footstep particles.
func (s *FootstepParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil || s.terrain == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}
}

// processEntity handles footstep spawning for a single entity.
func (s *FootstepParticleSystem) processEntity(entity *Entity, deltaTime float64) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	vel := entity.GetVelocity()
	if vel == nil {
		return
	}

	// Check if entity is moving fast enough
	speedSq := vel.VX*vel.VX + vel.VY*vel.VY
	if speedSq < s.minSpeedSq {
		return
	}

	// Check cooldown
	lastTime := s.lastFootstep[entity.ID]
	lastTime += deltaTime
	if lastTime < s.spawnInterval {
		s.lastFootstep[entity.ID] = lastTime
		return
	}

	// Get tile coordinates
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	// Check if we changed tiles (more noticeable footsteps on tile boundaries)
	last, hasLast := s.lastTile[entity.ID]
	s.lastTile[entity.ID] = tileCoord{x: tileX, y: tileY}

	// Only spawn on tile changes or periodic intervals
	onNewTile := !hasLast || last.x != tileX || last.y != tileY
	if !onNewTile && lastTime < s.spawnInterval*2 {
		s.lastFootstep[entity.ID] = lastTime
		return
	}

	// Reset cooldown
	s.lastFootstep[entity.ID] = 0

	// Get tile type and spawn appropriate particles
	tileType := s.terrain.GetTile(tileX, tileY)
	s.spawnFootstepParticles(entity.ID, pos.X, pos.Y, tileType)
}

// spawnFootstepParticles creates particles appropriate for the terrain type.
func (s *FootstepParticleSystem) spawnFootstepParticles(entityID uint64, x, y float64, tileType terrain.TileType) {
	config := s.getParticleConfig(tileType, x, y)
	if config == nil {
		return // No particles for this terrain type
	}

	s.particleSystem.SpawnParticles(s.world, *config, x, y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"tile_type": tileType.String(),
			"x":         x,
			"y":         y,
		}).Debug("footstep particles spawned")
	}
}

// getParticleConfig returns particle configuration for a terrain type.
func (s *FootstepParticleSystem) getParticleConfig(tileType terrain.TileType, x, y float64) *particles.Config {
	effectSeed := s.seed + int64(x*100) + int64(y*100)

	switch tileType {
	case terrain.TileFloor, terrain.TileCorridor, terrain.TileDoor:
		return s.getFloorConfig(effectSeed)
	case terrain.TileWaterShallow:
		return s.getWaterConfig(effectSeed)
	case terrain.TileLavaFlow:
		return s.getLavaConfig(effectSeed)
	case terrain.TileBridge:
		return s.getBridgeConfig(effectSeed)
	case terrain.TileRamp, terrain.TileRampUp, terrain.TileRampDown:
		return s.getRampConfig(effectSeed)
	case terrain.TilePlatform:
		return s.getPlatformConfig(effectSeed)
	default:
		return nil // No footstep effect for walls, trees, etc.
	}
}

// getFloorConfig returns dust particles for standard floors.
func (s *FootstepParticleSystem) getFloorConfig(seed int64) *particles.Config {
	count := 2
	// Genre variations
	switch s.genreID {
	case "horror":
		count = 3 // More dust in horror for atmosphere
	case "scifi":
		count = 1 // Clean sci-fi floors
	}

	return &particles.Config{
		Type:     particles.ParticleDust,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.4,
		SpreadX:  15.0,
		SpreadY:  8.0,
		Gravity:  20.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"footstep": true},
	}
}

// getWaterConfig returns splash particles for shallow water.
func (s *FootstepParticleSystem) getWaterConfig(seed int64) *particles.Config {
	return &particles.Config{
		Type:     particles.ParticleSpark, // Sparkle for water droplets
		Count:    4,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.5,
		SpreadX:  20.0,
		SpreadY:  25.0,
		Gravity:  60.0, // Falls back down
		MinSize:  2.0,
		MaxSize:  5.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"water_splash": true},
	}
}

// getLavaConfig returns ember particles for lava flows.
func (s *FootstepParticleSystem) getLavaConfig(seed int64) *particles.Config {
	return &particles.Config{
		Type:     particles.ParticleEmber,
		Count:    3,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.6,
		SpreadX:  12.0,
		SpreadY:  15.0,
		Gravity:  -30.0, // Embers rise
		MinSize:  3.0,
		MaxSize:  6.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"lava_step": true},
	}
}

// getBridgeConfig returns wood/creaking particles for bridges.
func (s *FootstepParticleSystem) getBridgeConfig(seed int64) *particles.Config {
	return &particles.Config{
		Type:     particles.ParticleDust,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.35,
		SpreadX:  10.0,
		SpreadY:  5.0,
		Gravity:  30.0,
		MinSize:  2.0,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"wood_dust": true},
	}
}

// getRampConfig returns particles for ramps.
func (s *FootstepParticleSystem) getRampConfig(seed int64) *particles.Config {
	count := 2
	particleType := particles.ParticleDust

	// Sci-fi ramps might have sparks
	if s.genreID == "scifi" || s.genreID == "cyberpunk" {
		particleType = particles.ParticleSpark
		count = 1
	}

	return &particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.3,
		SpreadX:  12.0,
		SpreadY:  10.0,
		Gravity:  25.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"ramp_step": true},
	}
}

// getPlatformConfig returns particles for elevated platforms.
func (s *FootstepParticleSystem) getPlatformConfig(seed int64) *particles.Config {
	// Platforms are typically metal/stone - subtle dust or sparks
	particleType := particles.ParticleDust
	if s.genreID == "scifi" || s.genreID == "cyberpunk" {
		particleType = particles.ParticleSpark
	}

	return &particles.Config{
		Type:     particleType,
		Count:    1,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.25,
		SpreadX:  8.0,
		SpreadY:  6.0,
		Gravity:  35.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"platform_step": true},
	}
}
