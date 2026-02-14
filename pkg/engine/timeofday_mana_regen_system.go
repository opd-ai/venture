// Package engine provides the TimeOfDayManaRegenSystem which bridges time-of-day
// cycles with mana regeneration. Players regenerate mana at different rates based
// on time of day and genre: night meditation boosts mana regen in fantasy, daytime
// solar energy aids recovery in sci-fi, and moonlight enhances dark magic in horror.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayManaRegenSystem modifies mana regeneration rates based on time of day.
// This connects the TimeOfDayLightingSystem with ManaComponent regen rates.
//
// Mana regen multipliers by time of day (base values before genre modifiers):
//   - Dawn: 1.05 (awakening magical energies)
//   - Day: 1.0 (baseline activity)
//   - Dusk: 1.15 (twilight mysticism)
//   - Night: 1.25 (meditation and starlight)
//
// Genre modifiers adjust these base values:
//   - Fantasy: Night +15% (starlight meditation), Dusk +10% (magical twilight)
//   - Scifi: Day +10% (solar collectors), Night -10% (reduced power grid)
//   - Horror: Night +20% (dark magic amplified), Dawn -10% (light disrupts)
//   - Cyberpunk: Night +10% (grid optimization), Dusk +5% (neon activation)
//   - Postapoc: Dawn +5% (calm before heat), Day -15% (energy conservation)
type TimeOfDayManaRegenSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// updateInterval controls how often we check time (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original mana regen rates to restore
	originalRegen map[uint64]float64

	// Track active regen multiplier per entity
	activeMultipliers map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre affects which times boost mana regen
	genreID string

	// Base mana regen multipliers by time of day
	baseMultipliers map[palette.TimeOfDay]float64

	// Genre-specific modifiers (added to base)
	genreModifiers map[string]map[palette.TimeOfDay]float64
}

// NewTimeOfDayManaRegenSystem creates a new time-of-day mana regen system.
func NewTimeOfDayManaRegenSystem(world *World, seed int64) *TimeOfDayManaRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_mana_regen")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayManaRegenSystem created")
	}

	return &TimeOfDayManaRegenSystem{
		world:             world,
		logger:            logEntry,
		rng:               rand.New(rand.NewSource(seed)),
		updateInterval:    1.0, // Check once per second
		originalRegen:     make(map[uint64]float64, 32),
		activeMultipliers: make(map[uint64]float64, 32),
		lastTimeOfDay:     palette.TimeOfDayDay,
		genreID:           "fantasy",
		baseMultipliers: map[palette.TimeOfDay]float64{
			palette.TimeOfDayDawn:  1.05, // Awakening energies
			palette.TimeOfDayDay:   1.0,  // Baseline
			palette.TimeOfDayDusk:  1.15, // Twilight mysticism
			palette.TimeOfDayNight: 1.25, // Meditation period
		},
		genreModifiers: map[string]map[palette.TimeOfDay]float64{
			"fantasy": {
				palette.TimeOfDayDusk:  0.10, // Magical twilight hour
				palette.TimeOfDayNight: 0.15, // Starlight meditation
			},
			"scifi": {
				palette.TimeOfDayDay:   0.10,  // Solar collectors active
				palette.TimeOfDayNight: -0.10, // Reduced power grid
			},
			"horror": {
				palette.TimeOfDayDawn:  -0.10, // Light disrupts dark magic
				palette.TimeOfDayNight: 0.20,  // Dark magic amplified
			},
			"cyberpunk": {
				palette.TimeOfDayDusk:  0.05, // Neon systems activation
				palette.TimeOfDayNight: 0.10, // Grid power optimization
			},
			"postapoc": {
				palette.TimeOfDayDawn: 0.05,  // Calm before heat
				palette.TimeOfDayDay:  -0.15, // Energy conservation mode
			},
		},
	}
}

