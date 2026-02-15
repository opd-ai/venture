// Package engine provides the MovementDustSystem which spawns speed-proportional
// dust and debris particles behind fast-moving entities based on terrain type.
// Terrain determines particle style: floors kick up dust, corridors scatter
// grit, bridges produce wood chips, lava surfaces emit ember sparks.
// Genre-aware styling: fantasy=earth tones, scifi=energy sparks,
// horror=ash/shadow wisps, cyberpunk=neon-grit, postapoc=rubble dust.
// Connects velocity magnitude with terrain lookup and ParticleSystem.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// MovementDustComponent tracks per-entity dust spawning state.
type MovementDustComponent struct {
	Cooldown     float64 // Time remaining before next dust spawn
	LastSpeedSq  float64 // Cached speed² from last frame
	Intensity    float64 // Current dust intensity (0-1) based on speed
	Suppressed   bool    // Temporarily suppressed (e.g. flying entities)
}

// Type returns the component type identifier.
func (c *MovementDustComponent) Type() string { return "movement_dust" }

// movementDustGenrePreset holds genre-specific dust visual parameters.
type movementDustGenrePreset struct {
	particleType particles.ParticleType
	minSize      float64
	maxSize      float64
	duration     float64
	gravity      float64
	spreadMul    float64 // Multiplier for velocity-based spread
}

// MovementDustSystem spawns terrain-aware dust particles behind fast-moving entities.
type MovementDustSystem struct {
	world          *World
	particleSystem *ParticleSystem
	terrainData    *terrain.Terrain
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry
	tileSize       int
	preset         movementDustGenrePreset

	// Speed thresholds (squared for fast comparison)
	minSpeedSq float64 // Below this: no dust
	maxSpeedSq float64 // Above this: maximum intensity

	// Spawn rate limits
	baseCooldown float64 // Cooldown at max speed
	maxCooldown  float64 // Cooldown at min speed
}

// NewMovementDustSystem creates a new movement dust system.
func NewMovementDustSystem(world *World, seed int64) *MovementDustSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "movement_dust")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("movement dust system created")
		}
	}

	s := &MovementDustSystem{
		world:        world,
		seed:         seed,
		rng:          rand.New(rand.NewSource(seed)),
		logger:       logEntry,
		tileSize:     32,
		minSpeedSq:   2500.0, // 50 px/s threshold
		maxSpeedSq:   40000.0, // 200 px/s max
		baseCooldown: 0.06,   // ~16 spawns/sec at max speed
		maxCooldown:  0.25,   // ~4 spawns/sec at min speed
	}
	s.preset = s.getGenrePreset("fantasy")
	return s
}

// SetParticleSystem sets the particle system for spawning effects.
func (s *MovementDustSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetTerrain sets the terrain data for tile lookups.
func (s *MovementDustSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrainData = terr
}

// SetGenre sets genre for visual styling.
func (s *MovementDustSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetTileSize sets the tile size in pixels.
func (s *MovementDustSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// Update processes entities and spawns dust particles for fast movers.
func (s *MovementDustSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}
}

// processEntity evaluates one entity for dust spawning.
func (s *MovementDustSystem) processEntity(entity *Entity, deltaTime float64) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}
	vel := entity.GetVelocity()
	if vel == nil {
		return
	}

	speedSq := vel.VX*vel.VX + vel.VY*vel.VY
	if speedSq < s.minSpeedSq {
		return
	}

	// Get or attach component
	dustComp := s.getDustComponent(entity)
	if dustComp.Suppressed {
		return
	}

	// Calculate intensity: 0 at minSpeed, 1 at maxSpeed
	intensity := (speedSq - s.minSpeedSq) / (s.maxSpeedSq - s.minSpeedSq)
	if intensity > 1.0 {
		intensity = 1.0
	}
	dustComp.Intensity = intensity
	dustComp.LastSpeedSq = speedSq

	// Cooldown check
	dustComp.Cooldown -= deltaTime
	if dustComp.Cooldown > 0 {
		return
	}

	// Set next cooldown inversely proportional to speed
	dustComp.Cooldown = s.maxCooldown - intensity*(s.maxCooldown-s.baseCooldown)

	// Determine terrain-based particle config
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	var tileType terrain.TileType
	if s.terrainData != nil {
		tileType = s.terrainData.GetTile(tileX, tileY)
	} else {
		tileType = terrain.TileFloor
	}

	if !s.isWalkableDustTerrain(tileType) {
		return
	}

	// Spawn behind entity (opposite of movement direction)
	speed := math.Sqrt(speedSq)
	dirX := -vel.VX / speed
	dirY := -vel.VY / speed
	spawnX := pos.X + dirX*4.0
	spawnY := pos.Y + dirY*4.0

	s.spawnDust(entity.ID, spawnX, spawnY, intensity, tileType, dirX, dirY)
}

