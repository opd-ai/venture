// Package engine provides the ItemPickupParticleSystem for visual item pickup feedback.
// This system connects ItemPickupSystem with ParticleSystem to spawn genre-aware particle
// effects when items are collected, enhancing loot collection visual feedback.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ItemPickupParticleSystem spawns particle effects when items are picked up.
// It connects ItemPickupSystem and ParticleSystem to provide visual feedback
// with genre-aware particle colors and behaviors based on item rarity.
type ItemPickupParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	baseParticleCount int
	spreadFactor      float64
}

// NewItemPickupParticleSystem creates a new item pickup particle system.
func NewItemPickupParticleSystem(world *World, seed int64) *ItemPickupParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "item_pickup_particle")
		logEntry.Debug("item pickup particle system created")
	}

	return &ItemPickupParticleSystem{
		world:             world,
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		baseParticleCount: 12,   // Default particle count for pickup effects
		spreadFactor:      60.0, // Default spread for pickup particles
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ItemPickupParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ItemPickupParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *ItemPickupParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnItemPickup, no per-frame processing needed
}

// OnItemPickup is called when an item is picked up to spawn particle effects.
// This method should be registered as a callback with the ItemPickupSystem.
func (s *ItemPickupParticleSystem) OnItemPickup(x, y float64, rarity int) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	// Spawn pickup particles at item location
	s.spawnPickupParticles(x, y, rarity)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":      x,
			"y":      y,
			"rarity": rarity,
		}).Debug("item pickup particles spawned")
	}
}

// spawnPickupParticles creates the item pickup particle effect.
func (s *ItemPickupParticleSystem) spawnPickupParticles(x, y float64, rarity int) {
	// Scale particle count based on item rarity (more particles for rarer items)
	count := s.baseParticleCount
	switch {
	case rarity >= 4: // Legendary
		count = int(float64(count) * 2.5)
	case rarity >= 3: // Epic
		count = int(float64(count) * 2.0)
	case rarity >= 2: // Rare
		count = int(float64(count) * 1.5)
	case rarity >= 1: // Uncommon
		count = int(float64(count) * 1.25)
	}
	// Cap at reasonable maximum
	if count > 40 {
		count = 40
	}

	// Use deterministic seed offset for this specific pickup
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(rarity*100)

	// Create sparkle config for pickup effect - particles rise upward
	config := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.5,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.5, // More vertical spread
		Gravity:  -120.0,               // Float upward
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as item pickup for potential special rendering
	config.Custom["pickup_effect"] = true
	config.Custom["rarity"] = rarity

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// SpawnPickupEffect allows external systems to trigger pickup particles directly.
func (s *ItemPickupParticleSystem) SpawnPickupEffect(x, y float64, rarity int) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnPickupParticles(x, y, rarity)
}
