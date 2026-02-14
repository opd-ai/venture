// Package engine provides the TimeOfDayFishingBonusParticleSystem for visual time-based
// fishing bonus feedback. This system connects TimeOfDayFishingBonusSystem with ParticleSystem
// to spawn genre-aware particle effects on fishing spots when time-of-day bonuses are active.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// TimeOfDayFishingBonusParticleSystem spawns particles on fishing spots when time-of-day
// bonuses are active. Dawn/dusk/night fishing windows are visualized with genre-aware
// particle effects, helping players identify optimal fishing times.
type TimeOfDayFishingBonusParticleSystem struct {
	world                       *World
	particleSystem              *ParticleSystem
	timeOfDayFishingBonusSystem *TimeOfDayFishingBonusSystem
	genreID                     string
	seed                        int64
	rng                         *rand.Rand
	logger                      *logrus.Entry

	// Configuration
	pulseInterval float64 // Time between bonus indicator particles
	timeSinceEmit float64 // Accumulator for pulse timing

	// Track last time of day to detect transitions
	lastTimeOfDay palette.TimeOfDay
}

// NewTimeOfDayFishingBonusParticleSystem creates a new time-of-day fishing bonus particle system.
func NewTimeOfDayFishingBonusParticleSystem(world *World, seed int64) *TimeOfDayFishingBonusParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_fishing_bonus_particle")
		logEntry.Debug("time-of-day fishing bonus particle system created")
	}

	return &TimeOfDayFishingBonusParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		genreID:       "fantasy",
		pulseInterval: 3.0, // Pulse every 3 seconds for active bonuses
		lastTimeOfDay: palette.TimeOfDayDay,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *TimeOfDayFishingBonusParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetTimeOfDayFishingBonusSystem sets the time-of-day fishing bonus system reference.
func (s *TimeOfDayFishingBonusParticleSystem) SetTimeOfDayFishingBonusSystem(tfbs *TimeOfDayFishingBonusSystem) {
	s.timeOfDayFishingBonusSystem = tfbs
	if s.logger != nil {
		s.logger.Debug("time-of-day fishing bonus system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *TimeOfDayFishingBonusParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities and spawns particles for time-based fishing bonuses.
func (s *TimeOfDayFishingBonusParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil || s.timeOfDayFishingBonusSystem == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	shouldPulse := s.timeSinceEmit >= s.pulseInterval

	// Get current time of day
	currentTime := s.timeOfDayFishingBonusSystem.GetCurrentTimeOfDay()

	// Check for time transition to spawn burst effect
	timeChanged := currentTime != s.lastTimeOfDay

	for _, entity := range entities {
		s.processEntity(entity, shouldPulse, timeChanged, currentTime)
	}

	if shouldPulse {
		s.timeSinceEmit = 0
	}

	if timeChanged {
		s.lastTimeOfDay = currentTime
		if s.logger != nil {
			s.logger.WithField("time_of_day", currentTime.String()).Debug("time of day changed")
		}
	}
}

// processEntity handles particle effects for a single fishing spot.
func (s *TimeOfDayFishingBonusParticleSystem) processEntity(entity *Entity, shouldPulse, timeChanged bool, currentTime palette.TimeOfDay) {
	comp, ok := entity.GetComponent("fishing_spot")
	if !ok {
		return
	}

	spot, ok := comp.(*FishingSpotComponent)
	if !ok {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Check if this time period provides a bonus
	bonusModifier := s.timeOfDayFishingBonusSystem.GetBonusModifier(spot.WaterType)
	hasBonus := bonusModifier > 1.05 // Only show particles if bonus > 5%

	if !hasBonus {
		return
	}

	// Spawn burst on time transition to bonus period
	if timeChanged {
		s.spawnTimeBonusAcquiredParticles(entity.ID, pos.X, pos.Y, currentTime, bonusModifier)
	}

	// Spawn pulse particles periodically for active bonus
	if shouldPulse {
		s.spawnActiveBonusPulse(entity.ID, pos.X, pos.Y, currentTime, bonusModifier)
	}
}

// spawnTimeBonusAcquiredParticles creates particles when a fishing time bonus starts.
func (s *TimeOfDayFishingBonusParticleSystem) spawnTimeBonusAcquiredParticles(entityID uint64, x, y float64, timeOfDay palette.TimeOfDay, modifier float64) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	particleType := s.getTimeParticleType(timeOfDay)
	count := s.getParticleCount(modifier)

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 1.0,
		SpreadX:  35.0,
		SpreadY:  25.0,
		Gravity:  s.getParticleGravity(timeOfDay),
		MinSize:  2.0,
		MaxSize:  5.0,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"time_fishing_bonus": true, "time_of_day": timeOfDay.String(), "modifier": modifier},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entityID,
			"time_of_day": timeOfDay.String(),
			"modifier":    modifier,
			"count":       count,
		}).Debug("time fishing bonus particles spawned")
	}
}