// getDustComponent retrieves or creates the dust component for an entity.
func (s *MovementDustSystem) getDustComponent(entity *Entity) *MovementDustComponent {
	if comp, ok := entity.GetComponent("movement_dust"); ok {
		return comp.(*MovementDustComponent)
	}
	c := &MovementDustComponent{}
	entity.AddComponent(c)
	return c
}

// isWalkableDustTerrain returns true for terrain types that produce dust.
func (s *MovementDustSystem) isWalkableDustTerrain(t terrain.TileType) bool {
	switch t {
	case terrain.TileFloor, terrain.TileCorridor, terrain.TileDoor,
		terrain.TileBridge, terrain.TileRamp, terrain.TileRampUp,
		terrain.TileRampDown, terrain.TilePlatform, terrain.TileStairsUp,
		terrain.TileStairsDown, terrain.TileLavaFlow:
		return true
	default:
		return false
	}
}

// spawnDust creates the actual dust particles.
func (s *MovementDustSystem) spawnDust(entityID uint64, x, y, intensity float64, tileType terrain.TileType, dirX, dirY float64) {
	pType, duration, gravity := s.getTerrainParticleStyle(tileType)

	count := 1 + int(intensity*3) // 1-4 particles
	effectSeed := s.seed + int64(x*73) + int64(y*97) + int64(entityID)

	config := particles.Config{
		Type:    pType,
		Count:   count,
		GenreID: s.genreID,
		Seed:    effectSeed,
		Duration: duration * (0.6 + intensity*0.4),
		SpreadX:  (3.0 + intensity*5.0) * s.preset.spreadMul,
		SpreadY:  (3.0 + intensity*5.0) * s.preset.spreadMul,
		Gravity:  gravity,
		MinSize:  s.preset.minSize,
		MaxSize:  s.preset.minSize + (s.preset.maxSize-s.preset.minSize)*intensity,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"movement_dust": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getTerrainParticleStyle returns particle type, duration, and gravity for a terrain.
func (s *MovementDustSystem) getTerrainParticleStyle(t terrain.TileType) (particles.ParticleType, float64, float64) {
	switch t {
	case terrain.TileLavaFlow:
		return particles.ParticleEmber, 0.5, -5.0 // Embers float up
	case terrain.TileBridge:
		return particles.ParticleDebris, 0.4, 15.0 // Wood chips fall
	case terrain.TileStairsUp, terrain.TileStairsDown:
		return particles.ParticleSpark, 0.3, 10.0 // Grit sparks on stone steps
	default:
		return s.preset.particleType, s.preset.duration, s.preset.gravity
	}
}

// getGenrePreset returns visual parameters for a genre.
func (s *MovementDustSystem) getGenrePreset(genreID string) movementDustGenrePreset {
	switch genreID {
	case "horror":
		return movementDustGenrePreset{
			particleType: particles.ParticleSmoke,
			minSize: 3.0, maxSize: 7.0,
			duration: 0.8, gravity: -2.0, spreadMul: 1.3,
		}
	case "scifi":
		return movementDustGenrePreset{
			particleType: particles.ParticleSpark,
			minSize: 2.0, maxSize: 5.0,
			duration: 0.4, gravity: 0.0, spreadMul: 1.5,
		}
	case "cyberpunk":
		return movementDustGenrePreset{
			particleType: particles.ParticleSpark,
			minSize: 2.0, maxSize: 6.0,
			duration: 0.5, gravity: 3.0, spreadMul: 1.2,
		}
	case "postapoc":
		return movementDustGenrePreset{
			particleType: particles.ParticleDebris,
			minSize: 3.0, maxSize: 8.0,
			duration: 0.7, gravity: 12.0, spreadMul: 1.0,
		}
	default: // fantasy
		return movementDustGenrePreset{
			particleType: particles.ParticleDust,
			minSize: 3.0, maxSize: 6.0,
			duration: 0.6, gravity: 8.0, spreadMul: 1.0,
		}
	}
}
