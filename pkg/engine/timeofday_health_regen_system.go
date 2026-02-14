// Package engine provides the TimeOfDayHealthRegenSystem which bridges time-of-day
// cycles with health regeneration. Players regenerate health at different rates based
// on time of day and genre: night resting boosts regen in fantasy, daytime safety
// aids recovery in horror, and nocturnal activity penalizes regen in sci-fi.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayHealthRegenSystem modifies health regeneration rates based on time of day.
// This connects the TimeOfDayLightingSystem with HealthComponent regen rates.
//
// Health regen multipliers by time of day (base values before genre modifiers):
//   - Dawn: 1.10 (fresh start to the day)
//   - Day: 1.0 (baseline activity)
//   - Dusk: 1.05 (winding down)
//   - Night: 1.20 (resting period)
//
// Genre modifiers adjust these base values:
//   - Fantasy: Night +10% (campfire rest), Dawn +5% (blessing of dawn)
//   - Scifi: Day +5% (medical facilities), Night -10% (reduced monitoring)
//   - Horror: Day +15% (safer to rest), Night -15% (constant threat)
//   - Cyberpunk: Night +5% (body augments active), Day -5% (urban stress)
//   - Postapoc: Dawn +10% (cooler temps), Day -10% (heat stress)
type TimeOfDayHealthRegenSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// updateInterval controls how often we check time (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original health regen rates to restore
	originalRegen map[uint64]float64

	// Track active regen multiplier per entity
	activeMultipliers map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre affects which times boost health regen
	genreID string

	// Base health regen multipliers by time of day
	baseMultipliers map[palette.TimeOfDay]float64

	// Genre-specific modifiers (added to base)
	genreModifiers map[string]map[palette.TimeOfDay]float64
}

// NewTimeOfDayHealthRegenSystem creates a new time-of-day health regen system.
func NewTimeOfDayHealthRegenSystem(world *World, seed int64) *TimeOfDayHealthRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_health_regen")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayHealthRegenSystem created")
	}

	return &TimeOfDayHealthRegenSystem{
		world:             world,
		logger:            logEntry,
		rng:               rand.New(rand.NewSource(seed)),
		updateInterval:    1.0, // Check once per second
		originalRegen:     make(map[uint64]float64, 32),
		activeMultipliers: make(map[uint64]float64, 32),
		lastTimeOfDay:     palette.TimeOfDayDay,
		genreID:           "fantasy",
		baseMultipliers: map[palette.TimeOfDay]float64{
			palette.TimeOfDayDawn:  1.10, // Fresh start
			palette.TimeOfDayDay:   1.0,  // Baseline
			palette.TimeOfDayDusk:  1.05, // Winding down
			palette.TimeOfDayNight: 1.20, // Rest period
		},
		genreModifiers: map[string]map[palette.TimeOfDay]float64{
			"fantasy": {
				palette.TimeOfDayDawn:  0.05, // Blessing of dawn
				palette.TimeOfDayNight: 0.10, // Campfire rest bonus
			},
			"scifi": {
				palette.TimeOfDayDay:   0.05,  // Medical facilities active
				palette.TimeOfDayNight: -0.10, // Reduced medical monitoring
			},
			"horror": {
				palette.TimeOfDayDay:   0.15,  // Safer to rest in daylight
				palette.TimeOfDayNight: -0.15, // Constant threat at night
			},
			"cyberpunk": {
				palette.TimeOfDayDay:   -0.05, // Urban stress
				palette.TimeOfDayNight: 0.05,  // Body augments optimize at night
			},
			"postapoc": {
				palette.TimeOfDayDawn:  0.10,  // Cooler temperatures
				palette.TimeOfDayDay:   -0.10, // Heat stress reduces regen
				palette.TimeOfDayNight: 0.05,  // Night is safer
			},
		},
	}
}

