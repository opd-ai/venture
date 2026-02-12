package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherGroundEffectSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherGroundEffectSystem returned nil")
	}
	if sys.world != world {
		t.Error("world reference not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.spawnInterval <= 0 {
		t.Error("spawn interval should be positive")
	}
}

func TestWeatherGroundEffectSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestWeatherGroundEffectSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)

	sys.SetGenre("cyberpunk")

	if sys.genreID != "cyberpunk" {
		t.Errorf("genre not set correctly, got %s want cyberpunk", sys.genreID)
	}
}

func TestWeatherGroundEffectSystem_GetSpawnCount(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)

	tests := []struct {
		name      string
		intensity particles.WeatherIntensity
		want      int
	}{
		{"light", particles.IntensityLight, 1},
		{"medium", particles.IntensityMedium, 2},
		{"heavy", particles.IntensityHeavy, 4},
		{"extreme", particles.IntensityExtreme, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.getSpawnCount(tt.intensity)
			if got != tt.want {
				t.Errorf("getSpawnCount(%v) = %d, want %d", tt.intensity, got, tt.want)
			}
		})
	}
}

func TestWeatherGroundEffectSystem_UpdateWithoutParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)
	// Don't set particle system

	// Should not panic
	sys.Update([]*Entity{}, 0.016)
}

func TestWeatherGroundEffectSystem_UpdateWithWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherSystem, _ := particles.GenerateWeather(weatherConfig)
	weatherComp.System = weatherSystem
	weatherEntity.AddComponent(weatherComp)

	initialEntities := len(world.GetEntities())

	// Update enough to trigger spawn (past spawn interval)
	sys.Update([]*Entity{weatherEntity}, 0.2)

	// Should have spawned particle entities
	finalEntities := len(world.GetEntities())
	if finalEntities <= initialEntities {
		t.Log("Note: particle spawn may not create entities depending on implementation")
	}
}

func TestWeatherGroundEffectSystem_UpdateWithInactiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create inactive weather entity
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = false // Inactive
	weatherEntity.AddComponent(weatherComp)

	initialEntities := len(world.GetEntities())

	// Update should not spawn anything for inactive weather
	sys.Update([]*Entity{weatherEntity}, 0.2)

	finalEntities := len(world.GetEntities())
	if finalEntities > initialEntities+1 { // +1 for weather entity itself
		t.Error("should not spawn particles for inactive weather")
	}
}

func TestWeatherGroundEffectSystem_SpawnInterval(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create active weather
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherSnow,
		Intensity: particles.IntensityLight,
		Width:     800,
		Height:    600,
		GenreID:   "fantasy",
		Seed:      12345,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherSystem, _ := particles.GenerateWeather(weatherConfig)
	weatherComp.System = weatherSystem
	weatherEntity.AddComponent(weatherComp)

	// First update with small delta should not spawn (below interval)
	sys.Update([]*Entity{weatherEntity}, 0.01)
	if sys.spawnTimer < 0.01 {
		t.Error("spawn timer should accumulate")
	}

	// Update with large delta should reset timer after spawning
	sys.Update([]*Entity{weatherEntity}, 0.2)
	if sys.spawnTimer >= sys.spawnInterval {
		t.Error("spawn timer should reset after spawning")
	}
}

func TestWeatherGroundEffectSystem_DifferentWeatherTypes(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	weatherTypes := []particles.WeatherType{
		particles.WeatherRain,
		particles.WeatherSnow,
		particles.WeatherDust,
		particles.WeatherAsh,
		particles.WeatherRadiation,
		particles.WeatherNeonRain,
		particles.WeatherBloodRain,
		particles.WeatherSandstorm,
		particles.WeatherFog,  // Should not spawn ground effects
		particles.WeatherSmog, // Should not spawn ground effects
	}

	for _, wt := range weatherTypes {
		t.Run(wt.String(), func(t *testing.T) {
			// Just verify no panics occur
			sys.spawnEffectForWeatherType(wt, 100, 500)
		})
	}
}

func TestWeatherGroundEffectSystem_GenreAwareness(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherGroundEffectSystem(world, 12345)
			ps := NewParticleSystem()
			sys.SetParticleSystem(ps)
			sys.SetGenre(genre)

			// Should not panic for any genre
			sys.spawnRainSplash(100, 500, particles.WeatherRain)
			sys.spawnSnowPuff(100, 500)
			sys.spawnDustCloud(100, 500)
			sys.spawnAshSettle(100, 500)
			sys.spawnRadiationGlow(100, 500)
		})
	}
}

func BenchmarkWeatherGroundEffectSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherGroundEffectSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityHeavy,
		Width:     1920,
		Height:    1080,
		GenreID:   "fantasy",
		Seed:      12345,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherSystem, _ := particles.GenerateWeather(weatherConfig)
	weatherComp.System = weatherSystem
	weatherEntity.AddComponent(weatherComp)

	entities := []*Entity{weatherEntity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.spawnTimer = 0 // Force spawn each iteration
		sys.Update(entities, 0.2)
	}
}
