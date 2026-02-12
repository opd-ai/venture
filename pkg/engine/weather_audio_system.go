// Package engine provides the weather-audio bridge system.
// WeatherAudioSystem connects WeatherSystem with AudioManager to play
// ambient sounds based on active weather conditions.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherAudioSystem bridges the weather system with audio to play ambient sounds.
// When weather is active (rain, snow, storm, etc.), this system triggers
// appropriate ambient sound effects through the AudioManager.
//
// This system:
//   - Monitors active weather on entities with WeatherComponent
//   - Triggers genre-aware ambient sounds at intervals
//   - Adjusts sound intensity based on weather intensity
type WeatherAudioSystem struct {
	world        *World
	audioManager *AudioManager
	logger       *logrus.Entry
	rng          *rand.Rand
	genreID      string

	// Sound timing control
	soundInterval  float64 // Minimum seconds between ambient sounds
	timeSinceSound float64 // Time since last ambient sound
	lastWeather    particles.WeatherType
	lastIntensity  particles.WeatherIntensity
}

// NewWeatherAudioSystem creates a new WeatherAudioSystem.
func NewWeatherAudioSystem(world *World, seed int64) *WeatherAudioSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "weather_audio")
	}
	return &WeatherAudioSystem{
		world:          world,
		audioManager:   nil,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		soundInterval:  2.0, // Play ambient sounds every 2 seconds minimum
		timeSinceSound: 0,
		lastWeather:    particles.WeatherRain,
		lastIntensity:  particles.IntensityLight,
	}
}

// SetAudioManager sets the audio manager reference for sound playback.
func (s *WeatherAudioSystem) SetAudioManager(am *AudioManager) {
	s.audioManager = am
}

// SetGenre sets the current genre for genre-aware sound variations.
func (s *WeatherAudioSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes weather entities and triggers ambient sounds.
func (s *WeatherAudioSystem) Update(entities []*Entity, deltaTime float64) {
	if s.audioManager == nil {
		return
	}

	// Find active weather
	weatherType, intensity, found := s.findActiveWeather(entities)
	if !found {
		s.timeSinceSound = 0
		return
	}

	// Track time for sound intervals
	s.timeSinceSound += deltaTime

	// Play ambient sound at intervals
	if s.timeSinceSound >= s.soundInterval {
		s.playWeatherAmbient(weatherType, intensity)
		s.timeSinceSound = 0

		// Vary interval based on intensity
		s.soundInterval = s.calculateInterval(intensity)
	}

	s.lastWeather = weatherType
	s.lastIntensity = intensity
}

// findActiveWeather searches for an active weather component and returns its type and intensity.
func (s *WeatherAudioSystem) findActiveWeather(entities []*Entity) (particles.WeatherType, particles.WeatherIntensity, bool) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}
		weather, ok := comp.(*WeatherComponent)
		if !ok || !weather.Active || weather.System == nil {
			continue
		}
		return weather.Config.Type, weather.Config.Intensity, true
	}
	return particles.WeatherRain, particles.IntensityLight, false
}

// playWeatherAmbient triggers the appropriate ambient sound for the weather type.
func (s *WeatherAudioSystem) playWeatherAmbient(weatherType particles.WeatherType, intensity particles.WeatherIntensity) {
	effectType := s.weatherToSFXType(weatherType)
	if effectType == "" {
		return
	}

	// Generate deterministic seed from weather type and rng
	soundSeed := s.rng.Int63()

	if err := s.audioManager.PlaySFX(effectType, soundSeed); err != nil {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"weather":   weatherType.String(),
				"intensity": intensity.String(),
				"effect":    effectType,
			}).WithError(err).Debug("failed to play weather ambient sound")
		}
	} else if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"weather":   weatherType.String(),
			"intensity": intensity.String(),
			"effect":    effectType,
		}).Debug("played weather ambient sound")
	}
}

// weatherToSFXType maps weather types to appropriate SFX effect types.
// Uses existing effect types from the SFX generator.
func (s *WeatherAudioSystem) weatherToSFXType(weatherType particles.WeatherType) string {
	switch weatherType {
	case particles.WeatherRain, particles.WeatherBloodRain, particles.WeatherNeonRain:
		return "impact" // Rain drops hitting surfaces
	case particles.WeatherSnow:
		return "magic" // Soft, ethereal wind sound
	case particles.WeatherSandstorm, particles.WeatherDust:
		return "explosion" // Rumbling wind
	case particles.WeatherFog, particles.WeatherSmog:
		return "magic" // Mysterious ambient
	case particles.WeatherAsh:
		return "death" // Ominous crackling
	case particles.WeatherRadiation:
		return "laser" // Electronic humming
	default:
		return "impact"
	}
}

// calculateInterval returns the sound interval based on weather intensity.
// More intense weather = more frequent sounds.
func (s *WeatherAudioSystem) calculateInterval(intensity particles.WeatherIntensity) float64 {
	switch intensity {
	case particles.IntensityHeavy:
		return 1.0 + s.rng.Float64()*0.5 // 1.0-1.5 seconds
	case particles.IntensityMedium:
		return 2.0 + s.rng.Float64()*1.0 // 2.0-3.0 seconds
	case particles.IntensityLight:
		return 3.0 + s.rng.Float64()*2.0 // 3.0-5.0 seconds
	default:
		return 2.0
	}
}
