// Package engine provides the WeatherEquipmentSheenSystem which modifies equipment
// material sheen based on active weather. Rain increases reflectivity on metal (wet look),
// snow adds frost that dampens sheen on cloth/leather, dust/sandstorms increase roughness,
// and fog slightly mutes all highlights. This bridges WeatherComponent with
// MaterialSheenComponent for environmental visual coherence.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// weatherSheenModifier holds per-weather-type modifications to material sheen.
type weatherSheenModifier struct {
	SheenScale      float64 // Multiplier on SheenIntensity
	ReflectivityAdd float64 // Added to Reflectivity (clamped 0-1)
	RoughnessAdd    float64 // Added to Roughness (clamped 0-1)
	ColorTintR      float64 // Tint multiplier on highlight color
	ColorTintG      float64
	ColorTintB      float64
	PulseSpeedScale float64 // Multiplier on PulseSpeed
}

// WeatherEquipmentSheenSystem adjusts MaterialSheenComponent values based on the
// active weather type and intensity. Metal equipment gains reflective wet-look in
// rain, cloth dampens in snow, and sandstorms roughen all surfaces.
type WeatherEquipmentSheenSystem struct {
	world   *World
	logger  *logrus.Entry
	rng     *rand.Rand
	genreID string
	seed    int64

	// Cached weather state to avoid recomputing modifiers every frame
	lastWeatherType   particles.WeatherType
	lastWeatherActive bool
	currentModifier   weatherSheenModifier

	// Throttle: only update modifiers when weather changes
	checkInterval  float64
	timeSinceCheck float64
}

// NewWeatherEquipmentSheenSystem creates a new weather equipment sheen system.
func NewWeatherEquipmentSheenSystem(world *World, seed int64) *WeatherEquipmentSheenSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "weather_equipment_sheen",
	})

	sys := &WeatherEquipmentSheenSystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logger,
		genreID:         "fantasy",
		lastWeatherType: particles.WeatherRain,
		currentModifier: neutralModifier(),
		checkInterval:   0.5,
	}

	logger.Debug("weather equipment sheen system created")
	return sys
}

