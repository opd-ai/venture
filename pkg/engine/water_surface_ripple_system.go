// Package engine provides the WaterSurfaceRippleSystem which spawns ripple and
// splash particles when entities move through shallow water tiles.
// Speed determines ripple intensity: slow wading produces gentle concentric rings,
// fast movement produces spray splashes with upward velocity.
// Genre-aware styling: fantasy=blue sparkles, scifi=energy ripples,
// horror=dark murky wisps, cyberpunk=neon reflections, postapoc=murky splashes.
// Connects velocity magnitude with terrain tile lookup and ParticleSystem.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WaterSurfaceRippleComponent tracks per-entity water ripple spawning state.
type WaterSurfaceRippleComponent struct {
	Cooldown  float64 // Time remaining before next ripple spawn
	Intensity float64 // Current ripple intensity (0-1) based on speed
	InWater   bool    // Whether entity is currently on a water tile
}

// Type returns the component type identifier.
func (c *WaterSurfaceRippleComponent) Type() string { return "water_surface_ripple" }

// waterRippleGenrePreset holds genre-specific water ripple visual parameters.
type waterRippleGenrePreset struct {
	particleType particles.ParticleType
	minSize      float64
	maxSize      float64
	duration     float64
	gravity      float64
	spreadMul    float64
}

// WaterSurfaceRippleSystem spawns ripple/splash particles for entities in water.
type WaterSurfaceRippleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	terrainData    *terrain.Terrain
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry
	tileSize       int
	preset         waterRippleGenrePreset

	// Speed thresholds (squared for fast comparison)
	minSpeedSq float64 // Below this: gentle idle ripple at slow rate
	maxSpeedSq float64 // Above this: maximum splash intensity

	// Spawn rate limits
	baseCooldown float64 // Cooldown at max speed
	maxCooldown  float64 // Cooldown at min speed / idle
}

// NewWaterSurfaceRippleSystem creates a new water surface ripple system.
func NewWaterSurfaceRippleSystem(world *World, seed int64) *WaterSurfaceRippleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "water_surface_ripple")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("water surface ripple system created")
		}
	}

	s := &WaterSurfaceRippleSystem{
		world:        world,
		seed:         seed,
		rng:          rand.New(rand.NewSource(seed)),
		logger:       logEntry,
		tileSize:     32,
		minSpeedSq:   400.0,   // 20 px/s — walking pace
		maxSpeedSq:   40000.0, // 200 px/s — full sprint
		baseCooldown: 0.08,    // ~12 spawns/sec at max speed
		maxCooldown:  0.5,     // ~2 spawns/sec when barely moving / idle
	}
	s.preset = s.getGenrePreset("fantasy")
	return s
}

// SetParticleSystem sets the particle system for spawning effects.
func (s *WaterSurfaceRippleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetTerrain sets the terrain data for tile lookups.
func (s *WaterSurfaceRippleSystem) SetTerrain(terr *terrain.Terrain) {
	s.terrainData = terr
}

// SetGenre sets genre for visual styling.
func (s *WaterSurfaceRippleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetTileSize sets the tile size in pixels.
func (s *WaterSurfaceRippleSystem) SetTileSize(size int) {
	if size > 0 {
		s.tileSize = size
	}
}

// Update processes entities and spawns ripple particles for those in water.
func (s *WaterSurfaceRippleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}
}

// processEntity evaluates one entity for water ripple spawning.
func (s *WaterSurfaceRippleSystem) processEntity(entity *Entity, deltaTime float64) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Determine tile under entity
	tileX := int(pos.X) / s.tileSize
	tileY := int(pos.Y) / s.tileSize

	var tileType terrain.TileType
	if s.terrainData != nil {
		tileType = s.terrainData.GetTile(tileX, tileY)
	} else {
		tileType = terrain.TileFloor
	}

	if !s.isWaterTile(tileType) {
		// Clear state if entity left water
		if comp, ok := entity.GetComponent("water_surface_ripple"); ok {
			comp.(*WaterSurfaceRippleComponent).InWater = false
		}
		return
	}

	// Get or attach component
	rippleComp := s.getRippleComponent(entity)

	// Mark entry into water
	wasInWater := rippleComp.InWater
	rippleComp.InWater = true

	// Entry splash on first frame in water
	if !wasInWater {
		s.spawnEntrySplash(entity.ID, pos.X, pos.Y, tileType)
		rippleComp.Cooldown = s.baseCooldown
		return
	}

	// Calculate speed
	vel := entity.GetVelocity()
	var speedSq float64
	if vel != nil {
		speedSq = vel.VX*vel.VX + vel.VY*vel.VY
	}

	// Compute intensity: idle entities still produce gentle ripples (intensity 0.1)
	var intensity float64
	if speedSq < s.minSpeedSq {
		intensity = 0.1
	} else {
		intensity = 0.1 + 0.9*(speedSq-s.minSpeedSq)/(s.maxSpeedSq-s.minSpeedSq)
		if intensity > 1.0 {
			intensity = 1.0
		}
	}
	rippleComp.Intensity = intensity

	// Cooldown check
	rippleComp.Cooldown -= deltaTime
	if rippleComp.Cooldown > 0 {
		return
	}

	// Set next cooldown inversely proportional to intensity
	rippleComp.Cooldown = s.maxCooldown - intensity*(s.maxCooldown-s.baseCooldown)

	s.spawnRipple(entity.ID, pos.X, pos.Y, intensity, tileType)
}

