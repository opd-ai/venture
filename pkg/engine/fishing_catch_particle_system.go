// Package engine provides the FishingCatchParticleSystem for visual fish catch feedback.
// This system connects FishingSystem with ParticleSystem to spawn genre-aware particle
// effects when fish are caught, with effects scaling based on fish rarity.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// FishingCatchParticleSystem spawns particle effects when fish are caught.
// It connects FishingSystem callbacks with ParticleSystem to provide visual
// feedback with genre-aware particle colors and behaviors based on fish rarity.
type FishingCatchParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	fishingSystem  *FishingSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	baseParticleCount int
	spreadFactor      float64
}

// NewFishingCatchParticleSystem creates a new fishing catch particle system.
func NewFishingCatchParticleSystem(world *World, seed int64) *FishingCatchParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "fishing_catch_particle")
		logEntry.Debug("fishing catch particle system created")
	}

	return &FishingCatchParticleSystem{
		world:             world,
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		genreID:           "fantasy",
		baseParticleCount: 15,   // Default particle count for catch effects
		spreadFactor:      50.0, // Default spread for catch particles
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *FishingCatchParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetFishingSystem sets the fishing system and wires up the catch callback.
func (s *FishingCatchParticleSystem) SetFishingSystem(fs *FishingSystem) {
	s.fishingSystem = fs
	if fs != nil {
		// Chain with existing callback if present
		existingCallback := fs.OnCatchCallback
		fs.OnCatchCallback = func(fisher *Entity, caught *CaughtFish) {
			s.OnFishCaught(fisher, caught)
			if existingCallback != nil {
				existingCallback(fisher, caught)
			}
		}
		if s.logger != nil {
			s.logger.Debug("fishing system catch callback wired")
		}
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *FishingCatchParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *FishingCatchParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnFishCaught, no per-frame processing needed
}

// OnFishCaught is called when a fish is caught to spawn particle effects.
func (s *FishingCatchParticleSystem) OnFishCaught(fisher *Entity, caught *CaughtFish) {
	if s.particleSystem == nil || s.world == nil || fisher == nil || caught == nil {
		return
	}

	// Get fisher position
	posComp, ok := fisher.GetComponent("position")
	if !ok {
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok || pos == nil {
		return
	}

	// Get fish rarity from fishing system
	rarity := s.getFishRarity(caught.FishTypeID)

	// Spawn catch particles at fisher location
	s.spawnCatchParticles(pos.X, pos.Y, rarity, caught.IsRecord, caught.Weight)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"fisher_id": fisher.ID,
			"fish_type": caught.FishTypeID,
			"rarity":    rarity,
			"weight":    caught.Weight,
			"is_record": caught.IsRecord,
		}).Debug("fishing catch particles spawned")
	}
}

// getFishRarity returns the rarity level (0-4) for a fish type.
func (s *FishingCatchParticleSystem) getFishRarity(fishTypeID string) int {
	if s.fishingSystem == nil {
		return 0
	}

	fishType := s.fishingSystem.GetFishType(fishTypeID)
	if fishType == nil {
		return 0
	}

	switch fishType.Rarity {
	case FishRarityLegendary:
		return 4
	case FishRarityEpic:
		return 3
	case FishRarityRare:
		return 2
	case FishRarityUncommon:
		return 1
	default:
		return 0
	}
}

// spawnCatchParticles creates the fish catch particle effect.
func (s *FishingCatchParticleSystem) spawnCatchParticles(x, y float64, rarity int, isRecord bool, weight float64) {
	// Scale particle count based on fish rarity
	count := s.baseParticleCount
	switch rarity {
	case 4: // Legendary
		count = int(float64(count) * 3.0)
	case 3: // Epic
		count = int(float64(count) * 2.5)
	case 2: // Rare
		count = int(float64(count) * 2.0)
	case 1: // Uncommon
		count = int(float64(count) * 1.5)
	}

	// Bonus particles for record catches
	if isRecord {
		count = int(float64(count) * 1.3)
	}

	// Cap at reasonable maximum
	if count > 60 {
		count = 60
	}

	// Use deterministic seed offset for this specific catch
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(rarity*100) + int64(weight*10)

	// Select particle type based on rarity - water splash base with sparkles for rare
	particleType := s.selectParticleType(rarity)

	// Create water splash config for catch effect
	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.selectDuration(rarity),
		SpreadX:  s.spreadFactor * (1.0 + float64(rarity)*0.2),
		SpreadY:  s.spreadFactor * 0.6,
		Gravity:  s.selectGravity(rarity),
		MinSize:  2.0 + float64(rarity)*0.5,
		MaxSize:  5.0 + float64(rarity)*1.0,
		Custom:   make(map[string]interface{}),
	}

	// Store metadata for potential special rendering
	config.Custom["fish_catch"] = true
	config.Custom["rarity"] = rarity
	config.Custom["is_record"] = isRecord

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	// Spawn extra sparkle burst for epic+ catches
	if rarity >= 3 {
		s.spawnRareSparkles(x, y, rarity, effectSeed+1)
	}
}

// selectParticleType returns the appropriate particle type for fish rarity.
func (s *FishingCatchParticleSystem) selectParticleType(rarity int) particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		if rarity >= 3 {
			return particles.ParticleSparkle
		}
		return particles.ParticleDust // Water droplets effect
	case "scifi":
		if rarity >= 3 {
			return particles.ParticleMagic // Energy effect
		}
		return particles.ParticleDust
	case "horror":
		return particles.ParticleSmoke // Mist effect
	case "cyberpunk":
		if rarity >= 2 {
			return particles.ParticleMagic // Neon energy effect
		}
		return particles.ParticleDust
	case "postapoc":
		return particles.ParticleDust
	default:
		return particles.ParticleDust
	}
}

// selectDuration returns particle duration based on rarity.
func (s *FishingCatchParticleSystem) selectDuration(rarity int) float64 {
	baseDuration := 0.6
	return baseDuration + float64(rarity)*0.15
}

// selectGravity returns gravity based on rarity (rarer fish have more upward float).
func (s *FishingCatchParticleSystem) selectGravity(rarity int) float64 {
	baseGravity := 80.0
	if rarity >= 3 {
		return -40.0 // Float upward for epic/legendary
	}
	return baseGravity - float64(rarity)*30.0
}

// spawnRareSparkles spawns additional sparkle burst for rare catches.
func (s *FishingCatchParticleSystem) spawnRareSparkles(x, y float64, rarity int, seed int64) {
	sparkleCount := 8 + rarity*4
	if sparkleCount > 24 {
		sparkleCount = 24
	}

	sparkleConfig := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    sparkleCount,
		GenreID:  s.genreID,
		Seed:     seed,
		Duration: 0.8,
		SpreadX:  70.0,
		SpreadY:  70.0,
		Gravity:  -60.0, // Float upward
		MinSize:  3.0,
		MaxSize:  7.0,
		Custom:   map[string]interface{}{"sparkle_burst": true, "rarity": rarity},
	}

	s.particleSystem.SpawnParticles(s.world, sparkleConfig, x, y-10)
}

// SpawnCatchEffect allows external systems to trigger catch particles directly.
func (s *FishingCatchParticleSystem) SpawnCatchEffect(x, y float64, rarity int, isRecord bool, weight float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnCatchParticles(x, y, rarity, isRecord, weight)
}
