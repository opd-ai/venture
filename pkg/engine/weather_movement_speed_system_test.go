package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// TestNewWeatherMovementSpeedSystem tests system creation.
func TestNewWeatherMovementSpeedSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherMovementSpeedSystem returned nil")
	}
	if sys.world != world {
		t.Error("World reference not set")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %q, want 'fantasy'", sys.genreID)
	}
}

// TestWeatherMovementSpeedSystem_SetGenre tests genre configuration.
func TestWeatherMovementSpeedSystem_SetGenre(t *testing.T) {
	sys := NewWeatherMovementSpeedSystem(NewWorld(), 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("SetGenre(%q): genreID = %q", genre, sys.genreID)
		}
	}
}

// TestWeatherMovementSpeedSystem_NoWeather tests behavior without weather.
func TestWeatherMovementSpeedSystem_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)

	// Create entity with velocity
	entity := world.CreateEntity()
	entity.AddComponent(&VelocityComponent{VX: 100.0, VY: 50.0})
	world.AddEntity(entity)

	// Update without weather - velocity should remain unchanged
	entities := world.GetEntities()
	sys.Update(entities, 0.5)

	vel := entity.GetVelocity()
	if vel.VX != 100.0 || vel.VY != 50.0 {
		t.Errorf("Velocity changed without weather: VX=%f, VY=%f", vel.VX, vel.VY)
	}
}

// TestWeatherMovementSpeedSystem_SnowSlowsMovement tests snow reducing speed.
func TestWeatherMovementSpeedSystem_SnowSlowsMovement(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSnow,
			Intensity: particles.IntensityMedium,
		},
	})
	world.AddEntity(weatherEntity)

	// Create player entity with velocity
	player := world.CreateEntity()
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 100.0})
	world.AddEntity(player)

	originalVX := 100.0
	originalVY := 100.0

	// Update with enough delta to pass update interval
	entities := world.GetEntities()
	sys.Update(entities, 0.5)

	vel := player.GetVelocity()

	// Snow should slow movement (multiplier < 1.0)
	if vel.VX >= originalVX || vel.VY >= originalVY {
		t.Errorf("Snow should slow movement: VX=%f (orig %f), VY=%f (orig %f)",
			vel.VX, originalVX, vel.VY, originalVY)
	}

	// Check cached modifier
	mod := sys.GetMovementSpeedModifier(player.ID)
	if mod >= 1.0 {
		t.Errorf("Snow modifier should be < 1.0, got %f", mod)
	}
}

// TestWeatherMovementSpeedSystem_RainSpeedsMovement tests rain increasing speed.
func TestWeatherMovementSpeedSystem_RainSpeedsMovement(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)
	sys.SetGenre("scifi") // Scifi doesn't reduce rain bonus

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityMedium,
		},
	})
	world.AddEntity(weatherEntity)

	// Create player entity with velocity
	player := world.CreateEntity()
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 100.0})
	world.AddEntity(player)

	originalVX := 100.0

	entities := world.GetEntities()
	sys.Update(entities, 0.5)

	vel := player.GetVelocity()

	// Rain should speed up movement (multiplier > 1.0)
	if vel.VX <= originalVX {
		t.Errorf("Rain should speed movement: VX=%f (orig %f)", vel.VX, originalVX)
	}
}

// TestWeatherMovementSpeedSystem_GenreModifiers tests genre-specific adjustments.
func TestWeatherMovementSpeedSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		weatherType particles.WeatherType
		expectFast  bool // true = speed > 1.0, false = speed < 1.0
	}{
		{"Fantasy snow slows", "fantasy", particles.WeatherSnow, false},
		{"Scifi snow less penalty", "scifi", particles.WeatherSnow, false},
		{"Horror snow double penalty", "horror", particles.WeatherSnow, false},
		{"Cyberpunk rain fast", "cyberpunk", particles.WeatherRain, true},
		{"Postapoc dust adapted", "postapoc", particles.WeatherDust, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherMovementSpeedSystem(world, 12345)
			sys.SetGenre(tt.genre)

			weatherEntity := world.CreateEntity()
			weatherEntity.AddComponent(&WeatherComponent{
				Active: true,
				Config: particles.WeatherConfig{
					Type:      tt.weatherType,
					Intensity: particles.IntensityMedium,
				},
			})
			world.AddEntity(weatherEntity)

			player := world.CreateEntity()
			player.AddComponent(&VelocityComponent{VX: 100.0, VY: 100.0})
			world.AddEntity(player)

			entities := world.GetEntities()
			sys.Update(entities, 0.5)

			mod := sys.GetMovementSpeedModifier(player.ID)
			if tt.expectFast && mod <= 1.0 {
				t.Errorf("%s: expected speed boost, got modifier %f", tt.name, mod)
			}
			if !tt.expectFast && mod >= 1.0 {
				t.Errorf("%s: expected speed penalty, got modifier %f", tt.name, mod)
			}
		})
	}
}

