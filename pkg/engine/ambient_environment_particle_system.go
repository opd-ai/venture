// Package engine provides the AmbientEnvironmentParticleSystem for atmospheric visuals.
// This system connects terrain data with ParticleSystem to spawn genre-aware ambient
// particles based on the terrain type near active entities (fireflies in forests,
// mist over water, embers near lava, dust in corridors), enhancing environmental atmosphere.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// AmbientEnvironmentParticleSystem spawns atmospheric particle effects based on
// terrain type and genre, providing passive environmental visual feedback.
type AmbientEnvironmentParticleSystem struct {
	world          *World
	terrain        *terrain.Terrain
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	tileSize      int
	spawnInterval float64 // Seconds between ambient spawn attempts per entity
	spawnRadius   float64 // Pixel radius around entity to spawn particles

	// Per-entity cooldown tracking
	lastSpawn map[uint64]float64
	// Accumulated time for global tick
	elapsed float64
}

// NewAmbientEnvironmentParticleSystem creates a new ambient environment particle system.
func NewAmbientEnvironmentParticleSystem(world *World, seed int64) *AmbientEnvironmentParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "ambient_environment_particle")
		logEntry.Debug("ambient environment particle system created")
	}

	return &AmbientEnvironmentParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		tileSize:      32,
		spawnInterval: 0.8, // Spawn attempt every 800ms per entity
		spawnRadius:   48.0,
		lastSpawn:     make(map[uint64]float64, 16),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *AmbientEnvironmentParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetTerrain sets the terrain data for tile lookups.
func (s *AmbientEnvironmentParticleSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrain = terr
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *AmbientEnvironmentParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// SetTileSize sets the tile size in pixels.
func (s *AmbientEnvironmentParticleSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// Update processes entities and spawns ambient particles based on nearby terrain.
func (s *AmbientEnvironmentParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil || s.terrain == nil {
		return
	}

	s.elapsed += deltaTime

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}
}

// processEntity checks terrain near entity and possibly spawns ambient particles.
func (s *AmbientEnvironmentParticleSystem) processEntity(entity *Entity, deltaTime float64) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}
	if !entity.HasComponent("health") {
		return
	}

	// Cooldown check
	lastTime, tracked := s.lastSpawn[entity.ID]
	if tracked && (s.elapsed-lastTime) < s.spawnInterval {
		return
	}
	s.lastSpawn[entity.ID] = s.elapsed

	// Sample terrain at entity position
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	if !s.terrain.IsInBounds(tileX, tileY) {
		return
	}

	tileType := s.terrain.GetTile(tileX, tileY)
	config := s.getAmbientConfig(tileType, pos.X, pos.Y)
	if config == nil {
		return
	}

	// Offset spawn position slightly for natural feel
	offsetX := (s.rng.Float64() - 0.5) * s.spawnRadius * 2
	offsetY := (s.rng.Float64() - 0.5) * s.spawnRadius * 2
	s.particleSystem.SpawnParticles(s.world, *config, pos.X+offsetX, pos.Y+offsetY)
}

// getAmbientConfig returns the appropriate ambient particle config for a terrain type.
func (s *AmbientEnvironmentParticleSystem) getAmbientConfig(tileType terrain.TileType, x, y float64) *particles.Config {
	seed := s.seed + int64(x*31) + int64(y*17)

	switch tileType {
	case terrain.TileFloor, terrain.TileCorridor:
		// Only spawn dust motes 30% of the time for subtlety
		if s.rng.Float64() > 0.3 {
			return nil
		}
		return s.getDustMoteConfig(seed)
	case terrain.TileWaterShallow:
		return s.getMistConfig(seed)
	case terrain.TileWaterDeep:
		return s.getDeepWaterMistConfig(seed)
	case terrain.TileLavaFlow:
		return s.getLavaEmberConfig(seed)
	case terrain.TileTree:
		return s.getForestAmbientConfig(seed)
	case terrain.TileBridge:
		// Occasional mist wisps under bridges
		if s.rng.Float64() > 0.4 {
			return nil
		}
		return s.getBridgeMistConfig(seed)
	case terrain.TileStructure:
		if s.rng.Float64() > 0.25 {
			return nil
		}
		return s.getStructureDustConfig(seed)
	default:
		return nil
	}
}

