// Package engine provides the WeatherMeleeDamageSystem that connects weather conditions
// with physical/melee damage calculations for environmental combat synergies.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherMeleeDamageSystem modifies physical/melee damage based on weather conditions.
// Creates environmental synergies: rain makes weapons slippery (reduced damage), cold
// stiffens joints (reduced attack speed), dust provides concealment (bonus critical),
// and genre-specific effects (cyberpunk smog corrodes, radiation weakens flesh).
//
// This system bridges WeatherComponent weather types with CombatSystem physical damage.
type WeatherMeleeDamageSystem struct {
	world    *World
	rng      *rand.Rand
	genreID  string
	logger   *logrus.Entry
	modCache map[weatherDamageKey]float64 // Cache computed modifiers
}

// weatherDamageKey combines weather and damage type for cache lookup.
type weatherDamageKey struct {
	weather    particles.WeatherType
	damageType combat.DamageType
}

// NewWeatherMeleeDamageSystem creates a new weather-melee damage modifier system.
func NewWeatherMeleeDamageSystem(world *World, seed int64) *WeatherMeleeDamageSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_melee_damage",
		"seed":        seed,
	})
	logger.Debug("Creating weather melee damage system")

	return &WeatherMeleeDamageSystem{
		world:    world,
		rng:      rand.New(rand.NewSource(seed)),
		genreID:  "fantasy",
		modCache: make(map[weatherDamageKey]float64),
		logger:   logger,
	}
}

// SetGenre sets the genre for genre-specific modifiers.
func (s *WeatherMeleeDamageSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.modCache = make(map[weatherDamageKey]float64) // Clear cache on genre change
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("Genre set for weather melee damage")
	}
}

// Update processes weather-melee damage modifiers for all relevant entities.
// This system is query-based; damage modifiers are computed on-demand.
func (s *WeatherMeleeDamageSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is query-based, not update-based.
	// Damage modifiers are computed on-demand via GetDamageModifier.
}

// GetDamageModifier returns the damage multiplier for a damage type given current weather.
// Returns 1.0 if no weather is active or no synergy exists.
// Values > 1.0 indicate bonus damage, < 1.0 indicate reduced damage.
func (s *WeatherMeleeDamageSystem) GetDamageModifier(damageType combat.DamageType) float64 {
	weather := s.getCurrentWeather()
	if weather == nil {
		return 1.0
	}

	key := weatherDamageKey{weather: weather.Config.Type, damageType: damageType}
	if cached, ok := s.modCache[key]; ok {
		return cached
	}

	modifier := s.calculateModifier(weather.Config.Type, damageType)
	s.modCache[key] = modifier

	if s.logger != nil && modifier != 1.0 {
		s.logger.WithFields(logrus.Fields{
			"weather":     weather.Config.Type.String(),
			"damage_type": damageType.String(),
			"modifier":    modifier,
		}).Debug("Weather melee damage modifier calculated")
	}

	return modifier
}

// getCurrentWeather finds the active weather component in the world.
func (s *WeatherMeleeDamageSystem) getCurrentWeather() *WeatherComponent {
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

// calculateModifier computes the damage modifier for a weather-damage type combination.
func (s *WeatherMeleeDamageSystem) calculateModifier(weather particles.WeatherType, damageType combat.DamageType) float64 {
	baseModifier := s.getBaseModifier(weather, damageType)
	genreBonus := s.getGenreBonus(weather, damageType)
	// Round to 2 decimal places to avoid IEEE 754 floating-point artifacts
	return math.Round((baseModifier+genreBonus)*100) / 100
}

// getBaseModifier returns the base damage modifier for weather-damage synergies.
func (s *WeatherMeleeDamageSystem) getBaseModifier(weather particles.WeatherType, damageType combat.DamageType) float64 {
	switch weather {
	case particles.WeatherRain:
		return s.getRainModifier(damageType)
	case particles.WeatherSnow:
		return s.getSnowModifier(damageType)
	case particles.WeatherFog:
		return s.getFogModifier(damageType)
	case particles.WeatherDust, particles.WeatherSandstorm:
		return s.getDustModifier(damageType)
	case particles.WeatherAsh:
		return s.getAshModifier(damageType)
	case particles.WeatherNeonRain:
		return s.getNeonRainModifier(damageType)
	case particles.WeatherSmog:
		return s.getSmogModifier(damageType)
	case particles.WeatherRadiation:
		return s.getRadiationModifier(damageType)
	case particles.WeatherBloodRain:
		return s.getBloodRainModifier(damageType)
	default:
		return 1.0
	}
}

// getRainModifier returns modifiers for rain weather.
// Rain makes weapons slippery, reduces physical damage but boosts lightning conductivity.
func (s *WeatherMeleeDamageSystem) getRainModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 0.90 // Slippery grip reduces physical impact
	case combat.DamageLightning:
		return 1.20 // Rain conducts lightning
	case combat.DamageFire:
		return 0.85 // Fire dampened by wet conditions
	default:
		return 1.0
	}
}