// TestWeatherMovementSpeedSystem_IntensityScaling tests intensity affects modifier strength.
func TestWeatherMovementSpeedSystem_IntensityScaling(t *testing.T) {
	intensities := []particles.WeatherIntensity{
		particles.IntensityLight,
		particles.IntensityMedium,
		particles.IntensityHeavy,
	}

	var lastMod float64 = 1.0
	for _, intensity := range intensities {
		world := NewWorld()
		sys := NewWeatherMovementSpeedSystem(world, 12345)
		sys.SetGenre("fantasy")

		weatherEntity := world.CreateEntity()
		weatherEntity.AddComponent(&WeatherComponent{
			Active: true,
			Config: particles.WeatherConfig{
				Type:      particles.WeatherSnow,
				Intensity: intensity,
			},
		})
		world.AddEntity(weatherEntity)

		player := world.CreateEntity()
		player.AddComponent(&VelocityComponent{VX: 100.0, VY: 100.0})
		world.AddEntity(player)

		entities := world.GetEntities()
		sys.Update(entities, 0.5)

		mod := sys.GetMovementSpeedModifier(player.ID)

		// Each intensity should produce a stronger penalty (lower modifier)
		if intensity != particles.IntensityLight && mod >= lastMod {
			t.Errorf("Intensity %v should produce stronger penalty than previous: %f >= %f",
				intensity, mod, lastMod)
		}
		lastMod = mod
	}
}

// TestWeatherMovementSpeedSystem_ClearModifier tests modifier removal.
func TestWeatherMovementSpeedSystem_ClearModifier(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)

	// Create weather and player
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSnow,
			Intensity: particles.IntensityMedium,
		},
	})
	world.AddEntity(weatherEntity)

	player := world.CreateEntity()
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 100.0})
	world.AddEntity(player)

	entities := world.GetEntities()
	sys.Update(entities, 0.5)

	// Verify modifier was applied
	if sys.GetMovementSpeedModifier(player.ID) == 1.0 {
		t.Fatal("Modifier should have been applied")
	}

	// Clear modifier
	sys.ClearModifier(player.ID)

	if sys.GetMovementSpeedModifier(player.ID) != 1.0 {
		t.Error("Modifier should be 1.0 after clearing")
	}
}

// TestWeatherMovementSpeedSystem_WeatherChange tests weather state transitions.
func TestWeatherMovementSpeedSystem_WeatherChange(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSnow,
			Intensity: particles.IntensityMedium,
		},
	}
	weatherEntity.AddComponent(weatherComp)
	world.AddEntity(weatherEntity)

	player := world.CreateEntity()
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 100.0})
	world.AddEntity(player)

	entities := world.GetEntities()
	sys.Update(entities, 0.5)

	snowMod := sys.GetMovementSpeedModifier(player.ID)

	// Change to rain
	weatherComp.Config.Type = particles.WeatherRain
	sys.Update(entities, 0.5)

	rainMod := sys.GetMovementSpeedModifier(player.ID)

	// Rain and snow should have different modifiers
	if math.Abs(snowMod-rainMod) < 0.01 {
		t.Errorf("Weather change should alter modifier: snow=%f, rain=%f", snowMod, rainMod)
	}
}

// TestWeatherMovementSpeedSystem_FogNoChange tests fog doesn't affect speed.
func TestWeatherMovementSpeedSystem_FogNoChange(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityHeavy,
		},
	})
	world.AddEntity(weatherEntity)

	player := world.CreateEntity()
	player.AddComponent(&VelocityComponent{VX: 100.0, VY: 100.0})
	world.AddEntity(player)

	entities := world.GetEntities()
	sys.Update(entities, 0.5)

	mod := sys.GetMovementSpeedModifier(player.ID)
	if mod != 1.0 {
		t.Errorf("Fog should not affect speed: modifier=%f", mod)
	}
}

// TestWeatherMovementSpeedSystem_NoVelocity tests entities without velocity are skipped.
func TestWeatherMovementSpeedSystem_NoVelocity(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSnow,
			Intensity: particles.IntensityMedium,
		},
	})
	world.AddEntity(weatherEntity)

	// Entity without velocity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.AddEntity(entity)

	entities := world.GetEntities()
	// Should not panic
	sys.Update(entities, 0.5)

	// No modifier should be cached
	if sys.GetMovementSpeedModifier(entity.ID) != 1.0 {
		t.Error("Entity without velocity should have no modifier")
	}
}

// BenchmarkWeatherMovementSpeedSystem_Update benchmarks system performance.
func BenchmarkWeatherMovementSpeedSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherMovementSpeedSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create weather
	weatherEntity := world.CreateEntity()
	weatherEntity.AddComponent(&WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSnow,
			Intensity: particles.IntensityMedium,
		},
	})
	world.AddEntity(weatherEntity)

	// Create 100 entities with velocity
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&VelocityComponent{VX: 100.0, VY: 100.0})
		world.AddEntity(entity)
	}

	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016) // 60 FPS delta
	}
}
