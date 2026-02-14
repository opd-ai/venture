// Package engine provides the WeatherElementalComboBonusSystem for modifying
// elemental combo damage based on current weather conditions.
// This connects WeatherSystem with ElementalComboDamageSystem to provide
// genre-aware environmental synergies for elemental combat.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherElementalComboBonusSystem modifies elemental combo damage multipliers
// based on active weather conditions. It bridges the WeatherSystem environmental
// state with ElementalComboDamageSystem combo calculations for tactical depth.
type WeatherElementalComboBonusSystem struct {
	world                      *World
	elementalComboDamageSystem *ElementalComboDamageSystem
	genreID                    string
	seed                       int64
	rng                        *rand.Rand
	logger                     *logrus.Entry

	// Weather-to-element bonus mappings
	weatherBonuses map[particles.WeatherType]map[string]float64

	// Genre-specific intensity scaling
	genreIntensityScale map[string]float64

	// Base multiplier applied to all bonuses
	baseMultiplier float64

	// Original damage to restore when weather clears
	originalBaseDamage float64
	hasModifiedDamage  bool
}

// NewWeatherElementalComboBonusSystem creates a new weather-elemental combo bonus system.
func NewWeatherElementalComboBonusSystem(world *World, seed int64) *WeatherElementalComboBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_elemental_combo_bonus")
		logEntry.Debug("weather elemental combo bonus system created")
	}

	sys := &WeatherElementalComboBonusSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		baseMultiplier: 1.0,
		genreIntensityScale: map[string]float64{
			"fantasy":   1.0,
			"scifi":     0.8, // Less elemental synergy in sci-fi
			"horror":    1.3, // Enhanced elemental effects in horror
			"cyberpunk": 0.7, // Reduced natural element synergy
			"postapoc":  1.1, // Slightly enhanced due to harsh environment
		},
	}

	sys.initializeWeatherBonuses()
	return sys
}

// initializeWeatherBonuses sets up the weather-to-combo damage bonus mappings.
// Positive values increase combo damage, negative values decrease it.
func (s *WeatherElementalComboBonusSystem) initializeWeatherBonuses() {
	s.weatherBonuses = map[particles.WeatherType]map[string]float64{
		particles.WeatherRain: {
			"steam_burst":  0.25,  // Rain + fire/ice = more steam
			"toxic_pool":   0.20,  // Rain + poison spreads toxicity
			"deep_freeze":  0.15,  // Rain + ice freezes deeper
			"toxic_flames": -0.10, // Rain dampens toxic flames
			"plasma_burst": -0.15, // Rain interferes with plasma
		},
		particles.WeatherSnow: {
			"deep_freeze":  0.30,  // Snow enhances ice combos
			"shatter":      0.25,  // Cold makes shatter more effective
			"steam_burst":  -0.15, // Cold reduces steam generation
			"toxic_flames": -0.25, // Cold dampens fire combos
		},
		particles.WeatherFog: {
			"toxic_pool":   0.20,  // Fog + poison lingers longer
			"steam_burst":  0.10,  // Humidity aids steam
			"plasma_burst": -0.10, // Fog disperses energy
		},
		particles.WeatherDust: {
			"toxic_flames": 0.20,  // Dust is combustible
			"plasma_burst": 0.15,  // Charged dust particles
			"toxic_pool":   -0.15, // Dust absorbs toxins
			"steam_burst":  -0.20, // Dry conditions reduce steam
		},
		particles.WeatherSandstorm: {
			"toxic_flames": 0.25,  // Sand is combustible
			"plasma_burst": 0.20,  // Charged sand particles
			"shatter":      0.15,  // Abrasive sand enhances shatter
			"toxic_pool":   -0.20, // Sand absorbs toxins
			"steam_burst":  -0.30, // Very dry conditions reduce steam
		},
		particles.WeatherBloodRain: {
			"toxic_flames": 0.35,  // Blood is fuel for dark fire
			"toxic_pool":   0.30,  // Blood spreads toxicity
			"steam_burst":  0.20,  // Blood creates dark steam
			"shatter":      -0.10, // Blood lubricates, less shatter
		},
		particles.WeatherRadiation: {
			"plasma_burst": 0.40,  // Radiation enhances energy combos
			"shatter":      0.25,  // Brittle irradiated materials
			"toxic_pool":   0.20,  // Radiation enhances toxicity
			"deep_freeze":  -0.20, // Heat interferes with ice
		},
	}
}

