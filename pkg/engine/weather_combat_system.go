package engine

import (
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherCombatSystem connects the WeatherSystem and StatusEffectSystem by
// applying weather-based status effects to entities. For example, rain applies
// a "wet" debuff (vulnerability), snow applies "chilled" (slow), and sandstorm
// applies periodic damage.
//
// This system iterates entities each frame. For each entity with a "health"
// component, it checks whether a weather entity is active and applies a
// matching status effect if one is not already present.
type WeatherCombatSystem struct {
	world  *World
	logger *logrus.Entry

	// Cooldown tracking to avoid applying effects every frame.
	// Maps entity ID to time remaining before next weather effect can be applied.
	cooldowns map[uint64]float64

	// How often (seconds) a new weather status effect can be applied per entity.
	applyCooldown float64
}

// NewWeatherCombatSystem creates a new WeatherCombatSystem.
func NewWeatherCombatSystem(world *World) *WeatherCombatSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_combat")
	}
	return &WeatherCombatSystem{
		world:         world,
		logger:        logEntry,
		cooldowns:     make(map[uint64]float64, 32),
		applyCooldown: 5.0, // Apply weather effects every 5 seconds
	}
}

// Update processes all entities and applies weather-based status effects.
func (s *WeatherCombatSystem) Update(entities []*Entity, deltaTime float64) {
	// First, find the active weather type from any weather entity.
	weatherType, intensity, found := s.findActiveWeather(entities)
	if !found {
		return
	}

	// Update cooldowns.
	for id, cd := range s.cooldowns {
		s.cooldowns[id] = cd - deltaTime
		if s.cooldowns[id] <= 0 {
			delete(s.cooldowns, id)
		}
	}

	// Apply weather effects to entities with health (players, NPCs, enemies).
	for _, entity := range entities {
		if !entity.HasComponent("health") {
			continue
		}

		// Skip entities still on cooldown.
		if _, onCooldown := s.cooldowns[entity.ID]; onCooldown {
			continue
		}

		// Skip if entity already has a weather-sourced status effect.
		if s.hasWeatherEffect(entity) {
			continue
		}

		effect := s.weatherEffect(weatherType, intensity)
		if effect == nil {
			continue
		}

		entity.AddComponent(effect)
		s.cooldowns[entity.ID] = s.applyCooldown

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entityID":    entity.ID,
				"weather":     weatherType.String(),
				"effect_type": effect.EffectType,
				"magnitude":   effect.Magnitude,
				"duration":    effect.Duration,
			}).Debug("applied weather status effect")
		}
	}
}

// findActiveWeather scans entities for an active WeatherComponent and returns
// the weather type and intensity. Returns found=false if no active weather.
func (s *WeatherCombatSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}
		weather, ok := comp.(*WeatherComponent)
		if !ok || !weather.Active {
			continue
		}
		return weather.Config.Type, weather.Config.Intensity, true
	}
	return 0, 0, false
}

// hasWeatherEffect checks if an entity already has a weather-sourced status effect.
func (s *WeatherCombatSystem) hasWeatherEffect(entity *Entity) bool {
	for _, comp := range entity.Components {
		effect, ok := comp.(*StatusEffectComponent)
		if !ok {
			continue
		}
		switch effect.EffectType {
		case "wet", "chilled", "sandblasted":
			return true
		}
	}
	return false
}

// intensityScale converts a WeatherIntensity enum to a 0.25–1.0 multiplier.
func intensityScale(i particles.WeatherIntensity) float64 {
	switch i {
	case particles.IntensityLight:
		return 0.25
	case particles.IntensityMedium:
		return 0.5
	case particles.IntensityHeavy:
		return 0.75
	case particles.IntensityExtreme:
		return 1.0
	default:
		return 0.25
	}
}

// weatherEffect returns a status effect matching the given weather type,
// scaled by intensity. Returns nil for weather types with no combat effect
// (e.g., fog, dust).
func (s *WeatherCombatSystem) weatherEffect(wt particles.WeatherType, intensity particles.WeatherIntensity) *StatusEffectComponent {
	scale := intensityScale(intensity)
	switch wt {
	case particles.WeatherRain:
		// Rain: "wet" debuff — vulnerability to damage, scales with intensity.
		return NewStatusEffectComponent("wet", 0.1*scale, 8.0, 0)

	case particles.WeatherSnow:
		// Snow: "chilled" debuff — slows movement, scales with intensity.
		return NewStatusEffectComponent("chilled", 0.15*scale, 6.0, 0)

	case particles.WeatherSandstorm:
		// Sandstorm: "sandblasted" — periodic minor damage.
		return NewStatusEffectComponent("sandblasted", 2.0*scale, 5.0, 1.0)

	case particles.WeatherAsh:
		// Ash: "burning" — periodic tick damage (reuses existing "burning" type).
		return NewStatusEffectComponent("burning", 1.0*scale, 4.0, 1.0)

	default:
		// Fog, Dust: no combat effect.
		return nil
	}
}