// SetGenre sets the genre for genre-aware mana regen modifiers.
func (s *TimeOfDayManaRegenSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayManaRegenSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// Update checks time of day and updates mana regeneration rates for entities.
func (s *TimeOfDayManaRegenSystem) Update(entities []*Entity, deltaTime float64) {
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
	multiplier := s.getManaRegenMultiplier(currentTime)

	// If time changed, log it
	if currentTime != s.lastTimeOfDay {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"time_of_day": currentTime.String(),
				"multiplier":  multiplier,
			}).Debug("time of day changed, updating mana regen")
		}
		s.lastTimeOfDay = currentTime
	}

	// Update mana regen for all entities with mana components
	for _, entity := range entities {
		s.updateEntityRegen(entity, multiplier)
	}
}

// updateEntityRegen applies mana regen modifier to a single entity.
func (s *TimeOfDayManaRegenSystem) updateEntityRegen(entity *Entity, multiplier float64) {
	manaComp, ok := entity.GetComponent("mana")
	if !ok {
		return
	}

	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return
	}

	// Store original regen if not tracked
	if _, exists := s.originalRegen[entity.ID]; !exists {
		s.originalRegen[entity.ID] = mana.Regen
	}

	// Apply multiplier to original regen
	originalRegen := s.originalRegen[entity.ID]
	newRegen := originalRegen * multiplier
	mana.Regen = newRegen

	// Track active multiplier
	s.activeMultipliers[entity.ID] = multiplier

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"original":   originalRegen,
			"modified":   newRegen,
			"multiplier": multiplier,
		}).Debug("mana regen updated")
	}
}

// getManaRegenMultiplier calculates the total mana regen multiplier for the current time of day.
func (s *TimeOfDayManaRegenSystem) getManaRegenMultiplier(timeOfDay palette.TimeOfDay) float64 {
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

// GetActiveMultiplier returns the current mana regen multiplier for an entity.
// Returns 1.0 if no multiplier is set.
func (s *TimeOfDayManaRegenSystem) GetActiveMultiplier(entityID uint64) float64 {
	if mult, ok := s.activeMultipliers[entityID]; ok {
		return mult
	}
	return 1.0
}

// GetCurrentMultiplier returns the current mana regen multiplier based on time of day.
// This is useful for UI display of active bonuses.
func (s *TimeOfDayManaRegenSystem) GetCurrentMultiplier() float64 {
	if s.lightingSystem == nil {
		return 1.0
	}
	return s.getManaRegenMultiplier(s.lightingSystem.GetCurrentTimeOfDay())
}

// GetBonusDescription returns a human-readable description of the current bonus.
func (s *TimeOfDayManaRegenSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()
	mult := s.getManaRegenMultiplier(timeOfDay)

	if mult > 1.0 {
		bonus := int((mult - 1.0) * 100)
		return timeOfDay.String() + " Meditation: +" + manaRegenItoa(bonus) + "% MP Regen"
	}
	if mult < 1.0 {
		penalty := int((1.0 - mult) * 100)
		return timeOfDay.String() + " Disruption: -" + manaRegenItoa(penalty) + "% MP Regen"
	}
	return ""
}

// RestoreOriginalRegen restores the original mana regen rate for an entity.
// Called when entity leaves the game or system is disabled.
func (s *TimeOfDayManaRegenSystem) RestoreOriginalRegen(entityID uint64) {
	if originalRegen, exists := s.originalRegen[entityID]; exists {
		if s.world != nil {
			entity, found := s.world.GetEntity(entityID)
			if found && entity != nil {
				if manaComp, ok := entity.GetComponent("mana"); ok {
					if mana, ok := manaComp.(*ManaComponent); ok {
						mana.Regen = originalRegen
					}
				}
			}
		}
		delete(s.originalRegen, entityID)
		delete(s.activeMultipliers, entityID)
	}
}

// manaRegenItoa converts an int to string without importing strconv.
func manaRegenItoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + manaRegenItoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return digits
}
