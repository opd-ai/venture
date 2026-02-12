package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// BenchmarkWeatherSystem_GetWeatherParticles benchmarks particle collection with buffer reuse.
// This measures the performance improvement from reusing the particleBuffer.
func BenchmarkWeatherSystem_GetWeatherParticles(b *testing.B) {
	world := NewWorld()
	system := NewWeatherSystem(world)
	system.SetViewport(0, 0, 1920, 1080)

	// Create entity with weather
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherRain
	config.Intensity = particles.IntensityHeavy // Heavy rain = ~200 particles
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)
	weather.StartWeather()

	// Complete transition
	weather.UpdateTransition(5.0)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		particles := system.GetWeatherParticles()
		_ = particles
	}
}

// BenchmarkWeatherSystem_GetWeatherParticles_MultipleWeather benchmarks with multiple weather entities.
func BenchmarkWeatherSystem_GetWeatherParticles_MultipleWeather(b *testing.B) {
	world := NewWorld()
	system := NewWeatherSystem(world)
	system.SetViewport(0, 0, 1920, 1080)

	// Create 5 weather entities (simulating weather zones)
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		config := particles.DefaultWeatherConfig()
		config.Type = particles.WeatherRain
		config.Intensity = particles.IntensityLight // Light to keep particle count reasonable
		weather := NewWeatherComponent(config)
		entity.AddComponent(weather)
		weather.StartWeather()
		weather.UpdateTransition(5.0) // Complete transition
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		particles := system.GetWeatherParticles()
		_ = particles
	}
}

// BenchmarkWeatherSystem_GetWeatherParticles_LargeViewport benchmarks with large viewport.
func BenchmarkWeatherSystem_GetWeatherParticles_LargeViewport(b *testing.B) {
	world := NewWorld()
	system := NewWeatherSystem(world)
	system.SetViewport(0, 0, 3840, 2160) // 4K resolution

	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherSnow
	config.Intensity = particles.IntensityHeavy
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)
	weather.StartWeather()
	weather.UpdateTransition(5.0)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		particles := system.GetWeatherParticles()
		_ = particles
	}
}

// BenchmarkWeatherSystem_Update benchmarks weather system update loop.
func BenchmarkWeatherSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Create 10 entities with weather
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		config := particles.DefaultWeatherConfig()
		config.Type = particles.WeatherRain
		config.Intensity = particles.IntensityMedium
		weather := NewWeatherComponent(config)
		entity.AddComponent(weather)
		weather.StartWeather()
		entities[i] = entity
	}

	deltaTime := 0.016 // 60 FPS

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		system.Update(entities, deltaTime)
	}
}
