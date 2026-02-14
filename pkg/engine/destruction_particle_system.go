// Package engine provides the DestructionParticleSystem for visual destruction feedback.
// This system connects DestructibleObjectSystem with ParticleSystem to spawn genre-aware
// particle effects when destructible objects are destroyed, enhancing environmental
// destruction visual feedback.
package engine

import (
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// DestructionParticleSystem spawns particle effects when destructible objects are destroyed.
// It provides visual feedback for explosions, debris scattering, and poison clouds
// with genre-aware particle colors and effect variations.
type DestructionParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	debrisParticleCount    int
	explosionParticleCount int
	poisonParticleCount    int
	spreadFactor           float64
	effectDuration         float64
}

// NewDestructionParticleSystem creates a new destruction particle system.
func NewDestructionParticleSystem(world *World, seed int64) *DestructionParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "destruction_particle")
		logEntry.Debug("destruction particle system created")
	}

	return &DestructionParticleSystem{
		world:                  world,
		genreID:                "fantasy",
		seed:                   seed,
		rng:                    rand.New(rand.NewSource(seed)),
		logger:                 logEntry,
		debrisParticleCount:    15,
		explosionParticleCount: 30,
		poisonParticleCount:    20,
		spreadFactor:           120.0,
		effectDuration:         0.6,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *DestructionParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *DestructionParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes destructible objects and spawns particles for newly destroyed ones.
func (s *DestructionParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		comp, ok := entity.GetComponent("destructibleObject")
		if !ok {
			continue
		}

		destructibleObj, ok := comp.(*DestructibleObjectComponent)
		if !ok {
			continue
		}

		// Check for newly destroyed objects that haven't had particles spawned yet
		if destructibleObj.IsDestroyed && !destructibleObj.ParticlesSpawned {
			s.spawnDestructionParticles(entity, destructibleObj)
			destructibleObj.ParticlesSpawned = true
		}
	}
}

// spawnDestructionParticles creates particles based on the object type.
func (s *DestructionParticleSystem) spawnDestructionParticles(entity *Entity, destructibleObj *DestructibleObjectComponent) {
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	x, y := pos.X, pos.Y
	objType := destructibleObj.ObjectType

	// Spawn type-specific particles
	switch objType {
	case ObjectExplosiveBarrel:
		s.spawnExplosionParticles(x, y, destructibleObj.ExplosionRadius)
	case ObjectPoisonContainer:
		s.spawnPoisonParticles(x, y, destructibleObj.PoisonRadius)
	default:
		s.spawnDebrisParticles(x, y, objType)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"object_type": objType.String(),
			"x":           x,
			"y":           y,
		}).Debug("destruction particles spawned")
	}
}