// getSnowModifier returns modifiers for snow weather.
// Cold stiffens joints, ice damage boosted, physical attacks hindered.
func (s *WeatherMeleeDamageSystem) getSnowModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 0.85 // Stiff joints reduce physical effectiveness
	case combat.DamageIce:
		return 1.25 // Cold empowers ice damage
	case combat.DamageFire:
		return 0.90 // Fire struggles in cold
	default:
		return 1.0
	}
}

// getFogModifier returns modifiers for fog weather.
// Obscured vision enables sneak attacks, poison spreads in moisture.
func (s *WeatherMeleeDamageSystem) getFogModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 1.10 // Concealment enables sneak attacks
	case combat.DamagePoison:
		return 1.15 // Poison spreads in moist air
	default:
		return 1.0
	}
}

// getDustModifier returns modifiers for dust/sandstorm weather.
// Abrasive particles enhance physical damage, fire ignites particulates.
func (s *WeatherMeleeDamageSystem) getDustModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 1.10 // Abrasive particles enhance cuts
	case combat.DamageFire:
		return 1.15 // Dust ignites easily
	case combat.DamageLightning:
		return 0.85 // Static interference
	default:
		return 1.0
	}
}

// getAshModifier returns modifiers for ash weather.
// Ash particles irritate and obscure, fire feeds on embers.
func (s *WeatherMeleeDamageSystem) getAshModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 1.05 // Irritated targets more vulnerable
	case combat.DamageFire:
		return 1.15 // Fire feeds on ash
	case combat.DamagePoison:
		return 1.10 // Ash particles carry toxins
	default:
		return 1.0
	}
}

// getNeonRainModifier returns modifiers for cyberpunk neon rain.
// Acidic rain corrodes armor, boosts lightning, weakens physical attacks.
func (s *WeatherMeleeDamageSystem) getNeonRainModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 0.95 // Slick surfaces reduce grip
	case combat.DamageLightning:
		return 1.30 // Highly conductive tech-rain
	case combat.DamagePoison:
		return 1.20 // Acidic rain enhances toxicity
	default:
		return 1.0
	}
}

// getSmogModifier returns modifiers for industrial smog.
// Thick air reduces visibility, poison thrives, fire ignites particulates.
func (s *WeatherMeleeDamageSystem) getSmogModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 1.05 // Obscured vision enables strikes
	case combat.DamageFire:
		return 1.10 // Smog is flammable
	case combat.DamagePoison:
		return 1.25 // Toxic atmosphere enhances poison
	default:
		return 1.0
	}
}

// getRadiationModifier returns modifiers for radioactive weather.
// Radiation weakens flesh, increases all damage received/dealt.
func (s *WeatherMeleeDamageSystem) getRadiationModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 1.15 // Weakened flesh takes more damage
	case combat.DamagePoison:
		return 1.20 // Radiation sickness compounds toxins
	case combat.DamageFire:
		return 1.10 // Irradiated tissue burns easier
	default:
		return 1.05 // General damage increase in radiation
	}
}

// getBloodRainModifier returns modifiers for horror blood rain.
// Dark atmosphere enhances violence, physical attacks empowered.
func (s *WeatherMeleeDamageSystem) getBloodRainModifier(damageType combat.DamageType) float64 {
	switch damageType {
	case combat.DamagePhysical:
		return 1.20 // Violence empowered by horror
	case combat.DamagePoison:
		return 1.15 // Blood carries disease
	case combat.DamageFire:
		return 0.85 // Blood dampens fire
	default:
		return 1.0
	}
}

// getGenreBonus returns additional bonus based on genre-weather-damage combinations.
func (s *WeatherMeleeDamageSystem) getGenreBonus(weather particles.WeatherType, damageType combat.DamageType) float64 {
	switch s.genreID {
	case "cyberpunk":
		if weather == particles.WeatherNeonRain && damageType == combat.DamageLightning {
			return 0.10 // Extra tech-rain lightning bonus
		}
		if weather == particles.WeatherSmog && damageType == combat.DamagePoison {
			return 0.10 // Industrial toxicity bonus
		}
	case "horror":
		if weather == particles.WeatherBloodRain && damageType == combat.DamagePhysical {
			return 0.10 // Extra violence bonus in horror
		}
		if weather == particles.WeatherFog && damageType == combat.DamagePoison {
			return 0.10 // Miasma bonus
		}
	case "postapoc":
		if weather == particles.WeatherRadiation && damageType == combat.DamagePhysical {
			return 0.10 // Wasteland brutality bonus
		}
		if weather == particles.WeatherSandstorm && damageType == combat.DamagePhysical {
			return 0.05 // Desert survival bonus
		}
	case "scifi":
		if damageType == combat.DamageLightning {
			return 0.05 // Tech-based damage boosted
		}
	case "fantasy":
		// Base modifiers are already balanced for fantasy
		return 0.0
	}
	return 0.0
}

// GetCurrentWeatherType returns the current weather type for external queries.
func (s *WeatherMeleeDamageSystem) GetCurrentWeatherType() (particles.WeatherType, bool) {
	weather := s.getCurrentWeather()
	if weather == nil {
		return 0, false
	}
	return weather.Config.Type, true
}

// ClearCache clears the modifier cache, useful when weather changes.
func (s *WeatherMeleeDamageSystem) ClearCache() {
	s.modCache = make(map[weatherDamageKey]float64)
}
