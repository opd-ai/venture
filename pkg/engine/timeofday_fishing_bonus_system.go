// Package engine provides the TimeOfDayFishingBonusSystem which bridges time-of-day
// with fishing mechanics. This system modifies fishing bonuses based on the current
// time period, making dawn/dusk boost rare fish chance and night enhance legendary catches.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayFishingBonusSystem modifies fishing spot bonuses based on time of day.
// Dawn increases rare fish chance by 15-25%, dusk by 20-30%, night enhances legendary
// fish by 40-60%. This connects TimeOfDayLightingSystem with FishingSystem to create
// immersive day/night fishing mechanics.
type TimeOfDayFishingBonusSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often we check time (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Cache original rare fish bonuses to restore
	originalBonuses map[uint64]float64

	// Current time state cache
	lastTimeOfDay palette.TimeOfDay

	// Genre affects time-of-day fishing bonuses
	genreID string

	// lightingSystem provides current time of day
	lightingSystem *TimeOfDayLightingSystem

	// fishingSystem receives time of day callbacks
	fishingSystem *FishingSystem
}

// NewTimeOfDayFishingBonusSystem creates a new time-of-day fishing bonus system.
func NewTimeOfDayFishingBonusSystem(world *World, seed int64) *TimeOfDayFishingBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_fishing_bonus")
		logEntry.Debug("TimeOfDayFishingBonusSystem created")
	}

	return &TimeOfDayFishingBonusSystem{
		world:           world,
		logger:          logEntry,
		rng:             rand.New(rand.NewSource(seed)),
		updateInterval:  1.0, // Check once per second
		originalBonuses: make(map[uint64]float64),
		lastTimeOfDay:   palette.TimeOfDayDay,
		genreID:         "fantasy",
	}
}

// SetGenre sets the genre for genre-aware fishing bonuses.
func (s *TimeOfDayFishingBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for time-of-day fishing bonus")
	}
}

// SetLightingSystem sets the reference to TimeOfDayLightingSystem.
func (s *TimeOfDayFishingBonusSystem) SetLightingSystem(ls *TimeOfDayLightingSystem) {
	s.lightingSystem = ls
}

// SetFishingSystem sets the reference to FishingSystem for time of day callback.
func (s *TimeOfDayFishingBonusSystem) SetFishingSystem(fs *FishingSystem) {
	s.fishingSystem = fs
	if fs != nil {
		// Set up the time of day callback to use our time detection
		fs.CurrentTimeOfDay = s.getCurrentTimeOfDayForFishing
	}
}

// Update checks time of day and modifies fishing spot bonuses.
func (s *TimeOfDayFishingBonusSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	if s.world == nil || s.lightingSystem == nil {
		return
	}

	// Get current time of day from lighting system
	currentTime := s.lightingSystem.GetCurrentTimeOfDay()

	// If time changed, update fishing bonuses
	if currentTime != s.lastTimeOfDay {
		s.updateFishingBonuses(entities, currentTime)
		s.lastTimeOfDay = currentTime

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"time_of_day":    currentTime.String(),
				"spots_affected": len(s.originalBonuses),
			}).Debug("Time of day changed, fishing bonuses updated")
		}
	}
}

// updateFishingBonuses modifies all fishing spot bonuses based on time of day.
func (s *TimeOfDayFishingBonusSystem) updateFishingBonuses(entities []*Entity, timeOfDay palette.TimeOfDay) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("fishing_spot")
		if !ok {
			continue
		}

		spot, ok := comp.(*FishingSpotComponent)
		if !ok {
			continue
		}

		// Store original bonus if not already stored
		if _, exists := s.originalBonuses[entity.ID]; !exists {
			s.originalBonuses[entity.ID] = spot.RareFishBonus
		}

		// Apply time-based modifier
		modifier := s.calculateBonusModifier(timeOfDay, spot.WaterType)
		spot.RareFishBonus = s.originalBonuses[entity.ID] * modifier
	}
}