// getDustMoteConfig returns floating dust mote particles for indoor areas.
func (s *AmbientEnvironmentParticleSystem) getDustMoteConfig(seed int64) *particles.Config {
	pType := particles.ParticleDust
	if s.genreID == "scifi" || s.genreID == "cyberpunk" {
		pType = particles.ParticleSpark
	}
	return &particles.Config{
		Type:     pType,
		Count:    1,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 2.0,
		SpreadX:  6.0,
		SpreadY:  6.0,
		Gravity:  -3.0, // Drift upward slowly
		MinSize:  1.0,
		MaxSize:  2.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"ambient_dust": true},
	}
}

// getMistConfig returns mist particles for shallow water areas.
func (s *AmbientEnvironmentParticleSystem) getMistConfig(seed int64) *particles.Config {
	pType := particles.ParticleSmoke
	if s.genreID == "horror" {
		pType = particles.ParticleSmokePlume
	}
	return &particles.Config{
		Type:     pType,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 2.5,
		SpreadX:  20.0,
		SpreadY:  10.0,
		Gravity:  -5.0,
		MinSize:  3.0,
		MaxSize:  6.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"ambient_mist": true},
	}
}

// getDeepWaterMistConfig returns denser mist for deep water.
func (s *AmbientEnvironmentParticleSystem) getDeepWaterMistConfig(seed int64) *particles.Config {
	return &particles.Config{
		Type:     particles.ParticleSmokePlume,
		Count:    3,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 3.0,
		SpreadX:  25.0,
		SpreadY:  15.0,
		Gravity:  -4.0,
		MinSize:  4.0,
		MaxSize:  8.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"ambient_deep_mist": true},
	}
}

// getLavaEmberConfig returns rising ember particles near lava.
func (s *AmbientEnvironmentParticleSystem) getLavaEmberConfig(seed int64) *particles.Config {
	return &particles.Config{
		Type:     particles.ParticleEmber,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 1.5,
		SpreadX:  15.0,
		SpreadY:  20.0,
		Gravity:  -40.0, // Embers rise quickly
		MinSize:  2.0,
		MaxSize:  5.0,
		ZLayer:   particles.ZLayerSky,
		Custom:   map[string]interface{}{"ambient_ember": true},
	}
}

// getForestAmbientConfig returns genre-aware forest ambient particles.
func (s *AmbientEnvironmentParticleSystem) getForestAmbientConfig(seed int64) *particles.Config {
	pType := particles.ParticleSparkle
	count := 2
	gravity := -8.0
	duration := 3.0

	switch s.genreID {
	case "horror":
		pType = particles.ParticleSmoke // Creepy fog wisps
		gravity = -2.0
		duration = 4.0
	case "scifi":
		pType = particles.ParticleSpark // Data fragments
		count = 1
		gravity = -15.0
	case "cyberpunk":
		pType = particles.ParticleSpark // Neon motes
		count = 1
		gravity = -10.0
	}

	return &particles.Config{
		Type:     pType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: duration,
		SpreadX:  18.0,
		SpreadY:  18.0,
		Gravity:  gravity,
		MinSize:  1.0,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"ambient_forest": true},
	}
}

// getBridgeMistConfig returns subtle mist wisps for bridge areas.
func (s *AmbientEnvironmentParticleSystem) getBridgeMistConfig(seed int64) *particles.Config {
	return &particles.Config{
		Type:     particles.ParticleSmoke,
		Count:    1,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 2.0,
		SpreadX:  15.0,
		SpreadY:  8.0,
		Gravity:  -3.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"ambient_bridge_mist": true},
	}
}

// getStructureDustConfig returns dust particles for structure/ruins areas.
func (s *AmbientEnvironmentParticleSystem) getStructureDustConfig(seed int64) *particles.Config {
	pType := particles.ParticleDust
	if s.genreID == "cyberpunk" || s.genreID == "scifi" {
		pType = particles.ParticleSpark
	}
	return &particles.Config{
		Type:     pType,
		Count:    1,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 1.8,
		SpreadX:  10.0,
		SpreadY:  10.0,
		Gravity:  5.0, // Settling dust
		MinSize:  1.0,
		MaxSize:  3.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"ambient_structure_dust": true},
	}
}