// SetGenre sets the genre for genre-aware health regen modifiers.
func (s *TimeOfDayHealthRegenSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayHealthRegenSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// Update checks time of day and updates health regeneration rates for entities.
func (s *TimeOfDayHealthRegenSystem) Update(entities []*Entity, deltaTime float64) {
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
	multiplier := s.getHealthRegenMultiplier(currentTime)

	// If time changed, log it
	if currentTime != s.lastTimeOfDay {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"time_of_day": currentTime.String(),
				"multiplier":  multiplier,
			}).Debug("time of day changed, updating health regen")
		}
		s.lastTimeOfDay = currentTime
	}

	// Update health regen for all entities with health components
	for _, entity := range entities {
		s.updateEntityRegen(entity, multiplier)
	}
}

// updateEntityRegen applies health regen modifier to a single entity.
func (s *TimeOfDayHealthRegenSystem) updateEntityRegen(entity *Entity, multiplier float64) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	// Store original regen if not already cached
	if _, exists := s.originalRegen[entity.ID]; !exists {
		s.originalRegen[entity.ID] = health.Regen
	}

	// Apply modified regen rate
	originalRegen := s.originalRegen[entity.ID]
	health.Regen = originalRegen * multiplier
	s.activeMultipliers[entity.ID] = multiplier

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"original_regen": originalRegen,
			"modified_regen": health.Regen,
			"multiplier":     multiplier,
		}).Debug("health regen modified by time of day")
	}
}

// getHealthRegenMultiplier calculates the total health regen multiplier for the current time of day.
func (s *TimeOfDayHealthRegenSystem) getHealthRegenMultiplier(timeOfDay palette.TimeOfDay) float64 {
	// Start with base multiplier
	multiplier := s.baseMultipliers[timeOfDay]
	if multiplier == 0 {
		multiplier = 1.0
	}

	// Apply genre-specific modifier
	if genreMods, ok := s.genreModifiers[s.genreID]; ok {
		if mod, ok := genreMods[timeOfDay]; ok {
			multiplier += mod
		}
	}

	// Clamp to reasonable range [0.5, 2.0]
	if multiplier < 0.5 {
		multiplier = 0.5
	}
	if multiplier > 2.0 {
		multiplier = 2.0
	}

	return multiplier
}

// GetActiveMultiplier returns the current health regen multiplier for an entity.
// Returns 1.0 if no multiplier is set.
func (s *TimeOfDayHealthRegenSystem) GetActiveMultiplier(entityID uint64) float64 {
	if mult, ok := s.activeMultipliers[entityID]; ok {
		return mult
	}
	return 1.0
}

// GetCurrentMultiplier returns the current health regen multiplier based on time of day.
// This is useful for UI display of active bonuses.
func (s *TimeOfDayHealthRegenSystem) GetCurrentMultiplier() float64 {
	if s.lightingSystem == nil {
		return 1.0
	}
	return s.getHealthRegenMultiplier(s.lightingSystem.GetCurrentTimeOfDay())
}

// GetBonusDescription returns a human-readable description of the current bonus.
func (s *TimeOfDayHealthRegenSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()
	mult := s.getHealthRegenMultiplier(timeOfDay)

	if mult > 1.0 {
		bonus := int((mult - 1.0) * 100)
		return timeOfDay.String() + " Rest: +" + healthRegenItoa(bonus) + "% HP Regen"
	}
	if mult < 1.0 {
		penalty := int((1.0 - mult) * 100)
		return timeOfDay.String() + " Strain: -" + healthRegenItoa(penalty) + "% HP Regen"
	}
	return ""
}

// RestoreOriginalRegen restores the original health regen rate for an entity.
// Called when entity leaves the game or system is disabled.
func (s *TimeOfDayHealthRegenSystem) RestoreOriginalRegen(entityID uint64) {
	if originalRegen, exists := s.originalRegen[entityID]; exists {
		if s.world != nil {
			entity := s.world.GetEntity(entityID)
			if entity != nil {
				if healthComp, ok := entity.GetComponent("health"); ok {
					if health, ok := healthComp.(*HealthComponent); ok {
						health.Regen = originalRegen
					}
				}
			}
		}
		delete(s.originalRegen, entityID)
		delete(s.activeMultipliers, entityID)
	}
}

// healthRegenItoa converts an int to string without importing strconv.
func healthRegenItoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + healthRegenItoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return digits
}
