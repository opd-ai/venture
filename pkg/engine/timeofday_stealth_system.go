// Package engine provides the TimeOfDayStealthSystem which bridges time-of-day
// lighting with AI detection ranges. At night entities are harder to detect,
// creating tactical stealth opportunities during darkness.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayStealthSystem modifies AI detection ranges based on the current
// time of day. This connects the TimeOfDayLightingSystem with the AISystem
// to create tactical stealth gameplay where darkness provides concealment.
//
// Detection range multipliers by time of day:
//   - Dawn: 0.85 (dim light, harder to see)
//   - Day: 1.0 (full visibility, baseline)
//   - Dusk: 0.80 (fading light, good for stealth)
//   - Night: 0.55 (darkness provides cover)
//
// Genre-specific modifiers:
//   - Fantasy: Night bonus +5% (magical moonlight varies)
//   - Scifi: Night penalty -10% (infrared/thermal sensors)
//   - Horror: Night bonus +15% (oppressive darkness)
//   - Cyberpunk: Night penalty -5% (neon glow, cameras)
//   - Postapoc: Night bonus +10% (no artificial lighting)
type TimeOfDayStealthSystem struct {
	world   *World
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string

	// Reference to time-of-day lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// updateInterval controls how often we check time of day (seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original AI detection ranges to restore on time change
	originalRanges map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre-specific night modifiers (added to base night multiplier)
	genreNightModifiers map[string]float64
}

// NewTimeOfDayStealthSystem creates a new time-of-day stealth system.
func NewTimeOfDayStealthSystem(world *World, seed int64) *TimeOfDayStealthSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_stealth")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayStealthSystem created")
	}

	return &TimeOfDayStealthSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		updateInterval: 1.0, // Check once per second (time changes slowly)
		originalRanges: make(map[uint64]float64, 64),
		lastTimeOfDay:  palette.TimeOfDayDay,
		genreNightModifiers: map[string]float64{
			"fantasy":   0.05,  // Magical forests have some moonlight
			"scifi":     -0.10, // Thermal/infrared sensors work at night
			"horror":    0.15,  // Oppressive darkness aids hiding
			"cyberpunk": -0.05, // Neon lights, security cameras
			"postapoc":  0.10,  // No infrastructure = true darkness
		},
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayStealthSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// SetGenre sets the genre ID for genre-specific stealth modifiers.
func (s *TimeOfDayStealthSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all AI entities and modifies detection ranges based on time.
func (s *TimeOfDayStealthSystem) Update(entities []*Entity, deltaTime float64) {
	if s.lightingSystem == nil {
		return
	}

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	// Get current time of day from lighting system
	currentTime := s.lightingSystem.GetCurrentTimeOfDay()

	// If time hasn't changed, nothing to update
	if currentTime == s.lastTimeOfDay {
		return
	}

	// Time changed - update all AI detection ranges
	s.updateAIDetectionRanges(entities, currentTime)
	s.lastTimeOfDay = currentTime

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"time_of_day": currentTime.String(),
			"multiplier":  s.getDetectionMultiplier(currentTime),
		}).Debug("time of day changed, updating AI detection")
	}
}

// updateAIDetectionRanges modifies AI detection ranges for all entities.
func (s *TimeOfDayStealthSystem) updateAIDetectionRanges(entities []*Entity, timeOfDay palette.TimeOfDay) {
	multiplier := s.getDetectionMultiplier(timeOfDay)

	for _, entity := range entities {
		if !entity.HasComponent("ai") {
			continue
		}

		aiComp, ok := entity.GetComponent("ai")
		if !ok {
			continue
		}

		ai, ok := aiComp.(*AIComponent)
		if !ok {
			continue
		}

		// Store original range if not already stored
		if _, exists := s.originalRanges[entity.ID]; !exists {
			s.originalRanges[entity.ID] = ai.DetectionRange
		}

		// Apply time-of-day modified detection range
		originalRange := s.originalRanges[entity.ID]
		ai.DetectionRange = originalRange * multiplier
	}
}

// getDetectionMultiplier returns the detection range multiplier for a time of day.
// Lower values = harder to detect (better for stealth).
func (s *TimeOfDayStealthSystem) getDetectionMultiplier(timeOfDay palette.TimeOfDay) float64 {
	var baseMultiplier float64

	switch timeOfDay {
	case palette.TimeOfDayDawn:
		baseMultiplier = 0.85 // Dim morning light
	case palette.TimeOfDayDay:
		baseMultiplier = 1.0 // Full visibility
	case palette.TimeOfDayDusk:
		baseMultiplier = 0.80 // Fading light
	case palette.TimeOfDayNight:
		baseMultiplier = 0.55 // Darkness provides cover
	default:
		baseMultiplier = 1.0
	}

	// Apply genre-specific modifier for night time
	if timeOfDay == palette.TimeOfDayNight && s.genreID != "" {
		if modifier, ok := s.genreNightModifiers[s.genreID]; ok {
			// Negative modifier = harder to hide (higher multiplier)
			// Positive modifier = easier to hide (lower multiplier)
			baseMultiplier -= modifier
		}
	}

	// Clamp to reasonable range
	if baseMultiplier < 0.30 {
		baseMultiplier = 0.30 // Can't be completely invisible
	}
	if baseMultiplier > 1.5 {
		baseMultiplier = 1.5 // Cap detection bonus
	}

	return baseMultiplier
}

// GetCurrentMultiplier returns the current detection multiplier for debugging.
func (s *TimeOfDayStealthSystem) GetCurrentMultiplier() float64 {
	return s.getDetectionMultiplier(s.lastTimeOfDay)
}

// GetOriginalRange returns the stored original detection range for an entity.
func (s *TimeOfDayStealthSystem) GetOriginalRange(entityID uint64) (float64, bool) {
	r, ok := s.originalRanges[entityID]
	return r, ok
}
