// Package engine provides weather ground effect visual feedback.
// This file implements a system that spawns particle effects when weather
// particles hit the ground (rain splashes, snow puffs, etc.).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// WeatherGroundEffectSystem spawns particle effects when weather impacts the ground.
// This connects WeatherSystem (tracks weather type) with ParticleSystem (visual effects)
// to create genre-aware ground impact visuals like rain splashes and snow puffs.
type WeatherGroundEffectSystem struct {
	world          *World
	particleSystem *ParticleSystem
	rng            *rand.Rand
	genreID        string

	// Spawn rate control to avoid overwhelming the particle system
	spawnTimer    float64
	spawnInterval float64 // seconds between spawn batches

	// Track last weather type to detect changes
	lastWeatherType string
}

// NewWeatherGroundEffectSystem creates a new weather ground effect system.
func NewWeatherGroundEffectSystem(world *World, seed int64) *WeatherGroundEffectSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "weather_ground_effect",
		"seed":        seed,
	}).Debug("Creating weather ground effect system")

	return &WeatherGroundEffectSystem{
		world:           world,
		rng:             rand.New(rand.NewSource(seed)),
		genreID:         "fantasy",
		spawnTimer:      0,
		spawnInterval:   0.15, // spawn effects every 150ms
		lastWeatherType: "",
	}
}

// SetParticleSystem sets the particle system for spawning effects.
func (s *WeatherGroundEffectSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetGenre sets the genre ID for color theming.
func (s *WeatherGroundEffectSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes weather entities and spawns ground effect particles.
func (s *WeatherGroundEffectSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.spawnTimer += deltaTime
	if s.spawnTimer < s.spawnInterval {
		return
	}
	s.spawnTimer = 0

	// Find active weather
	weatherEntities := s.world.GetEntitiesWith("weather")
	for _, entity := range weatherEntities {
		comp, ok := entity.GetComponent("weather")
		if !ok {
			continue
		}

		weather, ok := comp.(*WeatherComponent)
		if !ok || !weather.Active || weather.System == nil {
			continue
		}

		s.spawnGroundEffects(weather)
	}
}

// spawnGroundEffects spawns particle effects based on weather type.
func (s *WeatherGroundEffectSystem) spawnGroundEffects(weather *WeatherComponent) {
	weatherType := weather.Config.Type

	// Determine spawn count based on intensity
	spawnCount := s.getSpawnCount(weather.Config.Intensity)
	if spawnCount == 0 {
		return
	}

	// Get viewport bounds for spawning
	viewportWidth := float64(weather.Config.Width)
	viewportHeight := float64(weather.Config.Height)

	for i := 0; i < spawnCount; i++ {
		// Random position near ground level
		x := s.rng.Float64() * viewportWidth
		y := viewportHeight - s.rng.Float64()*50 // Bottom 50 pixels

		s.spawnEffectForWeatherType(weatherType, x, y)
	}
}

// getSpawnCount returns how many effects to spawn based on intensity.
func (s *WeatherGroundEffectSystem) getSpawnCount(intensity particles.WeatherIntensity) int {
	switch intensity {
	case particles.IntensityLight:
		return 1
	case particles.IntensityMedium:
		return 2
	case particles.IntensityHeavy:
		return 4
	case particles.IntensityExtreme:
		return 6
	default:
		return 1
	}
}

// spawnEffectForWeatherType spawns the appropriate particle effect for the weather.
func (s *WeatherGroundEffectSystem) spawnEffectForWeatherType(weatherType particles.WeatherType, x, y float64) {
	switch weatherType {
	case particles.WeatherRain, particles.WeatherNeonRain, particles.WeatherBloodRain:
		s.spawnRainSplash(x, y, weatherType)
	case particles.WeatherSnow:
		s.spawnSnowPuff(x, y)
	case particles.WeatherDust, particles.WeatherSandstorm:
		s.spawnDustCloud(x, y)
	case particles.WeatherAsh:
		s.spawnAshSettle(x, y)
	case particles.WeatherRadiation:
		s.spawnRadiationGlow(x, y)
	}
	// Fog and Smog don't have ground impacts
}

// spawnRainSplash creates a splash effect for rain impacts.
func (s *WeatherGroundEffectSystem) spawnRainSplash(x, y float64, weatherType particles.WeatherType) {
	config := particles.Config{
		Type:     particles.ParticleSpark,
		Count:    4,
		GenreID:  s.genreID,
		Seed:     s.rng.Int63(),
		Duration: 0.3,
		SpreadX:  20.0,
		SpreadY:  30.0,
		Gravity:  150.0,
		MinSize:  1.0,
		MaxSize:  2.0,
		Custom:   make(map[string]interface{}),
	}

	// Adjust color based on rain type via custom params
	switch weatherType {
	case particles.WeatherBloodRain:
		config.Custom["color"] = "blood"
	case particles.WeatherNeonRain:
		config.Custom["color"] = "neon"
	default:
		config.Custom["color"] = "water"
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnSnowPuff creates a soft puff effect when snow lands.
func (s *WeatherGroundEffectSystem) spawnSnowPuff(x, y float64) {
	config := particles.Config{
		Type:     particles.ParticleSmoke,
		Count:    3,
		GenreID:  s.genreID,
		Seed:     s.rng.Int63(),
		Duration: 0.5,
		SpreadX:  15.0,
		SpreadY:  10.0,
		Gravity:  -20.0, // Float up slightly
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom: map[string]interface{}{
			"color": "white",
		},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnDustCloud creates a small dust cloud effect.
func (s *WeatherGroundEffectSystem) spawnDustCloud(x, y float64) {
	config := particles.Config{
		Type:     particles.ParticleDust,
		Count:    5,
		GenreID:  s.genreID,
		Seed:     s.rng.Int63(),
		Duration: 0.8,
		SpreadX:  30.0,
		SpreadY:  15.0,
		Gravity:  -10.0,
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnAshSettle creates a settling ash effect.
func (s *WeatherGroundEffectSystem) spawnAshSettle(x, y float64) {
	config := particles.Config{
		Type:     particles.ParticleSmoke,
		Count:    2,
		GenreID:  s.genreID,
		Seed:     s.rng.Int63(),
		Duration: 1.0,
		SpreadX:  10.0,
		SpreadY:  5.0,
		Gravity:  0.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom: map[string]interface{}{
			"color": "gray",
		},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnRadiationGlow creates a brief glow effect for radiation particles.
func (s *WeatherGroundEffectSystem) spawnRadiationGlow(x, y float64) {
	config := particles.Config{
		Type:     particles.ParticleMagic,
		Count:    3,
		GenreID:  s.genreID,
		Seed:     s.rng.Int63(),
		Duration: 0.4,
		SpreadX:  15.0,
		SpreadY:  15.0,
		Gravity:  -30.0,
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom: map[string]interface{}{
			"color": "radioactive",
		},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}