// SetElementalComboDamageSystem sets the elemental combo damage system to modify.
func (s *WeatherElementalComboBonusSystem) SetElementalComboDamageSystem(sys *ElementalComboDamageSystem) {
	s.elementalComboDamageSystem = sys
	if sys != nil {
		s.originalBaseDamage = sys.GetBaseDamage()
	}
	if s.logger != nil {
		s.logger.Debug("elemental combo damage system linked")
	}
}

// SetGenre sets the genre ID for genre-aware bonus scaling.
func (s *WeatherElementalComboBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes weather conditions and applies combo damage bonuses.
func (s *WeatherElementalComboBonusSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil || s.elementalComboDamageSystem == nil {
		return
	}

	// Find current weather
	weatherComp := s.getCurrentWeather()
	if weatherComp == nil || !weatherComp.Active {
		// Restore original damage when no weather
		if s.hasModifiedDamage {
			s.elementalComboDamageSystem.SetBaseDamage(s.originalBaseDamage)
			s.hasModifiedDamage = false
		}
		return
	}

	weatherType := weatherComp.Config.Type

	// Get bonuses for current weather
	bonuses, exists := s.weatherBonuses[weatherType]
	if !exists {
		return
	}

	// Apply strongest bonus scaled by genre
	genreScale := s.getGenreIntensityScale()
	strongestBonus := 0.0

	for _, bonus := range bonuses {
		scaledBonus := bonus * genreScale * s.baseMultiplier
		if scaledBonus > strongestBonus || (strongestBonus == 0 && scaledBonus != 0) {
			strongestBonus = scaledBonus
		}
	}

	// Apply bonus to base damage
	if strongestBonus != 0 {
		adjustedDamage := s.originalBaseDamage * (1.0 + strongestBonus)

		// Clamp to reasonable bounds (50% to 200% of original)
		if adjustedDamage < s.originalBaseDamage*0.5 {
			adjustedDamage = s.originalBaseDamage * 0.5
		}
		if adjustedDamage > s.originalBaseDamage*2.0 {
			adjustedDamage = s.originalBaseDamage * 2.0
		}

		s.elementalComboDamageSystem.SetBaseDamage(adjustedDamage)
		s.hasModifiedDamage = true
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"weather_type":    weatherType.String(),
			"genre_scale":     genreScale,
			"strongest_bonus": strongestBonus,
			"combo_count":     len(bonuses),
		}).Debug("weather elemental combo bonuses applied")
	}
}

// getCurrentWeather finds the active weather component from world entities.
func (s *WeatherElementalComboBonusSystem) getCurrentWeather() *WeatherComponent {
	if s.world == nil {
		return nil
	}
	for _, entity := range s.world.GetEntities() {
		if comp, ok := entity.GetComponent("weather"); ok {
			if weather, ok := comp.(*WeatherComponent); ok && weather.Active {
				return weather
			}
		}
	}
	return nil
}

// getGenreIntensityScale returns the intensity scaling factor for the current genre.
func (s *WeatherElementalComboBonusSystem) getGenreIntensityScale() float64 {
	if scale, ok := s.genreIntensityScale[s.genreID]; ok {
		return scale
	}
	return 1.0
}

// GetWeatherBonus returns the damage bonus for a specific weather and combo type.
func (s *WeatherElementalComboBonusSystem) GetWeatherBonus(weatherType particles.WeatherType, comboName string) float64 {
	if bonuses, exists := s.weatherBonuses[weatherType]; exists {
		if bonus, ok := bonuses[comboName]; ok {
			return bonus * s.getGenreIntensityScale()
		}
	}
	return 0.0
}

// SetBaseMultiplier sets the base multiplier for all weather bonuses.
func (s *WeatherElementalComboBonusSystem) SetBaseMultiplier(mult float64) {
	if mult > 0 {
		s.baseMultiplier = mult
	}
}

// GetBaseMultiplier returns the current base multiplier.
func (s *WeatherElementalComboBonusSystem) GetBaseMultiplier() float64 {
	return s.baseMultiplier
}