// SetGenre configures genre-specific weather sheen behavior.
func (s *WeatherEquipmentSheenSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// neutralModifier returns a modifier that leaves sheen unchanged.
func neutralModifier() weatherSheenModifier {
	return weatherSheenModifier{
		SheenScale:      1.0,
		ReflectivityAdd: 0.0,
		RoughnessAdd:    0.0,
		ColorTintR:      1.0,
		ColorTintG:      1.0,
		ColorTintB:      1.0,
		PulseSpeedScale: 1.0,
	}
}

// getWeatherModifier returns sheen modifications for a given weather type and intensity.
func (s *WeatherEquipmentSheenSystem) getWeatherModifier(weatherType particles.WeatherType, intensity particles.WeatherIntensity) weatherSheenModifier {
	// Intensity multiplier: light=0.3, medium=0.6, heavy=0.85, extreme=1.0
	intensityFactor := 0.3
	switch intensity {
	case particles.IntensityMedium:
		intensityFactor = 0.6
	case particles.IntensityHeavy:
		intensityFactor = 0.85
	case particles.IntensityExtreme:
		intensityFactor = 1.0
	}

	var base weatherSheenModifier
	switch weatherType {
	case particles.WeatherRain:
		// Rain: wet surfaces increase reflectivity and sheen
		base = weatherSheenModifier{
			SheenScale:      1.0 + 0.4*intensityFactor,
			ReflectivityAdd: 0.25 * intensityFactor,
			RoughnessAdd:    -0.1 * intensityFactor, // Wet = smoother
			ColorTintR:      0.9,
			ColorTintG:      0.95,
			ColorTintB:      1.0,
			PulseSpeedScale: 1.0 + 0.3*intensityFactor,
		}
	case particles.WeatherSnow:
		// Snow: frost dampens sheen, adds cold tint
		base = weatherSheenModifier{
			SheenScale:      1.0 - 0.2*intensityFactor,
			ReflectivityAdd: 0.1 * intensityFactor,  // Frost is slightly reflective
			RoughnessAdd:    0.15 * intensityFactor, // Frost adds texture
			ColorTintR:      0.85,
			ColorTintG:      0.9,
			ColorTintB:      1.0,
			PulseSpeedScale: 1.0 - 0.2*intensityFactor, // Slower, frozen feel
		}
	case particles.WeatherDust, particles.WeatherSandstorm:
		// Dust/sand: roughens surfaces, mutes reflections
		base = weatherSheenModifier{
			SheenScale:      1.0 - 0.3*intensityFactor,
			ReflectivityAdd: -0.15 * intensityFactor,
			RoughnessAdd:    0.25 * intensityFactor,
			ColorTintR:      1.0,
			ColorTintG:      0.9,
			ColorTintB:      0.75,
			PulseSpeedScale: 1.0,
		}
	case particles.WeatherFog:
		// Fog: softly mutes everything
		base = weatherSheenModifier{
			SheenScale:      1.0 - 0.15*intensityFactor,
			ReflectivityAdd: -0.05 * intensityFactor,
			RoughnessAdd:    0.05 * intensityFactor,
			ColorTintR:      0.95,
			ColorTintG:      0.95,
			ColorTintB:      0.95,
			PulseSpeedScale: 1.0 - 0.1*intensityFactor,
		}
	default:
		base = neutralModifier()
	}

	// Genre-specific adjustments
	switch s.genreID {
	case "cyberpunk":
		// Neon reflections amplified in rain
		if weatherType == particles.WeatherRain {
			base.SheenScale *= 1.15
			base.ReflectivityAdd *= 1.2
		}
	case "horror":
		// Darker, more muted highlights
		base.SheenScale *= 0.85
		base.ColorTintR *= 0.9
		base.ColorTintG *= 0.85
		base.ColorTintB *= 0.85
	case "postapoc":
		// Extra roughness from environmental grime
		base.RoughnessAdd += 0.05
	}

	return base
}

// findActiveWeather locates the first active weather entity and returns its type and intensity.
func (s *WeatherEquipmentSheenSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
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
	return particles.WeatherRain, particles.IntensityLight, false
}

// Update modifies MaterialSheenComponent on entities based on current weather.
func (s *WeatherEquipmentSheenSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Throttle weather detection
	s.timeSinceCheck += deltaTime
	weatherChanged := false
	if s.timeSinceCheck >= s.checkInterval {
		s.timeSinceCheck = 0
		weatherType, intensity, active := s.findActiveWeather(entities)
		if active != s.lastWeatherActive || (active && weatherType != s.lastWeatherType) {
			s.lastWeatherActive = active
			s.lastWeatherType = weatherType
			if active {
				s.currentModifier = s.getWeatherModifier(weatherType, intensity)
			} else {
				s.currentModifier = neutralModifier()
			}
			weatherChanged = true
		}
	}

	// Only iterate entities when weather state changed
	if !weatherChanged {
		return
	}

	mod := s.currentModifier
	for _, entity := range entities {
		comp, ok := entity.GetComponent("material_sheen")
		if !ok {
			continue
		}
		sheen, ok := comp.(*MaterialSheenComponent)
		if !ok || !sheen.Enabled {
			continue
		}

		// Apply weather modifications
		sheen.SheenIntensity = clampF(sheen.SheenIntensity*mod.SheenScale, 0.0, 1.0)
		sheen.Reflectivity = clampF(sheen.Reflectivity+mod.ReflectivityAdd, 0.0, 1.0)
		sheen.Roughness = clampF(sheen.Roughness+mod.RoughnessAdd, 0.0, 1.0)
		sheen.ColorR *= mod.ColorTintR
		sheen.ColorG *= mod.ColorTintG
		sheen.ColorB *= mod.ColorTintB
		sheen.PulseSpeed *= mod.PulseSpeedScale
	}
}

// clampF clamps a float64 value between min and max.
func clampF(v, min, max float64) float64 {
	return math.Max(min, math.Min(max, v))
}