// spawnActiveBonusPulse creates subtle reminder particles for active time bonuses.
func (s *TimeOfDayFishingBonusParticleSystem) spawnActiveBonusPulse(entityID uint64, x, y float64, timeOfDay palette.TimeOfDay, modifier float64) {
	effectSeed := s.seed + int64(s.timeSinceEmit*1000) + int64(entityID)

	particleType := s.getTimeParticleType(timeOfDay)
	count := 3 // Subtle pulse
	if modifier >= 1.30 {
		count = 5 // More visible for strong bonus
	}

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.6,
		SpreadX:  20.0,
		SpreadY:  15.0,
		Gravity:  s.getParticleGravity(timeOfDay) * 0.5,
		MinSize:  1.5,
		MaxSize:  3.5,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"time_pulse": true, "time_of_day": timeOfDay.String()},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getTimeParticleType returns genre-appropriate particles for time of day.
func (s *TimeOfDayFishingBonusParticleSystem) getTimeParticleType(timeOfDay palette.TimeOfDay) particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		switch timeOfDay {
		case palette.TimeOfDayDawn:
			return particles.ParticleSparkle // Golden dawn sparkles
		case palette.TimeOfDayDusk:
			return particles.ParticleSparkle // Sunset shimmer
		case palette.TimeOfDayNight:
			return particles.ParticleMagic // Mystical night glow
		default:
			return particles.ParticleDust
		}
	case "scifi":
		return particles.ParticleSpark // Sensor indication
	case "horror":
		switch timeOfDay {
		case palette.TimeOfDayNight:
			return particles.ParticleSmoke // Eerie mist
		case palette.TimeOfDayDusk:
			return particles.ParticleSmoke // Creeping darkness
		default:
			return particles.ParticleDust
		}
	case "cyberpunk":
		return particles.ParticleSpark // Neon reflections
	case "postapoc":
		switch timeOfDay {
		case palette.TimeOfDayDawn:
			return particles.ParticleDust // Morning mist
		default:
			return particles.ParticleSmoke
		}
	default:
		return particles.ParticleSparkle
	}
}

// getParticleCount returns particle count scaled by bonus modifier.
func (s *TimeOfDayFishingBonusParticleSystem) getParticleCount(modifier float64) int {
	baseCount := 8

	if modifier >= 1.50 {
		return baseCount + 6 // Strong bonus (night+horror)
	}
	if modifier >= 1.30 {
		return baseCount + 4 // Good bonus
	}
	if modifier >= 1.15 {
		return baseCount + 2 // Moderate bonus
	}
	return baseCount
}

// getParticleGravity returns gravity based on time of day for thematic effect.
func (s *TimeOfDayFishingBonusParticleSystem) getParticleGravity(timeOfDay palette.TimeOfDay) float64 {
	switch timeOfDay {
	case palette.TimeOfDayDawn:
		return -30.0 // Rising dawn particles
	case palette.TimeOfDayDusk:
		return 20.0 // Settling dusk particles
	case palette.TimeOfDayNight:
		return -40.0 // Floating night particles
	default:
		return 0.0
	}
}
