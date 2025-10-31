// Package engine provides weather system management.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// TestWeatherSystem_New tests system creation.
func TestWeatherSystem_New(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	if system == nil {
		t.Fatal("NewWeatherSystem() returned nil")
	}
	if system.world != world {
		t.Error("World reference not set correctly")
	}
	if system.viewportWidth != 800 || system.viewportHeight != 600 {
		t.Error("Default viewport dimensions not set")
	}
}

// TestWeatherSystem_SetViewport tests viewport configuration.
func TestWeatherSystem_SetViewport(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	system.SetViewport(100, 200, 1920, 1080)

	if system.viewportX != 100 {
		t.Errorf("viewportX = %v, want 100", system.viewportX)
	}
	if system.viewportY != 200 {
		t.Errorf("viewportY = %v, want 200", system.viewportY)
	}
	if system.viewportWidth != 1920 {
		t.Errorf("viewportWidth = %v, want 1920", system.viewportWidth)
	}
	if system.viewportHeight != 1080 {
		t.Errorf("viewportHeight = %v, want 1080", system.viewportHeight)
	}
}

// TestWeatherSystem_Update_NoEntities tests update with no weather entities.
func TestWeatherSystem_Update_NoEntities(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Should not panic with no entities
	system.Update([]*Entity{}, 0.016)
}

// TestWeatherSystem_Update_WithWeather tests update with active weather.
func TestWeatherSystem_Update_WithWeather(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Create entity with weather component
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Intensity = particles.IntensityLight // Use light for fewer particles
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)

	// Start weather
	weather.StartWeather()

	// Update system (should update transition)
	entities := []*Entity{entity}
	system.Update(entities, 2.5) // Halfway through 5-second transition

	if weather.TransitionTime != 2.5 {
		t.Errorf("TransitionTime = %v, want 2.5", weather.TransitionTime)
	}
	if !weather.Transitioning {
		t.Error("Weather should still be transitioning")
	}
}

// TestWeatherSystem_Update_TransitionComplete tests transition completion handling.
func TestWeatherSystem_Update_TransitionComplete(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Create entity with weather
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)

	// Start weather
	weather.StartWeather()

	// Complete fade-in transition
	entities := []*Entity{entity}
	system.Update(entities, 5.0)

	if weather.Transitioning {
		t.Error("Weather should not be transitioning after completion")
	}
	if !weather.Active {
		t.Error("Weather should be active after fade-in completes")
	}
}

// TestWeatherSystem_Update_FadeOut tests fade-out handling.
func TestWeatherSystem_Update_FadeOut(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Create and start weather
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)
	weather.StartWeather()

	// Complete fade-in
	entities := []*Entity{entity}
	system.Update(entities, 5.0)

	// Start fade-out
	weather.StopWeather()

	// Complete fade-out
	system.Update(entities, 5.0)

	if weather.Active {
		t.Error("Weather should be inactive after fade-out completes")
	}
	if weather.System != nil {
		t.Error("Weather system should be cleared after fade-out")
	}
}

// TestWeatherSystem_Update_ParticlesMove tests that particles update position.
func TestWeatherSystem_Update_ParticlesMove(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Create weather with known config
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherRain
	config.Intensity = particles.IntensityLight
	config.Seed = 12345
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)

	// Start and complete transition
	weather.StartWeather()
	entities := []*Entity{entity}
	system.Update(entities, 5.0)

	// Capture initial particle position
	if weather.System == nil || len(weather.System.Particles) == 0 {
		t.Fatal("Weather system should have particles")
	}
	initialY := weather.System.Particles[0].Y

	// Update particles (rain falls downward)
	system.Update(entities, 0.1) // 100ms

	// Check that particle moved downward (Y increased)
	newY := weather.System.Particles[0].Y
	if newY <= initialY {
		t.Errorf("Rain particle should fall downward: initialY=%v, newY=%v", initialY, newY)
	}
}

// TestWeatherSystem_GetWeatherParticles tests particle retrieval for rendering.
func TestWeatherSystem_GetWeatherParticles(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Create weather
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Intensity = particles.IntensityLight
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)

	// Process entity additions
	world.Update(0)

	// Before activation, no particles
	particles := system.GetWeatherParticles()
	if len(particles) != 0 {
		t.Errorf("Should have 0 particles before activation, got %d", len(particles))
	}

	// Activate weather
	weather.StartWeather()
	entities := []*Entity{entity}
	system.Update(entities, 5.0) // Complete transition

	// After activation, should have particles
	particles = system.GetWeatherParticles()
	if len(particles) == 0 {
		t.Error("Should have particles after activation")
	}

	// Verify particle data structure
	p := particles[0]
	if p.Size <= 0 {
		t.Error("Particle size should be positive")
	}
	if p.Color.A == 0 {
		t.Error("Particle should have non-zero alpha")
	}
}