// calculateBonusModifier returns the rare fish bonus multiplier based on time of day.
func (s *TimeOfDayFishingBonusSystem) calculateBonusModifier(timeOfDay palette.TimeOfDay, waterType WaterType) float64 {
	baseModifier := s.getBaseModifier(timeOfDay, waterType)
	genreBonus := s.getGenreBonus(timeOfDay)

	return baseModifier * genreBonus
}

// getBaseModifier returns the base bonus modifier for time of day and water type.
func (s *TimeOfDayFishingBonusSystem) getBaseModifier(timeOfDay palette.TimeOfDay, waterType WaterType) float64 {
	switch timeOfDay {
	case palette.TimeOfDayDawn:
		// Dawn is excellent for fishing - fish are most active
		if waterType == WaterTypeFreshwater {
			return 1.35 // 35% bonus for freshwater at dawn
		}
		return 1.25 // 25% bonus for other water types

	case palette.TimeOfDayDay:
		// Day is standard fishing conditions
		return 1.0

	case palette.TimeOfDayDusk:
		// Dusk triggers feeding activity
		if waterType == WaterTypeSaltwater {
			return 1.40 // 40% bonus for saltwater at dusk
		}
		return 1.30 // 30% bonus for other water types

	case palette.TimeOfDayNight:
		// Night attracts rare and legendary fish
		if waterType == WaterTypeMagical {
			return 1.60 // 60% bonus for magical water at night
		}
		return 1.45 // 45% bonus for other water types

	default:
		return 1.0
	}
}

// getGenreBonus returns additional bonus based on genre synergies with time of day.
func (s *TimeOfDayFishingBonusSystem) getGenreBonus(timeOfDay palette.TimeOfDay) float64 {
	switch s.genreID {
	case "fantasy":
		// Fantasy benefits from mystical dawn and magical night
		if timeOfDay == palette.TimeOfDayDawn || timeOfDay == palette.TimeOfDayNight {
			return 1.15
		}
	case "scifi":
		// Sci-fi has artificial lighting, less affected by time
		return 0.95
	case "horror":
		// Horror greatly benefits from night fishing
		if timeOfDay == palette.TimeOfDayNight {
			return 1.35
		}
		if timeOfDay == palette.TimeOfDayDusk {
			return 1.20
		}
	case "cyberpunk":
		// Cyberpunk night life enhances night fishing
		if timeOfDay == palette.TimeOfDayNight {
			return 1.25
		}
	case "postapoc":
		// Post-apocalyptic dawn survival fishing bonus
		if timeOfDay == palette.TimeOfDayDawn {
			return 1.20
		}
	}
	return 1.0
}

// getCurrentTimeOfDayForFishing converts palette.TimeOfDay to engine.TimeOfDay.
func (s *TimeOfDayFishingBonusSystem) getCurrentTimeOfDayForFishing() TimeOfDay {
	if s.lightingSystem == nil {
		return TimeDay
	}

	paletteTime := s.lightingSystem.GetCurrentTimeOfDay()

	switch paletteTime {
	case palette.TimeOfDayDawn:
		return TimeDawn
	case palette.TimeOfDayDay:
		return TimeDay
	case palette.TimeOfDayDusk:
		return TimeDusk
	case palette.TimeOfDayNight:
		return TimeNight
	default:
		return TimeDay
	}
}

// GetCurrentTimeOfDay returns the current time of day for external queries.
func (s *TimeOfDayFishingBonusSystem) GetCurrentTimeOfDay() palette.TimeOfDay {
	return s.lastTimeOfDay
}

// GetBonusModifier returns the current bonus modifier for a given water type.
func (s *TimeOfDayFishingBonusSystem) GetBonusModifier(waterType WaterType) float64 {
	return s.calculateBonusModifier(s.lastTimeOfDay, waterType)
}