// getRippleComponent retrieves or creates the ripple component for an entity.
func (s *WaterSurfaceRippleSystem) getRippleComponent(entity *Entity) *WaterSurfaceRippleComponent {
	if comp, ok := entity.GetComponent("water_surface_ripple"); ok {
		return comp.(*WaterSurfaceRippleComponent)
	}
	c := &WaterSurfaceRippleComponent{}
	entity.AddComponent(c)
	return c
}

// isWaterTile returns true for water terrain types.
func (s *WaterSurfaceRippleSystem) isWaterTile(t terrain.TileType) bool {
	return t == terrain.TileWaterShallow || t == terrain.TileWaterDeep
}

// spawnEntrySplash creates a burst of particles when an entity first enters water.
func (s *WaterSurfaceRippleSystem) spawnEntrySplash(entityID uint64, x, y float64, tileType terrain.TileType) {
	effectSeed := s.seed + int64(x*53) + int64(y*79) + int64(entityID)
	count := 5
	if tileType == terrain.TileWaterDeep {
		count = 8
	}

	config := particles.Config{
		Type:     s.preset.particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.preset.duration * 1.2,
		SpreadX:  8.0 * s.preset.spreadMul,
		SpreadY:  8.0 * s.preset.spreadMul,
		Gravity:  s.preset.gravity,
		MinSize:  s.preset.minSize,
		MaxSize:  s.preset.maxSize,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"water_ripple": true, "entry_splash": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnRipple creates ongoing ripple particles.
func (s *WaterSurfaceRippleSystem) spawnRipple(entityID uint64, x, y, intensity float64, tileType terrain.TileType) {
	effectSeed := s.seed + int64(x*61) + int64(y*89) + int64(entityID)
	count := 1 + int(intensity*3) // 1-4 particles

	// Deep water produces larger, slower ripples
	durationMul := 1.0
	sizeMul := 1.0
	if tileType == terrain.TileWaterDeep {
		durationMul = 1.3
		sizeMul = 1.2
	}

	config := particles.Config{
		Type:     s.preset.particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.preset.duration * durationMul * (0.5 + intensity*0.5),
		SpreadX:  (4.0 + intensity*6.0) * s.preset.spreadMul,
		SpreadY:  (4.0 + intensity*6.0) * s.preset.spreadMul,
		Gravity:  s.preset.gravity,
		MinSize:  s.preset.minSize * sizeMul,
		MaxSize:  (s.preset.minSize + (s.preset.maxSize-s.preset.minSize)*intensity) * sizeMul,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"water_ripple": true},
	}

	// Offset spawn slightly outward from entity center for realism
	rng := rand.New(rand.NewSource(effectSeed))
	offsetAngle := rng.Float64() * 2 * math.Pi
	offsetDist := 2.0 + intensity*4.0
	spawnX := x + math.Cos(offsetAngle)*offsetDist
	spawnY := y + math.Sin(offsetAngle)*offsetDist

	s.particleSystem.SpawnParticles(s.world, config, spawnX, spawnY)
}

// getGenrePreset returns visual parameters for a genre.
func (s *WaterSurfaceRippleSystem) getGenrePreset(genreID string) waterRippleGenrePreset {
	switch genreID {
	case "horror":
		return waterRippleGenrePreset{
			particleType: particles.ParticleSmoke,
			minSize:      3.0, maxSize: 7.0,
			duration: 1.0, gravity: -1.0, spreadMul: 1.2,
		}
	case "scifi":
		return waterRippleGenrePreset{
			particleType: particles.ParticleSpark,
			minSize:      2.0, maxSize: 5.0,
			duration: 0.6, gravity: -0.5, spreadMul: 1.4,
		}
	case "cyberpunk":
		return waterRippleGenrePreset{
			particleType: particles.ParticleSpark,
			minSize:      2.0, maxSize: 6.0,
			duration: 0.7, gravity: -0.5, spreadMul: 1.3,
		}
	case "postapoc":
		return waterRippleGenrePreset{
			particleType: particles.ParticleSmokePlume,
			minSize:      3.0, maxSize: 8.0,
			duration: 0.9, gravity: -1.0, spreadMul: 1.0,
		}
	default: // fantasy
		return waterRippleGenrePreset{
			particleType: particles.ParticleSparkle,
			minSize:      3.0, maxSize: 6.0,
			duration: 0.8, gravity: -2.0, spreadMul: 1.0,
		}
	}
}
