package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherAudioSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)

	if system == nil {
		t.Fatal("NewWeatherAudioSystem returned nil")
	}
	if system.world != world {
		t.Error("world reference not set correctly")
	}
	if system.rng == nil {
		t.Error("rng not initialized")
	}
	if system.soundInterval != 2.0 {
		t.Errorf("soundInterval = %f, want 2.0", system.soundInterval)
	}
	if system.genreID != "fantasy" {
		t.Errorf("genreID = %s, want 'fantasy'", system.genreID)
	}
}

func TestWeatherAudioSystem_SetAudioManager(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)
	am := NewAudioManager(44100, 12345)

	system.SetAudioManager(am)

	if system.audioManager != am {
		t.Error("audioManager not set correctly")
	}
}

func TestWeatherAudioSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)

	system.SetGenre("scifi")
	if system.genreID != "scifi" {
		t.Errorf("genreID = %s, want 'scifi'", system.genreID)
	}

	system.SetGenre("horror")
	if system.genreID != "horror" {
		t.Errorf("genreID = %s, want 'horror'", system.genreID)
	}
}

func TestWeatherAudioSystem_UpdateNoAudioManager(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)

	// Should not panic with nil audio manager
	entities := []*Entity{}
	system.Update(entities, 0.016)
}

func TestWeatherAudioSystem_UpdateNoWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)
	am := NewAudioManager(44100, 12345)
	system.SetAudioManager(am)

	// Entity without weather component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// timeSinceSound should reset to 0 when no weather found
	if system.timeSinceSound != 0 {
		t.Errorf("timeSinceSound = %f, want 0", system.timeSinceSound)
	}
}

func TestWeatherAudioSystem_FindActiveWeather(t *testing.T) {
	tests := []struct {
		name          string
		setupWeather  func(*Entity)
		expectFound   bool
		expectType    particles.WeatherType
		expectIntense particles.WeatherIntensity
	}{
		{
			name:         "no weather component",
			setupWeather: func(e *Entity) {},
			expectFound:  false,
		},
		{
			name: "inactive weather",
			setupWeather: func(e *Entity) {
				config := particles.DefaultWeatherConfig()
				config.Type = particles.WeatherRain
				weather := NewWeatherComponent(config)
				weather.Active = false
				e.AddComponent(weather)
			},
			expectFound: false,
		},
		{
			name: "active rain weather",
			setupWeather: func(e *Entity) {
				config := particles.DefaultWeatherConfig()
				config.Type = particles.WeatherRain
				config.Intensity = particles.IntensityMedium
				weather := NewWeatherComponent(config)
				weather.Active = true
				// Create mock system
				system, _ := particles.GenerateWeather(config)
				weather.System = system
				e.AddComponent(weather)
			},
			expectFound:   true,
			expectType:    particles.WeatherRain,
			expectIntense: particles.IntensityMedium,
		},
		{
			name: "active snow weather",
			setupWeather: func(e *Entity) {
				config := particles.DefaultWeatherConfig()
				config.Type = particles.WeatherSnow
				config.Intensity = particles.IntensityHeavy
				weather := NewWeatherComponent(config)
				weather.Active = true
				system, _ := particles.GenerateWeather(config)
				weather.System = system
				e.AddComponent(weather)
			},
			expectFound:   true,
			expectType:    particles.WeatherSnow,
			expectIntense: particles.IntensityHeavy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewWeatherAudioSystem(world, 12345)

			entity := world.CreateEntity()
			tt.setupWeather(entity)

			entities := []*Entity{entity}
			weatherType, intensity, found := system.findActiveWeather(entities)

			if found != tt.expectFound {
				t.Errorf("found = %v, want %v", found, tt.expectFound)
			}
			if found && weatherType != tt.expectType {
				t.Errorf("weatherType = %v, want %v", weatherType, tt.expectType)
			}
			if found && intensity != tt.expectIntense {
				t.Errorf("intensity = %v, want %v", intensity, tt.expectIntense)
			}
		})
	}
}

func TestWeatherAudioSystem_WeatherToSFXType(t *testing.T) {
	system := NewWeatherAudioSystem(nil, 12345)

	tests := []struct {
		weather  particles.WeatherType
		expected string
	}{
		{particles.WeatherRain, "impact"},
		{particles.WeatherBloodRain, "impact"},
		{particles.WeatherNeonRain, "impact"},
		{particles.WeatherSnow, "magic"},
		{particles.WeatherSandstorm, "explosion"},
		{particles.WeatherDust, "explosion"},
		{particles.WeatherFog, "magic"},
		{particles.WeatherSmog, "magic"},
		{particles.WeatherAsh, "death"},
		{particles.WeatherRadiation, "laser"},
	}

	for _, tt := range tests {
		t.Run(tt.weather.String(), func(t *testing.T) {
			result := system.weatherToSFXType(tt.weather)
			if result != tt.expected {
				t.Errorf("weatherToSFXType(%v) = %s, want %s", tt.weather, result, tt.expected)
			}
		})
	}
}