// TestWeatherSystem_GetWeatherParticles_Opacity tests transition opacity.
func TestWeatherSystem_GetWeatherParticles_Opacity(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Create weather
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Intensity = particles.IntensityLight
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)

	// Process entity additions
	world.Update(0)

	// Start weather (fading in)
	weather.StartWeather()
	entities := []*Entity{entity}
	system.Update(entities, 2.5) // 50% transition

	// Get particles (should have 50% opacity)
	particles := system.GetWeatherParticles()
	if len(particles) == 0 {
		t.Error("Should have particles during transition")
	}

	// Check that alpha is reduced (exact value depends on original particle alpha)
	// Just verify it's not full opacity
	opacity := weather.GetOpacity()
	if opacity != 0.5 {
		t.Errorf("Opacity at 50%% transition = %v, want 0.5", opacity)
	}
}

// TestWeatherSystem_GetWeatherParticles_ViewportCulling tests particle culling.
func TestWeatherSystem_GetWeatherParticles_ViewportCulling(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Set small viewport
	system.SetViewport(0, 0, 100, 100)

	// Create weather with particles spread across larger area
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Width = 1000
	config.Height = 1000
	config.Intensity = particles.IntensityLight
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)

	// Process entity additions
	world.Update(0)

	// Activate
	weather.StartWeather()
	entities := []*Entity{entity}
	system.Update(entities, 5.0)

	// Get particles (should be culled to viewport)
	particles := system.GetWeatherParticles()

	// Not all particles should be returned (some outside viewport)
	totalParticles := len(weather.System.Particles)
	if len(particles) >= totalParticles {
		t.Errorf("Expected culling: got %d/%d particles", len(particles), totalParticles)
	}
}

// TestWeatherSystem_GetActiveWeatherType tests weather type retrieval.
func TestWeatherSystem_GetActiveWeatherType(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// No weather initially
	weatherType := system.GetActiveWeatherType()
	if weatherType != "" {
		t.Errorf("Should have no active weather initially, got %v", weatherType)
	}

	// Create rain weather
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherRain
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)

	// Process entity additions
	world.Update(0)

	// Start and complete transition
	weather.StartWeather()
	entities := []*Entity{entity}
	system.Update(entities, 5.0)

	// Should return "Rain"
	weatherType = system.GetActiveWeatherType()
	if weatherType != "Rain" {
		t.Errorf("Active weather type = %v, want Rain", weatherType)
	}
}

// TestWeatherSystem_GetWeatherCount tests counting active weather entities.
func TestWeatherSystem_GetWeatherCount(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Initially 0
	count := system.GetWeatherCount()
	if count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Add weather entity
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)
	weather.StartWeather()

	// Process entity additions
	world.Update(0)

	// Count should be 1
	count = system.GetWeatherCount()
	if count != 1 {
		t.Errorf("Count after adding weather = %d, want 1", count)
	}

	// Stop weather
	weather.StopWeather()
	entities := []*Entity{entity}
	system.Update(entities, 5.0) // Complete fade-out

	// Count should be 0
	count = system.GetWeatherCount()
	if count != 0 {
		t.Errorf("Count after stopping weather = %d, want 0", count)
	}
}

// TestWeatherSystem_Update_WeatherChange tests changing weather type.
func TestWeatherSystem_Update_WeatherChange(t *testing.T) {
	world := NewWorld()
	system := NewWeatherSystem(world)

	// Create rain weather
	entity := world.CreateEntity()
	config := particles.DefaultWeatherConfig()
	config.Type = particles.WeatherRain
	weather := NewWeatherComponent(config)
	entity.AddComponent(weather)

	// Process entity additions
	world.Update(0)

	// Start rain
	weather.StartWeather()
	entities := []*Entity{entity}
	system.Update(entities, 5.0) // Complete fade-in

	// Change to snow
	newConfig := config
	newConfig.Type = particles.WeatherSnow
	weather.ChangeWeather(newConfig)

	// Should be fading out
	if !weather.Transitioning || weather.FadingIn {
		t.Error("Should be fading out when changing weather")
	}

	// Complete fade-out (system should auto-start fade-in)
	system.Update(entities, 5.0)

	// Should be fading in new weather
	if !weather.Active || !weather.Transitioning || !weather.FadingIn {
		t.Error("Should be fading in new weather after fade-out completes")
	}
	if weather.Config.Type != particles.WeatherSnow {
		t.Error("Weather type should be updated to Snow")
	}
}
