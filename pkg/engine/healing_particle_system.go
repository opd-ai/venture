// Package engine provides the HealingParticleSystem for visual healing feedback.
// This system connects CombatSystem heal events with ParticleSystem to spawn
// genre-aware particle effects when entities receive healing, giving players
// immediate visual feedback that healing spells and abilities are taking effect.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// HealingParticleSystem spawns particle effects when entities receive healing.
// It connects CombatSystem heal callbacks with ParticleSystem to provide
// visual feedback when healing occurs, with genre-aware particle colors.
type HealingParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	particleCount    int
	spreadFactor     float64
	minHealAmount    float64 // Minimum heal to trigger particles
	particlesPerHeal float64 // Particles per point of healing (scaled)
}

// NewHealingParticleSystem creates a new healing particle system.
func NewHealingParticleSystem(world *World, seed int64) *HealingParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "healing_particle")
		logEntry.Debug("healing particle system created")
	}

	return &HealingParticleSystem{
		world:            world,
		seed:             seed,
		rng:              rand.New(rand.NewSource(seed)),
		logger:           logEntry,
		particleCount:    8,    // Base particle count for healing effects
		spreadFactor:     50.0, // Spread radius for healing particles
		minHealAmount:    1.0,  // Minimum 1 HP to trigger particles
		particlesPerHeal: 0.15, // Scale factor for larger heals
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *HealingParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *HealingParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetMinHealAmount sets the minimum heal amount to trigger particles.
func (s *HealingParticleSystem) SetMinHealAmount(amount float64) {
	if amount < 0 {
		amount = 0
	}
	s.minHealAmount = amount
}

// Update processes entities (no-op for this callback-driven system).
func (s *HealingParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnHeal, no per-frame processing needed
}

// OnHeal is called when an entity receives healing to spawn particles.
// This method should be registered as a callback with the CombatSystem.
//
// Parameters:
//   - healer: The entity that performed the heal (may be same as target for self-heal)
//   - target: The entity that received healing
//   - amount: The amount of HP healed
func (s *HealingParticleSystem) OnHeal(healer, target *Entity, amount float64) {
	if s.particleSystem == nil || s.world == nil || target == nil {
		return
	}

	// Check if heal amount meets minimum threshold
	if amount < s.minHealAmount {
		return
	}

	// Get target position for particle spawn
	pos := target.GetPosition()
	if pos == nil {
		return
	}

	// Spawn healing particles at target location
	s.spawnHealingParticles(pos.X, pos.Y, amount)

	if s.logger != nil {
		isSelfHeal := healer != nil && healer.ID == target.ID
		fields := logrus.Fields{
			"target_id": target.ID,
			"amount":    amount,
			"x":         pos.X,
			"y":         pos.Y,
			"self_heal": isSelfHeal,
		}
		if healer != nil {
			fields["healer_id"] = healer.ID
		}
		s.logger.WithFields(fields).Debug("healing particles spawned")
	}
}

// spawnHealingParticles creates the healing particle effect.
func (s *HealingParticleSystem) spawnHealingParticles(x, y, amount float64) {
	// Scale particle count based on heal amount
	count := s.particleCount + int(amount*s.particlesPerHeal)

	// Apply caps for performance
	if count < 4 {
		count = 4
	}
	if count > 20 {
		count = 20
	}

	// Use deterministic seed offset for this heal event
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(amount*100)

	// Create healing particle config - rising green/blue magical particles
	config := particles.Config{
		Type:     particles.ParticleMagic,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 1.0, // Longer duration for soothing heal effect
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.5, // Tighter vertical spread
		Gravity:  -80.0,                // Float upward gently
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as healing effect for special rendering (green/blue tint based on genre)
	config.Custom["healing_effect"] = true
	config.Custom["heal_amount"] = amount

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getGenreHealingParticleType returns genre-appropriate particle behavior.
// This is used internally for genre-aware visual variations.
func (s *HealingParticleSystem) getGenreHealingParticleType() string {
	switch s.genreID {
	case "fantasy":
		return "magical_sparkles" // Green/gold sparkles rising
	case "scifi":
		return "nanobots" // Blue tech particles
	case "horror":
		return "blood_mist" // Dark red/purple mist
	case "cyberpunk":
		return "stim_injection" // Neon green pulse
	case "postapoc":
		return "medkit_spray" // White/red medical particles
	default:
		return "magical_sparkles"
	}
}