func TestWeatherAudioSystem_CalculateInterval(t *testing.T) {
	system := NewWeatherAudioSystem(nil, 12345)

	// Test that intensity affects interval ranges
	heavyIntervals := make([]float64, 10)
	lightIntervals := make([]float64, 10)

	for i := 0; i < 10; i++ {
		heavyIntervals[i] = system.calculateInterval(particles.IntensityHeavy)
		lightIntervals[i] = system.calculateInterval(particles.IntensityLight)
	}

	// Heavy intensity should have shorter intervals on average
	heavyAvg := average(heavyIntervals)
	lightAvg := average(lightIntervals)

	if heavyAvg >= lightAvg {
		t.Errorf("heavy average (%f) should be less than light average (%f)", heavyAvg, lightAvg)
	}

	// Check bounds for heavy: 1.0-1.5
	for _, interval := range heavyIntervals {
		if interval < 1.0 || interval > 1.5 {
			t.Errorf("heavy interval %f out of bounds [1.0, 1.5]", interval)
		}
	}

	// Check bounds for light: 3.0-5.0
	for _, interval := range lightIntervals {
		if interval < 3.0 || interval > 5.0 {
			t.Errorf("light interval %f out of bounds [3.0, 5.0]", interval)
		}
	}
}

func TestWeatherAudioSystem_UpdateWithWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)
	am := NewAudioManager(44100, 12345)
	system.SetAudioManager(am)

	// Create weather entity
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherRain
	config.Intensity = particles.IntensityMedium
	weather := NewWeatherComponent(config)
	weather.Active = true
	weatherSys, _ := particles.GenerateWeather(config)
	weather.System = weatherSys
	entity.AddComponent(weather)

	entities := []*Entity{entity}

	// First update should accumulate time
	system.Update(entities, 0.5)
	if system.timeSinceSound != 0.5 {
		t.Errorf("timeSinceSound = %f, want 0.5", system.timeSinceSound)
	}

	// More updates to reach interval
	system.Update(entities, 1.0)
	system.Update(entities, 1.0) // Should trigger sound at 2.5s > 2.0s interval

	// After triggering, time resets
	if system.timeSinceSound >= 2.0 {
		t.Errorf("timeSinceSound = %f, should have reset after playing sound", system.timeSinceSound)
	}
}

func TestWeatherAudioSystem_Determinism(t *testing.T) {
	// Test that same seed produces same results
	world1 := NewWorld()
	world2 := NewWorld()

	sys1 := NewWeatherAudioSystem(world1, 99999)
	sys2 := NewWeatherAudioSystem(world2, 99999)

	// Both should produce same intervals
	for i := 0; i < 5; i++ {
		interval1 := sys1.calculateInterval(particles.IntensityMedium)
		interval2 := sys2.calculateInterval(particles.IntensityMedium)
		if interval1 != interval2 {
			t.Errorf("iteration %d: intervals differ (%f != %f)", i, interval1, interval2)
		}
	}
}

func BenchmarkWeatherAudioSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)
	am := NewAudioManager(44100, 12345)
	system.SetAudioManager(am)

	// Create weather entity
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherRain
	weather := NewWeatherComponent(config)
	weather.Active = true
	weatherSys, _ := particles.GenerateWeather(config)
	weather.System = weatherSys
	entity.AddComponent(weather)

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

func BenchmarkWeatherAudioSystem_FindActiveWeather(b *testing.B) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)

	// Create mix of entities
	entities := make([]*Entity, 100)
	for i := range entities {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i == 50 {
			// Add weather to one entity
			config := particles.DefaultWeatherConfig()
			weather := NewWeatherComponent(config)
			weather.Active = true
			weatherSys, _ := particles.GenerateWeather(config)
			weather.System = weatherSys
			entity.AddComponent(weather)
		}
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.findActiveWeather(entities)
	}
}

// Helper function
func average(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Ensure WeatherAudioSystem satisfies System interface
var _ System = (*WeatherAudioSystem)(nil)

// Test with nil world
func TestNewWeatherAudioSystem_NilWorld(t *testing.T) {
	system := NewWeatherAudioSystem(nil, 12345)
	if system == nil {
		t.Fatal("NewWeatherAudioSystem returned nil with nil world")
	}
	// Should not panic
	system.Update([]*Entity{}, 0.016)
}

// Test edge case: weather component without System field
func TestWeatherAudioSystem_WeatherWithoutSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherAudioSystem(world, 12345)
	am := NewAudioManager(44100, 12345)
	system.SetAudioManager(am)

	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	weather := NewWeatherComponent(config)
	weather.Active = true
	weather.System = nil // No weather system set
	entity.AddComponent(weather)

	entities := []*Entity{entity}
	_, _, found := system.findActiveWeather(entities)
	if found {
		t.Error("should not find weather when System is nil")
	}
}

// Suppress unused variable warning
var _ = rand.New
