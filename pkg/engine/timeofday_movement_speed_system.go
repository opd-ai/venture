// Package engine provides the TimeOfDayMovementSpeedSystem which bridges time-of-day
// lighting with entity movement speed. During daylight entities can travel faster,
// while night travel is slower but quieter (synergizing with stealth bonuses).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayMovementSpeedSystem modifies entity movement speed based on the current
// time of day. This connects the TimeOfDayLightingSystem with movement velocity
// to create realistic day/night travel dynamics.
//
// Speed multipliers by time of day:
//   - Dawn: 0.95 (warming up, slightly slower)
//   - Day: 1.0 (full speed, baseline)
//   - Dusk: 0.90 (fading light, cautious movement)
//   - Night: 0.80 (darkness slows travel, enhanced stealth)
//
// Genre-specific modifiers:
//   - Fantasy: Night penalty reduced (darkvision common)
//   - Scifi: Minimal time impact (artificial lighting)
//   - Horror: Night penalty increased (tension, fear)
//   - Cyberpunk: Night penalty reduced (neon lighting)
//   - Postapoc: Night penalty increased (danger awareness)
type TimeOfDayMovementSpeedSystem struct {
	world   *World
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// Reference to time-of-day lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// updateInterval controls how often we check time of day (seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original velocity multipliers to avoid compounding
	originalMultipliers map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Current speed multiplier based on time
	currentMultiplier float64

	// Genre-specific night modifiers (added to base night multiplier)
	genreNightModifiers map[string]float64
}

// NewTimeOfDayMovementSpeedSystem creates a new time-of-day movement speed system.
func NewTimeOfDayMovementSpeedSystem(world *World, seed int64) *TimeOfDayMovementSpeedSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_movement_speed")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayMovementSpeedSystem created")
	}

	return &TimeOfDayMovementSpeedSystem{
		world:               world,
		rng:                 rand.New(rand.NewSource(seed)),
		logger:              logEntry,
		updateInterval:      1.0, // Check once per second
		originalMultipliers: make(map[uint64]float64, 64),
		lastTimeOfDay:       palette.TimeOfDayDay,
		currentMultiplier:   1.0,
		genreNightModifiers: map[string]float64{
			"fantasy":   0.05,  // Darkvision races, magical light
			"scifi":     0.10,  // Night vision, artificial lighting
			"horror":    -0.10, // Fear, tension in darkness
			"cyberpunk": 0.08,  // Neon lighting helps navigation
			"postapoc":  -0.05, // Ruins are dangerous at night
		},
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayMovementSpeedSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// SetGenre sets the genre ID for genre-specific movement modifiers.
func (s *TimeOfDayMovementSpeedSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all entities and modifies movement speed based on time.
func (s *TimeOfDayMovementSpeedSystem) Update(entities []*Entity, deltaTime float64) {
	if s.lightingSystem == nil {
		return
	}

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		// Apply current multiplier to entities even if time hasn't changed
		s.applySpeedMultiplier(entities)
		return
	}
	s.timeSinceCheck = 0

	// Get current time of day from lighting system
	currentTime := s.lightingSystem.GetCurrentTimeOfDay()

	// If time hasn't changed, just apply existing multiplier
	if currentTime == s.lastTimeOfDay {
		s.applySpeedMultiplier(entities)
		return
	}

	// Time changed - calculate new multiplier
	s.currentMultiplier = s.getSpeedMultiplier(currentTime)
	s.lastTimeOfDay = currentTime

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"time_of_day": currentTime.String(),
			"multiplier":  s.currentMultiplier,
		}).Debug("time of day changed, updating movement speed")
	}

	s.applySpeedMultiplier(entities)
}

// applySpeedMultiplier applies the current time-based speed multiplier to entities.
func (s *TimeOfDayMovementSpeedSystem) applySpeedMultiplier(entities []*Entity) {
	if s.currentMultiplier == 1.0 {
		return // No modification needed
	}

	for _, entity := range entities {
		// Only affect entities with velocity (moving entities)
		vel := entity.GetVelocity()
		if vel == nil {
			continue
		}

		// Skip entities without position
		if entity.GetPosition() == nil {
			continue
		}

		// Apply speed multiplier to velocity
		vel.VX *= s.currentMultiplier
		vel.VY *= s.currentMultiplier
	}
}

// getSpeedMultiplier returns the movement speed multiplier for a given time of day.
func (s *TimeOfDayMovementSpeedSystem) getSpeedMultiplier(timeOfDay palette.TimeOfDay) float64 {
	var baseMultiplier float64

	switch timeOfDay {
	case palette.TimeOfDayDawn:
		baseMultiplier = 0.95 // Early morning, slightly slower
	case palette.TimeOfDayDay:
		baseMultiplier = 1.0 // Full daylight, baseline
	case palette.TimeOfDayDusk:
		baseMultiplier = 0.90 // Evening, cautious movement
	case palette.TimeOfDayNight:
		baseMultiplier = 0.80 // Night, slow and careful
	default:
		baseMultiplier = 1.0
	}

	// Apply genre-specific modifier for night time
	if timeOfDay == palette.TimeOfDayNight {
		if modifier, ok := s.genreNightModifiers[s.genreID]; ok {
			baseMultiplier += modifier
		}
	}

	// Clamp to reasonable range (0.5 to 1.1)
	if baseMultiplier < 0.5 {
		baseMultiplier = 0.5
	}
	if baseMultiplier > 1.1 {
		baseMultiplier = 1.1
	}

	return baseMultiplier
}

// GetCurrentMultiplier returns the current speed multiplier for external queries.
func (s *TimeOfDayMovementSpeedSystem) GetCurrentMultiplier() float64 {
	return s.currentMultiplier
}