// spawnExplosionParticles creates explosion effect particles.
func (s *DestructionParticleSystem) spawnExplosionParticles(x, y, radius float64) {
	count := s.explosionParticleCount
	spread := radius * 1.5
	if spread < s.spreadFactor {
		spread = s.spreadFactor
	}

	// Seed for this specific explosion
	seedOffset := int64(x*1000 + y*31)
	localRng := rand.New(rand.NewSource(s.seed + seedOffset))

	// Get genre-aware explosion colors
	primary, secondary := s.getExplosionColors()

	// Use deterministic seed for explosion effect
	effectSeed := s.seed + int64(x*1000) + int64(y*31)
	_ = localRng // Used for seed generation

	// Pick primary color for this explosion
	_ = primary
	_ = secondary

	// Create explosion config using particles.Config
	config := particles.Config{
		Type:     particles.ParticleEmber,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  spread,
		SpreadY:  spread,
		Gravity:  50.0, // Explosion debris falls
		MinSize:  4.0,
		MaxSize:  10.0,
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnPoisonParticles creates poison cloud effect particles.
func (s *DestructionParticleSystem) spawnPoisonParticles(x, y, radius float64) {
	count := s.poisonParticleCount
	spread := radius * 1.2
	if spread < s.spreadFactor*0.5 {
		spread = s.spreadFactor * 0.5
	}

	seedOffset := int64(x*1000 + y*37)
	localRng := rand.New(rand.NewSource(s.seed + seedOffset))

	// Get genre-aware poison colors
	primary, secondary := s.getPoisonColors()

	_ = primary
	_ = secondary
	_ = localRng

	// Use deterministic seed for poison effect
	effectSeed := s.seed + int64(x*1000) + int64(y*37)

	// Create poison cloud config using particles.Config
	config := particles.Config{
		Type:     particles.ParticleSmokePlume,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 1.5,
		SpreadX:  spread,
		SpreadY:  spread,
		Gravity:  -10.0, // Poison rises slowly
		MinSize:  6.0,
		MaxSize:  14.0,
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnDebrisParticles creates debris scattering particles.
func (s *DestructionParticleSystem) spawnDebrisParticles(x, y float64, objType ObjectType) {
	count := s.debrisParticleCount
	spread := s.spreadFactor

	seedOffset := int64(x*1000 + y*41)
	localRng := rand.New(rand.NewSource(s.seed + seedOffset))

	// Get colors based on object type
	primary, secondary := s.getDebrisColors(objType)

	_ = primary
	_ = secondary
	_ = localRng
	_ = objType

	// Use deterministic seed for debris effect
	effectSeed := s.seed + int64(x*1000) + int64(y*41)

	// Create debris config using particles.Config
	config := particles.Config{
		Type:     particles.ParticleDebris,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  spread,
		SpreadY:  spread,
		Gravity:  150.0, // Debris falls quickly
		MinSize:  3.0,
		MaxSize:  7.0,
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getExplosionColors returns genre-aware explosion particle colors.
func (s *DestructionParticleSystem) getExplosionColors() (primary, secondary color.RGBA) {
	switch s.genreID {
	case "fantasy":
		return color.RGBA{255, 150, 50, 255}, color.RGBA{255, 80, 20, 255} // Orange/red fire
	case "scifi":
		return color.RGBA{100, 200, 255, 255}, color.RGBA{200, 100, 255, 255} // Blue/purple plasma
	case "horror":
		return color.RGBA{180, 50, 50, 255}, color.RGBA{80, 20, 20, 255} // Dark red/crimson
	case "cyberpunk":
		return color.RGBA{255, 100, 200, 255}, color.RGBA{100, 255, 200, 255} // Neon pink/cyan
	case "postapoc":
		return color.RGBA{200, 150, 100, 255}, color.RGBA{150, 100, 50, 255} // Dusty brown/tan
	default:
		return color.RGBA{255, 150, 50, 255}, color.RGBA{255, 80, 20, 255} // Default orange fire
	}
}

// getPoisonColors returns genre-aware poison cloud particle colors.
func (s *DestructionParticleSystem) getPoisonColors() (primary, secondary color.RGBA) {
	switch s.genreID {
	case "fantasy":
		return color.RGBA{100, 200, 80, 200}, color.RGBA{50, 150, 50, 180} // Green poison
	case "scifi":
		return color.RGBA{150, 255, 150, 200}, color.RGBA{100, 200, 100, 180} // Bright green bio
	case "horror":
		return color.RGBA{80, 120, 50, 220}, color.RGBA{40, 80, 30, 200} // Sickly dark green
	case "cyberpunk":
		return color.RGBA{0, 255, 150, 200}, color.RGBA{100, 255, 200, 180} // Neon toxic green
	case "postapoc":
		return color.RGBA{150, 180, 50, 200}, color.RGBA{100, 130, 40, 180} // Radiation yellow-green
	default:
		return color.RGBA{100, 200, 80, 200}, color.RGBA{50, 150, 50, 180} // Default green
	}
}

// getDebrisColors returns colors based on object type and genre.
func (s *DestructionParticleSystem) getDebrisColors(objType ObjectType) (primary, secondary color.RGBA) {
	// Base colors for object types
	switch objType {
	case ObjectCrate:
		return color.RGBA{160, 120, 70, 255}, color.RGBA{120, 80, 40, 255} // Wood brown
	case ObjectBarrel:
		return color.RGBA{140, 100, 60, 255}, color.RGBA{100, 70, 40, 255} // Dark wood
	case ObjectFurniture:
		return color.RGBA{150, 110, 65, 255}, color.RGBA{110, 80, 45, 255} // Furniture wood
	case ObjectWeakWall:
		return s.getWallDebrisColors() // Genre-dependent wall materials
	default:
		return color.RGBA{130, 130, 130, 255}, color.RGBA{90, 90, 90, 255} // Generic gray
	}
}

// getWallDebrisColors returns genre-aware wall debris colors.
func (s *DestructionParticleSystem) getWallDebrisColors() (primary, secondary color.RGBA) {
	switch s.genreID {
	case "fantasy":
		return color.RGBA{150, 140, 120, 255}, color.RGBA{100, 95, 80, 255} // Stone gray
	case "scifi":
		return color.RGBA{180, 180, 200, 255}, color.RGBA{140, 140, 160, 255} // Metal gray-blue
	case "horror":
		return color.RGBA{80, 70, 65, 255}, color.RGBA{50, 45, 40, 255} // Dark stone
	case "cyberpunk":
		return color.RGBA{60, 60, 70, 255}, color.RGBA{40, 40, 50, 255} // Dark metal
	case "postapoc":
		return color.RGBA{140, 130, 110, 255}, color.RGBA{100, 90, 75, 255} // Crumbling concrete
	default:
		return color.RGBA{150, 140, 120, 255}, color.RGBA{100, 95, 80, 255} // Default stone
	}
}
