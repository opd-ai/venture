// Package engine provides the BlockParticleSystem for visual block feedback.
// This system connects CombatSystem block events with ParticleSystem to spawn
// genre-aware particle effects when entities block attacks, giving players
// immediate visual feedback that their block stats are working.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// BlockParticleSystem spawns particle effects when attacks are blocked.
// It connects CombatSystem block checks with ParticleSystem to provide
// visual feedback for successful blocks, with genre-aware particle colors.
type BlockParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	particleCount int
	spreadFactor  float64
}

// NewBlockParticleSystem creates a new block particle system.
func NewBlockParticleSystem(world *World, seed int64) *BlockParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "block_particle")
		logEntry.Debug("block particle system created")
	}

	return &BlockParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		particleCount: 8,    // Medium particle count for shield impact
		spreadFactor:  40.0, // Spread radius for block particles
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *BlockParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *BlockParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *BlockParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnBlock, no per-frame processing needed
}

// OnBlock is called when an attack is blocked to spawn particles.
// This method should be registered as a callback with the CombatSystem.
//
// Parameters:
//   - attacker: The entity that attacked
//   - target: The entity that blocked the attack
//   - blockChance: The block chance that resulted in the block
//   - originalDamage: The damage before block reduction
//   - reducedDamage: The damage after block reduction (50% of original)
func (s *BlockParticleSystem) OnBlock(
	attacker, target *Entity,
	blockChance, originalDamage, reducedDamage float64,
) {
	if s.particleSystem == nil || s.world == nil || target == nil {
		return
	}

	// Get target position for particle spawn
	pos := target.GetPosition()
	if pos == nil {
		return
	}

	// Spawn block particles at target location
	s.spawnBlockParticles(pos.X, pos.Y, blockChance, originalDamage)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"attacker_id":     attacker.ID,
			"target_id":       target.ID,
			"block_chance":    blockChance,
			"original_damage": originalDamage,
			"reduced_damage":  reducedDamage,
			"x":               pos.X,
			"y":               pos.Y,
		}).Debug("block particles spawned")
	}
}

// spawnBlockParticles creates the block particle effect.
func (s *BlockParticleSystem) spawnBlockParticles(
	x, y, blockChance, originalDamage float64,
) {
	// Scale particle count with damage blocked
	count := s.particleCount
	if originalDamage > 20 {
		count = int(float64(count) * 1.3)
	}
	if originalDamage > 50 {
		count = int(float64(count) * 1.5)
	}
	// Cap at reasonable maximum
	if count > 16 {
		count = 16
	}

	// Use deterministic seed offset based on position
	effectSeed := s.seed + int64(x*1000) + int64(y*1000)

	// Select particle type based on genre
	particleType := s.getParticleTypeForGenre()

	// Create particle config
	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.4, // Slightly longer than evasion for impact feel
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.6, // Less vertical spread for shield wall effect
		Gravity:  20.0,                 // Slight downward drift for weight
		MinSize:  3.0,
		MaxSize:  6.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as block effect for color selection
	config.Custom["block_effect"] = true
	config.Custom["block_chance"] = blockChance
	config.Custom["damage_blocked"] = originalDamage * 0.5

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *BlockParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Shield magic sparks
	case "scifi":
		return particles.ParticleSpark // Energy shield deflection
	case "horror":
		return particles.ParticleSmoke // Dark ward effect
	case "cyberpunk":
		return particles.ParticleSpark // Hardlight barrier
	case "postapoc":
		return particles.ParticleDebris // Scrap metal deflection
	default:
		return particles.ParticleSparkle
	}
}

// SpawnBlockEffect allows external systems to trigger block particles directly.
// Useful for abilities that grant temporary block visualization.
func (s *BlockParticleSystem) SpawnBlockEffect(x, y float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnBlockParticles(x, y, 0.2, 15.0) // Default moderate block
}
